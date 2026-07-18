package common

// State group constants for issue lifecycle (mirrors backend/internal/common/constants.go).
const (
	StateGroupBacklog    = "backlog"
	StateGroupUnstarted  = "unstarted"
	StateGroupStarted    = "started"
	StateGroupCompleted  = "completed"
	StateGroupCancelled  = "cancelled"
)
