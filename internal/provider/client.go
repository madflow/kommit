package provider

import (
	"fmt"
	"os"
	"strings"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/logger"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Client struct {
	openaiClients map[string]*openai.Client
	config        *config.Config
	streamEnabled bool
}

func NewClient(cfg *config.Config) *Client {
	clients := make(map[string]*openai.Client)

	for name, provider := range cfg.Providers {
		opts := []option.RequestOption{
			option.WithBaseURL(provider.BaseURL),
		}

		apiKey := resolveAPIKey(name, provider)
		if apiKey != "" {
			opts = append(opts, option.WithAPIKey(apiKey))
		}

		if os.Getenv("OPENAI_DEBUG") != "" {
			opts = append(opts, option.WithDebugLog(nil))
		}

		client := openai.NewClient(opts...)
		clients[name] = &client
	}

	return &Client{
		openaiClients: clients,
		config:        cfg,
		streamEnabled: true,
	}
}

func NewClientWithStreaming(cfg *config.Config, enabled bool) *Client {
	client := NewClient(cfg)
	client.streamEnabled = enabled
	return client
}

func resolveAPIKey(name string, provider config.ProviderConfig) string {
	envKey := fmt.Sprintf("%s_API_KEY", strings.ToUpper(name))
	if key := os.Getenv(envKey); key != "" {
		return key
	}

	if key := os.Getenv("OPENAI_API_KEY"); key != "" && provider.Type == "openai-compat" {
		return key
	}

	return provider.APIKey
}

func (c *Client) GetOpenAIClient(providerName string) (*openai.Client, error) {
	client, ok := c.openaiClients[providerName]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not found or not initialized", providerName)
	}
	return client, nil
}

func (c *Client) GetModel(providerName, modelID string) (*config.ModelConfig, error) {
	return c.config.GetModel(providerName, modelID)
}

func (c *Client) LogProviderError(providerName string, err error) {
	logger.Error("Provider '%s' failed: %v", providerName, err)
}
