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

func TestIssueTools_ReturnTheReferenceUsersSee(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Project{}, &model.Issue{}, &model.State{}, &model.IssueType{}))
	p := model.Project{Name: "p", Identifier: "TRI", WorkspaceID: 1}
	require.NoError(t, db.Create(&p).Error)
	iss := model.Issue{Name: "导出报表时提示服务器错误", SequenceID: 1, ProjectID: p.ID, WorkspaceID: 1, StateID: 1}
	require.NoError(t, db.Create(&iss).Error)

	s := &AIService{db: db}
	actx := &AIContext{WorkspaceID: 1, ProjectID: p.ID, Mode: "agent"}

	found, err := s.ExecuteTool("search_issues", json.RawMessage(`{}`), actx)
	require.NoError(t, err)
	list := found.([]map[string]interface{})
	require.Len(t, list, 1)
	assert.Equal(t, "TRI-1", list[0]["ref"], "agents cite database ids otherwise, which users cannot look up")

	raw, _ := json.Marshal(map[string]interface{}{"issue_id": iss.ID})
	got, err := s.ExecuteTool("get_issue", raw, actx)
	require.NoError(t, err)
	assert.Equal(t, "TRI-1", got.(map[string]interface{})["ref"])
}
