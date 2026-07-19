package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Pipeline represents a saved pipeline definition.
type Pipeline struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkspaceID uint64          `gorm:"not null;index" json:"workspace_id"`
	Name        string          `gorm:"size:255;not null" json:"name"`
	Description *string         `gorm:"type:text" json:"description,omitempty"`
	PipelineDef json.RawMessage `gorm:"type:jsonb;not null" json:"pipeline_def"`
	Version     string          `gorm:"size:50;default:1.0" json:"version"`
	Status      string          `gorm:"size:50;default:active" json:"status"`
	CreatedByID *uint64         `json:"created_by_id,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (Pipeline) TableName() string { return "agent_pipelines" }

// PipelineRun represents one execution of a pipeline.
type PipelineRun struct {
	ID             uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	PipelineID     uint64          `gorm:"not null;index" json:"pipeline_id"`
	TriggerType    string          `gorm:"size:50" json:"trigger_type"`
	TriggerContext json.RawMessage `gorm:"type:jsonb" json:"trigger_context,omitempty"`
	Status         string          `gorm:"size:50;default:pending" json:"status"`
	StagesResult   json.RawMessage `gorm:"type:jsonb" json:"stages_result,omitempty"`
	TokensUsed     int             `gorm:"default:0" json:"tokens_used"`
	CostUSD        float64         `gorm:"type:decimal(10,4);default:0" json:"cost_usd"`
	StartedAt      *time.Time      `json:"started_at,omitempty"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	ErrorMessage   *string         `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (PipelineRun) TableName() string { return "agent_pipeline_runs" }
