package registries

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// EncryptionConfig holds encryption settings
type EncryptionConfig struct {
	Type            string `json:"type" mapstructure:"type"`
	CredentialsPath string `json:"credentials_path,omitempty" mapstructure:"credentials_path"`
}

// LocalRegistry represents the structure of registry.json and implements RepositoryManager
type LocalRegistry struct {
	Repositories map[string]string `json:"repositories" mapstructure:"repositories"`
	Encryption   EncryptionConfig  `json:"encryption" mapstructure:"encryption"`
	LastUpdated  time.Time         `json:"last_updated" mapstructure:"last_updated"`
	mu           sync.RWMutex
}

// NewLocalRegistry creates a new registry instance
func NewLocalRegistry() *LocalRegistry {
	return &LocalRegistry{
		Repositories: make(map[string]string),
	}
}

// Load reads the registry from config
func (r *LocalRegistry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// It's okay if config doesn't exist yet
			return nil
		}
		return err
	}

	return viper.Unmarshal(r, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeHookFunc(time.RFC3339),
	)))
}

// Save writes the registry to config
func (r *LocalRegistry) Save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.LastUpdated = time.Now()
	viper.Set("repositories", r.Repositories)
	viper.Set("encryption", r.Encryption)
	viper.Set("last_updated", r.LastUpdated)

	// Ensure directory exists
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		// Fallback if no config file found yet
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".config", "zbwrap", "registry.json")
	} else {
		// viper.WriteConfig() works, but if we want to be safe about dir creation:
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	return viper.WriteConfigAs(configPath)
}

// Add adds a repository to the registry
func (r *LocalRegistry) Add(alias, path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Check if alias exists
	if _, exists := r.Repositories[alias]; exists {
		return fmt.Errorf("alias '%s' already exists", alias)
	}

	r.Repositories[alias] = path
	return nil
}

// Get retrieves a repository path by alias
func (r *LocalRegistry) Get(alias string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	path, ok := r.Repositories[alias]
	return path, ok
}

// List returns all repositories
func (r *LocalRegistry) List() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to be safe
	copy := make(map[string]string)
	for k, v := range r.Repositories {
		copy[k] = v
	}
	return copy
}
