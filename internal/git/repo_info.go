package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// RepoContext contains information about the current git repository context
type RepoContext struct {
	BranchName string
	FilesChanged int
	ChangeSummary string
	FileChanges []FileChange
}

// FileChange represents a single changed file in the repository
type FileChange struct {
	Status    string
	FilePath  string
	FileType  string
}

// GetRepoContext returns the current repository context including branch, changes, etc.
func GetRepoContext() (*RepoContext, error) {
	ctx := &RepoContext{}
	
	// Get current branch name
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchOut, err := branchCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get branch name: %w", err)
	}
	ctx.BranchName = strings.TrimSpace(string(branchOut))

	// Get number of changed files
	countCmd := exec.Command("git", "diff", "--staged", "--name-only")
	countOut, err := countCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to count changed files: %w", err)
	}
	files := strings.Split(strings.TrimSpace(string(countOut)), "\n")
	if len(files) == 1 && files[0] == "" {
		ctx.FilesChanged = 0
	} else {
		ctx.FilesChanged = len(files)
	}

	// Get change summary
	summaryCmd := exec.Command("git", "diff", "--staged", "--stat")
	summaryOut, err := summaryCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get change summary: %w", err)
	}
	ctx.ChangeSummary = string(summaryOut)

	// Get detailed file changes
	changesCmd := exec.Command("git", "diff", "--staged", "--name-status")
	changesOut, err := changesCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get file changes: %w", err)
	}

	changes := strings.Split(strings.TrimSpace(string(changesOut)), "\n")
	for _, change := range changes {
		if change == "" {
			continue
		}
		// Split on tab to separate status and file path
		parts := strings.Fields(change)
		if len(parts) < 2 {
			continue
		}
		status := parts[0]
		filePath := parts[1]
		
		// Get file extension
		fileType := ""
		if dotIndex := strings.LastIndex(filePath, "."); dotIndex != -1 && dotIndex < len(filePath)-1 {
			fileType = filePath[dotIndex+1:]
		}

		ctx.FileChanges = append(ctx.FileChanges, FileChange{
			Status:   status,
			FilePath: filePath,
			FileType: fileType,
		})
	}

	return ctx, nil
}

// String returns a formatted string representation of the repository context
func (r *RepoContext) String() string {
	var sb strings.Builder

	sb.WriteString("REPOSITORY CONTEXT:\n")
	sb.WriteString(fmt.Sprintf("Branch: %s\n", r.BranchName))
	sb.WriteString(fmt.Sprintf("Files changed: %d\n", r.FilesChanged))

	sb.WriteString("\nCHANGE SUMMARY:\n")
	sb.WriteString(r.ChangeSummary)

	if len(r.FileChanges) > 0 {
		sb.WriteString("\nFILE CHANGES:\n")
		for _, change := range r.FileChanges {
			sb.WriteString(fmt.Sprintf("[%s] %s (%s)\n", change.Status, change.FilePath, change.FileType))
		}
	}

	return sb.String()
}
