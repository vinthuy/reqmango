package model

import "time"

type Module struct {
	BaseModel
	Name        string    `json:"name" gorm:"type:varchar(100);not null"`
	Description string    `json:"description" gorm:"type:text"`
	ProjectID   uint64    `json:"project_id" gorm:"not null;index"`
	WorkspaceID uint64    `json:"workspace_id" gorm:"not null;index"`
	ParentID    *uint64   `json:"parent_id" gorm:"index"`
	Order       int       `json:"order" gorm:"default:0"`
	ArchivedAt  *time.Time `json:"archived_at"`
	IsArchived  bool      `json:"is_archived" gorm:"default:false"`
}
