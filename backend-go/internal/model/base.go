package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel provides common fields for all models.
// Uses uint64 primary key with autoincrement for PostgreSQL BIGINT compatibility.
type BaseModel struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	CreatedByID *uint64        `gorm:"column:created_by_id" json:"created_by_id"`
	UpdatedByID *uint64        `gorm:"column:updated_by_id" json:"updated_by_id"`
}
