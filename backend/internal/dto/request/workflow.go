package request

type WorkflowCreate struct {
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	IssueTypeID  *uint64 `json:"issue_type_id"`
	IssueTypeIDs string  `json:"issue_type_ids"`
}

type WorkflowUpdate struct {
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	IssueTypeID  *uint64 `json:"issue_type_id"`
	IssueTypeIDs *string `json:"issue_type_ids"`
	IsActive     *bool   `json:"is_active"`
}

type TransitionCreate struct {
	Name                 string  `json:"name"`
	FromStateID          uint64  `json:"from_state_id" binding:"required"`
	ToStateID            uint64  `json:"to_state_id" binding:"required"`
	Description          string  `json:"description"`
	RuleType             string  `json:"rule_type"`
	ApproverIDs          *string `json:"approver_ids"`
	RoleAllowed          string  `json:"role_allowed"`
	ApproveTargetStateID *uint64 `json:"approve_target_state_id"`
	RejectTargetStateID  *uint64 `json:"reject_target_state_id"`
	ApprovalMode         string  `json:"approval_mode"`
}

type TransitionUpdate struct {
	Description          *string `json:"description"`
	RuleType             *string `json:"rule_type"`
	ApproverIDs          *string `json:"approver_ids"`
	RoleAllowed          *string `json:"role_allowed"`
	ApproveTargetStateID *uint64 `json:"approve_target_state_id"`
	RejectTargetStateID  *uint64 `json:"reject_target_state_id"`
	ApprovalMode         *string `json:"approval_mode"`
}

type AutomationCreate struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	TriggerType string `json:"trigger_type" binding:"required"`
	Conditions  string `json:"conditions"`
	Actions     string `json:"actions" binding:"required"`
	Sequence    int    `json:"sequence"`
}

type AutomationUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	TriggerType *string `json:"trigger_type"`
	Conditions  *string `json:"conditions"`
	Actions     *string `json:"actions"`
	IsEnabled   *bool   `json:"is_enabled"`
	Sequence    *int    `json:"sequence"`
}
