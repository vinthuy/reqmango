package service

import (
	"testing"
)

// TestWorkflowExecute_PanicRecovery verifies S4: a panic during Execute marks
// the workflow run as "failed" instead of leaving it stuck in "running".
//
// A minimal workflow_runs row is inserted pointing at the existing workflow 1.
// The executor is constructed with a nil contextSvc so Execute panics when it
// first touches the context service (well after the deferred recover is
// registered). The deferred recover then calls failRun -> status "failed".
func TestWorkflowExecute_PanicRecovery(t *testing.T) {
	db := testDB(t)

	var runID uint64
	row := db.Raw(`INSERT INTO workflow_runs (workflow_id, status, issue_id) VALUES (1, 'pending', 1) RETURNING id`).Row()
	if err := row.Scan(&runID); err != nil {
		t.Fatalf("insert workflow_runs row: %v", err)
	}
	defer db.Exec("DELETE FROM workflow_runs WHERE id = ?", runID)

	exec := &WorkflowExecutor{
		db:          db,
		workflowSvc: NewWorkflowService(db, nil, nil, nil),
		contextSvc:  nil, // triggers a nil-pointer panic at BuildInitialContext
		// Safety net so the test can never hang on a nil channel if the panic
		// were somehow not reached.
		semaphore:   make(chan struct{}, 4),
		maxParallel: 4,
	}

	// Execute recovers from the panic internally and marks the run failed.
	_ = exec.Execute(runID, 1)

	var status string
	if err := db.Raw("SELECT status FROM workflow_runs WHERE id = ?", runID).Scan(&status).Error; err != nil {
		t.Fatalf("query run status: %v", err)
	}
	if status != "failed" {
		t.Errorf("expected run status 'failed', got %q", status)
	}

	var errInfo string
	db.Raw("SELECT error_info FROM workflow_runs WHERE id = ?", runID).Scan(&errInfo)
	if errInfo == "" {
		t.Errorf("expected non-empty error_info describing the panic")
	}
}
