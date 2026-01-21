package services

import (
	"bytes"
	"os/exec"
	"strings"
)

// MetadataSidecar represents the .meta.json file structure
type MetadataSidecar struct {
	MimeType    string `json:"mime_type"`
	Description string `json:"description"`
	Status      string `json:"status,omitempty"`
}

// DetectMimeType uses the 'file' command to detect MIME type from byte slice.
// It acts as a wrapper around "file -b --mime-type -".
func DetectMimeType(data []byte) string {
	cmd := exec.Command("file", "-b", "--mime-type", "-")
	cmd.Stdin = bytes.NewReader(data)
	out, err := cmd.Output()
	if err != nil {
		// Fallback to generic binary if 'file' command fails
		return "application/octet-stream"
	}
	return strings.TrimSpace(string(out))
}
