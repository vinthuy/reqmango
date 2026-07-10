package model

// CustomField represents a custom field definition.
// Workspace-scoped with optional project_id — when nil, the field is shared across all workspace projects.
// Supported field types: text, number, dropdown, boolean, date, member, url.
type CustomField struct {
	BaseModel
	Name         string  `gorm:"type:varchar(100);not null" json:"name"`
	Description  string  `gorm:"type:text" json:"description"`
	FieldType    string  `gorm:"type:varchar(20);not null" json:"field_type"`
	IsRequired   bool    `gorm:"default:false" json:"is_required"`
	DefaultValue string  `gorm:"type:text" json:"default_value"`
	Placeholder  string  `gorm:"type:varchar(255)" json:"placeholder"`
	IsActive     bool    `gorm:"default:true" json:"is_active"`
	ProjectID    *uint64 `gorm:"index" json:"project_id"`
	WorkspaceID  uint64  `gorm:"not null;index" json:"workspace_id"`

	// Relationships
	Workspace Workspace           `gorm:"foreignKey:WorkspaceID" json:"-"`
	Options   []CustomFieldOption `gorm:"foreignKey:FieldID" json:"-"`
	TypeLinks []IssueTypeField    `gorm:"foreignKey:FieldID" json:"-"`
}

func (CustomField) TableName() string {
	return "custom_fields"
}

// CustomFieldOption represents a dropdown/select option for a custom field.
type CustomFieldOption struct {
	BaseModel
	FieldID  uint64 `gorm:"not null;index" json:"field_id"`
	Value    string `gorm:"type:varchar(255);not null" json:"value"`
	Color    string `gorm:"type:varchar(20)" json:"color"`
	Sequence int    `gorm:"default:1" json:"sequence"`

	Field CustomField `gorm:"foreignKey:FieldID;constraint:OnDelete:CASCADE" json:"-"`
}

func (CustomFieldOption) TableName() string {
	return "custom_field_options"
}

// IssueCustomFieldValue stores a single custom field value for a specific issue.
type IssueCustomFieldValue struct {
	BaseModel
	IssueID uint64 `gorm:"not null;index" json:"issue_id"`
	FieldID uint64 `gorm:"not null;index" json:"field_id"`
	Value   string `gorm:"type:text" json:"value"`

	Issue Issue       `gorm:"foreignKey:IssueID" json:"-"`
	Field CustomField `gorm:"foreignKey:FieldID;constraint:OnDelete:CASCADE" json:"-"`
}

func (IssueCustomFieldValue) TableName() string {
	return "issue_custom_field_values"
}
