package cmd

import (
	"github.com/madflow/kommit/internal/logger"
	"github.com/madflow/kommit/internal/workflow"
)

type cliWorkflowOutput struct{}

func (cliWorkflowOutput) Emit(event workflow.Event) {
	switch event.Kind {
	case workflow.EventCommitNoChanges:
		logger.Success("No changes to commit")
	case workflow.EventRepositoryContext:
		renderRepositoryContext(event.Context)
	case workflow.EventCommitAnalyzing:
		logger.Info("Analyzing changes...")
	case workflow.EventCommitMessageGenerated:
		logger.Println("\n📝 Generated Commit Message:")
		logger.Printf("%s\n\n", event.Text)
	case workflow.EventCommitMessageUpdated:
		logger.Println("\n📝 Updated Commit Message:")
		logger.Printf("%s\n\n", event.Text)
	case workflow.EventCommitCancelled:
		logger.Error("Commit cancelled by user")
	case workflow.EventCommitSucceeded:
		logger.Success("Changes committed successfully!")
	case workflow.EventCommitYoloStarted:
		logger.Info("🚀 YOLO mode enabled - Automatically committing and pushing changes")
	case workflow.EventCommitYoloSucceeded:
		logger.Success("Changes committed and pushed successfully!")
	case workflow.EventPullRequestMainUnavailable:
		logger.Error("Error detecting origin main branch: %v", event.Err)
		logger.Info("Make sure remote tracking is set up: git remote set-head origin -a")
	case workflow.EventPullRequestPreparationFailed:
		logger.Error("Error preparing pull request: %v", event.Err)
	case workflow.EventPullRequestNotGitHub:
		logger.Error("This repository is not hosted on GitHub")
	case workflow.EventPullRequestSkippedOnMain:
		logger.Info("Current branch (%s) is the main branch. --pr flag has no effect.", event.Text)
	case workflow.EventPullRequestCliUnavailable:
		logger.Error("GitHub CLI (gh) is not installed or not available in PATH")
		logger.Info("Install it from: https://cli.github.com/")
	case workflow.EventPullRequestCliUnauthenticated:
		logger.Error("GitHub CLI is not authenticated")
		logger.Info("Run 'gh auth login' to authenticate")
	case workflow.EventPullRequestPublishing:
		logger.Info("Pushing current branch to remote...")
	case workflow.EventPullRequestPublishFailed:
		logger.Error("Error pushing branch: %v", event.Err)
	case workflow.EventPullRequestCreating:
		logger.Info("Creating pull request...")
	case workflow.EventPullRequestBodyFallback:
		logger.Error("Error generating PR body: %v", event.Err)
		logger.Info("Creating PR with empty body...")
	case workflow.EventPullRequestTitleFallback:
		logger.Error("Error generating PR title: %v", event.Err)
		logger.Info("Using commit message as title...")
	case workflow.EventPullRequestCreateFailed:
		logger.Error("Error creating pull request: %v", event.Err)
	case workflow.EventPullRequestCreated:
		logger.Success("Pull request created successfully!")
		logger.Info("URL: %s", event.Text)
	case workflow.EventBranchNoChanges:
		logger.Success("No changes to create a branch for")
	case workflow.EventBranchAnalyzing:
		logger.Info("Analyzing changes to generate a branch name...")
	case workflow.EventBranchNameGenerated:
		logger.Info("Generated branch name: %s", event.Text)
	case workflow.EventBranchCreated:
		logger.Success("Successfully created branch '%s'", event.Text)
	case workflow.EventProviderFailed:
		logger.Error("Provider '%s' failed: %v", event.Text, event.Err)
	}
}

func renderRepositoryContext(repoCtx *workflow.PromptContext) {
	if repoCtx == nil {
		return
	}

	logger.Println("📊 Repository Context:")
	logger.Printf("Branch name: %s\n", repoCtx.BranchName)
	logger.Printf("Files changed: %d\n", repoCtx.FilesChanged)

	if repoCtx.FilesChanged > 0 {
		logger.Println("\n📝 Change Summary:")
		logger.Println(repoCtx.ChangeSummary)

		if len(repoCtx.FileChanges) > 0 {
			logger.Println("\n📋 File Changes:")
			for _, change := range repoCtx.FileChanges {
				logger.Printf("[%s] %s (%s)\n", change.Status, change.Path, change.Type)
			}
		}
	}

	logger.Println()
}
