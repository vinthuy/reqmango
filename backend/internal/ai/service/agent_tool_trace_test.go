package service

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestToolTrace_KeepsFailuresAndBoundsSize(t *testing.T) {
	var tr toolTrace
	tr.add("list_states", json.RawMessage(`{"project_id":3015}`), nil, 12*time.Millisecond)
	tr.add("get_issue", json.RawMessage(`{"issue_id":9}`), errors.New("issue_id=9 超出当前范围"), time.Millisecond)
	tr.add("search_issues", json.RawMessage(`{"query":"`+strings.Repeat("长", 2000)+`"}`), nil, 0)
	for i := 0; i < maxTracedToolCalls+5; i++ {
		tr.add("list_labels", json.RawMessage(`{}`), nil, 0)
	}

	var calls []model.AgentToolCall
	require.NoError(t, json.Unmarshal(tr.JSON(), &calls))
	require.Len(t, calls, maxTracedToolCalls)
	assert.Equal(t, "list_states", calls[0].Name)
	assert.JSONEq(t, `{"project_id":3015}`, string(calls[0].Input))
	assert.EqualValues(t, 12, calls[0].DurationMs)
	assert.Equal(t, "issue_id=9 超出当前范围", calls[1].Error, "rejected calls are the ones worth auditing")
	assert.LessOrEqual(t, len(calls[2].Input), maxTracedInputBytes+64, "huge arguments are truncated")
	assert.True(t, json.Valid(calls[2].Input), "a truncated argument is still valid JSON")
}

func TestRecordActivity_StoresToolCalls(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentActivity{}))

	var tr toolTrace
	tr.add("list_states", json.RawMessage(`{}`), nil, 0)
	s := &AgentService{db: db}
	s.recordActivity(&model.Agent{BaseModel: model.BaseModel{ID: 3}, Name: "请求分诊"}, nil, "dispatch", "完成", "task", 7, tr.JSON())

	var stored string
	require.NoError(t, db.Raw("SELECT tool_calls FROM agent_activities LIMIT 1").Scan(&stored).Error)
	assert.Contains(t, stored, "list_states")
}
