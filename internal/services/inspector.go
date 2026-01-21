package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

// BackupItem represents a single backup artifact
type BackupItem struct {
	Filename    string    `json:"filename"`
	Date        time.Time `json:"date"`
	MimeType    string    `json:"mime_type"`
	Description string    `json:"description"`
	HasMetadata bool      `json:"has_metadata"`
}

// RepoDetails holds detailed information about a repository
type RepoDetails struct {
	Alias          string       `json:"repository_alias"`
	PhysicalPath   string       `json:"physical_path"`
	TotalSizeBytes int64        `json:"total_size_bytes"`
	Backups        []BackupItem `json:"backups"`
}

// RepositoryInspector handles inspection logic
type RepositoryInspector struct{}

// NewRepositoryInspector creates a new inspector
func NewRepositoryInspector() *RepositoryInspector {
	return &RepositoryInspector{}
}

// Inspect gathers details about a repository
func (i *RepositoryInspector) Inspect(alias, path string) (*RepoDetails, error) {
	details := &RepoDetails{
		Alias:        alias,
		PhysicalPath: path,
		Backups:      []BackupItem{},
	}

	// Calculate total size
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to calculate disk usage: %w", err)
	}
	details.TotalSizeBytes = size

	// Scan backups directory
	backupsDir := filepath.Join(path, "backups")
	if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
		// No backups directory, return empty backups list
		return details, nil
	}

	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backups directory: %w", err)
	}

	// Regex for "YYYY-MM-DD_HHMM"
	// Example: 2024-05-10_0800-initial.zbk
	reDate := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}_\d{4})`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := info.Name()
		// Basic filter for .zbk files, though spec implies specific naming schema
		// But let's assume anything in backups/ might be relevant, forcing .zbk extension check is safer
		if filepath.Ext(name) != ".zbk" {
			continue
		}

		item := BackupItem{
			Filename: name,
			MimeType: "unknown", // default
		}

		// Parse date
		matches := reDate.FindStringSubmatch(name)
		if len(matches) > 1 {
			// Layout: YYYY-MM-DD_HHMM
			parsed, err := time.Parse("2006-01-02_1504", matches[1])
			if err == nil {
				item.Date = parsed
			}
		} else {
			// Fallback to file mod time if naming convention isn't followed?
			// Spec says naming is enforced, but "unknown" items might exist
			item.Date = info.ModTime()
		}

		// Check for metadata sidecar
		metaPath := filepath.Join(backupsDir, name+".meta")
		if metaBytes, err := os.ReadFile(metaPath); err == nil {
			var meta MetadataSidecar
			if err := json.Unmarshal(metaBytes, &meta); err == nil {
				item.HasMetadata = true
				item.MimeType = meta.MimeType
				item.Description = meta.Description
			}
		}

		details.Backups = append(details.Backups, item)
	}

	// Sort backups by date descending
	sort.Slice(details.Backups, func(i, j int) bool {
		return details.Backups[i].Date.After(details.Backups[j].Date)
	})

	return details, nil
}

// Sync performs a synchronization of the repository, generating missing metadata.
// If deep is true, it attempts to detect MIME types by restoring the beginning of the backup.
func (i *RepositoryInspector) Sync(zbackupPath, repoPath string, deep bool, passwordFile string) error {
	backupsDir := filepath.Join(repoPath, "backups")
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read backups directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".zbk" {
			continue
		}

		zbkPath := filepath.Join(backupsDir, entry.Name())
		metaPath := zbkPath + ".meta"

		var meta MetadataSidecar
		metaExists := false

		if data, err := os.ReadFile(metaPath); err == nil {
			if err := json.Unmarshal(data, &meta); err == nil {
				metaExists = true
			}
		}

		if !metaExists {
			// Lazy: Create skeleton
			meta = MetadataSidecar{
				MimeType: "unknown",
				Status:   "complete",
			}
			if err := i.saveMetadata(metaPath, meta); err != nil {
				// Don't abort, just log or skip? For now, we return error as this is crucial.
				// Although failing one shouldn't stop others, we'll try to persist.
				return fmt.Errorf("failed to write metadata for %s: %w", entry.Name(), err)
			}
		}

		if deep && meta.MimeType == "unknown" {
			mime := i.SniffMimeType(zbackupPath, zbkPath, passwordFile)
			if mime != "unknown" {
				meta.MimeType = mime
				if err := i.saveMetadata(metaPath, meta); err != nil {
					return fmt.Errorf("failed to update metadata for %s: %w", entry.Name(), err)
				}
			}
		}
	}
	return nil
}

// SniffMimeType attempts to detect the MIME type of a backup by restoring the first 512 bytes.
func (i *RepositoryInspector) SniffMimeType(zbackupPath, zbkPath, passwordFile string) string {
	args := []string{"restore"}
	if passwordFile != "" {
		args = append(args, "--password-file", passwordFile)
	} else {
		// If no password file is provided, assume non-encrypted and pass the flag
		// zbackup requires explicit confirmation for non-encrypted restores
		args = append(args, "--non-encrypted")
	}
	args = append(args, zbkPath)

	cmd := exec.Command(zbackupPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "unknown"
	}

	if err := cmd.Start(); err != nil {
		return "unknown"
	}

	// We only need the first 512 bytes for MIME detection
	buf := make([]byte, 512)
	n, err := stdout.Read(buf)

	// Kill the process as we don't need the rest of the stream
	// Ignore errors from Kill/Wait as we are terminating forcibly
	_ = cmd.Process.Kill()
	_ = cmd.Wait()

	if n == 0 {
		return "unknown"
	}

	return DetectMimeType(buf[:n])
}

// saveMetadata writes the metadata to the specified path
func (i *RepositoryInspector) saveMetadata(path string, meta MetadataSidecar) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
