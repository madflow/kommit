package provider

import (
	"context"
	"fmt"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/git"
	"github.com/openai/openai-go/v3"
)

type StreamHandler func(chunk string)

func (c *Client) GenerateCommitMessageStreaming(ctx context.Context, diff, rules string, repoCtx *git.RepoContext, opts GenerateOptions, handler StreamHandler) (string, error) {
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

	return c.generateStreamingWithFallback(ctx, prompt, opts, "commit message", handler)
}

func (c *Client) GenerateBranchNameStreaming(ctx context.Context, diff string, opts GenerateOptions, handler StreamHandler) (string, error) {
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

	return c.generateStreamingWithFallback(ctx, prompt, opts, "branch name", handler)
}

func (c *Client) GeneratePullRequestBodyStreaming(ctx context.Context, diff, rules string, repoCtx *git.RepoContext, opts GenerateOptions, handler StreamHandler) (string, error) {
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

	return c.generateStreamingWithFallback(ctx, prompt, opts, "pull request body", handler)
}

func (c *Client) GeneratePullRequestTitleStreaming(ctx context.Context, diff, rules string, repoCtx *git.RepoContext, opts GenerateOptions, handler StreamHandler) (string, error) {
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

	return c.generateStreamingWithFallback(ctx, prompt, opts, "pull request title", handler)
}

func (c *Client) generateStreamingWithFallback(ctx context.Context, prompt string, opts GenerateOptions, description string, handler StreamHandler) (string, error) {
	providerOrder := c.config.GetProviderFallbackOrder(opts.Provider)

	var lastErr error
	for _, providerName := range providerOrder {
		response, err := c.generateStreamingCompletion(ctx, providerName, opts.Model, prompt, handler)
		if err != nil {
			c.LogProviderError(providerName, err)
			lastErr = err
			continue
		}
		return response, nil
	}

	return "", fmt.Errorf("all providers failed for %s generation: %w", description, lastErr)
}

func (c *Client) generateStreamingCompletion(ctx context.Context, providerName, modelID, prompt string, handler StreamHandler) (string, error) {
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

	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(model.ID),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		MaxTokens: openai.Int(int64(maxTokens)),
	})

	var fullResponse string
	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				fullResponse += content
				if handler != nil {
					handler(content)
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		return "", fmt.Errorf("streaming failed: %w", err)
	}

	return fullResponse, nil
}
