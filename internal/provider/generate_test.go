package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/madflow/kommit/internal/config"
)

func TestClientGenerateBuildsCommitPromptAndTruncatesDiff(t *testing.T) {
	client := NewClient(config.NewLoader("").LoadOrPanicForTest())

	var gotProvider string
	var gotModel string
	var gotPrompt string
	var gotStream bool
	client.execute = func(ctx context.Context, providerName, modelID, prompt string, stream bool, handler StreamHandler) (string, error) {
		gotProvider = providerName
		gotModel = modelID
		gotPrompt = prompt
		gotStream = stream
		return "generated", nil
	}

	diff := strings.Repeat("a", 4001)
	promptCtx := &PromptContext{
		BranchName:   "feature/deepen-provider",
		FilesChanged: 1,
		FileChanges:  []ChangedFile{{Status: "M", Path: "cmd/root.go", Type: "go"}},
	}

	result, err := client.Generate(context.Background(), Request{
		Kind:          KindCommitMessage,
		Diff:          diff,
		Rules:         "- Keep it concise",
		PromptContext: promptCtx,
	}, GenerateOptions{Provider: "ollama", Model: "custom-model"})
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}
	if result != "generated" {
		t.Fatalf("expected generated response, got %q", result)
	}
	if gotProvider != "ollama" {
		t.Fatalf("expected provider ollama, got %q", gotProvider)
	}
	if gotModel != "custom-model" {
		t.Fatalf("expected model custom-model, got %q", gotModel)
	}
	if gotStream {
		t.Fatalf("expected non-streaming execution")
	}
	if !strings.Contains(gotPrompt, "You are a git commit message generator.") {
		t.Fatalf("expected commit prompt, got %q", gotPrompt)
	}
	if !strings.Contains(gotPrompt, "feature/deepen-provider") {
		t.Fatalf("expected branch name in prompt, got %q", gotPrompt)
	}
	if !strings.Contains(gotPrompt, "[M] cmd/root.go (go)") {
		t.Fatalf("expected changed file summary in prompt, got %q", gotPrompt)
	}
	if !strings.Contains(gotPrompt, "... (truncated)") {
		t.Fatalf("expected truncated diff marker, got %q", gotPrompt)
	}
}

func TestClientGenerateFallsBackAcrossProviders(t *testing.T) {
	loader := config.NewLoader("")
	settings := loader.ResolveForTest(&config.Config{
		Providers: map[string]config.ProviderConfig{
			"primary":   {Name: "Primary"},
			"secondary": {Name: "Secondary"},
		},
		DefaultProvider: "primary",
	})
	client := NewClient(settings)

	var calls []string
	client.execute = func(ctx context.Context, providerName, modelID, prompt string, stream bool, handler StreamHandler) (string, error) {
		calls = append(calls, providerName)
		if providerName == "primary" {
			return "", errors.New("primary failed")
		}
		return "ok", nil
	}

	result, err := client.Generate(context.Background(), Request{
		Kind: KindBranchName,
		Diff: "diff",
	}, GenerateOptions{Provider: "primary"})
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}
	if result != "ok" {
		t.Fatalf("expected fallback result ok, got %q", result)
	}
	if len(calls) != 2 || calls[0] != "primary" || calls[1] != "secondary" {
		t.Fatalf("expected fallback order [primary secondary], got %v", calls)
	}
}

func TestClientGenerateUsesStreamingPath(t *testing.T) {
	client := NewClient(config.NewLoader("").LoadOrPanicForTest())

	var gotStream bool
	var chunks []string
	client.execute = func(ctx context.Context, providerName, modelID, prompt string, stream bool, handler StreamHandler) (string, error) {
		gotStream = stream
		if handler != nil {
			handler("chunk")
		}
		return "chunk", nil
	}

	result, err := client.Generate(context.Background(), Request{
		Kind:    KindPullRequestTitle,
		Diff:    "diff",
		Rules:   "- title rules",
		Handler: func(chunk string) { chunks = append(chunks, chunk) },
	}, GenerateOptions{Stream: true})
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}
	if result != "chunk" {
		t.Fatalf("expected chunk result, got %q", result)
	}
	if !gotStream {
		t.Fatalf("expected streaming execution")
	}
	if len(chunks) != 1 || chunks[0] != "chunk" {
		t.Fatalf("expected streamed chunk callback, got %v", chunks)
	}
}

func TestClientGenerateRejectsUnknownKind(t *testing.T) {
	client := NewClient(config.NewLoader("").LoadOrPanicForTest())

	_, err := client.Generate(context.Background(), Request{Kind: Kind("unknown")}, GenerateOptions{})
	if err == nil {
		t.Fatal("expected error for unknown generation kind")
	}
	if !strings.Contains(err.Error(), "unsupported generation kind") {
		t.Fatalf("expected unsupported generation kind error, got %v", err)
	}
}
