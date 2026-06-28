package service

import (
	"github.com/reqmanpy/backend-go/internal/common"
	"gorm.io/gorm"
)

type ReportService struct{ db *gorm.DB }

func NewReportService(db *gorm.DB) *ReportService { return &ReportService{db: db} }

type ReportRequest struct {
	GroupBy  string `json:"group_by"`  // state, priority, assignee, type, label
	Metric   string `json:"metric"`    // count (future: sum_estimate, avg_cycle)
	DateFrom string `json:"date_from"` // optional: YYYY-MM-DD
	DateTo   string `json:"date_to"`   // optional: YYYY-MM-DD
}

type ReportResponse struct {
	Labels []string `json:"labels"`
	Values []int    `json:"values"`
	Total  int      `json:"total"`
	GroupBy string `json:"group_by"`
	Colors map[string]string `json:"colors,omitempty"`
}

func (s *ReportService) Generate(projectID uint64, req *ReportRequest) (*ReportResponse, error) {
	type row struct { Name string; Cnt int }
	var rows []row

	selectExpr, joinClause := s.buildAggregation(req.GroupBy)
	sql := "SELECT " + selectExpr + ", COUNT(*) as cnt FROM issues"
	if joinClause != "" { sql += " " + joinClause }
	sql += " WHERE issues.project_id = ?"
	args := []interface{}{projectID}
	if req.DateFrom != "" { sql += " AND issues.created_at >= ?"; args = append(args, req.DateFrom) }
	if req.DateTo != "" { sql += " AND issues.created_at <= ?"; args = append(args, req.DateTo+" 23:59:59") }
	sql += " GROUP BY 1 ORDER BY cnt DESC"

	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, common.Internal("Failed to generate report")
	}

	resp := &ReportResponse{GroupBy: req.GroupBy, Colors: priorityColors}
	for _, r := range rows {
		resp.Labels = append(resp.Labels, r.Name)
		resp.Values = append(resp.Values, r.Cnt)
		resp.Total += r.Cnt
	}
	if resp.Labels == nil { resp.Labels = []string{}; resp.Values = []int{} }
	return resp, nil
}

func (s *ReportService) buildAggregation(groupBy string) (selectExpr, join string) {
	switch groupBy {
	case "state":
		return "COALESCE(s.name, 'Unknown') as name", "LEFT JOIN states s ON issues.state_id = s.id"
	case "priority":
		return "COALESCE(issues.priority, 'none') as name", ""
	case "assignee":
		return "COALESCE(u.display_name, 'Unassigned') as name", "LEFT JOIN issue_assignees ia ON issues.id = ia.issue_id LEFT JOIN users u ON ia.user_id = u.id"
	case "type":
		return "COALESCE(it.name, 'Untyped') as name", "LEFT JOIN issue_types it ON issues.issue_type_id = it.id"
	case "label":
		return "COALESCE(l.name, 'No Label') as name", "LEFT JOIN issue_labels il ON issues.id = il.issue_id LEFT JOIN labels l ON il.label_id = l.id"
	default:
		return "COALESCE(s.name, 'Unknown') as name", "LEFT JOIN states s ON issues.state_id = s.id"
	}
}

var priorityColors = map[string]string{
	"urgent": "#EF4444", "high": "#F59E0B", "medium": "#3B82F6",
	"low": "#6B7280", "none": "#9CA3AF",
	"Backlog": "#6B7280", "Todo": "#3B82F6", "In Progress": "#F59E0B",
	"In Review": "#8B5CF6", "Done": "#10B981", "Cancelled": "#EF4444",
}
