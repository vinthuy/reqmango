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

	svc := NewSquadService(db)
	svc.SetAgentExecutor(&panickingAgent{})

	// decomposeGoal calls the panicking mock -> panic -> deferred recover marks
	// the exec row failed. StartExecution returns (nil, nil) after recovery.
	_, _ = svc.StartExecution(squad.ID, request.SquadExecutionStart{
		Goal:   "test goal",
		UserID: 1,
	})

	var exec model.SquadExecution
	if err := db.Where("squad_id = ?", squad.ID).Order("id desc").First(&exec).Error; err != nil {
		t.Fatalf("query exec: %v", err)
	}
	if exec.Status != "failed" {
		t.Errorf("expected exec status 'failed', got %q (error_info: %s)", exec.Status, exec.ErrorInfo)
	}
	if exec.ErrorInfo == "" {
		t.Errorf("expected non-empty error_info describing the panic")
	}
}
