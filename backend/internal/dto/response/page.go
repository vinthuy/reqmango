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

	// Locking
	LockedByID   *uint64    `json:"locked_by_id"`
	LockedAt     *time.Time `json:"locked_at"`
	LockedByName string     `json:"locked_by_name,omitempty"`

	ProjectID   uint64 `json:"project_id"`
	WorkspaceID uint64 `json:"workspace_id"`

	CreatedByID *uint64   `json:"created_by_id"`
	UpdatedByID *uint64   `json:"updated_by_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Children []PageResponse `json:"children,omitempty"`
}
