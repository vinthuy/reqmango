package model

// ProjectTemplate defines a reusable set of issue types, states, and settings for a project.
type ProjectTemplate struct {
	BaseModel
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	WorkspaceID uint64 `gorm:"not null;index" json:"workspace_id"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`

	// Relations
	TypeLinks []ProjectTemplateType `gorm:"foreignKey:TemplateID" json:"-"`
}

func (ProjectTemplate) TableName() string {
	return "project_templates"
}

// ProjectTemplateType links an IssueType to a ProjectTemplate with additional config.
type ProjectTemplateType struct {
	TemplateID     uint64  `gorm:"primaryKey;autoIncrement:false" json:"template_id"`
	IssueTypeID    uint64  `gorm:"primaryKey;autoIncrement:false" json:"issue_type_id"`
	IsRequired     bool    `gorm:"default:false" json:"is_required"`
	DefaultStateID *uint64 `json:"default_state_id"`
	Sequence       int     `gorm:"default:1" json:"sequence"`

	Template  ProjectTemplate `gorm:"foreignKey:TemplateID;constraint:OnDelete:CASCADE" json:"-"`
	IssueType IssueType       `gorm:"foreignKey:IssueTypeID;constraint:OnDelete:CASCADE" json:"-"`
}

func (ProjectTemplateType) TableName() string {
	return "project_template_types"
}
