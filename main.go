package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type CommitMessages struct {
	Short string
	Long  string
}

func main() {
	fmt.Println("🤖 Ollama Git Auto-Commit Tool")
	fmt.Println("================================")

	// Check if we're in a git repository
	if !isGitRepo() {
		fmt.Println("❌ Error: Not in a git repository")
		os.Exit(1)
	}

	// Check git status
	status, err := getGitStatus()
	if err != nil {
		fmt.Printf("❌ Error getting git status: %v\n", err)
		os.Exit(1)
	}

	if strings.TrimSpace(status) == "" {
		fmt.Println("✅ No changes to commit")
		return
	}

	fmt.Println("📊 Git Status:")
	fmt.Println(status)
	fmt.Println()

	// Get git diff
	diff, err := getGitDiff()
	if err != nil {
		fmt.Printf("❌ Error getting git diff: %v\n", err)
		os.Exit(1)
	}

	if strings.TrimSpace(diff) == "" {
		fmt.Println("📝 No staged changes found. Staging all changes...")
		if err := stageAllChanges(); err != nil {
			fmt.Printf("❌ Error staging changes: %v\n", err)
			os.Exit(1)
		}

		// Get diff again after staging
		diff, err = getGitDiff()
		if err != nil {
			fmt.Printf("❌ Error getting git diff after staging: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("🔍 Analyzing changes with Ollama...")

	// Generate commit messages using Ollama
	messages, err := generateCommitMessages(diff)
	if err != nil {
		fmt.Printf("❌ Error generating commit messages: %v\n", err)
		os.Exit(1)
	}

	// Display generated messages
	fmt.Println("\n📝 Generated Commit Messages:")
	fmt.Println("Short:", messages.Short)
	fmt.Println("Long:", messages.Long)
	fmt.Println()

	// Ask user for confirmation
	if !askForConfirmation() {
		fmt.Println("❌ Commit cancelled by user")
		return
	}

	// Commit the changes
	if err := commitChanges(messages.Long); err != nil {
		fmt.Printf("❌ Error committing changes: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Changes committed successfully!")
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func getGitStatus() (string, error) {
	cmd := exec.Command("git", "status", "-v")
	output, err := cmd.Output()
	return string(output), err
}

func getGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// If no staged changes, get unstaged diff
	if strings.TrimSpace(string(output)) == "" {
		cmd = exec.Command("git", "diff")
		output, err = cmd.Output()
	}

	return string(output), err
}

func stageAllChanges() error {
	cmd := exec.Command("git", "add", ".")
	return cmd.Run()
}

func generateCommitMessages(diff string) (*CommitMessages, error) {
	// Truncate diff if it's too long (Ollama has token limits)
	maxDiffLength := 4000
	if len(diff) > maxDiffLength {
		diff = diff[:maxDiffLength] + "\n... (truncated)"
	}

	prompt := fmt.Sprintf(`Based on the following git diff, generate two commit messages:

1. A short commit message (max 50 characters, conventional commit format)
2. A longer commit message (max 72 characters per line, explaining what and why)

Git diff:
%s

Please respond in this exact format:
SHORT: [short message]
LONG: [longer message]

Focus on what changed and why it's important. Use conventional commit prefixes like feat:, fix:, docs:, refactor:, etc.`, diff)

	reqBody := OllamaRequest{
		Model:  "llama3.1", // You can change this to your preferred model
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama (make sure it's running): %v", err)
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&ollamaResp); err != nil {
		return nil, err
	}

	return parseCommitMessages(ollamaResp.Response)
}

func parseCommitMessages(response string) (*CommitMessages, error) {
	lines := strings.Split(response, "\n")
	messages := &CommitMessages{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(line), "SHORT:") {
			messages.Short = strings.TrimSpace(strings.TrimPrefix(line, "SHORT:"))
			messages.Short = strings.TrimSpace(strings.TrimPrefix(messages.Short, "short:"))
		} else if strings.HasPrefix(strings.ToUpper(line), "LONG:") {
			messages.Long = strings.TrimSpace(strings.TrimPrefix(line, "LONG:"))
			messages.Long = strings.TrimSpace(strings.TrimPrefix(messages.Long, "long:"))
		}
	}

	// Fallback if parsing fails
	if messages.Short == "" || messages.Long == "" {
		// Try to extract first meaningful line as short, rest as long
		meaningfulLines := []string{}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "Based on") {
				meaningfulLines = append(meaningfulLines, line)
			}
		}

		if len(meaningfulLines) > 0 {
			if messages.Short == "" {
				messages.Short = meaningfulLines[0]
				if len(messages.Short) > 50 {
					messages.Short = messages.Short[:47] + "..."
				}
			}
			if messages.Long == "" {
				messages.Long = strings.Join(meaningfulLines, "\n")
			}
		}
	}

	// Final fallback
	if messages.Short == "" {
		messages.Short = "chore: update files"
	}
	if messages.Long == "" {
		messages.Long = messages.Short
	}

	return messages, nil
}

func askForConfirmation() bool {
	fmt.Print("Do you want to commit these changes? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func commitChanges(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
