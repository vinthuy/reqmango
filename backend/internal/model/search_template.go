package model

import "encoding/json"

// SearchTemplate represents a preset search template for issues.
type SearchTemplate struct {
	BaseModel

	Name        string          `gorm:"size:100;not null" json:"name"`
	Description *string         `gorm:"size:255" json:"description"`
	Icon        string          `gorm:"size:50" json:"icon"`
	RQLTemplate string          `gorm:"type:text;not null" json:"rql_template"`
	ViewType    string          `gorm:"size:20;default:list" json:"view_type"`
	SortConfig  json.RawMessage `gorm:"type:jsonb" json:"sort_config"`
	GroupBy     *string         `gorm:"size:50" json:"group_by"`
	Columns     json.RawMessage `gorm:"type:jsonb" json:"columns"`

	IsBuiltIn bool    `gorm:"default:false" json:"is_built_in"`
	IsPublic  bool    `gorm:"default:true" json:"is_public"`
	OwnerID   *uint64 `gorm:"index" json:"owner_id"`
	ProjectID uint64  `gorm:"index" json:"project_id"`

	Owner   User    `gorm:"foreignKey:OwnerID" json:"-"`
	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

func (SearchTemplate) TableName() string {
	return "search_templates"
}