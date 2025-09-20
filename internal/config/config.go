package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madflow/kommit/internal/git"
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Ollama OllamaConfig `mapstructure:"ollama"`
	Rules  string       `mapstructure:"rules"`
}

// OllamaConfig holds configuration for the Ollama API
type OllamaConfig struct {
	ServerURL string `mapstructure:"server_url"`
	Model     string `mapstructure:"model"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Ollama: OllamaConfig{
			ServerURL: "http://localhost:11434/api/generate",
			Model:     "qwen2.5-coder:7b",
		},
		Rules: `
		Expected output format:

		---
		First line: summary under 80 characters

		Optional body paragraphs if needed
		---

		Do not deviate from this format.

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
	// Try to read the config file
	err := viper.ReadInConfig()
	if err != nil {
		return err // Return the original error to handle it in the caller
	}

	// Get the config file path for better error reporting
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		configFile = "(unknown file)"
	}

	// If we get here, we successfully read a config file
	// Now unmarshal it into our config struct
	appConfig = &Config{}
	if err := viper.Unmarshal(appConfig); err != nil {
		return fmt.Errorf("error parsing %s: %w", configFile, err)
	}
	return nil
}

// Init initializes the configuration
func Init(configFile string) error {
	// Set defaults
	defaults := DefaultConfig()
	viper.SetDefault("ollama.server_url", defaults.Ollama.ServerURL)
	viper.SetDefault("ollama.model", defaults.Ollama.Model)
	viper.SetDefault("rules", defaults.Rules)

	// If config file is explicitly specified, use that
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := readAndUnmarshalConfig(); err != nil {
			return fmt.Errorf("error loading config from %s: %w", configFile, err)
		}
		return nil
	}

	// Try each potential config file in order of preference
	for _, configPath := range getConfigFilePaths() {
		if _, err := os.Stat(configPath); err == nil {
			viper.SetConfigFile(configPath)
			if err := readAndUnmarshalConfig(); err == nil {
				// Successfully loaded a config file
				return nil
			}
			// Continue to next config source if there's an error reading this one
		}
	}

	// If we get here, no config file was found
	appConfig = DefaultConfig()

	// Apply environment variables on top of the default config
	viper.AutomaticEnv()

	// Apply environment variable overrides
	if err := viper.Unmarshal(appConfig); err != nil {
		return fmt.Errorf("error applying environment overrides: %w", err)
	}

	return nil
}

// Get returns the loaded configuration
func Get() *Config {
	if appConfig == nil {
		return DefaultConfig()
	}
	return appConfig
}

// getConfigFilePaths returns a list of potential config file paths in order of preference:
// 1. $PWD/.kommit.yaml
// 2. $GIT_DIR/.kommit.yaml (if inside a git repository)
// 3. $XDG_CONFIG_HOME/kommit/config.yaml
// 4. $HOME/.config/kommit/config.yaml
// 5. $HOME/.kommit.yaml
func getConfigFilePaths() []string {
	var paths []string

	// 1. Current working directory
	if pwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(pwd, StandaloneConfigFileName+"."+ConfigFileExt))
	}

	// 2. Git directory
	if gitDir, err := git.GetGitDir(); err == nil && gitDir != "" {
		paths = append(paths, filepath.Join(gitDir, StandaloneConfigFileName+"."+ConfigFileExt))
	}

	// 3. XDG config home
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		paths = append(paths, filepath.Join(xdgConfigHome, AppName, ConfigFileName+"."+ConfigFileExt))
	}

	// 4. Standard XDG config directory
	home, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(home, ".config", AppName, ConfigFileName+"."+ConfigFileExt))

		// 5. Home directory
		paths = append(paths, filepath.Join(home, StandaloneConfigFileName+"."+ConfigFileExt))
	}

	return paths
}

// GetString wraps viper.GetString
type Getter interface {
	GetString(key string) string
	GetStringMap(key string) map[string]any
	GetStringMapString(key string) map[string]string
	GetStringSlice(key string) []string
	GetInt(key string) int
	GetBool(key string) bool
}

// Viper returns the underlying viper instance
func Viper() Getter {
	return viper.GetViper()
}
