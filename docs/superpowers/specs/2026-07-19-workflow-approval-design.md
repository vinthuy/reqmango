# Workflow Approval Design (Plane AI Style)

> Date: 2026-07-19
> Status: Draft
> Owner: reqmango team

## 1. Overview

Reference: Plane AI's asynchronous approval flow pattern.

This feature replaces the current "permission-check only" approval logic with a complete asynchronous approval workflow:

1. User A submits a state change request that requires approval
2. The issue enters a virtual "pending approval" state
3. Any approver (OR logic) can approve or reject the request
4. Approval -> issue jumps to `approve_target_state_id`
5. Rejection -> issue jumps to `reject_target_state_id`
6. Full approval history is recorded and surfaced in the issue activity stream
7. Notifications are sent to approvers (on submit) and to the requester (on decision)

## 2. Current State

### 2.1 What exists
- `state_transitions` table has `rule_type`, `approver_ids`, `role_allowed` fields
- `rule_type` accepts `'allow'` or `'approval'`
- `issue_service.go` (lines 1650-1697) checks `rule_type='approval'`:
  - If user is in `approver_ids` -> allow direct transition
  - If user role >= `role_allowed` -> allow direct transition
  - Otherwise -> return 400 error
- This is **permission-check only**, not a real approval flow

### 2.2 Known bug to fix
- Frontend `WorkflowManager.vue` writes `approver_ids` via `JSON.stringify(arr)` (e.g. `"[1,2,3]"`)
- Backend `issue_service.go` parses via `strings.Split(s, ",")` (expecting `"1,2,3"`)
- Format mismatch. Unify on JSON array format during this redesign.

### 2.3 What is missing
- No `approvals` table
- No `approval_records` table
- No approval request/decision API
- No approval center UI
- No approval history in issue activity stream
- No "approve_target_state_id" / "reject_target_state_id" fields on transitions

## 3. Data Model

### 3.1 Extend `state_transitions` table

Add columns:

```sql
ALTER TABLE state_transitions
  ADD COLUMN approve_target_state_id BIGINT,
  ADD COLUMN reject_target_state_id BIGINT,
  ADD COLUMN approval_mode VARCHAR(20) DEFAULT 'any';
```

- `approve_target_state_id`: target state when approved. If NULL, fall back to `target_state_id` (backward compatibility).
- `reject_target_state_id`: target state when rejected. If NULL, fall back to `source_state_id` (rollback).
- `approval_mode`: `'any'` (default, any approver can decide) or `'all'` (all approvers must approve). Reserved for future; v1 implements `'any'` only.
- `approver_ids` format unified to JSON array string: `"[1,2,3]"`.

### 3.2 New `approvals` table

```sql
CREATE TABLE approvals (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by_id BIGINT,
    updated_by_id BIGINT,

    issue_id BIGINT NOT NULL,
    workflow_id BIGINT NOT NULL,
    transition_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    workspace_id BIGINT NOT NULL,

    requester_id BIGINT NOT NULL,
    request_note TEXT,

    source_state_id BIGINT NOT NULL,
    approve_target_state_id BIGINT NOT NULL,
    reject_target_state_id BIGINT NOT NULL,

    approver_ids JSONB NOT NULL DEFAULT '[]'::jsonb,

    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending | approved | rejected | cancelled

    decided_by BIGINT,
    decided_at TIMESTAMPTZ,
    decision_note TEXT
);

CREATE INDEX idx_approvals_issue_id ON approvals(issue_id);
CREATE INDEX idx_approvals_status ON approvals(status);
CREATE INDEX idx_approvals_project_id ON approvals(project_id);
CREATE INDEX idx_approvals_workspace_id ON approvals(workspace_id);
```

### 3.3 New `approval_records` table

Records each approver's individual decision. Supports future "all approvers" mode and audit trail.

```sql
CREATE TABLE approval_records (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    approval_id BIGINT NOT NULL,
    approver_id BIGINT NOT NULL,
    decision VARCHAR(20) NOT NULL,  -- approved | rejected
    note TEXT,
    decided_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_approval_records_approval_id ON approval_records(approval_id);
CREATE INDEX idx_approval_records_approver_id ON approval_records(approver_id);
```

### 3.4 Extend `issues` table

Add columns to track active approval state without changing `state_id`:

```sql
ALTER TABLE issues
  ADD COLUMN approval_status VARCHAR(20),
  ADD COLUMN active_approval_id BIGINT;
```

- `approval_status`: NULL (no active approval) | `'pending'` | `'approved'` | `'rejected'` | `'cancelled'`
- `active_approval_id`: FK to `approvals.id` when status is `pending`; NULL otherwise.

The issue's `state_id` stays at `source_state_id` during pending approval. UI shows a "pending approval" banner. This avoids needing a dedicated "pending" state in the state table.

## 4. Backend

### 4.1 Models (`backend/internal/model/approval.go` - new file)

```go
type Approval struct {
    BaseModel
    IssueID               uint64  `gorm:"index" json:"issue_id"`
    WorkflowID            uint64  `json:"workflow_id"`
    TransitionID          uint64  `gorm:"index" json:"transition_id"`
    ProjectID             uint64  `gorm:"index" json:"project_id"`
    WorkspaceID           uint64  `gorm:"index" json:"workspace_id"`
    RequesterID           uint64  `gorm:"not null" json:"requester_id"`
    RequestNote            string  `gorm:"type:text" json:"request_note"`
    SourceStateID         uint64  `gorm:"not null" json:"source_state_id"`
    ApproveTargetStateID  uint64  `gorm:"not null" json:"approve_target_state_id"`
    RejectTargetStateID   uint64  `gorm:"not null" json:"reject_target_state_id"`
    ApproverIDs           string  `gorm:"type:jsonb;not null;default:'[]'" json:"approver_ids"`
    Status                string  `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
    DecidedBy             *uint64 `json:"decided_by"`
    DecidedAt             *time.Time `json:"decided_at"`
    DecisionNote          string  `gorm:"type:text" json:"decision_note"`
}

type ApprovalRecord struct {
    BaseModel
    ApprovalID uint64    `gorm:"index;not null" json:"approval_id"`
    ApproverID uint64    `gorm:"index;not null" json:"approver_id"`
    Decision   string    `gorm:"type:varchar(20);not null" json:"decision"`
    Note       string    `gorm:"type:text" json:"note"`
    DecidedAt  time.Time `gorm:"not null" json:"decided_at"`
}
```

Extend `StateTransition` model with `ApproveTargetStateID *uint64`, `RejectTargetStateID *uint64`, `ApprovalMode string`.

Extend `Issue` model with `ApprovalStatus *string`, `ActiveApprovalID *uint64`.

### 4.2 Service layer (`backend/internal/service/approval_service.go` - new file)

Methods:
- `Create(issueID, requesterID, transitionID, requestNote) (*Approval, error)`
  - Validates: issue has no other `pending` approval
  - Validates: transition.rule_type == 'approval'
  - Resolves approve/reject target states (fall back to target_state_id / source_state_id)
  - Creates `Approval` row with status='pending'
  - Updates `issues.approval_status='pending'`, `issues.active_approval_id`
  - Sends notifications to all approvers
  - Adds issue activity "submitted approval"

- `List(filter) ([]Approval, error)` - filters: workspace_id, project_id, status, approver_id (current user)
- `Get(id) (*ApprovalDetail, error)` - includes approval_records

- `Decide(approvalID, approverID, decision, note) (*Approval, error)`
  - Validates: approver is in approver_ids
  - Validates: approval.status == 'pending'
  - Creates `ApprovalRecord`
  - Updates `Approval`: status, decided_by, decided_at, decision_note (OR logic: first decision wins)
  - Updates `Issue`: state_id (approve/reject target), approval_status, active_approval_id=NULL
  - Sends notification to requester
  - Adds issue activity "approved"/"rejected"

- `Cancel(approvalID, userID) (*Approval, error)`
  - Validates: userID == approval.requester_id
  - Validates: approval.status == 'pending'
  - Updates `Approval`: status='cancelled'
  - Updates `Issue`: approval_status=NULL, active_approval_id=NULL
  - Adds issue activity "cancelled approval"

### 4.3 Refactor `issue_service.go`

Modify `checkTransitionAllowed` (around line 1650):
- Remove the existing "permission-check" logic for `rule_type='approval'`.
- New behavior: when the requested transition has `rule_type='approval'`, return a specific error code `approval_required` (HTTP 409) with the `transition_id` in the response body. This signals the frontend to open the approval submission dialog.
- Additionally, before any state change, check `issue.approval_status`. If it equals `'pending'`, reject the state change with HTTP 409 `issue_pending_approval`. This backend guard prevents direct state changes during an active approval (defends against API misuse, not just frontend pre-checks).

Frontend pre-checks (`rule_type` lookup before submitting) are an optimization to avoid failed requests, but the backend guard is the source of truth.

### 4.4 Handler & routes (`backend/internal/handler/approval_handler.go` - new file)

```
POST   /api/v1/issues/:issueId/approvals          # submit approval request
GET    /api/v1/workspaces/:workspaceId/approvals  # list (current user as approver)
GET    /api/v1/projects/:projectId/approvals      # list by project
GET    /api/v1/approvals/:id                       # detail
POST   /api/v1/approvals/:id/decide                # approve/reject
POST   /api/v1/approvals/:id/cancel                # cancel
GET    /api/v1/workspaces/:workspaceId/approvals/count  # TopBar badge count
```

### 4.5 Workflow transition API extension

Update `TransitionCreate` / `TransitionUpdate` DTO and `AddTransition` / `UpdateTransition` service to accept:
- `approve_target_state_id`
- `reject_target_state_id`
- `approval_mode`
- `approver_ids` (JSON array string format)

## 5. Frontend

### 5.1 Workflow detail page (`WorkflowDetail.vue`)

Add transition dialog: when `rule_type='approval'`:
- Show "Approve target state" dropdown (default = target_state_id)
- Show "Reject target state" dropdown (default = source_state_id)
- Show "Approval mode" select (only `'any'` enabled in v1, `'all'` shown as disabled with "coming soon")
- Show approver multi-select checkbox list (existing)

Transition list rendering:
- For approval transitions, show: `Source -> [Approve -> approveTarget] / [Reject -> rejectTarget]`

### 5.2 Issue detail page (`IssueDetail.vue`)

#### 5.2.1 State change dialog
- User selects target state from dropdown
- Frontend looks up the matching transition
- If `rule_type='allow'`: call existing state change API
- If `rule_type='approval'`: open "Submit approval" dialog
  - Show: from state -> to state (approve target)
  - Show: approvers list (read-only)
  - Input: request note (textarea)
  - Buttons: Cancel / Submit

#### 5.2.2 Pending approval banner
When `issue.approval_status === 'pending'`:
- Top of issue detail page shows yellow banner:
  - "Pending approval: requested by {requester} at {time}"
  - "Approvers: {names}"
  - If current user is in approvers: show [Approve] [Reject] buttons
  - If current user is requester: show [Cancel] button
  - Click Approve/Reject opens dialog with note input

#### 5.2.3 Activity stream entries
New activity types:
- `approval_submitted`: "{user} submitted approval to move issue to {approve_target_state}"
- `approval_approved`: "{user} approved the approval, issue moved to {state}"
- `approval_rejected`: "{user} rejected the approval, issue moved to {state}"
- `approval_cancelled`: "{user} cancelled the approval"

### 5.3 TopBar approval icon (`TopBar.vue`)

- Add bell-shaped "approval" icon next to notification bell
- Fetch count from `/api/v1/workspaces/:wid/approvals/count?status=pending`
- Click -> dropdown showing latest 5 pending approvals (issue key, title, requester, time)
- Each item click -> navigate to issue detail page
- "View all" link -> navigate to approval list page

### 5.4 Approval list page (new: `views/ApprovalList.vue`)

Route: `/workspace/:slug/approvals`

- Filter bar: status (pending/approved/rejected/cancelled), project
- Table columns: issue key+title, project, requester, submitted at, status, action
- Action buttons (for pending items where current user is approver): Approve / Reject
- Row click -> navigate to issue detail page

### 5.5 i18n

Add `approvals.*` keys for all user-facing strings (zh-CN and en-US).

## 6. Business Flows

### 6.1 Submit approval
1. User opens issue detail, selects target state from dropdown
2. Frontend detects `rule_type='approval'` for matching transition
3. Frontend opens "Submit approval" dialog
4. User enters request note, clicks Submit
5. `POST /api/v1/issues/:issueId/approvals` with `{ transition_id, request_note }`
6. Backend creates Approval (status=pending), updates issue.approval_status=pending
7. Backend sends notifications to all approvers
8. Backend adds activity "approval_submitted"
9. Frontend refreshes issue detail (banner appears)

### 6.2 Approve / Reject
1. Approver sees TopBar badge or visits approval list page
2. Clicks Approve / Reject
3. Dialog opens, optional note input
4. `POST /api/v1/approvals/:id/decide` with `{ decision, note }`
5. Backend validates approver, creates approval_record
6. Backend updates Approval status, decided_by, decided_at
7. Backend updates Issue: state_id = approve/reject target, approval_status = approved/rejected, active_approval_id = NULL
8. Backend sends notification to requester
9. Backend adds activity "approval_approved" / "approval_rejected"
10. Frontend refreshes

### 6.3 Cancel (by requester)
1. Requester clicks Cancel on the pending approval banner
2. Confirm dialog
3. `POST /api/v1/approvals/:id/cancel`
4. Backend updates Approval status=cancelled, issue.approval_status=NULL
5. Backend adds activity "approval_cancelled"
6. Frontend refreshes

## 7. Edge Cases & Error Handling

- **Duplicate submission**: If issue already has a pending approval, reject new submission with 400 `issue_already_pending_approval`.
- **Approver leaves workspace**: Approver remains in `approver_ids`; if no remaining valid approvers, allow requester to cancel.
- **Incomplete transition config**: If `approve_target_state_id` and `target_state_id` are both NULL, return 400 `transition_config_incomplete`.
- **State change during pending approval**: Block direct state change when `issue.approval_status='pending'`. Return 400 `issue_pending_approval`.
- **Permission**: Only the requester can cancel; only listed approvers can decide. Return 403 otherwise.
- **Concurrent decision**: Use DB transaction with `SELECT ... FOR UPDATE` on the approval row. First decision wins; second decision returns 409 `approval_already_decided`.

## 8. Backward Compatibility

- Existing `state_transitions` rows with `rule_type='approval'`:
  - `target_state_id` is treated as `approve_target_state_id` (fallback)
  - `reject_target_state_id` falls back to `source_state_id` (rollback)
- Existing `approver_ids` in comma-separated format: run a one-time migration script to convert to JSON array format.
- Existing `issue_service.go` permission-check logic for `rule_type='approval'` is removed; the new flow always creates an approval request.

## 9. Testing

### 9.1 Unit tests
- `approval_service_test.go`: Create / Decide (approve/reject) / Cancel / Duplicate submission / Permission denied / Concurrent decision
- `issue_service_test.go`: Verify `rule_type='approval'` no longer blocks state change directly; instead returns `approval_required` indicator.

### 9.2 API tests
- All approval endpoints with happy path and error cases.
- Transition API with new fields.

### 9.3 Integration test
End-to-end flow: requester submits -> approver decides -> issue state updated -> notification sent -> activity recorded.

### 9.4 Frontend manual test
- Workflow config: create approval transition with approve/reject targets
- Issue detail: submit approval, see banner, cancel
- TopBar: badge count, dropdown
- Approval list page: filter, approve/reject
- Activity stream: all four activity types render correctly

## 10. Out of Scope (v1)

- Multi-level approval (chain of approvals)
- SLA / timeout for approvals
- Email notifications (only in-app notifications)
- Approval delegation (approver reassigns to another user)
- `approval_mode='all'` (all approvers must approve) - reserved in schema, not implemented in v1
- Bulk approval from list page

## 11. Rollout Plan

1. Phase 1 (backend): schema migration + models + service + handlers + unit tests
2. Phase 2 (backend integration): refactor issue_service.go, notifications, activity stream
3. Phase 3 (frontend): workflow detail page UI changes
4. Phase 4 (frontend): issue detail page state change dialog + pending banner + activity rendering
5. Phase 5 (frontend): TopBar icon + approval list page
6. Phase 6: integration testing, manual acceptance, bug fixes
7. Phase 7: data migration script for existing `approver_ids` format

Each phase should be merged and tested independently.
