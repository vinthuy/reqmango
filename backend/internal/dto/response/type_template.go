package response

import "time"

type TypeTemplateResponse struct {
	ID           uint64                        `json:"id"`
	Name         string                        `json:"name"`
	Color        string                        `json:"color"`
	Icon         string                        `json:"icon"`
	Description  string                        `json:"description"`
	Level        int                           `json:"level"`
	ParentTypeID *uint64                       `json:"parent_type_id"`
	WorkspaceID  uint64                        `json:"workspace_id"`
	Fields       []TypeTemplateFieldResponse   `json:"fields"`
	CreatedAt    time.Time                     `json:"created_at"`
	UpdatedAt    time.Time                     `json:"updated_at"`
}

type TypeTemplateFieldResponse struct {
	TemplateTypeID uint64 `json:"template_type_id"`
	FieldID        uint64 `json:"field_id"`
	IsRequired     bool   `json:"is_required"`
	Sequence       int    `json:"sequence"`
	FieldName      string `json:"field_name,omitempty"`
	FieldType      string `json:"field_type,omitempty"`
}
