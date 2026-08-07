package model

import (
	"encoding/json"
	"time"
)

// DeveloperJobStatus defines the lifecycle state of a Developer Agent job.
//
// Flow: pending → analyzing → generating → committing → opening_pr → completed
// Any step may transition to "failed".
type DeveloperJobStatus string

const (
	DeveloperJobPending    DeveloperJobStatus = "pending"
	DeveloperJobAnalyzing  DeveloperJobStatus = "analyzing"
	DeveloperJobGenerating DeveloperJobStatus = "generating"
	DeveloperJobCommitting DeveloperJobStatus = "committing"
	DeveloperJobOpeningPR  DeveloperJobStatus = "opening_pr"
	DeveloperJobCompleted  DeveloperJobStatus = "completed"
	DeveloperJobFailed     DeveloperJobStatus = "failed"
	DeveloperJobCancelled  DeveloperJobStatus = "cancelled"
)

// DeveloperJob tracks a single Developer Agent execution (PRD P4-001).
//
// The Developer Agent analyses a requirement (issue + design doc), generates
// source code, commits it to a Git provider, and opens a Pull Request. Each
// job records the inputs, intermediate artifacts, and final PR reference so
// the lifecycle can be observed and audited from the UI.
type DeveloperJob struct {
	BaseModel

	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
	ProjectID   *uint64 `gorm:"index" json:"project_id,omitempty"`
	IssueID     *uint64 `gorm:"index" json:"issue_id,omitempty"`
	AgentTaskID *uint64 `gorm:"index" json:"agent_task_id,omitempty"`

	// Git provider connection (GitHubConnection.ID, GitIntegration.ID, etc.)
	GitProvider     string  `gorm:"size:20;default:github" json:"git_provider"`
	GitConnectionID *uint64 `gorm:"index" json:"git_connection_id,omitempty"`

	// Inputs
	Title           string          `gorm:"size:255;not null" json:"title"`
	RequirementText string          `gorm:"type:text" json:"requirement_text"`
	DesignDocURL    *string         `gorm:"size:500" json:"design_doc_url,omitempty"`
	InputContext    json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"input_context"`

	// Branch + commit metadata
	BranchName    string `gorm:"size:255" json:"branch_name"`
	BaseBranch    string `gorm:"size:255;default:main" json:"base_branch"`
	CommitMessage string `gorm:"type:text" json:"commit_message"`

	// Generated code artifacts: array of {path, content, mode}
	GeneratedFiles json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"generated_files"`

	// PR reference returned by the git provider
	PRNumber  *int    `gorm:"index" json:"pr_number,omitempty"`
	PRURL     *string `gorm:"size:500" json:"pr_url,omitempty"`
	PRTitle   *string `gorm:"size:255" json:"pr_title,omitempty"`
	CommitSHA *string `gorm:"size:64" json:"commit_sha,omitempty"`

	// Lifecycle
	Status       DeveloperJobStatus `gorm:"size:20;default:pending;index" json:"status"`
	Progress     int                `gorm:"default:0" json:"progress"`
	CurrentStep  *string            `gorm:"size:100" json:"current_step,omitempty"`
	ErrorMessage *string            `gorm:"type:text" json:"error_message,omitempty"`

	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	// Relationships
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Project   *Project  `gorm:"foreignKey:ProjectID" json:"-"`
	Issue     *Issue    `gorm:"foreignKey:IssueID" json:"-"`
}

func (DeveloperJob) TableName() string { return "developer_jobs" }
