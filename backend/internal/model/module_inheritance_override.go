package model

import "time"

type ModuleInheritanceOverride struct {
	BaseModel
	ProjectID         uint64     `gorm:"not null;index" json:"project_id"`
	WorkspaceModuleID uint64     `gorm:"not null;index" json:"workspace_module_id"`
	IsExcluded        bool       `gorm:"default:false" json:"is_excluded"`
	OverrideName      *string    `gorm:"type:varchar(100)" json:"override_name"`
	OverrideDescription *string  `gorm:"type:text" json:"override_description"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (ModuleInheritanceOverride) TableName() string {
	return "module_inheritance_overrides"
}