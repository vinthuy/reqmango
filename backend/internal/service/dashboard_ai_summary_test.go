package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAISummary struct {
	summary string
	err     error
}

func (s *stubAISummary) ProjectAISummary(ctx context.Context, projectID uint64) (map[string]interface{}, error) {
	if s.err != nil {
		return nil, s.err
	}
	return map[string]interface{}{
		"summary":  s.summary,
		"insights": []string{"insight-a"},
		"bottlenecks": []map[string]interface{}{
			{"issue_id": 1, "issue_name": "Stuck", "days_in_state": 5, "state_name": "In Progress"},
		},
		"scope": "project",
	}, nil
}

func TestRenderWidget_AISummary(t *testing.T) {
	svc := NewDashboardService(nil)
	svc.SetAISummaryProvider(&stubAISummary{summary: "project looks healthy"})

	w := &model.DashboardWidget{WidgetType: "ai_summary", Title: "AI Summary", Config: json.RawMessage(`{}`)}
	raw, err := svc.renderWidget(4, w, nil)
	require.NoError(t, err)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &data))
	assert.Equal(t, "project looks healthy", data["summary"])
	assert.Equal(t, "project", data["scope"])
	insights, ok := data["insights"].([]interface{})
	require.True(t, ok)
	assert.Len(t, insights, 1)
}

func TestRenderWidget_AISummary_NoProvider(t *testing.T) {
	svc := NewDashboardService(nil)
	w := &model.DashboardWidget{WidgetType: "ai_summary", Title: "AI Summary", Config: json.RawMessage(`{}`)}
	raw, err := svc.renderWidget(4, w, nil)
	require.NoError(t, err)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &data))
	assert.Contains(t, data["error"], "AI")
}
