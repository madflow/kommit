package git

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestRepositoryLoadStagedSnapshotWithChanges(t *testing.T) {
	runner := &fakeRunner{
		output: map[string]commandResult{
			"git rev-parse --show-toplevel":   {stdout: "/tmp/repo\n"},
			"git branch --show-current":       {stdout: "feature/refactor\n"},
			"git diff --staged --name-only":   {stdout: "cmd/root.go\ninternal/git/git.go\n"},
			"git diff --staged --stat":        {stdout: " 2 files changed, 10 insertions(+), 2 deletions(-)\n"},
			"git diff --staged --name-status": {stdout: "M\tcmd/root.go\nA\tinternal/git/git.go\n"},
			"git diff --cached":               {stdout: "diff --git a/cmd/root.go b/cmd/root.go\n"},
		},
		run: map[string]error{
			"git rev-parse --git-dir": nil,
		},
	}
	repo := &Repository{runner: runner}

	snapshot, err := repo.LoadStagedSnapshot()
	if err != nil {
		t.Fatalf("LoadStagedSnapshot() returned error: %v", err)
	}
	if !snapshot.HasChanges {
		t.Fatal("expected snapshot to report staged changes")
	}
	if snapshot.Diff == "" {
		t.Fatal("expected staged diff to be populated")
	}
	if snapshot.Context == nil {
		t.Fatal("expected repository context")
	}
	if snapshot.Context.BranchName != "feature/refactor" {
		t.Fatalf("expected branch feature/refactor, got %q", snapshot.Context.BranchName)
	}
	if snapshot.Context.FilesChanged != 2 {
		t.Fatalf("expected 2 changed files, got %d", snapshot.Context.FilesChanged)
	}
	if len(snapshot.Context.FileChanges) != 2 {
		t.Fatalf("expected 2 file changes, got %d", len(snapshot.Context.FileChanges))
	}
}

func TestRepositoryLoadStagedSnapshotWithoutChanges(t *testing.T) {
	runner := &fakeRunner{
		output: map[string]commandResult{
			"git branch --show-current":       {stdout: "main\n"},
			"git diff --staged --name-only":   {stdout: ""},
			"git diff --staged --stat":        {stdout: ""},
			"git diff --staged --name-status": {stdout: ""},
			"git diff --cached":               {stdout: ""},
		},
		run: map[string]error{
			"git rev-parse --git-dir": nil,
		},
	}
	repo := &Repository{runner: runner}

	snapshot, err := repo.LoadStagedSnapshot()
	if err != nil {
		t.Fatalf("LoadStagedSnapshot() returned error: %v", err)
	}
	if snapshot.HasChanges {
		t.Fatal("expected snapshot without staged changes")
	}
	if snapshot.Context.FilesChanged != 0 {
		t.Fatalf("expected 0 changed files, got %d", snapshot.Context.FilesChanged)
	}
}

func TestRepositoryPreparePullRequestBuildsCombinedDiff(t *testing.T) {
	runner := &fakeRunner{
		output: map[string]commandResult{
			"git remote get-url origin":                 {stdout: "git@github.com:madflow/kommit.git\n"},
			"git rev-parse --abbrev-ref HEAD":           {stdout: "feature/refactor\n"},
			"git symbolic-ref refs/remotes/origin/HEAD": {stdout: "refs/remotes/origin/main\n"},
			"git diff origin/main...HEAD":               {stdout: "remote diff\n"},
		},
		run: map[string]error{
			"git rev-parse --git-dir": nil,
			"gh --version":            nil,
			"gh auth status":          nil,
		},
	}
	repo := &Repository{runner: runner}

	preparation, err := repo.PreparePullRequest("local diff")
	if err != nil {
		t.Fatalf("PreparePullRequest() returned error: %v", err)
	}
	if preparation.SkipReason != PullRequestSkipNone {
		t.Fatalf("expected no skip reason, got %q", preparation.SkipReason)
	}
	if preparation.CurrentBranch != "feature/refactor" {
		t.Fatalf("expected current branch feature/refactor, got %q", preparation.CurrentBranch)
	}
	if preparation.MainBranch != "main" {
		t.Fatalf("expected main branch main, got %q", preparation.MainBranch)
	}
	if !strings.Contains(preparation.CombinedDiff, "=== LOCAL CHANGES (staged) ===") {
		t.Fatalf("expected combined diff header, got %q", preparation.CombinedDiff)
	}
	if !strings.Contains(preparation.CombinedDiff, "remote diff") {
		t.Fatalf("expected remote diff, got %q", preparation.CombinedDiff)
	}
}

func TestRepositoryPreparePullRequestSkipReasons(t *testing.T) {
	tests := []struct {
		name      string
		runner    *fakeRunner
		expected  PullRequestSkipReason
		expectErr error
	}{
		{
			name: "not github repo",
			runner: &fakeRunner{
				output: map[string]commandResult{
					"git remote get-url origin": {stdout: "git@gitlab.com:madflow/kommit.git\n"},
				},
				run: map[string]error{"git rev-parse --git-dir": nil},
			},
			expected: PullRequestSkipNotGitHub,
		},
		{
			name: "main branch",
			runner: &fakeRunner{
				output: map[string]commandResult{
					"git remote get-url origin":                 {stdout: "git@github.com:madflow/kommit.git\n"},
					"git rev-parse --abbrev-ref HEAD":           {stdout: "main\n"},
					"git symbolic-ref refs/remotes/origin/HEAD": {stdout: "refs/remotes/origin/main\n"},
				},
				run: map[string]error{"git rev-parse --git-dir": nil},
			},
			expected: PullRequestSkipMainBranch,
		},
		{
			name: "gh unavailable",
			runner: &fakeRunner{
				output: map[string]commandResult{
					"git remote get-url origin":                 {stdout: "git@github.com:madflow/kommit.git\n"},
					"git rev-parse --abbrev-ref HEAD":           {stdout: "feature/refactor\n"},
					"git symbolic-ref refs/remotes/origin/HEAD": {stdout: "refs/remotes/origin/main\n"},
				},
				run: map[string]error{
					"git rev-parse --git-dir": nil,
					"gh --version":            errors.New("missing gh"),
				},
			},
			expected: PullRequestSkipGhUnavailable,
		},
		{
			name: "gh unauthenticated",
			runner: &fakeRunner{
				output: map[string]commandResult{
					"git remote get-url origin":                 {stdout: "git@github.com:madflow/kommit.git\n"},
					"git rev-parse --abbrev-ref HEAD":           {stdout: "feature/refactor\n"},
					"git symbolic-ref refs/remotes/origin/HEAD": {stdout: "refs/remotes/origin/main\n"},
				},
				run: map[string]error{
					"git rev-parse --git-dir": nil,
					"gh --version":            nil,
					"gh auth status":          errors.New("not authenticated"),
				},
			},
			expected: PullRequestSkipGhUnauthenticated,
		},
		{
			name: "origin main unavailable",
			runner: &fakeRunner{
				output: map[string]commandResult{
					"git remote get-url origin":                 {stdout: "git@github.com:madflow/kommit.git\n"},
					"git rev-parse --abbrev-ref HEAD":           {stdout: "feature/refactor\n"},
					"git symbolic-ref refs/remotes/origin/HEAD": {err: errors.New("origin/HEAD missing")},
				},
				run: map[string]error{"git rev-parse --git-dir": nil},
			},
			expectErr: ErrOriginMainBranchUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &Repository{runner: tt.runner}
			preparation, err := repo.PreparePullRequest("local diff")
			if tt.expectErr != nil {
				if !errors.Is(err, tt.expectErr) {
					t.Fatalf("expected error %v, got %v", tt.expectErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("PreparePullRequest() returned error: %v", err)
			}
			if preparation.SkipReason != tt.expected {
				t.Fatalf("expected skip reason %q, got %q", tt.expected, preparation.SkipReason)
			}
		})
	}
}

func TestRepositoryPublishCurrentBranch(t *testing.T) {
	runner := &fakeRunner{
		output: map[string]commandResult{
			"git rev-parse --abbrev-ref HEAD": {stdout: "feature/refactor\n"},
		},
		run: map[string]error{
			"git push --set-upstream origin feature/refactor": nil,
		},
	}

	repo := &Repository{runner: runner}
	if err := repo.PublishCurrentBranch(); err != nil {
		t.Fatalf("PublishCurrentBranch() returned error: %v", err)
	}
	if len(runner.runCalls) != 1 || runner.runCalls[0] != "git push --set-upstream origin feature/refactor" {
		t.Fatalf("expected push command, got %v", runner.runCalls)
	}
}

type commandResult struct {
	stdout string
	err    error
}

type fakeRunner struct {
	output           map[string]commandResult
	combinedOutput   map[string]commandResult
	run              map[string]error
	runCalls         []string
	combinedOutCalls []string
}

func (r *fakeRunner) Output(name string, args ...string) ([]byte, error) {
	key := commandKey(name, args...)
	result, ok := r.output[key]
	if !ok {
		return nil, fmt.Errorf("unexpected output command %s", key)
	}
	return []byte(result.stdout), result.err
}

func (r *fakeRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	key := commandKey(name, args...)
	result, ok := r.combinedOutput[key]
	if !ok {
		return nil, fmt.Errorf("unexpected combined output command %s", key)
	}
	r.combinedOutCalls = append(r.combinedOutCalls, key)
	return []byte(result.stdout), result.err
}

func (r *fakeRunner) Run(name string, args ...string) error {
	key := commandKey(name, args...)
	r.runCalls = append(r.runCalls, key)
	if r.run == nil {
		return nil
	}
	if err, ok := r.run[key]; ok {
		return err
	}
	return fmt.Errorf("unexpected run command %s", key)
}

func commandKey(name string, args ...string) string {
	return strings.TrimSpace(name + " " + strings.Join(args, " "))
}
