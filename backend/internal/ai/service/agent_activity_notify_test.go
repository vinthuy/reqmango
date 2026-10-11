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

func TestRecordActivity_NotifiesTheTriggeringUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentActivity{}))

	type sent struct {
		userID uint64
		event  string
		data   map[string]interface{}
	}
	var got []sent
	s := &AgentService{db: db}
	s.SetActivityNotifier(func(userID uint64, event string, data interface{}) {
		b, _ := json.Marshal(data)
		var m map[string]interface{}
		_ = json.Unmarshal(b, &m)
		got = append(got, sent{userID, event, m})
	})

	agent := &model.Agent{BaseModel: model.BaseModel{ID: 3}, Name: "请求分诊"}
	issueID := uint64(42)
	s.recordActivity(agent, &issueID, "auto_triage", "完成", "task", 7)
	s.recordActivity(agent, nil, "dispatch", "无工作项", "task", 7)

	require.Len(t, got, 1, "only issue-scoped activity is pushed")
	assert.EqualValues(t, 7, got[0].userID)
	assert.Equal(t, "agent_activity.created", got[0].event)
	assert.EqualValues(t, 42, got[0].data["issue_id"])
	assert.Equal(t, "auto_triage", got[0].data["action"])
}
