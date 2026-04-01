package models

// Pagination metadata embedded in list responses
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedResponse wraps a paginated list response
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    PaginationMeta `json:"meta,omitempty"`
}

// ExportPayload represents a full export of all data
type ExportPayload struct {
	Version   string   `json:"version"`
	ExportedAt string  `json:"exported_at"`
	Prompts   []Prompt `json:"prompts,omitempty"`
	Skills    []Skill  `json:"skills,omitempty"`
	Agents    []Agent  `json:"agents,omitempty"`
}

// FailedItem records a single import failure.
type FailedItem struct {
	Index int    `json:"index"`
	Title string `json:"title,omitempty"`
	Error string `json:"error"`
}

// ImportResult reports how many items were imported and which failed.
type ImportResult struct {
	Success    bool          `json:"success"`
	Imported   int           `json:"imported"`
	Failed     []FailedItem  `json:"failed,omitempty"`
	TotalCount int           `json:"total_count"`
}
