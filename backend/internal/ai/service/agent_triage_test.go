package service

import (
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFindTriageAgent_PrefersNamedTriage(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:triage_agent?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Agent{}, &model.Issue{}))

	capsAnalyze, _ := json.Marshal([]string{"analyze"})
	capsSuggest, _ := json.Marshal([]string{"suggest_issue_changes"})
	emptyJSON := json.RawMessage(`[]`)
	require.NoError(t, db.Create(&model.Agent{
		BaseModel: model.BaseModel{ID: 1}, WorkspaceID: 1, Name: "通用分析",
		Status: "active", Capabilities: capsAnalyze, InvocationTargets: emptyJSON,
	}).Error)
	require.NoError(t, db.Create(&model.Agent{
		BaseModel: model.BaseModel{ID: 2}, WorkspaceID: 1, Name: "请求分诊",
		Status: "active", Capabilities: capsSuggest, InvocationTargets: emptyJSON,
	}).Error)
	require.NoError(t, db.Create(&model.Agent{
		BaseModel: model.BaseModel{ID: 3}, WorkspaceID: 2, Name: "请求分诊",
		Status: "active", Capabilities: capsSuggest, InvocationTargets: emptyJSON,
	}).Error)

	s := &AgentService{db: db}
	got, err := s.findTriageAgent(1)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), got.ID)
	assert.Equal(t, "请求分诊", got.Name)
}

func TestFindTriageAgent_FallsBackToSuggestCapability(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:triage_fallback?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Agent{}))

	caps, _ := json.Marshal([]string{"suggest_issue_changes"})
	require.NoError(t, db.Create(&model.Agent{
		BaseModel: model.BaseModel{ID: 9}, WorkspaceID: 1, Name: "自定义分诊",
		Status: "active", Capabilities: caps, InvocationTargets: json.RawMessage(`[]`),
	}).Error)

	s := &AgentService{db: db}
	got, err := s.findTriageAgent(1)
	require.NoError(t, err)
	assert.Equal(t, "自定义分诊", got.Name)
}
