package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/git"
	"github.com/madflow/kommit/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type CommitMessage struct {
	Message string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kommit",
	Short: "Git commits for the disillusioned human being",
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
		message, err := generateCommitMessage(diff)
		if err != nil {
			logger.Fatal("Error generating commit message: %v", err)
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

func generateCommitMessage(diff string) (*CommitMessage, error) {
	// Truncate diff if it's too long (Ollama has token limits)
	maxDiffLength := 4000
	if len(diff) > maxDiffLength {
		diff = diff[:maxDiffLength] + "\n... (truncated)"
	}

	prompt := fmt.Sprintf(`You are a git commit message generator. Analyze the git diff and respond with ONLY the commit message.

Rules:
- Begin your message with a short summary of your changes (up to 80 characters as a guideline). 
- Separate it from the following body by including a blank line. 
- The body of your message should provide detailed answers to the following questions:
  * Why is this change being made?
  * How does it address the issue?
  * Any side effects or other important information?

Git diff:
%s`, diff)

	reqBody, err := json.Marshal(OllamaRequest{
		Model:  "qwen2.5-coder:7b",
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error making request to Ollama: %v", err)
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &CommitMessage{
		Message: strings.TrimSpace(ollamaResp.Response),
	}, nil
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
