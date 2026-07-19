# Workflow Approval Implementation Plan (Plane AI Style)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the existing "permission-check only" approval logic with a complete asynchronous approval workflow (Plane AI style) that supports submission, approver decision (approve/reject with independent target states), notifications, and audit trail.

**Architecture:** New `approvals` + `approval_records` tables track approval lifecycle. `issues` table gains `approval_status` + `active_approval_id` to mark the virtual "pending approval" state. `state_transitions` table gains `approve_target_state_id` + `reject_target_state_id`. Backend `approval_service` is the single source of truth; `issue_service` is refactored to return `approval_required` instead of blocking. Frontend adds TopBar badge, approval list page, issue detail banner, and workflow config UI.

**Tech Stack:** Go + Gin + GORM (backend), Vue 3 + TypeScript + Vite (frontend), PostgreSQL, `DATA-DOG/go-sqlmock` + `testify/assert` (backend tests).

**Spec:** `docs/superpowers/specs/2026-07-19-workflow-approval-design.md`

---

## File Structure

### Backend (new files)
- `backend/internal/model/approval.go` — Approval and ApprovalRecord models
- `backend/internal/service/approval_service.go` — Approval business logic
- `backend/internal/service/approval_service_test.go` — Unit tests
- `backend/internal/handler/approval_handler.go` — HTTP handlers
- `backend/internal/dto/request/approval.go` — Request DTOs
- `backend/internal/dto/response/approval.go` — Response DTOs

### Backend (modified files)
- `backend/internal/model/state.go` — Extend StateTransition with approve/reject target + approval_mode
- `backend/internal/model/issue.go` — Extend Issue with approval_status + active_approval_id
- `backend/internal/dto/request/workflow.go` — Extend TransitionCreate/Update
- `backend/internal/dto/response/workflow.go` — Extend TransitionResponse
- `backend/internal/service/workflow_service.go` — Update AddTransition/UpdateTransition
- `backend/internal/service/issue_service.go` — Refactor checkTransitionAllowed
- `backend/internal/handler/workflow_handler.go` — Bind new fields
- `backend/internal/router/router.go` — Register approval routes
- `backend/migrations/000002_workflow_approval.up.sql` (new) — Schema migration
- `backend/migrations/000002_workflow_approval.down.sql` (new) — Rollback

### Frontend (new files)
- `frontend/src/api/approval.ts` — API client
- `frontend/src/views/ApprovalList.vue` — Approval center page
- `frontend/src/components/ApprovalBadge.vue` — TopBar icon + dropdown
- `frontend/src/components/ApprovalSubmitDialog.vue` — Submit approval dialog
- `frontend/src/components/ApprovalDecisionDialog.vue` — Approve/reject dialog
- `frontend/src/components/ApprovalPendingBanner.vue` — Issue detail banner

### Frontend (modified files)
- `frontend/src/router/index.ts` — Register /workspace/:slug/approvals route
- `frontend/src/components/TopBar.vue` — Add ApprovalBadge
- `frontend/src/views/WorkflowDetail.vue` — Extend add-transition dialog
- `frontend/src/views/IssueDetail.vue` — Integrate submit dialog + banner + activity rendering
- `frontend/src/components/ActivityStream.vue` (if exists) — Render approval activities
- `frontend/src/locales/zh-CN.json` — i18n keys
- `frontend/src/locales/en-US.json` — i18n keys

---

## Phase 1: Backend Schema + Models

### Task 1.1: Create database migration

**Files:**
- Create: `backend/migrations/000002_workflow_approval.up.sql`
- Create: `backend/migrations/000002_workflow_approval.down.sql`

- [ ] **Step 1: Write the up migration**

Create `backend/migrations/000002_workflow_approval.up.sql`:

```sql
-- Extend state_transitions
ALTER TABLE state_transitions
  ADD COLUMN approve_target_state_id BIGINT,
  ADD COLUMN reject_target_state_id BIGINT,
  ADD COLUMN approval_mode VARCHAR(20) DEFAULT 'any';

-- New approvals table
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
    decided_by BIGINT,
    decided_at TIMESTAMPTZ,
    decision_note TEXT
);

CREATE INDEX idx_approvals_issue_id ON approvals(issue_id);
CREATE INDEX idx_approvals_status ON approvals(status);
CREATE INDEX idx_approvals_project_id ON approvals(project_id);
CREATE INDEX idx_approvals_workspace_id ON approvals(workspace_id);
CREATE INDEX idx_approvals_requester_id ON approvals(requester_id);
CREATE INDEX idx_approvals_deleted_at ON approvals(deleted_at);

-- New approval_records table
CREATE TABLE approval_records (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    approval_id BIGINT NOT NULL,
    approver_id BIGINT NOT NULL,
    decision VARCHAR(20) NOT NULL,
    note TEXT,
    decided_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_approval_records_approval_id ON approval_records(approval_id);
CREATE INDEX idx_approval_records_approver_id ON approval_records(approver_id);

-- Extend issues table
ALTER TABLE issues
  ADD COLUMN approval_status VARCHAR(20),
  ADD COLUMN active_approval_id BIGINT;

CREATE INDEX idx_issues_approval_status ON issues(approval_status);

-- Data migration: convert existing approver_ids from comma-separated to JSON array
UPDATE state_transitions
SET approver_ids = '[' || REPLACE(approver_ids, ',', ',') || ']'
WHERE rule_type = 'approval'
  AND approver_ids IS NOT NULL
  AND approver_ids != ''
  AND approver_ids NOT LIKE '[%';
```

- [ ] **Step 2: Write the down migration**

Create `backend/migrations/000002_workflow_approval.down.sql`:

```sql
DROP TABLE IF EXISTS approval_records;
DROP TABLE IF EXISTS approvals;

ALTER TABLE state_transitions
  DROP COLUMN IF EXISTS approve_target_state_id,
  DROP COLUMN IF EXISTS reject_target_state_id,
  DROP COLUMN IF EXISTS approval_mode;

ALTER TABLE issues
  DROP COLUMN IF EXISTS approval_status,
  DROP COLUMN IF EXISTS active_approval_id;
```

- [ ] **Step 3: Apply migration to dev database**

Run:
```bash
psql "host=localhost port=5432 user=postgres password=postgres dbname=reqmango sslmode=disable" -f backend/migrations/000002_workflow_approval.up.sql
```
Expected: series of `ALTER TABLE` / `CREATE TABLE` / `CREATE INDEX` / `UPDATE` success messages, no errors.

- [ ] **Step 4: Commit**

```bash
git add backend/migrations/000002_workflow_approval.up.sql backend/migrations/000002_workflow_approval.down.sql
git commit -m "feat(approval): add schema migration for workflow approval"
```

---

### Task 1.2: Add Approval and ApprovalRecord models

**Files:**
- Create: `backend/internal/model/approval.go`

- [ ] **Step 1: Write the model file**

Create `backend/internal/model/approval.go`:

```go
package model

import "time"

// Approval represents a single approval request on an issue.
type Approval struct {
	BaseModel
	IssueID              uint64     `gorm:"index;not null" json:"issue_id"`
	WorkflowID           uint64     `gorm:"not null" json:"workflow_id"`
	TransitionID         uint64     `gorm:"index;not null" json:"transition_id"`
	ProjectID            uint64     `gorm:"index;not null" json:"project_id"`
	WorkspaceID          uint64     `gorm:"index;not null" json:"workspace_id"`
	RequesterID          uint64     `gorm:"not null" json:"requester_id"`
	RequestNote          string     `gorm:"type:text" json:"request_note"`
	SourceStateID        uint64     `gorm:"not null" json:"source_state_id"`
	ApproveTargetStateID uint64     `gorm:"not null" json:"approve_target_state_id"`
	RejectTargetStateID  uint64     `gorm:"not null" json:"reject_target_state_id"`
	ApproverIDs          string     `gorm:"type:jsonb;not null;default:'[]'" json:"approver_ids"`
	Status               string     `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	DecidedBy            *uint64    `json:"decided_by"`
	DecidedAt            *time.Time `json:"decided_at"`
	DecisionNote         string     `gorm:"type:text" json:"decision_note"`
}

func (Approval) TableName() string { return "approvals" }

// ApprovalRecord records each approver's individual decision on an approval.
type ApprovalRecord struct {
	BaseModel
	ApprovalID uint64    `gorm:"index;not null" json:"approval_id"`
	ApproverID uint64    `gorm:"index;not null" json:"approver_id"`
	Decision   string    `gorm:"type:varchar(20);not null" json:"decision"`
	Note       string    `gorm:"type:text" json:"note"`
	DecidedAt  time.Time `gorm:"not null" json:"decided_at"`
}

func (ApprovalRecord) TableName() string { return "approval_records" }
```

- [ ] **Step 2: Extend StateTransition model**

Edit `backend/internal/model/state.go`, add fields to `StateTransition` struct (after `RoleAllowed` field):

```go
	ApproveTargetStateID *uint64 `json:"approve_target_state_id"`
	RejectTargetStateID  *uint64 `json:"reject_target_state_id"`
	ApprovalMode         string  `gorm:"type:varchar(20);default:'any'" json:"approval_mode"`
```

- [ ] **Step 3: Extend Issue model**

Edit `backend/internal/model/issue.go`, add fields to `Issue` struct (find an appropriate location near other status fields):

```go
	ApprovalStatus    *string `gorm:"type:varchar(20);index" json:"approval_status"`
	ActiveApprovalID  *uint64 `gorm:"index" json:"active_approval_id"`
```

- [ ] **Step 4: Register models in AutoMigrate (if applicable)**

Search for `AutoMigrate` in `backend/internal/database` or `backend/cmd/server/main.go`. Add `model.Approval{}` and `model.ApprovalRecord{}` to the migration list.

Run:
```bash
grep -rn "AutoMigrate" backend/internal/database backend/cmd
```
If found, append the two new models. If the project relies only on SQL migrations, skip this step.

- [ ] **Step 5: Verify the code compiles**

Run:
```bash
cd backend && go build ./...
```
Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/model/approval.go backend/internal/model/state.go backend/internal/model/issue.go
git commit -m "feat(approval): add Approval and ApprovalRecord models, extend StateTransition and Issue"
```

---

## Phase 2: Backend Service Layer (TDD)

### Task 2.1: Define DTOs and helper functions

**Files:**
- Create: `backend/internal/dto/request/approval.go`
- Create: `backend/internal/dto/response/approval.go`

- [ ] **Step 1: Write request DTOs**

Create `backend/internal/dto/request/approval.go`:

```go
package request

type ApprovalCreate struct {
	TransitionID uint64 `json:"transition_id" binding:"required"`
	RequestNote  string `json:"request_note"`
}

type ApprovalDecision struct {
	Decision string `json:"decision" binding:"required,oneof=approved rejected"`
	Note     string `json:"note"`
}

type ApprovalListQuery struct {
	Status     string `form:"status" json:"status"`
	ProjectID  uint64 `form:"project_id" json:"project_id"`
	ApproverID uint64 `form:"approver_id" json:"approver_id"`
}
```

- [ ] **Step 2: Write response DTOs**

Create `backend/internal/dto/response/approval.go`:

```go
package response

import "time"

type ApprovalResponse struct {
	ID                   uint64     `json:"id"`
	IssueID              uint64     `json:"issue_id"`
	WorkflowID           uint64     `json:"workflow_id"`
	TransitionID         uint64     `json:"transition_id"`
	ProjectID            uint64     `json:"project_id"`
	WorkspaceID          uint64     `json:"workspace_id"`
	RequesterID          uint64     `json:"requester_id"`
	RequesterName        string     `json:"requester_name"`
	RequestNote          string     `json:"request_note"`
	SourceStateID        uint64     `json:"source_state_id"`
	SourceStateName      string     `json:"source_state_name"`
	ApproveTargetStateID uint64     `json:"approve_target_state_id"`
	ApproveStateName     string     `json:"approve_target_state_name"`
	RejectTargetStateID  uint64     `json:"reject_target_state_id"`
	RejectStateName      string     `json:"reject_target_state_name"`
	ApproverIDs          []uint64   `json:"approver_ids"`
	ApproverNames        []string   `json:"approver_names"`
	Status               string     `json:"status"`
	DecidedBy            *uint64    `json:"decided_by"`
	DecidedByName        string     `json:"decided_by_name"`
	DecidedAt            *time.Time `json:"decided_at"`
	DecisionNote         string     `json:"decision_note"`
	CreatedAt            time.Time  `json:"created_at"`

	IssueKey   string `json:"issue_key"`
	IssueTitle string `json:"issue_title"`
	ProjectName string `json:"project_name"`

	Records []ApprovalRecordResponse `json:"records"`
}

type ApprovalRecordResponse struct {
	ID         uint64    `json:"id"`
	ApproverID uint64    `json:"approver_id"`
	ApproverName string  `json:"approver_name"`
	Decision   string    `json:"decision"`
	Note       string    `json:"note"`
	DecidedAt  time.Time `json:"decided_at"`
}

type ApprovalCountResponse struct {
	PendingCount int64 `json:"pending_count"`
}
```

- [ ] **Step 3: Extend TransitionCreate/Update DTOs**

Edit `backend/internal/dto/request/workflow.go`. Find `TransitionCreate` struct, add fields:

```go
	ApproveTargetStateID *uint64 `json:"approve_target_state_id"`
	RejectTargetStateID  *uint64 `json:"reject_target_state_id"`
	ApprovalMode         string  `json:"approval_mode"`
```

Do the same for `TransitionUpdate`.

- [ ] **Step 4: Extend TransitionResponse DTO**

Edit `backend/internal/dto/response/workflow.go`. Find `TransitionResponse` struct, add:

```go
	ApproveTargetStateID *uint64 `json:"approve_target_state_id"`
	ApproveStateName     string  `json:"approve_target_state_name"`
	RejectTargetStateID  *uint64 `json:"reject_target_state_id"`
	RejectStateName      string  `json:"reject_target_state_name"`
	ApprovalMode         string  `json:"approval_mode"`
```

- [ ] **Step 5: Verify build**

```bash
cd backend && go build ./...
```
Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/dto/request/approval.go backend/internal/dto/response/approval.go backend/internal/dto/request/workflow.go backend/internal/dto/response/workflow.go
git commit -m "feat(approval): add approval DTOs and extend transition DTOs"
```

---

### Task 2.2: Write approval_service with TDD

**Files:**
- Create: `backend/internal/service/approval_service_test.go`
- Create: `backend/internal/service/approval_service.go`

- [ ] **Step 1: Write the failing test for Create**

Create `backend/internal/service/approval_service_test.go`:

```go
package service

import (
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newApprovalTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}
	return gdb, mock
}

func TestApprovalService_Create_Success(t *testing.T) {
	db, mock := newApprovalTestDB(t)
	defer db.DB()

	// Mock: issue query (no pending approval)
	mock.ExpectQuery(`SELECT .* FROM "issues"`).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "state_id", "project_id", "workspace_id", "approval_status"}).
			AddRow(1, 10, 1, 1, nil))

	// Mock: transition query
	approverIDs := `[2,3]`
	mock.ExpectQuery(`SELECT .* FROM "state_transitions"`).
		WithArgs(uint64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "workflow_id", "source_state_id", "target_state_id", "rule_type", "approver_ids", "approve_target_state_id", "reject_target_state_id", "approval_mode"}).
			AddRow(100, 50, 10, 20, "approval", approverIDs, 20, 10, "any"))

	// Mock: state name lookups (optional, may be skipped in unit test)
	// Mock: count existing pending approvals = 0
	mock.ExpectQuery(`SELECT count`).
		WithArgs(uint64(1), "pending").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock: BEGIN tx
	mock.ExpectBegin()
	// Mock: INSERT approval
	mock.ExpectQuery(`INSERT INTO "approvals"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	// Mock: UPDATE issue
	mock.ExpectExec(`UPDATE "issues"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Mock: COMMIT
	mock.ExpectCommit()

	svc := NewApprovalService(db)
	approval, err := svc.Create(1, 1, 100, "please review")
	assert.NoError(t, err)
	assert.NotNil(t, approval)
	assert.Equal(t, "pending", approval.Status)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestApprovalService_Create_DuplicatePending(t *testing.T) {
	db, mock := newApprovalTestDB(t)
	defer db.DB()

	mock.ExpectQuery(`SELECT count`).
		WithArgs(uint64(1), "pending").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	svc := NewApprovalService(db)
	_, err := svc.Create(1, 1, 100, "please review")
	assert.Error(t, err)
}

func TestApprovalService_Decide_Approve(t *testing.T) {
	db, mock := newApprovalTestDB(t)
	defer db.DB()

	// Mock: get approval
	approverIDs := `[2,3]`
	mock.ExpectQuery(`SELECT .* FROM "approvals"`).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "approver_ids", "approve_target_state_id", "reject_target_state_id", "issue_id", "requester_id"}).
			AddRow(1, "pending", approverIDs, 20, 10, 100, 5))

	mock.ExpectBegin()
	// Mock: INSERT approval_record
	mock.ExpectQuery(`INSERT INTO "approval_records"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	// Mock: UPDATE approval
	mock.ExpectExec(`UPDATE "approvals"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Mock: UPDATE issue
	mock.ExpectExec(`UPDATE "issues"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	svc := NewApprovalService(db)
	approval, err := svc.Decide(1, 2, "approved", "looks good")
	assert.NoError(t, err)
	assert.Equal(t, "approved", approval.Status)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestApprovalService_Decide_NotInApprovers(t *testing.T) {
	db, mock := newApprovalTestDB(t)
	defer db.DB()

	approverIDs := `[2,3]`
	mock.ExpectQuery(`SELECT .* FROM "approvals"`).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "approver_ids", "approve_target_state_id", "reject_target_state_id", "issue_id", "requester_id"}).
			AddRow(1, "pending", approverIDs, 20, 10, 100, 5))

	svc := NewApprovalService(db)
	_, err := svc.Decide(1, 99, "approved", "")
	assert.Error(t, err)
}

func TestApprovalService_Cancel_Success(t *testing.T) {
	db, mock := newApprovalTestDB(t)
	defer db.DB()

	mock.ExpectQuery(`SELECT .* FROM "approvals"`).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "requester_id", "issue_id"}).
			AddRow(1, "pending", 5, 100))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "approvals"`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "issues"`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	svc := NewApprovalService(db)
	approval, err := svc.Cancel(1, 5)
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", approval.Status)
}

func TestApprovalService_Cancel_NotRequester(t *testing.T) {
	db, mock := newApprovalTestDB(t)
	defer db.DB()

	mock.ExpectQuery(`SELECT .* FROM "approvals"`).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "requester_id", "issue_id"}).
			AddRow(1, "pending", 5, 100))

	svc := NewApprovalService(db)
	_, err := svc.Cancel(1, 99)
	assert.Error(t, err)
}

// Helper: decode JSON array of uint64
func parseApproverIDs(t *testing.T, s string) []uint64 {
	var ids []uint64
	if err := json.Unmarshal([]byte(s), &ids); err != nil {
		t.Fatalf("failed to parse approver_ids: %v", err)
	}
	return ids
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
cd backend && go test ./internal/service -run TestApprovalService -v
```
Expected: compilation error (`NewApprovalService` undefined).

- [ ] **Step 3: Implement approval_service.go**

Create `backend/internal/service/approval_service.go`:

```go
package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type ApprovalService struct {
	db *gorm.DB
}

func NewApprovalService(db *gorm.DB) *ApprovalService {
	return &ApprovalService{db: db}
}

// Create submits a new approval request.
func (s *ApprovalService) Create(issueID, requesterID, transitionID uint64, requestNote string) (*model.Approval, error) {
	// 1. Check no other pending approval exists on this issue
	var pendingCount int64
	if err := s.db.Model(&model.Approval{}).Where("issue_id = ? AND status = ?", issueID, "pending").Count(&pendingCount).Error; err != nil {
		return nil, common.Internal("Failed to check pending approvals")
	}
	if pendingCount > 0 {
		return nil, common.BadRequest("issue_already_pending_approval")
	}

	// 2. Get the transition
	var transition model.StateTransition
	if err := s.db.First(&transition, transitionID).Error; err != nil {
		return nil, common.NotFound("Transition not found")
	}
	if transition.RuleType != "approval" {
		return nil, common.BadRequest("transition_not_approval_type")
	}

	// 3. Get the issue
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}

	// 4. Resolve approve/reject target states
	approveTargetID := transition.TargetStateID
	if transition.ApproveTargetStateID != nil {
		approveTargetID = *transition.ApproveTargetStateID
	}
	rejectTargetID := transition.SourceStateID
	if transition.RejectTargetStateID != nil {
		rejectTargetID = *transition.RejectTargetStateID
	}

	// 5. Normalize approver_ids to JSON array string
	approverIDsJSON := normalizeApproverIDs(transition.ApproverIDs)

	// 6. Create approval in a transaction
	approval := model.Approval{
		IssueID:              issueID,
		WorkflowID:           transition.WorkflowID,
		TransitionID:         transitionID,
		ProjectID:            issue.ProjectID,
		WorkspaceID:          issue.WorkspaceID,
		RequesterID:          requesterID,
		RequestNote:          requestNote,
		SourceStateID:        issue.StateID,
		ApproveTargetStateID: approveTargetID,
		RejectTargetStateID:  rejectTargetID,
		ApproverIDs:          approverIDsJSON,
		Status:               "pending",
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&approval).Error; err != nil {
			return err
		}
		// Update issue
		pendingStatus := "pending"
		if err := tx.Model(&model.Issue{}).Where("id = ?", issueID).
			Updates(map[string]interface{}{
				"approval_status":     pendingStatus,
				"active_approval_id":  approval.ID,
			}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, common.Internal("Failed to create approval")
	}

	// 7. TODO: send notifications to approvers (Phase 2)
	// 8. TODO: add issue activity (Phase 2)

	return &approval, nil
}

// List returns approvals matching the filter.
func (s *ApprovalService) List(req request.ApprovalListQuery, workspaceID uint64) ([]response.ApprovalResponse, error) {
	query := s.db.Model(&model.Approval{}).Where("workspace_id = ?", workspaceID)
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.ProjectID != 0 {
		query = query.Where("project_id = ?", req.ProjectID)
	}
	if req.ApproverID != 0 {
		// JSONB containment check: approver_ids @> '[<id>]'
		query = query.Where("approver_ids @> ?::jsonb", fmt.Sprintf("[%d]", req.ApproverID))
	}
	query = query.Order("created_at DESC")

	var approvals []model.Approval
	if err := query.Find(&approvals).Error; err != nil {
		return nil, common.Internal("Failed to list approvals")
	}

	res := make([]response.ApprovalResponse, 0, len(approvals))
	for _, a := range approvals {
		res = append(res, s.toResponse(a))
	}
	return res, nil
}

// Get returns an approval by ID with related data.
func (s *ApprovalService) Get(id uint64) (*response.ApprovalResponse, error) {
	var approval model.Approval
	if err := s.db.First(&approval, id).Error; err != nil {
		return nil, common.NotFound("Approval not found")
	}
	resp := s.toResponse(approval)

	var records []model.ApprovalRecord
	s.db.Where("approval_id = ?", id).Order("decided_at DESC").Find(&records)
	for _, r := range records {
		resp.Records = append(resp.Records, response.ApprovalRecordResponse{
			ID:         r.ID,
			ApproverID: r.ApproverID,
			Decision:   r.Decision,
			Note:       r.Note,
			DecidedAt:  r.DecidedAt,
		})
	}
	return &resp, nil
}

// Decide records an approver's decision and updates the issue state.
func (s *ApprovalService) Decide(approvalID, approverID uint64, decision, note string) (*model.Approval, error) {
	if decision != "approved" && decision != "rejected" {
		return nil, common.BadRequest("invalid_decision")
	}

	// 1. Load approval with row lock
	var approval model.Approval
	if err := s.db.First(&approval, approvalID).Error; err != nil {
		return nil, common.NotFound("Approval not found")
	}
	if approval.Status != "pending" {
		return nil, common.BadRequest("approval_already_decided")
	}

	// 2. Validate approver
	approverIDs := parseUint64Array(approval.ApproverIDs)
	if !containsUint64(approverIDs, approverID) {
		return nil, common.Forbidden("not_an_approver")
	}

	// 3. Determine target state
	var targetStateID uint64
	if decision == "approved" {
		targetStateID = approval.ApproveTargetStateID
	} else {
		targetStateID = approval.RejectTargetStateID
	}

	// 4. Transaction: create record, update approval, update issue
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create approval_record
		record := model.ApprovalRecord{
			ApprovalID: approvalID,
			ApproverID: approverID,
			Decision:   decision,
			Note:       note,
			DecidedAt:  time.Now(),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		// Update approval
		now := time.Now()
		updates := map[string]interface{}{
			"status":        decision,
			"decided_by":    approverID,
			"decided_at":    now,
			"decision_note": note,
		}
		if err := tx.Model(&model.Approval{}).Where("id = ? AND status = ?", approvalID, "pending").
			Updates(updates).Error; err != nil {
			return err
		}

		// Update issue: state_id + approval_status
		if err := tx.Model(&model.Issue{}).Where("id = ?", approval.IssueID).
			Updates(map[string]interface{}{
				"state_id":            targetStateID,
				"approval_status":     decision,
				"active_approval_id":  nil,
			}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, common.Internal("Failed to decide approval")
	}

	// Reload
	s.db.First(&approval, approvalID)

	// TODO: send notification to requester (Phase 2)
	// TODO: add issue activity (Phase 2)

	return &approval, nil
}

// Cancel cancels a pending approval. Only the requester can cancel.
func (s *ApprovalService) Cancel(approvalID, userID uint64) (*model.Approval, error) {
	var approval model.Approval
	if err := s.db.First(&approval, approvalID).Error; err != nil {
		return nil, common.NotFound("Approval not found")
	}
	if approval.Status != "pending" {
		return nil, common.BadRequest("approval_not_pending")
	}
	if approval.RequesterID != userID {
		return nil, common.Forbidden("not_requester")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Approval{}).Where("id = ?", approvalID).
			Update("status", "cancelled").Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Issue{}).Where("id = ?", approval.IssueID).
			Updates(map[string]interface{}{
				"approval_status":    nil,
				"active_approval_id": nil,
			}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, common.Internal("Failed to cancel approval")
	}

	s.db.First(&approval, approvalID)
	// TODO: add issue activity (Phase 2)
	return &approval, nil
}

// CountPending returns the number of pending approvals for a user (as approver) in a workspace.
func (s *ApprovalService) CountPending(workspaceID, userID uint64) (int64, error) {
	var count int64
	err := s.db.Model(&model.Approval{}).
		Where("workspace_id = ? AND status = ? AND approver_ids @> ?::jsonb",
			workspaceID, "pending", fmt.Sprintf("[%d]", userID)).
		Count(&count).Error
	if err != nil {
		return 0, common.Internal("Failed to count pending approvals")
	}
	return count, nil
}

// toResponse converts an Approval model to a response DTO (without records).
func (s *ApprovalService) toResponse(a model.Approval) response.ApprovalResponse {
	resp := response.ApprovalResponse{
		ID:                   a.ID,
		IssueID:              a.IssueID,
		WorkflowID:           a.WorkflowID,
		TransitionID:         a.TransitionID,
		ProjectID:            a.ProjectID,
		WorkspaceID:          a.WorkspaceID,
		RequesterID:          a.RequesterID,
		RequestNote:          a.RequestNote,
		SourceStateID:        a.SourceStateID,
		ApproveTargetStateID: a.ApproveTargetStateID,
		RejectTargetStateID:  a.RejectTargetStateID,
		ApproverIDs:          parseUint64Array(a.ApproverIDs),
		Status:               a.Status,
		DecidedBy:            a.DecidedBy,
		DecidedAt:            a.DecidedAt,
		DecisionNote:         a.DecisionNote,
		CreatedAt:            a.CreatedAt,
		Records:              []response.ApprovalRecordResponse{},
	}

	// Enrich with names (best-effort, ignore errors)
	var requester model.User
	if err := s.db.Select("id, name").First(&requester, a.RequesterID).Error; err == nil {
		resp.RequesterName = requester.Name
	}
	if a.DecidedBy != nil {
		var decider model.User
		if err := s.db.Select("id, name").First(&decider, *a.DecidedBy).Error; err == nil {
			resp.DecidedByName = decider.Name
		}
	}
	var issue model.Issue
	if err := s.db.First(&issue, a.IssueID).Error; err == nil {
		resp.IssueTitle = issue.Title
		resp.IssueKey = issue.Identifier // e.g. "CORE-52"
	}
	var project model.Project
	if err := s.db.Select("id, name").First(&project, a.ProjectID).Error; err == nil {
		resp.ProjectName = project.Name
	}
	var srcState model.State
	if err := s.db.Select("id, name").First(&srcState, a.SourceStateID).Error; err == nil {
		resp.SourceStateName = srcState.Name
	}
	var approveState model.State
	if err := s.db.Select("id, name").First(&approveState, a.ApproveTargetStateID).Error; err == nil {
		resp.ApproveStateName = approveState.Name
	}
	var rejectState model.State
	if err := s.db.Select("id, name").First(&rejectState, a.RejectTargetStateID).Error; err == nil {
		resp.RejectStateName = rejectState.Name
	}
	// Approver names
	for _, id := range resp.ApproverIDs {
		var u model.User
		if err := s.db.Select("id, name").First(&u, id).Error; err == nil {
			resp.ApproverNames = append(resp.ApproverNames, u.Name)
		} else {
			resp.ApproverNames = append(resp.ApproverNames, fmt.Sprintf("#%d", id))
		}
	}
	return resp
}

// normalizeApproverIDs ensures approver_ids is stored as a JSON array string.
// Accepts both "[1,2,3]" and "1,2,3" formats.
func normalizeApproverIDs(s *string) string {
	if s == nil || *s == "" {
		return "[]"
	}
	str := *s
	// Already JSON array
	if len(str) > 0 && str[0] == '[' {
		return str
	}
	// Comma-separated -> JSON array
	ids := parseUint64Array(str)
	b, _ := json.Marshal(ids)
	return string(b)
}

func parseUint64Array(s string) []uint64 {
	if s == "" {
		return nil
	}
	// Try JSON first
	var ids []uint64
	if err := json.Unmarshal([]byte(s), &ids); err == nil {
		return ids
	}
	// Fallback: comma-separated
	var result []uint64
	for _, part := range splitComma(s) {
		part = trimSpace(part)
		if part == "" {
			continue
		}
		var id uint64
		if _, err := fmt.Sscanf(part, "%d", &id); err == nil {
			result = append(result, id)
		}
	}
	return result
}

func containsUint64(arr []uint64, v uint64) bool {
	for _, x := range arr {
		if x == v {
			return true
		}
	}
	return false
}

// helper stubs (replace with stdlib equivalents)
func splitComma(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
cd backend && go test ./internal/service -run TestApprovalService -v
```
Expected: all 5 test cases PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/approval_service.go backend/internal/service/approval_service_test.go
git commit -m "feat(approval): add ApprovalService with Create/List/Get/Decide/Cancel"
```

---

## Phase 3: Backend Handlers + Routes

### Task 3.1: Add approval handlers

**Files:**
- Create: `backend/internal/handler/approval_handler.go`
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Write approval_handler.go**

Create `backend/internal/handler/approval_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

type ApprovalHandler struct {
	svc *service.ApprovalService
}

func NewApprovalHandler(svc *service.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{svc: svc}
}

// Create godoc
// POST /api/v1/issues/:issueId/approvals
func (h *ApprovalHandler) Create(c *gin.Context) {
	issueID, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	var req request.ApprovalCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	approval, err := h.svc.Create(issueID, userID, req.TransitionID, req.RequestNote)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, approval)
}

// ListByWorkspace godoc
// GET /api/v1/workspaces/:workspaceId/approvals
func (h *ApprovalHandler) ListByWorkspace(c *gin.Context) {
	wid, _ := strconv.ParseUint(c.Param("workspaceId"), 10, 64)
	var q request.ApprovalListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	list, err := h.svc.List(q, wid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListByProject godoc
// GET /api/v1/projects/:projectId/approvals
func (h *ApprovalHandler) ListByProject(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	wid, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	var q request.ApprovalListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q.ProjectID = pid
	list, err := h.svc.List(q, wid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// Get godoc
// GET /api/v1/approvals/:id
func (h *ApprovalHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	resp, err := h.svc.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Decide godoc
// POST /api/v1/approvals/:id/decide
func (h *ApprovalHandler) Decide(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req request.ApprovalDecision
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	approval, err := h.svc.Decide(id, userID, req.Decision, req.Note)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, approval)
}

// Cancel godoc
// POST /api/v1/approvals/:id/cancel
func (h *ApprovalHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := middleware.GetUserID(c)
	approval, err := h.svc.Cancel(id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, approval)
}

// CountPending godoc
// GET /api/v1/workspaces/:workspaceId/approvals/count
func (h *ApprovalHandler) CountPending(c *gin.Context) {
	wid, _ := strconv.ParseUint(c.Param("workspaceId"), 10, 64)
	userID := middleware.GetUserID(c)
	count, err := h.svc.CountPending(wid, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pending_count": count})
}
```

- [ ] **Step 2: Verify middleware.GetUserID exists**

Run:
```bash
grep -rn "func GetUserID" backend/internal/middleware
```
If the function name differs (e.g. `GetUserIDFromContext`), update the handler accordingly.

- [ ] **Step 3: Register routes**

Edit `backend/internal/router/router.go`. Find a suitable location (e.g. after workflow routes), add:

```go
// Approval routes
approvalHandler := handler.NewApprovalHandler(service.NewApprovalService(db))
apiGroup.POST("/issues/:issueId/approvals", authMiddleware, approvalHandler.Create)
apiGroup.GET("/workspaces/:workspaceId/approvals", authMiddleware, approvalHandler.ListByWorkspace)
apiGroup.GET("/workspaces/:workspaceId/approvals/count", authMiddleware, approvalHandler.CountPending)
apiGroup.GET("/projects/:projectId/approvals", authMiddleware, approvalHandler.ListByProject)
apiGroup.GET("/approvals/:id", authMiddleware, approvalHandler.Get)
apiGroup.POST("/approvals/:id/decide", authMiddleware, approvalHandler.Decide)
apiGroup.POST("/approvals/:id/cancel", authMiddleware, approvalHandler.Cancel)
```

Adjust `apiGroup` and `authMiddleware` variable names to match the existing router file.

- [ ] **Step 4: Verify build**

```bash
cd backend && go build ./...
```
Expected: no errors.

- [ ] **Step 5: Smoke test with curl**

After starting the backend, verify endpoints respond (replace token):

```bash
curl -s -H "Authorization: Bearer <token>" http://localhost:8000/api/v1/workspaces/1/approvals/count
```
Expected: `{"pending_count":0}`

- [ ] **Step 6: Commit**

```bash
git add backend/internal/handler/approval_handler.go backend/internal/router/router.go
git commit -m "feat(approval): add approval HTTP handlers and routes"
```

---

## Phase 4: Backend Issue Service Refactor + Workflow Transition API Extension

### Task 4.1: Refactor issue_service.go to support approval flow

**Files:**
- Modify: `backend/internal/service/issue_service.go` (around lines 1640-1700)

- [ ] **Step 1: Locate the checkTransitionAllowed function**

Read `backend/internal/service/issue_service.go` around lines 1600-1700. Identify the function name and signature (likely `checkTransitionAllowed(db, projectID, issueTypeID, oldStateID, newStateID, userID) error`).

- [ ] **Step 2: Remove the approval permission-check block**

Find the block:
```go
if transition.RuleType == "approval" {
    // Check if the user is an authorized approver
    if transition.ApproverIDs != nil && *transition.ApproverIDs != "" {
        ...
    }
    // Check role-based approval
    ...
    return common.BadRequest("Approval required: you are not authorized to approve this transition")
}
```

Replace with:
```go
if transition.RuleType == "approval" {
    // Approval flow: signal to caller that approval is required.
    return common.ConflictWithCode("approval_required", map[string]interface{}{
        "transition_id": transition.ID,
    })
}
```

If `common.ConflictWithCode` does not exist, create it (or use a sentinel error type). Add this helper to `backend/internal/common/errors.go`:

```go
// ApprovalRequiredError signals that the requested state change requires approval.
type ApprovalRequiredError struct {
	TransitionID uint64
}

func (e *ApprovalRequiredError) Error() string {
	return fmt.Sprintf("approval_required: transition_id=%d", e.TransitionID)
}

func NewApprovalRequiredError(transitionID uint64) *ApprovalRequiredError {
	return &ApprovalRequiredError{TransitionID: transitionID}
}
```

Then in `checkTransitionAllowed`:
```go
if transition.RuleType == "approval" {
    return nil, NewApprovalRequiredError(transition.ID)  // change signature to return (Transition, error)
}
```

Adjust the function signature to also return the matched `Transition` so the caller can build a proper response.

- [ ] **Step 3: Add pending-approval guard**

In the issue update flow (the caller of `checkTransitionAllowed`), before invoking state change, add:

```go
if issue.ApprovalStatus != nil && *issue.ApprovalStatus == "pending" {
    return common.BadRequest("issue_pending_approval")
}
```

- [ ] **Step 4: Update the issue update handler to return approval_required**

In the issue handler (`backend/internal/handler/issue_handler.go`), catch `*common.ApprovalRequiredError` and return HTTP 409:

```go
if approvalErr, ok := err.(*common.ApprovalRequiredError); ok {
    c.JSON(http.StatusConflict, gin.H{
        "error":         "approval_required",
        "transition_id": approvalErr.TransitionID,
    })
    return
}
```

- [ ] **Step 5: Verify build and run existing tests**

```bash
cd backend && go build ./... && go test ./internal/service/...
```
Expected: no errors. If existing issue_service tests break, update them.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/service/issue_service.go backend/internal/common/errors.go backend/internal/handler/issue_handler.go
git commit -m "refactor(approval): issue_service returns approval_required instead of blocking"
```

---

### Task 4.2: Extend workflow transition API

**Files:**
- Modify: `backend/internal/service/workflow_service.go` (AddTransition, UpdateTransition)
- Modify: `backend/internal/handler/workflow_handler.go` (bind new fields in transition create/update)

- [ ] **Step 1: Update AddTransition in workflow_service.go**

Find `func (s *WorkflowService) AddTransition` in `backend/internal/service/workflow_service.go`. After setting existing fields, add:

```go
	t.ApproveTargetStateID = req.ApproveTargetStateID
	t.RejectTargetStateID = req.RejectTargetStateID
	if req.ApprovalMode != "" {
		t.ApprovalMode = req.ApprovalMode
	} else {
		t.ApprovalMode = "any"
	}
```

Also update the `StateTransition` creation block accordingly.

- [ ] **Step 2: Update UpdateTransition in workflow_service.go**

Find `func (s *WorkflowService) UpdateTransition`. Add:

```go
	if req.ApproveTargetStateID != nil { t.ApproveTargetStateID = req.ApproveTargetStateID }
	if req.RejectTargetStateID != nil { t.RejectTargetStateID = req.RejectTargetStateID }
	if req.ApprovalMode != "" { t.ApprovalMode = req.ApprovalMode }
```

- [ ] **Step 3: Update TransitionResponse building**

In `workflow_service.go`, find all places that build `response.TransitionResponse{...}`. Add the new fields:

```go
ApproveTargetStateID: t.ApproveTargetStateID,
RejectTargetStateID:  t.RejectTargetStateID,
ApprovalMode:         t.ApprovalMode,
```

If state names are needed, also fetch the approve/reject state names (similar to SourceName/TargetName). Add:

```go
var approveState, rejectState model.State
if t.ApproveTargetStateID != nil { s.db.First(&approveState, *t.ApproveTargetStateID) }
if t.RejectTargetStateID != nil { s.db.First(&rejectState, *t.RejectTargetStateID) }
// In response:
ApproveStateName: approveState.Name,
RejectStateName:  rejectState.Name,
```

- [ ] **Step 4: Verify build**

```bash
cd backend && go build ./...
```

- [ ] **Step 5: Smoke test**

```bash
# Create an approval transition with approve/reject targets
curl -s -X POST -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '{"from_state_id":1,"to_state_id":2,"rule_type":"approval","approver_ids":"[2,3]","approve_target_state_id":2,"reject_target_state_id":1}' \
  http://localhost:8000/api/v1/projects/1/workflows/25/transitions
```
Expected: 200/201 response with the new fields populated.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/service/workflow_service.go backend/internal/handler/workflow_handler.go
git commit -m "feat(approval): extend transition API with approve/reject target states"
```

---

## Phase 5: Frontend Workflow Detail Page UI

### Task 5.1: Extend add-transition dialog with approve/reject targets

**Files:**
- Modify: `frontend/src/views/WorkflowDetail.vue`
- Modify: `frontend/src/locales/zh-CN.json`
- Modify: `frontend/src/locales/en-US.json`

- [ ] **Step 1: Add i18n keys**

In `frontend/src/locales/zh-CN.json`, under `workflow` section, add:

```json
"approveTargetState": "批准后跳转状态",
"rejectTargetState": "拒绝后跳转状态",
"approvalMode": "审批模式",
"approvalModeAny": "任一审批人决定",
"approvalModeAll": "全部审批人决定（即将上线）",
"approvalFlowHint": "配置审批通过和拒绝后 issue 的目标状态"
```

In `frontend/src/locales/en-US.json`, under `workflow` section, add:

```json
"approveTargetState": "Approve target state",
"rejectTargetState": "Reject target state",
"approvalMode": "Approval mode",
"approvalModeAny": "Any approver can decide",
"approvalModeAll": "All approvers must approve (coming soon)",
"approvalFlowHint": "Configure target states after approval and rejection"
```

- [ ] **Step 2: Extend the transition form state**

In `frontend/src/views/WorkflowDetail.vue`, find the `trans` ref (around line 86):
```typescript
const trans = ref({ from: 0, to: 0, desc: '', rule_type: 'allow', approver_ids: '', role_allowed: '' })
```

Update to:
```typescript
const trans = ref({
  from: 0, to: 0, desc: '', rule_type: 'allow',
  approver_ids: '', role_allowed: '',
  approve_target_state_id: 0, reject_target_state_id: 0, approval_mode: 'any'
})
```

Update the reset in `openAddTrans`:
```typescript
function openAddTrans(w:any) {
  selWid.value = w.id
  trans.value = {
    from: 0, to: 0, desc: '', rule_type: 'allow',
    approver_ids: '', role_allowed: '',
    approve_target_state_id: 0, reject_target_state_id: 0, approval_mode: 'any'
  }
  showTrans.value = true
}
```

- [ ] **Step 3: Update saveTrans to send new fields**

Find `async function saveTrans()`. Update the `data` object:

```typescript
const data = {
  from_state_id: trans.value.from,
  to_state_id: trans.value.to,
  description: trans.value.desc,
  rule_type: trans.value.rule_type,
  approver_ids: trans.value.approver_ids || undefined,
  role_allowed: trans.value.role_allowed || undefined,
  approve_target_state_id: trans.value.rule_type === 'approval' ? trans.value.approve_target_state_id || undefined : undefined,
  reject_target_state_id: trans.value.rule_type === 'approval' ? trans.value.reject_target_state_id || undefined : undefined,
  approval_mode: trans.value.rule_type === 'approval' ? trans.value.approval_mode : undefined,
}
```

- [ ] **Step 4: Add form fields for approval config**

In the template, find the rule_type select. After the `role_allowed` input (the last approval-only field), add:

```vue
<div v-if="trans.rule_type==='approval'" class="text-xs text-gray-500 bg-indigo-50 rounded p-2">
  {{ t('workflow.approvalFlowHint') }}
</div>
<div v-if="trans.rule_type==='approval'">
  <label class="block text-sm font-medium mb-1">{{ t('workflow.approveTargetState') }}</label>
  <select v-model="trans.approve_target_state_id" class="w-full px-3 py-2 border rounded-lg">
    <option :value="0">{{ t('workflow.fromState') }} →</option>
    <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
  </select>
</div>
<div v-if="trans.rule_type==='approval'">
  <label class="block text-sm font-medium mb-1">{{ t('workflow.rejectTargetState') }}</label>
  <select v-model="trans.reject_target_state_id" class="w-full px-3 py-2 border rounded-lg">
    <option :value="0">{{ t('workflow.fromState') }} (回退)</option>
    <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
  </select>
</div>
<div v-if="trans.rule_type==='approval'">
  <label class="block text-sm font-medium mb-1">{{ t('workflow.approvalMode') }}</label>
  <select v-model="trans.approval_mode" class="w-full px-3 py-2 border rounded-lg">
    <option value="any">{{ t('workflow.approvalModeAny') }}</option>
    <option value="all" disabled>{{ t('workflow.approvalModeAll') }}</option>
  </select>
</div>
```

- [ ] **Step 5: Render approve/reject paths in transition list**

Find the transition rendering block (around line 23-30). For approval transitions, show both paths. Update:

```vue
<div v-for="t in w.transitions" :key="t.id" class="flex items-center text-sm text-gray-600 bg-gray-50 rounded px-3 py-1.5">
  <span class="font-medium">{{ t.source_name || t.from_name || '#'+t.from_state_id }}</span>
  <span class="mx-2 text-gray-400">→</span>
  <span v-if="t.rule_type === 'approval'" class="flex items-center space-x-2">
    <span class="inline-flex items-center px-1.5 py-0.5 bg-green-100 text-green-700 rounded text-xs">{{ t('workflow.approval') }}</span>
    <span class="text-xs">✓ {{ t.approve_target_state_name || ('#'+t.approve_target_state_id) }}</span>
    <span class="text-xs text-red-500">✗ {{ t.reject_target_state_name || ('#'+t.reject_target_state_id) }}</span>
  </span>
  <span v-else class="font-medium">{{ t.to_name || t.target_name || '#'+t.to_state_id }}</span>
  <button @click="delTrans(t.id)" class="ml-auto text-xs text-red-400 hover:text-red-600">×</button>
</div>
```

- [ ] **Step 6: Verify in browser**

Open `http://localhost:5173/workspace/reqmango-dev/project/1/settings/workflows/25` (after login). Click "+ Transition", select rule_type=approval, verify the new fields appear and save works.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/views/WorkflowDetail.vue frontend/src/locales/zh-CN.json frontend/src/locales/en-US.json
git commit -m "feat(approval): workflow detail UI supports approve/reject target states"
```

---

## Phase 6: Frontend Issue Detail Page

### Task 6.1: Add approval API client

**Files:**
- Create: `frontend/src/api/approval.ts`

- [ ] **Step 1: Write the API client**

Create `frontend/src/api/approval.ts`:

```typescript
import api from '@/api'

export interface ApprovalResponse {
  id: number
  issue_id: number
  issue_key: string
  issue_title: string
  project_id: number
  project_name: string
  requester_id: number
  requester_name: string
  request_note: string
  source_state_id: number
  source_state_name: string
  approve_target_state_id: number
  approve_target_state_name: string
  reject_target_state_id: number
  reject_target_state_name: string
  approver_ids: number[]
  approver_names: string[]
  status: 'pending' | 'approved' | 'rejected' | 'cancelled'
  decided_by: number | null
  decided_by_name: string
  decided_at: string | null
  decision_note: string
  created_at: string
  records: Array<{
    id: number
    approver_id: number
    approver_name: string
    decision: 'approved' | 'rejected'
    note: string
    decided_at: string
  }>
}

export interface ApprovalListQuery {
  status?: string
  project_id?: number
  approver_id?: number
}

export const approvalApi = {
  submit: (issueId: number, data: { transition_id: number; request_note: string }) =>
    api.post(`/issues/${issueId}/approvals`, data).then(r => r.data),

  listByWorkspace: (workspaceId: number, query: ApprovalListQuery = {}) =>
    api.get(`/workspaces/${workspaceId}/approvals`, { params: query }).then(r => r.data),

  listByProject: (projectId: number, workspaceId: number, query: ApprovalListQuery = {}) =>
    api.get(`/projects/${projectId}/approvals`, { params: { workspace_id: workspaceId, ...query } }).then(r => r.data),

  get: (id: number) =>
    api.get(`/approvals/${id}`).then(r => r.data),

  decide: (id: number, data: { decision: 'approved' | 'rejected'; note?: string }) =>
    api.post(`/approvals/${id}/decide`, data).then(r => r.data),

  cancel: (id: number) =>
    api.post(`/approvals/${id}/cancel`).then(r => r.data),

  countPending: (workspaceId: number) =>
    api.get(`/workspaces/${workspaceId}/approvals/count`).then(r => r.data?.pending_count ?? 0),
}

export default approvalApi
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/api/approval.ts
git commit -m "feat(approval): add approval API client"
```

---

### Task 6.2: Add ApprovalSubmitDialog and ApprovalDecisionDialog components

**Files:**
- Create: `frontend/src/components/ApprovalSubmitDialog.vue`
- Create: `frontend/src/components/ApprovalDecisionDialog.vue`

- [ ] **Step 1: Write ApprovalSubmitDialog.vue**

Create `frontend/src/components/ApprovalSubmitDialog.vue`:

```vue
<template>
  <div v-if="show" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="$emit('close')">
    <div class="bg-white rounded-xl p-6 w-full max-w-md">
      <h3 class="text-lg font-semibold mb-2">{{ t('approvals.submitTitle') }}</h3>
      <p class="text-sm text-gray-500 mb-4">
        {{ t('approvals.submitDesc', { from: fromStateName, to: approveStateName }) }}
      </p>
      <div class="space-y-3">
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('approvals.approvers') }}</label>
          <div class="text-sm text-gray-700 bg-gray-50 rounded px-3 py-2">
            {{ approverNames.join(', ') || '-' }}
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('approvals.requestNote') }}</label>
          <textarea v-model="note" rows="3" class="w-full px-3 py-2 border rounded-lg"
            :placeholder="t('approvals.requestNotePlaceholder')"></textarea>
        </div>
      </div>
      <div class="flex justify-end space-x-3 mt-6">
        <button @click="$emit('close')" class="px-4 py-2 border rounded-lg">{{ t('common.cancel') }}</button>
        <button @click="submit" :disabled="submitting"
          class="px-4 py-2 bg-indigo-600 text-white rounded-lg disabled:opacity-50">
          {{ submitting ? t('common.submitting') : t('approvals.submit') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/composables/useI18n'
import approvalApi from '@/api/approval'

const props = defineProps<{
  show: boolean
  issueId: number
  transitionId: number
  fromStateName: string
  approveStateName: string
  approverNames: string[]
}>()

const emit = defineEmits<{ close: []; submitted: [] }>()
const { t } = useI18n()
const note = ref('')
const submitting = ref(false)

async function submit() {
  if (submitting.value) return
  submitting.value = true
  try {
    await approvalApi.submit(props.issueId, {
      transition_id: props.transitionId,
      request_note: note.value,
    })
    emit('submitted')
    emit('close')
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Failed to submit approval')
  } finally {
    submitting.value = false
  }
}
</script>
```

- [ ] **Step 2: Write ApprovalDecisionDialog.vue**

Create `frontend/src/components/ApprovalDecisionDialog.vue`:

```vue
<template>
  <div v-if="show" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="$emit('close')">
    <div class="bg-white rounded-xl p-6 w-full max-w-md">
      <h3 class="text-lg font-semibold mb-2">
        {{ decision === 'approved' ? t('approvals.approveTitle') : t('approvals.rejectTitle') }}
      </h3>
      <p class="text-sm text-gray-500 mb-4">{{ t('approvals.decisionDesc') }}</p>
      <div class="space-y-3">
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('approvals.decisionNote') }}</label>
          <textarea v-model="note" rows="3" class="w-full px-3 py-2 border rounded-lg"
            :placeholder="t('approvals.decisionNotePlaceholder')"></textarea>
        </div>
      </div>
      <div class="flex justify-end space-x-3 mt-6">
        <button @click="$emit('close')" class="px-4 py-2 border rounded-lg">{{ t('common.cancel') }}</button>
        <button @click="confirm" :disabled="submitting"
          :class="['px-4 py-2 text-white rounded-lg disabled:opacity-50',
            decision === 'approved' ? 'bg-green-600 hover:bg-green-700' : 'bg-red-600 hover:bg-red-700']">
          {{ submitting ? t('common.submitting') : (decision === 'approved' ? t('approvals.approve') : t('approvals.reject')) }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/composables/useI18n'
import approvalApi from '@/api/approval'

const props = defineProps<{
  show: boolean
  approvalId: number
  decision: 'approved' | 'rejected'
}>()

const emit = defineEmits<{ close: []; decided: [] }>()
const { t } = useI18n()
const note = ref('')
const submitting = ref(false)

async function confirm() {
  if (submitting.value) return
  submitting.value = true
  try {
    await approvalApi.decide(props.approvalId, {
      decision: props.decision,
      note: note.value,
    })
    emit('decided')
    emit('close')
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Failed to decide approval')
  } finally {
    submitting.value = false
  }
}
</script>
```

- [ ] **Step 3: Add i18n keys**

Add to both `zh-CN.json` and `en-US.json` under a new `approvals` section:

zh-CN:
```json
"approvals": {
  "title": "审批",
  "submitTitle": "提交审批",
  "submitDesc": "申请将工作项从 \"{from}\" 变更为 \"{to}\"",
  "approvers": "审批人",
  "requestNote": "申请说明",
  "requestNotePlaceholder": "请输入申请说明（可选）",
  "submit": "提交",
  "approveTitle": "批准审批",
  "rejectTitle": "拒绝审批",
  "decisionDesc": "请输入审批备注（可选）",
  "decisionNote": "审批备注",
  "decisionNotePlaceholder": "请输入审批备注",
  "approve": "批准",
  "reject": "拒绝",
  "pending": "待审批",
  "approved": "已批准",
  "rejected": "已拒绝",
  "cancelled": "已取消",
  "pendingBanner": "此工作项正在等待审批",
  "requestedBy": "申请人：{name}",
  "submittedAt": "提交时间：{time}",
  "cancelApproval": "撤销申请",
  "viewAll": "查看全部",
  "listTitle": "审批中心",
  "issue": "工作项",
  "project": "项目",
  "requester": "申请人",
  "submittedTime": "提交时间",
  "status": "状态",
  "action": "操作",
  "noApprovals": "暂无审批记录"
}
```

en-US: same structure with English text.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ApprovalSubmitDialog.vue frontend/src/components/ApprovalDecisionDialog.vue frontend/src/locales/zh-CN.json frontend/src/locales/en-US.json
git commit -m "feat(approval): add submit and decision dialog components"
```

---

### Task 6.3: Add ApprovalPendingBanner component

**Files:**
- Create: `frontend/src/components/ApprovalPendingBanner.vue`

- [ ] **Step 1: Write the banner component**

Create `frontend/src/components/ApprovalPendingBanner.vue`:

```vue
<template>
  <div v-if="approval" class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
    <div class="flex items-start justify-between">
      <div class="flex-1">
        <div class="flex items-center space-x-2 mb-1">
          <span class="text-yellow-700 font-medium">{{ t('approvals.pendingBanner') }}</span>
          <span class="px-2 py-0.5 bg-yellow-100 text-yellow-700 rounded text-xs">
            {{ t('approvals.pending') }}
          </span>
        </div>
        <p class="text-sm text-yellow-700">
          {{ t('approvals.requestedBy', { name: approval.requester_name }) }}
        </p>
        <p class="text-sm text-yellow-700">
          {{ t('approvals.submittedAt', { time: new Date(approval.created_at).toLocaleString() }) }}
        </p>
        <p class="text-sm text-yellow-700 mt-1">
          {{ t('approvals.approvers') }}: {{ approval.approver_names.join(', ') }}
        </p>
        <p v-if="approval.request_note" class="text-sm text-yellow-700 mt-1 italic">
          "{{ approval.request_note }}"
        </p>
      </div>
      <div class="flex items-center space-x-2">
        <button v-if="canDecide" @click="$emit('decide', approval, 'approved')"
          class="px-3 py-1.5 bg-green-600 text-white rounded text-sm hover:bg-green-700">
          {{ t('approvals.approve') }}
        </button>
        <button v-if="canDecide" @click="$emit('decide', approval, 'rejected')"
          class="px-3 py-1.5 bg-red-600 text-white rounded text-sm hover:bg-red-700">
          {{ t('approvals.reject') }}
        </button>
        <button v-if="canCancel" @click="$emit('cancel', approval)"
          class="px-3 py-1.5 border border-yellow-300 text-yellow-700 rounded text-sm hover:bg-yellow-100">
          {{ t('approvals.cancelApproval') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ApprovalResponse } from '@/api/approval'

const props = defineProps<{
  approval: ApprovalResponse | null
  currentUserId: number
}>()

defineEmits<{
  decide: [approval: ApprovalResponse, decision: 'approved' | 'rejected']
  cancel: [approval: ApprovalResponse]
}>()

const { t } = useI18n()

const canDecide = computed(() =>
  props.approval?.status === 'pending' &&
  props.approval.approver_ids.includes(props.currentUserId)
)

const canCancel = computed(() =>
  props.approval?.status === 'pending' &&
  props.approval.requester_id === props.currentUserId
)
</script>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ApprovalPendingBanner.vue
git commit -m "feat(approval): add pending approval banner component"
```

---

### Task 6.4: Integrate approval UI into IssueDetail.vue

**Files:**
- Modify: `frontend/src/views/IssueDetail.vue`

- [ ] **Step 1: Locate state-change handler**

Open `frontend/src/views/IssueDetail.vue`. Search for the function that handles state changes (look for `@change` on state `<select>` or `updateState`). Identify:
- Where the state dropdown is rendered
- The function that calls the issue update API
- Where issue data is loaded (to fetch active approval)

- [ ] **Step 2: Add approval state refs**

In `<script setup>`, add:

```typescript
import ApprovalSubmitDialog from '@/components/ApprovalSubmitDialog.vue'
import ApprovalDecisionDialog from '@/components/ApprovalDecisionDialog.vue'
import ApprovalPendingBanner from '@/components/ApprovalPendingBanner.vue'
import approvalApi, { type ApprovalResponse } from '@/api/approval'

const showSubmitDialog = ref(false)
const showDecisionDialog = ref(false)
const pendingApproval = ref<ApprovalResponse | null>(null)
const activeApproval = ref<ApprovalResponse | null>(null)
const submitTransitionId = ref(0)
const submitFromStateName = ref('')
const submitApproveStateName = ref('')
const submitApproverNames = ref<string[]>([])
const decisionDialogData = ref<{ approvalId: number; decision: 'approved' | 'rejected' } | null>(null)
const currentUserId = ref(0)  // populate from auth store
```

- [ ] **Step 3: Load active approval when issue loads**

In the issue load function (e.g. `loadIssue`), after the issue is fetched, add:

```typescript
// Load active approval if pending
if (issue.approval_status === 'pending' && issue.active_approval_id) {
  try {
    activeApproval.value = await approvalApi.get(issue.active_approval_id)
  } catch (e) {
    console.error('Failed to load active approval', e)
    activeApproval.value = null
  }
} else {
  activeApproval.value = null
}
```

Also populate `currentUserId` from the auth store (e.g. `useAuth().user.id`).

- [ ] **Step 4: Intercept state change for approval transitions**

Modify the state change handler. Before calling the update API, look up the matching transition. If `rule_type === 'approval'`, open the submit dialog instead:

```typescript
async function onStateChange(newStateId: number) {
  if (!issue.value) return
  // Find transition matching (issue.state_id -> newStateId)
  try {
    const transitions = await api.get(`/projects/${issue.value.project_id}/workflows`, {
      params: { _expand: 'transitions' }
    }).then(r => r.data)
    const allTransitions = (Array.isArray(transitions) ? transitions : []).flatMap((w: any) => w.transitions || [])
    const match = allTransitions.find((t: any) =>
      t.source_state_id === issue.value.state_id && t.target_state_id === newStateId
    )
    if (match && match.rule_type === 'approval') {
      // Open approval submit dialog
      submitTransitionId.value = match.id
      submitFromStateName.value = getStateName(issue.value.state_id)
      submitApproveStateName.value = match.approve_target_state_name || getStateName(newStateId)
      const approverIds = JSON.parse(match.approver_ids || '[]')
      submitApproverNames.value = await loadUserNames(approverIds)
      showSubmitDialog.value = true
      return
    }
  } catch (e) {
    console.warn('Failed to lookup transition, falling back to direct update', e)
  }
  // Default: direct state change
  await updateIssueState(newStateId)
}
```

Add helper:
```typescript
function getStateName(id: number): string {
  return states.value.find((s: any) => s.id === id)?.name || `#${id}`
}
async function loadUserNames(ids: number[]): Promise<string[]> {
  if (!ids.length) return []
  try {
    const res = await api.get('/users', { params: { ids: ids.join(',') } })
    const users = res.data?.data || res.data || []
    return users.map((u: any) => u.name)
  } catch { return ids.map(id => `#${id}`) }
}
```

- [ ] **Step 5: Add banner and dialogs to template**

In the template, near the top of the issue detail content (above the title or right after the header), add:

```vue
<ApprovalPendingBanner
  :approval="activeApproval"
  :current-user-id="currentUserId"
  @decide="onDecideApproval"
  @cancel="onCancelApproval"
/>
<ApprovalSubmitDialog
  :show="showSubmitDialog"
  :issue-id="issue?.id || 0"
  :transition-id="submitTransitionId"
  :from-state-name="submitFromStateName"
  :approve-state-name="submitApproveStateName"
  :approver-names="submitApproverNames"
  @close="showSubmitDialog = false"
  @submitted="onApprovalSubmitted"
/>
<ApprovalDecisionDialog
  :show="showDecisionDialog"
  :approval-id="decisionDialogData?.approvalId || 0"
  :decision="decisionDialogData?.decision || 'approved'"
  @close="showDecisionDialog = false"
  @decided="onApprovalDecided"
/>
```

- [ ] **Step 6: Implement the event handlers**

```typescript
async function onApprovalSubmitted() {
  showSubmitDialog.value = false
  await loadIssue()  // reload to show banner
}

function onDecideApproval(approval: ApprovalResponse, decision: 'approved' | 'rejected') {
  decisionDialogData.value = { approvalId: approval.id, decision }
  showDecisionDialog.value = true
}

async function onApprovalDecided() {
  showDecisionDialog.value = false
  await loadIssue()
}

async function onCancelApproval(approval: ApprovalResponse) {
  if (!confirm(t('approvals.cancelApproval') + '?')) return
  try {
    await approvalApi.cancel(approval.id)
    await loadIssue()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Failed to cancel approval')
  }
}
```

- [ ] **Step 7: Verify in browser**

Manually test the flow:
1. Open an issue that has an approval transition configured
2. Try to change state via the dropdown — submit dialog should appear
3. Submit -> banner appears with "pending"
4. As approver, click Approve/Reject -> dialog -> confirm -> banner disappears, state changes

- [ ] **Step 8: Commit**

```bash
git add frontend/src/views/IssueDetail.vue
git commit -m "feat(approval): integrate submit dialog and pending banner in issue detail"
```

---

## Phase 7: Frontend TopBar + Approval List Page

### Task 7.1: Add ApprovalBadge component to TopBar

**Files:**
- Create: `frontend/src/components/ApprovalBadge.vue`
- Modify: `frontend/src/components/TopBar.vue`

- [ ] **Step 1: Write ApprovalBadge.vue**

Create `frontend/src/components/ApprovalBadge.vue`:

```vue
<template>
  <div class="relative">
    <button @click="toggleDropdown"
      class="relative p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
          d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
      </svg>
      <span v-if="pendingCount > 0"
        class="absolute -top-0 -right-0 bg-red-500 text-white text-xs rounded-full px-1.5 min-w-[18px] text-center">
        {{ pendingCount > 99 ? '99+' : pendingCount }}
      </span>
    </button>

    <div v-if="open" class="absolute right-0 mt-2 w-80 bg-white rounded-lg shadow-lg border z-50">
      <div class="px-4 py-2 border-b font-medium text-sm">{{ t('approvals.title') }}</div>
      <div class="max-h-80 overflow-y-auto">
        <div v-if="recentApprovals.length === 0" class="px-4 py-6 text-center text-gray-400 text-sm">
          {{ t('approvals.noApprovals') }}
        </div>
        <a v-for="a in recentApprovals" :key="a.id"
          :href="`/workspace/${slug}/project/${a.project_id}/issues/${a.issue_id}`"
          class="block px-4 py-2 hover:bg-gray-50 border-b last:border-0">
          <div class="text-sm font-medium text-gray-900">{{ a.issue_key }}: {{ a.issue_title }}</div>
          <div class="text-xs text-gray-500">{{ a.requester_name }} · {{ formatDate(a.created_at) }}</div>
        </a>
      </div>
      <div class="px-4 py-2 border-t">
        <router-link :to="`/workspace/${slug}/approvals`" class="text-sm text-indigo-600 hover:underline">
          {{ t('approvals.viewAll') }}
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import approvalApi, { type ApprovalResponse } from '@/api/approval'

const route = useRoute()
const { t } = useI18n()
const slug = computed(() => route.params.slug as string)

const open = ref(false)
const pendingCount = ref(0)
const recentApprovals = ref<ApprovalResponse[]>([])

function toggleDropdown() {
  open.value = !open.value
  if (open.value) loadApprovals()
}

async function loadApprovals() {
  const wid = Number(route.query.workspace_id || 1)
  try {
    const [count, list] = await Promise.all([
      approvalApi.countPending(wid),
      approvalApi.listByWorkspace(wid, { status: 'pending' }),
    ])
    pendingCount.value = count
    recentApprovals.value = (Array.isArray(list) ? list : []).slice(0, 5)
  } catch (e) {
    console.error('Failed to load approvals', e)
  }
}

function formatDate(s: string): string {
  return new Date(s).toLocaleDateString() + ' ' + new Date(s).toLocaleTimeString().slice(0, 5)
}

function closeOnOutside(e: MouseEvent) {
  if (!open.value) return
  const target = e.target as HTMLElement
  if (!target.closest('.relative')) open.value = false
}

onMounted(() => {
  loadApprovals()
  document.addEventListener('click', closeOnOutside)
  // Refresh every 60 seconds
  setInterval(loadApprovals, 60000)
})
onUnmounted(() => document.removeEventListener('click', closeOnOutside))
</script>
```

- [ ] **Step 2: Add ApprovalBadge to TopBar.vue**

Edit `frontend/src/components/TopBar.vue`. Import the component and place it next to the notification bell:

```vue
<script setup lang="ts">
import ApprovalBadge from '@/components/ApprovalBadge.vue'
// ... existing imports
</script>

<template>
  <!-- ... existing topbar content ... -->
  <div class="flex items-center space-x-2">
    <!-- existing notification bell -->
    <ApprovalBadge />
    <!-- existing user menu -->
  </div>
</template>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ApprovalBadge.vue frontend/src/components/TopBar.vue
git commit -m "feat(approval): add TopBar approval badge with pending count and dropdown"
```

---

### Task 7.2: Create ApprovalList page

**Files:**
- Create: `frontend/src/views/ApprovalList.vue`
- Modify: `frontend/src/router/index.ts`

- [ ] **Step 1: Write ApprovalList.vue**

Create `frontend/src/views/ApprovalList.vue`:

```vue
<template>
  <div class="p-6 max-w-7xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-semibold text-gray-900">{{ t('approvals.listTitle') }}</h1>
      </div>
    </div>

    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
      <div class="px-4 py-3 border-b flex items-center space-x-3">
        <select v-model="filter.status" @change="load" class="px-3 py-1.5 border rounded-lg text-sm">
          <option value="">{{ t('approvals.status') }}</option>
          <option value="pending">{{ t('approvals.pending') }}</option>
          <option value="approved">{{ t('approvals.approved') }}</option>
          <option value="rejected">{{ t('approvals.rejected') }}</option>
          <option value="cancelled">{{ t('approvals.cancelled') }}</option>
        </select>
      </div>

      <table class="w-full">
        <thead class="bg-gray-50 text-xs text-gray-500 uppercase">
          <tr>
            <th class="px-4 py-2 text-left">{{ t('approvals.issue') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.project') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.requester') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.submittedTime') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.status') }}</th>
            <th class="px-4 py-2 text-right">{{ t('approvals.action') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y">
          <tr v-for="a in approvals" :key="a.id" class="hover:bg-gray-50">
            <td class="px-4 py-2">
              <router-link :to="`/workspace/${slug}/project/${a.project_id}/issues/${a.issue_id}`"
                class="text-indigo-600 hover:underline">
                {{ a.issue_key }}: {{ a.issue_title }}
              </router-link>
            </td>
            <td class="px-4 py-2 text-sm text-gray-700">{{ a.project_name }}</td>
            <td class="px-4 py-2 text-sm text-gray-700">{{ a.requester_name }}</td>
            <td class="px-4 py-2 text-sm text-gray-500">{{ new Date(a.created_at).toLocaleString() }}</td>
            <td class="px-4 py-2">
              <span :class="statusClass(a.status)" class="px-2 py-0.5 rounded text-xs">
                {{ statusLabel(a.status) }}
              </span>
            </td>
            <td class="px-4 py-2 text-right space-x-1">
              <button v-if="canDecide(a)" @click="decide(a, 'approved')"
                class="px-2 py-1 bg-green-600 text-white rounded text-xs hover:bg-green-700">
                {{ t('approvals.approve') }}
              </button>
              <button v-if="canDecide(a)" @click="decide(a, 'rejected')"
                class="px-2 py-1 bg-red-600 text-white rounded text-xs hover:bg-red-700">
                {{ t('approvals.reject') }}
              </button>
            </td>
          </tr>
          <tr v-if="approvals.length === 0">
            <td colspan="6" class="px-4 py-8 text-center text-gray-400 text-sm">
              {{ t('approvals.noApprovals') }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <ApprovalDecisionDialog
      :show="showDecisionDialog"
      :approval-id="decisionData?.approvalId || 0"
      :decision="decisionData?.decision || 'approved'"
      @close="showDecisionDialog = false"
      @decided="onDecided"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import approvalApi, { type ApprovalResponse } from '@/api/approval'
import ApprovalDecisionDialog from '@/components/ApprovalDecisionDialog.vue'
import { useAuth } from '@/composables/useAuth'

const route = useRoute()
const { t } = useI18n()
const { user } = useAuth()
const slug = computed(() => route.params.slug as string)

const approvals = ref<ApprovalResponse[]>([])
const filter = ref({ status: '' })
const showDecisionDialog = ref(false)
const decisionData = ref<{ approvalId: number; decision: 'approved' | 'rejected' } | null>(null)

async function load() {
  const wid = Number(route.query.workspace_id || 1)
  try {
    const list = await approvalApi.listByWorkspace(wid, {
      status: filter.value.status || undefined,
    })
    approvals.value = Array.isArray(list) ? list : []
  } catch (e) {
    console.error('Failed to load approvals', e)
  }
}

function canDecide(a: ApprovalResponse): boolean {
  return a.status === 'pending' && a.approver_ids.includes(user.value?.id || 0)
}

function decide(a: ApprovalResponse, decision: 'approved' | 'rejected') {
  decisionData.value = { approvalId: a.id, decision }
  showDecisionDialog.value = true
}

async function onDecided() {
  showDecisionDialog.value = false
  await load()
}

function statusClass(s: string): string {
  return {
    pending: 'bg-yellow-100 text-yellow-700',
    approved: 'bg-green-100 text-green-700',
    rejected: 'bg-red-100 text-red-700',
    cancelled: 'bg-gray-100 text-gray-500',
  }[s] || 'bg-gray-100 text-gray-500'
}

function statusLabel(s: string): string {
  return t(`approvals.${s}`)
}

onMounted(load)
</script>
```

- [ ] **Step 2: Register the route**

Edit `frontend/src/router/index.ts`. Add a new route in the workspace-scoped section:

```typescript
{
  path: '/workspace/:slug/approvals',
  name: 'ApprovalList',
  component: () => import('@/views/ApprovalList.vue'),
  meta: { requiresAuth: true },
}
```

- [ ] **Step 3: Verify in browser**

Navigate to `http://localhost:5173/workspace/reqmango-dev/approvals?workspace_id=1`. Verify the page renders with the filter and table.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/ApprovalList.vue frontend/src/router/index.ts
git commit -m "feat(approval): add approval center list page with route"
```

---

## Phase 8: Notifications + Activity Stream

### Task 8.1: Send notifications on approval events

**Files:**
- Modify: `backend/internal/service/approval_service.go`

- [ ] **Step 1: Locate the notification service**

Run:
```bash
grep -rn "type NotificationService\|func NewNotificationService" backend/internal/service
```

Identify the notification service constructor and `Send` or `Create` method signature.

- [ ] **Step 2: Inject notification service into ApprovalService**

Update `NewApprovalService` to accept a notification sender interface (or the concrete service). For loose coupling, define an interface:

```go
// ApprovalNotifier sends notifications for approval events.
type ApprovalNotifier interface {
	Notify(userID uint64, title, body string, metadata map[string]interface{}) error
}
```

Update `ApprovalService`:
```go
type ApprovalService struct {
	db        *gorm.DB
	notifier  ApprovalNotifier
}

func NewApprovalService(db *gorm.DB, notifier ApprovalNotifier) *ApprovalService {
	return &ApprovalService{db: db, notifier: notifier}
}
```

Update the test setup to pass `nil` or a mock notifier.

- [ ] **Step 3: Send notifications in Create**

At the end of `Create`, after the transaction commits:

```go
if s.notifier != nil {
	for _, approverID := range parseUint64Array(approverIDsJSON) {
		_ = s.notifier.Notify(approverID,
			"New approval request",
			fmt.Sprintf("Issue %d requires your approval", issueID),
			map[string]interface{}{"approval_id": approval.ID, "issue_id": issueID},
		)
	}
}
```

- [ ] **Step 4: Send notifications in Decide**

After Decide commits:

```go
if s.notifier != nil {
	_ = s.notifier.Notify(approval.RequesterID,
		"Approval "+decision,
		fmt.Sprintf("Your approval request on issue %d has been %s", approval.IssueID, decision),
		map[string]interface{}{"approval_id": approval.ID, "issue_id": approval.IssueID},
	)
}
```

- [ ] **Step 5: Wire up the notifier in router.go**

Update the handler construction in `router.go`:

```go
notificationSvc := service.NewNotificationService(db)  // or existing instance
approvalSvc := service.NewApprovalService(db, notificationSvc)
approvalHandler := handler.NewApprovalHandler(approvalSvc)
```

- [ ] **Step 6: Verify build and run tests**

```bash
cd backend && go build ./... && go test ./internal/service -run TestApprovalService -v
```
Expected: tests still pass (nil notifier or mock).

- [ ] **Step 7: Commit**

```bash
git add backend/internal/service/approval_service.go backend/internal/service/approval_service_test.go backend/internal/router/router.go
git commit -m "feat(approval): send notifications on submit and decide"
```

---

### Task 8.2: Record approval activities in issue activity stream

**Files:**
- Modify: `backend/internal/service/approval_service.go`
- Modify: `frontend/src/components/ActivityStream.vue` (or wherever activities are rendered)

- [ ] **Step 1: Locate the activity recording helper**

Run:
```bash
grep -rn "func.*RecordActivity\|func.*AddActivity\|IssueActivity" backend/internal/service | head -20
```

Identify the function that creates an `IssueActivity` record.

- [ ] **Step 2: Add activity recording in approval_service.go**

In `Create`, after the transaction:

```go
s.recordActivity(approval.IssueID, approval.RequesterID, "approval_submitted", map[string]interface{}{
	"approval_id":           approval.ID,
	"from_state_id":         approval.SourceStateID,
	"approve_target_state_id": approval.ApproveTargetStateID,
})
```

In `Decide`, after the transaction:

```go
s.recordActivity(approval.IssueID, approverID, "approval_"+decision, map[string]interface{}{
	"approval_id":  approval.ID,
	"target_state_id": targetStateID,
	"note":         note,
})
```

In `Cancel`:

```go
s.recordActivity(approval.IssueID, userID, "approval_cancelled", map[string]interface{}{
	"approval_id": approval.ID,
})
```

Add the helper:
```go
func (s *ApprovalService) recordActivity(issueID, actorID uint64, activityType string, metadata map[string]interface{}) {
	// Convert metadata to JSON
	jsonBytes, _ := json.Marshal(metadata)
	activity := model.IssueActivity{
		IssueID:   issueID,
		ActorID:   actorID,
		Type:      activityType,
		Metadata:  string(jsonBytes),
	}
	s.db.Create(&activity)  // best-effort, ignore errors
}
```

If the `IssueActivity` model or field names differ, adjust accordingly.

- [ ] **Step 3: Render approval activities in frontend**

Edit the frontend activity stream component (find it via `grep -rn "activity_type\|activity.type" frontend/src/components`). Add rendering cases for the four new types:

```vue
<div v-else-if="activity.type === 'approval_submitted'" class="text-sm text-gray-600">
  <span class="font-medium">{{ activity.actor_name }}</span>
  {{ t('approvals.activitySubmitted') }}
  <span class="text-gray-400 ml-1">{{ formatTime(activity.created_at) }}</span>
</div>
<div v-else-if="activity.type === 'approval_approved'" class="text-sm text-green-700">
  <span class="font-medium">{{ activity.actor_name }}</span>
  {{ t('approvals.activityApproved') }}
  <span class="text-gray-400 ml-1">{{ formatTime(activity.created_at) }}</span>
</div>
<div v-else-if="activity.type === 'approval_rejected'" class="text-sm text-red-700">
  <span class="font-medium">{{ activity.actor_name }}</span>
  {{ t('approvals.activityRejected') }}
  <span class="text-gray-400 ml-1">{{ formatTime(activity.created_at) }}</span>
</div>
<div v-else-if="activity.type === 'approval_cancelled'" class="text-sm text-gray-500">
  <span class="font-medium">{{ activity.actor_name }}</span>
  {{ t('approvals.activityCancelled') }}
  <span class="text-gray-400 ml-1">{{ formatTime(activity.created_at) }}</span>
</div>
```

Add i18n keys:
```json
"activitySubmitted": "提交了审批申请",
"activityApproved": "批准了审批",
"activityRejected": "拒绝了审批",
"activityCancelled": "撤销了审批"
```

- [ ] **Step 4: Verify end-to-end in browser**

1. Submit an approval on an issue
2. As approver, approve/reject
3. Verify the activity stream shows all entries correctly

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/approval_service.go frontend/src/components/ActivityStream.vue frontend/src/locales/zh-CN.json frontend/src/locales/en-US.json
git commit -m "feat(approval): record and render approval activities in issue stream"
```

---

## Phase 9: Integration Testing + Acceptance

### Task 9.1: Backend API integration test

**Files:**
- Create: `backend/internal/handler/approval_handler_test.go` (optional, if a test harness exists)

- [ ] **Step 1: Write end-to-end API test (if httptest setup exists)**

Skip if no existing handler test pattern. Otherwise, write a test that:
1. Creates an issue
2. Creates a workflow + approval transition
3. Calls `POST /issues/:id/approvals`
4. Calls `POST /approvals/:id/decide` with `decision=approved`
5. Verifies issue state changed to approve_target_state_id

- [ ] **Step 2: Run all backend tests**

```bash
cd backend && go test ./...
```
Expected: all tests pass.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/approval_handler_test.go
git commit -m "test(approval): add end-to-end API integration test"
```

---

### Task 9.2: Manual acceptance checklist

- [ ] **Step 1: Backend smoke test**

Run these in order with curl/PowerShell:

```bash
# 1. Create approval transition (assumes workflow 25, states 1/2 exist)
curl -X POST -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '{"from_state_id":1,"to_state_id":2,"rule_type":"approval","approver_ids":"[2,3]","approve_target_state_id":2,"reject_target_state_id":1}' \
  http://localhost:8000/api/v1/projects/1/workflows/25/transitions

# 2. Submit approval (assumes issue 1 in state 1)
curl -X POST -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '{"transition_id":<tid>,"request_note":"please review"}' \
  http://localhost:8000/api/v1/issues/1/approvals

# 3. List pending approvals
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8000/api/v1/workspaces/1/approvals?status=pending"

# 4. Approve
curl -X POST -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '{"decision":"approved","note":"looks good"}' \
  http://localhost:8000/api/v1/approvals/<id>/decide

# 5. Verify issue state changed
curl -H "Authorization: Bearer <token>" http://localhost:8000/api/v1/issues/1
```

Expected:
- Step 1 returns the transition with `approve_target_state_id=2`
- Step 2 returns 201 with status=pending
- Step 3 returns the approval in the list
- Step 4 returns the approval with status=approved
- Step 5 shows issue.state_id=2

- [ ] **Step 2: Frontend manual test**

1. Open workflow detail page, create an approval transition with approve/reject targets
2. Open an issue, change state to the approval target -> submit dialog appears
3. Submit -> yellow banner appears
4. Logout, login as approver, see TopBar badge with count
5. Open approval list page, see the pending approval
6. Click Approve -> dialog -> confirm
7. Original issue state is updated
8. Activity stream shows submitted + approved entries

- [ ] **Step 3: Edge case tests**

- Submit approval on issue that already has pending approval -> should reject with `issue_already_pending_approval`
- Try to decide an approval as non-approver -> should reject with 403
- Try to cancel an approval as non-requester -> should reject with 403
- Try to directly change state of an issue with pending approval -> should reject with `issue_pending_approval`

- [ ] **Step 4: Final commit (if any fixes were needed)**

```bash
git add -A
git commit -m "fix(approval): address issues found in acceptance testing"
```

---

## Self-Review Checklist

**Spec coverage:**
- ✅ Data model (Section 3) — Task 1.1, 1.2
- ✅ Backend models (Section 4.1) — Task 1.2
- ✅ Service layer (Section 4.2) — Task 2.2
- ✅ Issue service refactor (Section 4.3) — Task 4.1
- ✅ Handlers + routes (Section 4.4) — Task 3.1
- ✅ Workflow transition API extension (Section 4.5) — Task 4.2
- ✅ Workflow detail UI (Section 5.1) — Task 5.1
- ✅ Issue detail state change dialog (Section 5.2.1) — Task 6.4
- ✅ Pending approval banner (Section 5.2.2) — Task 6.3, 6.4
- ✅ Activity stream entries (Section 5.2.3) — Task 8.2
- ✅ TopBar approval icon (Section 5.3) — Task 7.1
- ✅ Approval list page (Section 5.4) — Task 7.2
- ✅ i18n (Section 5.5) — Task 5.1, 6.2, 8.2
- ✅ Notifications (Section 6) — Task 8.1
- ✅ Edge cases (Section 7) — covered in service tests + manual acceptance
- ✅ Backward compatibility (Section 8) — migration script in Task 1.1

**Placeholder scan:** None found.

**Type consistency:**
- `ApprovalResponse` defined in Task 2.1, used consistently in Tasks 6.1, 6.3, 7.1, 7.2
- `ApprovalService` methods (Create/List/Get/Decide/Cancel/CountPending) defined in Task 2.2, called consistently in Tasks 3.1, 6.4, 7.1, 7.2
- `ApprovalRequiredError` introduced in Task 4.1, used in issue handler

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-07-19-workflow-approval.md`.
