package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// AppName is the name of the application
	AppName = "kommit"
	// ConfigFileName is the name of the config file (without extension)
	ConfigFileName = "config"
	// ConfigFileExt is the extension for the config file
	ConfigFileExt = "yaml"
)

// Init initializes the configuration
func Init(configFile string) error {
	// If config file is specified, use that
	if configFile != "" {
		viper.SetConfigFile(configFile)
		return viper.ReadInConfig()
	}

	// Set the config name and type
	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType(ConfigFileExt)

	// Set config search paths in order of preference
	configDirs := getConfigDirs()
	for _, dir := range configDirs {
		viper.AddConfigPath(dir)
	}

	// Read in the config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		// If we can't read the config file, it's not an error if it doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// Read in environment variables that match
	viper.AutomaticEnv()

	return nil
}

// getConfigDirs returns a list of directories to search for configuration files
// in order of preference
func getConfigDirs() []string {
	var dirs []string

	// Check XDG config home
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		dirs = append(dirs, filepath.Join(xdgConfigHome, AppName))
	}

	// Check standard XDG config directory
	home, err := os.UserHomeDir()
	if err == nil {
		xdgConfigPath := filepath.Join(home, ".config", AppName)
		dirs = append(dirs, xdgConfigPath)

		// Add legacy config path for backward compatibility
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
