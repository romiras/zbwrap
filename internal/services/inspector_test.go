package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryInspector_Sync_Lazy(t *testing.T) {
	// Setup temp repository structure
	repoDir, err := os.MkdirTemp("", "zbwrap-test-inspector")
	require.NoError(t, err)
	defer os.RemoveAll(repoDir)

	backupsDir := filepath.Join(repoDir, "backups")
	err = os.MkdirAll(backupsDir, 0755)
	require.NoError(t, err)

	// Create a dummy .zbk file without metadata
	zbkPath := filepath.Join(backupsDir, "2024-01-01_1000-manual.zbk")
	err = os.WriteFile(zbkPath, []byte("zbackup data content"), 0644)
	require.NoError(t, err)

	inspector := NewRepositoryInspector()

	// Test: Lazy sync (no deep inspection)
	// zbackupPath is irrelevant for lazy sync
	err = inspector.Sync("mock-zbackup", repoDir, false, "")
	assert.NoError(t, err)

	// Verify metadata creation
	metaPath := zbkPath + ".meta"
	assert.FileExists(t, metaPath)

	var meta MetadataSidecar
	data, err := os.ReadFile(metaPath)
	require.NoError(t, err)
	err = json.Unmarshal(data, &meta)
	require.NoError(t, err)

	assert.Equal(t, "unknown", meta.MimeType)
	assert.Equal(t, "complete", meta.Status)
}

func TestRepositoryInspector_Inspect(t *testing.T) {
	// Setup temp repository
	repoDir, err := os.MkdirTemp("", "zbwrap-test-inspect")
	require.NoError(t, err)
	defer os.RemoveAll(repoDir)

	backupsDir := filepath.Join(repoDir, "backups")
	err = os.MkdirAll(backupsDir, 0755)
	require.NoError(t, err)

	// 1. Create a dated file with metadata
	date1 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	name1 := date1.Format("2006-01-02_1504") + "-test1.zbk"
	path1 := filepath.Join(backupsDir, name1)
	err = os.WriteFile(path1, []byte("content"), 0644)
	require.NoError(t, err)

	meta1 := MetadataSidecar{
		MimeType:    "application/json",
		Description: "First backup",
	}
	data1, _ := json.Marshal(meta1)
	err = os.WriteFile(path1+".meta", data1, 0644)
	require.NoError(t, err)

	// 2. Create a dated file WITHOUT metadata (should be unknown)
	date2 := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
	name2 := date2.Format("2006-01-02_1504") + "-test2.zbk"
	path2 := filepath.Join(backupsDir, name2)
	err = os.WriteFile(path2, []byte("content"), 0644)
	require.NoError(t, err)

	// 3. Create a junk file (should be ignored)
	err = os.WriteFile(filepath.Join(backupsDir, "junk.txt"), []byte("junk"), 0644)
	require.NoError(t, err)

	inspector := NewRepositoryInspector()
	details, err := inspector.Inspect("test-alias", repoDir)
	assert.NoError(t, err)

	assert.Equal(t, "test-alias", details.Alias)
	assert.Equal(t, repoDir, details.PhysicalPath)
	assert.Equal(t, 2, len(details.Backups)) // Only 2 valid backups

	// Validation sort order (descending date)
	assert.Equal(t, name2, details.Backups[0].Filename)
	assert.Equal(t, name1, details.Backups[1].Filename)

	// Validate metadata parsing
	assert.Equal(t, "application/json", details.Backups[1].MimeType)
	assert.Equal(t, "First backup", details.Backups[1].Description)
	assert.True(t, details.Backups[1].HasMetadata)

	// Validate missing metadata handling
	assert.Equal(t, "unknown", details.Backups[0].MimeType)
	assert.False(t, details.Backups[0].HasMetadata)
}

func TestRepositoryInspector_Sync_Deep(t *testing.T) {
	// Setup temp repository
	repoDir, err := os.MkdirTemp("", "zbwrap-test-deep")
	require.NoError(t, err)
	defer os.RemoveAll(repoDir)

	backupsDir := filepath.Join(repoDir, "backups")
	err = os.MkdirAll(backupsDir, 0755)
	require.NoError(t, err)

	// Create a backup file
	zbkPath := filepath.Join(backupsDir, "2024-01-01_1200-deep.zbk")
	err = os.WriteFile(zbkPath, []byte("ignored content in file itself"), 0644)
	require.NoError(t, err)

	// Create a .meta file with "unknown"
	initialMeta := MetadataSidecar{MimeType: "unknown", Status: "complete"}
	metaData, _ := json.Marshal(initialMeta)
	err = os.WriteFile(zbkPath+".meta", metaData, 0644)
	require.NoError(t, err)

	// Create a mock zbackup script
	// This script ignores arguments and outputs a fixed string "Detected Content"
	// which will be sniffed as text/plain.
	mockScript := filepath.Join(repoDir, "mock_zbackup.sh")
	scriptContent := `#!/bin/sh
# Output enough text to fill part of the buffer and ensure text detection
echo "This is some mock content that should be detected as text/plain."
echo "Repeating content to ensure we have enough bytes if needed."
`
	err = os.WriteFile(mockScript, []byte(scriptContent), 0755)
	require.NoError(t, err)

	inspector := NewRepositoryInspector()

	// Test: Deep sync
	err = inspector.Sync(mockScript, repoDir, true, "")
	assert.NoError(t, err)

	// Verify metadata update
	var updatedMeta MetadataSidecar
	data, err := os.ReadFile(zbkPath + ".meta")
	require.NoError(t, err)
	err = json.Unmarshal(data, &updatedMeta)
	require.NoError(t, err)

	// http.DetectContentType for simple text usually returns "text/plain; charset=utf-8"
	assert.Contains(t, updatedMeta.MimeType, "text/plain")
	// Make sure it didn't stay unknown
	assert.NotEqual(t, "unknown", updatedMeta.MimeType)
}
