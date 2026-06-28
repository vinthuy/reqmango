package request

type CreateInitiativeReq struct {
	WorkspaceID uint64   `json:"workspace_id"`
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description,omitempty"`
	Color       string   `json:"color,omitempty"`
	Status      string   `json:"status,omitempty"`
	TargetDate  string   `json:"target_date,omitempty"`
	StartDate   string   `json:"start_date,omitempty"`
	ProjectIDs  []uint64 `json:"project_ids,omitempty"`
}

type UpdateInitiativeReq struct {
	Name        string   `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Color       *string  `json:"color,omitempty"`
	Status      string   `json:"status,omitempty"`
	TargetDate  *string  `json:"target_date,omitempty"`
	StartDate   *string  `json:"start_date,omitempty"`
	ProjectIDs  []uint64 `json:"project_ids,omitempty"`
}
