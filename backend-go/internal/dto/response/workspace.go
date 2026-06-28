package response

import "time"

// WorkspaceResponse is the full workspace representation.
type WorkspaceResponse struct {
	ID               uint64         `json:"id"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	LogoURL          *string        `json:"logo_url"`
	OrganizationSize *string        `json:"organization_size"`
	Timezone         string         `json:"timezone"`
	OwnerID          uint64         `json:"owner_id"`
	Owner            UserResponse   `json:"owner"`
	TotalMembers     int64          `json:"total_members"`
	TotalProjects    int64          `json:"total_projects"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	CreatedByID      *uint64        `json:"created_by_id"`
	UpdatedByID      *uint64        `json:"updated_by_id"`
	DeletedAt        *time.Time     `json:"deleted_at"`
	IsDeleted        bool           `json:"is_deleted"`
}

// WorkspaceLite is a compact workspace representation.
type WorkspaceLite struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// WorkspaceMemberResponse represents a workspace member.
type WorkspaceMemberResponse struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspace_id"`
	UserID      uint64    `json:"user_id"`
	Role        int       `json:"role"`
	IsActive    bool      `json:"is_active"`
	User        UserLite  `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
