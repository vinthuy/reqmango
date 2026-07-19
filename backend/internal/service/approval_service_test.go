package service

import (
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

	// Mock: count existing pending approvals = 0
	mock.ExpectQuery(`SELECT count`).
		WithArgs(uint64(1), "pending").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock: transition query
	approverIDs := `[2,3]`
	mock.ExpectQuery(`SELECT .* FROM "state_transitions"`).
		WithArgs(uint64(100), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "workflow_id", "source_state_id", "target_state_id", "rule_type", "approver_ids", "approve_target_state_id", "reject_target_state_id", "approval_mode"}).
			AddRow(100, 50, 10, 20, "approval", approverIDs, 20, 10, "any"))

	// Mock: issue query
	mock.ExpectQuery(`SELECT .* FROM "issues"`).
		WithArgs(uint64(1), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "state_id", "project_id", "workspace_id", "approval_status"}).
			AddRow(1, 10, 1, 1, nil))

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
		WithArgs(uint64(1), 1).
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

	// Mock: reload approval
	mock.ExpectQuery(`SELECT .* FROM "approvals"`).
		WithArgs(uint64(1), uint64(1), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "approver_ids", "approve_target_state_id", "reject_target_state_id", "issue_id", "requester_id"}).
			AddRow(1, "approved", approverIDs, 20, 10, 100, 5))

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
		WithArgs(uint64(1), 1).
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
		WithArgs(uint64(1), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "requester_id", "issue_id"}).
			AddRow(1, "pending", 5, 100))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "approvals"`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "issues"`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Mock: reload approval
	mock.ExpectQuery(`SELECT .* FROM "approvals"`).
		WithArgs(uint64(1), uint64(1), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "requester_id", "issue_id"}).
			AddRow(1, "cancelled", 5, 100))

	svc := NewApprovalService(db)
	approval, err := svc.Cancel(1, 5)
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", approval.Status)
}

func TestApprovalService_Cancel_NotRequester(t *testing.T) {
	db, mock := newApprovalTestDB(t)
	defer db.DB()

	mock.ExpectQuery(`SELECT .* FROM "approvals"`).
		WithArgs(uint64(1), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "requester_id", "issue_id"}).
			AddRow(1, "pending", 5, 100))

	svc := NewApprovalService(db)
	_, err := svc.Cancel(1, 99)
	assert.Error(t, err)
}
