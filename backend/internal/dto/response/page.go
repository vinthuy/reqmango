package response

import "time"

// PageResponse is the API response for a page.
type PageResponse struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	ContentJSON *string    `json:"content_json"`
	Published   bool       `json:"published"`
	ArchivedAt  *time.Time `json:"archived_at"`
	Sequence    int        `json:"sequence"`
	ParentID    *uint64    `json:"parent_id"`
	Depth       int        `json:"depth"`
	ProjectID   uint64     `json:"project_id"`
	WorkspaceID uint64     `json:"workspace_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Children    []PageResponse `json:"children,omitempty"`
}
