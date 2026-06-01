package workflow

import (
	"context"
	"testing"

	"github.com/madflow/kommit/internal/config"
)

func TestCommitWorkflowRunInteractiveEdit(t *testing.T) {
	repo := &fakeCommitRepository{
		stagedSnapshot: &ChangeSet{
			HasChanges: true,
			Context: &PromptContext{
				BranchName:   "feature/refactor",
				FilesChanged: 1,
			},
			Diff: "diff --git a/root.go b/root.go",
		},
	}
	generator := &fakeCommitGenerator{commitMessage: "initial message"}
	prompter := &fakeCommitPrompter{
		actions: []CommitAction{CommitActionEdit, CommitActionYes},
		edits:   []string{"edited message"},
	}
	output := &fakeOutput{}

	workflow := NewCommitWorkflow(config.NewLoader("").LoadOrPanicForTest(), CommitDependencies{
		Repository: repo,
		Generator:  generator,
		Prompter:   prompter,
		Output:     output,
	})

	err := workflow.Run(context.Background(), CommitRequest{})
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if repo.commitCalls != 1 {
		t.Fatalf("expected 1 commit, got %d", repo.commitCalls)
	}
	if repo.commitMessage != "edited message" {
		t.Fatalf("expected edited message to be committed, got %q", repo.commitMessage)
	}
	if repo.pushCalls != 0 {
		t.Fatalf("expected no push, got %d", repo.pushCalls)
	}
	if prompter.askCalls != 2 {
		t.Fatalf("expected 2 confirmation prompts, got %d", prompter.askCalls)
	}
	if prompter.editCalls != 1 {
		t.Fatalf("expected 1 edit, got %d", prompter.editCalls)
	}
	if generator.commitCalls != 1 {
		t.Fatalf("expected 1 commit message generation, got %d", generator.commitCalls)
	}

	assertEventKinds(t, output.events, []EventKind{
		EventRepositoryContext,
		EventCommitAnalyzing,
		EventCommitMessageGenerated,
		EventCommitMessageUpdated,
		EventCommitSucceeded,
	})
	if output.events[2].Text != "initial message" {
		t.Fatalf("expected generated message event, got %q", output.events[2].Text)
	}
	if output.events[3].Text != "edited message" {
		t.Fatalf("expected updated message event, got %q", output.events[3].Text)
	}
}

func TestCommitWorkflowRunYoloCreatesPullRequestOnFeatureBranch(t *testing.T) {
	repo := &fakeCommitRepository{
		stagedSnapshot: &ChangeSet{
			HasChanges: true,
			Context:    &PromptContext{BranchName: "feature/refactor", FilesChanged: 2},
			Diff:       "local diff",
		},
		pullRequestPreparation: &PullRequestPlan{
			CurrentBranch: "feature/refactor",
			MainBranch:    "main",
			CombinedDiff:  "combined diff",
		},
		pullRequestURL: "https://example.com/pr/1",
	}
	generator := &fakeCommitGenerator{
		commitMessage:    "ship workflow module",
		pullRequestBody:  "## Summary\n\n- deepen commit workflow",
		pullRequestTitle: "Deepen commit workflow",
	}
	output := &fakeOutput{}

	workflow := NewCommitWorkflow(config.NewLoader("").LoadOrPanicForTest(), CommitDependencies{
		Repository: repo,
		Generator:  generator,
		Output:     output,
	})

	err := workflow.Run(context.Background(), CommitRequest{
		Yolo:              true,
		CreatePullRequest: true,
	})
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if repo.addAllCalls != 1 {
		t.Fatalf("expected 1 add call in YOLO mode, got %d", repo.addAllCalls)
	}
	if repo.commitCalls != 1 {
		t.Fatalf("expected 1 commit, got %d", repo.commitCalls)
	}
	if repo.pushCalls != 1 {
		t.Fatalf("expected 1 push in YOLO mode, got %d", repo.pushCalls)
	}
	if repo.createPullRequestCalls != 1 {
		t.Fatalf("expected 1 pull request creation, got %d", repo.createPullRequestCalls)
	}
	if repo.pullRequestTitle != "Deepen commit workflow" {
		t.Fatalf("expected PR title %q, got %q", "Deepen commit workflow", repo.pullRequestTitle)
	}
	if repo.pullRequestBody != "## Summary\n\n- deepen commit workflow" {
		t.Fatalf("expected PR body to be preserved, got %q", repo.pullRequestBody)
	}
	if generator.pullRequestBodyCalls != 1 {
		t.Fatalf("expected PR body generation once, got %d", generator.pullRequestBodyCalls)
	}
	if generator.pullRequestTitleCalls != 1 {
		t.Fatalf("expected PR title generation once, got %d", generator.pullRequestTitleCalls)
	}

	assertEventKinds(t, output.events, []EventKind{
		EventRepositoryContext,
		EventCommitAnalyzing,
		EventCommitMessageGenerated,
		EventCommitYoloStarted,
		EventCommitYoloSucceeded,
		EventPullRequestCreating,
		EventPullRequestCreated,
	})
	if output.events[6].Text != "https://example.com/pr/1" {
		t.Fatalf("expected PR URL event, got %q", output.events[6].Text)
	}
}

func TestCommitWorkflowRunSkipsPullRequestOnMainBranch(t *testing.T) {
	repo := &fakeCommitRepository{
		stagedSnapshot: &ChangeSet{
			HasChanges: true,
			Context:    &PromptContext{BranchName: "main", FilesChanged: 1},
			Diff:       "diff",
		},
		pullRequestPreparation: &PullRequestPlan{
			CurrentBranch: "main",
			MainBranch:    "main",
			SkipReason:    PullRequestSkipMainBranch,
		},
	}
	generator := &fakeCommitGenerator{commitMessage: "commit message"}
	prompter := &fakeCommitPrompter{actions: []CommitAction{CommitActionYes}}
	output := &fakeOutput{}

	workflow := NewCommitWorkflow(config.NewLoader("").LoadOrPanicForTest(), CommitDependencies{
		Repository: repo,
		Generator:  generator,
		Prompter:   prompter,
		Output:     output,
	})

	err := workflow.Run(context.Background(), CommitRequest{CreatePullRequest: true})
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if repo.commitCalls != 1 {
		t.Fatalf("expected 1 commit, got %d", repo.commitCalls)
	}
	if repo.pushCalls != 0 {
		t.Fatalf("expected no push when pull request is skipped on main, got %d", repo.pushCalls)
	}
	if repo.createPullRequestCalls != 0 {
		t.Fatalf("expected no pull request on main, got %d", repo.createPullRequestCalls)
	}

	assertEventKinds(t, output.events, []EventKind{
		EventRepositoryContext,
		EventCommitAnalyzing,
		EventCommitMessageGenerated,
		EventCommitSucceeded,
		EventPullRequestSkippedOnMain,
	})
	if output.events[4].Text != "main" {
		t.Fatalf("expected main branch skip event, got %q", output.events[4].Text)
	}
}

type fakeCommitRepository struct {
	ensureRepositoryErr    error
	stagedSnapshot         *ChangeSet
	addAllCalls            int
	commitCalls            int
	commitMessage          string
	pushCalls              int
	pullRequestPreparation *PullRequestPlan
	pullRequestError       error
	createPullRequestCalls int
	pullRequestTitle       string
	pullRequestBody        string
	pullRequestURL         string
}

func (r *fakeCommitRepository) EnsureRepository() error {
	return r.ensureRepositoryErr
}

func (r *fakeCommitRepository) StageAll() error {
	r.addAllCalls++
	return nil
}

func (r *fakeCommitRepository) LoadStagedSnapshot() (*ChangeSet, error) {
	if r.stagedSnapshot == nil {
		return &ChangeSet{Context: &PromptContext{}}, nil
	}
	return r.stagedSnapshot, nil
}

func (r *fakeCommitRepository) Commit(message string) error {
	r.commitCalls++
	r.commitMessage = message
	return nil
}

func (r *fakeCommitRepository) PublishCurrentBranch() error {
	r.pushCalls++
	return nil
}

func (r *fakeCommitRepository) PreparePullRequest(localDiff string) (*PullRequestPlan, error) {
	if r.pullRequestError != nil {
		return nil, r.pullRequestError
	}
	if r.pullRequestPreparation == nil {
		return &PullRequestPlan{SkipReason: PullRequestSkipNotGitHub}, nil
	}
	return r.pullRequestPreparation, nil
}

func (r *fakeCommitRepository) CreatePullRequest(title string, body string) (string, error) {
	r.createPullRequestCalls++
	r.pullRequestTitle = title
	r.pullRequestBody = body
	if r.pullRequestURL == "" {
		return "https://example.com/pr/1", nil
	}
	return r.pullRequestURL, nil
}

type fakeCommitGenerator struct {
	commitMessage         string
	pullRequestBody       string
	pullRequestTitle      string
	commitCalls           int
	pullRequestBodyCalls  int
	pullRequestTitleCalls int
}

func (g *fakeCommitGenerator) Generate(ctx context.Context, req GenerationRequest, opts GenerationOptions) (string, error) {
	switch req.Kind {
	case GenerationKindCommitMessage:
		g.commitCalls++
		if req.Context == nil {
			return "", nil
		}
		return g.commitMessage, nil
	case GenerationKindPullRequestBody:
		g.pullRequestBodyCalls++
		return g.pullRequestBody, nil
	case GenerationKindPullRequestTitle:
		g.pullRequestTitleCalls++
		return g.pullRequestTitle, nil
	default:
		return "", nil
	}
}

type fakeCommitPrompter struct {
	actions   []CommitAction
	edits     []string
	askCalls  int
	editCalls int
}

func (p *fakeCommitPrompter) AskForConfirmation() (CommitAction, error) {
	if p.askCalls >= len(p.actions) {
		return CommitActionNo, nil
	}

	action := p.actions[p.askCalls]
	p.askCalls++
	return action, nil
}

func (p *fakeCommitPrompter) EditCommitMessage(message string) (string, error) {
	if p.editCalls >= len(p.edits) {
		return message, nil
	}

	edited := p.edits[p.editCalls]
	p.editCalls++
	return edited, nil
}
