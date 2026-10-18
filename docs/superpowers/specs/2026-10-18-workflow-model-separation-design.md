# Workflow Model Separation Design

> Date: 2026-10-18
> Status: Approved (design), implementation not started
> Owner: reqmango team
> Supersedes part of: `docs/superpowers/specs/2026-07-19-workflow-approval-design.md`
> (that spec's approval flow stays valid; only table ownership changes)

## 1. Problem

Two unrelated concepts are both called "workflow" and both own rows in
`state_transitions`:

| Concept | Master table | Child tables | Purpose |
|---------|--------------|--------------|---------|
| State machine | `workflows` | `state_transitions` | Decide whether an issue may move between states (allow / approval) |
| Agent orchestration | `agent_workflows` | `workflow_nodes`, `workflow_edges`, `workflow_runs` | Execute a DAG of agents (manual / event / cron / webhook) |

`state_transitions.workflow_id` used to point at either table. Their id ranges
overlap, so the same id resolved two different workflows. `workflow_kind`
(`state` / `agent`) was added in October 2026 to disambiguate, which fixed the
worst symptom but left the underlying problem: a child table with two possible
parents, and a `kind` argument every caller must remember to pass.

Callers did not remember. Three defects are live today.

## 2. Current state (evidence)

### 2.1 Only the state machine has real data and readers

| Table | Rows in `reqmango_test` | Rows in `reqmango_audit` |
|-------|------------------------|--------------------------|
| `workflows` | 1 | 1 |
| `state_transitions` (`state`) | 10 | 10 |
| `agent_workflows` | 0 | 1 |
| `state_transitions` (`agent`) | 0 | 0 |
| `workflow_nodes` / `workflow_edges` / `workflow_runs` | 0 | — |

The production database was not reachable during this review (the local
PostgreSQL 18 service was stopped and the Docker instance only hosts the test
and audit databases). The migration therefore archives rows instead of
assuming the count is zero; see §6.

Live readers of the state machine:

- `backend/internal/service/issue_service.go` (`workflow.Load` / `Check`) —
  rejects illegal issue state changes.
- `backend/internal/ai/service/ai_service.go` (`allowed_next`) — tells the
  model which moves are legal.
- `frontend/src/views/WorkspaceSettings.vue` → `WorkflowManager` — the only
  place transitions are configured.
- `IssueDetail.vue` / `IssueDetailPanel.vue` read `/workspaces/:ws/workflows`
  to render the state picker.

Agent orchestration has **no** reader of state transitions: `WorkflowsPage.vue`
and `WorkflowDesigner.vue` do not mention transitions, and `StateTransition.vue`
— the only component that did — is not imported anywhere.

### 2.2 Defect 1: every workspace workflow detail route is 404

`RequireWorkspaceWorkflow` validates `:workflowId` against `agent_workflows`
(`backend/internal/middleware/access.go:207-215`), but the routes it guards
address `workflows`:

```go
// backend/internal/router/router.go:602-608
workspaces.GET("/:wsParam/workflows/:workflowId", wsWorkflow, workflowH.GetWorkflow)
workspaces.PUT("/:wsParam/workflows/:workflowId", wsWorkflow, workflowH.UpdateWorkspaceWorkflow)
workspaces.DELETE("/:wsParam/workflows/:workflowId", wsWorkflow, workflowH.DeleteWorkspaceWorkflow)
workspaces.GET("/:wsParam/workflows/:workflowId/transitions", wsWorkflow, workflowH.ListTransitions)
```

A workspace state workflow id never matches `agent_workflows`, so read, update,
delete and transition routes all return 404. Only list and create work.

### 2.3 Defect 2: transitions are created with the wrong kind

The same handler serves both route groups and hardcodes the agent kind:

```go
// backend/internal/handler/workflow_handler.go:449
transition, err := h.workflowSvc.AddTransition(model.TransitionKindAgent, workflowID, req)
```

`resolveWorkflowScope(kind=agent, id)` then looks the id up in
`agent_workflows`. If an agent workflow happens to share the id, the transition
is written with that workflow's project and workspace — a cross-project
misattribution. Otherwise the request fails with `workflow not found`.

### 2.4 Defect 3: the workspace list still mixes both kinds

```go
// backend/internal/service/workflow_service.go:837
s.db.Where("workflow_id = ? AND deleted_at IS NULL", wf.ID).Order("id ASC").Find(&transitions)
```

No `workflow_kind` filter, so `WorkflowManager` can show agent transitions under
a state workflow — the exact class of bug `workflow_kind` was introduced to fix.

### 2.5 Defect 4: the agent workflow detail page edits state transitions

`/workspace/:slug/project/:id/settings/workflows/:workflowId` is reachable from
`ProjectSettings.vue` (`handleViewWorkflow`) and renders `WorkflowDetail.vue`.
That page is a state-machine editor pointed at an `agent_workflows` row:

- it reads `wf.transitions` (`views/WorkflowDetail.vue:113`), which for an agent
  workflow is the agent-kind set that no other reader uses;
- it reads `wf.issue_type_ids` (`:115`), a column `agent_workflows` does not
  have, so the issue-type selection is always empty;
- its "add transition" call sends **state ids as node ids**
  (`views/WorkflowDetail.vue:222-233`):

  ```js
  await workflowApi.addEdge(projectId, workflowId, {
    source_node_id: selectedFromState.value,
    target_node_id: newTransitionTo.value,
    ...
  })
  ```

  `AddEdge` validates that both node ids belong to the workflow, so the request
  fails with `source or target node not found` (500). Deleting a "transition"
  calls the node-edge delete endpoint with a transition id.

The page cannot work under either model. It is the UI-level version of the same
conflation this design removes.

### 2.6 Dead code that makes the confusion look intentional

| Item | Why it is dead |
|------|----------------|
| `frontend/src/components/StateTransition.vue` | No importer; only user of `listStateTransitions` and the `StateTransition` types |
| `WorkflowManager.vue` project branch | Only `WorkspaceSettings.vue` mounts the component, always with `workspace-id` |
| `workflowApi.addEdge` in that branch | Sends `from_state_id` / `to_state_id` to the node-edge endpoint |
| `GET/POST/PUT/DELETE /projects/:id/workflows/:id/transitions` | No UI, no service reader; only tests create agent-kind rows |

## 3. Decision

**Keep the two domains, give each table exactly one owner, and remove the
discriminator instead of threading it through more code.**

They are not variations of one thing:

- A transition is a **constraint**: evaluated synchronously on every issue
  update and every AI tool call, and it may deny the request.
- An agent workflow is an **action**: scheduled, long-running, costs tokens, and
  has retries and run state.

Coupling them would make the hot constraint path depend on the execution
subsystem. The relationship is therefore:

```
issue state change ──(constraint)──> state_transitions
        │
        └──(future event, not built)──> agent_workflows where trigger_type = 'event'
```

`agent_workflows.trigger_type` already accepts `event` and `webhook`, and
`trigger_config` is free-form JSON. When a transition should launch an
orchestration, publish `issue.transitioned` and let an event-triggered workflow
subscribe. No foreign key, no shared table, no ordering guarantee needed in the
write path.

## 4. Target model

```
workflows (state machine)
  ├── state_transitions.workflow_id  ──FK──> workflows.id
  ├── states (existing)
  └── approvals / approval_records (existing)

agent_workflows (orchestration)
  ├── workflow_nodes
  ├── workflow_edges
  ├── workflow_runs
  └── workflow_node_runs
```

- `state_transitions.workflow_kind` is **removed**. There is no second parent.
- `state_transitions.workflow_id` regains a real foreign key to
  `workflows(id)`. The FK was removed in October 2026 precisely because
  agent transitions could not satisfy it; once agent transitions no longer
  exist, the FK is correct and prevents orphan rows.
- Soft deletes stay. `ON DELETE CASCADE` only fires on a hard delete, so
  `DeleteStateWorkflow` keeps deleting its transitions explicitly.

## 5. Backend changes

### 5.1 Model

`backend/internal/model/state.go`

- Drop `WorkflowKind` from `StateTransition` and the `TransitionKind*` constants.
- Restore the relationship field so GORM maintains the FK:

  ```go
  Workflow Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
  ```

  The October comment that removal was needed no longer applies.

`backend/internal/model/workflow.go` — unchanged (`Workflow`,
`AgentWorkflow`, nodes, edges, runs).

### 5.2 Service split

`backend/internal/service/workflow_service.go` is 1,257 lines and owns both
domains. Split along the table boundary:

| New file | Owns | Moved methods |
|----------|------|---------------|
| `state_workflow_service.go` | `workflows`, `state_transitions` | `ListStateWorkflowsByWorkspace`, `CreateStateWorkflow`, `UpdateStateWorkflow`, `DeleteStateWorkflow`, `ListTransitions`, `AddTransition`, `UpdateTransition`, `DeleteTransition`, `resolveWorkflowScope` (state only), `toStateWorkflowResponse`, `toTransitionResponse`, `ResolveWorkspaceID` |
| `agent_workflow_service.go` | `agent_workflows`, nodes, edges, runs | `ListByProject`, `Get`, `GetByNodeID`, `Create`, `Update`, `Delete`, `AddNode`, `UpdateNode`, `DeleteNode`, `AddEdge`, `UpdateEdge`, `DeleteEdge`, `GetRuns`, `GetRun`, `ExecuteWorkflow` |

Signature changes:

- `ListTransitions(kind, workflowID)` → `ListTransitions(workflowID)`.
- `AddTransition(kind, workflowID, req)` → `AddTransition(workflowID, req)`.
- `UpdateTransition(kind, id, req)` → `UpdateTransition(id, req)`.
- `DeleteTransition(kind, id)` → `DeleteTransition(id)`.
- `resolveWorkflowScope` always reads `workflows`; delete the agent branch.
- `toStateWorkflowResponse` filters by `workflow_id` alone, which is now
  unambiguous. Add `deleted_at IS NULL` as it already does.
- `StateTransitionResponse.WorkflowKind` removed.
- `WorkflowDetail.Transitions` removed (an agent workflow has no transitions).

The agent service keeps `contextSvc`, `decisionSvc`, `budgetSvc`; the state
service only needs `db`.

### 5.3 Handlers and routes

`backend/internal/handler/workflow_handler.go` → split:

- `state_workflow_handler.go` — workspace list/create/update/delete and the
  transition CRUD. Passes no kind.
- `agent_workflow_handler.go` — project list/create/update/delete, nodes,
  edges, runs, execute, cancel.

Routes (`backend/internal/router/router.go`):

- Remove `/projects/:projectId/workflows/:workflowId/transitions` (GET, POST,
  PUT, DELETE). Those were agent-kind CRUD with no consumer.
- Keep `/projects/:projectId/workflows/:workflowId/edges` — it is the node DAG
  edge, driven by `WorkflowDesigner`.
- Keep the workspace `.../transitions` and the `.../edges` alias, now
  backed by the state service.
- `GET /workspaces/:ws/workflows/:workflowId` must return the state workflow
  (`workflowH.GetStateWorkflow`), not call the agent `Get`.

### 5.4 Middleware

`backend/internal/middleware/access.go`

- `RequireProjectWorkflow`: drop the `/transitions` branch and the
  `transitionChildOf(..., TransitionKindAgent, ...)` call. It validates
  `agent_workflows` + nodes/edges/runs only.
- `RequireWorkspaceWorkflow`: validate `:workflowId` against `workflows` where
  `workspace_id = wsID`; validate `:transitionId` against `state_transitions`
  where `workflow_id = :workflowId`.
- `transitionChildOf` loses its `kind` parameter.
- `AllowProject` / role rules unchanged.

### 5.5 Other call sites

- `backend/cmd/server/main.go:211-221` — remove the `workflow_kind` backfill.
  It would fail once the column is gone.
- `backend/cmd/route_audit/main.go:353` — `"workflows": {"agent_workflows",
  "workflows"}` is wrong for the project route group; keep both entries only if
  a route still addresses both. After this change no route does.
- `backend/cmd/route_audit/main.go:492` — the `workflow_id` union resolver
  becomes `SELECT id FROM workflows WHERE workspace_id = ?`.
- `backend/internal/seed/seed.go` — unchanged; it seeds `workflows` and
  transitions with no kind once the column defaults are removed.

## 6. Migration

`backend/migrations/000002_separate_state_transitions.up.sql`

```sql
-- 1. Preserve agent-kind rows before dropping the discriminator.
CREATE TABLE IF NOT EXISTS state_transitions_agent_archive (
    LIKE state_transitions INCLUDING DEFAULTS
);
ALTER TABLE state_transitions_agent_archive ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ DEFAULT now();

INSERT INTO state_transitions_agent_archive
SELECT * FROM state_transitions WHERE workflow_kind = 'agent';

-- 2. Drop them from the live table (including already soft-deleted rows).
DELETE FROM state_transitions WHERE workflow_kind = 'agent';

-- 3. One parent only.
ALTER TABLE state_transitions DROP COLUMN IF EXISTS workflow_kind;

-- 4. Restore the foreign key. Guarded because the constraint may exist on
--    databases created by an older baseline.
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

Down migration restores the column with default `'state'`, re-inserts archived
rows (preserving their ids), and drops the archive table.

Notes:

- `LIKE ... INCLUDING DEFAULTS` copies column types and defaults but not
  constraints, which is what we want for an archive.
- `INSERT ... SELECT *` assumes column order matches; generate the statement
  from `information_schema.columns` in the migration review, or list columns
  explicitly. Prefer listing columns explicitly to survive baseline drift.
- The migration is safe on a database with zero agent rows: it archives
  nothing and still drops the column.
- Existing databases created from `000001_baseline` have `workflow_kind`
  already; databases created by an older chain also have it (the column was
  added in archived migration `000022`).

## 7. Frontend

| Change | File |
|--------|------|
| Delete dead component and its types/imports | `src/components/StateTransition.vue`, `src/types/workflow.ts` (`StateTransition*`) |
| Delete dead API helpers and their tests | `src/api/workflow.ts` (`listStateTransitions`, `createStateTransition`, `updateStateTransition`, `deleteStateTransition`), `src/api/workflow.api.test.ts` |
| Delete the broken agent-detail editor and its route | `src/views/WorkflowDetail.vue`, `src/router/index.ts:223-228`, `handleViewWorkflow` in `src/views/ProjectSettings.vue` |
| Collapse `WorkflowManager.vue` to workspace-only, rename | `src/components/StateWorkflowManager.vue`, `src/views/WorkspaceSettings.vue` |
| Remove the project-mode branches (`workflowApi.list/addEdge/deleteEdge`) | same file |
| Agent pages unchanged | `src/views/agent/*` |

Removing `WorkflowDetail.vue` rather than repairing it: its only unique function
was editing transitions, which belongs to workspace state workflows; agent
workflow name/description/active already edit in `WorkflowsPage.vue`, and nodes
and edges in `WorkflowDesigner.vue`. Pointing `handleViewWorkflow` at the
designer route keeps the ProjectSettings row functional.

`ProjectSettings.vue` already renders a banner pointing state-machine
configuration at workspace settings; no change needed beyond confirming the
link still works.

## 8. Testing

Unit / service:

1. Workspace transition created through the workspace route lands in
   `state_transitions` with `workflow_id` belonging to that workspace's
   workflow, and no other project's rows change.
2. `ListStateWorkflowsByWorkspace` never returns a transition whose
   `workflow_id` is an agent workflow id (fixture with colliding ids).
3. `ListTransitions` for a workspace workflow returns its own rows only.
4. `RequireWorkspaceWorkflow` allows a state workflow id, rejects an agent
   workflow id, and rejects a transition id belonging to another workflow.
5. `RequireProjectWorkflow` no longer accepts `:transitionId`, and a request to
   a removed route is 404.
6. Rules engine (`internal/workflow`) unchanged: existing tests must pass
   without modification.
7. `GET /projects/:id/workflows/:id` no longer returns a `transitions` field;
   `WorkflowDesigner` still loads nodes, edges and runs.

Migration:

7. `cmd/migrate up` twice on a database with agent-kind rows → column gone,
   rows present in the archive, second run is a no-op.
8. `AutoMigrate` on a database already migrated → no attempt to re-add
   `workflow_kind`, no FK churn.
9. `cmd/migrate down` once → column and rows restored.

End-to-end:

10. Workspace settings: create workflow → add allow transition → open an issue
    and confirm the move is offered; add an approval transition and confirm the
    approval path still starts.

CI gates that already exist and will exercise this: `Tenancy audit` (route
group changes) and `Migration path` (new migration).

## 9. Rollout order

1. **Fix the live breakage first** (defects 1–3) as a small patch, without
   waiting for the split: point the workspace middleware and handlers at
   `workflows` / `state` and add the kind filter in `toStateWorkflowResponse`.
   This is user-visible and currently 404s.
2. Schema migration `000002`, with the archive table.
3. Service and handler split, `kind` removal, model FK.
4. Route cleanup and middleware simplification.
5. Frontend dead-code deletion, `WorkflowDetail.vue` removal, and
   `StateWorkflowManager` rename.
6. Docs: `docs/security.md` §11 rewritten to describe ownership instead of
   `workflow_kind`; `docs/bug-list.md` BUG-58 gains the residual fixes.

Step 1 can ship alone; steps 2–4 should land together because the schema change
and the signature change are the same commit boundary.

## 10. Non-goals

- Automatically running an agent workflow from a state transition. The event
  contract is described in §3 but not implemented.
- Unifying the two master tables under a `kind` column. Rejected in design
  review: different scope rules (`workflows.project_id` nullable for
  workspace-level defaults; `agent_workflows.project_id` required), different
  lifecycles, and every query would still branch.
- Renaming `agent_workflows` or the `/workflows` routes. The UI labels already
  distinguish "state workflow" from "agent orchestration".
- Multi-level approval, approval timeouts and `approval_mode='all'` — still out
  of scope in the 2026-07-19 approval spec.

## 11. Risks

| Risk | Mitigation |
|------|------------|
| Production has agent-kind transitions with real configuration | Archive table preserves every row and id; down migration restores them. Verify the count before and after in the deploy log. |
| `AutoMigrate` re-adds `workflow_kind` | The field is removed from the model in the same commit; CI's migration job boots the server against the migrated schema and would fail loudly. |
| Route removal breaks an unmonitored client | The removed project transition routes have no frontend caller and no server-side reader; the audit harness covers every registered route. |
| `StateTransition.vue` or `WorkflowDetail.vue` is used by a downstream fork | Neither is reachable from a working flow; deletion is recorded in the commit message so both can be restored from history. |
