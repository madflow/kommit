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

type ResolvedSettings struct {
	Config          *Config
	ConfigFileUsed  string
	Provider        string
	Model           string
	ProviderOrder   []string
	ProviderAPIKeys map[string]string
}

type Loader struct {
	gitDirLookup func() (string, error)
	envLookup    func(string) string
	configFile   string
}

const (
	AppName                  = "kommit"
	ConfigFileName           = "config"
	StandaloneConfigFileName = ".kommit"
	ConfigFileExt            = "yaml"
)

var currentSettings *ResolvedSettings

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

func Init(configFile string) error {
	settings, err := NewLoader(configFile).Load()
	if err != nil {
		return err
	}

	currentSettings = settings
	return nil
}

func Get() *ResolvedSettings {
	if currentSettings == nil {
		settings, err := NewLoader("").Load()
		if err != nil {
			return defaultResolvedSettings("")
		}
		currentSettings = settings
	}

	return currentSettings
}

func NewLoader(configFile string) *Loader {
	return &Loader{
		gitDirLookup: git.GetGitDir,
		envLookup:    os.Getenv,
		configFile:   configFile,
	}
}

func (l *Loader) ResolveForTest(cfg *Config) *ResolvedSettings {
	return l.resolve(cfg, "")
}

func (l *Loader) LoadOrPanicForTest() *ResolvedSettings {
	settings, err := l.Load()
	if err != nil {
		panic(err)
	}
	return settings
}

func (l *Loader) Load() (*ResolvedSettings, error) {
	v := viper.New()
	applyDefaults(v, DefaultConfig())
	v.AutomaticEnv()

	loadedConfigFile := ""
	if l.configFile != "" {
		v.SetConfigFile(l.configFile)
		cfg, err := readConfig(v)
		if err != nil {
			return nil, fmt.Errorf("error loading config from %s: %w", l.configFile, err)
		}
		loadedConfigFile = v.ConfigFileUsed()
		return l.resolve(cfg, loadedConfigFile), nil
	}

	for _, path := range l.configFilePaths() {
		if _, err := os.Stat(path); err == nil {
			v.SetConfigFile(path)
			cfg, err := readConfig(v)
			if err == nil {
				loadedConfigFile = v.ConfigFileUsed()
				return l.resolve(cfg, loadedConfigFile), nil
			}
		}
	}

	defaultConfig := DefaultConfig()
	if err := v.Unmarshal(defaultConfig); err != nil {
		return nil, fmt.Errorf("error applying environment overrides: %w", err)
	}

	return l.resolve(defaultConfig, loadedConfigFile), nil
}

func readConfig(v *viper.Viper) (*Config, error) {
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	configFile := v.ConfigFileUsed()
	if configFile == "" {
		configFile = "(unknown file)"
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error parsing %s: %w", configFile, err)
	}

	migrateLegacyConfig(v, cfg)
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error applying environment overrides: %w", err)
	}

	return cfg, nil
}

func migrateLegacyConfig(v *viper.Viper, cfg *Config) {
	if len(cfg.Providers) != 0 {
		return
	}

	var legacyOllama OllamaConfig
	if err := v.UnmarshalKey("ollama", &legacyOllama); err != nil {
		return
	}
	if legacyOllama.ServerURL == "" || legacyOllama.Model == "" {
		return
	}

	baseURL := legacyOllama.ServerURL
	if strings.HasSuffix(baseURL, "/api/generate") {
		baseURL = strings.TrimSuffix(baseURL, "/api/generate") + "/v1/"
	}

	cfg.Providers = map[string]ProviderConfig{
		"ollama": {
			Name:    "Ollama",
			BaseURL: baseURL,
			Type:    "openai-compat",
			Models: []ModelConfig{{
				ID:               legacyOllama.Model,
				Name:             legacyOllama.Model,
				ContextWindow:    32768,
				DefaultMaxTokens: 4096,
			}},
		},
	}
	cfg.DefaultProvider = "ollama"
	cfg.DefaultModel = legacyOllama.Model
}

func applyDefaults(v *viper.Viper, defaults *Config) {
	if len(defaults.Providers) > 0 {
		for providerName, provider := range defaults.Providers {
			v.SetDefault(fmt.Sprintf("providers.%s", providerName), provider)
		}
	}
	v.SetDefault("default_provider", defaults.DefaultProvider)
	v.SetDefault("default_model", defaults.DefaultModel)
	v.SetDefault("rules", defaults.Rules)
	v.SetDefault("pr_rules", defaults.PRRules)
	v.SetDefault("pr_title_rules", defaults.PRTitleRules)
}

func (l *Loader) resolve(cfg *Config, configFileUsed string) *ResolvedSettings {
	provider := firstNonEmpty(l.envLookup("KOMMIT_PROVIDER"), cfg.DefaultProvider, firstProvider(cfg.Providers))
	model := firstNonEmpty(l.envLookup("KOMMIT_MODEL"), cfg.DefaultModel)

	return &ResolvedSettings{
		Config:          cfg,
		ConfigFileUsed:  configFileUsed,
		Provider:        provider,
		Model:           model,
		ProviderOrder:   providerOrder(cfg.Providers, provider),
		ProviderAPIKeys: resolveProviderAPIKeys(cfg.Providers, l.envLookup),
	}
}

func defaultResolvedSettings(configFileUsed string) *ResolvedSettings {
	loader := NewLoader("")
	return loader.resolve(DefaultConfig(), configFileUsed)
}

func (l *Loader) configFilePaths() []string {
	var paths []string

	if pwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(pwd, StandaloneConfigFileName+"."+ConfigFileExt))
	}

	if l.gitDirLookup != nil {
		if gitDir, err := l.gitDirLookup(); err == nil && gitDir != "" {
			paths = append(paths, filepath.Join(gitDir, StandaloneConfigFileName+"."+ConfigFileExt))
		}
	}

	if xdgConfigHome := l.envLookup("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		paths = append(paths, filepath.Join(xdgConfigHome, AppName, ConfigFileName+"."+ConfigFileExt))
	}

	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", AppName, ConfigFileName+"."+ConfigFileExt))
		paths = append(paths, filepath.Join(home, StandaloneConfigFileName+"."+ConfigFileExt))
	}

	return paths
}

func (s *ResolvedSettings) Rules() string {
	return s.Config.Rules
}

func (s *ResolvedSettings) PullRequestRules() string {
	return s.Config.PRRules
}

func (s *ResolvedSettings) PullRequestTitleRules() string {
	return s.Config.PRTitleRules
}

func (s *ResolvedSettings) ProviderConfig(name string) (*ProviderConfig, error) {
	if name == "" {
		name = s.Provider
	}
	provider, ok := s.Config.Providers[name]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}

	provider.APIKey = s.ProviderAPIKeys[name]
	return &provider, nil
}

func (s *ResolvedSettings) ModelConfig(providerName, modelID string) (*ModelConfig, error) {
	if providerName == "" {
		providerName = s.Provider
	}
	if modelID == "" {
		modelID = s.Model
	}

	provider, err := s.ProviderConfig(providerName)
	if err != nil {
		return nil, err
	}

	for i := range provider.Models {
		if provider.Models[i].ID == modelID {
			return &provider.Models[i], nil
		}
	}

	if modelID == s.Model && len(provider.Models) > 0 {
		return &provider.Models[0], nil
	}

	return nil, fmt.Errorf("model '%s' not found in provider '%s'", modelID, providerName)
}

func firstProvider(providers map[string]ProviderConfig) string {
	for name := range providers {
		return name
	}
	return ""
}

func providerOrder(providers map[string]ProviderConfig, preferred string) []string {
	var order []string
	if preferred != "" {
		order = append(order, preferred)
	}

	for name := range providers {
		if name != preferred {
			order = append(order, name)
		}
	}

	return order
}

func resolveProviderAPIKeys(providers map[string]ProviderConfig, env func(string) string) map[string]string {
	keys := make(map[string]string, len(providers))
	for name, provider := range providers {
		envKey := fmt.Sprintf("%s_API_KEY", strings.ToUpper(name))
		switch {
		case env(envKey) != "":
			keys[name] = env(envKey)
		case provider.Type == "openai-compat" && env("OPENAI_API_KEY") != "":
			keys[name] = env("OPENAI_API_KEY")
		default:
			keys[name] = provider.APIKey
		}
	}
	return keys
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
