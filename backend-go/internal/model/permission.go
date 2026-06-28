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
}

// RolePermission join table between Role and Permission.
type RolePermission struct {
	RoleID       uint64 `gorm:"primaryKey;index" json:"role_id"`
	PermissionID uint64 `gorm:"primaryKey;index" json:"permission_id"`
}
