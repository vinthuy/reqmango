package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestExtractMemoryIssueHint(t *testing.T) {
	assert.Equal(t, memoryIssueHint{IssueID: 81}, extractMemoryIssueHint("任务：分诊\nissue_id=81\n结论"))
	assert.Equal(t, memoryIssueHint{IssueID: 704}, extractMemoryIssueHint("Issue ID: 704 Analyze"))
	hint := extractMemoryIssueHint("重复项 CORE-8 / CORE-68，本项 CORE-81 建议降级")
	assert.Equal(t, "CORE", hint.Identifier)
	assert.EqualValues(t, 81, hint.SequenceID)
	assert.EqualValues(t, 42, extractMemoryIssueHint("请分析工作项 #42").SequenceID)
	assert.Equal(t, memoryIssueHint{}, extractMemoryIssueHint("没有任何引用"))
}

func TestBackfillAgentMemoryIssueIDs_FillsAndExpires(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Project{}, &model.Issue{}, &model.MemoryEntry{}))

	require.NoError(t, db.Create(&model.Project{
		BaseModel: model.BaseModel{ID: 1}, Name: "P", Identifier: "CORE", WorkspaceID: 1,
	}).Error)
	require.NoError(t, db.Create(&model.Issue{
		BaseModel: model.BaseModel{ID: 81}, Name: "慢查询", ProjectID: 1, WorkspaceID: 1, SequenceID: 81, StateID: 1,
	}).Error)

	agentID := uint64(7)
	projectID := uint64(1)
	old := time.Now().Add(-40 * 24 * time.Hour)
	m1 := model.MemoryEntry{
		BaseModel: model.BaseModel{ID: 1}, WorkspaceID: 1, ProjectID: &projectID, AgentID: &agentID,
		MemoryType: model.MemoryMediumTerm, Scope: model.ScopeAgent,
		Content:  "任务：分诊 CORE-81\n执行结果：建议 medium",
		Metadata: json.RawMessage(`{}`), Tags: json.RawMessage(`[]`),
	}
	m2 := model.MemoryEntry{
		BaseModel: model.BaseModel{ID: 2}, WorkspaceID: 1, AgentID: &agentID,
		MemoryType: model.MemoryMediumTerm, Scope: model.ScopeAgent,
		Content:  "无关闲聊，没有工作项引用",
		Metadata: json.RawMessage(`{}`), Tags: json.RawMessage(`[]`),
	}
	require.NoError(t, db.Create(&m1).Error)
	require.NoError(t, db.Create(&m2).Error)
	require.NoError(t, db.Model(&m1).Update("created_at", old).Error)
	require.NoError(t, db.Model(&m2).Update("created_at", old).Error)

	s := &AgentService{db: db}
	filled, expired, err := s.BackfillAgentMemoryIssueIDs(1)
	require.NoError(t, err)
	assert.Equal(t, 1, filled)
	assert.Equal(t, 1, expired)

	var got1, got2 model.MemoryEntry
	require.NoError(t, db.First(&got1, 1).Error)
	require.NoError(t, db.First(&got2, 2).Error)
	require.NotNil(t, got1.IssueID)
	assert.EqualValues(t, 81, *got1.IssueID)
	assert.Nil(t, got2.IssueID)
	require.NotNil(t, got2.ExpiresAt)
}
