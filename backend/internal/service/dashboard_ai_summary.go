package service

import (
	"context"
	"encoding/json"
	"fmt"

	aiservice "github.com/reqmango/backend/internal/ai/service"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// AISummaryProvider supplies project-level AI analysis for dashboard widgets.
type AISummaryProvider interface {
	ProjectAISummary(ctx context.Context, projectID uint64) (map[string]interface{}, error)
}

// SetAISummaryProvider wires the AI analyzer used by ai_summary widgets.
func (s *DashboardService) SetAISummaryProvider(p AISummaryProvider) {
	s.aiSummary = p
}

// AIServiceSummaryProvider adapts aiservice.AIService for dashboard widgets.
type AIServiceSummaryProvider struct {
	AI *aiservice.AIService
	DB *gorm.DB
}

func (p *AIServiceSummaryProvider) ProjectAISummary(ctx context.Context, projectID uint64) (map[string]interface{}, error) {
	if p == nil || p.AI == nil || p.DB == nil {
		return nil, fmt.Errorf("AI summary provider not configured")
	}
	var project model.Project
	if err := p.DB.First(&project, projectID).Error; err != nil {
		return nil, fmt.Errorf("load project: %w", err)
	}
	actx := &aiservice.AIContext{
		ProjectID:         project.ID,
		WorkspaceID:       project.WorkspaceID,
		ProjectName:       project.Name,
		ProjectIdentifier: project.Identifier,
	}
	result, err := p.AI.Analyze(ctx, actx)
	if err != nil {
		return nil, err
	}
	bottlenecks := make([]map[string]interface{}, 0, len(result.Bottlenecks))
	for _, b := range result.Bottlenecks {
		bottlenecks = append(bottlenecks, map[string]interface{}{
			"issue_id":      b.IssueID,
			"issue_name":    b.IssueName,
			"days_in_state": b.DaysInState,
			"state_name":    b.StateName,
		})
	}
	out := map[string]interface{}{
		"summary":     result.Summary,
		"insights":    result.Insights,
		"bottlenecks": bottlenecks,
		"stats":       result.Stats,
		"scope":       "project",
	}
	return out, nil
}

func (s *DashboardService) renderAISummary(projectID uint64, w *model.DashboardWidget) (json.RawMessage, error) {
	if s.aiSummary == nil {
		data, _ := json.Marshal(map[string]interface{}{
			"error": "AI summary unavailable",
		})
		return json.RawMessage(data), nil
	}
	payload, err := s.aiSummary.ProjectAISummary(context.Background(), projectID)
	if err != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"error": err.Error(),
		})
		return json.RawMessage(data), nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}
