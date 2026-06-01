package workflow

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/madflow/kommit/internal/config"
)

var invalidBranchNameChars = regexp.MustCompile(`[^a-zA-Z0-9-_/]+`)

type BranchRequest struct {
	Generation GenerationOptions
}

type BranchWorkflow struct {
	settings   *config.ResolvedSettings
	repository BranchRepository
	generator  Generator
	output     Output
}

func NewBranchWorkflow(settings *config.ResolvedSettings, deps BranchDependencies) *BranchWorkflow {
	settings = ensureSettings(settings)
	deps.Output = ensureOutput(deps.Output)
	if deps.Repository == nil {
		deps.Repository = gitRepository{}
	}
	if deps.Generator == nil {
		deps.Generator = providerGenerator{settings: settings, output: deps.Output}
	}

	return &BranchWorkflow{
		settings:   settings,
		repository: deps.Repository,
		generator:  deps.Generator,
		output:     deps.Output,
	}
}

func (w *BranchWorkflow) Run(ctx context.Context, req BranchRequest) error {
	if err := w.repository.EnsureRepository(); err != nil {
		return fmt.Errorf("not in a git repository")
	}

	if err := w.repository.StageAll(); err != nil {
		return fmt.Errorf("error staging changes: %w", err)
	}

	snapshot, err := w.repository.LoadStagedSnapshot()
	if err != nil {
		return fmt.Errorf("error loading staged snapshot: %w", err)
	}

	if !snapshot.HasChanges {
		w.output.Emit(Event{Kind: EventBranchNoChanges})
		return nil
	}

	w.output.Emit(Event{Kind: EventBranchAnalyzing})

	branchName, err := w.generator.Generate(ctx, GenerationRequest{
		Kind: GenerationKindBranchName,
		Diff: snapshot.Diff,
	}, req.Generation)
	if err != nil {
		return fmt.Errorf("error generating branch name: %w", err)
	}

	branchName = sanitizeBranchName(branchName)
	w.output.Emit(Event{Kind: EventBranchNameGenerated, Text: branchName})

	if err := w.repository.CreateBranch(branchName); err != nil {
		return fmt.Errorf("error creating branch: %w", err)
	}

	w.output.Emit(Event{Kind: EventBranchCreated, Text: branchName})
	return nil
}

func sanitizeBranchName(branchName string) string {
	branchName = strings.TrimSpace(branchName)
	if len(branchName) > 50 {
		branchName = branchName[:50]
	}

	return invalidBranchNameChars.ReplaceAllString(branchName, "")
}
