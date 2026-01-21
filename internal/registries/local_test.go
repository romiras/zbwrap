package registries

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalRegistry_Add_Get_List(t *testing.T) {
	// Create a temporary directory to act as a valid repository target
	tempDir, err := os.MkdirTemp("", "zbwrap-test-repo")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := NewLocalRegistry()

	// Test: Add valid repository
	err = registry.Add("test-repo", tempDir)
	assert.NoError(t, err)

	// Test: Add duplicate alias
	err = registry.Add("test-repo", tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test: Add non-existent path
	err = registry.Add("bad-repo", "/path/to/nothing/hopefully")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path does not exist")

	// Test: Add file as path (should fail, must be dir)
	tempFile := filepath.Join(tempDir, "file.txt")
	err = os.WriteFile(tempFile, []byte("content"), 0644)
	require.NoError(t, err)
	err = registry.Add("file-repo", tempFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")

	// Test: Get existing
	path, ok := registry.Get("test-repo")
	assert.True(t, ok)
	assert.Equal(t, tempDir, path)

	// Test: Get non-existent
	_, ok = registry.Get("unknown")
	assert.False(t, ok)

	// Test: List
	list := registry.List()
	assert.Len(t, list, 1)
	assert.Equal(t, tempDir, list["test-repo"])
}

func TestLocalRegistry_Save_Load(t *testing.T) {
	// Reset viper for this test
	viper.Reset()

	// Create a temporary config file
	tempDir, err := os.MkdirTemp("", "zbwrap-test-config")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "registry.json")
	viper.SetConfigFile(configFile)

	// Setup registry with some data
	registry := NewLocalRegistry()
	registry.ZBackupPath = "/usr/local/bin/zbackup"
	registry.Encryption.Type = "none"

	// Create a dummy repo dir
	repoDir, err := os.MkdirTemp("", "zbwrap-repo")
	require.NoError(t, err)
	defer os.RemoveAll(repoDir)

	err = registry.Add("my-repo", repoDir)
	require.NoError(t, err)

	// Test: Save
	err = registry.Save()
	assert.NoError(t, err)
	assert.FileExists(t, configFile)

	// Test: Load into a new instance
	// We need to reset viper or confirm it reads from the file we just wrote.
	// Since viper singleton is used in Load(), we just ensure it points to our file.
	newRegistry := NewLocalRegistry()
	err = newRegistry.Load()
	assert.NoError(t, err)

	// Verify data persistence
	assert.Equal(t, "/usr/local/bin/zbackup", newRegistry.ZBackupPath)
	assert.Equal(t, "none", newRegistry.Encryption.Type)

	path, ok := newRegistry.Get("my-repo")
	assert.True(t, ok)
	assert.Equal(t, repoDir, path)

	// Check LastUpdated is populated
	assert.False(t, newRegistry.LastUpdated.IsZero())
}
