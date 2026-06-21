package model

// IssueType represents a work item type (e.g., Bug, Task, Story, Epic).
// Workspace-scoped with optional project_id — when nil, the type is shared across all workspace projects.
type IssueType struct {
	BaseModel
	Name        string  `gorm:"type:varchar(100);not null" json:"name"`
	Color       string  `gorm:"type:varchar(20);default:'#6366F1'" json:"color"`
	Icon        string  `gorm:"type:varchar(50);default:'circle'" json:"icon"`
	IsDefault   bool    `gorm:"default:false" json:"is_default"`
	Sequence    int     `gorm:"default:1" json:"sequence"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`

	// Relationships
	Workspace    Workspace        `gorm:"foreignKey:WorkspaceID" json:"-"`
	FieldLinks   []IssueTypeField `gorm:"foreignKey:TypeID" json:"-"`
}

func (IssueType) TableName() string {
	return "issue_types"
}

// IssueTypeField is a join table for many-to-many association between IssueType and CustomField.
type IssueTypeField struct {
	TypeID     uint64 `gorm:"primaryKey;autoIncrement:false" json:"type_id"`
	FieldID    uint64 `gorm:"primaryKey;autoIncrement:false" json:"field_id"`
	IsRequired bool   `gorm:"default:false" json:"is_required"`
	Sequence   int    `gorm:"default:1" json:"sequence"`

	IssueType  IssueType   `gorm:"foreignKey:TypeID;constraint:OnDelete:CASCADE" json:"-"`
	Field      CustomField `gorm:"foreignKey:FieldID;constraint:OnDelete:CASCADE" json:"-"`
}

func (IssueTypeField) TableName() string {
	return "issue_type_fields"
}
