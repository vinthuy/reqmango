package model

import (
	"time"
)

type EstimateMode string

const (
	EstimateModePoints    EstimateMode = "points"
	EstimateModeCategories EstimateMode = "categories"
	EstimateModeTime      EstimateMode = "time"
)

type EstimatePoint struct {
	BaseModel
	Name       string       `gorm:"size:100;not null" json:"name"`
	Value      int          `json:"value"`
	Mode       EstimateMode `gorm:"size:20;default:points" json:"mode"`
	IsDefault  bool         `gorm:"default:false" json:"is_default"`
	Sequence   int          `gorm:"default:1" json:"sequence"`
	ProjectID  uint64       `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64      `gorm:"not null;index" json:"workspace_id"`
}

type EstimateCategory struct {
	BaseModel
	Name       string       `gorm:"size:100;not null" json:"name"`
	Mode       EstimateMode `gorm:"size:20;default:categories" json:"mode"`
	IsDefault  bool         `gorm:"default:false" json:"is_default"`
	Sequence   int          `gorm:"default:1" json:"sequence"`
	ProjectID  uint64       `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64      `gorm:"not null;index" json:"workspace_id"`
}

type EstimateTime struct {
	BaseModel
	Name       string       `gorm:"size:100;not null" json:"name"`
	Minutes    int          `json:"minutes"`
	Mode       EstimateMode `gorm:"size:20;default:time" json:"mode"`
	IsDefault  bool         `gorm:"default:false" json:"is_default"`
	Sequence   int          `gorm:"default:1" json:"sequence"`
	ProjectID  uint64       `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64      `gorm:"not null;index" json:"workspace_id"`
}

type ProjectEstimateSettings struct {
	BaseModel
	ProjectID      uint64       `gorm:"not null;unique;index" json:"project_id"`
	WorkspaceID    uint64       `gorm:"not null;index" json:"workspace_id"`
	Mode           EstimateMode `gorm:"size:20;default:points" json:"mode"`
	PointsEnabled  bool         `gorm:"default:true" json:"points_enabled"`
	CategoriesEnabled bool      `gorm:"default:false" json:"categories_enabled"`
	TimeEnabled    bool         `gorm:"default:false" json:"time_enabled"`
}

func (e *EstimatePoint) CreatedAtTime() time.Time {
	return e.CreatedAt
}

func (e *EstimatePoint) UpdatedAtTime() time.Time {
	return e.UpdatedAt
}

func (e *EstimateCategory) CreatedAtTime() time.Time {
	return e.CreatedAt
}

func (e *EstimateCategory) UpdatedAtTime() time.Time {
	return e.UpdatedAt
}

func (e *EstimateTime) CreatedAtTime() time.Time {
	return e.CreatedAt
}

func (e *EstimateTime) UpdatedAtTime() time.Time {
	return e.UpdatedAt
}

func (e *ProjectEstimateSettings) CreatedAtTime() time.Time {
	return e.CreatedAt
}

func (e *ProjectEstimateSettings) UpdatedAtTime() time.Time {
	return e.UpdatedAt
}