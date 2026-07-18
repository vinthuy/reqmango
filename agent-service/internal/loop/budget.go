// backend/internal/agent/loop/budget.go
package loop

import (
	"fmt"
	"sync"
	"time"
)

// BudgetController enforces hard limits on Loop execution.
type BudgetController struct {
	mu sync.Mutex

	MaxTokens     int
	UsedTokens    int
	MaxCost       float64
	UsedCost      float64
	MaxIterations int
	Iteration     int
	MaxDuration   time.Duration
	StartTime     time.Time

	// Progress tracking for stuck detection
	LastMetrics     map[string]float64
	NoProgressCount int
	MaxNoProgress   int
}

func NewBudgetController(maxTokens int, maxCost float64, maxIterations int, maxDuration time.Duration) *BudgetController {
	return &BudgetController{
		MaxTokens:      maxTokens,
		MaxCost:        maxCost,
		MaxIterations:  maxIterations,
		MaxDuration:    maxDuration,
		StartTime:      time.Now(),
		LastMetrics:    make(map[string]float64),
		MaxNoProgress:  3,
	}
}

// CanContinue checks all budget constraints.
func (b *BudgetController) CanContinue() (bool, string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.MaxTokens > 0 && b.UsedTokens >= b.MaxTokens {
		return false, "token budget exhausted"
	}
	if b.MaxCost > 0 && b.UsedCost >= b.MaxCost {
		return false, "cost budget exhausted"
	}
	if b.MaxIterations > 0 && b.Iteration >= b.MaxIterations {
		return false, "max iterations reached"
	}
	if b.MaxDuration > 0 && time.Since(b.StartTime) >= b.MaxDuration {
		return false, "max duration exceeded"
	}
	return true, ""
}

// RecordIteration increments iteration counter and returns whether stuck.
func (b *BudgetController) RecordIteration(tokens int, cost float64, metrics map[string]float64) (bool, string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Iteration++
	b.UsedTokens += tokens
	b.UsedCost += cost

	improved := false
	for k, v := range metrics {
		if last, ok := b.LastMetrics[k]; ok {
			if v > last {
				improved = true
				break
			}
		}
	}
	if improved || len(b.LastMetrics) == 0 {
		b.NoProgressCount = 0
	} else {
		b.NoProgressCount++
	}

	b.LastMetrics = metrics
	if b.NoProgressCount >= b.MaxNoProgress {
		return true, fmt.Sprintf("no progress for %d consecutive iterations", b.NoProgressCount)
	}
	return false, ""
}

// RecordTokens adds token usage without incrementing iteration.
func (b *BudgetController) RecordTokens(tokens int, cost float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.UsedTokens += tokens
	b.UsedCost += cost
}

// Summary returns a human-readable budget summary.
func (b *BudgetController) Summary() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := fmt.Sprintf("iter %d", b.Iteration)
	if b.MaxIterations > 0 {
		result += fmt.Sprintf("/%d", b.MaxIterations)
	}
	if b.MaxTokens > 0 {
		result += fmt.Sprintf(" | tokens %d/%d", b.UsedTokens, b.MaxTokens)
	}
	if b.MaxCost > 0 {
		result += fmt.Sprintf(" | cost $%.4f/$%.2f", b.UsedCost, b.MaxCost)
	}
	if b.MaxDuration > 0 {
		result += fmt.Sprintf(" | elapsed %s/%s",
			time.Since(b.StartTime).Round(time.Second),
			b.MaxDuration.Round(time.Second))
	}
	return result
}
