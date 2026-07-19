package harness

import (
	"strings"
	"testing"
)

func TestNewModelRouter(t *testing.T) {
	r := NewModelRouter()
	if r == nil {
		t.Fatal("NewModelRouter() should not return nil")
	}
	if len(r.defaults) != 4 {
		t.Errorf("expected 4 default stage types, got %d", len(r.defaults))
	}
}

func TestModelRouter_Route_ExplicitModel(t *testing.T) {
	r := NewModelRouter()
	stage := StageConfig{
		Name:      "custom",
		StageType: StageExecutor,
		Model:     "gpt-4",
		Effort:    "max",
	}
	result := r.Route(stage, 50000)
	if result.Model != "gpt-4" {
		t.Errorf("Model = %q, want 'gpt-4'", result.Model)
	}
	if result.Effort != "max" {
		t.Errorf("Effort = %q, want 'max'", result.Effort)
	}
}

func TestModelRouter_Route_ExplicitModelNoEffort(t *testing.T) {
	r := NewModelRouter()
	stage := StageConfig{
		Name:      "custom",
		StageType: StageExecutor,
		Model:     "custom-model",
	}
	result := r.Route(stage, 50000)
	if result.Model != "custom-model" {
		t.Errorf("Model = %q, want 'custom-model'", result.Model)
	}
	if result.Effort != "high" {
		t.Errorf("Effort = %q, want 'high' (default)", result.Effort)
	}
}

func TestModelRouter_Route_DefaultPlanner(t *testing.T) {
	r := NewModelRouter()
	stage := StageConfig{
		Name:      "plan",
		StageType: StagePlanner,
	}
	result := r.Route(stage, 50000)
	if !strings.Contains(strings.ToLower(result.Model), "opus") {
		t.Errorf("planner should use opus, got %s", result.Model)
	}
	if result.Effort != "xhigh" {
		t.Errorf("planner effort = %q, want 'xhigh'", result.Effort)
	}
}

func TestModelRouter_Route_DefaultExecutor(t *testing.T) {
	r := NewModelRouter()
	stage := StageConfig{
		Name:      "execute",
		StageType: StageExecutor,
	}
	result := r.Route(stage, 50000)
	if !strings.Contains(strings.ToLower(result.Model), "sonnet") {
		t.Errorf("executor should use sonnet, got %s", result.Model)
	}
	if result.Effort != "high" {
		t.Errorf("executor effort = %q, want 'high'", result.Effort)
	}
}

func TestModelRouter_Route_DefaultReviewer(t *testing.T) {
	r := NewModelRouter()
	stage := StageConfig{
		Name:      "review",
		StageType: StageReviewer,
	}
	result := r.Route(stage, 50000)
	if !strings.Contains(strings.ToLower(result.Model), "sonnet") {
		t.Errorf("reviewer should use sonnet, got %s", result.Model)
	}
}

func TestModelRouter_Route_DefaultJudge(t *testing.T) {
	r := NewModelRouter()
	stage := StageConfig{
		Name:      "judge",
		StageType: StageJudge,
	}
	result := r.Route(stage, 50000)
	if !strings.Contains(strings.ToLower(result.Model), "opus") {
		t.Errorf("judge should use opus, got %s", result.Model)
	}
}

func TestModelRouter_Route_BudgetDowngrade(t *testing.T) {
	r := NewModelRouter()
	stage := StageConfig{
		Name:      "plan",
		StageType: StagePlanner,
	}
	// Low budget should trigger downgrade
	result := r.Route(stage, 5000)
	// Opus should downgrade to Sonnet
	if strings.Contains(strings.ToLower(result.Model), "opus") {
		t.Error("opus should be downgraded under low budget")
	}
	if result.Effort == "xhigh" {
		t.Error("effort should be reduced under low budget")
	}
}

func TestModelRouter_Route_BudgetDowngrade_Sonnet(t *testing.T) {
	r := NewModelRouter()
	stage := StageConfig{
		Name:      "execute",
		StageType: StageExecutor,
	}
	// Low budget should downgrade Sonnet to Haiku
	result := r.Route(stage, 5000)
	if strings.Contains(strings.ToLower(result.Model), "sonnet") {
		t.Error("sonnet should be downgraded under low budget")
	}
}

func TestModelRouter_Route_UnknownStageType(t *testing.T) {
	r := NewModelRouter()
	stage := StageConfig{
		Name:      "unknown",
		StageType: "custom_role", // not in defaults
	}
	result := r.Route(stage, 50000)
	if result.Model != "claude-sonnet-4-6" {
		t.Errorf("unknown stage type should default to claude-sonnet-4-6, got %s", result.Model)
	}
}

func TestModelRouter_EstimateTokens(t *testing.T) {
	r := NewModelRouter()
	tests := []struct {
		stageType StageType
		wantMin  int
	}{
		{StagePlanner, 2000},
		{StageExecutor, 4000},
		{StageReviewer, 3000},
		{StageJudge, 2000},
		{StageType("unknown"), 2000},
	}
	for _, tt := range tests {
		t.Run(string(tt.stageType), func(t *testing.T) {
			got := r.EstimateTokens(tt.stageType)
			if got != tt.wantMin {
				t.Errorf("EstimateTokens(%q) = %d, want %d", tt.stageType, got, tt.wantMin)
			}
		})
	}
}

func TestModelRouter_GetModelTier(t *testing.T) {
	r := NewModelRouter()
	tests := []struct {
		model  string
		effort string
		want   ModelTier
	}{
		{"claude-opus-4-8", "max", TierMax},
		{"fable-5", "high", TierMax},
		{"claude-opus-4-8", "high", TierHigh},
		{"claude-sonnet-4-6", "high", TierHigh},
		{"claude-sonnet-4-6", "medium", TierMedium},
		{"claude-haiku-4-5", "high", TierLow},
		{"claude-haiku-4-5", "low", TierLow},
	}
	for _, tt := range tests {
		t.Run(tt.model+"/"+tt.effort, func(t *testing.T) {
			got := r.GetModelTier(ModelAssignment{Model: tt.model, Effort: tt.effort})
			if got != tt.want {
				t.Errorf("GetModelTier() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestModelRouter_Summary(t *testing.T) {
	r := NewModelRouter()
	stage := StageConfig{
		Name:      "planner",
		StageType: StagePlanner,
	}
	summary := r.Summary(stage, 50000)
	if len(summary) == 0 {
		t.Error("summary should not be empty")
	}
	if !strings.Contains(summary, "planner") {
		t.Errorf("summary should mention stage name. Got: %s", summary)
	}
}
