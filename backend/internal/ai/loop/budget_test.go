package loop

import (
	"testing"
	"time"
)

func TestNewBudgetController(t *testing.T) {
	bc := NewBudgetController(10000, 5.0, 10, 30*time.Second)
	if bc.MaxTokens != 10000 {
		t.Errorf("MaxTokens = %d, want 10000", bc.MaxTokens)
	}
	if bc.MaxCost != 5.0 {
		t.Errorf("MaxCost = %f, want 5.0", bc.MaxCost)
	}
	if bc.MaxIterations != 10 {
		t.Errorf("MaxIterations = %d, want 10", bc.MaxIterations)
	}
	if bc.MaxDuration != 30*time.Second {
		t.Errorf("MaxDuration = %v, want 30s", bc.MaxDuration)
	}
	if bc.Iteration != 0 {
		t.Errorf("initial Iteration = %d, want 0", bc.Iteration)
	}
}

func TestBudgetController_CanContinue_Initial(t *testing.T) {
	bc := NewBudgetController(10000, 5.0, 10, 30*time.Second)
	ok, reason := bc.CanContinue()
	if !ok {
		t.Errorf("should be able to continue, got reason: %s", reason)
	}
}

func TestBudgetController_CanContinue_TokenExhausted(t *testing.T) {
	bc := NewBudgetController(1000, 0, 0, 0)
	bc.UsedTokens = 1000
	ok, reason := bc.CanContinue()
	if ok {
		t.Error("should not be able to continue when tokens exhausted")
	}
	if reason != "token budget exhausted" {
		t.Errorf("reason = %q, want 'token budget exhausted'", reason)
	}
}

func TestBudgetController_CanContinue_CostExhausted(t *testing.T) {
	bc := NewBudgetController(0, 1.0, 0, 0)
	bc.UsedCost = 1.0
	ok, reason := bc.CanContinue()
	if ok {
		t.Error("should not be able to continue when cost exhausted")
	}
	if reason != "cost budget exhausted" {
		t.Errorf("reason = %q, want 'cost budget exhausted'", reason)
	}
}

func TestBudgetController_CanContinue_MaxIterations(t *testing.T) {
	bc := NewBudgetController(0, 0, 5, 0)
	bc.Iteration = 5
	ok, reason := bc.CanContinue()
	if ok {
		t.Error("should not be able to continue when max iterations reached")
	}
	if reason != "max iterations reached" {
		t.Errorf("reason = %q, want 'max iterations reached'", reason)
	}
}

func TestBudgetController_CanContinue_DurationExceeded(t *testing.T) {
	bc := NewBudgetController(0, 0, 0, 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	ok, reason := bc.CanContinue()
	if ok {
		t.Error("should not be able to continue when duration exceeded")
	}
	if reason != "max duration exceeded" {
		t.Errorf("reason = %q, want 'max duration exceeded'", reason)
	}
}

func TestBudgetController_CanContinue_ZeroLimits(t *testing.T) {
	// Zero limits means unlimited
	bc := NewBudgetController(0, 0, 0, 0)
	bc.UsedTokens = 999999
	bc.UsedCost = 999999
	bc.Iteration = 999999
	ok, _ := bc.CanContinue()
	if !ok {
		t.Error("should be able to continue when all limits are zero (unlimited)")
	}
}

func TestBudgetController_RecordIteration(t *testing.T) {
	bc := NewBudgetController(10000, 5.0, 10, 30*time.Second)
	stuck, reason := bc.RecordIteration(500, 0.25, map[string]float64{"progress": 0.5})
	if stuck {
		t.Errorf("should not be stuck, got reason: %s", reason)
	}
	if bc.Iteration != 1 {
		t.Errorf("Iteration = %d, want 1", bc.Iteration)
	}
	if bc.UsedTokens != 500 {
		t.Errorf("UsedTokens = %d, want 500", bc.UsedTokens)
	}
	if bc.UsedCost != 0.25 {
		t.Errorf("UsedCost = %f, want 0.25", bc.UsedCost)
	}
}

func TestBudgetController_RecordIteration_ProgressDetection(t *testing.T) {
	bc := NewBudgetController(10000, 5.0, 10, 30*time.Second)

	// First iteration: no progress detection (empty LastMetrics)
	stuck, _ := bc.RecordIteration(100, 0.05, map[string]float64{"score": 0.3})
	if stuck {
		t.Error("first iteration should not be stuck")
	}

	// Second iteration: improved metric
	stuck, _ = bc.RecordIteration(100, 0.05, map[string]float64{"score": 0.5})
	if stuck {
		t.Error("improved metric should not trigger stuck")
	}
	if bc.NoProgressCount != 0 {
		t.Errorf("NoProgressCount = %d, want 0", bc.NoProgressCount)
	}

	// Third iteration: same metric (no progress)
	stuck, _ = bc.RecordIteration(100, 0.05, map[string]float64{"score": 0.5})
	if stuck {
		t.Error("first no-progress should not trigger stuck")
	}
	if bc.NoProgressCount != 1 {
		t.Errorf("NoProgressCount = %d, want 1", bc.NoProgressCount)
	}
}

func TestBudgetController_RecordIteration_Stuck(t *testing.T) {
	bc := NewBudgetController(10000, 5.0, 10, 30*time.Second)

	// Seed with initial metrics
	bc.LastMetrics = map[string]float64{"score": 0.5}

	// 3 consecutive no-progress iterations
	for i := 1; i <= 3; i++ {
		stuck, reason := bc.RecordIteration(100, 0.05, map[string]float64{"score": 0.5})
		if i < 3 && stuck {
			t.Errorf("iteration %d: should not be stuck yet", i)
		}
		if i == 3 {
			if !stuck {
				t.Error("should be stuck after 3 no-progress iterations")
			}
			if reason == "" {
				t.Error("stuck reason should not be empty")
			}
		}
	}
}

func TestBudgetController_RecordTokens(t *testing.T) {
	bc := NewBudgetController(10000, 5.0, 10, 30*time.Second)
	bc.RecordTokens(200, 0.10)
	if bc.UsedTokens != 200 {
		t.Errorf("UsedTokens = %d, want 200", bc.UsedTokens)
	}
	if bc.UsedCost != 0.10 {
		t.Errorf("UsedCost = %f, want 0.10", bc.UsedCost)
	}
	// Iteration should NOT be incremented
	if bc.Iteration != 0 {
		t.Errorf("Iteration = %d, want 0 (should not change on RecordTokens)", bc.Iteration)
	}
}

func TestBudgetController_Summary(t *testing.T) {
	bc := NewBudgetController(10000, 5.0, 10, 30*time.Second)
	bc.Iteration = 3
	bc.UsedTokens = 2500
	bc.UsedCost = 1.25

	summary := bc.Summary()
	if len(summary) == 0 {
		t.Error("summary should not be empty")
	}
	// Check for expected substrings
	expectedParts := []string{"iter 3/10", "tokens 2500/10000", "cost $1.2500/$5.00"}
	for _, part := range expectedParts {
		if !contains(summary, part) {
			t.Errorf("summary missing %q. Got: %s", part, summary)
		}
	}
}

func TestBudgetController_Summary_Unlimited(t *testing.T) {
	bc := NewBudgetController(0, 0, 0, 0)
	summary := bc.Summary()
	if !contains(summary, "iter 0") {
		t.Errorf("summary should show iteration. Got: %s", summary)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
