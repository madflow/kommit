package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/git"
)

// Client represents an Ollama API client
type Client struct {
	BaseURL string
	Model   string
}

// Request represents a request to the Ollama API
type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// Response represents a response from the Ollama API
type Response struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewClient creates a new Ollama client with the given configuration
func NewClient(cfg *config.OllamaConfig) *Client {
	return &Client{
		BaseURL: cfg.ServerURL,
		Model:   cfg.Model,
	}
}

// GenerateCommitMessage generates a commit message using the Ollama API
func (c *Client) GenerateCommitMessage(diff, rules string, repoCtx *git.RepoContext) (string, error) {
	// Truncate diff if it's too long (Ollama has token limits)
	maxDiffLength := 4000
	if len(diff) > maxDiffLength {
		diff = diff[:maxDiffLength] + "\n... (truncated)"
	}

	// Build the prompt using the rules and repository context
	prompt := fmt.Sprintf(`
You are a git commit message generator. 
Output ONLY the commit message in plain text format with no additional text, headers, or formatting.

Repository Context:
- Branch: %s
- Files changed: %d
- Changed files:%s

IMPORTANT Rules:
%s

Git diff:
%s`,
		repoCtx.BranchName,
		repoCtx.FilesChanged,
		func() string {
			if len(repoCtx.FileChanges) == 0 {
				return " (none)"
			}
			var files []string
			for _, change := range repoCtx.FileChanges {
				files = append(files, fmt.Sprintf("\n  - [%s] %s (%s)", change.Status, change.FilePath, change.FileType))
			}
			return strings.Join(files, "")
		}(),
		rules,
		diff)

	fmt.Println(prompt)

	reqBody, err := json.Marshal(Request{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Make the request
	resp, err := http.Post(c.BaseURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("error making request to Ollama: %v", err)
	}
	defer resp.Body.Close()

	var ollamaResp Response
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return ollamaResp.Response, nil
}
