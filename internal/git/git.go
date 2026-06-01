package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var (
	ErrNotGitRepository            = errors.New("not in a git repository")
	ErrOriginMainBranchUnavailable = errors.New("origin main branch is unavailable")
)

type PullRequestSkipReason string

const (
	PullRequestSkipNone              PullRequestSkipReason = ""
	PullRequestSkipNotGitHub         PullRequestSkipReason = "not_github"
	PullRequestSkipMainBranch        PullRequestSkipReason = "main_branch"
	PullRequestSkipGhUnavailable     PullRequestSkipReason = "gh_unavailable"
	PullRequestSkipGhUnauthenticated PullRequestSkipReason = "gh_unauthenticated"
)

type StagedSnapshot struct {
	HasChanges bool
	Diff       string
	Context    *RepoContext
}

type PullRequestPreparation struct {
	CurrentBranch string
	MainBranch    string
	CombinedDiff  string
	SkipReason    PullRequestSkipReason
}

type commandRunner interface {
	Output(name string, args ...string) ([]byte, error)
	CombinedOutput(name string, args ...string) ([]byte, error)
	Run(name string, args ...string) error
}

type execRunner struct{}

func (execRunner) Output(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

func (execRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func (execRunner) Run(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

type Repository struct {
	runner commandRunner
}

func Current() *Repository {
	return &Repository{runner: execRunner{}}
}

func GetGitDir() (string, error) {
	return Current().GitDir()
}

func (r *Repository) EnsureRepository() error {
	if err := r.runnerOrDefault().Run("git", "rev-parse", "--git-dir"); err != nil {
		return ErrNotGitRepository
	}

	return nil
}

func (r *Repository) GitDir() (string, error) {
	if err := r.EnsureRepository(); err != nil {
		return "", err
	}

	output, err := r.runnerOrDefault().Output("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) StageAll() error {
	return r.runnerOrDefault().Run("git", "add", ".")
}

func (r *Repository) LoadStagedSnapshot() (*StagedSnapshot, error) {
	if err := r.EnsureRepository(); err != nil {
		return nil, err
	}

	repoCtx, err := r.loadRepoContext()
	if err != nil {
		return nil, err
	}

	diff, err := r.stagedDiff()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged diff: %w", err)
	}

	return &StagedSnapshot{
		HasChanges: repoCtx.FilesChanged > 0,
		Diff:       diff,
		Context:    repoCtx,
	}, nil
}

func (r *Repository) Commit(message string) error {
	return r.runnerOrDefault().Run("git", "commit", "-m", message)
}

func (r *Repository) PublishCurrentBranch() error {
	branch, err := r.currentBranch()
	if err != nil {
		return err
	}

	return r.runnerOrDefault().Run("git", "push", "--set-upstream", "origin", branch)
}

func (r *Repository) CreateBranch(branchName string) error {
	return r.runnerOrDefault().Run("git", "checkout", "-b", branchName)
}

func (r *Repository) PreparePullRequest(localDiff string) (*PullRequestPreparation, error) {
	if err := r.EnsureRepository(); err != nil {
		return nil, err
	}

	isGitHubRepo, err := r.isGitHubRepo()
	if err != nil {
		return nil, fmt.Errorf("check GitHub repository: %w", err)
	}
	if !isGitHubRepo {
		return &PullRequestPreparation{SkipReason: PullRequestSkipNotGitHub}, nil
	}

	currentBranch, err := r.currentBranch()
	if err != nil {
		return nil, fmt.Errorf("get current branch: %w", err)
	}

	mainBranch, err := r.originMainBranch()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOriginMainBranchUnavailable, err)
	}

	preparation := &PullRequestPreparation{
		CurrentBranch: currentBranch,
		MainBranch:    mainBranch,
	}

	if currentBranch == mainBranch {
		preparation.SkipReason = PullRequestSkipMainBranch
		return preparation, nil
	}

	if !r.isGhCliAvailable() {
		preparation.SkipReason = PullRequestSkipGhUnavailable
		return preparation, nil
	}

	if !r.isGhAuthenticated() {
		preparation.SkipReason = PullRequestSkipGhUnauthenticated
		return preparation, nil
	}

	remoteDiff, err := r.diffFromOriginMain(mainBranch)
	if err != nil {
		return nil, fmt.Errorf("get remote diff: %w", err)
	}

	preparation.CombinedDiff = fmt.Sprintf("=== LOCAL CHANGES (staged) ===\n%s\n\n=== REMOTE CHANGES (since %s) ===\n%s", localDiff, mainBranch, remoteDiff)
	return preparation, nil
}

func (r *Repository) CreatePullRequest(title string, body string) (string, error) {
	var args []string
	if body != "" {
		args = []string{"pr", "create", "--title", title, "--body", body}
	} else {
		args = []string{"pr", "create", "--fill"}
	}

	output, err := r.runnerOrDefault().CombinedOutput("gh", args...)
	if err != nil {
		return "", fmt.Errorf("gh pr create failed: %v\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) stagedDiff() (string, error) {
	output, err := r.runnerOrDefault().Output("git", "diff", "--cached")
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (r *Repository) currentBranch() (string, error) {
	output, err := r.runnerOrDefault().Output("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) originMainBranch() (string, error) {
	output, err := r.runnerOrDefault().Output("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", err
	}

	fullRef := strings.TrimSpace(string(output))
	prefix := "refs/remotes/origin/"
	if after, ok := strings.CutPrefix(fullRef, prefix); ok {
		return after, nil
	}

	return fullRef, nil
}

func (r *Repository) diffFromOriginMain(mainBranch string) (string, error) {
	output, err := r.runnerOrDefault().Output("git", "diff", fmt.Sprintf("origin/%s...HEAD", mainBranch))
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (r *Repository) isGitHubRepo() (bool, error) {
	output, err := r.runnerOrDefault().Output("git", "remote", "get-url", "origin")
	if err != nil {
		return false, err
	}

	return strings.Contains(strings.TrimSpace(string(output)), "github.com"), nil
}

func (r *Repository) isGhCliAvailable() bool {
	return r.runnerOrDefault().Run("gh", "--version") == nil
}

func (r *Repository) isGhAuthenticated() bool {
	return r.runnerOrDefault().Run("gh", "auth", "status") == nil
}

func (r *Repository) runnerOrDefault() commandRunner {
	if r != nil && r.runner != nil {
		return r.runner
	}

	return execRunner{}
}
