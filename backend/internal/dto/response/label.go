package response

import "time"

// LabelResponse is the label representation returned by the API.
type LabelResponse struct {
	ID          *uint64    `json:"id"`
	Name        string     `json:"name"`
	Color       string     `json:"color"`
	Description *string    `json:"description"`
	ProjectID   *uint64    `json:"project_id"`
	WorkspaceID uint64     `json:"workspace_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedByID *uint64    `json:"created_by_id"`
	UpdatedByID *uint64    `json:"updated_by_id"`
	DeletedAt   *time.Time `json:"deleted_at"`
	IsDeleted   bool       `json:"is_deleted"`
}
