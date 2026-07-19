package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Loop represents a saved Loop definition (YAML/JSON DSL).
type Loop struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkspaceID uint64          `gorm:"not null;index" json:"workspace_id"`
	Name        string          `gorm:"size:255;not null" json:"name"`
	Description *string         `gorm:"type:text" json:"description,omitempty"`
	LoopDef     json.RawMessage `gorm:"type:jsonb;not null" json:"loop_def"`
	Version     string          `gorm:"size:50;default:1.0" json:"version"`
	Status      string          `gorm:"size:50;default:active" json:"status"`
	CreatedByID *uint64         `json:"created_by_id,omitempty"`
	UpdatedByID *uint64         `json:"updated_by_id,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (Loop) TableName() string { return "agent_loops" }

// LoopRun represents one active/completed execution of a Loop.
type LoopRun struct {
	ID               uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	LoopID           uint64          `gorm:"not null;index" json:"loop_id"`
	Status           string          `gorm:"size:50;default:running" json:"status"`
	CurrentIteration int             `gorm:"default:0" json:"current_iteration"`
	MaxIterations    int             `gorm:"default:100" json:"max_iterations"`
	Goal             string          `gorm:"type:text;not null" json:"goal"`
	GoalMetrics      json.RawMessage `gorm:"type:jsonb" json:"goal_metrics,omitempty"`
	WorkingMemory    json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"working_memory"`
	TokensUsed       int             `gorm:"default:0" json:"tokens_used"`
	CostUSD          float64         `gorm:"type:decimal(10,4);default:0" json:"cost_usd"`
	StoppedReason    *string         `gorm:"size:100" json:"stopped_reason,omitempty"`
	StartedAt        time.Time       `gorm:"autoCreateTime" json:"started_at"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func (LoopRun) TableName() string { return "agent_loop_runs" }

// LoopIteration records each Act->Observe->Reason cycle within a LoopRun.
type LoopIteration struct {
	ID             uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	LoopRunID      uint64          `gorm:"not null;index" json:"loop_run_id"`
	IterationNum   int             `gorm:"not null" json:"iteration_num"`
	ActionTaken    json.RawMessage `gorm:"type:jsonb;not null" json:"action_taken"`
	ResultObserved json.RawMessage `gorm:"type:jsonb;not null" json:"result_observed"`
	Reasoning      *string         `gorm:"type:text" json:"reasoning,omitempty"`
	Decision       string          `gorm:"size:50;not null" json:"decision"`
	TokensUsed     int             `gorm:"default:0" json:"tokens_used"`
	DurationMs     *int            `json:"duration_ms,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (LoopIteration) TableName() string { return "agent_loop_iterations" }
