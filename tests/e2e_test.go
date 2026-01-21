package tests

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"zbwrap/internal/registries"
	"zbwrap/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createMockZBackup creates a dummy script that mimics zbackup's success behavior
func createMockZBackup(t *testing.T, path string) {
	content := `#!/bin/sh
# Mock zbackup: Consume stdin and exit 0
cat > /dev/null
exit 0
`
	err := os.WriteFile(path, []byte(content), 0755)
	require.NoError(t, err)
}

func TestE2E_Backup_MimeTypes(t *testing.T) {
	// 1. Setup Environment
	tempDir, err := os.MkdirTemp("", "zbwrap-e2e")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repoDir := filepath.Join(tempDir, "repo")
	err = os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	zbackupPath := filepath.Join(tempDir, "zbackup-mock")
	createMockZBackup(t, zbackupPath)

	// 2. Setup Registry
	registry := registries.NewLocalRegistry()
	registry.ZBackupPath = zbackupPath
	registry.Encryption.Type = "none"
	err = registry.Add("test-repo", repoDir)
	require.NoError(t, err)

	runner := services.NewBackupRunner(registry)

	tests := []struct {
		name         string
		generator    func() []byte
		expectedMime string
	}{
		{
			name: "TAR Archive",
			generator: func() []byte {
				buf := new(bytes.Buffer)
				tw := tar.NewWriter(buf)
				hdr := &tar.Header{
					Name: "test.txt",
					Mode: 0600,
					Size: 11,
				}
				if err := tw.WriteHeader(hdr); err != nil {
					panic(err)
				}
				if _, err := tw.Write([]byte("hello world")); err != nil {
					panic(err)
				}
				if err := tw.Close(); err != nil {
					panic(err)
				}
				return buf.Bytes()
			},
			expectedMime: "application/x-tar",
		},
		{
			name: "ZIP Archive (Stored)",
			generator: func() []byte {
				buf := new(bytes.Buffer)
				zw := zip.NewWriter(buf)
				f, err := zw.Create("test.txt")
				if err != nil {
					panic(err)
				}
				if _, err := f.Write([]byte("hello world")); err != nil {
					panic(err)
				}
				if err := zw.Close(); err != nil {
					panic(err)
				}
				return buf.Bytes()
			},
			expectedMime: "application/zip",
		},
		{
			name: "Plain Text",
			generator: func() []byte {
				return []byte("Just some plain text content here.")
			},
			// Go's detection often adds charset
			expectedMime: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputData := tt.generator()
			suffix := strings.ReplaceAll(tt.name, " ", "_")
			desc := fmt.Sprintf("Testing %s", tt.name)

			// Clean previous backups to easy find the new one (or rely on suffix)

			err := runner.Backup(repoDir, suffix, desc, bytes.NewReader(inputData))
			assert.NoError(t, err)

			// Find the generated .zbk.meta file
			// We scan the backups dir
			backupsDir := filepath.Join(repoDir, "backups")
			entries, err := os.ReadDir(backupsDir)
			require.NoError(t, err)

			var foundMeta *services.MetadataSidecar
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), suffix+".zbk.meta") {
					data, err := os.ReadFile(filepath.Join(backupsDir, e.Name()))
					require.NoError(t, err)

					var meta services.MetadataSidecar
					err = json.Unmarshal(data, &meta)
					require.NoError(t, err)

					foundMeta = &meta
					break
				}
			}

			require.NotNil(t, foundMeta, "Metadata file not found for suffix %s", suffix)
			assert.Contains(t, foundMeta.MimeType, tt.expectedMime)
			assert.Equal(t, desc, foundMeta.Description)
			assert.Equal(t, "success", foundMeta.Status)
		})
	}
}
