package model

// Permission defines a single permission (e.g., "issue:create").
type Permission struct {
	BaseModel
	Code        string `gorm:"uniqueIndex;size:100;not null" json:"code"`
	Name        string `gorm:"size:200;not null" json:"name"`
	Description string `gorm:"size:500" json:"description"`
	Resource    string `gorm:"size:100;not null" json:"resource"`
	Action      string `gorm:"size:50;not null" json:"action"`
	Scope       string `gorm:"size:20;default:project" json:"scope"`
	FieldName   string `gorm:"size:100;default:''" json:"field_name,omitempty"`
}

// RolePermission join table between Role and Permission.
type RolePermission struct {
	RoleID       uint64 `gorm:"primaryKey;index" json:"role_id"`
	PermissionID uint64 `gorm:"primaryKey;index" json:"permission_id"`
}

// FieldPermission represents a field-level permission rule.
type FieldPermission struct {
	BaseModel
	Resource     string `gorm:"size:100;not null;index" json:"resource"`
	FieldName    string `gorm:"size:100;not null" json:"field_name"`
	RoleID       uint64 `gorm:"index" json:"role_id"`
	CanRead      bool   `gorm:"default:true" json:"can_read"`
	CanWrite     bool   `gorm:"default:false" json:"can_write"`
	ProjectID    uint64 `gorm:"index" json:"project_id"`
	WorkspaceID  uint64 `gorm:"index" json:"workspace_id"`
}
