package git

import (
	"os/exec"
	"strings"
)

func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func HasChangesToCommit() (bool, error) {
	// Check for staged changes
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	if err := cmd.Run(); err == nil {
		// No staged changes, check for unstaged changes
		cmd = exec.Command("git", "diff", "--quiet")
		if err := cmd.Run(); err == nil {
			// Check for untracked files
			cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
			output, err := cmd.Output()
			if err != nil {
				return false, err
			}
			return strings.TrimSpace(string(output)) != "", nil
		}
		return true, nil // Has unstaged changes
	}
	return true, nil // Has staged changes
}

func GetGitStatus() (string, error) {
	cmd := exec.Command("git", "status", "-v")
	output, err := cmd.Output()
	return string(output), err
}

func GetGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// If no staged changes, get unstaged diff
	if strings.TrimSpace(string(output)) == "" {
		cmd = exec.Command("git", "diff")
		output, err = cmd.Output()
	}

	return string(output), err
}

func StageAllChanges() error {
	cmd := exec.Command("git", "add", ".")
	return cmd.Run()
}

func CommitChanges(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}
