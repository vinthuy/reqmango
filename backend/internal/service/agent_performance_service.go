package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// AgentPerformanceService aggregates agent task execution metrics for the
// performance analytics dashboard (PRD P4-010: 执行效率统计、成功率分析、报表).
type AgentPerformanceService struct {
	db *gorm.DB
}

// NewAgentPerformanceService creates a new AgentPerformanceService.
func NewAgentPerformanceService(db *gorm.DB) *AgentPerformanceService {
	return &AgentPerformanceService{db: db}
}

// parsePeriodRange resolves the query window from the from/to query params.
// When both are empty, the function returns nil,nil to signal "no filter".
func parsePeriodRange(from, to string) (*time.Time, *time.Time, error) {
	var fromT, toT *time.Time
	if from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return nil, nil, common.BadRequest("Invalid 'from' datetime format (expect RFC3339)")
		}
		fromT = &t
	}
	if to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return nil, nil, common.BadRequest("Invalid 'to' datetime format (expect RFC3339)")
		}
		toT = &t
	}
	if fromT != nil && toT != nil && fromT.After(*toT) {
		return nil, nil, common.BadRequest("'from' must not be later than 'to'")
	}
	return fromT, toT, nil
}

// applyPeriodFilter adds created_at bounds to a query when period is provided.
func applyPeriodFilter(q *gorm.DB, fromT, toT *time.Time) *gorm.DB {
	if fromT != nil {
		q = q.Where("created_at >= ?", *fromT)
	}
	if toT != nil {
		q = q.Where("created_at <= ?", *toT)
	}
	return q
}

// Overview returns workspace-wide aggregated metrics.
//
// Period bounds (RFC3339) are optional. When omitted, the metrics cover the
// entire task history of the workspace.
func (s *AgentPerformanceService) Overview(workspaceID uint64, from, to string) (*response.AgentPerformanceOverview, error) {
	fromT, toT, err := parsePeriodRange(from, to)
	if err != nil {
		return nil, err
	}

	// Status counts in a single scan.
	type statusCount struct {
		Status string
		Count  int64
	}
	var counts []statusCount
	q := s.db.Model(&model.AgentTask{}).Where("workspace_id = ?", workspaceID)
	q = applyPeriodFilter(q, fromT, toT)
	if err := q.Select("status, COUNT(*) as count").Group("status").Scan(&counts).Error; err != nil {
		return nil, common.Internal("Failed to aggregate agent task stats")
	}

	overview := &response.AgentPerformanceOverview{
		WorkspaceID: workspaceID,
		PeriodStart: fromT,
		PeriodEnd:   toT,
	}
	var completedCount, failedCount int64
	for _, c := range counts {
		overview.TotalTasks += c.Count
		switch c.Status {
		case "completed":
			overview.CompletedTasks = c.Count
			completedCount = c.Count
		case "failed":
			overview.FailedTasks = c.Count
			failedCount = c.Count
		case "cancelled":
			overview.CancelledTasks = c.Count
		case "running":
			overview.RunningTasks = c.Count
		case "enqueue", "claimed":
			overview.PendingTasks += c.Count
		}
	}

	// Success rate: completed / (completed + failed). Guard divide-by-zero.
	if denom := completedCount + failedCount; denom > 0 {
		overview.SuccessRate = round2(float64(completedCount) / float64(denom) * 100.0)
	}

	// Average duration of completed tasks in seconds.
	type avgRow struct {
		AvgSec   float64
		TotalSec float64
	}
	var avg avgRow
	durQ := s.db.Model(&model.AgentTask{}).
		Where("workspace_id = ? AND status = ? AND started_at IS NOT NULL AND completed_at IS NOT NULL", workspaceID, "completed")
	durQ = applyPeriodFilter(durQ, fromT, toT)
	if err := durQ.Select(`
			COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - started_at))), 0) AS avg_sec,
			COALESCE(SUM(EXTRACT(EPOCH FROM (completed_at - started_at))), 0) AS total_sec
		`).Scan(&avg).Error; err != nil {
		return nil, common.Internal("Failed to aggregate task durations")
	}
	overview.AvgDurationSeconds = round2(avg.AvgSec)
	overview.TotalDurationSec = round2(avg.TotalSec)

	return overview, nil
}

// ByTemplate returns per-template performance breakdown.
// Tasks without a template are aggregated under template_name "(unassigned)".
func (s *AgentPerformanceService) ByTemplate(workspaceID uint64, from, to string) ([]response.TemplatePerformance, error) {
	fromT, toT, err := parsePeriodRange(from, to)
	if err != nil {
		return nil, err
	}

	type row struct {
		AgentTemplateID *uint64
		TemplateName    string
		TaskType        string
		TotalTasks      int64
		CompletedTasks  int64
		FailedTasks     int64
		CancelledTasks  int64
		AvgDurationSec  float64
		LastRunAt       *time.Time
	}

	q := s.db.Model(&model.AgentTask{}).Table("agent_tasks AS t").
		Select(`
			t.agent_template_id AS agent_template_id,
			COALESCE(tpl.name, '(unassigned)') AS template_name,
			COALESCE(t.task_type, '') AS task_type,
			COUNT(*) AS total_tasks,
			COUNT(CASE WHEN t.status = 'completed' THEN 1 END) AS completed_tasks,
			COUNT(CASE WHEN t.status = 'failed' THEN 1 END) AS failed_tasks,
			COUNT(CASE WHEN t.status = 'cancelled' THEN 1 END) AS cancelled_tasks,
			COALESCE(AVG(CASE WHEN t.status = 'completed' AND t.started_at IS NOT NULL AND t.completed_at IS NOT NULL
				THEN EXTRACT(EPOCH FROM (t.completed_at - t.started_at)) END), 0) AS avg_duration_sec,
			MAX(t.completed_at) AS last_run_at
		`).
		Joins("LEFT JOIN agent_templates tpl ON tpl.id = t.agent_template_id").
		Where("t.workspace_id = ?", workspaceID).
		Group("t.agent_template_id, tpl.name, t.task_type").
		Order("total_tasks DESC")
	q = applyPeriodFilter(q, fromT, toT)

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, common.Internal("Failed to aggregate per-template performance")
	}

	out := make([]response.TemplatePerformance, 0, len(rows))
	for _, r := range rows {
		tp := response.TemplatePerformance{
			AgentTemplateID:    r.AgentTemplateID,
			TemplateName:       r.TemplateName,
			TaskType:           r.TaskType,
			TotalTasks:         r.TotalTasks,
			CompletedTasks:     r.CompletedTasks,
			FailedTasks:        r.FailedTasks,
			CancelledTasks:     r.CancelledTasks,
			AvgDurationSeconds: round2(r.AvgDurationSec),
			LastRunAt:          r.LastRunAt,
		}
		if denom := r.CompletedTasks + r.FailedTasks; denom > 0 {
			tp.SuccessRate = round2(float64(r.CompletedTasks) / float64(denom) * 100.0)
		}
		out = append(out, tp)
	}
	return out, nil
}

// Timeline returns daily aggregated metrics for trend charts.
//
// bucket must be one of: "day" (default), "week", "month".
func (s *AgentPerformanceService) Timeline(workspaceID uint64, bucket, from, to string) ([]response.TimelinePoint, error) {
	bucket = normalizeBucket(bucket)
	fromT, toT, err := parsePeriodRange(from, to)
	if err != nil {
		return nil, err
	}
	// Default window: last 30 days when neither bound supplied.
	if fromT == nil && toT == nil {
		end := time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour)
		start := end.AddDate(0, 0, -30)
		fromT = &start
		toT = &end
	}

	type row struct {
		BucketStart     time.Time
		TotalTasks      int64
		CompletedTasks  int64
		FailedTasks     int64
		AvgDurationSec  float64
	}

	q := s.db.Model(&model.AgentTask{}).Table("agent_tasks AS t").
		Select(fmt.Sprintf(`
			date_trunc('%s', t.created_at) AS bucket_start,
			COUNT(*) AS total_tasks,
			COUNT(CASE WHEN t.status = 'completed' THEN 1 END) AS completed_tasks,
			COUNT(CASE WHEN t.status = 'failed' THEN 1 END) AS failed_tasks,
			COALESCE(AVG(CASE WHEN t.status = 'completed' AND t.started_at IS NOT NULL AND t.completed_at IS NOT NULL
				THEN EXTRACT(EPOCH FROM (t.completed_at - t.started_at)) END), 0) AS avg_duration_sec
		`, bucket)).
		Where("t.workspace_id = ?", workspaceID).
		Group("bucket_start").
		Order("bucket_start ASC")
	q = applyPeriodFilter(q, fromT, toT)

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, common.Internal("Failed to aggregate timeline metrics")
	}

	points := make([]response.TimelinePoint, 0, len(rows))
	for _, r := range rows {
		p := response.TimelinePoint{
			BucketStart:        r.BucketStart,
			BucketEnd:          bucketEnd(r.BucketStart, bucket),
			TotalTasks:         r.TotalTasks,
			CompletedTasks:     r.CompletedTasks,
			FailedTasks:        r.FailedTasks,
			AvgDurationSeconds: round2(r.AvgDurationSec),
		}
		if denom := r.CompletedTasks + r.FailedTasks; denom > 0 {
			p.SuccessRate = round2(float64(r.CompletedTasks) / float64(denom) * 100.0)
		}
		points = append(points, p)
	}
	return points, nil
}

// FailureBreakdown returns counts of each failure_reason across failed tasks.
func (s *AgentPerformanceService) FailureBreakdown(workspaceID uint64, from, to string) ([]response.FailureBreakdown, error) {
	fromT, toT, err := parsePeriodRange(from, to)
	if err != nil {
		return nil, err
	}

	type row struct {
		Reason string
		Count  int64
	}
	q := s.db.Model(&model.AgentTask{}).
		Where("workspace_id = ? AND status = ?", workspaceID, "failed").
		Select(fmt.Sprintf(`
			COALESCE(NULLIF(failure_reason, ''), '%s') AS reason,
			COUNT(*) AS count
		`, string(model.FailureReasonUnknown))).
		Group("reason").
		Order("count DESC")
	q = applyPeriodFilter(q, fromT, toT)

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, common.Internal("Failed to aggregate failure reasons")
	}

	var total int64
	for _, r := range rows {
		total += r.Count
	}
	out := make([]response.FailureBreakdown, 0, len(rows))
	for _, r := range rows {
		pct := 0.0
		if total > 0 {
			pct = round2(float64(r.Count) / float64(total) * 100.0)
		}
		out = append(out, response.FailureBreakdown{
			FailureReason: r.Reason,
			Count:         r.Count,
			Percentage:    pct,
		})
	}
	return out, nil
}

// GetMonitorWorkspaceExists guards against queries on unknown workspaces.
// Returns gorm.ErrRecordNotFound if the workspace does not exist.
func (s *AgentPerformanceService) GetMonitorWorkspaceExists(id uint64) error {
	var ws model.Workspace
	err := s.db.Select("id").First(&ws, id).Error
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NotFound("Workspace not found")
	}
	return common.Internal("Database error")
}

// round2 rounds a float to 2 decimal places.
func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100.0
}

// normalizeBucket returns a valid PostgreSQL date_trunc unit.
func normalizeBucket(bucket string) string {
	switch bucket {
	case "week":
		return "week"
	case "month":
		return "month"
	default:
		return "day"
	}
}

// bucketEnd returns the exclusive end of a bucket for client rendering.
func bucketEnd(start time.Time, bucket string) time.Time {
	switch bucket {
	case "week":
		return start.AddDate(0, 0, 7)
	case "month":
		return start.AddDate(0, 1, 0)
	default:
		return start.AddDate(0, 0, 1)
	}
}
