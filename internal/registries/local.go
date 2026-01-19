package registries

import (
	"time"
)

// EncryptionConfig holds encryption settings
type EncryptionConfig struct {
	Type            string `json:"type" mapstructure:"type"`
	CredentialsPath string `json:"credentials_path,omitempty" mapstructure:"credentials_path"`
}

// Registry represents the structure of registry.json
type Registry struct {
	Repositories map[string]string `json:"repositories" mapstructure:"repositories"`
	Encryption   EncryptionConfig  `json:"encryption" mapstructure:"encryption"`
	LastUpdated  time.Time         `json:"last_updated" mapstructure:"last_updated"`
}
