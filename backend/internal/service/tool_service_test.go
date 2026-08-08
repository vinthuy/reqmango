package service

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var testCounter uint64

func nextID() uint64 {
	return atomic.AddUint64(&testCounter, 1)
}

func setupToolTestDB(t *testing.T) *gorm.DB {
	db := setupTestDB(t)
	err := db.AutoMigrate(
		&model.Tool{},
		&model.ToolPermission{},
		&model.ToolCallLog{},
		&model.WorkspaceMember{},
		&model.MCPConfig{},
	)
	require.NoError(t, err)
	return db
}

func makeUser(t *testing.T, db *gorm.DB, name string) *model.User {
	id := nextID()
	u := &model.User{
		Username:    fmt.Sprintf("user_%d_%s", id, name),
		Email:       fmt.Sprintf("user_%d_%s@test.com", id, name),
		DisplayName: name,
	}
	require.NoError(t, db.Create(u).Error)
	return u
}

func makeTool(t *testing.T, db *gorm.DB, wid uint64, name, category string, rateLimit int) *model.Tool {
	id := nextID()
	tool := &model.Tool{
		Name:        fmt.Sprintf("%s_%d", name, id),
		Category:    category,
		ToolType:    "function",
		Status:      "active",
		WorkspaceID: &wid,
		RateLimit:   rateLimit,
	}
	require.NoError(t, db.Create(tool).Error)
	return tool
}

func makeBuiltinTool(t *testing.T, db *gorm.DB, wid uint64, name string) *model.Tool {
	id := nextID()
	tool := &model.Tool{
		Name:        fmt.Sprintf("%s_%d", name, id),
		Category:    "general",
		ToolType:    "function",
		Status:      "active",
		IsBuiltin:   true,
		WorkspaceID: &wid,
	}
	require.NoError(t, db.Create(tool).Error)
	return tool
}

func clearRateLimiter() {
	globalRateLimiter.store.Range(func(key, _ interface{}) bool {
		globalRateLimiter.store.Delete(key)
		return true
	})
}

func resetRateLimiter() {
	globalRateLimiter.store.Range(func(key, _ interface{}) bool {
		globalRateLimiter.store.Delete(key)
		return true
	})
}
// ==================== Rate Limiter Tests ====================

func TestRateLimiter_SlidingWindowBasic(t *testing.T) {
	rl := &rateLimiter{}
	toolID := uint64(1)
	callerID := uint64(100)
	// callerLimit = ceil(25/5) = 5, global limit = 25
	for i := 0; i < 5; i++ {
		assert.True(t, rl.tryAcquire(toolID, callerID, 25))
	}
	assert.False(t, rl.tryAcquire(toolID, callerID, 25))
}

func TestRateLimiter_PerCallerLimit(t *testing.T) {
	rl := &rateLimiter{}
	toolID := uint64(2)
	callerID := uint64(200)
	for i := 0; i < 2; i++ {
		assert.True(t, rl.tryAcquire(toolID, callerID, 10))
	}
	assert.False(t, rl.tryAcquire(toolID, callerID, 10))
}

func TestRateLimiter_Unlimited(t *testing.T) {
	rl := &rateLimiter{}
	toolID := uint64(3)
	callerID := uint64(300)
	for i := 0; i < 20; i++ {
		assert.True(t, rl.tryAcquire(toolID, callerID, 0))
	}
}

func TestRateLimiter_DBToolRateLimit(t *testing.T) {
	resetRateLimiter()
	db := setupToolTestDB(t)
	svc := NewToolService(db)

	ws := &model.Workspace{Name: "ws_rl1", Slug: "ws_rl1"}
	require.NoError(t, db.Create(ws).Error)
	user := makeUser(t, db, "rl1_user")
	tool := makeTool(t, db, ws.ID, "rl1_tool", "general", 10) // 10 rpm, callerLimit = ceil(10/5) = 2
	member := &model.WorkspaceMember{WorkspaceID: ws.ID, UserID: user.ID, Role: common.RoleMember, IsActive: true}
	require.NoError(t, db.Create(member).Error)

	// Test permission check succeeds for workspace member
	req := request.CallToolRequest{ToolID: tool.ID, CallerUserID: user.ID, InputParams: json.RawMessage(`{}`)}
	_, err := svc.checkPermissions(ws.ID, tool, &req)
	assert.NoError(t, err)

	// Test rate limiter directly: callerLimit=2, so first 2 acquire succeed, 3rd fails
	for i := 0; i < 2; i++ {
		assert.True(t, svc.checkRateLimit(tool.ID, user.ID, tool.RateLimit), "call %d should succeed", i)
	}
	assert.False(t, svc.checkRateLimit(tool.ID, user.ID, tool.RateLimit), "3rd call should be rate limited")
}
// ==================== Permission Tests ====================

func TestPermissionStep1_NotWorkspaceMember(t *testing.T) {
	db := setupToolTestDB(t)
	svc := NewToolService(db)

	ws := &model.Workspace{Name: "ws_perm1", Slug: "ws_perm1"}
	require.NoError(t, db.Create(ws).Error)
	tool := makeTool(t, db, ws.ID, "perm1_tool", "general", 0)

	// Create a user who is NOT a member of the workspace
	outsider := makeUser(t, db, "outsider")

	req := request.CallToolRequest{ToolID: tool.ID, CallerUserID: outsider.ID, InputParams: json.RawMessage(`{}`)}
	_, err := svc.checkPermissions(ws.ID, tool, &req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Workspace member required")
}

func TestPermissionStep1_WorkspaceMember(t *testing.T) {
	db := setupToolTestDB(t)
	svc := NewToolService(db)

	ws := &model.Workspace{Name: "ws_perm1b", Slug: "ws_perm1b"}
	require.NoError(t, db.Create(ws).Error)
	user := makeUser(t, db, "member1")
	tool := makeTool(t, db, ws.ID, "perm1b_tool", "general", 0)
	member := &model.WorkspaceMember{WorkspaceID: ws.ID, UserID: user.ID, Role: common.RoleMember, IsActive: true}
	require.NoError(t, db.Create(member).Error)

	req := request.CallToolRequest{ToolID: tool.ID, CallerUserID: user.ID, InputParams: json.RawMessage(`{}`)}
	_, err := svc.checkPermissions(ws.ID, tool, &req)
	assert.NoError(t, err)
}

func TestPermissionStep2_DangerousAdmin(t *testing.T) {
	db := setupToolTestDB(t)
	svc := NewToolService(db)

	ws := &model.Workspace{Name: "ws_perm2", Slug: "ws_perm2"}
	require.NoError(t, db.Create(ws).Error)
	tool := makeTool(t, db, ws.ID, "perm2_danger", "dangerous", 0)

	// Non-admin member should be denied
	member := makeUser(t, db, "non_admin")
	wm := &model.WorkspaceMember{WorkspaceID: ws.ID, UserID: member.ID, Role: common.RoleMember, IsActive: true}
	require.NoError(t, db.Create(wm).Error)

	req := request.CallToolRequest{ToolID: tool.ID, CallerUserID: member.ID, InputParams: json.RawMessage(`{}`)}
	_, err := svc.checkPermissions(ws.ID, tool, &req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Admin required for dangerous tools")

	// Admin should be allowed
	admin := makeUser(t, db, "admin_user")
	wa := &model.WorkspaceMember{WorkspaceID: ws.ID, UserID: admin.ID, Role: common.RoleAdmin, IsActive: true}
	require.NoError(t, db.Create(wa).Error)

	req2 := request.CallToolRequest{ToolID: tool.ID, CallerUserID: admin.ID, InputParams: json.RawMessage(`{}`)}
	_, err = svc.checkPermissions(ws.ID, tool, &req2)
	assert.NoError(t, err)
}

func TestPermissionStep2_SafeTool(t *testing.T) {
	db := setupToolTestDB(t)
	svc := NewToolService(db)

	ws := &model.Workspace{Name: "ws_perm2b", Slug: "ws_perm2b"}
	require.NoError(t, db.Create(ws).Error)
	tool := makeTool(t, db, ws.ID, "perm2b_safe", "general", 0)

	member := makeUser(t, db, "safe_member")
	wm := &model.WorkspaceMember{WorkspaceID: ws.ID, UserID: member.ID, Role: common.RoleMember, IsActive: true}
	require.NoError(t, db.Create(wm).Error)

	req := request.CallToolRequest{ToolID: tool.ID, CallerUserID: member.ID, InputParams: json.RawMessage(`{}`)}
	_, err := svc.checkPermissions(ws.ID, tool, &req)
	assert.NoError(t, err)
}

func TestPermissionStep3_WhitelistAllow(t *testing.T) {
	db := setupToolTestDB(t)
	svc := NewToolService(db)

	ws := &model.Workspace{Name: "ws_perm3a", Slug: "ws_perm3a"}
	require.NoError(t, db.Create(ws).Error)
	tool := makeTool(t, db, ws.ID, "perm3a_tool", "general", 0)

	member := makeUser(t, db, "perm3a_member")
	wm := &model.WorkspaceMember{WorkspaceID: ws.ID, UserID: member.ID, Role: common.RoleMember, IsActive: true}
	require.NoError(t, db.Create(wm).Error)

	agentID := uint64(42)
	// Create a whitelist (Allowed=true) for this agent
	tp := model.ToolPermission{
		WorkspaceID:    ws.ID,
		ToolID:         tool.ID,
		AgentTemplateID: &agentID,
		Allowed:        true,
	}
	require.NoError(t, db.Create(&tp).Error)

	req := request.CallToolRequest{
		ToolID:          tool.ID,
		CallerUserID:    member.ID,
		AgentTemplateID: &agentID,
		InputParams:     json.RawMessage(`{}`),
	}
	_, err := svc.checkPermissions(ws.ID, tool, &req)
	assert.NoError(t, err)
}
func TestPermissionStep3_WhitelistDeny(t *testing.T) {
	db := setupToolTestDB(t)
	svc := NewToolService(db)

	ws := &model.Workspace{Name: "ws_perm3b", Slug: "ws_perm3b"}
	require.NoError(t, db.Create(ws).Error)
	tool := makeTool(t, db, ws.ID, "perm3b_tool", "general", 0)

	member := makeUser(t, db, "perm3b_member")
	wm := &model.WorkspaceMember{WorkspaceID: ws.ID, UserID: member.ID, Role: common.RoleMember, IsActive: true}
	require.NoError(t, db.Create(wm).Error)

	agentID := uint64(99)
	// Use raw SQL to insert with Allowed=false, bypassing GORM's default:true tag
	err := db.Exec(
		"INSERT INTO tool_permissions (workspace_id, tool_id, agent_template_id, allowed, created_at, updated_at) VALUES (?, ?, ?, 0, datetime('now'), datetime('now'))",
		ws.ID, tool.ID, agentID,
	).Error
	require.NoError(t, err)

	req := request.CallToolRequest{
		ToolID:          tool.ID,
		CallerUserID:    member.ID,
		AgentTemplateID: &agentID,
		InputParams:     json.RawMessage(`{}`),
	}
	_, err = svc.checkPermissions(ws.ID, tool, &req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Tool permission denied")
}

func TestPermissionStep3_WildcardDeny(t *testing.T) {
	db := setupToolTestDB(t)
	svc := NewToolService(db)

	ws := &model.Workspace{Name: "ws_perm3c", Slug: "ws_perm3c"}
	require.NoError(t, db.Create(ws).Error)
	tool := makeTool(t, db, ws.ID, "perm3c_tool", "general", 0)

	member := makeUser(t, db, "perm3c_member")
	wm := &model.WorkspaceMember{WorkspaceID: ws.ID, UserID: member.ID, Role: common.RoleMember, IsActive: true}
	require.NoError(t, db.Create(wm).Error)

	// Create a wildcard deny (agent_template_id = NULL, Allowed=false) using raw SQL
	err := db.Exec(
		"INSERT INTO tool_permissions (workspace_id, tool_id, agent_template_id, allowed, created_at, updated_at) VALUES (?, ?, NULL, 0, datetime('now'), datetime('now'))",
		ws.ID, tool.ID,
	).Error
	require.NoError(t, err)

	agentID := uint64(123)
	req := request.CallToolRequest{
		ToolID:          tool.ID,
		CallerUserID:    member.ID,
		AgentTemplateID: &agentID,
		InputParams:     json.RawMessage(`{}`),
	}
	_, err = svc.checkPermissions(ws.ID, tool, &req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Tool permission denied")
}

func TestPermissionStep3_NoAgentTemplate(t *testing.T) {
	db := setupToolTestDB(t)
	svc := NewToolService(db)

	ws := &model.Workspace{Name: "ws_perm3d", Slug: "ws_perm3d"}
	require.NoError(t, db.Create(ws).Error)
	tool := makeTool(t, db, ws.ID, "perm3d_tool", "general", 0)

	member := makeUser(t, db, "perm3d_member")
	wm := &model.WorkspaceMember{WorkspaceID: ws.ID, UserID: member.ID, Role: common.RoleMember, IsActive: true}
	require.NoError(t, db.Create(wm).Error)

	// No AgentTemplateID → permission check is skipped
	req := request.CallToolRequest{
		ToolID:          tool.ID,
		CallerUserID:    member.ID,
		AgentTemplateID: nil,
		InputParams:     json.RawMessage(`{}`),
	}
	_, err := svc.checkPermissions(ws.ID, tool, &req)
	assert.NoError(t, err)
}

// ==================== Builtin Function Tests ====================

func TestBuiltin_GetPRDiff(t *testing.T) {
	db := setupToolTestDB(t)
	svc := NewToolService(db)

	ws := &model.Workspace{Name: "ws_builtin", Slug: "ws_builtin"}
	require.NoError(t, db.Create(ws).Error)
	tool := makeBuiltinTool(t, db, ws.ID, "get_pr_diff")
	member := makeUser(t, db, "builtin_member")
	wm := &model.WorkspaceMember{WorkspaceID: ws.ID, UserID: member.ID, Role: common.RoleMember, IsActive: true}
	require.NoError(t, db.Create(wm).Error)

	// This builtin function checks for env vars, so it will fail gracefully
	// but the tool type routing should work
	req := request.CallToolRequest{
		ToolID:       tool.ID,
		CallerUserID: member.ID,
		InputParams:  json.RawMessage(`{"repo_owner": "test", "repo_name": "repo", "pr_number": 1}`),
	}
	resp, err := svc.Call(ws.ID, req)
	// The function fails due to missing GITHUB_TOKEN env var, but the tool routing works
	assert.NoError(t, err)
	assert.Equal(t, "failed", resp.Status)
	assert.NotNil(t, resp.ErrorMessage)
}