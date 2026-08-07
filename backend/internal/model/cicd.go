package model

import (
	"encoding/json"
	"time"
)

// CICDProvider identifies the CI/CD platform a config targets.
//
// Supported values map to PRD §14.1:
//
//	github_actions  — GitHub Actions (Webhook + API)
//	gitlab_ci       — GitLab CI (Webhook + API)
//	jenkins         — Jenkins (API)
//	generic         — any webhook-driven pipeline (default)
type CICDProvider string

const (
	CICDProviderGitHubActions CICDProvider = "github_actions"
	CICDProviderGitLabCI      CICDProvider = "gitlab_ci"
	CICDProviderJenkins       CICDProvider = "jenkins"
	CICDProviderGeneric       CICDProvider = "generic"
)

// CICDConfig captures the connection details for a CI/CD pipeline
// (PRD P4-003: cicd_configs table).
//
// A config is workspace-scoped and may optionally be project-scoped. It
// records the provider, API endpoint, secret token reference, and the events
// that should auto-trigger a build. The token is stored as an opaque
// reference (e.g. an env-var key or secret manager handle) rather than the
// raw credential — the service resolves it at trigger time.
type CICDConfig struct {
	BaseModel

	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
	ProjectID   *uint64 `gorm:"index" json:"project_id,omitempty"`

	Name string `gorm:"size:120;not null" json:"name"`

	// Provider: github_actions | gitlab_ci | jenkins | generic
	Provider CICDProvider `gorm:"size:30;not null;default:generic" json:"provider"`

	// APIEndpoint is the provider's API base URL or webhook trigger URL.
	APIEndpoint string `gorm:"size:512" json:"api_endpoint"`

	// ProjectSlug is the provider-specific project/repo identifier
	// (e.g. "owner/repo" for GitHub, project path for GitLab, job name for Jenkins).
	ProjectSlug string `gorm:"size:255" json:"project_slug"`

	// DefaultBranch is the branch to build when none is supplied.
	DefaultBranch string `gorm:"size:120;default:main" json:"default_branch"`

	// AuthTokenRef references the stored credential. We never persist the raw
	// token — only an opaque reference (env var name, secret key, etc.).
	AuthTokenRef string `gorm:"size:255" json:"auth_token_ref,omitempty"`

	// TriggerEvents: array of event names that should auto-trigger a build
	// (e.g. ["push","pull_request","manual"]). Stored as JSONB.
	TriggerEvents json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"trigger_events"`

	// ExtraConfig: provider-specific free-form config (JSONB).
	ExtraConfig json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"extra_config"`

	Enabled bool `gorm:"default:true" json:"enabled"`

	// Relationships
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Project   *Project  `gorm:"foreignKey:ProjectID" json:"-"`
}

func (CICDConfig) TableName() string { return "cicd_configs" }

// BuildStatus defines the lifecycle state of a build (PRD P4-003 build_records).
//
// Flow: pending → queued → running → success | failed | cancelled
// Unknown is used when the provider returns an unrecognized state.
type BuildStatus string

const (
	BuildPending   BuildStatus = "pending"
	BuildQueued    BuildStatus = "queued"
	BuildRunning   BuildStatus = "running"
	BuildSuccess   BuildStatus = "success"
	BuildFailed    BuildStatus = "failed"
	BuildCancelled BuildStatus = "cancelled"
	BuildUnknown   BuildStatus = "unknown"
)

// BuildTrigger identifies what initiated a build.
type BuildTrigger string

const (
	BuildTriggerManual   BuildTrigger = "manual"
	BuildTriggerPush     BuildTrigger = "push"
	BuildTriggerPull     BuildTrigger = "pull_request"
	BuildTriggerSchedule BuildTrigger = "schedule"
	BuildTriggerAgent    BuildTrigger = "agent"
	BuildTriggerWebhook  BuildTrigger = "webhook"
)

// BuildRecord tracks a single CI/CD pipeline execution
// (PRD P4-003: build_records table).
//
// A build is created when a CICDConfig is triggered (manually or by an event).
// The service records the trigger source, branch, commit, provider build id,
// per-stage progress, and the final outcome. Status updates flow through SSE
// as cicd_build.* events so the UI can render live progress.
type BuildRecord struct {
	BaseModel

	WorkspaceID  uint64  `gorm:"not null;index" json:"workspace_id"`
	ProjectID    *uint64 `gorm:"index" json:"project_id,omitempty"`
	CICDConfigID uint64  `gorm:"not null;index" json:"cicd_config_id"`

	// Trigger: manual | push | pull_request | schedule | agent | webhook
	Trigger BuildTrigger `gorm:"size:20;not null;default:manual" json:"trigger"`

	// Source references
	Branch     string  `gorm:"size:120" json:"branch"`
	CommitSHA  string  `gorm:"size:64" json:"commit_sha"`
	IssueID    *uint64 `gorm:"index" json:"issue_id,omitempty"`
	AgentTaskID *uint64 `gorm:"index" json:"agent_task_id,omitempty"`
	TriggeredByID uint64 `gorm:"column:triggered_by_id" json:"triggered_by_id"`

	// Provider-side identifiers
	ExternalBuildID string `gorm:"size:128" json:"external_build_id"`
	BuildURL        string `gorm:"size:512" json:"build_url"`

	// Stages: array of {name, status, duration_ms, started_at, completed_at}
	Stages json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"stages"`

	// Lifecycle
	Status       BuildStatus `gorm:"size:20;default:pending;index" json:"status"`
	Progress     int         `gorm:"default:0" json:"progress"`
	CurrentStage *string     `gorm:"size:100" json:"current_stage,omitempty"`
	ErrorMessage *string     `gorm:"type:text" json:"error_message,omitempty"`

	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	// Roll-up
	DurationMs int64 `gorm:"default:0" json:"duration_ms"`

	// Relationships
	Workspace  Workspace  `gorm:"foreignKey:WorkspaceID" json:"-"`
	Project    *Project   `gorm:"foreignKey:ProjectID" json:"-"`
	CICDConfig CICDConfig `gorm:"foreignKey:CICDConfigID" json:"-"`
}

func (BuildRecord) TableName() string { return "build_records" }
