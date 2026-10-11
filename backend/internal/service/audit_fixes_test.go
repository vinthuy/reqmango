package service

import (
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openAuditTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	return db
}

func appErrorCode(t *testing.T, err error) int {
	t.Helper()
	ae, ok := err.(*common.AppError)
	require.True(t, ok, "expected *common.AppError, got %T: %v", err, err)
	return ae.Code
}

func TestIssueAgentAssignRejectsForeignAgent(t *testing.T) {
	db := openAuditTestDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE issues (id integer primary key, project_id integer, workspace_id integer, agent_assignee_id integer, deleted_at datetime)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE agents (id integer primary key, workspace_id integer, deleted_at datetime)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO issues (id, project_id, workspace_id) VALUES (1, 7, 10)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO agents (id, workspace_id) VALUES (5, 20)`).Error)

	svc := &IssueAgentService{db: db}

	_, err := svc.Assign(1, AssignRequest{AgentID: 0})
	require.Error(t, err)
	assert.Equal(t, 400, appErrorCode(t, err))

	_, err = svc.Assign(1, AssignRequest{AgentID: 5})
	require.Error(t, err)
	assert.Equal(t, 404, appErrorCode(t, err))

	_, err = svc.Assign(99, AssignRequest{AgentID: 5})
	require.Error(t, err)
	assert.Equal(t, 404, appErrorCode(t, err))

	var assignee *uint64
	require.NoError(t, db.Raw(`SELECT agent_assignee_id FROM issues WHERE id = 1`).Scan(&assignee).Error)
	assert.Nil(t, assignee)
}

func TestEstimatePointUsesProjectWorkspace(t *testing.T) {
	db := openAuditTestDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE projects (id integer primary key, workspace_id integer, deleted_at datetime)`).Error)
	require.NoError(t, db.AutoMigrate(&model.EstimatePoint{}))
	require.NoError(t, db.Exec(`INSERT INTO projects (id, workspace_id) VALUES (3, 10)`).Error)

	svc := NewEstimateService(db)
	p, err := svc.CreatePoint(3, 999, "1", 1, false, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(10), p.WorkspaceID)

	_, err = svc.CreatePoint(404, 10, "1", 1, false, 1)
	require.Error(t, err)
	assert.Equal(t, 404, appErrorCode(t, err))
}

func TestSkillExecutionLogSortIsWhitelisted(t *testing.T) {
	db := openAuditTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.SkillExecutionLog{}))
	require.NoError(t, db.Create(&model.SkillExecutionLog{
		SkillID: 1, WorkspaceID: 10, Status: "completed",
		InputParams: json.RawMessage(`{}`), OutputResult: json.RawMessage(`{}`),
	}).Error)

	svc := NewSkillService(db)
	logs, total, err := svc.ListExecutionLogs(SkillExecutionLogQueryParams{
		WorkspaceID: 10,
		SortBy:      "id; DROP TABLE skill_execution_logs; --",
		SortOrder:   "desc; --",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, logs, 1)
	assert.True(t, db.Migrator().HasTable(&model.SkillExecutionLog{}))
}
