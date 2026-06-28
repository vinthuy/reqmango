package common

// Role constants for workspace and project members.
const (
	RoleGuest  = 5
	RoleMember = 15
	RoleAdmin  = 20
)

// Priority constants for issues.
const (
	PriorityUrgent = "urgent"
	PriorityHigh   = "high"
	PriorityMedium = "medium"
	PriorityLow    = "low"
	PriorityNone   = "none"
)

// State group constants.
const (
	StateGroupBacklog   = "backlog"
	StateGroupUnstarted = "unstarted"
	StateGroupStarted   = "started"
	StateGroupCompleted = "completed"
	StateGroupCancelled = "cancelled"
)

// DefaultStates defines the 6 default states created for each new project.
var DefaultStates = []struct {
	Name      string
	Color     string
	Group     string
	Sequence  int
	IsDefault bool
}{
	{Name: "Backlog", Color: "#6B7280", Group: StateGroupBacklog, Sequence: 1, IsDefault: true},
	{Name: "Todo", Color: "#3B82F6", Group: StateGroupUnstarted, Sequence: 2, IsDefault: false},
	{Name: "In Progress", Color: "#F59E0B", Group: StateGroupStarted, Sequence: 3, IsDefault: false},
	{Name: "In Review", Color: "#8B5CF6", Group: StateGroupStarted, Sequence: 4, IsDefault: false},
	{Name: "Done", Color: "#10B981", Group: StateGroupCompleted, Sequence: 5, IsDefault: false},
	{Name: "Cancelled", Color: "#EF4444", Group: StateGroupCancelled, Sequence: 6, IsDefault: false},
}
