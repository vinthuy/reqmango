// backend/internal/agent/loop/budget_test.go
package loop

import (
	"testing"
	"time"
)

func TestBudgetTokenLimit(t *testing.T) {
	b := NewBudgetController(1000, 0, 0, 0)
	b.RecordTokens(900, 0)

	ok, reason := b.CanContinue()
	if !ok {
		t.Fatalf("expected continue, got: %s", reason)
	}

	b.RecordTokens(200, 0)
	ok, reason = b.CanContinue()
	if ok {
		t.Fatal("expected budget exhausted")
	}
	if reason != "token budget exhausted" {
		t.Fatalf("unexpected reason: %s", reason)
	}
}

func TestBudgetIterationLimit(t *testing.T) {
	b := NewBudgetController(0, 0, 5, 0)
	for i := 0; i < 5; i++ {
		b.RecordIteration(100, 0, map[string]float64{"progress": float64(i)})
	}
	ok, _ := b.CanContinue()
	if ok {
		t.Fatal("expected max iterations reached")
	}
}

func TestStuckDetection(t *testing.T) {
	b := NewBudgetController(0, 0, 0, 0)

	isStuck, _ := b.RecordIteration(100, 0, map[string]float64{"p": 0.5})
	if isStuck {
		t.Fatal("should not be stuck on first iteration")
	}

	isStuck, _ = b.RecordIteration(100, 0, map[string]float64{"p": 0.5})
	if isStuck {
		t.Fatal("should not be stuck after 2 same metrics")
	}

	isStuck, _ = b.RecordIteration(100, 0, map[string]float64{"p": 0.5})
	if isStuck {
		t.Fatal("should not be stuck after 3 same metrics")
	}

	// 4th consecutive no-progress should trigger stuck
	isStuck, msg := b.RecordIteration(100, 0, map[string]float64{"p": 0.5})
	if !isStuck {
		t.Fatal("should detect stuck after 3+ no-progress iterations")
	}
	if msg == "" {
		t.Fatal("expected stuck message")
	}
}

func TestBudgetDurationExceeded(t *testing.T) {
	b := NewBudgetController(0, 0, 0, 1*time.Nanosecond)
	time.Sleep(10 * time.Millisecond)
	ok, reason := b.CanContinue()
	if ok {
		t.Fatal("expected duration exceeded")
	}
	if reason != "max duration exceeded" {
		t.Fatalf("unexpected reason: %s", reason)
	}
}
