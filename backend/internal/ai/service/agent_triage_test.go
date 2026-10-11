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

func TestDispatchActivityAction_UsesTriggeredBy(t *testing.T) {
	assert.Equal(t, "dispatch", dispatchActivityAction(nil))
	assert.Equal(t, "dispatch", dispatchActivityAction(&DispatchContext{}))
	assert.Equal(t, "auto_triage", dispatchActivityAction(&DispatchContext{TriggeredBy: "auto_triage"}))
	assert.Equal(t, "mention", dispatchActivityAction(&DispatchContext{TriggeredBy: "  mention  "}))
}

func TestFindAgentByCapability_StaysInTheWorkspace(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Agent{}))

	caps, _ := json.Marshal([]string{"analyze"})
	for _, a := range []model.Agent{
		{BaseModel: model.BaseModel{ID: 1}, WorkspaceID: 2, Name: "别人的分析"},
		{BaseModel: model.BaseModel{ID: 2}, WorkspaceID: 1, Name: "自己的分析"},
	} {
		a.Status, a.Capabilities, a.InvocationTargets = "active", caps, json.RawMessage(`[]`)
		require.NoError(t, db.Create(&a).Error)
	}

	s := &AgentService{db: db}
	got, err := s.findAgentByCapability(1, "analyze")
	require.NoError(t, err)
	assert.Equal(t, "自己的分析", got.Name)

	_, err = s.findAgentByCapability(3, "analyze")
	assert.Error(t, err, "no agent in the workspace means no dispatch, not a foreign agent")
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
