package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madflow/kommit/internal/git"
	"github.com/spf13/viper"
)

type Config struct {
	Providers       map[string]ProviderConfig `mapstructure:"providers"`
	DefaultProvider string                    `mapstructure:"default_provider"`
	DefaultModel    string                    `mapstructure:"default_model"`
	Rules           string                    `mapstructure:"rules"`
	PRRules         string                    `mapstructure:"pr_rules"`
	PRTitleRules    string                    `mapstructure:"pr_title_rules"`
}

type ProviderConfig struct {
	Name    string        `mapstructure:"name"`
	BaseURL string        `mapstructure:"base_url"`
	Type    string        `mapstructure:"type"`
	APIKey  string        `mapstructure:"api_key"`
	Models  []ModelConfig `mapstructure:"models"`
}

type ModelConfig struct {
	Name             string `mapstructure:"name"`
	ID               string `mapstructure:"id"`
	ContextWindow    int    `mapstructure:"context_window"`
	DefaultMaxTokens int    `mapstructure:"default_max_tokens"`
}

type OllamaConfig struct {
	ServerURL string `mapstructure:"server_url"`
	Model     string `mapstructure:"model"`
}

func DefaultConfig() *Config {
	return &Config{
		Providers: map[string]ProviderConfig{
			"ollama": {
				Name:    "Ollama",
				BaseURL: "http://localhost:11434/v1/",
				Type:    "openai-compat",
				Models: []ModelConfig{
					{
						Name:             "Qwen 2.5 Coder 7B",
						ID:               "qwen2.5-coder:7b",
						ContextWindow:    32768,
						DefaultMaxTokens: 4096,
					},
				},
			},
		},
		DefaultProvider: "ollama",
		DefaultModel:    "qwen2.5-coder:7b",
		Rules: `
		Expected output format:

		First line: summary under 80 characters

		Optional body paragraphs if needed

		IMPORTANT: Do not deviate from this format.

	- Begin the message with a short summary of your changes (up to 80 characters as a guideline).
	- Do not use any emoji or markdown in the commit message.
	- Do not use any formatting characters including asterisks (*), underscores (_), backticks, brackets [] or any other markup symbols.
	- Write in plain text only - no bold, italic, or code formatting.
	- Use simple, direct language without any text decoration or emphasis markers.
	- Write as if you're typing in a plain text editor with no formatting options.
	- Do not wrap words or phrases in any special characters.
	- Avoid using quotation marks around technical terms unless they are part of the actual code/file names.
  - For longer commit messages, create a separate message body.
  - Separate the message body by including a blank line.
  - The body of your message should provide a more detailed answers how the changes differ from the previous implementation.
  - Use the imperative, present tense («change», not «changed» or «changes») to be consistent with generated messages from commands like git merge.
  - Be direct, try to eliminate filler words and phrases in these sentences (examples: though, maybe, I think, kind of).`,
		PRRules: `
		Expected output format:

		## Summary

		<1-3 bullet points>

		Do not deviate from this format.

	- Create a concise summary for a pull request in the format "## Summary" followed by 1-3 bullet points.
	- Each bullet point should highlight a key change or improvement made in this pull request.
	- Focus on the value and impact of the changes, not just what files were modified.
	- Write in plain text only - no bold, italic, or code formatting except for the ## Summary header.
	- Use simple, direct language that explains what was accomplished.
	- Be specific about features added, bugs fixed, or improvements made.
	- Avoid generic descriptions like "updated files" or "made changes".
	- Keep bullet points concise but informative (aim for 1-2 lines each).
	- Use the past tense for describing what was accomplished in this pull request.
	- Focus on the user-facing or developer-facing benefits of the changes.`,
		PRTitleRules: `
	- Create a concise and descriptive pull request title.
	- Maximum length: 50 characters (aim for under 50 characters).
	- Do no use any emoji or markdown in the title.
	- Do not use any formatting characters including asterisks (*), underscores (_), backticks, brackets [] or any other markup symbols.
	- Do not use newline characters in the title.
	- Always start with a capital letter.
	- Use imperative mood ("Add feature" not "Added feature" or "Adds feature").
	- Start with a verb when possible (Add, Fix, Update, Remove, etc.).
	- Be specific about what was changed or accomplished.
	- Do not use conventional commit prefixes (feat:, fix:, etc.) unless explicitly requested.
	- Avoid articles (a, an, the) when possible to save space.
	- Do not end with a period.
	- Focus on the primary change or most important aspect.
	- Use title case for the first word only.
	- Examples: "Add user authentication system", "Fix memory leak in parser", "Update API documentation".`,
	}
}

const (
	AppName                  = "kommit"
	ConfigFileName           = "config"
	StandaloneConfigFileName = ".kommit"
	ConfigFileExt            = "yaml"
)

var appConfig *Config

func readAndUnmarshalConfig() error {
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		configFile = "(unknown file)"
	}

	appConfig = &Config{}
	if err := viper.Unmarshal(appConfig); err != nil {
		return fmt.Errorf("error parsing %s: %w", configFile, err)
	}

	migrateLegacyConfig(appConfig)

	return nil
}

func migrateLegacyConfig(cfg *Config) {
	if len(cfg.Providers) == 0 {
		var legacyOllama OllamaConfig
		if err := viper.UnmarshalKey("ollama", &legacyOllama); err == nil {
			if legacyOllama.ServerURL != "" && legacyOllama.Model != "" {
				baseURL := legacyOllama.ServerURL
				if strings.HasSuffix(baseURL, "/api/generate") {
					baseURL = strings.TrimSuffix(baseURL, "/api/generate") + "/v1/"
				}

				cfg.Providers = map[string]ProviderConfig{
					"ollama": {
						Name:    "Ollama",
						BaseURL: baseURL,
						Type:    "openai-compat",
						Models: []ModelConfig{
							{
								ID:               legacyOllama.Model,
								Name:             legacyOllama.Model,
								ContextWindow:    32768,
								DefaultMaxTokens: 4096,
							},
						},
					},
				}
				cfg.DefaultProvider = "ollama"
				cfg.DefaultModel = legacyOllama.Model
			}
		}
	}
}

func Init(configFile string) error {
	defaults := DefaultConfig()
	if len(defaults.Providers) > 0 {
		for providerName, provider := range defaults.Providers {
			viper.SetDefault(fmt.Sprintf("providers.%s", providerName), provider)
		}
	}
	viper.SetDefault("default_provider", defaults.DefaultProvider)
	viper.SetDefault("default_model", defaults.DefaultModel)
	viper.SetDefault("rules", defaults.Rules)
	viper.SetDefault("pr_rules", defaults.PRRules)
	viper.SetDefault("pr_title_rules", defaults.PRTitleRules)

	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := readAndUnmarshalConfig(); err != nil {
			return fmt.Errorf("error loading config from %s: %w", configFile, err)
		}
		return nil
	}

	for _, configPath := range getConfigFilePaths() {
		if _, err := os.Stat(configPath); err == nil {
			viper.SetConfigFile(configPath)
			if err := readAndUnmarshalConfig(); err == nil {
				return nil
			}
		}
	}

	appConfig = DefaultConfig()
	viper.AutomaticEnv()
	if err := viper.Unmarshal(appConfig); err != nil {
		return fmt.Errorf("error applying environment overrides: %w", err)
	}

	return nil
}

func Get() *Config {
	if appConfig == nil {
		return DefaultConfig()
	}
	return appConfig
}

func (c *Config) GetProvider(name string) (*ProviderConfig, error) {
	if name == "" {
		name = c.DefaultProvider
	}
	if name == "" && len(c.Providers) > 0 {
		for pname := range c.Providers {
			name = pname
			break
		}
	}

	provider, ok := c.Providers[name]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}
	return &provider, nil
}

func (c *Config) GetModel(providerName, modelID string) (*ModelConfig, error) {
	provider, err := c.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	if modelID == "" {
		modelID = c.DefaultModel
	}

	for i := range provider.Models {
		if provider.Models[i].ID == modelID {
			return &provider.Models[i], nil
		}
	}

	if modelID == c.DefaultModel && len(provider.Models) > 0 {
		return &provider.Models[0], nil
	}

	return nil, fmt.Errorf("model '%s' not found in provider '%s'", modelID, providerName)
}

func (c *Config) ResolveProviderModel(providerFlag, modelFlag string) (string, string) {
	provider := providerFlag
	model := modelFlag

	if provider == "" {
		provider = os.Getenv("KOMMIT_PROVIDER")
	}
	if model == "" {
		model = os.Getenv("KOMMIT_MODEL")
	}

	if provider == "" {
		provider = c.DefaultProvider
	}
	if model == "" {
		model = c.DefaultModel
	}

	if provider == "" && len(c.Providers) > 0 {
		for pname := range c.Providers {
			provider = pname
			break
		}
	}

	return provider, model
}

func (c *Config) GetProviderFallbackOrder(preferredProvider string) []string {
	var order []string
	if preferredProvider != "" {
		order = append(order, preferredProvider)
	}

	for name := range c.Providers {
		if name != preferredProvider {
			order = append(order, name)
		}
	}

	return order
}

func getConfigFilePaths() []string {
	var paths []string

	if pwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(pwd, StandaloneConfigFileName+"."+ConfigFileExt))
	}

	if gitDir, err := git.GetGitDir(); err == nil && gitDir != "" {
		paths = append(paths, filepath.Join(gitDir, StandaloneConfigFileName+"."+ConfigFileExt))
	}

	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		paths = append(paths, filepath.Join(xdgConfigHome, AppName, ConfigFileName+"."+ConfigFileExt))
	}

	home, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(home, ".config", AppName, ConfigFileName+"."+ConfigFileExt))
		paths = append(paths, filepath.Join(home, StandaloneConfigFileName+"."+ConfigFileExt))
	}

	return paths
}

type Getter interface {
	GetString(key string) string
	GetStringMap(key string) map[string]any
	GetStringMapString(key string) map[string]string
	GetStringSlice(key string) []string
	GetInt(key string) int
	GetBool(key string) bool
}

func Viper() Getter {
	return viper.GetViper()
}
