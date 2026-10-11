package response

import "time"

// AgentPerformanceOverview aggregates workspace-wide agent task execution metrics.
type AgentPerformanceOverview struct {
	WorkspaceID        uint64     `json:"workspace_id"`
	TotalTasks         int64      `json:"total_tasks"`
	CompletedTasks     int64      `json:"completed_tasks"`
	FailedTasks        int64      `json:"failed_tasks"`
	CancelledTasks     int64      `json:"cancelled_tasks"`
	RunningTasks       int64      `json:"running_tasks"`
	PendingTasks       int64      `json:"pending_tasks"`
	SuccessRate        float64    `json:"success_rate"`         // 0-100, completed/(completed+failed)
	AvgDurationSeconds float64    `json:"avg_duration_seconds"` // mean of completed tasks
	TotalDurationSec   float64    `json:"total_duration_seconds"`
	PeriodStart        *time.Time `json:"period_start,omitempty"`
	PeriodEnd          *time.Time `json:"period_end,omitempty"`
}

// TemplatePerformance groups metrics per agent template.
type TemplatePerformance struct {
	AgentTemplateID    *uint64    `json:"agent_template_id,omitempty"`
	TemplateName       string     `json:"template_name"`
	TaskType           string     `json:"task_type,omitempty"`
	TotalTasks         int64      `json:"total_tasks"`
	CompletedTasks     int64      `json:"completed_tasks"`
	FailedTasks        int64      `json:"failed_tasks"`
	CancelledTasks     int64      `json:"cancelled_tasks"`
	SuccessRate        float64    `json:"success_rate"`
	AvgDurationSeconds float64    `json:"avg_duration_seconds"`
	LastRunAt          *time.Time `json:"last_run_at,omitempty"`
}

// TimelinePoint represents an aggregated metrics bucket over a time range.
type TimelinePoint struct {
	BucketStart        time.Time `json:"bucket_start"`
	BucketEnd          time.Time `json:"bucket_end"`
	TotalTasks         int64     `json:"total_tasks"`
	CompletedTasks     int64     `json:"completed_tasks"`
	FailedTasks        int64     `json:"failed_tasks"`
	SuccessRate        float64   `json:"success_rate"`
	AvgDurationSeconds float64   `json:"avg_duration_seconds"`
}

// SuggestionCounts tallies what users did with agent proposals.
type SuggestionCounts struct {
	Proposed  int64 `json:"proposed"` // actionable proposals (keep-as-is excluded)
	Applied   int64 `json:"applied"`
	Dismissed int64 `json:"dismissed"`
	Failed    int64 `json:"failed"` // the user tried to apply but it was refused
	Pending   int64 `json:"pending"`
	// AcceptanceRate is applied/(applied+dismissed) in 0-100; refusals and
	// undecided proposals say nothing about the agent's judgement.
	AcceptanceRate float64 `json:"acceptance_rate"`
}

// AgentSuggestionStats groups suggestion outcomes per agent.
type AgentSuggestionStats struct {
	AgentID   uint64 `json:"agent_id"`
	AgentName string `json:"agent_name"`
	SuggestionCounts
	ByField map[string]SuggestionCounts `json:"by_field"`
}

// FailureBreakdown tallies failure reasons for diagnostic dashboards.
type FailureBreakdown struct {
	FailureReason string  `json:"failure_reason"`
	Count         int64   `json:"count"`
	Percentage    float64 `json:"percentage"` // share of total failures
}
