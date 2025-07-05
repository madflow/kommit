package cmd

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/git"
	"github.com/madflow/kommit/internal/logger"
	"github.com/madflow/kommit/internal/ollama"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	yolo    bool
	add     bool
)

type CommitMessage struct {
	Message string
}

// yoloCommit performs an automatic commit and push without confirmation
func yoloCommit(message string) {
	logger.Info("🚀 YOLO mode enabled - Automatically committing and pushing changes")

	// Commit the changes (changes already staged in the main flow)
	if err := git.CommitChanges(message); err != nil {
		logger.Fatal("Error committing changes: %v", err)
	}

	// Push to remote
	if err := git.PushCurrentBranch(); err != nil {
		logger.Fatal("Error pushing changes: %v", err)
	}

	logger.Success("Changes committed and pushed successfully!")
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

		// If add flag is set, stage all changes
		if add {
			if err := git.AddAll(); err != nil {
				logger.Fatal("Error staging changes: %v", err)
			}
		}

		// In yolo mode, stage all changes first, then check for staged changes
		if yolo {
			if err := git.AddAll(); err != nil {
				logger.Fatal("Error staging changes: %v", err)
			}
		}

		// Check for staged changes to commit
		hasChanges, err := git.HasStagedChanges()
		if err != nil {
			logger.Fatal("Error checking for changes: %v", err)
		}

		if !hasChanges {
			logger.Success("No changes to commit")
			return
		}

		// Get repository context
		repoCtx, err := git.GetRepoContext()
		if err != nil {
			logger.Fatal("Error getting repository context: %v", err)
		}

		// Display repository context
		logger.Println("📊 Repository Context:")
		logger.Printf("Branch name: %s\n", repoCtx.BranchName)
		logger.Printf("Files changed: %d\n", repoCtx.FilesChanged)

		if repoCtx.FilesChanged > 0 {
			logger.Println("\n📝 Change Summary:")
			logger.Println(repoCtx.ChangeSummary)

			if len(repoCtx.FileChanges) > 0 {
				logger.Println("\n📋 File Changes:")
				for _, change := range repoCtx.FileChanges {
					logger.Printf("[%s] %s (%s)\n", change.Status, change.FilePath, change.FileType)
				}
			}
		}

		logger.Println()

		// Get git diff for AI analysis
		diff, err := git.GetGitDiff()
		if err != nil {
			logger.Fatal("Error getting git diff: %v", err)
		}

		logger.Info("Analyzing changes...")

		// Generate commit message using Ollama
		cfg := config.Get()
		ollamaClient := ollama.NewClient(&cfg.Ollama)
		messageText, err := ollamaClient.GenerateCommitMessage(diff, cfg.Rules, repoCtx)
		if err != nil {
			logger.Fatal("Error generating commit message: %v", err)
		}
		message := &CommitMessage{
			Message: strings.TrimSpace(messageText),
		}

		// Display generated message
		logger.Println("\n📝 Generated Commit Message:")
		logger.Printf("%s\n\n", message.Message)

		if yolo {
			yoloCommit(message.Message)
		} else {
			// Loop until user confirms or cancels
			for {
				switch askForConfirmation() {
				case "no":
					logger.Error("Commit cancelled by user")
					return
				case "edit":
					tempFile, err := os.CreateTemp("", "kommit-*.md")
					if err != nil {
						logger.Fatal("Error creating temporary file: %v", err)
					}
					defer os.Remove(tempFile.Name())

					// Write the current message to the temp file
					if _, err := tempFile.WriteString(message.Message); err != nil {
						tempFile.Close()
						logger.Fatal("Error writing to temporary file: %v", err)
					}
					tempFile.Close()

					// Open the editor
					editor := os.Getenv("EDITOR")
					if editor == "" {
						editor = "vi" // Default to vi if no editor is set
					}

					// Execute the editor
					cmd := exec.Command(editor, tempFile.Name())
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						logger.Fatal("Error opening editor: %v", err)
					}

					// Read the edited message
					editedMessage, err := os.ReadFile(tempFile.Name())
					if err != nil {
						logger.Fatal("Error reading edited message: %v", err)
					}

					message.Message = strings.TrimSpace(string(editedMessage))
					logger.Println("\n📝 Updated Commit Message:")
					logger.Printf("%s\n\n", message.Message)
					continue // Go back to the confirmation prompt

				case "yes":
					// Break out of the loop to proceed with the commit
					goto commit
				}
			}

		commit:

			// Commit the changes
			if err := git.CommitChanges(message.Message); err != nil {
				logger.Fatal("Error committing changes: %v", err)
			}

			logger.Success("Changes committed successfully!")
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logger.Fatal("Command failed: %v", err)
	}
}

// askForConfirmation prompts the user to confirm the commit message
// Returns: "yes", "edit", or "no"
func askForConfirmation() string {
	reader := bufio.NewReader(os.Stdin)
	for {
		logger.Printf("Do you want to commit with this message? [y/e/N] ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(strings.ToLower(text))
		switch text {
		case "y", "yes":
			return "yes"
		case "e", "edit":
			return "edit"
		case "", "n", "no":
			return "no"
		default:
			logger.Printf("Please enter 'y' for yes, 'e' to edit, or 'N' for no\n")
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/kommit/config.yaml or $HOME/.config/kommit/config.yaml)")
	rootCmd.Flags().BoolVarP(&yolo, "yolo", "y", false, "Automatically stage all changes, commit, and push without confirmation")
	rootCmd.Flags().BoolVarP(&add, "add", "a", false, "Stage all changes before committing")
}

// initConfig initializes the configuration
func initConfig() {
	// Initialize configuration
	if err := config.Init(cfgFile); err != nil {
		if cfgFile != "" {
			logger.Fatal("Failed to initialize config from %s: %v", cfgFile, err)
		} else {
			logger.Fatal("Failed to initialize config: %v", err)
		}
	}

	// Log the config file being used if any
	if viper.ConfigFileUsed() != "" {
		logger.Info("Using config file: %s", viper.ConfigFileUsed())
	} else {
		logger.Info("No configuration file found, using defaults")
	}
}
