package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeMemories records the filters it is asked for and answers from canned lists.
type fakeMemories struct {
	byIssue  []*model.MemoryEntry
	semantic []*model.MemoryEntry
	byAgent  []*model.MemoryEntry
	queries  []map[string]interface{}
}

func (f *fakeMemories) CreateMemory(_ context.Context, e *model.MemoryEntry) (*model.MemoryEntry, error) {
	return e, nil
}

func (f *fakeMemories) ListMemories(_ context.Context, _ uint64, filters interface{}) ([]*model.MemoryEntry, error) {
	m := filters.(map[string]interface{})
	f.queries = append(f.queries, m)
	if _, ok := m["issue_id"]; ok {
		return f.byIssue, nil
	}
	return f.byAgent, nil
}

func (f *fakeMemories) SemanticSearchByText(context.Context, uint64, string, int) ([]*model.MemoryEntry, error) {
	if f.semantic == nil {
		return nil, errors.New("no embeddings")
	}
	return f.semantic, nil
}

func mem(content string) *model.MemoryEntry { return &model.MemoryEntry{Content: content} }

func TestRetrieveAgentMemories_PrefersTheCurrentIssue(t *testing.T) {
	f := &fakeMemories{byIssue: []*model.MemoryEntry{mem("同一工作项")}, semantic: []*model.MemoryEntry{mem("别的工作项")}}
	s := &AgentService{memSvc: f}

	got, err := s.retrieveAgentMemories(context.Background(), &model.Agent{BaseModel: model.BaseModel{ID: 7}},
		&AIContext{WorkspaceID: 1, ProjectID: 1, IssueID: 81}, "分诊")
	require.NoError(t, err)

	require.Len(t, got, 1)
	assert.Equal(t, "同一工作项", got[0].Content)
	require.NotEmpty(t, f.queries)
	assert.EqualValues(t, 81, f.queries[0]["issue_id"])
	assert.EqualValues(t, 7, f.queries[0]["agent_id"])
}

func TestRetrieveAgentMemories_FallsBackWhenTheIssueHasNoHistory(t *testing.T) {
	f := &fakeMemories{semantic: []*model.MemoryEntry{mem("相似任务")}}
	s := &AgentService{memSvc: f}

	got, err := s.retrieveAgentMemories(context.Background(), &model.Agent{BaseModel: model.BaseModel{ID: 7}},
		&AIContext{WorkspaceID: 1, ProjectID: 1, IssueID: 81}, "分诊")
	require.NoError(t, err)

	require.Len(t, got, 1)
	assert.Equal(t, "相似任务", got[0].Content)
}

func TestRetrieveAgentMemories_WithoutAnIssueSkipsTheIssueLookup(t *testing.T) {
	f := &fakeMemories{byAgent: []*model.MemoryEntry{mem("代理历史")}}
	s := &AgentService{memSvc: f}

	got, err := s.retrieveAgentMemories(context.Background(), &model.Agent{BaseModel: model.BaseModel{ID: 9}},
		&AIContext{WorkspaceID: 1, ProjectID: 1}, "sprint 总结")
	require.NoError(t, err)

	require.Len(t, got, 1)
	for _, q := range f.queries {
		assert.NotContains(t, q, "issue_id")
	}
}

func TestNewAgentTaskMemory_RecordsTheIssue(t *testing.T) {
	agent := &model.Agent{BaseModel: model.BaseModel{ID: 7}, Name: "请求分诊"}

	withIssue := newAgentTaskMemory(agent, &AIContext{WorkspaceID: 1, ProjectID: 1, IssueID: 81}, "分诊", "结论", nil)
	require.NotNil(t, withIssue.IssueID)
	assert.EqualValues(t, 81, *withIssue.IssueID)
	assert.EqualValues(t, 7, *withIssue.AgentID)

	withoutIssue := newAgentTaskMemory(agent, &AIContext{WorkspaceID: 1, ProjectID: 1}, "总结", "结论", nil)
	assert.Nil(t, withoutIssue.IssueID)
}

func TestAgentMemoryBlock_WarnsThatHistoryMayBeStale(t *testing.T) {
	block := agentMemoryBlock([]*model.MemoryEntry{{ContextName: "请求分诊", Content: "上次建议流转到待办"}})
	assert.Contains(t, block, "上次建议流转到待办")
	assert.Contains(t, block, "以工具查询到的当前数据为准")
	assert.Empty(t, agentMemoryBlock(nil))
	assert.False(t, strings.Contains(agentMemoryBlock([]*model.MemoryEntry{{Content: strings.Repeat("长", 400)}}), strings.Repeat("长", 200)),
		"each memory is truncated")
}
