package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/madflow/kommit/internal/git"
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
		fmt.Println("🤖 Kommit ")
		fmt.Println("================================")

		// Check if we're in a git repository
		if !git.IsGitRepo() {
			fmt.Println("❌ Error: Not in a git repository")
			os.Exit(1)
		}

		// Check if there are any changes to commit
		hasChanges, err := git.HasChangesToCommit()
		if err != nil {
			fmt.Printf("❌ Error checking for changes: %v\n", err)
			os.Exit(1)
		}

		if !hasChanges {
			fmt.Println("✅ No changes to commit")
			return
		}

		// Check git status
		status, err := git.GetGitStatus()
		if err != nil {
			fmt.Printf("❌ Error getting git status: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("📊 Git Status:")
		fmt.Println(status)
		fmt.Println()

		// Get git diff
		diff, err := git.GetGitDiff()
		if err != nil {
			fmt.Printf("❌ Error getting git diff: %v\n", err)
			os.Exit(1)
		}

		if strings.TrimSpace(diff) == "" {
			fmt.Println("📝 No staged changes found. Staging all changes...")
			if err := git.StageAllChanges(); err != nil {
				fmt.Printf("❌ Error staging changes: %v\n", err)
				os.Exit(1)
			}

			// Get diff again after staging
			diff, err = git.GetGitDiff()
			if err != nil {
				fmt.Printf("❌ Error getting git diff after staging: %v\n", err)
				os.Exit(1)
			}

			// Double-check we have actual changes after staging
			if strings.TrimSpace(diff) == "" {
				fmt.Println("✅ No actual changes found after staging")
				return
			}
		}

		// Final check: ensure we have meaningful diff content
		if len(strings.TrimSpace(diff)) < 10 {
			fmt.Println("✅ No meaningful changes to commit")
			return
		}

		fmt.Println("🔍 Analyzing changes...")

		// Generate commit message using Ollama
		message, err := generateCommitMessage(diff)
		if err != nil {
			fmt.Printf("❌ Error generating commit message: %v\n", err)
			os.Exit(1)
		}

		// Display generated message
		fmt.Println("\n📝 Generated Commit Message:")
		fmt.Printf("Message: %s\n", message.Message)
		fmt.Println()

		// Ask user for confirmation
		if !askForConfirmation() {
			fmt.Println("❌ Commit cancelled by user")
			return
		}

		// Commit the changes
		if err := git.CommitChanges(message.Message); err != nil {
			fmt.Printf("❌ Error committing changes: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Changes committed successfully!")
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
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
	- How does it differ from the previous implementation? 
- Use the imperative, present tense («change», not «changed» or «changes») to be consistent with generated messages from commands like git merge
- NO prefixes like "feat:" or "fix:"
- NO explanatory text
- NO quotes
- Be concise and specific without to much detail

Git diff:
%s

COMMIT MESSAGE:`, diff)

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
	fmt.Print("Do you want to commit with this message? [y/N] ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))
	return text == "y" || text == "yes"
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kommit.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".kommit" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kommit")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
