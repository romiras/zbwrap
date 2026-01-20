package services

// MetadataSidecar represents the .meta.json file structure
type MetadataSidecar struct {
	MimeType    string `json:"mime_type"`
	Description string `json:"description"`
	Status      string `json:"status,omitempty"`
}
