package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderLoadAppliesEnvOverridesAndResolvesSettings(t *testing.T) {
	loader := &Loader{
		envLookup: func(key string) string {
			switch key {
			case "KOMMIT_PROVIDER":
				return "other"
			case "KOMMIT_MODEL":
				return "other-model"
			case "OTHER_API_KEY":
				return "other-key"
			case "OPENAI_API_KEY":
				return "shared-key"
			default:
				return ""
			}
		},
		gitDirLookup: func() (string, error) { return "", nil },
	}

	settings := loader.resolve(&Config{
		Providers: map[string]ProviderConfig{
			"ollama": {
				Name:    "Ollama",
				Type:    "openai-compat",
				BaseURL: "http://localhost:11434/v1/",
				Models:  []ModelConfig{{ID: "default-model"}},
			},
			"other": {
				Name:    "Other",
				Type:    "openai-compat",
				BaseURL: "https://example.com/v1/",
				Models:  []ModelConfig{{ID: "other-model"}},
			},
		},
		DefaultProvider: "ollama",
		DefaultModel:    "default-model",
	}, "/tmp/config.yaml")

	if settings.Provider != "other" {
		t.Fatalf("expected provider override, got %q", settings.Provider)
	}
	if settings.Model != "other-model" {
		t.Fatalf("expected model override, got %q", settings.Model)
	}
	if settings.ConfigFileUsed != "/tmp/config.yaml" {
		t.Fatalf("expected config file provenance, got %q", settings.ConfigFileUsed)
	}
	if len(settings.ProviderOrder) != 2 || settings.ProviderOrder[0] != "other" {
		t.Fatalf("expected provider order starting with override, got %v", settings.ProviderOrder)
	}

	provider, err := settings.ProviderConfig("other")
	if err != nil {
		t.Fatalf("ProviderConfig() returned error: %v", err)
	}
	if provider.APIKey != "other-key" {
		t.Fatalf("expected provider-specific API key, got %q", provider.APIKey)
	}

	sharedProvider, err := settings.ProviderConfig("ollama")
	if err != nil {
		t.Fatalf("ProviderConfig() returned error: %v", err)
	}
	if sharedProvider.APIKey != "shared-key" {
		t.Fatalf("expected shared OPENAI_API_KEY, got %q", sharedProvider.APIKey)
	}
}

func TestLoaderLoadReadsConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".kommit.yaml")
	configContents := []byte("default_provider: test\ndefault_model: test-model\nproviders:\n  test:\n    name: Test\n    base_url: https://example.com/v1/\n    type: openai-compat\n    models:\n      - id: test-model\n        name: Test Model\n")
	if err := os.WriteFile(configPath, configContents, 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	loader := &Loader{
		configFile: configPath,
		envLookup:  func(string) string { return "" },
	}

	settings, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if settings.ConfigFileUsed != configPath {
		t.Fatalf("expected config file used %q, got %q", configPath, settings.ConfigFileUsed)
	}
	if settings.Provider != "test" {
		t.Fatalf("expected provider test, got %q", settings.Provider)
	}
	if settings.Model != "test-model" {
		t.Fatalf("expected model test-model, got %q", settings.Model)
	}
}

func TestLoaderConfigFilePathsIncludesGitDir(t *testing.T) {
	loader := &Loader{
		gitDirLookup: func() (string, error) { return "/tmp/repo", nil },
		envLookup: func(key string) string {
			if key == "XDG_CONFIG_HOME" {
				return "/tmp/xdg"
			}
			return ""
		},
	}

	paths := loader.configFilePaths()
	var hasGitDir bool
	for _, path := range paths {
		if path == "/tmp/repo/.kommit.yaml" {
			hasGitDir = true
			break
		}
	}
	if !hasGitDir {
		t.Fatalf("expected git dir config path in %v", paths)
	}
}
