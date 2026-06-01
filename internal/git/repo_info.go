package git

import (
	"fmt"
	"strings"
)

// RepoContext contains information about the current git repository context
type RepoContext struct {
	BranchName    string
	FilesChanged  int
	ChangeSummary string
	FileChanges   []FileChange
}

// FileChange represents a single changed file in the repository
type FileChange struct {
	Status   string
	FilePath string
	FileType string
}

func (r *Repository) loadRepoContext() (*RepoContext, error) {
	ctx := &RepoContext{}

	branchOut, err := r.runnerOrDefault().Output("git", "branch", "--show-current")
	if err != nil {
		return nil, fmt.Errorf("failed to get branch name: %w", err)
	}
	ctx.BranchName = strings.TrimSpace(string(branchOut))

	countOut, err := r.runnerOrDefault().Output("git", "diff", "--staged", "--name-only")
	if err != nil {
		return nil, fmt.Errorf("failed to count changed files: %w", err)
	}
	files := strings.Split(strings.TrimSpace(string(countOut)), "\n")
	if len(files) == 1 && files[0] == "" {
		ctx.FilesChanged = 0
	} else {
		ctx.FilesChanged = len(files)
	}

	summaryOut, err := r.runnerOrDefault().Output("git", "diff", "--staged", "--stat")
	if err != nil {
		return nil, fmt.Errorf("failed to get change summary: %w", err)
	}
	ctx.ChangeSummary = string(summaryOut)

	changesOut, err := r.runnerOrDefault().Output("git", "diff", "--staged", "--name-status")
	if err != nil {
		return nil, fmt.Errorf("failed to get file changes: %w", err)
	}

	for _, change := range strings.Split(strings.TrimSpace(string(changesOut)), "\n") {
		if change == "" {
			continue
		}
		parts := strings.Fields(change)
		if len(parts) < 2 {
			continue
		}
		status := parts[0]
		filePath := parts[1]

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
	fmt.Fprintf(&sb, "Branch: %s\n", r.BranchName)
	fmt.Fprintf(&sb, "Files changed: %d\n", r.FilesChanged)

	sb.WriteString("\nCHANGE SUMMARY:\n")
	sb.WriteString(r.ChangeSummary)

	if len(r.FileChanges) > 0 {
		sb.WriteString("\nFILE CHANGES:\n")
		for _, change := range r.FileChanges {
			fmt.Fprintf(&sb, "[%s] %s (%s)\n", change.Status, change.FilePath, change.FileType)
		}
	}

	return sb.String()
}
