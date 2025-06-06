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

type CommitMessage struct {
	Message string
}

func main() {
	fmt.Println("🤖 Kommit ")
	fmt.Println("================================")

	// Check if we're in a git repository
	if !isGitRepo() {
		fmt.Println("❌ Error: Not in a git repository")
		os.Exit(1)
	}

	// Check if there are any changes to commit
	hasChanges, err := hasChangesToCommit()
	if err != nil {
		fmt.Printf("❌ Error checking for changes: %v\n", err)
		os.Exit(1)
	}

	if !hasChanges {
		fmt.Println("✅ No changes to commit")
		return
	}

	// Check git status
	status, err := getGitStatus()
	if err != nil {
		fmt.Printf("❌ Error getting git status: %v\n", err)
		os.Exit(1)
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
	if err := commitChanges(message.Message); err != nil {
		fmt.Printf("❌ Error committing changes: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Changes committed successfully!")
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func hasChangesToCommit() (bool, error) {
	// Check for staged changes
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	if err := cmd.Run(); err == nil {
		// No staged changes, check for unstaged changes
		cmd = exec.Command("git", "diff", "--quiet")
		if err := cmd.Run(); err == nil {
			// Check for untracked files
			cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
			output, err := cmd.Output()
			if err != nil {
				return false, err
			}
			return strings.TrimSpace(string(output)) != "", nil
		}
		return true, nil // Has unstaged changes
	}
	return true, nil // Has staged changes
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

	reqBody := OllamaRequest{
		Model:  "qwen2.5-coder:7b",
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

	return parseCommitMessage(ollamaResp.Response)
}

func parseCommitMessage(response string) (*CommitMessage, error) {
	lines := strings.Split(response, "\n")

	// Find the first non-empty, meaningful line that doesn't contain meta text
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Remove any quotes
		line = strings.Trim(line, `"'`)

		// Truncate to 80 characters if necessary
		if len(line) > 80 {
			line = line[:77] + "..."
		}

		// Must have some actual content
		if len(line) > 3 {
			return &CommitMessage{Message: line}, nil
		}
	}

	// Fallback
	return &CommitMessage{Message: "Update files"}, nil
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
