package provider

import (
	"context"
	"fmt"

	"github.com/madflow/kommit/internal/config"
	"github.com/openai/openai-go/v3"
)

type GenerateOptions struct {
	Provider        string
	Model           string
	Stream          bool
	OnProviderError func(providerName string, err error)
}

type Kind string

const (
	KindCommitMessage    Kind = "commit_message"
	KindBranchName       Kind = "branch_name"
	KindPullRequestBody  Kind = "pull_request_body"
	KindPullRequestTitle Kind = "pull_request_title"
)

type Request struct {
	Kind          Kind
	Diff          string
	Rules         string
	PromptContext *PromptContext
	Handler       StreamHandler
}

type PromptContext struct {
	BranchName   string
	FilesChanged int
	FileChanges  []ChangedFile
	Summary      string
}

type ChangedFile struct {
	Status string
	Path   string
	Type   string
}

type requestSpec struct {
	prompt      string
	description string
}

func (c *Client) Generate(ctx context.Context, req Request, opts GenerateOptions) (string, error) {
	spec, err := buildRequestSpec(req)
	if err != nil {
		return "", err
	}

	return c.generateWithFallback(ctx, spec, opts, req.Handler)
}

func buildRequestSpec(req Request) (requestSpec, error) {
	diff := truncateDiff(req.Diff)
	promptCtx := req.PromptContext
	if promptCtx == nil {
		promptCtx = &PromptContext{}
	}

	switch req.Kind {
	case KindCommitMessage:
		return requestSpec{
			prompt: fmt.Sprintf(`
You are a git commit message generator. 
Output ONLY the commit message in plain text format with no additional text, headers, or formatting.

Repository Context:
- Branch name: %s
- Number of files changed: %d
- Changed files:%s

IMPORTANT Rules:
%s

Git diff:
%s`,
				promptCtx.BranchName,
				promptCtx.FilesChanged,
				formatFileChanges(promptCtx),
				req.Rules,
				diff,
			),
			description: "commit message",
		}, nil
	case KindBranchName:
		return requestSpec{
			prompt: fmt.Sprintf(`
You are a git branch name generator. 
Output ONLY the branch name in plain text format with no additional text, headers, or formatting.
The branch name should only contain alphanumeric characters and the symbols: -, _, /.
The branch name should not be longer than 40 characters.

Git diff:
%s`, diff),
			description: "branch name",
		}, nil
	case KindPullRequestBody:
		return requestSpec{
			prompt: fmt.Sprintf(`
You are a pull request body generator.
Output ONLY the pull request body in plain text format with no additional text, headers, or formatting.

Repository Context:
- Branch name: %s
- Number of files changed: %d
- Changed files:%s

IMPORTANT Rules:
%s

Git diff:
%s`,
				promptCtx.BranchName,
				promptCtx.FilesChanged,
				formatFileChanges(promptCtx),
				req.Rules,
				diff,
			),
			description: "pull request body",
		}, nil
	case KindPullRequestTitle:
		return requestSpec{
			prompt: fmt.Sprintf(`
You are a pull request title generator.
Output ONLY the pull request title in plain text format with no additional text, headers, or formatting.
The pull request title should only contain alphanumeric characters and the symbols: -, _, /.
The pull request title should not be longer than 50 characters.

Repository Context:
- Branch name: %s
- Number of files changed: %d
- Changed files:%s

IMPORTANT Rules:
%s

Git diff:
%s`,
				promptCtx.BranchName,
				promptCtx.FilesChanged,
				formatFileChanges(promptCtx),
				req.Rules,
				diff,
			),
			description: "pull request title",
		}, nil
	default:
		return requestSpec{}, fmt.Errorf("unsupported generation kind %q", req.Kind)
	}
}

func truncateDiff(diff string) string {
	const maxDiffLength = 4000
	if len(diff) > maxDiffLength {
		return diff[:maxDiffLength] + "\n... (truncated)"
	}

	return diff
}

func formatFileChanges(promptCtx *PromptContext) string {
	if promptCtx == nil || len(promptCtx.FileChanges) == 0 {
		return " (none)"
	}

	var files []string
	for _, change := range promptCtx.FileChanges {
		files = append(files, fmt.Sprintf("\n  - [%s] %s (%s)", change.Status, change.Path, change.Type))
	}

	return stringOrNil(files)
}

func stringOrNil(files []string) string {
	result := ""
	for _, f := range files {
		result += f
	}
	return result
}

func (c *Client) generateWithFallback(ctx context.Context, spec requestSpec, opts GenerateOptions, handler StreamHandler) (string, error) {
	providerOrder := c.settings.ProviderOrder
	if opts.Provider != "" {
		providerOrder = make([]string, 0, len(c.settings.ProviderOrder))
		providerOrder = append(providerOrder, opts.Provider)
		for _, providerName := range c.settings.ProviderOrder {
			if providerName != opts.Provider {
				providerOrder = append(providerOrder, providerName)
			}
		}
	}
	if len(providerOrder) == 0 {
		return "", fmt.Errorf("no providers configured for %s generation", spec.description)
	}

	var lastErr error
	for _, providerName := range providerOrder {
		response, err := c.executor()(ctx, providerName, opts.Model, spec.prompt, opts.Stream, handler)
		if err != nil {
			if opts.OnProviderError != nil {
				opts.OnProviderError(providerName, err)
			} else {
				c.LogProviderError(providerName, err)
			}
			lastErr = err
			continue
		}
		return response, nil
	}

	return "", fmt.Errorf("all providers failed for %s generation: %w", spec.description, lastErr)
}

func (c *Client) executeRequest(ctx context.Context, providerName, modelID, prompt string, stream bool, handler StreamHandler) (string, error) {
	client, err := c.GetOpenAIClient(providerName)
	if err != nil {
		return "", err
	}

	model := c.resolveModel(providerName, modelID)
	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(model.ID),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		MaxTokens: openai.Int(int64(defaultMaxTokens(model))),
	}

	if stream {
		return generateStreamingCompletion(ctx, client, params, handler)
	}

	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("chat completion failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	return completion.Choices[0].Message.Content, nil
}

func (c *Client) resolveModel(providerName, modelID string) *config.ModelConfig {
	model, err := c.GetModel(providerName, modelID)
	if err == nil {
		return model
	}

	return &config.ModelConfig{
		ID:               modelID,
		Name:             modelID,
		ContextWindow:    32768,
		DefaultMaxTokens: 4096,
	}
}

func defaultMaxTokens(model *config.ModelConfig) int {
	if model.DefaultMaxTokens == 0 {
		return 4096
	}

	return model.DefaultMaxTokens
}

func generateStreamingCompletion(ctx context.Context, client *openai.Client, params openai.ChatCompletionNewParams, handler StreamHandler) (string, error) {
	stream := client.Chat.Completions.NewStreaming(ctx, params)

	var fullResponse string
	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) == 0 {
			continue
		}

		content := chunk.Choices[0].Delta.Content
		if content == "" {
			continue
		}

		fullResponse += content
		if handler != nil {
			handler(content)
		}
	}

	if err := stream.Err(); err != nil {
		return "", fmt.Errorf("streaming failed: %w", err)
	}

	return fullResponse, nil
}
