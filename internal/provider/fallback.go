package provider

import (
	"fmt"

	"github.com/madflow/kommit/internal/config"
)

func (c *Client) ListProviders() []string {
	var providers []string
	for name := range c.config.Providers {
		providers = append(providers, name)
	}
	return providers
}

func (c *Client) ListModels(providerName string) ([]config.ModelConfig, error) {
	provider, err := c.config.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return provider.Models, nil
}

func (c *Client) ValidateProviderAndModel(providerName, modelID string) error {
	provider, err := c.config.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("provider validation failed: %w", err)
	}

	if modelID == "" {
		modelID = c.config.DefaultModel
	}

	found := false
	for _, model := range provider.Models {
		if model.ID == modelID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("model '%s' not found in provider '%s'", modelID, providerName)
	}

	return nil
}

func (c *Client) GetDefaultMaxTokens(providerName, modelID string) int {
	model, err := c.GetModel(providerName, modelID)
	if err != nil || model.DefaultMaxTokens == 0 {
		return 4096
	}
	return model.DefaultMaxTokens
}
