package service

import (
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
)

// panickingAgent is an AgentExecutorInterface whose DispatchAgent always panics.
type panickingAgent struct{}

func (p *panickingAgent) DispatchAgent(agentID, userID uint64, task string, ctx *AgentDispatchContext) (*AgentDispatchResult, error) {
	panic("intentional test panic")
}

// testDB connects to the local PostgreSQL instance used for development. The
// test is skipped (not failed) when the DB is unavailable.
func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open("postgres://postgres:postgres@localhost:5432/reqmango?sslmode=disable"), &gorm.Config{})
	if err != nil {
		t.Skipf("cannot connect to test DB: %v", err)
	}
	return db
}

// TestSquadStartExecution_PanicRecovery verifies S2: a panic during squad
// execution marks the execution row as "failed" instead of leaving it stuck
// in "running".
func TestSquadStartExecution_PanicRecovery(t *testing.T) {
	db := testDB(t)

	leaderID := uint64(1)
	squad := &model.Squad{
		WorkspaceID:   1,
		Name:          "test-panic-squad",
		LeaderAgentID: &leaderID,
		Status:        "active",
	}
	if err := db.Create(squad).Error; err != nil {
		t.Fatalf("create squad: %v", err)
	}
	member := &model.SquadMember{
		SquadID:    squad.ID,
		AgentID:    1,
		Role:       "leader",
		Status:     "active",
		AssignedAt: time.Now(),
	}
	if err := db.Create(member).Error; err != nil {
		t.Fatalf("create member: %v", err)
	}
	defer func() {
		db.Where("squad_id = ?", squad.ID).Delete(&model.SquadExecution{})
		db.Where("squad_id = ?", squad.ID).Delete(&model.SquadTask{})
		db.Where("squad_id = ?", squad.ID).Delete(&model.SquadMember{})
		db.Delete(squad)
	}()

	// Ensure new columns exist (cancelled_at, cancel_reason)
	db.AutoMigrate(&model.SquadExecution{})

	svc := NewSquadService(db)
	svc.SetAgentExecutor(&panickingAgent{})

	// StartExecution is now async — it returns immediately with status=pending.
	// The goroutine will panic and recover, marking execution as "failed".
	resp, err := svc.StartExecution(squad.ID, request.SquadExecutionStart{
		Goal:   "test goal",
		UserID: 1,
	})
	if err != nil {
		t.Fatalf("start execution: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Status != "pending" {
		t.Errorf("expected initial status 'pending', got %q", resp.Status)
	}

	// Wait for the async goroutine to finish (panic recovery + DB update).
	// Poll with a short timeout.
	var exec model.SquadExecution
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if err := db.Where("squad_id = ?", squad.ID).Order("id desc").First(&exec).Error; err == nil {
			if exec.Status == "failed" {
				break
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	if exec.Status != "failed" {
		t.Errorf("expected exec status 'failed', got %q (error_info: %s)", exec.Status, exec.ErrorInfo)
	}
	if exec.ErrorInfo == "" {
		t.Errorf("expected non-empty error_info describing the panic")
	}
}
