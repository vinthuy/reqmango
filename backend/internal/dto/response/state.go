package response

import "time"

// StateResponse is the state representation returned by the API.
type StateResponse struct {
	ID          *uint64    `json:"id"`
	Name        string     `json:"name"`
	Color       string     `json:"color"`
	Group       string     `json:"group"`
	Sequence    int        `json:"sequence"`
	IsDefault   bool       `json:"is_default"`
	IsActive    bool       `json:"is_active"`
	ProjectID   *uint64    `json:"project_id"`
	WorkspaceID uint64     `json:"workspace_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedByID *uint64    `json:"created_by_id"`
	UpdatedByID *uint64    `json:"updated_by_id"`
	DeletedAt   *time.Time `json:"deleted_at"`
	IsDeleted   bool       `json:"is_deleted"`
	IsInherited bool       `json:"is_inherited"`
}
