package response

import "time"

// ProjectResponse is the full project representation.
type ProjectResponse struct {
	ID                uint64         `json:"id"`
	Name              string         `json:"name"`
	Identifier        string         `json:"identifier"`
	Description       *string        `json:"description"`
	IsPublic          bool           `json:"is_public"`
	Timezone          string         `json:"timezone"`
	ArchivedAt        *time.Time     `json:"archived_at"`
	WorkspaceID       uint64         `json:"workspace_id"`
	Workspace         *WorkspaceLite `json:"workspace"`
	DefaultAssigneeID *uint64        `json:"default_assignee_id"`
	DefaultAssignee   *UserLite      `json:"default_assignee"`
	ProjectLeadID     *uint64        `json:"project_lead_id"`
	ProjectLead       *UserLite      `json:"project_lead"`
	TotalIssues       int64          `json:"total_issues"`
	TotalMembers      int64          `json:"total_members"`
	LogoURL           *string        `json:"logo_url"`
	IsFavorite        bool           `json:"is_favorite"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	CreatedByID       *uint64        `json:"created_by_id"`
	UpdatedByID       *uint64        `json:"updated_by_id"`
	DeletedAt         *time.Time     `json:"deleted_at"`
	IsDeleted         bool           `json:"is_deleted"`
}

// ProjectMemberResponse represents a project member.
type ProjectMemberResponse struct {
	ID        uint64    `json:"id"`
	ProjectID uint64    `json:"project_id"`
	UserID    uint64    `json:"user_id"`
	Role      int       `json:"role"`
	IsActive  bool      `json:"is_active"`
	User      *UserLite `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProjectStatistics represents project statistics.
type ProjectStatistics struct {
	ProjectID      uint64         `json:"project_id"`
	ProjectName    string         `json:"project_name"`
	TotalIssues    int64          `json:"total_issues"`
	CompletedIssues int64         `json:"completed_issues"`
	ActiveMembers  int64          `json:"active_members"`
	States         map[string]int `json:"states"`
	Priorities     map[string]int `json:"priorities"`
}

// IssuesSummary represents issue counts by state group.
type IssuesSummary struct {
	ProjectID   uint64            `json:"project_id"`
	ProjectName string            `json:"project_name"`
	Issues      map[string]int    `json:"issues"` // todo, started, completed, cancelled
}

// ProjectSubscriberResponse represents a project subscriber.
type ProjectSubscriberResponse struct {
	ID        uint64    `json:"id"`
	ProjectID uint64    `json:"project_id"`
	UserID    uint64    `json:"user_id"`
	User      *UserLite `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
