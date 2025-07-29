package cmd

import (
	"fmt"
	"strings"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/git"
	"github.com/madflow/kommit/internal/logger"
	"github.com/madflow/kommit/internal/ollama"
	"github.com/spf13/cobra"
)

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Generate a branch name based on the current changes",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Println("🤖 Kommit Branch")
		logger.Println("====================================")

		if !git.IsGitRepo() {
			logger.Fatal("Not in a git repository")
		}

		if err := git.AddAll(); err != nil {
			logger.Fatal("Error staging changes: %v", err)
		}

		hasChanges, err := git.HasStagedChanges()
		if err != nil {
			logger.Fatal("Error checking for changes: %v", err)
		}

		if !hasChanges {
			logger.Success("No changes to create a branch for")
			return
		}

		diff, err := git.GetGitDiff()
		if err != nil {
			logger.Fatal("Error getting git diff: %v", err)
		}

		logger.Info("Analyzing changes to generate a branch name...")

		cfg := config.Get()
		ollamaClient := ollama.NewClient(&cfg.Ollama)
		branchName, err := ollamaClient.GenerateBranchName(diff)
		if err != nil {
			logger.Fatal("Error generating branch name: %v", err)
		}

		branchName = strings.TrimSpace(branchName)
		if len(branchName) > 50 {
			branchName = branchName[:50]
		}

		logger.Info("Generated branch name: %s", branchName)

		if err := git.CreateBranch(branchName); err != nil {
			logger.Fatal("Error creating branch: %v", err)
		}

		logger.Success(fmt.Sprintf("Successfully created branch '%s'", branchName))
	},
}

func init() {
	rootCmd.AddCommand(branchCmd)
}
