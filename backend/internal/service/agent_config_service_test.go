package service

import (
	"fmt"
	"testing"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupAgentConfigTestDB(t *testing.T) *gorm.DB {
	db := openTestDB(t)
	err := db.AutoMigrate(
		&model.User{},
		&model.Workspace{},
		&model.AgentConfig{},
		&model.WorkspaceMember{},
	)
	require.NoError(t, err)
	return db
}

func createConfigAdmin(t *testing.T, db *gorm.DB, wsID uint64) *model.User {
	user := makeUser(t, db, "cfg_admin")
	member := &model.WorkspaceMember{
		WorkspaceID: wsID,
		UserID:      user.ID,
		Role:        common.RoleAdmin,
		IsActive:    true,
	}
	require.NoError(t, db.Create(member).Error)
	return user
}

func createConfigMember(t *testing.T, db *gorm.DB, wsID uint64) *model.User {
	user := makeUser(t, db, "cfg_member")
	member := &model.WorkspaceMember{
		WorkspaceID: wsID,
		UserID:      user.ID,
		Role:        common.RoleMember,
		IsActive:    true,
	}
	require.NoError(t, db.Create(member).Error)
	return user
}

func makeAgentConfig(t *testing.T, db *gorm.DB, wsID uint64, name, provider, modelName string) *model.AgentConfig {
	id := nextID()
	cfg := &model.AgentConfig{
		Name:           fmt.Sprintf("%s_%d", name, id),
		Provider:       provider,
		Model:          modelName,
		APIKey:         "test-api-key",
		InferenceLevel: "normal",
		ServiceLevel:   "standard",
		MaxTokens:      4096,
		Temperature:    0.7,
		TopP:           1.0,
		IsDefault:      false,
		IsActive:       true,
		WorkspaceID:    wsID,
	}
	require.NoError(t, db.Create(cfg).Error)
	return cfg
}

// ==================== Create ====================

func TestAgentConfigService_Create(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg1", Slug: "ws_cfg1", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createConfigAdmin(t, db, ws.ID)

	req := request.AgentConfigCreate{
		Name:           "DeepSeek Config",
		Provider:       "deepseek",
		Model:          "deepseek-coder",
		APIKey:         "sk-test-123",
		InferenceLevel: "advanced",
		ServiceLevel:   "premium",
		MaxTokens:      8192,
		Temperature:    0.5,
		TopP:           0.9,
		IsDefault:      true,
	}

	resp, err := svc.Create(ws.ID, admin.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "DeepSeek Config", resp.Name)
	assert.Equal(t, "deepseek", resp.Provider)
	assert.Equal(t, "deepseek-coder", resp.Model)
	assert.Equal(t, "advanced", resp.InferenceLevel)
	assert.Equal(t, "premium", resp.ServiceLevel)
	assert.Equal(t, 8192, resp.MaxTokens)
	assert.Equal(t, 0.5, resp.Temperature)
	assert.Equal(t, 0.9, resp.TopP)
	assert.True(t, resp.IsDefault)
	assert.True(t, resp.IsActive)
	assert.Equal(t, ws.ID, resp.WorkspaceID)
}

func TestAgentConfigService_Create_NonAdmin(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg_na", Slug: "ws_cfg_na", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	member := createConfigMember(t, db, ws.ID)

	req := request.AgentConfigCreate{
		Name:     "Should Fail",
		Provider: "openai",
		Model:    "gpt-4",
		APIKey:   "key",
	}

	_, err := svc.Create(ws.ID, member.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestAgentConfigService_Create_NoMember(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg_nm", Slug: "ws_cfg_nm", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	outsider := makeUser(t, db, "outsider_cfg")

	req := request.AgentConfigCreate{
		Name:     "Outsider Config",
		Provider: "anthropic",
		Model:    "claude-3",
		APIKey:   "key",
	}

	_, err := svc.Create(ws.ID, outsider.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

// ==================== Get ====================

func TestAgentConfigService_Get(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg_get", Slug: "ws_cfg_get", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	cfg := makeAgentConfig(t, db, ws.ID, "get_cfg", "openai", "gpt-4")

	resp, err := svc.Get(cfg.ID)
	require.NoError(t, err)
	assert.Equal(t, cfg.ID, resp.ID)
	assert.Equal(t, cfg.Name, resp.Name)
	assert.Equal(t, "openai", resp.Provider)
	assert.Equal(t, ws.ID, resp.WorkspaceID)
}

func TestAgentConfigService_Get_NotFound(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	_, err := svc.Get(99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== List ====================

func TestAgentConfigService_List(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg_list", Slug: "ws_cfg_list", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	makeAgentConfig(t, db, ws.ID, "cfg_a", "openai", "gpt-4")
	makeAgentConfig(t, db, ws.ID, "cfg_b", "anthropic", "claude-3")

	resp, err := svc.List(ws.ID)
	require.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestAgentConfigService_List_Empty(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	resp, err := svc.List(99999)
	require.NoError(t, err)
	assert.Len(t, resp, 0)
}

func TestAgentConfigService_List_OtherWorkspace(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws1 := &model.Workspace{Name: "ws_cfg_l1", Slug: "ws_cfg_l1", OwnerID: 1}
	ws2 := &model.Workspace{Name: "ws_cfg_l2", Slug: "ws_cfg_l2", OwnerID: 1}
	require.NoError(t, db.Create(ws1).Error)
	require.NoError(t, db.Create(ws2).Error)

	makeAgentConfig(t, db, ws1.ID, "ws1_cfg", "openai", "gpt-4")
	makeAgentConfig(t, db, ws2.ID, "ws2_cfg", "anthropic", "claude-3")

	resp1, err := svc.List(ws1.ID)
	require.NoError(t, err)
	assert.Len(t, resp1, 1)

	resp2, err := svc.List(ws2.ID)
	require.NoError(t, err)
	assert.Len(t, resp2, 1)
}

// ==================== Update ====================

func TestAgentConfigService_Update(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg_upd", Slug: "ws_cfg_upd", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createConfigAdmin(t, db, ws.ID)

	cfg := makeAgentConfig(t, db, ws.ID, "upd_cfg", "openai", "gpt-4")

	newName := "Updated Config"
	newModel := "gpt-4-turbo"
	maxTokens := 16384
	newTemp := 0.3
	req := request.AgentConfigUpdate{
		Name:        &newName,
		Model:       &newModel,
		MaxTokens:   &maxTokens,
		Temperature: &newTemp,
	}

	resp, err := svc.Update(cfg.ID, admin.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Updated Config", resp.Name)
	assert.Equal(t, "gpt-4-turbo", resp.Model)
	assert.Equal(t, 16384, resp.MaxTokens)
	assert.Equal(t, 0.3, resp.Temperature)
}

func TestAgentConfigService_Update_NotFound(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	newName := "Ghost"
	req := request.AgentConfigUpdate{Name: &newName}

	_, err := svc.Update(99999, 1, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAgentConfigService_Update_NonAdmin(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg_una", Slug: "ws_cfg_una", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	member := createConfigMember(t, db, ws.ID)

	cfg := makeAgentConfig(t, db, ws.ID, "una_cfg", "openai", "gpt-4")

	newName := "Should Not Update"
	req := request.AgentConfigUpdate{Name: &newName}

	_, err := svc.Update(cfg.ID, member.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestAgentConfigService_Update_SetDefault(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg_sd", Slug: "ws_cfg_sd", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createConfigAdmin(t, db, ws.ID)

	cfg1 := makeAgentConfig(t, db, ws.ID, "cfg1", "openai", "gpt-4")
	cfg1.IsDefault = true
	require.NoError(t, db.Save(cfg1).Error)

	cfg2 := makeAgentConfig(t, db, ws.ID, "cfg2", "anthropic", "claude-3")

	// Set cfg2 as default, should unset cfg1
	isDefault := true
	req := request.AgentConfigUpdate{
		IsDefault: &isDefault,
	}

	resp, err := svc.Update(cfg2.ID, admin.ID, req)
	require.NoError(t, err)
	assert.True(t, resp.IsDefault)

	// Verify cfg1 is no longer default
	cfg1Resp, err := svc.Get(cfg1.ID)
	require.NoError(t, err)
	assert.False(t, cfg1Resp.IsDefault)
}

// ==================== Delete ====================

func TestAgentConfigService_Delete(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg_del", Slug: "ws_cfg_del", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createConfigAdmin(t, db, ws.ID)

	cfg := makeAgentConfig(t, db, ws.ID, "del_cfg", "openai", "gpt-4")

	err := svc.Delete(cfg.ID, admin.ID)
	require.NoError(t, err)

	_, err = svc.Get(cfg.ID)
	assert.Error(t, err)
}

func TestAgentConfigService_Delete_NonAdmin(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg_dna", Slug: "ws_cfg_dna", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	member := createConfigMember(t, db, ws.ID)

	cfg := makeAgentConfig(t, db, ws.ID, "dna_cfg", "openai", "gpt-4")

	err := svc.Delete(cfg.ID, member.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestAgentConfigService_Delete_NotFound(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	err := svc.Delete(99999, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== GetDefault ====================

func TestAgentConfigService_GetDefault(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg_gd", Slug: "ws_cfg_gd", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	cfg := makeAgentConfig(t, db, ws.ID, "default_cfg", "deepseek", "deepseek-coder")
	cfg.IsDefault = true
	require.NoError(t, db.Save(cfg).Error)

	resp, err := svc.GetDefault(ws.ID)
	require.NoError(t, err)
	assert.Equal(t, cfg.ID, resp.ID)
	assert.True(t, resp.IsDefault)
}

func TestAgentConfigService_GetDefault_NotFound(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	_, err := svc.GetDefault(99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No default agent config found")
}

func TestAgentConfigService_GetDefault_NoDefaultSet(t *testing.T) {
	db := setupAgentConfigTestDB(t)
	svc := NewAgentConfigService(db)

	ws := &model.Workspace{Name: "ws_cfg_gd2", Slug: "ws_cfg_gd2", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	// Create config without setting as default
	makeAgentConfig(t, db, ws.ID, "not_default", "openai", "gpt-4")

	_, err := svc.GetDefault(ws.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No default agent config found")
}
