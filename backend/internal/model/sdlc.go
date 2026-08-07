package model

import (
	"encoding/json"
	"time"
)

// SDLCWorkflowStatus defines the lifecycle state of an SDLC pipeline run
// (PRD P4-006: 完整 SDLC 流程编排引擎).
//
// Flow: pending → running → completed | failed | partial_failed | cancelled
type SDLCWorkflowStatus string

const (
	SDLCWorkflowPending   SDLCWorkflowStatus = "pending"
	SDLCWorkflowRunning   SDLCWorkflowStatus = "running"
	SDLCWorkflowCompleted SDLCWorkflowStatus = "completed"
	SDLCWorkflowFailed    SDLCWorkflowStatus = "failed"
	SDLCWorkflowPartial   SDLCWorkflowStatus = "partial_failed"
	SDLCWorkflowCancelled SDLCWorkflowStatus = "cancelled"
)

// SDLCStageStatus defines the lifecycle state of a single SDLC stage.
//
// Flow: pending → running → success | failed | skipped
type SDLCStageStatus string

const (
	SDLCStagePending SDLCStageStatus = "pending"
	SDLCStageRunning SDLCStageStatus = "running"
	SDLCStageSuccess SDLCStageStatus = "success"
	SDLCStageFailed  SDLCStageStatus = "failed"
	SDLCStageSkipped SDLCStageStatus = "skipped"
)

// SDLCWorkflow tracks a single end-to-end software delivery lifecycle run
// (PRD §3: 需求分析 → 需求设计 → 分派 Feature → 功能设计 → 拆解 US →
// 迭代排期 → US 开发 → 代码审查 → US 测试 → FE 测试 → 上线).
//
// A workflow is workspace-scoped and may optionally be project-scoped. It
// records the natural-language requirement, the runtime configuration (which
// stages to run, agent mappings, optional squad), and the accumulated
// artifacts produced by each stage (analysis report, PRD, issue IDs, PR URL,
// build id, …). Status updates flow through SSE as sdlc_workflow.* events so
// the UI can render live progress without polling.
type SDLCWorkflow struct {
	BaseModel

	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
	ProjectID   *uint64 `gorm:"index" json:"project_id,omitempty"`

	// SquadID optionally ties the workflow to a Squad so stages can dispatch
	// to the squad's agents (P3-004 Leader Agent integration).
	SquadID *uint64 `gorm:"index" json:"squad_id,omitempty"`

	// Title is a short human-readable label for the requirement.
	Title string `gorm:"size:200;not null" json:"title"`

	// Requirement is the natural-language requirement description that seeds
	// stage 1 (需求分析).
	Requirement string `gorm:"type:text" json:"requirement"`

	// Status: pending | running | completed | failed | partial_failed | cancelled
	Status SDLCWorkflowStatus `gorm:"size:20;not null;default:pending;index" json:"status"`

	Progress     int     `gorm:"default:0" json:"progress"`
	CurrentStage *string `gorm:"size:100" json:"current_stage,omitempty"`

	// Config: runtime configuration JSONB — selected stage keys, agent
	// mappings, options (e.g. {"stages":["requirement_analysis",...],
	// "fail_fast":true}).
	Config json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"config"`

	// Artifacts: accumulated outputs across stages (JSONB). Each stage's
	// output is merged here so downstream stages and the UI can read the
	// pipeline's evolving state (e.g. analysis_report, feature_issue_id,
	// pr_url, build_id).
	Artifacts json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"artifacts"`

	ErrorMessage *string    `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CancelledAt  *time.Time `json:"cancelled_at,omitempty"`

	// Relationships
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Project   *Project  `gorm:"foreignKey:ProjectID" json:"-"`
	Stages    []SDLCStage `gorm:"foreignKey:WorkflowID" json:"stages,omitempty"`
}

func (SDLCWorkflow) TableName() string { return "sdlc_workflows" }

// SDLCStage tracks the execution of a single stage within an SDLC workflow.
//
// Stages are created up-front (one row per canonical stage, in order) when
// the workflow starts. Each row records its canonical key/name/agent-role,
// the input snapshot, the produced output, and per-stage logs. This gives
// the UI a stable, queryable view of pipeline progress.
type SDLCStage struct {
	BaseModel

	WorkflowID  uint64 `gorm:"not null;index" json:"workflow_id"`
	WorkspaceID uint64 `gorm:"not null;index" json:"workspace_id"`

	// Order is the 1-based position of the stage in the canonical SDLC flow.
	Order int `gorm:"not null" json:"order"`

	// Key is the stable machine identifier (e.g. "requirement_analysis").
	Key string `gorm:"size:60;not null" json:"key"`

	// Name is the human-readable stage name (e.g. "需求分析").
	Name string `gorm:"size:100;not null" json:"name"`

	// AgentRole is the agent role responsible for the stage
	// (e.g. "需求分析师", "Developer", "Tester", "Leader").
	AgentRole string `gorm:"size:60" json:"agent_role"`

	// Status: pending | running | success | failed | skipped
	Status SDLCStageStatus `gorm:"size:20;not null;default:pending;index" json:"status"`

	Progress int `gorm:"default:0" json:"progress"`

	// Input: snapshot of the inputs handed to the stage (JSONB).
	Input json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"input"`

	// Output: artifacts produced by the stage (JSONB), merged into the
	// workflow's Artifacts on success.
	Output json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"output"`

	// Logs: append-only log lines for the stage (JSONB array of strings).
	Logs json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"logs"`

	ErrorMessage *string    `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	DurationMs   int64      `gorm:"default:0" json:"duration_ms"`

	// Relationships
	Workflow SDLCWorkflow `gorm:"foreignKey:WorkflowID" json:"-"`
}

func (SDLCStage) TableName() string { return "sdlc_stages" }
