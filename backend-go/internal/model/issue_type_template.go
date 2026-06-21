package model

// IssueTypeTemplate defines a work item type blueprint at workspace level.
// When applied to a project, it creates a concrete IssueType.
type IssueTypeTemplate struct {
	BaseModel
	Name         string  `gorm:"type:varchar(100);not null" json:"name"`
	Color        string  `gorm:"type:varchar(20);default:'#6366F1'" json:"color"`
	Icon         string  `gorm:"type:varchar(50);default:'circle'" json:"icon"`
	Description  string  `gorm:"type:text" json:"description"`
	Level        int     `gorm:"default:0" json:"level"`
	ParentTypeID *uint64 `gorm:"index" json:"parent_type_id"` // parent in template hierarchy
	WorkspaceID  uint64  `gorm:"not null;index" json:"workspace_id"`

	// Relationships
	ParentType *IssueTypeTemplate       `gorm:"foreignKey:ParentTypeID" json:"-"`
	ChildTypes []IssueTypeTemplate      `gorm:"foreignKey:ParentTypeID" json:"-"`
	FieldLinks []IssueTypeTemplateField `gorm:"foreignKey:TemplateTypeID" json:"-"`
}

func (IssueTypeTemplate) TableName() string {
	return "issue_type_templates"
}

// IssueTypeTemplateField binds a workspace custom field to a type template.
type IssueTypeTemplateField struct {
	TemplateTypeID uint64 `gorm:"primaryKey;autoIncrement:false" json:"template_type_id"`
	FieldID        uint64 `gorm:"primaryKey;autoIncrement:false" json:"field_id"`
	IsRequired     bool   `gorm:"default:false" json:"is_required"`
	Sequence       int    `gorm:"default:1" json:"sequence"`

	Template  IssueTypeTemplate `gorm:"foreignKey:TemplateTypeID;constraint:OnDelete:CASCADE" json:"-"`
	Field     CustomField       `gorm:"foreignKey:FieldID;constraint:OnDelete:CASCADE" json:"-"`
}

func (IssueTypeTemplateField) TableName() string {
	return "issue_type_template_fields"
}
