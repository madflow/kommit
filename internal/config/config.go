package config

import (
	"fmt"
	"os"
	"path/filepath"

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
		Rules: `- Begin the message with a short summary of your changes (up to 80 characters as a guideline).
  - Capitalization and Punctuation: Capitalize the first word in the sentence and do not end in punctuation.
  - Separate it from the following body by including a blank line.
  - The body of your message should provide a more detailed answers how the changes differ from the previous implementation.
  - Use the imperative, present tense («change», not «changed» or «changes») to be consistent with generated messages from commands like git merge.
  - Be direct, try to eliminate filler words and phrases in these sentences (examples: though, maybe, I think, kind of).`,
	}
}

const (
	// AppName is the name of the application
	AppName = "kommit"
	// ConfigFileName is the name of the config file (without extension)
	ConfigFileName = "config"
	// LegacyConfigFileName is the name of the config file in the current directory
	LegacyConfigFileName = ".kommit"
	// ConfigFileExt is the extension for the config file
	ConfigFileExt = "yaml"
)

// appConfig holds the loaded configuration
var appConfig *Config

// readAndUnmarshalConfig reads the current config file and unmarshals it into appConfig
func readAndUnmarshalConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	// Unmarshal the config
	appConfig = &Config{}
	if err := viper.Unmarshal(appConfig); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
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

	// First try to load .kommit.yaml from current directory
	if pwd, err := os.Getwd(); err == nil {
		legacyConfig := filepath.Join(pwd, LegacyConfigFileName+"."+ConfigFileExt)
		if _, err := os.Stat(legacyConfig); err == nil {
			viper.SetConfigFile(legacyConfig)
			if err := readAndUnmarshalConfig(); err != nil {
				return fmt.Errorf("error loading config from %s: %w", legacyConfig, err)
			}
			return nil
		}
	}

	// Set up search paths for config.yaml
	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType(ConfigFileExt)
	configDirs := getConfigDirs()
	for _, dir := range configDirs {
		viper.AddConfigPath(dir)
	}

	// Try to read the config
	if err := readAndUnmarshalConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// Read in environment variables that match
	viper.AutomaticEnv()

	// Unmarshal the config
	appConfig = &Config{}
	if err := viper.Unmarshal(appConfig); err != nil {
		return err
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

// getConfigDirs returns a list of directories to search for configuration files
// in order of preference:
// 1. $PWD (for .kommit.yaml)
// 2. $XDG_CONFIG_HOME/kommit (for config.yaml)
// 3. $HOME/.config/kommit (for config.yaml)
// 4. $HOME (for .kommit.yaml)
func getConfigDirs() []string {
	var dirs []string

	// 1. Current working directory (for .kommit.yaml)
	if pwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, pwd)
	}

	// 2. XDG config home (for config.yaml)
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		dirs = append(dirs, filepath.Join(xdgConfigHome, AppName))
	}

	// 3. Standard XDG config directory (for config.yaml)
	home, err := os.UserHomeDir()
	if err == nil {
		dirs = append(dirs, filepath.Join(home, ".config", AppName))
	}

	// 4. Home directory (for .kommit.yaml)
	if home != "" {
		dirs = append(dirs, home)
	}

	return dirs
}

// GetString wraps viper.GetString
type Getter interface {
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringSlice(key string) []string
	GetInt(key string) int
	GetBool(key string) bool
}

// Viper returns the underlying viper instance
func Viper() Getter {
	return viper.GetViper()
}
