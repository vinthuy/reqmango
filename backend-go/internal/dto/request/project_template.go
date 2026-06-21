package request

type ProjectTemplateCreate struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

type ProjectTemplateUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsDefault   *bool   `json:"is_default"`
}

type ProjectTemplateAddType struct {
	TypeTemplateID uint64  `json:"type_template_id" binding:"required"`
	IsRequired     bool    `json:"is_required"`
	DefaultStateID *uint64 `json:"default_state_id"`
	Sequence       int     `json:"sequence"`
}

type ProjectTemplateApply struct {
	ProjectID uint64 `json:"project_id" binding:"required"`
}
