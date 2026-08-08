package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/reqmango/backend/internal/model"
)

func TestCancelExecution_NotRunning(t *testing.T) {
	svc := NewSquadService(nil)
	err := svc.CancelExecution(999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCancelExecution_RemovesFromStore(t *testing.T) {
	svc := NewSquadService(nil)
	// Register a fake cancel func
	called := false
	svc.cancelStore.Store(uint64(1), context.CancelFunc(func() { called = true }))
	// CancelExecution will: call cancel func, delete from store, then try DB (panic).
	// Use recover to handle the expected panic from nil DB access.
	func() {
		defer func() { recover() }() // swallow expected panic from nil db
		_ = svc.CancelExecution(1)
	}()
	assert.True(t, called, "cancel func should have been called")
	// Verify removed from store (delete happens before DB access)
	_, ok := svc.cancelStore.Load(uint64(1))
	assert.False(t, ok, "cancel func should be removed from store")
}

func TestCheckPermissions_NoDB(t *testing.T) {
	svc := NewSquadService(nil)
	// db=nil => test mode: skip check
	err := svc.checkPermissions(1, 999)
	assert.NoError(t, err)
}

func TestTruncateStr(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
	}
	for _, tt := range tests {
		got := truncateStr(tt.input, tt.n)
		assert.Equal(t, tt.want, got, "truncateStr(%q, %d)", tt.input, tt.n)
	}
}

func TestNewSquadService_NilDB(t *testing.T) {
	svc := NewSquadService(nil)
	assert.NotNil(t, svc)
	assert.Nil(t, svc.db)
}

func TestExecuteSubtaskWithRetry_ContextCancelled(t *testing.T) {
	svc := NewSquadService(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	squad := &model.Squad{}
	member := &model.SquadMember{}
	logs := []string{}
	result := svc.executeSubtaskWithRetry(ctx, 1, squad, member, "task", "title", 1, &logs)
	assert.Empty(t, result)
	assert.Len(t, logs, 1)
	assert.Contains(t, logs[0], "cancelled")
}

func TestBuildResponse_WithMembers(t *testing.T) {
	svc := NewSquadService(nil)
	now := time.Now()
	squad := &model.Squad{
		WorkspaceID: 1,
		Name:        "Test Squad",
		Status:      "active",
		Members: []model.SquadMember{
			{
				SquadID:    1,
				AgentID:    10,
				Role:       "contributor",
				Status:     "active",
				AssignedAt: now,
			},
		},
	}
	squad.CreatedAt = now
	squad.UpdatedAt = now
	resp := svc.buildResponse(squad)
	assert.Equal(t, squad.ID, resp.ID)
	assert.Equal(t, squad.Name, resp.Name)
	assert.Len(t, resp.Members, 1)
	assert.Equal(t, uint64(10), resp.Members[0].AgentID)
}

func TestBuildExecutionResponse_WithData(t *testing.T) {
	svc := NewSquadService(nil)
	now := time.Now()
	exec := &model.SquadExecution{
		SquadID:      1,
		Status:       "completed",
		Goal:         "test goal",
		StartedAt:    &now,
		CompletedAt:  &now,
		CancelReason: "",
	}
	exec.CreatedAt = now
	resp := svc.buildExecutionResponse(exec)
	assert.Equal(t, exec.Status, resp.Status)
	assert.Equal(t, exec.Goal, resp.Goal)
	assert.Equal(t, exec.SquadID, resp.SquadID)
}

func TestBuildMemberResponse_WithData(t *testing.T) {
	svc := NewSquadService(nil)
	now := time.Now()
	member := &model.SquadMember{
		SquadID:       1,
		AgentID:       20,
		Role:          "reviewer",
		AgentConfigID: 5,
		Status:        "active",
		AssignedAt:    now,
	}
	resp := svc.buildMemberResponse(member)
	assert.Equal(t, member.AgentID, resp.AgentID)
	assert.Equal(t, member.Role, resp.Role)
	assert.Equal(t, member.AgentConfigID, resp.AgentConfigID)
}
