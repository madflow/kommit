package provider

import (
	"context"
	"fmt"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/git"
	"github.com/openai/openai-go/v3"
)

type GenerateOptions struct {
	Provider string
	Model    string
	Stream   bool
}

func (c *Client) GenerateCommitMessage(ctx context.Context, diff, rules string, repoCtx *git.RepoContext, opts GenerateOptions) (string, error) {
	maxDiffLength := 4000
	if len(diff) > maxDiffLength {
		diff = diff[:maxDiffLength] + "\n... (truncated)"
	}

	prompt := fmt.Sprintf(`
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
		repoCtx.BranchName,
		repoCtx.FilesChanged,
		formatFileChanges(repoCtx),
		rules,
		diff)

	return c.generateWithFallback(ctx, prompt, opts, "commit message")
}

func (c *Client) GenerateBranchName(ctx context.Context, diff string, opts GenerateOptions) (string, error) {
	maxDiffLength := 4000
	if len(diff) > maxDiffLength {
		diff = diff[:maxDiffLength] + "\n... (truncated)"
	}

	prompt := fmt.Sprintf(`
You are a git branch name generator. 
Output ONLY the branch name in plain text format with no additional text, headers, or formatting.
The branch name should only contain alphanumeric characters and the symbols: -, _, /.
The branch name should not be longer than 40 characters.

Git diff:
%s`, diff)

	return c.generateWithFallback(ctx, prompt, opts, "branch name")
}

func (c *Client) GeneratePullRequestBody(ctx context.Context, diff, rules string, repoCtx *git.RepoContext, opts GenerateOptions) (string, error) {
	maxDiffLength := 4000
	if len(diff) > maxDiffLength {
		diff = diff[:maxDiffLength] + "\n... (truncated)"
	}

	prompt := fmt.Sprintf(`
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
		repoCtx.BranchName,
		repoCtx.FilesChanged,
		formatFileChanges(repoCtx),
		rules,
		diff)

	return c.generateWithFallback(ctx, prompt, opts, "pull request body")
}

func (c *Client) GeneratePullRequestTitle(ctx context.Context, diff, rules string, repoCtx *git.RepoContext, opts GenerateOptions) (string, error) {
	maxDiffLength := 4000
	if len(diff) > maxDiffLength {
		diff = diff[:maxDiffLength] + "\n... (truncated)"
	}

	prompt := fmt.Sprintf(`
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
		repoCtx.BranchName,
		repoCtx.FilesChanged,
		formatFileChanges(repoCtx),
		rules,
		diff)

	return c.generateWithFallback(ctx, prompt, opts, "pull request title")
}

func formatFileChanges(repoCtx *git.RepoContext) string {
	if len(repoCtx.FileChanges) == 0 {
		return " (none)"
	}
	var files []string
	for _, change := range repoCtx.FileChanges {
		files = append(files, fmt.Sprintf("\n  - [%s] %s (%s)", change.Status, change.FilePath, change.FileType))
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

func (c *Client) generateWithFallback(ctx context.Context, prompt string, opts GenerateOptions, description string) (string, error) {
	providerOrder := c.config.GetProviderFallbackOrder(opts.Provider)

	var lastErr error
	for _, providerName := range providerOrder {
		response, err := c.generateCompletion(ctx, providerName, opts.Model, prompt)
		if err != nil {
			c.LogProviderError(providerName, err)
			lastErr = err
			continue
		}
		return response, nil
	}

	return "", fmt.Errorf("all providers failed for %s generation: %w", description, lastErr)
}

func (c *Client) generateCompletion(ctx context.Context, providerName, modelID, prompt string) (string, error) {
	client, err := c.GetOpenAIClient(providerName)
	if err != nil {
		return "", err
	}

	model, err := c.GetModel(providerName, modelID)
	if err != nil {
		model = &config.ModelConfig{
			ID:               modelID,
			Name:             modelID,
			ContextWindow:    32768,
			DefaultMaxTokens: 4096,
		}
	}

	maxTokens := model.DefaultMaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(model.ID),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		MaxTokens: openai.Int(int64(maxTokens)),
	})
	if err != nil {
		return "", fmt.Errorf("chat completion failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	return completion.Choices[0].Message.Content, nil
}
