# Workflow Model Separation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give `workflows` and `agent_workflows` exactly one owner each, drop the `state_transitions.workflow_kind` discriminator, and delete the code paths that conflated them.

**Architecture:** `workflows` + `state_transitions` become the state machine; `agent_workflows` + nodes/edges/runs become orchestration. The FK from transitions to workflows is restored, the `kind` argument disappears from every signature, and the 1,257-line `WorkflowService` splits along the table boundary. A future transition→run link uses the existing `trigger_type=event` contract, not a shared row.

**Tech Stack:** Go 1.25, Gin, GORM, PostgreSQL 16/18, Vue 3 + TypeScript, golang-migrate.

**Spec:** `docs/superpowers/specs/2026-10-18-workflow-model-separation-design.md`

## Global Constraints

- Migrations live in `backend/migrations/`; the current baseline is `000001_baseline.up.sql`, so the new file is `000002_separate_state_transitions.up.sql`.
- `cmd/server` runs `AutoMigrate` on every start. Any model change must leave that path and the SQL path converging on the same schema.
- `cmd/route_audit` must stay green in all four modes (`base`, `no-query`, `member`, `cross`) with `-strict`; it is a CI gate.
- Do not remove or rename `agent_workflows`, `workflow_nodes`, `workflow_edges`, `workflow_runs`, `workflow_node_runs`.
- Frontend types are strict: `vue-tsc --noEmit` must stay at 0 errors, and the vitest suite must stay green.
- Commit after every task. Run `gofmt -w` on every touched Go file before committing; the working tree is CRLF, so use the content diff, not `gofmt -l`, as the signal.

## Preconditions

Task 0 landed in `21ba91f`: workspace routes now bind `workflows` / `state`, the workspace list and detail filter by kind, `WorkflowDetail.vue` is gone, and ProjectSettings opens the designer. Tasks below start from that commit.

---

### Task 1: Retire the `workflow_kind` discriminator

**Files:**
- Create: `backend/migrations/000002_separate_state_transitions.up.sql`
- Create: `backend/migrations/000002_separate_state_transitions.down.sql`
- Modify: `backend/internal/model/state.go:28-66`
- Modify: `backend/internal/service/workflow_service.go` (transition methods)
- Modify: `backend/internal/handler/workflow_handler.go`
- Modify: `backend/internal/router/router.go:1137-1160`
- Modify: `backend/internal/middleware/access.go:133-215`
- Modify: `backend/cmd/server/main.go:210-222`
- Modify: `backend/cmd/route_audit/main.go:353,492`
- Test: `backend/internal/service/workflow_transition_kind_test.go`
- Test: `backend/internal/middleware/workflow_access_test.go`
- Test: `backend/internal/workflow/rules_test.go`

**Interfaces:**
- Consumes: nothing new.
- Produces:
  - `func (s *WorkflowService) ListTransitions(workflowID uint64) ([]StateTransitionResponse, error)`
  - `func (s *WorkflowService) AddTransition(workflowID uint64, req AddTransitionRequest) (*StateTransitionResponse, error)`
  - `func (s *WorkflowService) UpdateTransition(transitionID uint64, req UpdateTransitionRequest) error`
  - `func (s *WorkflowService) DeleteTransition(transitionID uint64) error`
  - `StateTransition` has no `WorkflowKind` field and no `TransitionKind*` constants exist.
  - `GET/POST/PUT/DELETE /api/v1/projects/:projectId/workflows/:workflowId/transitions` return 404 (routes removed).

- [ ] **Step 1: Write the failing schema test**

Add to `backend/internal/service/workflow_transition_kind_test.go` (replace the file's kind-specific assertions with ownership assertions):

```go
func TestStateTransitionHasOneParent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Workflow{}, &model.StateTransition{}))

	wf := model.Workflow{Name: "state", WorkspaceID: 10}
	require.NoError(t, db.Create(&wf).Error)
	tr := model.StateTransition{Name: "edge", WorkflowID: wf.ID, SourceStateID: 1, TargetStateID: 2, WorkspaceID: 10}
	require.NoError(t, db.Create(&tr).Error)

	var stored model.StateTransition
	require.NoError(t, db.First(&stored, tr.ID).Error)
	assert.Equal(t, wf.ID, stored.WorkflowID)

	// The column is gone: inserting a value for it must fail.
	err = db.Exec("INSERT INTO state_transitions (workflow_kind) VALUES ('agent')").Error
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd backend && go test ./internal/service/ -run TestStateTransitionHasOneParent -count=1 -v`
Expected: FAIL at compile time — `model.StateTransition` still has `WorkflowKind`, so the test's premise (column absent) does not hold yet. That is the intended failure; proceed.

- [ ] **Step 3: Write the migration**

`backend/migrations/000002_separate_state_transitions.up.sql`:

```sql
-- Preserve agent-kind rows before removing the discriminator. The live table
-- keeps the state machine only; agent orchestration has its own children.
CREATE TABLE IF NOT EXISTS state_transitions_agent_archive (
    LIKE state_transitions INCLUDING DEFAULTS
);
ALTER TABLE state_transitions_agent_archive
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ DEFAULT now();

INSERT INTO state_transitions_agent_archive (
    id, created_at, updated_at, deleted_at, created_by_id, updated_by_id,
    name, description, workflow_id, workflow_kind, source_state_id, target_state_id,
    project_id, workspace_id, rule_type, approver_ids, role_allowed,
    approve_target_state_id, reject_target_state_id, approval_mode
)
SELECT
    id, created_at, updated_at, deleted_at, created_by_id, updated_by_id,
    name, description, workflow_id, workflow_kind, source_state_id, target_state_id,
    project_id, workspace_id, rule_type, approver_ids, role_allowed,
    approve_target_state_id, reject_target_state_id, approval_mode
FROM state_transitions
WHERE workflow_kind = 'agent';

DELETE FROM state_transitions WHERE workflow_kind = 'agent';

ALTER TABLE state_transitions DROP COLUMN IF EXISTS workflow_kind;

DO $migration$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_state_transitions_workflow'
    ) THEN
        ALTER TABLE state_transitions
            ADD CONSTRAINT fk_state_transitions_workflow
            FOREIGN KEY (workflow_id) REFERENCES workflows(id);
    END IF;
END
$migration$;
```

Confirm the archive column list matches the live table before committing:

```bash
docker exec reqmango-pg psql -U reqmango -d reqmango_test -At -c \
  "SELECT column_name FROM information_schema.columns WHERE table_name='state_transitions' ORDER BY ordinal_position"
```

`backend/migrations/000002_separate_state_transitions.down.sql`:

```sql
ALTER TABLE state_transitions ADD COLUMN IF NOT EXISTS workflow_kind VARCHAR(16) NOT NULL DEFAULT 'state';
ALTER TABLE state_transitions DROP CONSTRAINT IF EXISTS fk_state_transitions_workflow;

INSERT INTO state_transitions (
    id, created_at, updated_at, deleted_at, created_by_id, updated_by_id,
    name, description, workflow_id, workflow_kind, source_state_id, target_state_id,
    project_id, workspace_id, rule_type, approver_ids, role_allowed,
    approve_target_state_id, reject_target_state_id, approval_mode
)
SELECT
    id, created_at, updated_at, deleted_at, created_by_id, updated_by_id,
    name, description, workflow_id, workflow_kind, source_state_id, target_state_id,
    project_id, workspace_id, rule_type, approver_ids, role_allowed,
    approve_target_state_id, reject_target_state_id, approval_mode
FROM state_transitions_agent_archive
ON CONFLICT (id) DO NOTHING;

DROP TABLE IF EXISTS state_transitions_agent_archive;
```

- [ ] **Step 4: Remove the field and the backfill**

In `backend/internal/model/state.go`, delete `WorkflowKind` and the `TransitionKindState` / `TransitionKindAgent` constants. Restore the relationship so GORM maintains the FK:

```go
	Workflow Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
	SourceState State `gorm:"foreignKey:SourceStateID" json:"-"`
	TargetState State `gorm:"foreignKey:TargetStateID" json:"-"`
```

In `backend/cmd/server/main.go`, delete the whole `// Backfill state_transitions.workflow_kind...` block (both `db.Exec` statements).

- [ ] **Step 5: Remove `kind` from every call site**

`backend/internal/service/workflow_service.go`:

```go
func (s *WorkflowService) ListTransitions(workflowID uint64) ([]StateTransitionResponse, error) {
	var transitions []model.StateTransition
	if err := s.db.Where("workflow_id = ? AND deleted_at IS NULL", workflowID).
		Order("id ASC").Find(&transitions).Error; err != nil {
		return nil, err
	}
	result := make([]StateTransitionResponse, len(transitions))
	for i, t := range transitions {
		result[i] = s.toTransitionResponse(&t)
	}
	return result, nil
}

func (s *WorkflowService) AddTransition(workflowID uint64, req AddTransitionRequest) (*StateTransitionResponse, error) {
	if err := req.normalize(); err != nil {
		return nil, err
	}
	workspaceID, projectID, err := s.resolveWorkflowScope(workflowID)
	// ... unchanged body, but transition.WorkflowKind is not set
}

func (s *WorkflowService) UpdateTransition(transitionID uint64, req UpdateTransitionRequest) error {
	// ... unchanged, Where("id = ?", transitionID)
}

func (s *WorkflowService) DeleteTransition(transitionID uint64) error {
	result := s.db.Where("id = ?", transitionID).Delete(&model.StateTransition{})
	if result.RowsAffected == 0 {
		return errors.New("transition not found")
	}
	return result.Error
}
```

`resolveWorkflowScope` loses its kind branch and always reads `workflows`:

```go
func (s *WorkflowService) resolveWorkflowScope(workflowID uint64) (uint64, *uint64, error) {
	var sw model.Workflow
	if err := s.db.First(&sw, workflowID).Error; err != nil {
		return 0, nil, errors.New("workflow not found")
	}
	return sw.WorkspaceID, sw.ProjectID, nil
}
```

Also in the same file:

- `toStateWorkflowResponse`: `Where("workflow_id = ? AND deleted_at IS NULL", wf.ID)` (drop the kind predicate added in `21ba91f`).
- `DeleteStateWorkflow`: `Where("workflow_id = ?", workflowID)`.
- `WorkflowDetail.Transitions` and its population in `Get`: remove both.
- `StateTransitionResponse.WorkflowKind`: remove.

`backend/internal/handler/workflow_handler.go`:

- Delete `listTransitions/addTransition/updateTransition/deleteTransition` wrappers and the agent variants.
- Keep one method each that calls the new signatures. The workspace and project route groups no longer share transition handlers; project transition routes are removed, so only the workspace methods remain:

```go
func (h *WorkflowHandler) ListWorkspaceTransitions(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}
	transitions, err := h.workflowSvc.ListTransitions(workflowID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, transitions)
}
```

Do the same for `AddWorkspaceTransition`, `UpdateWorkspaceTransition`, `DeleteWorkspaceTransition`. Delete the project-level `ListTransitions`, `AddTransition`, `UpdateTransition`, `DeleteTransition`.

`backend/internal/router/router.go`:

- Delete these four lines from the project group (around 1145-1147):

```go
workflows.GET("/:workflowId/transitions", ...)
workflows.POST("/:workflowId/transitions", ...)
workflows.PUT("/:workflowId/transitions/:transitionId", ...)
workflows.DELETE("/:workflowId/transitions/:transitionId", ...)
```

- Keep the node edge routes (`/nodes`, `/edges` with `AddEdge`).
- Workspace group keeps `transitions` and the `edges` alias, pointing at the `...WorkspaceTransition` methods.

`backend/internal/middleware/access.go`:

- `RequireProjectWorkflow`: delete the `strings.Contains(c.FullPath(), "/transitions")` branch; validate nodes/edges/runs only.
- `RequireWorkspaceWorkflow`: workflow check in `workflows`, transition check in `state_transitions` by `workflow_id` only (no kind).
- `transitionChildOf` loses its `kind` parameter:

```go
func transitionChildOf(db *gorm.DB, transitionID, workflowID uint64) bool {
	if transitionID == 0 {
		return true
	}
	var n int64
	db.Model(&model.StateTransition{}).
		Where("id = ? AND workflow_id = ?", transitionID, workflowID).Count(&n)
	return n > 0
}
```

- Remove the now-unused `strings` import if nothing else uses it.

`backend/cmd/route_audit/main.go`:

- Line ~353: `"workflows": {"workflows"}` — the segment now maps to one table.
- Line ~492: replace the union with `SELECT id FROM workflows WHERE workspace_id = ?`.

- [ ] **Step 6: Rewrite the affected tests**

`workflow_transition_kind_test.go`: delete `TestTransitionKindSeparation` (the discriminator no longer exists) and keep the ownership tests from Step 1 plus the read-path test renamed to `TestStateWorkflowReadPathsStayInWorkflows`, with kind fields removed from fixtures.

`workflow_access_test.go`:

- Delete the project transition route registrations and their assertions.
- Remove `WorkflowKind` from every `StateTransition` literal.
- Add:

```go
assert.Equal(t, 404, status("bob", "POST", p+uitoa(awNew.ID)+"/transitions"), "project transitions are gone")
assert.Equal(t, 200, status("bob", "POST", "/workspaces/10/workflows/"+uitoa(swWs.ID)+"/transitions"))
assert.Equal(t, 404, status("bob", "POST", "/workspaces/10/workflows/"+uitoa(swWs.ID)+"/transitions/"+uitoa(tForeign.ID)))
```

`backend/internal/workflow/rules_test.go`: remove `WorkflowKind: model.TransitionKindAgent` from the "ignores agent workflow transitions with a colliding id" fixture. The test's purpose disappears with the discriminator; replace it with a test that a transition whose `workflow_id` has no `workflows` row is ignored:

```go
func TestLoad_IgnoresTransitionsWithoutAWorkflowRow(t *testing.T) {
	// fixture with a transition whose workflow_id has no workflows row
	rules, err := Load(f.db, f.project.ID, nil)
	require.NoError(t, err)
	assert.False(t, rules.Restricts, "an orphan transition must not constrain anything")
}
```

- [ ] **Step 7: Run the backend suites**

Run: `cd backend && go build ./... && go test ./internal/service/ ./internal/middleware/ ./internal/workflow/ ./internal/handler/ ./cmd/route_audit/ -count=1`
Expected: PASS. Fix compile errors from removed signatures rather than reintroducing the parameter.

- [ ] **Step 8: Verify the migration both ways**

On a scratch database (do not use `reqmango`):

```bash
createdb -h 127.0.0.1 -U reqmango reqmango_mig_smoke
DATABASE_URL='postgres://reqmango:reqmango@127.0.0.1:5432/reqmango_mig_smoke?sslmode=disable' go run ./cmd/migrate up
DATABASE_URL='...' psql -At -c "SELECT count(*) FROM information_schema.columns WHERE table_name='state_transitions' AND column_name='workflow_kind'"
DATABASE_URL='...' go run ./cmd/migrate down 1
DATABASE_URL='...' psql -At -c "SELECT count(*) FROM information_schema.columns WHERE table_name='state_transitions' AND column_name='workflow_kind'"
```

Expected: `0` after up, `1` after down. If the role cannot create a database, use a dedicated schema (`search_path`) or ask for a scratch instance; do not point the migration at `reqmango`.

- [ ] **Step 9: Commit**

```bash
gofmt -w backend/internal/model/state.go backend/internal/service/workflow_service.go backend/internal/handler/workflow_handler.go backend/internal/router/router.go backend/internal/middleware/access.go backend/cmd/server/main.go backend/cmd/route_audit/main.go
git add backend/migrations/000002_*.sql backend/internal backend/cmd
git commit -m "refactor(workflows): give state_transitions one parent"
```

---

### Task 2: Split the state-machine half out of `WorkflowService`

**Files:**
- Create: `backend/internal/service/state_workflow_service.go`
- Modify: `backend/internal/service/workflow_service.go` (shrinks to agent orchestration)
- Modify: `backend/internal/service/workflow_executor.go:26` (field type)
- Modify: `backend/internal/router/router.go:137-161`
- Modify: `backend/internal/handler/workflow_handler.go`
- Test: `backend/internal/service/state_workflow_service_test.go`

**Interfaces:**
- Consumes: the `WorkflowService` transition signatures from Task 1.
- Produces:
  - `type StateWorkflowService struct{ db *gorm.DB }`
  - `func NewStateWorkflowService(db *gorm.DB) *StateWorkflowService`
  - Methods moved verbatim: `ResolveWorkspaceID`, `ListStateWorkflowsByWorkspace`, `GetStateWorkflow`, `CreateStateWorkflow`, `UpdateStateWorkflow`, `DeleteStateWorkflow`, `ListTransitions`, `AddTransition`, `UpdateTransition`, `DeleteTransition`, `resolveWorkflowScope`, `toStateWorkflowResponse`, `toTransitionResponse`, `parseIssueTypeIDs`, `encodeIssueTypeIDs`.
  - DTOs stay in a shared file: `StateWorkflowResponse`, `StateTransitionUIItem`, `StateTransitionResponse`, `AddTransitionRequest`, `UpdateTransitionRequest`, `CreateStateWorkflowRequest`, `UpdateStateWorkflowRequest`.

- [ ] **Step 1: Write the failing wiring test**

`backend/internal/service/state_workflow_service_test.go`:

```go
func TestStateWorkflowServiceOwnsTransitions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Workflow{}, &model.StateTransition{}, &model.Workspace{}))

	svc := NewStateWorkflowService(db)
	require.NotNil(t, svc)

	wf, err := svc.CreateStateWorkflow(10, nil, CreateStateWorkflowRequest{Name: "flow"})
	require.NoError(t, err)

	created, err := svc.AddTransition(wf.ID, AddTransitionRequest{SourceStateID: 1, TargetStateID: 2})
	require.NoError(t, err)
	assert.Equal(t, wf.ID, created.WorkflowID)

	list, err := svc.ListTransitions(wf.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd backend && go test ./internal/service/ -run TestStateWorkflowServiceOwnsTransitions -count=1`
Expected: FAIL — `undefined: NewStateWorkflowService`.

- [ ] **Step 3: Move the methods**

Create `state_workflow_service.go` with the struct, constructor, and the method list from **Produces**, copied from `workflow_service.go` with the receiver changed. Move the string/JSON helpers with them. Keep the DTOs in `workflow_service.go` (they are shared by both handlers) or move them to `workflow_dto.go`; either is fine as long as both services compile.

- [ ] **Step 4: Point handlers at both services**

`workflow_handler.go`:

```go
type WorkflowHandler struct {
	stateSvc *service.StateWorkflowService
	agentSvc *service.WorkflowService
	executor *service.WorkflowExecutor
}

func NewWorkflowHandler(stateSvc *service.StateWorkflowService, agentSvc *service.WorkflowService, executor *service.WorkflowExecutor) *WorkflowHandler {
	return &WorkflowHandler{stateSvc: stateSvc, agentSvc: agentSvc, executor: executor}
}
```

Mechanical replacements:

| Handler | Service |
|---------|---------|
| `ListWorkspaceWorkflows`, `CreateWorkspaceWorkflow`, `GetStateWorkflow`, `UpdateWorkspaceWorkflow`, `DeleteWorkspaceWorkflow`, `*WorkspaceTransition` | `h.stateSvc` |
| `ListWorkflows`, `CreateWorkflow`, `GetWorkflow`, `UpdateWorkflow`, `DeleteWorkflow`, `AddNode`, `UpdateNode`, `DeleteNode`, `AddEdge`, `UpdateEdge`, `DeleteEdge`, `ExecuteWorkflow`, `ListRuns`, `GetRun`, `CancelRun` | `h.agentSvc` |

`router.go`:

```go
stateWorkflowSvc := service.NewStateWorkflowService(db)
agentWorkflowSvc := service.NewWorkflowService(db, contextPayloadSvc, agentDecisionSvc, agentBudgetSvc)
workflowExecutor := service.NewWorkflowExecutor(db, agentWorkflowSvc, contextPayloadSvc, agentDecisionSvc, agentBudgetSvc, agentSLASvc)
workflowH := handler.NewWorkflowHandler(stateWorkflowSvc, agentWorkflowSvc, workflowExecutor)
```

`workflow_executor.go`: no change. The executor consumes only the orchestration half (`GetByNodeID`, `GetRun`, run/node-run updates), which stays on `WorkflowService`. Renaming that type to `AgentWorkflowService` is deliberately out of scope for this task, to keep the diff mechanical.

- [ ] **Step 5: Run everything**

Run: `cd backend && go build ./... && go test ./internal/service/ ./internal/handler/ ./internal/router/ -count=1`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
gofmt -w backend/internal/service/state_workflow_service.go backend/internal/service/workflow_service.go backend/internal/handler/workflow_handler.go backend/internal/router/router.go backend/internal/service/workflow_executor.go
git add backend/internal
git commit -m "refactor(workflows): separate the state-machine service from orchestration"
```

---

### Task 3: Remove the frontend paths that assumed one workflow type

**Files:**
- Delete: `frontend/src/components/StateTransition.vue`
- Modify: `frontend/src/api/workflow.ts:218-233` (remove the four state-transition helpers)
- Modify: `frontend/src/api/workflow.api.test.ts` (remove their tests)
- Modify: `frontend/src/types/workflow.ts` (remove `StateTransition*` interfaces)
- Rename: `frontend/src/components/WorkflowManager.vue` → `frontend/src/components/StateWorkflowManager.vue`
- Modify: `frontend/src/views/WorkspaceSettings.vue:17,368`
- Test: `frontend/src/__tests__/StateWorkflowManager.spec.ts` (new)

**Interfaces:**
- Consumes: the workspace routes from Task 0/1 (`/workspaces/:ws/workflows...`).
- Produces: `StateWorkflowManager` with props `{ workspaceId: number; canManage?: boolean }`; it no longer accepts `projectId`.

- [ ] **Step 1: Write the failing component test**

`frontend/src/__tests__/StateWorkflowManager.spec.ts`:

```ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import StateWorkflowManager from '@/components/StateWorkflowManager.vue'

const get = vi.fn()
vi.mock('@/api', () => ({ default: { get: (...a: any[]) => get(...a) } }))
vi.mock('@/composables/useI18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ user: { id: 1, is_superuser: true } }) }))

describe('StateWorkflowManager', () => {
  beforeEach(() => { get.mockReset() })

  it('only lists workspace workflows', async () => {
    get.mockImplementation((url: string) => {
      if (url.endsWith('/workflows')) return Promise.resolve({ data: [] })
      if (url.endsWith('/settings/states')) return Promise.resolve({ data: { data: [] } })
      if (url === '/issue-types') return Promise.resolve({ data: { data: [] } })
      return Promise.resolve({ data: [] })
    })
    mount(StateWorkflowManager, { props: { workspaceId: 1 } })
    await flushPromises()
    expect(get).toHaveBeenCalledWith('/workspaces/1/workflows')
    expect(get.mock.calls.some(([u]) => String(u).includes('/projects/'))).toBe(false)
  })
})
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd frontend && npx vitest run src/__tests__/StateWorkflowManager.spec.ts`
Expected: FAIL — module `@/components/StateWorkflowManager.vue` not found.

- [ ] **Step 3: Rename and delete**

```bash
git mv frontend/src/components/WorkflowManager.vue frontend/src/components/StateWorkflowManager.vue
git rm frontend/src/components/StateTransition.vue
```

In the renamed file:

- Props become `{ workspaceId: number; canManage?: boolean }`.
- Delete `isWorkspaceMode`, the `projectId` branches in `load`, `save`, `confirmDel`, `saveTrans`, `delTrans`, and the `workflowApi` import.
- `load` always calls `listWorkspaceWorkflows(props.workspaceId)` and `/workspaces/${props.workspaceId}/settings/states`.
- `canManageWorkflows` keeps the superuser/role check and drops the `can('workflow:manage')` branch.

In `WorkspaceSettings.vue`, update the import and the tag. In `api/workflow.ts`, delete the four helpers under `// ==================== State Transitions ====================`; in the test file, delete the matching `it(...)` blocks; in `types/workflow.ts`, delete the three `StateTransition*` interfaces.

- [ ] **Step 4: Run frontend checks**

Run: `cd frontend && npx vue-tsc --noEmit && npx vitest run`
Expected: 0 type errors; all tests pass. The suite has one timing-sensitive test (`WorkflowsPage`), which passes alone; if it times out only in the full run, re-run it isolated before treating it as a regression.

- [ ] **Step 5: Commit**

```bash
git add frontend
git commit -m "refactor(workflows): drop the frontend paths for a single workflow type"
```

---

### Task 4: Update the documentation and re-run the gates

**Files:**
- Modify: `docs/security.md` §11
- Modify: `docs/bug-list.md` BUG-58
- Modify: `docs/superpowers/specs/2026-10-18-workflow-model-separation-design.md` (status line)
- Modify: `docs/dev/pipeline-status.md` (entry for this work)

- [ ] **Step 1: Rewrite `docs/security.md` §11**

Replace the "stopgap" framing with ownership: one table per domain, transitions FK to `workflows`, no discriminator. Keep the historical paragraph explaining why the column existed, wrapped in past tense.

- [ ] **Step 2: Update BUG-58**

Add a third-round line: the workspace detail/transition 404s, the unfiltered list, and `WorkflowDetail.vue`'s state-as-node ids; link to the spec.

- [ ] **Step 3: Mark the spec implemented**

Change the header to `Status: Implemented` and add the commit range.

- [ ] **Step 4: Run the gates locally**

```bash
cd backend && go build ./... && go test ./... -count=1
cd ../frontend && npx vue-tsc --noEmit && npx vitest run
```

Then push and confirm the GitHub Actions run is green in all seven jobs: Lint, Security, Test, E2E, Tenancy audit, Migration path, Build.

- [ ] **Step 5: Commit**

```bash
git add docs
git commit -m "docs(workflows): record the single-owner model"
```

---

## Self-Review Notes

- Every spec section maps to a task: §4 model → Task 1; §5.2 service split → Task 2; §5.3/5.4 routes and middleware → Tasks 1-2; §6 migration → Task 1; §7 frontend → Task 3; §8 testing → embedded per task plus Task 4; §9 rollout → Task 0 (done) + tasks above; §10 non-goals → not implemented by design.
- `WorkflowDetail.Transitions` removal is covered in Task 1 Step 5; the frontend type never declared it, so Task 3 has nothing to change there.
- The migration's column list must be checked against a real database before commit (Task 1 Step 3); this is the only place where a wrong assumption would silently lose data.
