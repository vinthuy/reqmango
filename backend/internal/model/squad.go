package model

import (
	"encoding/json"
	"time"
)

// Squad represents a multi-agent collaboration team
type Squad struct {
	BaseModel
	WorkspaceID    uint64         `gorm:"not null;index" json:"workspace_id"`
	ProjectID      *uint64        `gorm:"index" json:"project_id,omitempty"`
	Name           string         `gorm:"size:100;not null" json:"name"`
	Description    string         `gorm:"type:text" json:"description"`
	LeaderAgentID  *uint64        `gorm:"index" json:"leader_agent_id,omitempty"`
	Status         string         `gorm:"size:20;not null;default:active" json:"status"`
	Config         json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"config"`
	Goal           string         `gorm:"type:text" json:"goal"`
	Members        []SquadMember  `json:"members"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      *time.Time     `gorm:"index" json:"deleted_at,omitempty"`
}

// SquadMember represents a member agent in a squad
type SquadMember struct {
	BaseModel
	SquadID     uint64     `gorm:"not null;index" json:"squad_id"`
	AgentID     uint64     `gorm:"not null;index" json:"agent_id"`
	Role        string     `gorm:"size:50;not null" json:"role"` // leader, member, observer
	AgentConfigID uint64   `gorm:"index" json:"agent_config_id"`
	Status      string     `gorm:"size:20;not null;default:active" json:"status"`
	AssignedAt  time.Time  `json:"assigned_at"`
	RemovedAt   *time.Time `json:"removed_at,omitempty"`
}

// SquadTask represents a task assigned to a squad member
type SquadTask struct {
	BaseModel
	SquadID         uint64     `gorm:"not null;index" json:"squad_id"`
	AgentTaskID     uint64     `gorm:"not null;index" json:"agent_task_id"`
	MemberID        uint64     `gorm:"not null;index" json:"member_id"`
	Status          string     `gorm:"size:20;not null;default:pending" json:"status"`
	Priority        string     `gorm:"size:20" json:"priority"`
	Progress        int        `gorm:"default:0" json:"progress"`
	TaskDescription string     `gorm:"type:text" json:"task_description"` // What the member was asked to do (decomposed subtask)
	Feedback        string     `gorm:"type:text" json:"feedback"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

// SquadExecution represents a squad's execution session
type SquadExecution struct {
	BaseModel
	SquadID       uint64         `gorm:"not null;index" json:"squad_id"`
	Status        string         `gorm:"size:20;not null;default:pending" json:"status"`
	Goal          string         `gorm:"type:text" json:"goal"`
	InputData     json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"input_data"`
	OutputData    json.RawMessage `gorm:"type:jsonb" json:"output_data"`
	Logs          json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"logs"`
	StartedAt     *time.Time     `json:"started_at,omitempty"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	FailedAt      *time.Time     `json:"failed_at,omitempty"`
	ErrorInfo     string         `json:"error_info"`
}
