package model

import (
	"encoding/json"
	"time"
)

// TesterJobStatus defines the lifecycle state of a Tester Agent job.
//
// Flow: pending → generating_cases → executing → reporting → completed
// Any step may transition to "failed".
type TesterJobStatus string

const (
	TesterJobPending         TesterJobStatus = "pending"
	TesterJobGeneratingCases TesterJobStatus = "generating_cases"
	TesterJobExecuting       TesterJobStatus = "executing"
	TesterJobReporting       TesterJobStatus = "reporting"
	TesterJobCompleted       TesterJobStatus = "completed"
	TesterJobFailed          TesterJobStatus = "failed"
	TesterJobCancelled       TesterJobStatus = "cancelled"
)

// TesterJob tracks a single Tester Agent execution (PRD P4-002).
//
// The Tester Agent generates test cases from a requirement + acceptance
// criteria, executes them, and reports failures as Bug work items. Each job
// records the generated cases, execution results, and the IDs of any Bug
// issues created so the lifecycle can be observed and audited from the UI.
type TesterJob struct {
	BaseModel

	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
	ProjectID   *uint64 `gorm:"index" json:"project_id,omitempty"`
	IssueID     *uint64 `gorm:"index" json:"issue_id,omitempty"`
	AgentTaskID *uint64 `gorm:"index" json:"agent_task_id,omitempty"`

	// Inputs
	Title              string          `gorm:"size:255;not null" json:"title"`
	RequirementText    string          `gorm:"type:text" json:"requirement_text"`
	AcceptanceCriteria string          `gorm:"type:text" json:"acceptance_criteria"`
	TestScope          string          `gorm:"size:20;default:unit" json:"test_scope"` // unit | integration | e2e
	InputContext       json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"input_context"`

	// Generated test cases: array of {id, name, description, steps, expected}
	GeneratedCases json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"generated_cases"`

	// Execution results: array of {case_id, name, status, duration_ms, error}
	TestResults json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"test_results"`

	// Roll-up summary
	TotalCases int `gorm:"default:0" json:"total_cases"`
	PassCount  int `gorm:"default:0" json:"pass_count"`
	FailCount  int `gorm:"default:0" json:"fail_count"`
	SkipCount  int `gorm:"default:0" json:"skip_count"`

	// Bug issue IDs created from failed cases (linked as sub-issues of IssueID)
	BugIssueIDs json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"bug_issue_ids"`

	// Lifecycle
	Status       TesterJobStatus `gorm:"size:20;default:pending;index" json:"status"`
	Progress     int             `gorm:"default:0" json:"progress"`
	CurrentStep  *string         `gorm:"size:100" json:"current_step,omitempty"`
	ErrorMessage *string         `gorm:"type:text" json:"error_message,omitempty"`

	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	// Relationships
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Project   *Project  `gorm:"foreignKey:ProjectID" json:"-"`
	Issue     *Issue    `gorm:"foreignKey:IssueID" json:"-"`
}

func (TesterJob) TableName() string { return "tester_jobs" }
