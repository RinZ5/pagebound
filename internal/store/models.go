package store

type Book struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Creator     *string `json:"creator,omitempty"`
	Language    *string `json:"language,omitempty"`
	Identifier  *string `json:"identifier,omitempty"`
	Description *string `json:"description,omitempty"`
	Publisher   *string `json:"publisher,omitempty"`
	FilePath    string  `json:"file_path"`
	CoverPath   *string `json:"cover_path,omitempty"`
	FileSize    int64   `json:"file_size"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
