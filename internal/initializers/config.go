package initializers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Load initializes configuration via Viper
func Load() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(home, ".config", "zbwrap")

	viper.AddConfigPath(configDir)
	viper.SetConfigName("registry")
	viper.SetConfigType("json")

	// Create config directory if it doesn't exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		_ = os.MkdirAll(configDir, 0755)
	}

	// Try to read config, but don't fail if it doesn't exist yet (first run)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error produced
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}
	}
}
