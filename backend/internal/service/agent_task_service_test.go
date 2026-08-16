package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupAgentTaskTestDB(t *testing.T) *gorm.DB {
	db := openTestDB(t)
	err := db.AutoMigrate(
		&model.User{},
		&model.Workspace{},
		&model.AgentTask{},
		&model.AgentTemplate{},
		&model.AgentConfig{},
		&model.Runtime{},
		&model.TaskLog{},
		&model.Tool{},
		&model.ToolPermission{},
		&model.ToolCallLog{},
		&model.MCPConfig{},
		&model.WorkspaceMember{},
		&model.Skill{},
		&model.SkillExecutionLog{},
	)
	require.NoError(t, err)
	return db
}

// createAdminForWorkspace creates a user and workspace member with admin role for the given workspace.
func createAdminForWorkspace(t *testing.T, db *gorm.DB, wsID uint64) *model.User {
	user := makeUser(t, db, "task_admin")
	member := &model.WorkspaceMember{
		WorkspaceID: wsID,
		UserID:      user.ID,
		Role:        common.RoleAdmin,
		IsActive:    true,
	}
	require.NoError(t, db.Create(member).Error)
	return user
}

// createMemberForWorkspace creates a user and workspace member with member (non-admin) role.
func createMemberForWorkspace(t *testing.T, db *gorm.DB, wsID uint64) *model.User {
	user := makeUser(t, db, "task_member")
	member := &model.WorkspaceMember{
		WorkspaceID: wsID,
		UserID:      user.ID,
		Role:        common.RoleMember,
		IsActive:    true,
	}
	require.NoError(t, db.Create(member).Error)
	return user
}

// createEnqueueTask creates a task in "enqueue" status for testing claim/complete/fail/cancel.
func createEnqueueTask(t *testing.T, db *gorm.DB, wsID uint64, title string) *model.AgentTask {
	task := &model.AgentTask{
		Title:       title,
		Status:      "enqueue",
		Priority:    "normal",
		Progress:    0,
		WorkspaceID: wsID,
	}
	require.NoError(t, db.Create(task).Error)
	return task
}

// ==================== Create ====================

func TestAgentTaskService_Create(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task1", Slug: "ws_task1", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createAdminForWorkspace(t, db, ws.ID)

	req := request.AgentTaskCreate{
		Title:     "Implement Feature X",
		Priority:  "high",
		TaskType:  "generate_code",
		InputData: json.RawMessage(`{"repo": "frontend"}`),
	}

	resp, err := svc.Create(ws.ID, admin.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Implement Feature X", resp.Title)
	assert.Equal(t, "enqueue", resp.Status)
	assert.Equal(t, "high", resp.Priority)
	assert.Equal(t, 0, resp.Progress)
	assert.Equal(t, ws.ID, resp.WorkspaceID)
}

func TestAgentTaskService_Create_NonAdmin(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_na", Slug: "ws_task_na", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	member := createMemberForWorkspace(t, db, ws.ID)

	req := request.AgentTaskCreate{
		Title: "Should Fail",
	}

	_, err := svc.Create(ws.ID, member.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestAgentTaskService_Create_NoMember(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_nm", Slug: "ws_task_nm", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	outsider := makeUser(t, db, "outsider_task")

	req := request.AgentTaskCreate{
		Title: "Outsider Task",
	}

	_, err := svc.Create(ws.ID, outsider.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestAgentTaskService_Create_WithOptionalFields(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_opt", Slug: "ws_task_opt", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createAdminForWorkspace(t, db, ws.ID)

	estTime := 30
	req := request.AgentTaskCreate{
		Title:         "Full Task",
		Description:   strPtr("A full task with all fields"),
		Priority:      "urgent",
		TaskType:      "analyze_requirement",
		InputData:     json.RawMessage(`{"source": "jira"}`),
		EstimatedTime: &estTime,
	}

	resp, err := svc.Create(ws.ID, admin.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Full Task", resp.Title)
	assert.Equal(t, "urgent", resp.Priority)
	assert.Equal(t, "analyze_requirement", resp.TaskType)
	assert.NotNil(t, resp.EstimatedTime)
	assert.Equal(t, 30, *resp.EstimatedTime)
}

// ==================== Get ====================

func TestAgentTaskService_Get(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_get", Slug: "ws_task_get", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Get Me")

	resp, err := svc.Get(task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, resp.ID)
	assert.Equal(t, "Get Me", resp.Title)
	assert.Equal(t, ws.ID, resp.WorkspaceID)
}

func TestAgentTaskService_Get_NotFound(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	_, err := svc.Get(99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== List ====================

func TestAgentTaskService_List(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_list", Slug: "ws_task_list", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	createEnqueueTask(t, db, ws.ID, "Task A")
	createEnqueueTask(t, db, ws.ID, "Task B")
	createEnqueueTask(t, db, ws.ID, "Task C")

	resp, err := svc.List(ws.ID, "")
	require.NoError(t, err)
	assert.Len(t, resp, 3)
}

func TestAgentTaskService_List_FilterByStatus(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_lf", Slug: "ws_task_lf", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	createEnqueueTask(t, db, ws.ID, "Enqueue 1")
	createEnqueueTask(t, db, ws.ID, "Enqueue 2")

	// Create a completed task
	runningTask := createEnqueueTask(t, db, ws.ID, "Running 1")
	runningTask.Status = "running"
	require.NoError(t, db.Save(runningTask).Error)

	enqueueResp, err := svc.List(ws.ID, "enqueue")
	require.NoError(t, err)
	assert.Len(t, enqueueResp, 2)

	runningResp, err := svc.List(ws.ID, "running")
	require.NoError(t, err)
	assert.Len(t, runningResp, 1)
}

func TestAgentTaskService_List_Empty(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	resp, err := svc.List(99999, "")
	require.NoError(t, err)
	assert.Len(t, resp, 0)
}

func TestAgentTaskService_List_OtherWorkspace(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws1 := &model.Workspace{Name: "ws_task_lo1", Slug: "ws_task_lo1", OwnerID: 1}
	ws2 := &model.Workspace{Name: "ws_task_lo2", Slug: "ws_task_lo2", OwnerID: 1}
	require.NoError(t, db.Create(ws1).Error)
	require.NoError(t, db.Create(ws2).Error)

	createEnqueueTask(t, db, ws1.ID, "WS1 Task")
	createEnqueueTask(t, db, ws1.ID, "WS1 Task 2")
	createEnqueueTask(t, db, ws2.ID, "WS2 Task")

	resp1, err := svc.List(ws1.ID, "")
	require.NoError(t, err)
	assert.Len(t, resp1, 2)

	resp2, err := svc.List(ws2.ID, "")
	require.NoError(t, err)
	assert.Len(t, resp2, 1)
}

// ==================== Claim ====================

func TestAgentTaskService_Claim(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_claim", Slug: "ws_task_claim", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Claim Me")

	resp, err := svc.Claim(task.ID, 0)
	require.NoError(t, err)
	assert.Equal(t, "claimed", resp.Status)
	assert.NotNil(t, resp.ClaimedAt)
}

func TestAgentTaskService_Claim_WithRuntime(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_cr", Slug: "ws_task_cr", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	rt := &model.Runtime{
		Name:        "Claim Runtime",
		RuntimeType: "local_daemon",
		Status:      "online",
		Capacity:    4,
		CurrentLoad: 0,
		WorkspaceID: ws.ID,
	}
	require.NoError(t, db.Create(rt).Error)

	task := createEnqueueTask(t, db, ws.ID, "Claim With Runtime")

	resp, err := svc.Claim(task.ID, rt.ID)
	require.NoError(t, err)
	assert.Equal(t, "claimed", resp.Status)
	assert.NotNil(t, resp.RuntimeID)
	assert.Equal(t, rt.ID, *resp.RuntimeID)

	// Verify runtime load increased
	var updated model.Runtime
	require.NoError(t, db.First(&updated, rt.ID).Error)
	assert.Equal(t, 1, updated.CurrentLoad)
}

func TestAgentTaskService_Claim_NotEnqueue(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_cne", Slug: "ws_task_cne", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Already Claimed")
	task.Status = "claimed"
	require.NoError(t, db.Save(task).Error)

	_, err := svc.Claim(task.ID, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not available for claiming")
}

func TestAgentTaskService_Claim_NotFound(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	_, err := svc.Claim(99999, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Complete ====================

func TestAgentTaskService_Complete(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_cmp", Slug: "ws_task_cmp", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Complete Me")
	// Transition to running
	task.Status = "running"
	require.NoError(t, db.Save(task).Error)

	outputData := json.RawMessage(`{"result": "success"}`)
	req := request.AgentTaskComplete{
		OutputData: outputData,
		ActualTime: 120,
	}

	resp, err := svc.Complete(task.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, 100, resp.Progress)
	assert.NotNil(t, resp.CompletedAt)
	assert.NotNil(t, resp.ActualTime)
	assert.Equal(t, 120, *resp.ActualTime)
}

func TestAgentTaskService_Complete_NotRunning(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_cnr", Slug: "ws_task_cnr", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Not Running")

	req := request.AgentTaskComplete{}
	_, err := svc.Complete(task.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be running to complete")
}

func TestAgentTaskService_Complete_NotFound(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	req := request.AgentTaskComplete{}
	_, err := svc.Complete(99999, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Fail ====================

func TestAgentTaskService_Fail(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_fail", Slug: "ws_task_fail", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Fail Me")
	// Transition to running
	task.Status = "running"
	require.NoError(t, db.Save(task).Error)

	req := request.AgentTaskFail{
		ErrorInfo:     "Agent execution error occurred",
		FailureReason: "agent_error",
	}

	resp, err := svc.Fail(task.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "failed", resp.Status)
	assert.NotNil(t, resp.ErrorInfo)
	assert.Equal(t, "Agent execution error occurred", *resp.ErrorInfo)
	assert.NotNil(t, resp.CompletedAt)
	assert.Equal(t, "agent_error", resp.FailureReason)
}

func TestAgentTaskService_Fail_AutoDetectReason(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_fdr", Slug: "ws_task_fdr", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	testCases := []struct {
		errorInfo      string
		expectedReason string
	}{
		{"task timeout exceeded", "timeout"},
		{"runtime is offline", "runtime_offline"},
		{"invalid input parameters", "invalid_input"},
		{"API rate limit exceeded", "rate_limit"},
		{"model API error", "model_error"},
		{"some random error", "agent_error"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedReason, func(t *testing.T) {
			task := createEnqueueTask(t, db, ws.ID, fmt.Sprintf("Fail_%s", tc.expectedReason))
			task.Status = "running"
			require.NoError(t, db.Save(task).Error)

			req := request.AgentTaskFail{
				ErrorInfo: tc.errorInfo,
			}

			resp, err := svc.Fail(task.ID, req)
			require.NoError(t, err)
			assert.Equal(t, "failed", resp.Status)
			assert.Equal(t, tc.expectedReason, resp.FailureReason)
		})
	}
}

func TestAgentTaskService_Fail_NotFound(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	req := request.AgentTaskFail{ErrorInfo: "error"}
	_, err := svc.Fail(99999, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Cancel ====================

func TestAgentTaskService_Cancel(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_cancel", Slug: "ws_task_cancel", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Cancel Me")

	resp, err := svc.Cancel(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", resp.Status)
	assert.NotNil(t, resp.CancelledAt)
}

func TestAgentTaskService_Cancel_ClaimedTask(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_cc", Slug: "ws_task_cc", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Cancel Claimed")
	// Transition to claimed
	task.Status = "claimed"
	require.NoError(t, db.Save(task).Error)

	resp, err := svc.Cancel(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", resp.Status)
}

func TestAgentTaskService_Cancel_CompletedTask(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_ccomp", Slug: "ws_task_ccomp", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Cancel Completed")
	task.Status = "completed"
	require.NoError(t, db.Save(task).Error)

	_, err := svc.Cancel(task.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot cancel completed or failed task")
}

func TestAgentTaskService_Cancel_FailedTask(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_cfail", Slug: "ws_task_cfail", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Cancel Failed")
	task.Status = "failed"
	require.NoError(t, db.Save(task).Error)

	_, err := svc.Cancel(task.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot cancel completed or failed task")
}

func TestAgentTaskService_Cancel_NotFound(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	_, err := svc.Cancel(99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAgentTaskService_Cancel_WithRuntime(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_crt", Slug: "ws_task_crt", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	rt := &model.Runtime{
		Name:        "Cancel Runtime",
		RuntimeType: "local_daemon",
		Status:      "online",
		Capacity:    4,
		CurrentLoad: 2,
		WorkspaceID: ws.ID,
	}
	require.NoError(t, db.Create(rt).Error)

	task := createEnqueueTask(t, db, ws.ID, "Cancel With Runtime")
	task.Status = "running"
	task.RuntimeID = &rt.ID
	require.NoError(t, db.Save(task).Error)

	_, err := svc.Cancel(task.ID)
	require.NoError(t, err)

	// Verify runtime load decreased
	var updated model.Runtime
	require.NoError(t, db.First(&updated, rt.ID).Error)
	assert.Equal(t, 1, updated.CurrentLoad)
}

// ==================== AddLog ====================

func TestAgentTaskService_AddLog(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_log", Slug: "ws_task_log", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Log Task")

	err := svc.AddLog(task.ID, "info", "Task started", nil)
	require.NoError(t, err)

	logs, err := svc.GetLogs(task.ID)
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "info", logs[0].Level)
	assert.Equal(t, "Task started", logs[0].Message)
}

func TestAgentTaskService_AddLog_WithMetadata(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_logm", Slug: "ws_task_logm", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	task := createEnqueueTask(t, db, ws.ID, "Log Meta Task")

	meta := []byte(`{"step": 1}`)
	err := svc.AddLog(task.ID, "debug", "Step 1 complete", meta)
	require.NoError(t, err)

	logs, err := svc.GetLogs(task.ID)
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "debug", logs[0].Level)
	assert.Equal(t, meta, []byte(logs[0].Metadata))
}

// ==================== Delete ====================

func TestAgentTaskService_Delete(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_del", Slug: "ws_task_del", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createAdminForWorkspace(t, db, ws.ID)

	task := createEnqueueTask(t, db, ws.ID, "Delete Me")

	err := svc.Delete(task.ID, admin.ID)
	require.NoError(t, err)

	_, err = svc.Get(task.ID)
	assert.Error(t, err)
}

func TestAgentTaskService_Delete_NonAdmin(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_dna", Slug: "ws_task_dna", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	member := createMemberForWorkspace(t, db, ws.ID)

	task := createEnqueueTask(t, db, ws.ID, "Delete NonAdmin")

	err := svc.Delete(task.ID, member.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestAgentTaskService_Delete_NotFound(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	err := svc.Delete(99999, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Update ====================

func TestAgentTaskService_Update(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_upd", Slug: "ws_task_upd", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createAdminForWorkspace(t, db, ws.ID)

	task := createEnqueueTask(t, db, ws.ID, "Update Me")

	newTitle := "Updated Task Title"
	newPriority := "urgent"
	progress := 50
	req := request.AgentTaskUpdate{
		Title:    &newTitle,
		Priority: &newPriority,
		Progress: &progress,
	}

	resp, err := svc.Update(task.ID, admin.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Updated Task Title", resp.Title)
	assert.Equal(t, "urgent", resp.Priority)
	assert.Equal(t, 50, resp.Progress)
}

func TestAgentTaskService_Update_NotFound(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	newTitle := "Ghost"
	req := request.AgentTaskUpdate{Title: &newTitle}

	_, err := svc.Update(99999, 1, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAgentTaskService_Update_NonAdmin(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_upna", Slug: "ws_task_upna", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	member := createMemberForWorkspace(t, db, ws.ID)

	task := createEnqueueTask(t, db, ws.ID, "Update NonAdmin")

	newTitle := "Should Not Update"
	req := request.AgentTaskUpdate{Title: &newTitle}

	_, err := svc.Update(task.ID, member.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

// ==================== Full Lifecycle ====================

func TestAgentTaskService_FullLifecycle(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_lc", Slug: "ws_task_lc", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createAdminForWorkspace(t, db, ws.ID)

	// 1. Create
	createReq := request.AgentTaskCreate{
		Title:    "Lifecycle Task",
		Priority: "normal",
	}
	task, err := svc.Create(ws.ID, admin.ID, createReq)
	require.NoError(t, err)
	assert.Equal(t, "enqueue", task.Status)

	// 2. Claim
	claimed, err := svc.Claim(task.ID, 0)
	require.NoError(t, err)
	assert.Equal(t, "claimed", claimed.Status)

	// 3. Transition to running (simulates runtime picking up the task)
	require.NoError(t, db.Model(&model.AgentTask{}).Where("id = ?", task.ID).Update("status", "running").Error)

	// 4. Complete
	completeReq := request.AgentTaskComplete{
		OutputData: json.RawMessage(`{"result": "done"}`),
		ActualTime: 60,
	}
	completed, err := svc.Complete(task.ID, completeReq)
	require.NoError(t, err)
	assert.Equal(t, "completed", completed.Status)
	assert.Equal(t, 100, completed.Progress)
}

func TestAgentTaskService_FullLifecycle_FailAndRetry(t *testing.T) {
	db := setupAgentTaskTestDB(t)
	svc := NewAgentTaskService(db)

	ws := &model.Workspace{Name: "ws_task_fr", Slug: "ws_task_fr", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createAdminForWorkspace(t, db, ws.ID)

	// Create and start
	createReq := request.AgentTaskCreate{
		Title: "Fail Retry Task",
	}
	task, err := svc.Create(ws.ID, admin.ID, createReq)
	require.NoError(t, err)

	claimed, err := svc.Claim(task.ID, 0)
	require.NoError(t, err)
	assert.Equal(t, "claimed", claimed.Status)

	// Fail
	failReq := request.AgentTaskFail{ErrorInfo: "Something went wrong"}
	failed, err := svc.Fail(task.ID, failReq)
	require.NoError(t, err)
	assert.Equal(t, "failed", failed.Status)
	assert.Equal(t, "agent_error", failed.FailureReason)
}
