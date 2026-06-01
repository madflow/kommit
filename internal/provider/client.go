package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/logger"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type StreamHandler func(chunk string)

type requestExecutor func(ctx context.Context, providerName, modelID, prompt string, stream bool, handler StreamHandler) (string, error)

type Client struct {
	openaiClients map[string]*openai.Client
	settings      *config.ResolvedSettings
	execute       requestExecutor
}

func NewClient(settings *config.ResolvedSettings) *Client {
	if settings == nil {
		settings = config.Get()
	}

	clients := make(map[string]*openai.Client)

	for name := range settings.Config.Providers {
		provider, err := settings.ProviderConfig(name)
		if err != nil {
			continue
		}
		opts := []option.RequestOption{
			option.WithBaseURL(provider.BaseURL),
		}

		if provider.APIKey != "" {
			opts = append(opts, option.WithAPIKey(provider.APIKey))
		}

		if os.Getenv("OPENAI_DEBUG") != "" {
			opts = append(opts, option.WithDebugLog(nil))
		}

		client := openai.NewClient(opts...)
		clients[name] = &client
	}

	return &Client{
		openaiClients: clients,
		settings:      settings,
	}
}

func (c *Client) GetOpenAIClient(providerName string) (*openai.Client, error) {
	client, ok := c.openaiClients[providerName]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not found or not initialized", providerName)
	}
	return client, nil
}

func (c *Client) GetModel(providerName, modelID string) (*config.ModelConfig, error) {
	return c.settings.ModelConfig(providerName, modelID)
}

func (c *Client) LogProviderError(providerName string, err error) {
	logger.Error("Provider '%s' failed: %v", providerName, err)
}

func (c *Client) executor() requestExecutor {
	if c.execute != nil {
		return c.execute
	}

	return c.executeRequest
}
