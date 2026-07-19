package request

type ApprovalCreate struct {
	TransitionID uint64 `json:"transition_id" binding:"required"`
	RequestNote  string `json:"request_note"`
}

type ApprovalDecision struct {
	Decision string `json:"decision" binding:"required,oneof=approved rejected"`
	Note     string `json:"note"`
}

type ApprovalListQuery struct {
	Status     string `form:"status" json:"status"`
	ProjectID  uint64 `form:"project_id" json:"project_id"`
	ApproverID uint64 `form:"approver_id" json:"approver_id"`
}
