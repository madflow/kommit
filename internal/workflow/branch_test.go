package workflow

import (
	"context"
	"testing"

	"github.com/madflow/kommit/internal/config"
)

func TestSanitizeBranchName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid characters",
			input:    "feat/test-branch_123",
			expected: "feat/test-branch_123",
		},
		{
			name:     "invalid characters",
			input:    "feat/test branch with spaces and invalid chars!@#$",
			expected: "feat/testbranchwithspacesandinvalidchars",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only invalid characters",
			input:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "branch name too long",
			input:    "feat/this-is-a-very-long-branch-name-that-should-be-truncated",
			expected: "feat/this-is-a-very-long-branch-name-that-should-b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := sanitizeBranchName(tt.input)
			if actual != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestBranchWorkflowRunCreatesSanitizedBranch(t *testing.T) {
	repo := &fakeBranchRepository{
		stagedSnapshot: &ChangeSet{
			HasChanges: true,
			Diff:       "diff",
		},
	}
	generator := &fakeBranchGenerator{branchName: "feat/test branch with spaces and invalid chars!@#$"}
	output := &fakeOutput{}

	workflow := NewBranchWorkflow(config.NewLoader("").LoadOrPanicForTest(), BranchDependencies{
		Repository: repo,
		Generator:  generator,
		Output:     output,
	})

	err := workflow.Run(context.Background(), BranchRequest{})
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if repo.addAllCalls != 1 {
		t.Fatalf("expected 1 add call, got %d", repo.addAllCalls)
	}
	if generator.calls != 1 {
		t.Fatalf("expected 1 branch generation, got %d", generator.calls)
	}
	if repo.createBranchCalls != 1 {
		t.Fatalf("expected 1 branch creation, got %d", repo.createBranchCalls)
	}
	if repo.createdBranch != "feat/testbranchwithspacesandinvalidchars" {
		t.Fatalf("expected sanitized branch name, got %q", repo.createdBranch)
	}

	assertEventKinds(t, output.events, []EventKind{
		EventBranchAnalyzing,
		EventBranchNameGenerated,
		EventBranchCreated,
	})
	if output.events[1].Text != "feat/testbranchwithspacesandinvalidchars" {
		t.Fatalf("expected generated branch name event, got %q", output.events[1].Text)
	}
}

func TestBranchWorkflowRunNoChanges(t *testing.T) {
	repo := &fakeBranchRepository{}
	generator := &fakeBranchGenerator{}
	output := &fakeOutput{}

	workflow := NewBranchWorkflow(config.NewLoader("").LoadOrPanicForTest(), BranchDependencies{
		Repository: repo,
		Generator:  generator,
		Output:     output,
	})

	err := workflow.Run(context.Background(), BranchRequest{})
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if generator.calls != 0 {
		t.Fatalf("expected no generation when there are no changes, got %d", generator.calls)
	}
	if repo.createBranchCalls != 0 {
		t.Fatalf("expected no branch creation when there are no changes, got %d", repo.createBranchCalls)
	}

	assertEventKinds(t, output.events, []EventKind{EventBranchNoChanges})
}

type fakeBranchRepository struct {
	ensureRepositoryErr error
	stagedSnapshot      *ChangeSet
	addAllCalls         int
	createBranchCalls   int
	createdBranch       string
}

func (r *fakeBranchRepository) EnsureRepository() error {
	return r.ensureRepositoryErr
}

func (r *fakeBranchRepository) StageAll() error {
	r.addAllCalls++
	return nil
}

func (r *fakeBranchRepository) LoadStagedSnapshot() (*ChangeSet, error) {
	if r.stagedSnapshot == nil {
		return &ChangeSet{}, nil
	}
	return r.stagedSnapshot, nil
}

func (r *fakeBranchRepository) CreateBranch(branchName string) error {
	r.createBranchCalls++
	r.createdBranch = branchName
	return nil
}

type fakeBranchGenerator struct {
	branchName string
	calls      int
}

func (g *fakeBranchGenerator) Generate(ctx context.Context, req GenerationRequest, opts GenerationOptions) (string, error) {
	g.calls++
	return g.branchName, nil
}
