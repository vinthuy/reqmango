package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// IssueType represents a work item type (e.g., Bug, Task, Story, Epic).
// Workspace-scoped with optional project_id — when nil, the type is shared across all workspace projects.
// Supports hierarchical levels (L0-L5) via Level + ParentTypeID fields.
type IssueType struct {
	BaseModel
	Name               string          `gorm:"type:varchar(100);not null" json:"name"`
	Color              string          `gorm:"type:varchar(20);default:'#6366F1'" json:"color"`
	Icon               string          `gorm:"type:varchar(50);default:'circle'" json:"icon"`
	Description        string          `gorm:"type:text" json:"description"`
	Level              int             `gorm:"default:0" json:"level"`           // L0-L5, 0=top-level (Epic)
	ParentTypeID       *uint64         `gorm:"index" json:"parent_type_id"`    // parent type in hierarchy
	AllowedChildTypeIDs JSONUint64Array `gorm:"type:jsonb" json:"allowed_child_type_ids"` // [2, 3, 5] — allowed child type IDs
	IsDefault          bool            `gorm:"default:false" json:"is_default"`
	Sequence           int             `gorm:"default:1" json:"sequence"`
	IsActive           bool            `gorm:"default:true" json:"is_active"`
	ProjectID          *uint64         `gorm:"index" json:"project_id"`
	WorkspaceID        uint64          `gorm:"not null;index" json:"workspace_id"`

	// Relationships
	Workspace  Workspace        `gorm:"foreignKey:WorkspaceID" json:"-"`
	ParentType *IssueType       `gorm:"foreignKey:ParentTypeID" json:"-"`
	ChildTypes []IssueType      `gorm:"foreignKey:ParentTypeID" json:"-"`
	FieldLinks []IssueTypeField `gorm:"foreignKey:TypeID" json:"-"`
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

	IssueType IssueType   `gorm:"foreignKey:TypeID;constraint:OnDelete:CASCADE" json:"-"`
	Field     CustomField `gorm:"foreignKey:FieldID;constraint:OnDelete:CASCADE" json:"-"`
}

func (IssueTypeField) TableName() string {
	return "issue_type_fields"
}

type JSONUint64Array []uint64

func (j JSONUint64Array) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONUint64Array) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	data, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONUint64Array: expected []byte, got %T", value)
	}
	return json.Unmarshal(data, j)
}
