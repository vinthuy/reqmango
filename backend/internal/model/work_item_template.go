package model

import "encoding/json"

type WorkItemTemplate struct {
	BaseModel
	Name        string          `gorm:"size:100;not null" json:"name"`
	IssueTypeID *uint64         `gorm:"index" json:"issue_type_id"`
	Defaults    json.RawMessage `gorm:"type:jsonb" json:"defaults"`
	IsDefault   bool            `gorm:"default:false" json:"is_default"`
	ProjectID   uint64          `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64          `gorm:"not null" json:"workspace_id"`
}

func (WorkItemTemplate) TableName() string {
	return "work_item_templates"
}
