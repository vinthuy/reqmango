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

func setupAgentTemplateTestDB(t *testing.T) *gorm.DB {
	db := openTestDB(t)
	err := db.AutoMigrate(
		&model.User{},
		&model.Workspace{},
		&model.AgentTemplate{},
		&model.WorkspaceMember{},
		&model.Skill{},
		&model.SkillExecutionLog{},
	)
	require.NoError(t, err)
	return db
}

func createTemplateAdmin(t *testing.T, db *gorm.DB, wsID uint64) *model.User {
	user := makeUser(t, db, "tmpl_admin")
	member := &model.WorkspaceMember{
		WorkspaceID: wsID,
		UserID:      user.ID,
		Role:        common.RoleAdmin,
		IsActive:    true,
	}
	require.NoError(t, db.Create(member).Error)
	return user
}

func createTemplateMember(t *testing.T, db *gorm.DB, wsID uint64) *model.User {
	user := makeUser(t, db, "tmpl_member")
	member := &model.WorkspaceMember{
		WorkspaceID: wsID,
		UserID:      user.ID,
		Role:        common.RoleMember,
		IsActive:    true,
	}
	require.NoError(t, db.Create(member).Error)
	return user
}

func makeAgentTemplate(t *testing.T, db *gorm.DB, wsID uint64, name, systemPrompt string) *model.AgentTemplate {
	id := nextID()
	tmpl := &model.AgentTemplate{
		Name:         fmt.Sprintf("%s_%d", name, id),
		SystemPrompt: systemPrompt,
		Status:       "active",
		Version:      "1.0",
		WorkspaceID:  &wsID,
	}
	require.NoError(t, db.Create(tmpl).Error)
	return tmpl
}

func makePresetTemplate(t *testing.T, db *gorm.DB, name, systemPrompt string) *model.AgentTemplate {
	id := nextID()
	tmpl := &model.AgentTemplate{
		Name:         fmt.Sprintf("preset_%s_%d", name, id),
		SystemPrompt: systemPrompt,
		IsPreset:     true,
		Status:       "active",
		Version:      "1.0",
	}
	require.NoError(t, db.Create(tmpl).Error)
	return tmpl
}

// ==================== Create ====================

func TestAgentTemplateService_Create(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl1", Slug: "ws_tmpl1", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createTemplateAdmin(t, db, ws.ID)

	req := request.AgentTemplateCreate{
		Name:         "Code Reviewer",
		SystemPrompt: "You are a code reviewer.",
		Icon:         "🔍",
	}

	resp, err := svc.Create(ws.ID, admin.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Code Reviewer", resp.Name)
	assert.Equal(t, "You are a code reviewer.", resp.SystemPrompt)
	assert.Equal(t, "active", resp.Status)
	assert.Equal(t, "1.0", resp.Version)
	assert.False(t, resp.IsPreset)
	assert.Equal(t, ws.ID, *resp.WorkspaceID)
}

func TestAgentTemplateService_Create_NonAdmin(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl_na", Slug: "ws_tmpl_na", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	member := createTemplateMember(t, db, ws.ID)

	req := request.AgentTemplateCreate{
		Name:         "Should Fail Template",
		SystemPrompt: "test",
	}

	_, err := svc.Create(ws.ID, member.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestAgentTemplateService_Create_NoMember(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl_nm", Slug: "ws_tmpl_nm", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	outsider := makeUser(t, db, "outsider_tmpl")

	req := request.AgentTemplateCreate{
		Name:         "Outsider Template",
		SystemPrompt: "test",
	}

	_, err := svc.Create(ws.ID, outsider.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestAgentTemplateService_Create_WithSkills(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl_sk", Slug: "ws_tmpl_sk", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createTemplateAdmin(t, db, ws.ID)

	// Create some skills
	skill1 := &model.Skill{
		Name:        "Skill A",
		SkillType:   "custom",
		SkillMD:     "# Skill A",
		WorkspaceID: ws.ID,
		Status:      "active",
	}
	skill2 := &model.Skill{
		Name:        "Skill B",
		SkillType:   "custom",
		SkillMD:     "# Skill B",
		WorkspaceID: ws.ID,
		Status:      "active",
	}
	require.NoError(t, db.Create(skill1).Error)
	require.NoError(t, db.Create(skill2).Error)

	skillsJSON, _ := json.Marshal([]uint64{skill1.ID, skill2.ID})
	req := request.AgentTemplateCreate{
		Name:            "Skilled Agent",
		SystemPrompt:    "You have skills.",
		AvailableSkills: skillsJSON,
	}

	resp, err := svc.Create(ws.ID, admin.ID, req)
	require.NoError(t, err)
	assert.NotNil(t, resp.AvailableSkills)
}

func TestAgentTemplateService_Create_InvalidSkillIDs(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl_isi", Slug: "ws_tmpl_isi", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createTemplateAdmin(t, db, ws.ID)

	skillsJSON, _ := json.Marshal([]uint64{99999, 88888})
	req := request.AgentTemplateCreate{
		Name:            "Bad Skills Agent",
		SystemPrompt:    "test",
		AvailableSkills: skillsJSON,
	}

	_, err := svc.Create(ws.ID, admin.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "do not exist")
}

// ==================== Get ====================

func TestAgentTemplateService_Get(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl_get", Slug: "ws_tmpl_get", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	tmpl := makeAgentTemplate(t, db, ws.ID, "get_tmpl", "Get prompt")

	resp, err := svc.Get(tmpl.ID)
	require.NoError(t, err)
	assert.Equal(t, tmpl.ID, resp.ID)
	assert.Equal(t, tmpl.Name, resp.Name)
}

func TestAgentTemplateService_Get_NotFound(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	_, err := svc.Get(99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAgentTemplateService_Get_PresetTemplate(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	preset := makePresetTemplate(t, db, "preset_get", "Preset prompt")

	resp, err := svc.Get(preset.ID)
	require.NoError(t, err)
	assert.True(t, resp.IsPreset)
}

// ==================== List ====================

func TestAgentTemplateService_List(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl_list", Slug: "ws_tmpl_list", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	makeAgentTemplate(t, db, ws.ID, "tmpl_a", "Prompt A")
	makeAgentTemplate(t, db, ws.ID, "tmpl_b", "Prompt B")

	resp, err := svc.List(ws.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(resp), 2)
}

func TestAgentTemplateService_List_IncludesPresets(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl_lp", Slug: "ws_tmpl_lp", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	makeAgentTemplate(t, db, ws.ID, "user_tmpl", "User prompt")
	makePresetTemplate(t, db, "preset1", "Preset prompt 1")

	resp, err := svc.List(ws.ID)
	require.NoError(t, err)

	// Should include both workspace templates AND preset templates
	hasPreset := false
	hasUser := false
	for _, r := range resp {
		if r.IsPreset {
			hasPreset = true
		}
		if r.Name != "" {
			hasUser = true
		}
	}
	assert.True(t, hasPreset, "should include preset templates")
	assert.True(t, hasUser, "should include user templates")
}

func TestAgentTemplateService_List_Empty(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	resp, err := svc.List(99999)
	require.NoError(t, err)
	// May have preset templates even for empty workspace
	assert.NotNil(t, resp)
}

// ==================== Update ====================

func TestAgentTemplateService_Update(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl_upd", Slug: "ws_tmpl_upd", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createTemplateAdmin(t, db, ws.ID)

	tmpl := makeAgentTemplate(t, db, ws.ID, "upd_tmpl", "Old prompt")

	newName := "Updated Template"
	newPrompt := "New system prompt"
	newIcon := "🚀"
	req := request.AgentTemplateUpdate{
		Name:         &newName,
		SystemPrompt: &newPrompt,
		Icon:         &newIcon,
	}

	resp, err := svc.Update(tmpl.ID, admin.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Updated Template", resp.Name)
	assert.Equal(t, "New system prompt", resp.SystemPrompt)
	assert.Equal(t, "🚀", resp.Icon)
}

func TestAgentTemplateService_Update_NotFound(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	newName := "Ghost"
	req := request.AgentTemplateUpdate{Name: &newName}

	_, err := svc.Update(99999, 1, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAgentTemplateService_Update_NonAdmin(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl_una", Slug: "ws_tmpl_una", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	member := createTemplateMember(t, db, ws.ID)

	tmpl := makeAgentTemplate(t, db, ws.ID, "una_tmpl", "Prompt")

	newName := "Should Not Update"
	req := request.AgentTemplateUpdate{Name: &newName}

	_, err := svc.Update(tmpl.ID, member.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

// ==================== Delete ====================

func TestAgentTemplateService_Delete(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl_del", Slug: "ws_tmpl_del", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	admin := createTemplateAdmin(t, db, ws.ID)

	tmpl := makeAgentTemplate(t, db, ws.ID, "del_tmpl", "Delete me")

	err := svc.Delete(tmpl.ID, admin.ID)
	require.NoError(t, err)

	_, err = svc.Get(tmpl.ID)
	assert.Error(t, err)
}

func TestAgentTemplateService_Delete_PresetTemplate(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	preset := makePresetTemplate(t, db, "preset_del", "Cannot delete me")

	err := svc.Delete(preset.ID, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot delete preset")
}

func TestAgentTemplateService_Delete_NonAdmin(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	ws := &model.Workspace{Name: "ws_tmpl_dna", Slug: "ws_tmpl_dna", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)
	member := createTemplateMember(t, db, ws.ID)

	tmpl := makeAgentTemplate(t, db, ws.ID, "dna_tmpl", "Non-admin delete")

	err := svc.Delete(tmpl.ID, member.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestAgentTemplateService_Delete_NotFound(t *testing.T) {
	db := setupAgentTemplateTestDB(t)
	svc := NewAgentTemplateService(db)

	err := svc.Delete(99999, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
