package cmd

import (
	"context"
	"fmt"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/logger"
	"github.com/madflow/kommit/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	yolo    bool
	add     bool
	pr      bool

	flagProvider string
	flagModel    string
	flagStream   bool

	configInit = config.Init
	configGet  = config.Get
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kommit",
	Short: "Git commits for the rest of us",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Println("🤖 Kommit")
		logger.Println("================================")

		commitWorkflow := workflow.NewCommitWorkflow(configGet(), workflow.CommitDependencies{
			Prompter: cliCommitPrompter{},
			Output:   cliWorkflowOutput{},
		})

		return commitWorkflow.Run(context.Background(), workflow.CommitRequest{
			Add:               add,
			Yolo:              yolo,
			CreatePullRequest: pr,
			Generation: workflow.GenerationOptions{
				Provider: flagProvider,
				Model:    flagModel,
				Stream:   flagStream,
			},
		})
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		exitWithError("Command failed: %v", err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/kommit/config.yaml or $HOME/.config/kommit/config.yaml)")
	rootCmd.Flags().BoolVarP(&yolo, "yolo", "y", false, "Automatically stage all changes, commit, and push without confirmation")
	rootCmd.Flags().BoolVarP(&add, "add", "a", false, "Stage all changes before committing")
	rootCmd.Flags().BoolVar(&pr, "pr", false, "Create pull request after committing")
	rootCmd.PersistentFlags().StringVar(&flagProvider, "provider", "", "Specify provider (overrides config/default)")
	rootCmd.PersistentFlags().StringVar(&flagModel, "model", "", "Specify model (overrides config/default)")
	rootCmd.PersistentFlags().BoolVar(&flagStream, "stream", true, "Enable streaming output (token-by-token)")
}

func initializeConfig() error {
	if err := configInit(cfgFile); err != nil {
		if cfgFile != "" {
			return fmt.Errorf("failed to initialize config from %s: %w", cfgFile, err)
		}

		return fmt.Errorf("failed to initialize config: %w", err)
	}

	settings := configGet()
	if settings.ConfigFileUsed != "" {
		logger.Info("Using config file: %s", settings.ConfigFileUsed)
	} else {
		logger.Info("No configuration file found, using defaults")
	}

	return nil
}
