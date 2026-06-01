package cmd

import (
	"context"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/logger"
	"github.com/madflow/kommit/internal/workflow"
	"github.com/spf13/cobra"
)

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Generate a branch name based on the current changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Println("🤖 Kommit Branch")
		logger.Println("====================================")

		branchWorkflow := workflow.NewBranchWorkflow(config.Get(), workflow.BranchDependencies{Output: cliWorkflowOutput{}})

		return branchWorkflow.Run(context.Background(), workflow.BranchRequest{
			Generation: workflow.GenerationOptions{
				Provider: flagProvider,
				Model:    flagModel,
				Stream:   flagStream,
			},
		})
	},
}

func init() {
	rootCmd.AddCommand(branchCmd)
}
