package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"testing"

	gosqlite "github.com/glebarez/go-sqlite"
	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ---------- JSON-safe database/sql driver wrapper ----------
// SQLite returns string for TEXT columns; json.RawMessage (which is []byte)
// cannot scan from string.  This wrapper converts every string in result
// sets to []byte so that json.RawMessage scans correctly.

type jsonSafeDriver struct {
	underlying driver.Driver
}

func (d *jsonSafeDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.underlying.Open(name)
	if err != nil {
		return nil, err
	}
	return &jsonSafeConn{Conn: conn}, nil
}

type jsonSafeConn struct {
	driver.Conn
}

func (c *jsonSafeConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if qc, ok := c.Conn.(driver.QueryerContext); ok {
		rows, err := qc.QueryContext(ctx, query, args)
		if err != nil {
			return nil, err
		}
		return &jsonSafeRows{Rows: rows}, nil
	}
	return nil, driver.ErrSkip
}

func (c *jsonSafeConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if ec, ok := c.Conn.(driver.ExecerContext); ok {
		return ec.ExecContext(ctx, query, args)
	}
	return nil, driver.ErrSkip
}

func (c *jsonSafeConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if pc, ok := c.Conn.(driver.ConnPrepareContext); ok {
		return pc.PrepareContext(ctx, query)
	}
	return c.Conn.Prepare(query)
}

type jsonSafeRows struct {
	driver.Rows
}

func (r *jsonSafeRows) Next(dest []driver.Value) error {
	err := r.Rows.Next(dest)
	if err != nil {
		return err
	}
	for i, v := range dest {
		if s, ok := v.(string); ok {
			dest[i] = []byte(s)
		}
	}
	return nil
}

// Register the wrapper driver once.
func init() {
	sql.Register("json_safe_sqlite", &jsonSafeDriver{underlying: &gosqlite.Driver{}})
}

// openTestDB opens an in-memory SQLite database with JSON-safe scanning.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	sqlDB, err := sql.Open("json_safe_sqlite", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, sqlDB)

	gormDB, err := gorm.Open(&sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NotNil(t, gormDB)
	return gormDB
}

func setupSkillTestDB(t *testing.T) *gorm.DB {
	db := openTestDB(t)
	err := db.AutoMigrate(
		&model.User{},
		&model.Workspace{},
		&model.Skill{},
		&model.SkillExecutionLog{},
	)
	require.NoError(t, err)
	return db
}

func makeSkill(t *testing.T, db *gorm.DB, wid uint64, name string) *model.Skill {
	id := nextID()
	skill := &model.Skill{
		Name:        fmt.Sprintf("%s_%d", name, id),
		SkillType:   "custom",
		Version:     "1.0",
		Status:      "active",
		SkillMD:     "# Test Skill\nExecute this skill",
		WorkspaceID: wid,
		UsageCount:  0,
		IsShared:    false,
	}
	require.NoError(t, db.Create(skill).Error)
	return skill
}

// ==================== Create ====================

func TestSkillService_Create(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	ws := &model.Workspace{Name: "ws_skill1", Slug: "ws_skill1", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	desc := "A test skill for code review"
	req := request.SkillCreate{
		Name:        "Code Review Skill",
		Description: &desc,
		SkillType:   "custom",
		SkillMD:     "# Code Review\nAutomated code review",
		IsShared:    false,
	}

	resp, err := svc.Create(ws.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Code Review Skill", resp.Name)
	assert.Equal(t, "custom", resp.SkillType)
	assert.Equal(t, "1.0", resp.Version)
	assert.Equal(t, "active", resp.Status)
	assert.Equal(t, ws.ID, resp.WorkspaceID)
	assert.Equal(t, 0, resp.UsageCount)
	assert.Equal(t, "# Code Review\nAutomated code review", resp.SkillMD)
}

func TestSkillService_Create_MinimalFields(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	ws := &model.Workspace{Name: "ws_skill_min", Slug: "ws_skill_min", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	req := request.SkillCreate{
		Name:    "Minimal Skill",
		SkillMD: "# Minimal",
	}

	resp, err := svc.Create(ws.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Minimal Skill", resp.Name)
	assert.Equal(t, "custom", resp.SkillType)
}

// ==================== Get ====================

func TestSkillService_Get(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	ws := &model.Workspace{Name: "ws_skill_get", Slug: "ws_skill_get", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	skill := makeSkill(t, db, ws.ID, "get_test")

	resp, err := svc.Get(skill.ID)
	require.NoError(t, err)
	assert.Equal(t, skill.ID, resp.ID)
	assert.Equal(t, skill.Name, resp.Name)
	assert.Equal(t, ws.ID, resp.WorkspaceID)
}

func TestSkillService_Get_NotFound(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	_, err := svc.Get(99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== List ====================

func TestSkillService_List(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	ws := &model.Workspace{Name: "ws_skill_list", Slug: "ws_skill_list", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	makeSkill(t, db, ws.ID, "list_skill1")
	makeSkill(t, db, ws.ID, "list_skill2")
	makeSkill(t, db, ws.ID, "list_skill3")

	resp, err := svc.List(ws.ID)
	require.NoError(t, err)
	assert.Len(t, resp, 3)
}

func TestSkillService_List_OtherWorkspace(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	ws1 := &model.Workspace{Name: "ws_skill_list1", Slug: "ws_skill_list1", OwnerID: 1}
	ws2 := &model.Workspace{Name: "ws_skill_list2", Slug: "ws_skill_list2", OwnerID: 1}
	require.NoError(t, db.Create(ws1).Error)
	require.NoError(t, db.Create(ws2).Error)

	makeSkill(t, db, ws1.ID, "ws1_skill")
	makeSkill(t, db, ws2.ID, "ws2_skill")

	resp1, err := svc.List(ws1.ID)
	require.NoError(t, err)
	assert.Len(t, resp1, 1)

	resp2, err := svc.List(ws2.ID)
	require.NoError(t, err)
	assert.Len(t, resp2, 1)
}

func TestSkillService_List_Empty(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	resp, err := svc.List(99999)
	require.NoError(t, err)
	assert.Len(t, resp, 0)
}

// ==================== Update ====================

func TestSkillService_Update(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	ws := &model.Workspace{Name: "ws_skill_upd", Slug: "ws_skill_upd", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	skill := makeSkill(t, db, ws.ID, "upd_skill")

	newName := "Updated Skill Name"
	newSkillMD := "# Updated\nNew content"
	isShared := true
	req := request.SkillUpdate{
		Name:     &newName,
		SkillMD:  &newSkillMD,
		IsShared: &isShared,
	}

	resp, err := svc.Update(skill.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Updated Skill Name", resp.Name)
	assert.Equal(t, "# Updated\nNew content", resp.SkillMD)
	assert.True(t, resp.IsShared)
}

func TestSkillService_Update_NotFound(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	newName := "Ghost"
	req := request.SkillUpdate{Name: &newName}

	_, err := svc.Update(99999, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Delete ====================

func TestSkillService_Delete(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	ws := &model.Workspace{Name: "ws_skill_del", Slug: "ws_skill_del", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	skill := makeSkill(t, db, ws.ID, "del_skill")

	err := svc.Delete(skill.ID)
	require.NoError(t, err)

	_, err = svc.Get(skill.ID)
	assert.Error(t, err)
}

func TestSkillService_Delete_NotFound(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	err := svc.Delete(99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== IncrementUsage ====================

func TestSkillService_IncrementUsage(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	ws := &model.Workspace{Name: "ws_skill_inc", Slug: "ws_skill_inc", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	skill := makeSkill(t, db, ws.ID, "inc_skill")
	assert.Equal(t, 0, skill.UsageCount)

	err := svc.IncrementUsage(skill.ID)
	require.NoError(t, err)

	resp, err := svc.Get(skill.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.UsageCount)

	err = svc.IncrementUsage(skill.ID)
	require.NoError(t, err)

	resp, err = svc.Get(skill.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.UsageCount)
}

// ==================== Execute ====================

func TestSkillService_Execute(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	ws := &model.Workspace{Name: "ws_skill_exec", Slug: "ws_skill_exec", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	skill := makeSkill(t, db, ws.ID, "exec_skill")

	params := json.RawMessage(`{"input": "test data"}`)
	req := request.SkillExecute{
		Parameters: params,
	}

	resp, err := svc.Execute(context.Background(), skill.ID, req)
	require.NoError(t, err)
	assert.Equal(t, skill.ID, resp.SkillID)
	assert.Equal(t, skill.Name, resp.SkillName)
	assert.NotEmpty(t, resp.Steps)
	assert.Equal(t, "completed", resp.Steps[0].Status)
}

func TestSkillService_Execute_NotFound(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	req := request.SkillExecute{}
	_, err := svc.Execute(context.Background(), 99999, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSkillService_Execute_InactiveSkill(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	ws := &model.Workspace{Name: "ws_skill_exec2", Slug: "ws_skill_exec2", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	skill := makeSkill(t, db, ws.ID, "inactive_skill")
	err := db.Model(skill).Update("status", "deprecated").Error
	require.NoError(t, err)

	req := request.SkillExecute{}
	_, err = svc.Execute(context.Background(), skill.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

func TestSkillService_Execute_RecordsLog(t *testing.T) {
	db := setupSkillTestDB(t)
	svc := NewSkillService(db)

	ws := &model.Workspace{Name: "ws_skill_log", Slug: "ws_skill_log", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	skill := makeSkill(t, db, ws.ID, "log_skill")

	req := request.SkillExecute{
		Parameters: json.RawMessage(`{"key": "value"}`),
	}

	_, err := svc.Execute(context.Background(), skill.ID, req)
	require.NoError(t, err)

	var count int64
	db.Model(&model.SkillExecutionLog{}).Where("skill_id = ?", skill.ID).Count(&count)
	assert.Equal(t, int64(1), count)

	resp, err := svc.Get(skill.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.UsageCount)
}

// ==================== WorkspaceMember (for admin checks) ====================

func setupSkillTestDBWithMember(t *testing.T) (*gorm.DB, uint64, uint64) {
	db := setupSkillTestDB(t)
	err := db.AutoMigrate(&model.WorkspaceMember{})
	require.NoError(t, err)

	user := makeUser(t, db, "skill_admin")
	ws := &model.Workspace{Name: "ws_skill_admin", Slug: "ws_skill_admin", OwnerID: user.ID}
	require.NoError(t, db.Create(ws).Error)

	member := &model.WorkspaceMember{
		WorkspaceID: ws.ID,
		UserID:      user.ID,
		Role:        common.RoleAdmin,
		IsActive:    true,
	}
	require.NoError(t, db.Create(member).Error)

	return db, ws.ID, user.ID
}

func TestSkillService_Create_WithWorkspaceMember(t *testing.T) {
	db, wsID, _ := setupSkillTestDBWithMember(t)
	svc := NewSkillService(db)

	req := request.SkillCreate{
		Name:    "Member Skill",
		SkillMD: "# Skill with member context",
	}

	resp, err := svc.Create(wsID, req)
	require.NoError(t, err)
	assert.Equal(t, "Member Skill", resp.Name)
}
