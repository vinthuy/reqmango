package service

import (
	"time"

	"gorm.io/gorm"
)

type WorkspaceAnalyticsService struct {
	db *gorm.DB
}

func NewWorkspaceAnalyticsService(db *gorm.DB) *WorkspaceAnalyticsService {
	return &WorkspaceAnalyticsService{db: db}
}

type AnalyticsSummary struct {
	Total            int64    `json:"total"`
	Open             int64    `json:"open"`
	Overdue          int64    `json:"overdue"`
	CreatedInRange   int64    `json:"created_in_range"`
	CompletedInRange int64    `json:"completed_in_range"`
	CompletionRate   float64  `json:"completion_rate"`
	AvgCycleDays     *float64 `json:"avg_cycle_days"`
	ProjectCount     int64    `json:"project_count"`
}

type AnalyticsTrendPoint struct {
	Date      string `json:"date"`
	Created   int64  `json:"created"`
	Completed int64  `json:"completed"`
}

type AnalyticsBucket struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

type AnalyticsProjectRow struct {
	ID               uint64  `json:"id"`
	Name             string  `json:"name"`
	Identifier       string  `json:"identifier"`
	Color            string  `json:"color"`
	Total            int64   `json:"total"`
	Open             int64   `json:"open"`
	Completed        int64   `json:"completed"`
	Overdue          int64   `json:"overdue"`
	CreatedInRange   int64   `json:"created_in_range"`
	CompletedInRange int64   `json:"completed_in_range"`
	CompletionRate   float64 `json:"completion_rate"`
}

type AnalyticsAssigneeRow struct {
	UserID           uint64 `json:"user_id"`
	Name             string `json:"name"`
	Open             int64  `json:"open"`
	Overdue          int64  `json:"overdue"`
	CompletedInRange int64  `json:"completed_in_range"`
}

type WorkspaceAnalytics struct {
	Days         int                    `json:"days"`
	Granularity  string                 `json:"granularity"`
	Summary      AnalyticsSummary       `json:"summary"`
	Trend        []AnalyticsTrendPoint  `json:"trend"`
	StateGroups  []AnalyticsBucket      `json:"state_groups"`
	Priorities   []AnalyticsBucket      `json:"priorities"`
	IssueTypes   []AnalyticsBucket      `json:"issue_types"`
	Projects     []AnalyticsProjectRow  `json:"projects"`
	Assignees    []AnalyticsAssigneeRow `json:"assignees"`
	GeneratedAt  time.Time              `json:"generated_at"`
}

const openStateCond = "s.\"group\" NOT IN ('completed','cancelled')"

// baseIssues selects live issues of the workspace (optionally one project), excluding drafts,
// archived items and intake requests that have not been accepted.
func (s *WorkspaceAnalyticsService) baseIssues(workspaceID uint64, projectID uint64) *gorm.DB {
	q := s.db.Table("issues i").
		Joins("JOIN states s ON s.id = i.state_id").
		Joins("JOIN projects p ON p.id = i.project_id AND p.deleted_at IS NULL AND p.archived_at IS NULL").
		Where("i.workspace_id = ? AND i.deleted_at IS NULL AND i.archived_at IS NULL AND i.is_draft = false", workspaceID).
		Where("i.intake_status IS NULL OR i.intake_status = 'accepted'")
	if projectID > 0 {
		q = q.Where("i.project_id = ?", projectID)
	}
	return q
}

func (s *WorkspaceAnalyticsService) Get(workspaceID, projectID uint64, days int) (*WorkspaceAnalytics, error) {
	now := time.Now()
	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	since := today.AddDate(0, 0, -(days - 1))

	out := &WorkspaceAnalytics{Days: days, GeneratedAt: now}

	var sum struct {
		Total            int64
		Open             int64
		Overdue          int64
		CreatedInRange   int64
		CompletedInRange int64
		AvgCycleDays     *float64
		ProjectCount     int64
	}
	err := s.baseIssues(workspaceID, projectID).Select(`
		COUNT(*) AS total,
		COUNT(*) FILTER (WHERE `+openStateCond+`) AS open,
		COUNT(*) FILTER (WHERE `+openStateCond+` AND i.target_date IS NOT NULL AND i.target_date < ?) AS overdue,
		COUNT(*) FILTER (WHERE i.created_at >= ?) AS created_in_range,
		COUNT(*) FILTER (WHERE s."group" = 'completed' AND i.completed_at >= ?) AS completed_in_range,
		AVG(EXTRACT(EPOCH FROM (i.completed_at - i.created_at)) / 86400.0)
			FILTER (WHERE s."group" = 'completed' AND i.completed_at >= ? AND i.completed_at >= i.created_at) AS avg_cycle_days,
		COUNT(DISTINCT i.project_id) AS project_count`, today, since, since, since).
		Scan(&sum).Error
	if err != nil {
		return nil, err
	}
	out.Summary = AnalyticsSummary{
		Total: sum.Total, Open: sum.Open, Overdue: sum.Overdue,
		CreatedInRange: sum.CreatedInRange, CompletedInRange: sum.CompletedInRange,
		AvgCycleDays: sum.AvgCycleDays, ProjectCount: sum.ProjectCount,
	}
	if sum.Total > 0 {
		out.Summary.CompletionRate = float64(sum.Total-sum.Open) / float64(sum.Total)
	}

	if err := s.fillTrend(out, workspaceID, projectID, since, today, days); err != nil {
		return nil, err
	}

	if err := s.baseIssues(workspaceID, projectID).
		Select(`s."group" AS key, COUNT(*) AS count`).
		Group(`s."group"`).Order("count DESC").Scan(&out.StateGroups).Error; err != nil {
		return nil, err
	}
	if err := s.baseIssues(workspaceID, projectID).Where(openStateCond).
		Select(`COALESCE(NULLIF(i.priority, ''), 'none') AS key, COUNT(*) AS count`).
		Group("key").Order("count DESC").Scan(&out.Priorities).Error; err != nil {
		return nil, err
	}
	if err := s.baseIssues(workspaceID, projectID).Where(openStateCond).
		Joins("LEFT JOIN issue_types it ON it.id = i.issue_type_id").
		Select(`COALESCE(it.name, '') AS key, COUNT(*) AS count`).
		Group("key").Order("count DESC").Scan(&out.IssueTypes).Error; err != nil {
		return nil, err
	}

	if err := s.baseIssues(workspaceID, projectID).Select(`
		p.id, p.name, p.identifier, COALESCE(p.color, '') AS color,
		COUNT(*) AS total,
		COUNT(*) FILTER (WHERE `+openStateCond+`) AS open,
		COUNT(*) FILTER (WHERE s."group" = 'completed') AS completed,
		COUNT(*) FILTER (WHERE `+openStateCond+` AND i.target_date IS NOT NULL AND i.target_date < ?) AS overdue,
		COUNT(*) FILTER (WHERE i.created_at >= ?) AS created_in_range,
		COUNT(*) FILTER (WHERE s."group" = 'completed' AND i.completed_at >= ?) AS completed_in_range`,
		today, since, since).
		Group("p.id, p.name, p.identifier, p.color").Order("open DESC, total DESC").
		Scan(&out.Projects).Error; err != nil {
		return nil, err
	}
	for i := range out.Projects {
		if out.Projects[i].Total > 0 {
			out.Projects[i].CompletionRate = float64(out.Projects[i].Total-out.Projects[i].Open) / float64(out.Projects[i].Total)
		}
	}

	if err := s.baseIssues(workspaceID, projectID).
		Joins("JOIN issue_assignees ia ON ia.issue_id = i.id").
		Joins("JOIN users u ON u.id = ia.user_id").
		Select(`u.id AS user_id,
			COALESCE(NULLIF(u.display_name, ''), NULLIF(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''), u.email) AS name,
			COUNT(*) FILTER (WHERE `+openStateCond+`) AS open,
			COUNT(*) FILTER (WHERE `+openStateCond+` AND i.target_date IS NOT NULL AND i.target_date < ?) AS overdue,
			COUNT(*) FILTER (WHERE s."group" = 'completed' AND i.completed_at >= ?) AS completed_in_range`, today, since).
		Group("u.id, u.display_name, u.first_name, u.last_name, u.email").
		Having(`COUNT(*) FILTER (WHERE `+openStateCond+`) > 0 OR COUNT(*) FILTER (WHERE s."group" = 'completed' AND i.completed_at >= ?) > 0`, since).
		Order("open DESC, completed_in_range DESC").Limit(10).
		Scan(&out.Assignees).Error; err != nil {
		return nil, err
	}

	ensure := func(b []AnalyticsBucket) []AnalyticsBucket {
		if b == nil {
			return []AnalyticsBucket{}
		}
		return b
	}
	out.StateGroups, out.Priorities, out.IssueTypes = ensure(out.StateGroups), ensure(out.Priorities), ensure(out.IssueTypes)
	if out.Projects == nil {
		out.Projects = []AnalyticsProjectRow{}
	}
	if out.Assignees == nil {
		out.Assignees = []AnalyticsAssigneeRow{}
	}
	return out, nil
}

// fillTrend buckets created/completed counts per day (<= 31 days) or per ISO week.
func (s *WorkspaceAnalyticsService) fillTrend(out *WorkspaceAnalytics, workspaceID, projectID uint64, since, today time.Time, days int) error {
	unit := "day"
	if days > 31 {
		unit = "week"
	}
	out.Granularity = unit

	type row struct {
		Bucket time.Time
		Count  int64
	}
	// Bucket on the server's local wall-clock time so keys line up with the Go-side calendar.
	_, offset := today.Zone()
	localExpr := func(col string) string {
		return "date_trunc('" + unit + "', (" + col + " AT TIME ZONE 'UTC') + (? * interval '1 second'))"
	}
	var created, completed []row
	if err := s.baseIssues(workspaceID, projectID).
		Where("i.created_at >= ?", since).
		Select(localExpr("i.created_at")+" AS bucket, COUNT(*) AS count", offset).
		Group("bucket").Scan(&created).Error; err != nil {
		return err
	}
	if err := s.baseIssues(workspaceID, projectID).
		Where(`s."group" = 'completed' AND i.completed_at >= ?`, since).
		Select(localExpr("i.completed_at")+" AS bucket, COUNT(*) AS count", offset).
		Group("bucket").Scan(&completed).Error; err != nil {
		return err
	}

	bucketKey := func(t time.Time) string { return t.UTC().Format("2006-01-02") }
	key := func(t time.Time) string { return t.Format("2006-01-02") }
	idx := map[string]int{}
	start := since
	if unit == "week" {
		wd := (int(start.Weekday()) + 6) % 7
		start = start.AddDate(0, 0, -wd)
	}
	for d := start; !d.After(today); {
		idx[key(d)] = len(out.Trend)
		out.Trend = append(out.Trend, AnalyticsTrendPoint{Date: key(d)})
		if unit == "week" {
			d = d.AddDate(0, 0, 7)
		} else {
			d = d.AddDate(0, 0, 1)
		}
	}
	for _, r := range created {
		if i, ok := idx[bucketKey(r.Bucket)]; ok {
			out.Trend[i].Created = r.Count
		}
	}
	for _, r := range completed {
		if i, ok := idx[bucketKey(r.Bucket)]; ok {
			out.Trend[i].Completed = r.Count
		}
	}
	return nil
}
