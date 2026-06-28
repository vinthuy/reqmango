package response

import "time"

type ProjectTemplateResponse struct {
	ID          uint64                        `json:"id"`
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	WorkspaceID uint64                        `json:"workspace_id"`
	IsDefault   bool                          `json:"is_default"`
	States      *string                       `json:"states,omitempty"`
	Labels      *string                       `json:"labels,omitempty"`
	Types       []ProjectTemplateTypeResponse `json:"types"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`
}

type ProjectTemplateTypeResponse struct {
	TemplateID     uint64  `json:"template_id"`
	TypeTemplateID uint64  `json:"type_template_id"`
	IsRequired     bool    `json:"is_required"`
	DefaultStateID *uint64 `json:"default_state_id"`
	Sequence       int     `json:"sequence"`
	TypeName       string  `json:"type_name,omitempty"`
	TypeColor      string  `json:"type_color,omitempty"`
	TypeIcon       string  `json:"type_icon,omitempty"`
	TypeLevel      int     `json:"type_level,omitempty"`
}
