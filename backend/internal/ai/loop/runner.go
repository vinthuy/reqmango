package loop

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentExecutor is the interface LoopRunner uses to execute agent actions.
type AgentExecutor interface {
	Execute(ctx context.Context, task string, context map[string]interface{}) (result string, tokensUsed int, cost float64, err error)
}

// MetricsCollector collects observable metrics after an action.
type MetricsCollector interface {
	Collect(ctx context.Context, context map[string]interface{}) (map[string]float64, error)
}

// GoalEvaluator checks whether the collected metrics satisfy the loop's goal.
type GoalEvaluator interface {
	Evaluate(goal string, metrics map[string]float64) (achieved bool, reason string)
}

// LoopConfig holds configuration for a single LoopRun execution.
type LoopConfig struct {
	Goal          string
	MaxIterations int
	MaxTokens     int
	MaxCost       float64
	MaxDuration   time.Duration
	Context       map[string]interface{}
}

// LoopRunner orchestrates the execution of a single LoopRun.
type LoopRunner struct {
	db           *gorm.DB
	stateMachine *StateMachine
	budget       *BudgetController
	executor     AgentExecutor
	collector    MetricsCollector
	evaluator    GoalEvaluator
	config       LoopConfig
	iterations   []IterationRecord
	sessionID    string
}

// IterationRecord stores the result of one loop iteration.
type IterationRecord struct {
	Num        int
	Action     string
	Result     string
	Metrics    map[string]float64
	Decision   Decision
	Reasoning  string
	TokensUsed int
	Cost       float64
	DurationMs int
}

// NewLoopRunner creates a new LoopRunner.
func NewLoopRunner(db *gorm.DB, executor AgentExecutor, collector MetricsCollector, evaluator GoalEvaluator, config LoopConfig) *LoopRunner {
	return &LoopRunner{
		db:           db,
		stateMachine: NewStateMachine(),
		budget:       NewBudgetController(config.MaxTokens, config.MaxCost, config.MaxIterations, config.MaxDuration),
		executor:     executor,
		collector:    collector,
		evaluator:    evaluator,
		config:       config,
		sessionID:    uuid.New().String(),
	}
}

// SessionID returns the unique session identifier.
func (r *LoopRunner) SessionID() string {
	return r.sessionID
}

// Budget returns the budget controller for external inspection.
func (r *LoopRunner) Budget() *BudgetController {
	return r.budget
}

// Run executes the loop until a stopping condition is met.
func (r *LoopRunner) Run(ctx context.Context) ([]IterationRecord, string, error) {
	// IDLE -> PLANNING
	if err := r.stateMachine.Transition(StatePlanning); err != nil {
		return nil, "", fmt.Errorf("loop start failed: %w", err)
	}

	for !r.stateMachine.IsTerminal() {
		// Check budget before each iteration
		if ok, reason := r.budget.CanContinue(); !ok {
			r.stateMachine.Transition(StateFailed)
			return r.iterations, fmt.Sprintf("budget_exhausted: %s", reason), nil
		}

		select {
		case <-ctx.Done():
			r.stateMachine.Transition(StateFailed)
			return r.iterations, "cancelled", ctx.Err()
		default:
		}

		// ACTING: Execute the action
		if err := r.stateMachine.Transition(StateActing); err != nil {
			return nil, "", err
		}

		start := time.Now()
		task := r.buildTask()
		result, tokens, cost, err := r.executor.Execute(ctx, task, r.config.Context)
		duration := int(time.Since(start).Milliseconds())

		if err != nil {
			log.Printf("[LoopRunner] executor error: %v", err)
			r.stateMachine.Transition(StateFailed)
			return r.iterations, "executor_error", err
		}

		r.budget.RecordTokens(tokens, cost)

		// OBSERVING
		if err := r.stateMachine.Transition(StateObserving); err != nil {
			return nil, "", err
		}

		metrics, err := r.collector.Collect(ctx, r.config.Context)
		if err != nil {
			log.Printf("[LoopRunner] collector error: %v", err)
			metrics = map[string]float64{}
		}

		// REASONING
		if err := r.stateMachine.Transition(StateReasoning); err != nil {
			return nil, "", err
		}

		achieved, evalReason := r.evaluator.Evaluate(r.config.Goal, metrics)

		var decision Decision
		var reasoning string
		if achieved {
			decision = DecideStop
			reasoning = fmt.Sprintf("goal achieved: %s", evalReason)
		} else {
			isStuck, stuckMsg := r.budget.RecordIteration(tokens, cost, metrics)
			if isStuck {
				decision = DecideEscalate
				reasoning = fmt.Sprintf("stuck: %s", stuckMsg)
			} else {
				decision = DecideContinue
				reasoning = fmt.Sprintf("not yet achieved: %s", evalReason)
			}
		}

		iter := IterationRecord{
			Num:        r.budget.Iteration,
			Action:     task,
			Result:     result,
			Metrics:    metrics,
			Decision:   decision,
			Reasoning:  reasoning,
			TokensUsed: tokens,
			Cost:       cost,
			DurationMs: duration,
		}
		r.iterations = append(r.iterations, iter)

		switch decision {
		case DecideStop:
			r.stateMachine.Transition(StateCompleted)
			return r.iterations, "goal_achieved", nil
		case DecideEscalate:
			r.stateMachine.Transition(StateFailed)
			return r.iterations, fmt.Sprintf("escalated: %s", reasoning), nil
		case DecideWait:
			r.stateMachine.Transition(StateWaiting)
			time.Sleep(30 * time.Second)
			r.stateMachine.Transition(StateActing)
		default:
			// DecideContinue: loop back to acting
			r.stateMachine.Transition(StateActing)
		}
	}

	return r.iterations, string(r.stateMachine.Current()), nil
}

func (r *LoopRunner) buildTask() string {
	if len(r.iterations) == 0 {
		return fmt.Sprintf("Initial execution for goal: %s", r.config.Goal)
	}
	last := r.iterations[len(r.iterations)-1]
	return fmt.Sprintf(
		"Continue working toward goal: %s\nPrevious result: %s\nPrevious metrics: %v\nPrevious reasoning: %s",
		r.config.Goal, last.Result, last.Metrics, last.Reasoning,
	)
}
