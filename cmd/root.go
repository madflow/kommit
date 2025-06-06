package cmd

import (
	"bufio"
	"os"
	"strings"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/git"
	"github.com/madflow/kommit/internal/logger"
	"github.com/madflow/kommit/internal/ollama"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

type CommitMessage struct {
	Message string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kommit",
	Short: "Git commits for the rest of us",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Println("🤖 Kommit")
		logger.Println("================================")

		// Check if we're in a git repository
		if !git.IsGitRepo() {
			logger.Fatal("Not in a git repository")
		}

		// Check if there are any changes to commit
		hasChanges, err := git.HasChangesToCommit()
		if err != nil {
			logger.Fatal("Error checking for changes: %v", err)
		}

		if !hasChanges {
			logger.Success("No changes to commit")
			return
		}

		// Check git status
		status, err := git.GetGitStatus()
		if err != nil {
			logger.Fatal("Error getting git status: %v", err)
		}

		logger.Println("📊 Git Status:")
		logger.Println(status)
		logger.Println()

		// Get git diff
		diff, err := git.GetGitDiff()
		if err != nil {
			logger.Fatal("Error getting git diff: %v", err)
		}

		if strings.TrimSpace(diff) == "" {
			logger.Info("No staged changes found. Staging all changes...")
			if err := git.StageAllChanges(); err != nil {
				logger.Fatal("Error staging changes: %v", err)
			}

			// Get diff again after staging
			diff, err = git.GetGitDiff()
			if err != nil {
				logger.Fatal("Error getting git diff after staging: %v", err)
			}

			// Double-check we have actual changes after staging
			if strings.TrimSpace(diff) == "" {
				logger.Success("No actual changes found after staging")
				return
			}
		}

		// Final check: ensure we have meaningful diff content
		if len(strings.TrimSpace(diff)) < 10 {
			logger.Success("No meaningful changes to commit")
			return
		}

		logger.Info("Analyzing changes...")

		// Generate commit message using Ollama
		cfg := config.Get()
		ollamaClient := ollama.NewClient(&cfg.Ollama)
		messageText, err := ollamaClient.GenerateCommitMessage(diff, cfg.Rules)
		if err != nil {
			logger.Fatal("Error generating commit message: %v", err)
		}
		message := &CommitMessage{
			Message: strings.TrimSpace(messageText),
		}

		// Display generated message
		logger.Println("\n📝 Generated Commit Message:")
		logger.Printf("Message: %s\n\n", message.Message)

		// Ask user for confirmation
		if !askForConfirmation() {
			logger.Error("Commit cancelled by user")
			return
		}

		// Commit the changes
		if err := git.CommitChanges(message.Message); err != nil {
			logger.Fatal("Error committing changes: %v", err)
		}

		logger.Success("Changes committed successfully!")
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logger.Fatal("Command failed: %v", err)
	}
}



func askForConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	logger.Printf("Do you want to commit with this message? [y/N] ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))
	return text == "y" || text == "yes"
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/kommit/config.yaml or $HOME/.config/kommit/config.yaml)")
}

// initConfig initializes the configuration
func initConfig() {
	// Initialize configuration
	if err := config.Init(cfgFile); err != nil {
		logger.Fatal("Failed to initialize config: %v", err)
	}

	// Log the config file being used if any
	if viper.ConfigFileUsed() != "" {
		logger.Info("Using config file: %s", viper.ConfigFileUsed())
	} else {
		logger.Info("No configuration file found, using defaults")
	}
}
