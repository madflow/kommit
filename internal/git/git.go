package git

import (
	"os/exec"
	"strings"
)

// execCommand is defined as a variable so it can be mocked in tests
var execCommand = exec.Command

func IsGitRepo() bool {
	cmd := execCommand("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func GetGitStatus() (string, error) {
	cmd := exec.Command("git", "status", "-v")
	output, err := cmd.Output()
	return string(output), err
}

// GetGitDiff returns the diff of changes that are currently staged for commit.
// It only shows changes that have been added to the staging area with 'git add'.
func GetGitDiff() (string, error) {
	cmd := execCommand("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// HasStagedChanges checks if there are any staged changes in the git repository.
// It returns true if there are staged changes, false otherwise.
// If there is an error running the git command, it returns false and the error.
func HasStagedChanges() (bool, error) {
	// Use execCommand to allow mocking in tests
	cmd := execCommand("git", "diff-index", "--cached", "HEAD", "--")
	output, err := cmd.Output()
	if err != nil {
		// If HEAD doesn't exist yet (new repository), check if there are any files in the index
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 128 {
			// Try to list files in the index directly
			cmd = execCommand("git", "ls-files", "--cached", "--error-unmatch", ".")
			_, err := cmd.Output()
			if err == nil {
				return true, nil // Files are staged but no HEAD yet
			}
			return false, nil // No files in index
		}
		return false, err
	}

	// If there are staged changes, there will be output lines
	return strings.TrimSpace(string(output)) != "", nil
}

func StageAllChanges() error {
	cmd := exec.Command("git", "add", ".")
	return cmd.Run()
}

func CommitChanges(message string) error {
	cmd := execCommand("git", "commit", "-m", message)
	return cmd.Run()
}

// GetGitDir returns the absolute path to the root directory of the current git repository.
// Returns an empty string if not in a git repository.
func GetGitDir() (string, error) {
	// First try to get the git directory to check if we're in a git repo
	cmd := execCommand("git", "rev-parse", "--absolute-git-dir")
	_, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Now get the root directory of the repository
	cmd = execCommand("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// AddAll stages all changes in the working directory for commit.
func AddAll() error {
	cmd := execCommand("git", "add", ".")
	return cmd.Run()
}

// HasAnyChanges checks if there are any changes in the working directory (staged or unstaged).
func HasAnyChanges() (bool, error) {
	// Check for any changes in the working tree (unstaged changes)
	cmd := execCommand("git", "diff", "--quiet")
	unstagedChanges := cmd.Run() != nil

	// Check for any staged changes
	cmd = execCommand("git", "diff", "--cached", "--quiet")
	stagedChanges := cmd.Run() != nil

	return unstagedChanges || stagedChanges, nil
}

// PushCurrentBranch pushes the current branch to its remote tracking branch.
func PushCurrentBranch() error {
	// First, get the current branch name
	cmd := execCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	branch := strings.TrimSpace(string(output))

	// Push the current branch to its upstream branch
	pushCmd := execCommand("git", "push", "--set-upstream", "origin", branch)
	return pushCmd.Run()
}
