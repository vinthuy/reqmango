package harness

import (
	"fmt"
	"strings"
)

// ModelTier represents a model capability level.
type ModelTier string

const (
	TierMax    ModelTier = "max"    // Fable 5, Opus max-effort
	TierHigh   ModelTier = "high"   // Opus 4.x, Sonnet high-effort
	TierMedium ModelTier = "medium" // Sonnet, Haiku high-effort
	TierLow    ModelTier = "low"    // Haiku
)

// ModelAssignment maps a stage type to a specific model and effort.
type ModelAssignment struct {
	Model  string `json:"model"`
	Effort string `json:"effort"` // low, medium, high, xhigh, max
}

// ModelRouter determines which model and effort to use for a given stage.
type ModelRouter struct {
	defaults map[StageType]ModelAssignment
}

// NewModelRouter creates a router with sensible defaults.
func NewModelRouter() *ModelRouter {
	return &ModelRouter{
		defaults: map[StageType]ModelAssignment{
			StagePlanner:  {Model: "claude-opus-4-8", Effort: "xhigh"},
			StageExecutor: {Model: "claude-sonnet-4-6", Effort: "high"},
			StageReviewer: {Model: "claude-sonnet-4-6", Effort: "high"},
			StageJudge:    {Model: "claude-opus-4-8", Effort: "high"},
		},
	}
}

// Route returns the model assignment for a stage, considering budget constraints.
func (r *ModelRouter) Route(stage StageConfig, budgetRemaining int) ModelAssignment {
	// If stage explicitly specifies model, use it
	if stage.Model != "" && stage.Effort != "" {
		return ModelAssignment{Model: stage.Model, Effort: stage.Effort}
	}
	if stage.Model != "" {
		return ModelAssignment{Model: stage.Model, Effort: "high"}
	}

	assignment, ok := r.defaults[stage.StageType]
	if !ok {
		// Default: executor with medium effort
		return ModelAssignment{Model: "claude-sonnet-4-6", Effort: "medium"}
	}

	// Budget-based downgrade
	if budgetRemaining > 0 && budgetRemaining < 10000 {
		assignment = r.downgrade(assignment)
	}

	return assignment
}

// downgrade reduces model tier to conserve budget.
func (r *ModelRouter) downgrade(a ModelAssignment) ModelAssignment {
	model := strings.ToLower(a.Model)

	if strings.Contains(model, "opus") || strings.Contains(model, "fable") {
		return ModelAssignment{Model: "claude-sonnet-4-6", Effort: "medium"}
	}
	if strings.Contains(model, "sonnet") {
		return ModelAssignment{Model: "claude-haiku-4-5", Effort: "medium"}
	}

	// Already at minimum — reduce effort
	return ModelAssignment{Model: a.Model, Effort: "low"}
}

// EstimateTokens estimates token usage for a stage type.
func (r *ModelRouter) EstimateTokens(stageType StageType) int {
	switch stageType {
	case StagePlanner:
		return 2000
	case StageExecutor:
		return 4000
	case StageReviewer:
		return 3000
	case StageJudge:
		return 2000
	default:
		return 2000
	}
}

// GetModelTier categorizes a model assignment into a tier.
func (r *ModelRouter) GetModelTier(a ModelAssignment) ModelTier {
	model := strings.ToLower(a.Model)
	if strings.Contains(model, "fable") || (strings.Contains(model, "opus") && a.Effort == "max") {
		return TierMax
	}
	if strings.Contains(model, "opus") {
		return TierHigh
	}
	if strings.Contains(model, "sonnet") && a.Effort == "high" {
		return TierHigh
	}
	if strings.Contains(model, "sonnet") {
		return TierMedium
	}
	return TierLow
}

// Summary returns a human-readable summary of the routing decision.
func (r *ModelRouter) Summary(stage StageConfig, budgetRemaining int) string {
	assignment := r.Route(stage, budgetRemaining)
	tier := r.GetModelTier(assignment)
	return fmt.Sprintf("[%s] %s → %s/%s (tier: %s, budget: %d remaining)",
		stage.StageType, stage.Name, assignment.Model, assignment.Effort, tier, budgetRemaining)
}
