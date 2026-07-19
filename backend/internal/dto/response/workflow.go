package response

import "time"

type WorkflowResponse struct {
	ID           uint64                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	ProjectID    uint64                 `json:"project_id"`
	WorkspaceID  uint64                 `json:"workspace_id"`
	IssueTypeID  *uint64                `json:"issue_type_id"`
	IssueTypeIDs string                 `json:"issue_type_ids"`
	IsActive     bool                   `json:"is_active"`
	IsInherited  bool                   `json:"is_inherited"`
	Transitions  []TransitionResponse   `json:"transitions"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type TransitionResponse struct {
	ID          uint64 `json:"id"`
	WorkflowID  uint64 `json:"workflow_id"`
	SourceStateID uint64 `json:"from_state_id"`
	TargetStateID   uint64 `json:"to_state_id"`
	Description string `json:"description"`
	RuleType    string `json:"rule_type"`
	ApproverIDs *string `json:"approver_ids"`
	RoleAllowed string `json:"role_allowed"`
	SourceName    string `json:"from_name,omitempty"`
	TargetName      string `json:"to_name,omitempty"`
}

type AutomationResponse struct {
	ID             uint64    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ProjectID      uint64    `json:"project_id"`
	IsEnabled      bool      `json:"is_enabled"`
	Sequence       int       `json:"sequence"`
	ExecutionCount int       `json:"execution_count"`
	TriggerType    string    `json:"trigger_type"`
	Conditions     string    `json:"conditions"`
	Actions        string    `json:"actions"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
