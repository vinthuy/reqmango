# Phase 2: Harness Engine — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development.

**Goal:** Build the multi-agent Harness Engine — Pipeline orchestration, adversarial verification, Worktree isolation, and model routing.

**Architecture:** New `backend/internal/agent/harness/` package with pipeline, planner, executor, reviewer, router, and worktree modules. Extends existing Loop engine from Phase 1. Pipeline API + Registry routes. Frontend Pipeline Builder DAG view.

**Tech Stack:** Go 1.21+ + Gin + GORM + PostgreSQL | Vue 3 + TypeScript + Cytoscape.js (DAG) | YAML v3 parser

**Prerequisites:** Phase 1 Loop Engine (completed)

---

## Task 1: Pipeline Engine Core

**Files:**
- Create: `backend/internal/agent/harness/pipeline.go`
- Create: `backend/internal/agent/harness/pipeline_test.go`

Implement the Pipeline orchestrator:

```go
// pipeline.go
package harness

type PipelineMode string
const (
    ModeSequential PipelineMode = "sequential"  // Planner → Executor → Reviewer
    ModeFanOut     PipelineMode = "fan_out"     // Planner → [E1, E2, E3] → Reviewer
    ModeTournament PipelineMode = "tournament"  // Planner → [E1, E2, E3] → Judge
    ModeClassify   PipelineMode = "classify"    // Classifier → Routed Executor
)

type StageType string
const (
    StagePlanner  StageType = "planner"
    StageExecutor StageType = "executor"
    StageReviewer StageType = "reviewer"
    StageJudge    StageType = "judge"
)

type StageConfig struct {
    Name       string
    StageType  StageType
    AgentID    uint64
    Model      string
    Effort     string
    Mode       string // "adversarial" for reviewer stages
    InputFrom  string // which stage output to use as input
}

type PipelineConfig struct {
    Name        string
    Mode        PipelineMode
    Stages      []StageConfig
    MaxRetries  int
    Budget      BudgetConfig
}

type StageResult struct {
    StageName   string
    Output      string
    TokensUsed  int
    Cost        float64
    Verdicts    []ReviewVerdict // for adversarial stages
    Error       error
}

type PipelineRunner struct { ... }
func (r *PipelineRunner) Run(ctx context.Context, config PipelineConfig, input map[string]interface{}) ([]StageResult, error)
```

Key methods: `RunSequential`, `RunFanOut`, `RunTournament`, `RunClassify`.

Commit: `feat(harness): add Pipeline engine with 4 orchestration modes`

## Task 2: Adversarial Verification

**Files:**
- Create: `backend/internal/agent/harness/reviewer.go`
- Create: `backend/internal/agent/harness/reviewer_test.go`

Implement 3-vote adversarial review:

```go
type ReviewVerdict struct {
    Pass       bool     `json:"pass"`
    Issues     []Issue  `json:"issues"`
    Confidence float64  `json:"confidence"`
}

type AdversarialReviewer struct { ... }

// RunReview spawns 3 independent reviewer agents, each prompted to REFUTE.
// Returns: passes if >= 2 reviewers approve.
func (r *AdversarialReviewer) RunReview(ctx, spec, artifact) ([]ReviewVerdict, bool, error)

// buildRefutePrompt creates the "Try to REFUTE" prompt template
func buildRefutePrompt(spec, artifact string) string
```

Default prompt: "You are a QA Reviewer. CRITICAL: Assume this output contains errors. FIND them. Default to REFUTED."

Commit: `feat(harness): add adversarial verification with 3-vote protocol`

## Task 3: Worktree Isolation

**Files:**
- Create: `backend/internal/agent/harness/worktree.go`

```go
type WorktreePool struct {
    repoPath   string
    maxWorktrees int
    sem        chan struct{}
}

func (p *WorktreePool) Acquire(ctx context.Context) (*Worktree, error)
func (p *WorktreePool) Release(wt *Worktree) error
func (p *WorktreePool) Cleanup() error
```

Uses `git worktree add` / `git worktree remove`. Max 16 concurrent (CPU-bound cap).

Commit: `feat(harness): add Git Worktree pool for parallel agent isolation`

## Task 4: Dynamic Model Router

**Files:**
- Create: `backend/internal/agent/harness/router.go`

Maps task complexity → model + effort:
```
Planner     → Opus/Fable + xhigh
Executor    → Sonnet + high (simple → Haiku + medium)
Reviewer    → Sonnet + high
Classifier  → Haiku + low
Judge       → Opus/Sonnet + high
```

Commit: `feat(harness): add dynamic model router by task type and complexity`

## Task 5: Harness DSL Parser

**Files:**
- Create: `backend/internal/agent/harness/dsl.go`
- Create: `backend/internal/agent/harness/dsl_test.go`

Parse YAML Pipeline definitions (like sprint-review.yml from spec §5.6).

```go
type PipelineDSL struct {
    Name        string                 `yaml:"name"`
    Description string                 `yaml:"description"`
    Trigger     TriggerDSL             `yaml:"trigger"`
    Pipeline    PipelineStagesDSL      `yaml:"pipeline"`
    Retry       RetryDSL               `yaml:"retry"`
    Budget      BudgetDSL              `yaml:"budget"`
}

func ParsePipelineDSL(yamlBytes []byte) (*PipelineDSL, error)
func (d *PipelineDSL) ToPipelineConfig() (*PipelineConfig, error)
```

Commit: `feat(harness): add YAML Pipeline DSL parser`

## Task 6: Agent Registry

**Files:**
- Create: `backend/internal/agent/registry/registry.go`
- Create: `backend/internal/agent/registry/capabilities.go`
- Create: `backend/internal/agent/model/registry.go` (GORM model)

```go
type AgentEntry struct {
    BaseModel
    WorkspaceID  *uint64        // nil = global
    Name         string
    Capabilities []string       // ["triage", "planning", "code_review", ...]
    AgentDef     json.RawMessage
    Version      string
    IsVerified   bool
    Installs     int
    Rating       float64
    Source       string         // local | marketplace
}

type Registry struct { db *gorm.DB }

func (r *Registry) Register(entry *AgentEntry) error
func (r *Registry) Find(capabilities ...string) ([]AgentEntry, error)
func (r *Registry) GetByID(id uint64) (*AgentEntry, error)
func (r *Registry) ListByWorkspace(wsID uint64) ([]AgentEntry, error)
```

Commit: `feat(harness): add Agent Registry with capability-based discovery`

## Task 7: Pipeline API + Routes + DB Models

**Files:**
- Create: `backend/internal/agent/model/pipeline.go` (Pipeline, PipelineRun GORM models)
- Create: `backend/internal/handler/agent_pipeline.go`
- Modify: `backend/internal/router/router.go` and `backend/cmd/server/main.go`

Add Pipeline CRUD + run endpoints:
```
POST   /api/v1/workspaces/:ws/pipelines
GET    /api/v1/workspaces/:ws/pipelines
GET    /api/v1/workspaces/:ws/pipelines/:id
PUT    /api/v1/workspaces/:ws/pipelines/:id
DELETE /api/v1/workspaces/:ws/pipelines/:id
POST   /api/v1/workspaces/:ws/pipelines/:id/run
GET    /api/v1/workspaces/:ws/pipelines/:id/runs
GET    /api/v1/workspaces/:ws/pipelines/runs/:runId
```

+ Agent Registry routes:
```
GET    /api/v1/workspaces/:ws/agents/registry
POST   /api/v1/workspaces/:ws/agents/registry
```

Auto-migrate Pipeline/PipelineRun/AgentEntry models.

Commit: `feat(harness): add Pipeline API, Registry API, routes, and data models`

## Task 8: Frontend Pipeline Builder

**Files:**
- Create: `frontend/src/api/agent-pipeline.ts`
- Create: `frontend/src/api/agent-registry.ts`
- Create: `frontend/src/stores/agentPipeline.ts`
- Create: `frontend/src/views/agents/PipelineBuilder.vue`
- Create: `frontend/src/views/agents/PipelineRunDetail.vue`
- Create: `frontend/src/components/agents/PipelineDAG.vue`
- Modify: `frontend/src/router/index.ts`

Pipeline Builder with DAG visualization (simple SVG-based, no heavy dependency), stage configuration, and run button.

Commit: `feat(frontend): add Pipeline Builder with DAG visualization`

## Task 9: Integration Tests + Build Verification

Run all tests, build backend + frontend, final commit.

Commit: `feat: Phase 2 Harness Engine — complete implementation`
