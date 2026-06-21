package response

import "time"

type ModuleResponse struct {
	ID          uint64     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ProjectID   uint64     `json:"project_id"`
	WorkspaceID uint64     `json:"workspace_id"`
	ParentID    *uint64    `json:"parent_id"`
	Order       int        `json:"order"`
	IsArchived  bool       `json:"is_archived"`
	ArchivedAt  *time.Time `json:"archived_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ModuleTreeNode struct {
	ModuleResponse
	Children []*ModuleTreeNode `json:"children"`
}

type ModuleProgress struct {
	ModuleID     uint64 `json:"module_id"`
	ModuleName   string `json:"module_name"`
	TotalIssues  int64  `json:"total_issues"`
	Completed    int64  `json:"completed"`
	Progress     int    `json:"progress"`
}

type ModuleStatistics struct {
	ModuleID      uint64            `json:"module_id"`
	ModuleName    string            `json:"module_name"`
	TotalIssues   int64             `json:"total_issues"`
	ActiveIssues  int64             `json:"active_issues"`
	Completed     int64             `json:"completed"`
	Cancelled     int64             `json:"cancelled"`
	ByPriority    map[string]int64  `json:"by_priority"`
	ByState       map[string]int64  `json:"by_state"`
}
