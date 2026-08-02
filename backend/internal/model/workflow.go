package model

import (
	"encoding/json"
	"time"
)

// Workflow represents a state machine workflow for issue transitions.
type Workflow struct {
	BaseModel

	Name        string  `gorm:"size:100;not null" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`

	// Relationships
	Project     Project          `gorm:"foreignKey:ProjectID" json:"-"`
	Transitions []StateTransition `gorm:"foreignKey:WorkflowID" json:"transitions,omitempty"`
}

func (Workflow) TableName() string {
	return "workflows"
}

// AgentWorkflow represents a workflow definition for multi-agent orchestration.
type AgentWorkflow struct {
	BaseModel

	Name          string          `gorm:"size:100;not null" json:"name"`
	Description   string          `gorm:"type:text" json:"description"`
	ProjectID     uint64          `gorm:"not null;index" json:"project_id"`
	WorkspaceID   uint64          `gorm:"not null;index" json:"workspace_id"`
	Version       int             `gorm:"default:1" json:"version"`
	IsActive      bool            `gorm:"default:true" json:"is_active"`
	TriggerType   string          `gorm:"size:20;default:manual" json:"trigger_type"` // manual|event|cron|webhook
	TriggerConfig json.RawMessage `gorm:"type:jsonb" json:"trigger_config"`
	Config        json.RawMessage `gorm:"type:jsonb" json:"config"`

	// Relationships
	Nodes []WorkflowNode `gorm:"foreignKey:WorkflowID" json:"nodes,omitempty"`
	Edges []WorkflowEdge `gorm:"foreignKey:WorkflowID" json:"edges,omitempty"`
}

func (AgentWorkflow) TableName() string {
	return "agent_workflows"
}

// WorkflowNode represents a node in a workflow (typically an agent task).
type WorkflowNode struct {
	BaseModel

	WorkflowID    uint64          `gorm:"not null;index" json:"workflow_id"`
	AgentID       uint64          `gorm:"not null" json:"agent_id"`
	NodeType      string          `gorm:"size:20;default:agent" json:"node_type"` // agent|condition|parallel|loop|gate
	Name          string          `gorm:"size:100;not null" json:"name"`
	Config        json.RawMessage `gorm:"type:jsonb" json:"config"`
	ContextConfig json.RawMessage `gorm:"type:jsonb" json:"context_config"`
	SortOrder     int             `gorm:"default:0" json:"sort_order"`
	Timeout       int             `gorm:"default:1800" json:"timeout"` // seconds
	RetryPolicy   string          `gorm:"size:20;default:retry" json:"retry_policy"` // retry|skip|abort
	MaxRetries    int             `gorm:"default:3" json:"max_retries"`
}

func (WorkflowNode) TableName() string {
	return "workflow_nodes"
}

// WorkflowEdge represents a connection between two nodes in a workflow.
type WorkflowEdge struct {
	BaseModel

	WorkflowID     uint64          `gorm:"not null;index" json:"workflow_id"`
	SourceNodeID   uint64          `gorm:"not null" json:"source_node_id"`
	TargetNodeID   uint64          `gorm:"not null" json:"target_node_id"`
	Condition      string          `gorm:"type:text" json:"condition"` // condition expression
	ContextMapping json.RawMessage `gorm:"type:jsonb" json:"context_mapping"`

	// Relationships
	SourceNode WorkflowNode `gorm:"foreignKey:SourceNodeID" json:"-"`
	TargetNode WorkflowNode `gorm:"foreignKey:TargetNodeID" json:"-"`
}

func (WorkflowEdge) TableName() string {
	return "workflow_edges"
}

// WorkflowRun represents an execution instance of a workflow.
type WorkflowRun struct {
	BaseModel

	WorkflowID  uint64          `gorm:"not null;index" json:"workflow_id"`
	IssueID     *uint64         `gorm:"index" json:"issue_id"`
	Status      string          `gorm:"size:20;default:pending" json:"status"` // pending|running|completed|failed|cancelled
	CurrentNode *uint64         `gorm:"index" json:"current_node"`
	Context     json.RawMessage `gorm:"type:jsonb" json:"context"` // global context
	StartedAt   *time.Time      `json:"started_at"`
	CompletedAt *time.Time      `json:"completed_at"`
	TotalTokens int             `json:"total_tokens"`
	TotalCost   float64         `json:"total_cost"`
	ErrorInfo   string          `gorm:"type:text" json:"error_info"`

	// Relationships
	Workflow AgentWorkflow  `gorm:"foreignKey:WorkflowID" json:"-"`
	Issue    *Issue         `gorm:"foreignKey:IssueID" json:"-"`
	NodeRuns []WorkflowNodeRun `gorm:"foreignKey:WorkflowRunID" json:"node_runs,omitempty"`
}

func (WorkflowRun) TableName() string {
	return "workflow_runs"
}

// WorkflowNodeRun represents the execution of a single node in a workflow.
type WorkflowNodeRun struct {
	BaseModel

	WorkflowRunID uint64          `gorm:"not null;index" json:"workflow_run_id"`
	NodeID        uint64          `gorm:"not null" json:"node_id"`
	AgentID       uint64          `gorm:"not null" json:"agent_id"`
	AgentTaskID   *uint64         `gorm:"index" json:"agent_task_id"`
	Status        string          `gorm:"size:20;default:pending" json:"status"` // pending|running|completed|failed|skipped
	InputContext  json.RawMessage `gorm:"type:jsonb" json:"input_context"`
	OutputContext json.RawMessage `gorm:"type:jsonb" json:"output_context"`
	StartedAt     *time.Time      `json:"started_at"`
	CompletedAt   *time.Time      `json:"completed_at"`
	TokensUsed    int             `json:"tokens_used"`
	Cost          float64         `json:"cost"`
	ErrorInfo     string          `gorm:"type:text" json:"error_info"`
	RetryCount    int             `gorm:"default:0" json:"retry_count"`

	// Relationships
	WorkflowRun WorkflowRun `gorm:"foreignKey:WorkflowRunID" json:"-"`
	Node        WorkflowNode `gorm:"foreignKey:NodeID" json:"-"`
	AgentTask   *AgentTask   `gorm:"foreignKey:AgentTaskID" json:"-"`
}

func (WorkflowNodeRun) TableName() string {
	return "workflow_node_runs"
}
