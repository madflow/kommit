package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/madflow/kommit/internal/config"
)

// Client represents an Ollama API client
type Client struct {
	BaseURL string
	Model  string
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
		Model:  cfg.Model,
	}
}

// GenerateCommitMessage generates a commit message using the Ollama API
func (c *Client) GenerateCommitMessage(diff, rules string) (string, error) {
	// Truncate diff if it's too long (Ollama has token limits)
	maxDiffLength := 4000
	if len(diff) > maxDiffLength {
		diff = diff[:maxDiffLength] + "\n... (truncated)"
	}

	// Build the prompt using the rules
	prompt := fmt.Sprintf(`You are a git commit message generator. Analyze the git diff and respond with ONLY the commit message.

Rules:
%s

Git diff:
%s`, rules, diff)

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
