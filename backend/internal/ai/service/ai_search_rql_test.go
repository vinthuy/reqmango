package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/ai/llm"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestValidateAISearchRQL(t *testing.T) {
	valid := []string{
		`priority IN ("urgent", "high")`,
		`state_group = "started"`,
		`name LIKE "%登录%"`,
		`assignee_id = $CURRENT_USER AND state_group IN ("unstarted", "started")`,
		`(priority = "high" OR priority = "urgent") AND target_date < $TODAY`,
		`cycle_id IS NULL`,
		`assignee_id = $CURRENT_USER AND state_group IN (backlog, unstarted, started)`,
	}
	for _, q := range valid {
		assert.NoError(t, validateAISearchRQL(q), q)
	}

	invalid := []string{
		"",
		`description LIKE "%x%"`,
		`NOT priority = "high"`,
		`priority = "high" foo`,
		`priority ==`,
		`哈哈`,
	}
	for _, q := range invalid {
		assert.Error(t, validateAISearchRQL(q), q)
	}
}

// fakeSearchLLM answers the first (translation) request with rqlJSON and every
// later request with a plain reply.
func fakeSearchLLM(t *testing.T, rqlJSON string) (*llm.LLMClient, *int32) {
	t.Helper()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		content := "没有找到"
		if n == 1 {
			content = rqlJSON
		}
		_, _ = fmt.Fprintf(w, `{"choices":[{"finish_reason":"stop","message":{"content":%s}}]}`, strconv.Quote(content))
	}))
	t.Cleanup(srv.Close)
	return llm.NewLLMClient("k-live", "deepseek-chat", srv.URL, "deepseek"), &calls
}

func searchFixtureDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Project{}, &model.State{}, &model.Issue{}, &model.Label{},
		&model.IssueType{}, &model.IssueTypeImport{}, &model.ProjectMember{}, &model.User{}))
	require.NoError(t, db.Create(&model.Project{
		BaseModel: model.BaseModel{ID: 1}, Name: "P", Identifier: "CORE", WorkspaceID: 1,
	}).Error)
	for i, p := range []string{"high", "low"} {
		require.NoError(t, db.Create(&model.Issue{
			Name: "任务" + p, Priority: p, ProjectID: 1, WorkspaceID: 1, SequenceID: i + 1,
		}).Error)
	}
	return db
}

func TestSearch_ReturnsGeneratedFilterWhenItMatches(t *testing.T) {
	client, calls := fakeSearchLLM(t, `{"rql": "priority = \"high\"", "explanation": "高优先级"}`)
	s := NewAIService(searchFixtureDB(t), client)

	res, err := s.Search(context.Background(), &AISearchRequest{Query: "高优先级的"}, &AIContext{ProjectID: 1, WorkspaceID: 1, UserID: 1})

	require.NoError(t, err)
	assert.Equal(t, `priority = "high"`, res.RQL)
	assert.Equal(t, "高优先级", res.Explanation)
	require.Len(t, res.Issues, 1)
	assert.Equal(t, "任务high", res.Issues[0]["name"])
	assert.EqualValues(t, 1, *calls, "no tool loop when the filter matched")
}

func TestSearch_FallsBackWhenGeneratedFilterIsInvalid(t *testing.T) {
	client, calls := fakeSearchLLM(t, `{"rql": "description LIKE \"%x%\"", "explanation": "x"}`)
	s := NewAIService(searchFixtureDB(t), client)

	res, err := s.Search(context.Background(), &AISearchRequest{Query: "描述里有 x"}, &AIContext{ProjectID: 1, WorkspaceID: 1, UserID: 1})

	require.NoError(t, err)
	assert.Empty(t, res.RQL, "an unsafe field is never handed to the FilterBar")
	assert.Equal(t, "没有找到", res.Explanation)
	assert.EqualValues(t, 2, *calls)
}

func TestSearch_KeepsValidFilterWhenNothingMatches(t *testing.T) {
	client, _ := fakeSearchLLM(t, `{"rql": "priority = \"urgent\"", "explanation": "紧急"}`)
	s := NewAIService(searchFixtureDB(t), client)

	res, err := s.Search(context.Background(), &AISearchRequest{Query: "紧急的"}, &AIContext{ProjectID: 1, WorkspaceID: 1, UserID: 1})

	require.NoError(t, err)
	assert.Equal(t, `priority = "urgent"`, res.RQL)
	assert.Empty(t, res.Issues)
}
