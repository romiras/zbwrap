package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"zbwrap/internal/registries"
)

// BackupRunner handles the execution of ZBackup processes
type BackupRunner struct {
	registry *registries.LocalRegistry
}

// NewBackupRunner creates a new backup runner
func NewBackupRunner(registry *registries.LocalRegistry) *BackupRunner {
	return &BackupRunner{
		registry: registry,
	}
}

// Backup performs a backup operation
func (r *BackupRunner) Backup(repoPath, suffix, description string, reader io.Reader) error {
	// 1. Generate filename
	timestamp := time.Now().Format("2006-01-02_1504")
	filename := fmt.Sprintf("%s-%s.zbk", timestamp, suffix)
	backupsDir := filepath.Join(repoPath, "backups")

	// Ensure backups directory exists
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		return fmt.Errorf("failed to create backups directory: %w", err)
	}

	filePath := filepath.Join(backupsDir, filename)
	metaPath := filePath + ".meta"

	// 2. Sniff MIME type from the first 512 bytes
	sniffBuf := make([]byte, 512)
	n, err := io.ReadFull(reader, sniffBuf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("failed to read for MIME detection: %w", err)
	}
	sniffBuf = sniffBuf[:n]
	mimeType := DetectMimeType(sniffBuf)

	// Combine sniffBuf and the rest of reader
	combinedReader := io.MultiReader(bytes.NewReader(sniffBuf), reader)

	// 3. Prepare ZBackup command
	args := []string{"backup", filePath}

	// Add encryption flags if necessary
	if r.registry.Encryption.Type == "password-file" && r.registry.Encryption.CredentialsPath != "" {
		args = append([]string{"--password-file", r.registry.Encryption.CredentialsPath}, args...)
	} else {
		// Default to non-encrypted if not specified or explicitly set to "none"
		args = append([]string{"--non-encrypted"}, args...)
	}

	// Determine zbackup binary path
	zbackupPath := r.registry.ZBackupPath
	if zbackupPath == "" {
		zbackupPath = "zbackup"
	}

	cmd := exec.Command(zbackupPath, args...)
	cmd.Stdin = combinedReader
	cmd.Stderr = os.Stderr // Pipe stderr to our stderr for visibility

	// 4. Create metadata sidecar (temporarily, will be kept on success)
	meta := MetadataSidecar{
		MimeType:    mimeType,
		Description: description,
		Status:      "success",
	}

	metaFile, err := os.Create(metaPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}

	if err := json.NewEncoder(metaFile).Encode(meta); err != nil {
		metaFile.Close()
		os.Remove(metaPath)
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	metaFile.Close()

	// 5. Run ZBackup
	if err := cmd.Run(); err != nil {
		// Cleanup metadata on failure
		os.Remove(metaPath)
		// Also cleanup the partial .zbk file if it exists?
		// ZBackup might leave it. Spec says "Clean up .meta on zbackup failure."
		return fmt.Errorf("zbackup failed: %w", err)
	}

	return nil
}
