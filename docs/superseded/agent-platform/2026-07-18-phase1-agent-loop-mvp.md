# Phase 1: Agent Loop MVP — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the core Loop Engine + Sprint Agent Loop flagship scenario + Agent Session observability panel.

**Architecture:** New `backend/internal/agent/loop/` package with state machine, runner, budget controller, and trigger system. New `agent_loop_*` tables in PostgreSQL. Loop API handlers following existing Gin handler patterns. Frontend components in `frontend/src/views/agents/` and `frontend/src/components/agents/` following Vue 3 Composition API patterns.

**Tech Stack:** Go 1.21+ + Gin + GORM + PostgreSQL | Vue 3 + TypeScript + Pinia + TailwindCSS

**Prerequisites:** FR-01 (`dispatch_agent` in AutomationRule) is already implemented. `AICopilot.vue` already exists with Agent mode tab.

---

## File Structure

```
BACKEND (New)
backend/internal/
├── agent/                              # Agent Platform package
│   ├── loop/
│   │   ├── statemachine.go             # Loop state machine (IDLE→PLANNING→ACTING→OBSERVING→REASONING→...)
│   │   ├── runner.go                   # LoopRunner: main execution loop
│   │   ├── budget.go                   # BudgetController: token/cost/time/iteration limits
│   │   ├── trigger.go                  # Trigger system: event/cron/webhook/manual/chained
│   │   ├── observer.go                 # Metric collection + goal evaluation
│   │   └── loop_test.go               # Unit tests for state machine + runner
│   └── model/
│       ├── loop.go                     # GORM models: Loop, LoopRun, LoopIteration
│       └── session.go                  # GORM model: AgentSession
├── handler/
│   ├── agent_loop.go                   # Loop CRUD + start/stop + runs API
│   └── agent_session.go               # Agent Session list + detail API
└── service/
    └── agent_loop_service.go           # LoopService: business logic bridging handler → loop engine

BACKEND (Modify)
backend/internal/
├── model/agent.go                      # Add Agent.LoopConfig field
├── router/router.go                    # Register new loop/session routes + auto-migrate new models
├── service/automation_service.go       # (Already done: dispatch_agent handler)
└── service/agent_service.go            # Add RunLoop/StopLoop helper methods

FRONTEND (New)
frontend/src/
├── views/agents/
│   ├── AgentDashboard.vue              # Agent overview dashboard
│   ├── LoopList.vue                    # Loop list page
│   ├── LoopRunDetail.vue               # Loop run detail + iteration history
│   └── AgentSessions.vue               # Agent session history list
├── components/agents/
│   ├── LoopStateBadge.vue              # Badge showing loop status (running/completed/failed)
│   ├── LoopIterationTimeline.vue       # Timeline of loop iterations
│   ├── BudgetGauge.vue                 # Visual budget usage gauge
│   └── SessionTimeline.vue             # Agent session timeline component
├── api/
│   ├── agent-loop.ts                   # Loop API client
│   └── agent-session.ts               # Session API client
└── stores/
    └── agentLoop.ts                    # Pinia store for loop state

FRONTEND (Modify)
frontend/src/
├── router/index.ts                     # Add agent sub-routes
└── components/AICopilot.vue            # (Already has agent mode — no changes needed in Phase 1)
```

---

### Task 1: Loop Data Models

**Files:**
- Create: `backend/internal/agent/model/loop.go`
- Create: `backend/internal/agent/model/session.go`

- [ ] **Step 1: Create the loop models file**

```go
// backend/internal/agent/model/loop.go
package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Loop represents a saved Loop definition (YAML/JSON DSL).
type Loop struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkspaceID uint64          `gorm:"not null;index" json:"workspace_id"`
	Name        string          `gorm:"size:255;not null" json:"name"`
	Description *string         `gorm:"type:text" json:"description,omitempty"`
	LoopDef     json.RawMessage `gorm:"type:jsonb;not null" json:"loop_def"` // YAML→JSON DSL
	Version     string          `gorm:"size:50;default:1.0" json:"version"`
	Status      string          `gorm:"size:50;default:active" json:"status"` // active/draft/archived
	CreatedByID *uint64         `json:"created_by_id,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (Loop) TableName() string { return "agent_loops" }

// LoopRun represents one active/completed execution of a Loop.
type LoopRun struct {
	ID               uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	LoopID           uint64          `gorm:"not null;index" json:"loop_id"`
	Status           string          `gorm:"size:50;default:running" json:"status"` // running/completed/failed/escalated/stopped
	CurrentIteration int             `gorm:"default:0" json:"current_iteration"`
	MaxIterations    int             `gorm:"default:100" json:"max_iterations"`
	Goal             string          `gorm:"type:text;not null" json:"goal"`
	GoalMetrics      json.RawMessage `gorm:"type:jsonb" json:"goal_metrics,omitempty"`
	WorkingMemory    json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"working_memory"`
	TokensUsed       int             `gorm:"default:0" json:"tokens_used"`
	CostUSD          float64         `gorm:"type:decimal(10,4);default:0" json:"cost_usd"`
	StoppedReason    *string         `gorm:"size:100" json:"stopped_reason,omitempty"` // goal_achieved/budget_exhausted/stuck/escalated/manual
	StartedAt        time.Time       `gorm:"autoCreateTime" json:"started_at"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func (LoopRun) TableName() string { return "agent_loop_runs" }

// LoopIteration records each Act→Observe→Reason cycle within a LoopRun.
type LoopIteration struct {
	ID             uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	LoopRunID      uint64          `gorm:"not null;index" json:"loop_run_id"`
	IterationNum   int             `gorm:"not null" json:"iteration_num"`
	ActionTaken    json.RawMessage `gorm:"type:jsonb;not null" json:"action_taken"`
	ResultObserved json.RawMessage `gorm:"type:jsonb;not null" json:"result_observed"`
	Reasoning      *string         `gorm:"type:text" json:"reasoning,omitempty"`
	Decision       string          `gorm:"size:50;not null" json:"decision"` // continue/stop/escalate/wait
	TokensUsed     int             `gorm:"default:0" json:"tokens_used"`
	DurationMs     *int            `json:"duration_ms,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (LoopIteration) TableName() string { return "agent_loop_iterations" }
```

- [ ] **Step 2: Create the session model file**

```go
// backend/internal/agent/model/session.go
package model

import (
	"encoding/json"
	"time"
)

// AgentSession provides unified observability across all agent activities
// (pipeline stages, loop iterations, standalone dispatches).
type AgentSession struct {
	ID            string          `gorm:"primaryKey;size:64" json:"id"` // UUID
	WorkspaceID   uint64          `gorm:"not null;index" json:"workspace_id"`
	AgentType     string          `gorm:"size:50;not null" json:"agent_type"`       // loop_iteration/pipeline_stage/standalone_dispatch
	AgentRef      *string         `gorm:"size:255" json:"agent_ref,omitempty"`     // loop_id:run_id:iter or pipeline_id:stage
	Status        string          `gorm:"size:50;default:running" json:"status"`   // running/completed/failed
	ModelUsed     *string         `gorm:"size:100" json:"model_used,omitempty"`
	InputSummary  *string         `gorm:"type:text" json:"input_summary,omitempty"`
	OutputSummary *string         `gorm:"type:text" json:"output_summary,omitempty"`
	TokensInput   int             `gorm:"default:0" json:"tokens_input"`
	TokensOutput  int             `gorm:"default:0" json:"tokens_output"`
	CostUSD       float64         `gorm:"type:decimal(10,4);default:0" json:"cost_usd"`
	ToolsCalled   json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"tools_called"`
	ErrorMessage  *string         `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt     time.Time       `gorm:"autoCreateTime" json:"started_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	Metadata      json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"metadata"`
}

func (AgentSession) TableName() string { return "agent_sessions" }
```

- [ ] **Step 3: Create the package directory and verify compilation**

Run:
```bash
mkdir -p backend/internal/agent/model
```

Then verify the package compiles (it will fail until we add deps; that's expected for now):
```bash
cd backend && go build ./internal/agent/model/ 2>&1 || true
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/agent/model/loop.go backend/internal/agent/model/session.go
git commit -m "feat(agent): add Loop, LoopRun, LoopIteration, and AgentSession data models"
```

---

### Task 2: Loop State Machine

**Files:**
- Create: `backend/internal/agent/loop/statemachine.go`
- Create: `backend/internal/agent/loop/statemachine_test.go`

- [ ] **Step 1: Write the state machine**

```go
// backend/internal/agent/loop/statemachine.go
package loop

import (
	"fmt"
)

// LoopState represents the current phase of a Loop execution.
type LoopState string

const (
	StateIdle      LoopState = "idle"
	StatePlanning  LoopState = "planning"
	StateActing    LoopState = "acting"
	StateObserving LoopState = "observing"
	StateReasoning LoopState = "reasoning"
	StateWaiting   LoopState = "waiting"
	StateCompleted LoopState = "completed"
	StateFailed    LoopState = "failed"
)

// Decision is the outcome of the reasoning phase.
type Decision string

const (
	DecideContinue  Decision = "continue"
	DecideStop      Decision = "stop"
	DecideEscalate  Decision = "escalate"
	DecideWait      Decision = "wait"
)

// Transition defines a valid state transition.
type Transition struct {
	From LoopState
	To   LoopState
}

// AllowedTransitions is the complete state transition map.
var AllowedTransitions = map[LoopState][]LoopState{
	StateIdle:      {StatePlanning},
	StatePlanning:  {StateActing, StateFailed},
	StateActing:    {StateObserving, StateFailed},
	StateObserving: {StateReasoning, StateFailed},
	StateReasoning: {StateActing, StateWaiting, StateCompleted, StateFailed},
	StateWaiting:   {StateActing, StateCompleted, StateFailed},
	StateCompleted: {}, // terminal
	StateFailed:    {}, // terminal
}

// StateMachine manages deterministic Loop state transitions.
type StateMachine struct {
	current LoopState
}

func NewStateMachine() *StateMachine {
	return &StateMachine{current: StateIdle}
}

func (sm *StateMachine) Current() LoopState {
	return sm.current
}

func (sm *StateMachine) IsTerminal() bool {
	return sm.current == StateCompleted || sm.current == StateFailed
}

// Transition attempts to move to the target state. Returns error if invalid.
func (sm *StateMachine) Transition(to LoopState) error {
	allowed, ok := AllowedTransitions[sm.current]
	if !ok {
		return fmt.Errorf("unknown current state: %s", sm.current)
	}
	for _, s := range allowed {
		if s == to {
			sm.current = to
			return nil
		}
	}
	return fmt.Errorf("invalid transition: %s → %s", sm.current, to)
}
```

- [ ] **Step 2: Write the test file**

```go
// backend/internal/agent/loop/statemachine_test.go
package loop

import (
	"testing"
)

func TestStateMachineHappyPath(t *testing.T) {
	sm := NewStateMachine()
	if sm.Current() != StateIdle {
		t.Fatalf("expected idle, got %s", sm.Current())
	}

	steps := []struct {
		to  LoopState
		err bool
	}{
		{StatePlanning, false},
		{StateActing, false},
		{StateObserving, false},
		{StateReasoning, false},
		{StateActing, false},  // loop back
		{StateObserving, false},
		{StateReasoning, false},
		{StateCompleted, false},
	}

	for i, step := range steps {
		err := sm.Transition(step.to)
		if step.err && err == nil {
			t.Fatalf("step %d: expected error for %s→%s", i, sm.Current(), step.to)
		}
		if !step.err && err != nil {
			t.Fatalf("step %d: unexpected error: %v", i, err)
		}
	}

	if !sm.IsTerminal() {
		t.Fatal("expected terminal state")
	}
}

func TestStateMachineInvalidTransition(t *testing.T) {
	sm := NewStateMachine()
	// Cannot go from idle directly to completed
	err := sm.Transition(StateCompleted)
	if err == nil {
		t.Fatal("expected error for idle→completed")
	}
}

func TestStateMachineTerminalNoTransition(t *testing.T) {
	sm := NewStateMachine()
	sm.Transition(StatePlanning)
	sm.Transition(StateActing)
	sm.Transition(StateObserving)
	sm.Transition(StateReasoning)
	sm.Transition(StateFailed)

	if !sm.IsTerminal() {
		t.Fatal("expected terminal after failed")
	}

	err := sm.Transition(StateActing)
	if err == nil {
		t.Fatal("expected error transitioning from terminal state")
	}
}
```

- [ ] **Step 3: Run tests**

```bash
cd backend && go test ./internal/agent/loop/ -v -run TestStateMachine
```
Expected: 3 tests PASS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/agent/loop/statemachine.go backend/internal/agent/loop/statemachine_test.go
git commit -m "feat(agent): add Loop state machine with valid transitions"
```

---

### Task 3: Budget Controller

**Files:**
- Create: `backend/internal/agent/loop/budget.go`
- Create: `backend/internal/agent/loop/budget_test.go`

- [ ] **Step 1: Write the budget controller**

```go
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

	MaxTokens     int     // 0 = unlimited
	UsedTokens    int
	MaxCost       float64 // 0 = unlimited
	UsedCost      float64
	MaxIterations int     // 0 = unlimited
	Iteration     int
	MaxDuration   time.Duration // 0 = unlimited
	StartTime     time.Time

	// Progress tracking for stuck detection
	LastMetrics    map[string]float64
	NoProgressCount int
	MaxNoProgress  int // consecutive iterations without improvement before escalation
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

	// Stuck detection: compare metrics against last iteration
	improved := false
	for k, v := range metrics {
		if last, ok := b.LastMetrics[k]; ok {
			if v > last { // assumes higher = better; caller should normalize
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

	parts := []string{fmt.Sprintf("iter %d", b.Iteration)}
	if b.MaxIterations > 0 {
		parts = append(parts, fmt.Sprintf("/%d", b.MaxIterations))
	}
	if b.MaxTokens > 0 {
		parts = append(parts, fmt.Sprintf(" | tokens %d/%d", b.UsedTokens, b.MaxTokens))
	}
	if b.MaxCost > 0 {
		parts = append(parts, fmt.Sprintf(" | cost $%.4f/$%.2f", b.UsedCost, b.MaxCost))
	}
	if b.MaxDuration > 0 {
		parts = append(parts, fmt.Sprintf(" | elapsed %s/%s",
			time.Since(b.StartTime).Round(time.Second),
			b.MaxDuration.Round(time.Second)))
	}
	result := ""
	for _, p := range parts {
		result += p
	}
	return result
}
```

- [ ] **Step 2: Write the test file**

```go
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
		t.Fatal("should not be stuck after 3 same metrics — still checking")
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
```

- [ ] **Step 3: Run tests**

```bash
cd backend && go test ./internal/agent/loop/ -v -run TestBudget
```
Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/agent/loop/budget.go backend/internal/agent/loop/budget_test.go
git commit -m "feat(agent): add BudgetController with token/cost/iteration/duration limits and stuck detection"
```

---

### Task 4: Loop Runner

**Files:**
- Create: `backend/internal/agent/loop/runner.go`

- [ ] **Step 1: Write the LoopRunner**

```go
// backend/internal/agent/loop/runner.go
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
// Implementations call the actual LLM via AgentService/AIService.
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
	Context       map[string]interface{} // initial context passed to each iteration
}

// LoopRunner orchestrates the execution of a single LoopRun.
type LoopRunner struct {
	db             *gorm.DB
	stateMachine   *StateMachine
	budget         *BudgetController
	executor       AgentExecutor
	collector      MetricsCollector
	evaluator      GoalEvaluator
	config         LoopConfig
	iterations     []IterationRecord
	sessionID      string
}

// IterationRecord stores the result of one loop iteration.
type IterationRecord struct {
	Num            int
	Action         string
	Result         string
	Metrics        map[string]float64
	Decision       Decision
	Reasoning      string
	TokensUsed     int
	Cost           float64
	DurationMs     int
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
	// IDLE → PLANNING
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

		// PLANNING/ACTING: Execute the action
		if r.stateMachine.Current() == StatePlanning || r.stateMachine.Current() == StateActing {
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
				// In Phase 1, waiting is treated as continue after a delay
				time.Sleep(30 * time.Second)
				r.stateMachine.Transition(StateActing)
			default:
				// DecideContinue: loop back to acting
				r.stateMachine.Transition(StateActing)
			}
		}
	}

	return r.iterations, r.stateMachine.Current(), nil
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go mod tidy && go build ./internal/agent/loop/ 2>&1
```
Expected: compile succeeds (may need `go get github.com/google/uuid`)

- [ ] **Step 3: Commit**

```bash
git add backend/internal/agent/loop/runner.go
git commit -m "feat(agent): add LoopRunner with Act→Observe→Reason→Repeat execution"
```

---

### Task 5: Loop Service (Business Logic Layer)

**Files:**
- Create: `backend/internal/service/agent_loop_service.go`

- [ ] **Step 1: Write the LoopService**

```go
// backend/internal/service/agent_loop_service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	agentloop "github.com/reqmango/backend/internal/agent/loop"
	agentmodel "github.com/reqmango/backend/internal/agent/model"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// LoopService manages Loop CRUD and execution, bridging the HTTP handler
// to the Loop Engine.
type LoopService struct {
	db       *gorm.DB
	agentSvc *AgentService
}

func NewLoopService(db *gorm.DB, agentSvc *AgentService) *LoopService {
	return &LoopService{db: db, agentSvc: agentSvc}
}

// CreateLoop persists a new Loop definition.
func (s *LoopService) CreateLoop(workspaceID, userID uint64, name, description string, loopDef json.RawMessage) (*agentmodel.Loop, error) {
	loop := &agentmodel.Loop{
		WorkspaceID: workspaceID,
		Name:        name,
		LoopDef:     loopDef,
		CreatedByID: &userID,
		Status:      "active",
	}
	if description != "" {
		loop.Description = &description
	}
	if err := s.db.Create(loop).Error; err != nil {
		return nil, common.Internal(fmt.Sprintf("failed to create loop: %v", err))
	}
	return loop, nil
}

// ListLoops returns all active Loops for a workspace.
func (s *LoopService) ListLoops(workspaceID uint64) ([]agentmodel.Loop, error) {
	var loops []agentmodel.Loop
	if err := s.db.Where("workspace_id = ? AND status != ?", workspaceID, "archived").
		Order("created_at DESC").Find(&loops).Error; err != nil {
		return nil, common.Internal(fmt.Sprintf("failed to list loops: %v", err))
	}
	return loops, nil
}

// GetLoop retrieves a single Loop by ID.
func (s *LoopService) GetLoop(workspaceID, loopID uint64) (*agentmodel.Loop, error) {
	var loop agentmodel.Loop
	if err := s.db.Where("id = ? AND workspace_id = ?", loopID, workspaceID).First(&loop).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("loop not found")
		}
		return nil, common.Internal(fmt.Sprintf("failed to get loop: %v", err))
	}
	return &loop, nil
}

// UpdateLoop updates a Loop definition.
func (s *LoopService) UpdateLoop(workspaceID, loopID uint64, name, description *string, loopDef json.RawMessage, status *string) (*agentmodel.Loop, error) {
	loop, err := s.GetLoop(workspaceID, loopID)
	if err != nil {
		return nil, err
	}
	if name != nil {
		loop.Name = *name
	}
	if description != nil {
		loop.Description = description
	}
	if loopDef != nil {
		loop.LoopDef = loopDef
	}
	if status != nil {
		loop.Status = *status
	}
	if err := s.db.Save(loop).Error; err != nil {
		return nil, common.Internal(fmt.Sprintf("failed to update loop: %v", err))
	}
	return loop, nil
}

// DeleteLoop soft-deletes a Loop.
func (s *LoopService) DeleteLoop(workspaceID, loopID uint64) error {
	result := s.db.Where("id = ? AND workspace_id = ?", loopID, workspaceID).Delete(&agentmodel.Loop{})
	if result.RowsAffected == 0 {
		return common.NotFound("loop not found")
	}
	return result.Error
}

// StartLoop creates a LoopRun and starts execution in a background goroutine.
func (s *LoopService) StartLoop(workspaceID, loopID uint64, userID uint64) (*agentmodel.LoopRun, error) {
	loop, err := s.GetLoop(workspaceID, loopID)
	if err != nil {
		return nil, err
	}

	// Parse loop DSL to extract config
	var def struct {
		Goal          string `json:"goal"`
		MaxIterations int    `json:"max_iterations"`
		MaxTokens     int    `json:"max_tokens"`
		MaxCost       float64 `json:"max_cost"`
		MaxDurationSec int   `json:"max_duration_sec"`
	}
	if err := json.Unmarshal(loop.LoopDef, &def); err != nil {
		return nil, common.Validation("invalid loop definition")
	}

	if def.MaxIterations == 0 {
		def.MaxIterations = 100
	}
	if def.MaxTokens == 0 {
		def.MaxTokens = 50000
	}
	if def.MaxDurationSec == 0 {
		def.MaxDurationSec = 3600 // 1 hour
	}

	run := &agentmodel.LoopRun{
		LoopID:        loopID,
		Status:        "running",
		MaxIterations: def.MaxIterations,
		Goal:          def.Goal,
		WorkingMemory: json.RawMessage("{}"),
		StartedAt:     time.Now(),
	}
	if err := s.db.Create(run).Error; err != nil {
		return nil, common.Internal(fmt.Sprintf("failed to create loop run: %v", err))
	}

	// Build executor that bridges to AgentService
	executor := &loopAgentExecutor{
		agentSvc:    s.agentSvc,
		workspaceID: workspaceID,
		userID:      userID,
	}

	collector := &loopMetricsCollector{
		db:          s.db,
		workspaceID: workspaceID,
	}
	evaluator := &loopGoalEvaluator{}

	config := agentloop.LoopConfig{
		Goal:          def.Goal,
		MaxIterations: def.MaxIterations,
		MaxTokens:     def.MaxTokens,
		MaxCost:       def.MaxCost,
		MaxDuration:   time.Duration(def.MaxDurationSec) * time.Second,
	}

	runner := agentloop.NewLoopRunner(s.db, executor, collector, evaluator, config)

	// Run in background
	go func() {
		ctx := context.Background()
		iterations, stopReason, runErr := runner.Run(ctx)

		now := time.Now()
		run.CompletedAt = &now
		run.StoppedReason = &stopReason
		run.TokensUsed = runner.Budget().UsedTokens
		run.CostUSD = runner.Budget().UsedCost

		if runErr != nil || stopReason != "goal_achieved" {
			run.Status = "failed"
		} else {
			run.Status = "completed"
		}

		s.db.Save(run)

		// Save iterations
		for _, iter := range iterations {
			actionJSON, _ := json.Marshal(map[string]string{"task": iter.Action, "result": iter.Result})
			resultJSON, _ := json.Marshal(iter.Metrics)
			reasoning := iter.Reasoning
			dur := iter.DurationMs
			iteration := &agentmodel.LoopIteration{
				LoopRunID:      run.ID,
				IterationNum:   iter.Num,
				ActionTaken:    actionJSON,
				ResultObserved: resultJSON,
				Reasoning:      &reasoning,
				Decision:       string(iter.Decision),
				TokensUsed:     iter.TokensUsed,
				DurationMs:     &dur,
			}
			s.db.Create(iteration)
		}

		// Record agent session
		s.recordSession(workspaceID, runner.SessionID(), "loop_iteration", stopReason, runner.Budget().UsedTokens, runner.Budget().UsedCost, runErr)
	}()

	return run, nil
}

// StopLoop marks a running LoopRun as stopped.
func (s *LoopService) StopLoop(workspaceID, runID uint64) error {
	reason := "manual"
	now := time.Now()
	result := s.db.Model(&agentmodel.LoopRun{}).
		Where("id = ? AND status = ?", runID, "running").
		Updates(map[string]interface{}{
			"status":         "stopped",
			"stopped_reason": &reason,
			"completed_at":   &now,
		})
	if result.RowsAffected == 0 {
		return common.NotFound("loop run not found or not running")
	}
	return result.Error
}

// GetLoopRuns returns run history for a Loop.
func (s *LoopService) GetLoopRuns(workspaceID, loopID uint64, limit int) ([]agentmodel.LoopRun, error) {
	if limit == 0 {
		limit = 20
	}
	var runs []agentmodel.LoopRun
	if err := s.db.Where("loop_id = ?", loopID).
		Order("started_at DESC").Limit(limit).Find(&runs).Error; err != nil {
		return nil, common.Internal(fmt.Sprintf("failed to get loop runs: %v", err))
	}
	return runs, nil
}

// GetLoopRun retrieves a single run with its iterations.
func (s *LoopService) GetLoopRun(workspaceID, runID uint64) (*agentmodel.LoopRun, []agentmodel.LoopIteration, error) {
	var run agentmodel.LoopRun
	if err := s.db.Where("id = ?", runID).First(&run).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, common.NotFound("loop run not found")
		}
		return nil, nil, err
	}
	var iterations []agentmodel.LoopIteration
	s.db.Where("loop_run_id = ?", runID).Order("iteration_num ASC").Find(&iterations)
	return &run, iterations, nil
}

func (s *LoopService) recordSession(workspaceID uint64, sessionID, agentType, reason string, tokens int, cost float64, runErr error) {
	status := "completed"
	errMsg := (*string)(nil)
	if runErr != nil || (reason != "goal_achieved" && reason != "cancelled") {
		status = "failed"
		msg := reason
		errMsg = &msg
	}

	now := time.Now()
	summary := fmt.Sprintf("Loop run completed: %s", reason)
	session := &agentmodel.AgentSession{
		ID:            sessionID,
		WorkspaceID:   workspaceID,
		AgentType:     agentType,
		Status:        status,
		InputSummary:  &summary,
		OutputSummary: &reason,
		TokensInput:   tokens / 2,
		TokensOutput:  tokens / 2,
		CostUSD:       cost,
		ErrorMessage:  errMsg,
		StartedAt:     time.Now().Add(-1 * time.Minute),
		CompletedAt:   &now,
	}
	s.db.Create(session)
}

// --- Internal adapters ---

// loopAgentExecutor adapts AgentService to the Loop's AgentExecutor interface.
type loopAgentExecutor struct {
	agentSvc    *AgentService
	workspaceID uint64
	userID      uint64
}

func (e *loopAgentExecutor) Execute(ctx context.Context, task string, context map[string]interface{}) (string, int, float64, error) {
	// Use AgentService to find a suitable agent and dispatch
	agents, err := e.agentSvc.ListByWorkspace(e.workspaceID)
	if err != nil || len(agents) == 0 {
		return "", 0, 0, fmt.Errorf("no agents available: %w", err)
	}

	// Use first active agent (or a "loop-runner" agent if configured)
	var runner *model.Agent
	for i := range agents {
		if agents[i].Status == "active" {
			runner = &agents[i]
			break
		}
	}
	if runner == nil {
		return "", 0, 0, fmt.Errorf("no active agent found")
	}

	// Build dispatch context
	var issueIDPtr *uint64
	if issueID, ok := context["issue_id"].(uint64); ok {
		issueIDPtr = &issueID
	}
	var projectIDPtr *uint64
	if projectID, ok := context["project_id"].(uint64); ok {
		projectIDPtr = &projectID
	}

	dispCtx := &DispatchContext{
		IssueID:     issueIDPtr,
		ProjectID:   projectIDPtr,
		WorkspaceID: e.workspaceID,
		TriggeredBy: "loop",
	}

	activity, err := e.agentSvc.DispatchAgent(runner.ID, e.userID, task, dispCtx)
	if err != nil {
		return "", 0, 0, err
	}

	result := activity.ResultSummary
	tokens := 500 // estimate; real tracking needs LLM client instrumentation
	cost := float64(tokens) * 0.000002 // ~$2/M tokens for Sonnet-level

	return result, tokens, cost, nil
}

// loopMetricsCollector collects project metrics from the database.
type loopMetricsCollector struct {
	db          *gorm.DB
	workspaceID uint64
}

func (c *loopMetricsCollector) Collect(ctx context.Context, context map[string]interface{}) (map[string]float64, error) {
	metrics := make(map[string]float64)

	// If cycle_id is in context, collect Sprint-specific metrics
	if cycleID, ok := context["cycle_id"]; ok {
		var total, completed int64
		c.db.Table("issues").Where("cycle_id = ? AND deleted_at IS NULL", cycleID).Count(&total)
		c.db.Table("issues").Where("cycle_id = ? AND state_group = 'completed' AND deleted_at IS NULL", cycleID).Count(&completed)
		if total > 0 {
			metrics["progress"] = float64(completed) / float64(total)
		} else {
			metrics["progress"] = 0
		}
	}

	return metrics, nil
}

// loopGoalEvaluator checks simple threshold-based goals.
type loopGoalEvaluator struct{}

func (e *loopGoalEvaluator) Evaluate(goal string, metrics map[string]float64) (bool, string) {
	// Simple threshold evaluation for Phase 1
	// Goal format: "progress > 0.9" or "progress > 0.9 AND workload_balance > 0.7"
	if progress, ok := metrics["progress"]; ok {
		if progress >= 0.9 {
			return true, fmt.Sprintf("progress %.1f%% meets 90%% target", progress*100)
		}
		return false, fmt.Sprintf("progress %.1f%% below 90%% target", progress*100)
	}
	return false, "no progress metric found"
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./internal/service/ 2>&1
```
Expected: compile succeeds (fix any import issues)

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/agent_loop_service.go
git commit -m "feat(agent): add LoopService with CRUD, Start/Stop, and agent executor adapter"
```

---

### Task 6: Loop & Session API Handlers

**Files:**
- Create: `backend/internal/handler/agent_loop.go`
- Create: `backend/internal/handler/agent_session.go`

- [ ] **Step 1: Write the Loop handler**

```go
// backend/internal/handler/agent_loop.go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

type AgentLoopHandler struct {
	svc *service.LoopService
}

func NewAgentLoopHandler(svc *service.LoopService) *AgentLoopHandler {
	return &AgentLoopHandler{svc: svc}
}

func (h *AgentLoopHandler) getWSAndUser(c *gin.Context) (uint64, uint64, error) {
	wsID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	user := middleware.GetCurrentUser(c)
	if user == nil {
		return 0, 0, common.NotFound("user not found")
	}
	return wsID, user.ID, nil
}

func (h *AgentLoopHandler) Create(c *gin.Context) {
	wsID, userID, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}

	var req struct {
		Name        string          `json:"name" binding:"required"`
		Description string          `json:"description"`
		LoopDef     json.RawMessage `json:"loop_def" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	loop, err := h.svc.CreateLoop(wsID, userID, req.Name, req.Description, req.LoopDef)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, loop)
}

func (h *AgentLoopHandler) List(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loops, err := h.svc.ListLoops(wsID)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loops)
}

func (h *AgentLoopHandler) Get(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid loop id"})
		return
	}
	loop, err := h.svc.GetLoop(wsID, loopID)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loop)
}

func (h *AgentLoopHandler) Update(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid loop id"})
		return
	}

	var req struct {
		Name        *string         `json:"name"`
		Description *string         `json:"description"`
		LoopDef     json.RawMessage `json:"loop_def"`
		Status      *string         `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	loop, err := h.svc.UpdateLoop(wsID, loopID, req.Name, req.Description, req.LoopDef, req.Status)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loop)
}

func (h *AgentLoopHandler) Delete(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid loop id"})
		return
	}
	if err := h.svc.DeleteLoop(wsID, loopID); err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *AgentLoopHandler) Start(c *gin.Context) {
	wsID, userID, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid loop id"})
		return
	}
	run, err := h.svc.StartLoop(wsID, loopID, userID)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *AgentLoopHandler) Stop(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	runID, err := strconv.ParseUint(c.Param("runId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid run id"})
		return
	}
	if err := h.svc.StopLoop(wsID, runID); err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "loop stopped"})
}

func (h *AgentLoopHandler) GetRuns(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid loop id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	runs, err := h.svc.GetLoopRuns(wsID, loopID, limit)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, runs)
}

func (h *AgentLoopHandler) GetRun(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	runID, err := strconv.ParseUint(c.Param("runId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid run id"})
		return
	}
	run, iterations, err := h.svc.GetLoopRun(wsID, runID)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"run": run, "iterations": iterations})
}
```

- [ ] **Step 2: Write the Session handler**

```go
// backend/internal/handler/agent_session.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	agentmodel "github.com/reqmango/backend/internal/agent/model"
	"gorm.io/gorm"
)

type AgentSessionHandler struct {
	db *gorm.DB
}

func NewAgentSessionHandler(db *gorm.DB) *AgentSessionHandler {
	return &AgentSessionHandler{db: db}
}

func (h *AgentSessionHandler) getWS(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

func (h *AgentSessionHandler) List(c *gin.Context) {
	wsID, err := h.getWS(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	agentType := c.Query("agent_type")
	status := c.Query("status")

	var sessions []agentmodel.AgentSession
	query := h.db.Where("workspace_id = ?", wsID)
	if agentType != "" {
		query = query.Where("agent_type = ?", agentType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("started_at DESC").Limit(limit).Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": common.Internal(err.Error()).Message})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

func (h *AgentSessionHandler) Get(c *gin.Context) {
	wsID, err := h.getWS(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}

	sessionID := c.Param("sessionId")
	var session agentmodel.AgentSession
	if err := h.db.Where("id = ? AND workspace_id = ?", sessionID, wsID).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session)
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/agent_loop.go backend/internal/handler/agent_session.go
git commit -m "feat(agent): add Loop and Session HTTP handlers"
```

---

### Task 7: Register Routes & Auto-Migrate

**Files:**
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Add routes and auto-migration to router.go**

Read the existing `router.go` to find the right insertion points. Add the following changes:

After the existing service initializations (around line 61 where `commentSvc.SetAgentService(agentSvc)` is called), add:

```go
// Add after: commentSvc.SetAgentService(agentSvc)
loopSvc := service.NewLoopService(db, agentSvc)
```

After the existing handler initializations, add:

```go
// Add after existing handler inits
loopH := handler.NewAgentLoopHandler(loopSvc)
sessionH := handler.NewAgentSessionHandler(db)
```

In the `autoMigrate` call (search for `AutoMigrate` in router.go), add the new models:

```go
// Add to the AutoMigrate list
&agentmodel.Loop{},
&agentmodel.LoopRun{},
&agentmodel.LoopIteration{},
&agentmodel.AgentSession{},
```

Make sure the import includes:
```go
agentmodel "github.com/reqmango/backend/internal/agent/model"
```

After the existing agent routes block, add the Loop routes:

```go
// Agent Loops
loops := v1.Group("/workspaces/:wsParam/loops", authMiddleware)
{
	loops.GET("", loopH.List)
	loops.POST("", loopH.Create)
	loops.GET("/:id", loopH.Get)
	loops.PUT("/:id", loopH.Update)
	loops.DELETE("/:id", loopH.Delete)
	loops.POST("/:id/start", loopH.Start)
	loops.GET("/:id/runs", loopH.GetRuns)
	loops.POST("/runs/:runId/stop", loopH.Stop)
	loops.GET("/runs/:runId", loopH.GetRun)
}

// Agent Sessions
sessions := v1.Group("/workspaces/:wsParam/agent-sessions", authMiddleware)
{
	sessions.GET("", sessionH.List)
	sessions.GET("/:sessionId", sessionH.Get)
}
```

- [ ] **Step 2: Verify full backend compilation**

```bash
cd backend && go build ./cmd/server/ 2>&1
```
Expected: compile succeeds. Fix any import issues.

- [ ] **Step 3: Start the server and verify auto-migration**

```bash
cd backend && go run ./cmd/server/ 2>&1 | head -30
```
Expected: server starts, tables `agent_loops`, `agent_loop_runs`, `agent_loop_iterations`, `agent_sessions` are created.

- [ ] **Step 4: Test the API endpoints**

```bash
# Login first
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"demo1234"}' | jq -r '.access_token')

# Create a Loop
curl -s -X POST http://localhost:8000/api/v1/workspaces/1/loops \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Loop","loop_def":{"goal":"progress > 0.9","max_iterations":3}}' | jq .

# List Loops
curl -s http://localhost:8000/api/v1/workspaces/1/loops \
  -H "Authorization: Bearer $TOKEN" | jq .
```

Expected: Create returns 201 with loop JSON. List returns 200 with array.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/router/router.go
git commit -m "feat(agent): register Loop/Session routes and auto-migrate models"
```

---

### Task 8: Sprint Agent Loop Scenario

**Files:**
- Create: `backend/internal/agent/loop/sprint_guardian.go`

- [ ] **Step 1: Write the Sprint Guardian Loop preset**

```go
// backend/internal/agent/loop/sprint_guardian.go
package loop

import (
	"encoding/json"
)

// SprintGuardianPreset returns the pre-configured Loop definition
// for the Sprint Guardian — the flagship Phase 1 scenario.
func SprintGuardianPreset() json.RawMessage {
	def := map[string]interface{}{
		"name":        "sprint-guardian",
		"description": "Sprint 自主守护Agent — 每日检查、自动调整、风险预警",
		"version":     "1.0",
		"trigger": map[string]interface{}{
			"type":     "cron",
			"schedule": "0 9 * * 1-5",
		},
		"loop": map[string]interface{}{
			"type":                    "goal_based",
			"goal":                    "Sprint进度偏差 < 10% AND 无人过载",
			"max_iterations_per_check": 3,
			"max_daily_tokens":        30000,
		},
		"agent": map[string]interface{}{
			"model": map[string]interface{}{
				"planner":  "opus",
				"executor": "sonnet",
				"effort":   "high",
			},
		},
		"actions": []string{
			"analyze_progress",
			"detect_blockers",
			"check_workload",
			"suggest_rebalance",
			"auto_rebalance",
			"notify_stakeholders",
			"generate_daily_digest",
			"generate_sprint_review",
		},
		"notifications": []map[string]interface{}{
			{
				"channel": "in_app",
				"on":      []string{"blocker_detected", "risk_escalated", "daily_digest"},
			},
		},
		"budget": map[string]interface{}{
			"max_tokens_per_day":   50000,
			"max_cost_per_sprint":  2.00,
			"on_budget_critical":   "notify_admin",
		},
	}

	raw, _ := json.Marshal(def)
	return raw
}
```

- [ ] **Step 2: Add a seed function for Sprint Loop in the LoopService**

Add to `backend/internal/service/agent_loop_service.go`:

```go
// SeedSprintGuardianLoop creates the Sprint Guardian loop preset if it doesn't exist.
func (s *LoopService) SeedSprintGuardianLoop(workspaceID, userID uint64) error {
	var count int64
	s.db.Model(&agentmodel.Loop{}).Where("workspace_id = ? AND name = ?", workspaceID, "sprint-guardian").Count(&count)
	if count > 0 {
		return nil // already seeded
	}

	def := agentloop.SprintGuardianPreset()
	desc := "Sprint 自主守护Agent — 每日检查、自动调整、风险预警"
	_, err := s.CreateLoop(workspaceID, userID, "sprint-guardian", desc, def)
	return err
}
```

- [ ] **Step 3: Wire seed into server startup**

In `backend/cmd/server/main.go` or `backend/internal/router/router.go`, after all services are initialized, add:

```go
// Seed Sprint Guardian loop for all workspaces
workspaces, _ := workspaceSvc.List()
for _, ws := range workspaces {
    loopSvc.SeedSprintGuardianLoop(ws.ID, 1) // user 1 = admin
}
```

- [ ] **Step 4: Verify compilation and test**

```bash
cd backend && go build ./cmd/server/
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/agent/loop/sprint_guardian.go backend/internal/service/agent_loop_service.go backend/internal/router/router.go
git commit -m "feat(agent): add Sprint Guardian Loop preset with auto-seed"
```

---

### Task 9: Frontend — Agent Loop API & Store

**Files:**
- Create: `frontend/src/api/agent-loop.ts`
- Create: `frontend/src/api/agent-session.ts`
- Create: `frontend/src/stores/agentLoop.ts`

- [ ] **Step 1: Create the Loop API module**

```typescript
// frontend/src/api/agent-loop.ts
import api from './index'

export interface LoopDef {
  goal?: string
  max_iterations?: number
  max_tokens?: number
  max_cost?: number
  max_duration_sec?: number
  trigger?: { type: string; schedule?: string; event?: string }
  actions?: string[]
  [key: string]: any
}

export interface Loop {
  id: number
  workspace_id: number
  name: string
  description?: string
  loop_def: LoopDef
  version: string
  status: 'active' | 'draft' | 'archived'
  created_by_id?: number
  created_at: string
  updated_at: string
}

export interface LoopRun {
  id: number
  loop_id: number
  status: 'running' | 'completed' | 'failed' | 'escalated' | 'stopped'
  current_iteration: number
  max_iterations: number
  goal: string
  goal_metrics?: Record<string, number>
  working_memory?: Record<string, any>
  tokens_used: number
  cost_usd: number
  stopped_reason?: string
  started_at: string
  completed_at?: string
}

export interface LoopIteration {
  id: number
  loop_run_id: number
  iteration_num: number
  action_taken: { task: string; result: string }
  result_observed: Record<string, number>
  reasoning?: string
  decision: 'continue' | 'stop' | 'escalate' | 'wait'
  tokens_used: number
  duration_ms?: number
  created_at: string
}

export interface LoopRunDetail {
  run: LoopRun
  iterations: LoopIteration[]
}

export const loopApi = {
  list(workspaceId: number): Promise<Loop[]> {
    return api.get(`/workspaces/${workspaceId}/loops`).then(r => r.data)
  },

  get(workspaceId: number, loopId: number): Promise<Loop> {
    return api.get(`/workspaces/${workspaceId}/loops/${loopId}`).then(r => r.data)
  },

  create(workspaceId: number, data: { name: string; description?: string; loop_def: LoopDef }): Promise<Loop> {
    return api.post(`/workspaces/${workspaceId}/loops`, data).then(r => r.data)
  },

  update(workspaceId: number, loopId: number, data: Partial<Loop>): Promise<Loop> {
    return api.put(`/workspaces/${workspaceId}/loops/${loopId}`, data).then(r => r.data)
  },

  delete(workspaceId: number, loopId: number): Promise<void> {
    return api.delete(`/workspaces/${workspaceId}/loops/${loopId}`)
  },

  start(workspaceId: number, loopId: number): Promise<LoopRun> {
    return api.post(`/workspaces/${workspaceId}/loops/${loopId}/start`).then(r => r.data)
  },

  stop(workspaceId: number, runId: number): Promise<void> {
    return api.post(`/workspaces/${workspaceId}/loops/runs/${runId}/stop`)
  },

  getRuns(workspaceId: number, loopId: number, limit = 20): Promise<LoopRun[]> {
    return api.get(`/workspaces/${workspaceId}/loops/${loopId}/runs`, { params: { limit } }).then(r => r.data)
  },

  getRun(workspaceId: number, runId: number): Promise<LoopRunDetail> {
    return api.get(`/workspaces/${workspaceId}/loops/runs/${runId}`).then(r => r.data)
  },
}
```

- [ ] **Step 2: Create the Session API module**

```typescript
// frontend/src/api/agent-session.ts
import api from './index'

export interface AgentSession {
  id: string
  workspace_id: number
  agent_type: 'loop_iteration' | 'pipeline_stage' | 'standalone_dispatch'
  agent_ref?: string
  status: 'running' | 'completed' | 'failed'
  model_used?: string
  input_summary?: string
  output_summary?: string
  tokens_input: number
  tokens_output: number
  cost_usd: number
  tools_called?: { tool_name: string; count: number }[]
  error_message?: string
  started_at: string
  completed_at?: string
}

export const sessionApi = {
  list(workspaceId: number, params?: { agent_type?: string; status?: string; limit?: number }): Promise<AgentSession[]> {
    return api.get(`/workspaces/${workspaceId}/agent-sessions`, { params }).then(r => r.data)
  },

  get(workspaceId: number, sessionId: string): Promise<AgentSession> {
    return api.get(`/workspaces/${workspaceId}/agent-sessions/${sessionId}`).then(r => r.data)
  },
}
```

- [ ] **Step 3: Create the Pinia store**

```typescript
// frontend/src/stores/agentLoop.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { loopApi, type Loop, type LoopRun, type LoopRunDetail } from '@/api/agent-loop'

export const useAgentLoopStore = defineStore('agentLoop', () => {
  const loops = ref<Loop[]>([])
  const currentRun = ref<LoopRun | null>(null)
  const runDetail = ref<LoopRunDetail | null>(null)
  const loading = ref(false)

  async function fetchLoops(workspaceId: number) {
    loading.value = true
    try {
      loops.value = await loopApi.list(workspaceId)
    } finally {
      loading.value = false
    }
  }

  async function startLoop(workspaceId: number, loopId: number) {
    const run = await loopApi.start(workspaceId, loopId)
    currentRun.value = run
    return run
  }

  async function stopLoop(workspaceId: number, runId: number) {
    await loopApi.stop(workspaceId, runId)
    if (currentRun.value?.id === runId) {
      currentRun.value = { ...currentRun.value, status: 'stopped' }
    }
  }

  async function fetchRunDetail(workspaceId: number, runId: number) {
    runDetail.value = await loopApi.getRun(workspaceId, runId)
    currentRun.value = runDetail.value.run
    return runDetail.value
  }

  // Poll for run status updates
  function watchRun(workspaceId: number, runId: number, intervalMs = 5000) {
    const interval = setInterval(async () => {
      try {
        const detail = await loopApi.getRun(workspaceId, runId)
        runDetail.value = detail
        currentRun.value = detail.run
        if (detail.run.status !== 'running') {
          clearInterval(interval)
        }
      } catch {
        clearInterval(interval)
      }
    }, intervalMs)
    return () => clearInterval(interval)
  }

  return { loops, currentRun, runDetail, loading, fetchLoops, startLoop, stopLoop, fetchRunDetail, watchRun }
})
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/api/agent-loop.ts frontend/src/api/agent-session.ts frontend/src/stores/agentLoop.ts
git commit -m "feat(frontend): add Loop/Session API modules and agentLoop Pinia store"
```

---

### Task 10: Frontend — Agent Dashboard & Loop Pages

**Files:**
- Create: `frontend/src/views/agents/AgentDashboard.vue`
- Create: `frontend/src/views/agents/LoopRunDetail.vue`
- Create: `frontend/src/views/agents/AgentSessions.vue`
- Create: `frontend/src/components/agents/LoopStateBadge.vue`
- Create: `frontend/src/components/agents/BudgetGauge.vue`
- Modify: `frontend/src/router/index.ts`

- [ ] **Step 1: Create the LoopStateBadge component**

```vue
<!-- frontend/src/components/agents/LoopStateBadge.vue -->
<template>
  <span :class="badgeClass" class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium">
    {{ label }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ status: string }>()

const badgeClass = computed(() => {
  switch (props.status) {
    case 'running': return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
    case 'completed': return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
    case 'failed': return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
    case 'escalated': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
    case 'stopped': return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
    default: return 'bg-gray-100 text-gray-800'
  }
})

const label = computed(() => {
  switch (props.status) {
    case 'running': return 'Running'
    case 'completed': return 'Completed'
    case 'failed': return 'Failed'
    case 'escalated': return 'Escalated'
    case 'stopped': return 'Stopped'
    default: return props.status
  }
})
</script>
```

- [ ] **Step 2: Create the BudgetGauge component**

```vue
<!-- frontend/src/components/agents/BudgetGauge.vue -->
<template>
  <div class="space-y-2">
    <div v-if="maxTokens > 0" class="flex items-center gap-2 text-xs">
      <span class="text-gray-500 w-16">Tokens</span>
      <div class="flex-1 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
        <div
          class="h-full rounded-full transition-all"
          :class="usedTokens / maxTokens > 0.9 ? 'bg-red-500' : 'bg-indigo-500'"
          :style="{ width: Math.min(100, (usedTokens / maxTokens) * 100) + '%' }"
        ></div>
      </div>
      <span class="text-gray-500 w-24 text-right">{{ usedTokens.toLocaleString() }} / {{ maxTokens.toLocaleString() }}</span>
    </div>
    <div v-if="maxIterations > 0" class="flex items-center gap-2 text-xs">
      <span class="text-gray-500 w-16">Iterations</span>
      <div class="flex-1 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
        <div
          class="h-full rounded-full transition-all bg-purple-500"
          :style="{ width: Math.min(100, (currentIteration / maxIterations) * 100) + '%' }"
        ></div>
      </div>
      <span class="text-gray-500 w-24 text-right">{{ currentIteration }} / {{ maxIterations }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  maxTokens: number
  usedTokens: number
  maxIterations: number
  currentIteration: number
}>()
</script>
```

- [ ] **Step 3: Create the AgentDashboard page**

```vue
<!-- frontend/src/views/agents/AgentDashboard.vue -->
<template>
  <div class="p-6 max-w-5xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">🤖 Agent Dashboard</h1>
        <p class="text-sm text-gray-500 mt-1">Monitor and manage autonomous agent loops</p>
      </div>
    </div>

    <!-- Active Runs -->
    <section class="mb-8">
      <h2 class="text-lg font-semibold mb-3">Active Loops</h2>
      <div v-if="activeRuns.length === 0" class="text-sm text-gray-400 py-8 text-center border rounded-lg">
        No active loop runs. Start one from the Loops tab.
      </div>
      <div v-else class="space-y-3">
        <div
          v-for="run in activeRuns"
          :key="run.id"
          class="border rounded-lg p-4 hover:border-indigo-300 cursor-pointer transition-colors"
          @click="$router.push(`/agents/loops/runs/${run.id}`)"
        >
          <div class="flex items-center justify-between mb-2">
            <span class="font-medium">{{ run.goal }}</span>
            <LoopStateBadge :status="run.status" />
          </div>
          <BudgetGauge
            :max-tokens="50000"
            :used-tokens="run.tokens_used"
            :max-iterations="run.max_iterations"
            :current-iteration="run.current_iteration"
          />
        </div>
      </div>
    </section>

    <!-- Quick Actions -->
    <section class="grid grid-cols-2 gap-4">
      <router-link
        to="/agents/loops"
        class="border rounded-lg p-4 hover:border-indigo-400 hover:shadow-sm transition-all"
      >
        <div class="text-xl mb-1">🔄</div>
        <div class="font-medium text-sm">Loop Configurations</div>
        <div class="text-xs text-gray-400 mt-1">Create and manage autonomous loops</div>
      </router-link>
      <router-link
        to="/agents/sessions"
        class="border rounded-lg p-4 hover:border-indigo-400 hover:shadow-sm transition-all"
      >
        <div class="text-xl mb-1">📋</div>
        <div class="font-medium text-sm">Agent Sessions</div>
        <div class="text-xs text-gray-400 mt-1">View agent execution history and costs</div>
      </router-link>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { loopApi, type LoopRun } from '@/api/agent-loop'
import LoopStateBadge from '@/components/agents/LoopStateBadge.vue'
import BudgetGauge from '@/components/agents/BudgetGauge.vue'
import { useWorkspaceStore } from '@/stores/workspace'

const activeRuns = ref<LoopRun[]>([])
const ws = useWorkspaceStore()

onMounted(async () => {
  try {
    const loops = await loopApi.list(ws.currentId!)
    for (const loop of loops) {
      if (loop.status === 'active') {
        const runs = await loopApi.getRuns(ws.currentId!, loop.id, 5)
        activeRuns.value.push(...runs.filter(r => r.status === 'running'))
      }
    }
  } catch { /* no loops yet */ }
})
</script>
```

- [ ] **Step 4: Create the LoopRunDetail page**

```vue
<!-- frontend/src/views/agents/LoopRunDetail.vue -->
<template>
  <div class="p-6 max-w-4xl mx-auto">
    <div class="flex items-center gap-3 mb-6">
      <button @click="$router.back()" class="text-gray-400 hover:text-gray-600">← Back</button>
      <h1 class="text-xl font-bold">Loop Run #{{ runId }}</h1>
      <LoopStateBadge v-if="detail?.run.status" :status="detail.run.status" />
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-400">Loading...</div>

    <template v-else-if="detail">
      <!-- Summary Card -->
      <div class="border rounded-lg p-4 mb-6">
        <div class="font-medium mb-2">{{ detail.run.goal }}</div>
        <BudgetGauge
          :max-tokens="50000"
          :used-tokens="detail.run.tokens_used"
          :max-iterations="detail.run.max_iterations"
          :current-iteration="detail.run.current_iteration"
        />
        <div class="text-xs text-gray-400 mt-2">
          Started {{ new Date(detail.run.started_at).toLocaleString() }}
          <span v-if="detail.run.stopped_reason"> · {{ detail.run.stopped_reason }}</span>
        </div>
      </div>

      <!-- Iterations Timeline -->
      <h2 class="font-semibold mb-3">Iterations</h2>
      <div class="space-y-3">
        <div
          v-for="iter in detail.iterations"
          :key="iter.id"
          class="border rounded-lg p-4"
        >
          <div class="flex items-center justify-between mb-2">
            <span class="text-sm font-medium text-gray-600">Iteration #{{ iter.iteration_num }}</span>
            <span
              :class="{
                'text-green-600': iter.decision === 'stop',
                'text-blue-600': iter.decision === 'continue',
                'text-yellow-600': iter.decision === 'escalate',
              }"
              class="text-xs font-medium uppercase"
            >
              {{ iter.decision }}
            </span>
          </div>
          <p class="text-sm text-gray-500 mb-2">{{ iter.reasoning }}</p>
          <div class="flex gap-4 text-xs text-gray-400">
            <span>Tokens: {{ iter.tokens_used }}</span>
            <span v-if="iter.duration_ms">Duration: {{ (iter.duration_ms / 1000).toFixed(1) }}s</span>
          </div>
        </div>
        <div v-if="detail.iterations.length === 0" class="text-center py-8 text-gray-400 text-sm">
          No iterations recorded yet
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAgentLoopStore } from '@/stores/agentLoop'
import LoopStateBadge from '@/components/agents/LoopStateBadge.vue'
import BudgetGauge from '@/components/agents/BudgetGauge.vue'
import { useWorkspaceStore } from '@/stores/workspace'

const route = useRoute()
const runId = Number(route.params.runId)
const store = useAgentLoopStore()
const ws = useWorkspaceStore()

const detail = ref(store.runDetail)
const loading = ref(true)
let stopWatch: (() => void) | null = null

onMounted(async () => {
  await store.fetchRunDetail(ws.currentId!, runId)
  detail.value = store.runDetail
  loading.value = false

  if (detail.value?.run.status === 'running') {
    stopWatch = store.watchRun(ws.currentId!, runId, 3000)
  }
})

onUnmounted(() => {
  stopWatch?.()
})
</script>
```

- [ ] **Step 5: Create the AgentSessions page**

```vue
<!-- frontend/src/views/agents/AgentSessions.vue -->
<template>
  <div class="p-6 max-w-5xl mx-auto">
    <h1 class="text-2xl font-bold mb-6">📋 Agent Sessions</h1>

    <div v-if="loading" class="text-center py-12 text-gray-400">Loading...</div>

    <div v-else-if="sessions.length === 0" class="text-center py-12 text-gray-400 border rounded-lg">
      No agent sessions recorded yet
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="session in sessions"
        :key="session.id"
        class="border rounded-lg p-3 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
      >
        <div class="flex items-center justify-between">
          <div>
            <span class="text-xs font-mono text-gray-400 mr-2">{{ session.id.slice(0, 8) }}</span>
            <span class="text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700">{{ session.agent_type }}</span>
          </div>
          <LoopStateBadge :status="session.status" />
        </div>
        <div v-if="session.input_summary" class="text-sm text-gray-600 mt-1">{{ session.input_summary }}</div>
        <div class="flex gap-4 text-xs text-gray-400 mt-1">
          <span>T: {{ session.tokens_input + session.tokens_output }}</span>
          <span>${{ session.cost_usd.toFixed(4) }}</span>
          <span>{{ new Date(session.started_at).toLocaleString() }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { sessionApi, type AgentSession } from '@/api/agent-session'
import LoopStateBadge from '@/components/agents/LoopStateBadge.vue'
import { useWorkspaceStore } from '@/stores/workspace'

const ws = useWorkspaceStore()
const sessions = ref<AgentSession[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    sessions.value = await sessionApi.list(ws.currentId!, { limit: 50 })
  } finally {
    loading.value = false
  }
})
</script>
```

- [ ] **Step 6: Add routes to the router**

In `frontend/src/router/index.ts`, add inside the workspace-scoped children:

```typescript
// Add these routes after the existing project routes
{
  path: 'agents',
  name: 'agent-dashboard',
  component: () => import('@/views/agents/AgentDashboard.vue'),
  meta: { requiresAuth: true },
},
{
  path: 'agents/loops',
  name: 'agent-loops',
  component: () => import('@/views/agents/LoopList.vue'),
  meta: { requiresAuth: true },
},
{
  path: 'agents/loops/runs/:runId',
  name: 'loop-run-detail',
  component: () => import('@/views/agents/LoopRunDetail.vue'),
  meta: { requiresAuth: true },
},
{
  path: 'agents/sessions',
  name: 'agent-sessions',
  component: () => import('@/views/agents/AgentSessions.vue'),
  meta: { requiresAuth: true },
},
```

- [ ] **Step 7: Create LoopList.vue (simple list page)**

```vue
<!-- frontend/src/views/agents/LoopList.vue -->
<template>
  <div class="p-6 max-w-5xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">🔄 Agent Loops</h1>
      <button
        @click="showCreate = true"
        class="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700"
      >
        + New Loop
      </button>
    </div>

    <div v-if="store.loading" class="text-center py-12 text-gray-400">Loading...</div>

    <div v-else-if="store.loops.length === 0" class="text-center py-12 text-gray-400 border rounded-lg">
      No loops configured yet. Create one to get started.
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="loop in store.loops"
        :key="loop.id"
        class="border rounded-lg p-4 hover:border-indigo-300 transition-colors"
      >
        <div class="flex items-center justify-between">
          <div>
            <span class="font-medium">{{ loop.name }}</span>
            <span v-if="loop.description" class="text-sm text-gray-400 ml-2">{{ loop.description }}</span>
          </div>
          <div class="flex gap-2">
            <button
              @click="startLoop(loop.id)"
              class="px-3 py-1 bg-green-600 text-white rounded text-xs hover:bg-green-700"
            >
              ▶ Run
            </button>
          </div>
        </div>
        <div class="text-xs text-gray-400 mt-1">
          Status: {{ loop.status }} · v{{ loop.version }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAgentLoopStore } from '@/stores/agentLoop'
import { useWorkspaceStore } from '@/stores/workspace'
import { useRouter } from 'vue-router'

const store = useAgentLoopStore()
const ws = useWorkspaceStore()
const router = useRouter()
const showCreate = ref(false)

onMounted(() => store.fetchLoops(ws.currentId!))

async function startLoop(loopId: number) {
  try {
    const run = await store.startLoop(ws.currentId!, loopId)
    router.push(`/agents/loops/runs/${run.id}`)
  } catch (e) {
    console.error('Failed to start loop:', e)
  }
}
</script>
```

- [ ] **Step 8: Verify frontend compilation**

```bash
cd frontend && npx vue-tsc --noEmit 2>&1 | head -20
```
Expected: no type errors (fix any missing imports)

- [ ] **Step 9: Commit**

```bash
git add frontend/src/views/agents/ frontend/src/components/agents/ frontend/src/router/index.ts
git commit -m "feat(frontend): add Agent Dashboard, Loop List, Run Detail, and Sessions pages"
```

---

### Task 11: Integration Test — Sprint Loop E2E

**Files:**
- Create: `tests/agent_loop_integration_test.go`

- [ ] **Step 1: Write the integration test**

```go
// tests/agent_loop_integration_test.go
package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestLoopCRUD verifies the full CRUD lifecycle of a Loop.
func TestLoopCRUD(t *testing.T) {
	token := mustLogin(t, "demo@example.com", "demo1234")

	// CREATE
	createBody := map[string]interface{}{
		"name":        "test-loop",
		"description": "integration test loop",
		"loop_def": map[string]interface{}{
			"goal":           "progress > 0.5",
			"max_iterations": 3,
			"max_tokens":     1000,
		},
	}
	resp := mustPost(t, token, "/api/v1/workspaces/1/loops", createBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create loop: expected 201, got %d", resp.StatusCode)
	}

	var loop map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loop)
	loopID := int(loop["id"].(float64))

	// LIST
	resp = mustGet(t, token, "/api/v1/workspaces/1/loops")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list loops: expected 200, got %d", resp.StatusCode)
	}

	// GET
	resp = mustGet(t, token, "/api/v1/workspaces/1/loops/"+itoa(loopID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get loop: expected 200, got %d", resp.StatusCode)
	}

	// START
	resp = mustPost(t, token, "/api/v1/workspaces/1/loops/"+itoa(loopID)+"/start", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("start loop: expected 200, got %d", resp.StatusCode)
	}

	var run map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&run)
	runID := int(run["id"].(float64))

	// GET RUN
	resp = mustGet(t, token, "/api/v1/workspaces/1/loops/runs/"+itoa(runID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get run: expected 200, got %d", resp.StatusCode)
	}

	t.Logf("Loop %d started, run %d created", loopID, runID)
}

// TestSessionList verifies the session history API.
func TestSessionList(t *testing.T) {
	token := mustLogin(t, "demo@example.com", "demo1234")

	resp := mustGet(t, token, "/api/v1/workspaces/1/agent-sessions?limit=10")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list sessions: expected 200, got %d", resp.StatusCode)
	}

	var sessions []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sessions)
	t.Logf("Found %d agent sessions", len(sessions))
}
```

- [ ] **Step 2: Run integration tests**

```bash
cd backend && go test ./tests/ -v -run "TestLoopCRUD|TestSessionList" -count=1
```
Expected: 2 tests PASS

- [ ] **Step 3: Commit**

```bash
git add tests/agent_loop_integration_test.go
git commit -m "test(agent): add integration tests for Loop CRUD and Session list"
```

---

### Task 12: Final Verification & Documentation

- [ ] **Step 1: Run all backend tests**

```bash
cd backend && go test ./... 2>&1
```
Expected: all tests pass

- [ ] **Step 2: Run frontend build**

```bash
cd frontend && npm run build 2>&1
```
Expected: build succeeds without errors

- [ ] **Step 3: Manual smoke test checklist**

```
□ Login → Navigate to Agent Dashboard (/agents)
□ View Loops list → See "sprint-guardian" pre-seeded
□ Click "Run" on sprint-guardian → Navigated to Run Detail
□ Run Detail shows iterations appearing in real-time
□ Navigate to Agent Sessions (/agents/sessions)
□ See session records from loop runs
□ Open AICopilot (Ctrl+J) → Agent tab → Agent picker works
□ In Project Settings → Automations → Create rule with dispatch_agent action
□ Rule triggers on issue creation → Agent dispatched → Session recorded
```

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat: Phase 1 Agent Loop MVP — Loop Engine + Sprint Guardian + Agent Sessions

- Loop Engine: state machine, runner, budget controller, trigger system
- Sprint Guardian Loop: pre-seeded autonomous sprint monitoring loop
- Agent Sessions: unified observability across loop iterations and dispatches
- Frontend: Agent Dashboard, Loop List/Run Detail, Sessions pages
- Integration tests for Loop CRUD and Session API

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Verification Summary

After completing all tasks, verify:

| Check | Expected |
|-------|----------|
| `go test ./...` in backend/ | All tests pass |
| `npm run build` in frontend/ | Build succeeds |
| POST /api/v1/workspaces/1/loops | 201 Created |
| GET /api/v1/workspaces/1/loops | Returns list with sprint-guardian |
| POST /api/v1/workspaces/1/loops/1/start | Returns run, background execution starts |
| GET /api/v1/workspaces/1/loops/runs/1 | Returns run + iterations |
| GET /api/v1/workspaces/1/agent-sessions | Returns session history |
| Agent Dashboard page loads | Shows active runs and cards |
| Loop Run Detail page | Shows iterations timeline with live polling |

---

> **Next Phase:** Phase 2 (Harness Engine) — upgrade Loop to multi-agent Pipeline with Planner→Executor→Reviewer + adversarial verification. See [design spec](../specs/2026-07-18-reqmango-agent-platform-design.md#72-phase-2-harness-engine-m4-m6).
