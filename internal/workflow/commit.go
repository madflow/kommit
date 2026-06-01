package workflow

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/madflow/kommit/internal/config"
)

var prTitleMaxLengthPattern = regexp.MustCompile(`Maximum length:\s*(\d+)`)

type CommitRequest struct {
	Add               bool
	Yolo              bool
	CreatePullRequest bool
	Generation        GenerationOptions
}

type CommitWorkflow struct {
	settings   *config.ResolvedSettings
	repository CommitRepository
	generator  Generator
	prompter   CommitPrompter
	output     Output
}

func NewCommitWorkflow(settings *config.ResolvedSettings, deps CommitDependencies) *CommitWorkflow {
	settings = ensureSettings(settings)
	deps.Output = ensureOutput(deps.Output)
	if deps.Repository == nil {
		deps.Repository = gitRepository{}
	}
	if deps.Generator == nil {
		deps.Generator = providerGenerator{settings: settings, output: deps.Output}
	}

	return &CommitWorkflow{
		settings:   settings,
		repository: deps.Repository,
		generator:  deps.Generator,
		prompter:   deps.Prompter,
		output:     deps.Output,
	}
}

func (w *CommitWorkflow) Run(ctx context.Context, req CommitRequest) error {
	if err := w.repository.EnsureRepository(); err != nil {
		return fmt.Errorf("not in a git repository")
	}

	if req.Add || req.Yolo {
		if err := w.repository.StageAll(); err != nil {
			return fmt.Errorf("error staging changes: %w", err)
		}
	}

	snapshot, err := w.repository.LoadStagedSnapshot()
	if err != nil {
		return fmt.Errorf("error loading staged snapshot: %w", err)
	}

	if !snapshot.HasChanges {
		w.output.Emit(Event{Kind: EventCommitNoChanges})
		return nil
	}

	repoCtx := snapshot.Context
	w.output.Emit(Event{Kind: EventRepositoryContext, Context: repoCtx})

	diff := snapshot.Diff

	w.output.Emit(Event{Kind: EventCommitAnalyzing})

	messageText, err := w.generator.Generate(ctx, GenerationRequest{
		Kind:    GenerationKindCommitMessage,
		Diff:    diff,
		Rules:   w.settings.Rules(),
		Context: repoCtx,
	}, req.Generation)
	if err != nil {
		return fmt.Errorf("error generating commit message: %w", err)
	}
	message := strings.TrimSpace(messageText)

	w.output.Emit(Event{Kind: EventCommitMessageGenerated, Text: message})

	if req.Yolo {
		return w.yoloCommit(ctx, req, repoCtx, diff, message)
	}

	message, approved, err := w.confirmCommit(message)
	if err != nil {
		return err
	}
	if !approved {
		return nil
	}

	if err := w.repository.Commit(message); err != nil {
		return fmt.Errorf("error committing changes: %w", err)
	}

	w.output.Emit(Event{Kind: EventCommitSucceeded})
	w.createPullRequest(ctx, req, repoCtx, diff, message)

	return nil
}

func (w *CommitWorkflow) confirmCommit(message string) (string, bool, error) {
	if w.prompter == nil {
		return "", false, fmt.Errorf("interactive confirmation is unavailable")
	}

	for {
		action, err := w.prompter.AskForConfirmation()
		if err != nil {
			return "", false, fmt.Errorf("error reading confirmation: %w", err)
		}

		switch action {
		case CommitActionNo:
			w.output.Emit(Event{Kind: EventCommitCancelled})
			return message, false, nil
		case CommitActionYes:
			return message, true, nil
		case CommitActionEdit:
			editedMessage, err := w.prompter.EditCommitMessage(message)
			if err != nil {
				return "", false, fmt.Errorf("error editing commit message: %w", err)
			}
			message = editedMessage
			w.output.Emit(Event{Kind: EventCommitMessageUpdated, Text: message})
		default:
			return "", false, fmt.Errorf("unsupported confirmation action %q", action)
		}
	}
}

func (w *CommitWorkflow) yoloCommit(ctx context.Context, req CommitRequest, repoCtx *PromptContext, diff, message string) error {
	w.output.Emit(Event{Kind: EventCommitYoloStarted})

	if err := w.repository.Commit(message); err != nil {
		return fmt.Errorf("error committing changes: %w", err)
	}

	if err := w.repository.PublishCurrentBranch(); err != nil {
		return fmt.Errorf("error pushing changes: %w", err)
	}

	w.output.Emit(Event{Kind: EventCommitYoloSucceeded})
	w.createPullRequest(ctx, req, repoCtx, diff, message)

	return nil
}

func (w *CommitWorkflow) createPullRequest(ctx context.Context, req CommitRequest, repoCtx *PromptContext, localDiff, commitMessage string) {
	if !req.CreatePullRequest {
		return
	}

	preparation, err := w.repository.PreparePullRequest(localDiff)
	if err != nil {
		if errors.Is(err, ErrPullRequestMainBranchUnavailable) {
			w.output.Emit(Event{Kind: EventPullRequestMainUnavailable, Err: err})
			return
		}

		w.output.Emit(Event{Kind: EventPullRequestPreparationFailed, Err: err})
		return
	}

	switch preparation.SkipReason {
	case PullRequestSkipNotGitHub:
		w.output.Emit(Event{Kind: EventPullRequestNotGitHub})
		return
	case PullRequestSkipMainBranch:
		w.output.Emit(Event{Kind: EventPullRequestSkippedOnMain, Text: preparation.CurrentBranch})
		return
	case PullRequestSkipCliUnavailable:
		w.output.Emit(Event{Kind: EventPullRequestCliUnavailable})
		return
	case PullRequestSkipCliUnauthenticated:
		w.output.Emit(Event{Kind: EventPullRequestCliUnauthenticated})
		return
	}

	if !req.Yolo {
		w.output.Emit(Event{Kind: EventPullRequestPublishing})
		if err := w.repository.PublishCurrentBranch(); err != nil {
			w.output.Emit(Event{Kind: EventPullRequestPublishFailed, Err: err})
			return
		}
	}

	w.output.Emit(Event{Kind: EventPullRequestCreating})

	prBody, err := w.generator.Generate(ctx, GenerationRequest{
		Kind:    GenerationKindPullRequestBody,
		Diff:    preparation.CombinedDiff,
		Rules:   w.settings.PullRequestRules(),
		Context: repoCtx,
	}, req.Generation)
	if err != nil {
		w.output.Emit(Event{Kind: EventPullRequestBodyFallback, Err: err})
		prBody = ""
	}

	prTitle, err := w.generator.Generate(ctx, GenerationRequest{
		Kind:    GenerationKindPullRequestTitle,
		Diff:    preparation.CombinedDiff,
		Rules:   w.settings.PullRequestTitleRules(),
		Context: repoCtx,
	}, req.Generation)
	if err != nil {
		w.output.Emit(Event{Kind: EventPullRequestTitleFallback, Err: err})
		prTitle = pullRequestTitleFallback(commitMessage, w.settings.PullRequestTitleRules())
	}

	prURL, err := w.repository.CreatePullRequest(prTitle, prBody)
	if err != nil {
		w.output.Emit(Event{Kind: EventPullRequestCreateFailed, Err: err})
		return
	}

	w.output.Emit(Event{Kind: EventPullRequestCreated, Text: prURL})
}

func pullRequestTitleFallback(commitMessage, rules string) string {
	title, _, _ := strings.Cut(commitMessage, "\n")
	title = strings.TrimSpace(title)

	maxLength := 50
	matches := prTitleMaxLengthPattern.FindStringSubmatch(rules)
	if len(matches) == 2 {
		if parsed, err := strconv.Atoi(matches[1]); err == nil && parsed > 0 {
			maxLength = parsed
		}
	}

	if len(title) <= maxLength {
		return title
	}
	if maxLength <= 3 {
		return title[:maxLength]
	}

	return title[:maxLength-3] + "..."
}
