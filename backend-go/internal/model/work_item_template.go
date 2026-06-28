package model

import "encoding/json"

type WorkItemTemplate struct {
	BaseModel

	Name         string          `gorm:"size:100;not null" json:"name"`
	Description  string          `gorm:"size:500" json:"description"`
	IssueTypeID  *uint64         `gorm:"index" json:"issue_type_id"`
	DefaultsJSON json.RawMessage `gorm:"type:jsonb;column:defaults" json:"defaults"`
	IsDefault    bool            `gorm:"default:false" json:"is_default"`
	ProjectID    uint64          `gorm:"not null;index" json:"project_id"`
	WorkspaceID  uint64          `gorm:"not null;index" json:"workspace_id"`

	IssueType *IssueType `gorm:"foreignKey:IssueTypeID" json:"issue_type,omitempty"`
}

func (WorkItemTemplate) TableName() string {
	return "work_item_templates"
}
