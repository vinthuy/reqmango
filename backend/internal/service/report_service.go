package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/rql"
	"gorm.io/gorm"
)

type ReportService struct {
	db         *gorm.DB
	rqlService *rql.RQLService
}

func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{db: db, rqlService: rql.NewRQLService()}
}

// ReportRequest 报表请求
type ReportRequest struct {
	RQL        string `json:"rql"`         // RQL 筛选条件
	ReportType string `json:"report_type"` // distribution | created_vs_resolved | avg_age | current_age | created_trend
	GroupBy    string `json:"group_by"`    // state | priority | assignee | type | label | cycle | module
	Chart      string `json:"chart"`       // bar | pie | doughnut | table
	DateFrom   string `json:"date_from"`   // YYYY-MM-DD
	DateTo     string `json:"date_to"`     // YYYY-MM-DD
	Interval   string `json:"interval"`    // day | week | month (for trend reports)
}

// ReportV2Request V2 报表请求 — 显式指定 x_axis (维度) 和 y_axis (指标)
type ReportV2Request struct {
	XAxis    string `json:"x_axis"`              // 维度: state, priority, assignee, type, label, cycle, module, created_day, created_week, created_month, completed_day, completed_week, completed_month, updated_day, updated_week, updated_month
	YAxis    string `json:"y_axis"`              // 指标: count, avg_processing_time, current_retention, created_vs_resolved
	Interval string `json:"interval,omitempty"`  // 仅用于时间轴: day, week, month (默认 week)
	RQL      string `json:"rql,omitempty"`
	DateFrom string `json:"date_from,omitempty"`
	DateTo   string `json:"date_to,omitempty"`
}

// ReportResponse 报表响应
type ReportResponse struct {
	Type    string            `json:"type"`
	Labels  []string          `json:"labels"`
	Values  []int             `json:"values"`
	Values2 []int             `json:"values2,omitempty"` // 第二组数据 (created_vs_resolved)
	Total   int               `json:"total"`
	GroupBy string            `json:"group_by"`
	Colors  map[string]string `json:"colors,omitempty"`
	Summary *ReportSummary    `json:"summary,omitempty"`
}

type ReportSummary struct {
	AvgDays   float64 `json:"avg_days,omitempty"`
	MaxDays   int     `json:"max_days,omitempty"`
	MinDays   int     `json:"min_days,omitempty"`
	MedianDay int     `json:"median_days,omitempty"`
}

// Generate 根据报表类型生成报表
func (s *ReportService) Generate(projectID uint64, req *ReportRequest) (*ReportResponse, error) {
	switch req.ReportType {
	case "created_vs_resolved":
		return s.createdVsResolved(projectID, req)
	case "avg_age":
		return s.avgAge(projectID, req)
	case "current_age":
		return s.currentAge(projectID, req)
	case "created_trend":
		return s.createdTrend(projectID, req)
	default:
		return s.distribution(projectID, req)
	}
}

// ========================
// V2: 基于 x_axis / y_axis 的统一报表
// ========================

// GenerateV2 根据 x_axis (维度) 和 y_axis (指标) 生成报表
func (s *ReportService) GenerateV2(projectID uint64, req *ReportV2Request) (*ReportResponse, error) {
	// 1. 获取 RQL 筛选后的 issue IDs
	ids, err := s.getFilteredIssueIDs(projectID, req.RQL)
	if err != nil {
		return nil, common.BadRequest("Invalid RQL: " + err.Error())
	}

	// 2. 验证 x_axis 和 y_axis
	xAxis := req.XAxis
	yAxis := req.YAxis
	if xAxis == "" {
		xAxis = "state"
	}
	if yAxis == "" {
		yAxis = "count"
	}

	// 3. created_vs_resolved 需要特殊处理：运行两个独立查询
	if yAxis == "created_vs_resolved" {
		return s.v2CreatedVsResolved(projectID, xAxis, req, ids)
	}

	// 4. 构建 X 轴表达式
	selectExpr, joinClause, timeColumn := s.resolveXAxis(xAxis, req.Interval)
	_ = timeColumn // 仅分类轴需要 groupBy 名称

	// 5. 构建 Y 轴表达式
	yExpr, extraWhere := s.resolveYAxis(yAxis)

	// 6. 组装 SQL
	sql := "SELECT " + selectExpr + ", (" + yExpr + ") AS cnt FROM issues"
	if joinClause != "" {
		sql += " " + joinClause
	}
	sql += " WHERE issues.project_id = ? AND issues.archived_at IS NULL"
	args := []interface{}{projectID}

	// 追加 y_axis 额外 WHERE 条件
	if extraWhere != "" {
		sql += " AND " + extraWhere
	}

	// 追加 RQL 过滤
	if ids != nil {
		if len(ids) == 0 {
			return &ReportResponse{Type: yAxis, GroupBy: xAxis, Labels: []string{}, Values: []int{}}, nil
		}
		placeholders := make([]string, len(ids))
		vals := make([]interface{}, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			vals[i] = id
		}
		sql += " AND issues.id IN (" + strings.Join(placeholders, ",") + ")"
		args = append(args, vals...)
	}

	// 追加日期过滤
	if req.DateFrom != "" {
		sql += " AND issues.created_at >= ?"
		args = append(args, req.DateFrom)
	}
	if req.DateTo != "" {
		sql += " AND issues.created_at <= ?"
		args = append(args, req.DateTo+" 23:59:59")
	}
	sql += " GROUP BY 1 ORDER BY 1"

	// 7. 执行查询
	type row struct {
		Name string
		Cnt  int
	}
	var rows []row
	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, common.Internal("Failed to generate v2 report: " + err.Error())
	}

	// 8. 组装响应
	resp := &ReportResponse{
		Type:    yAxis,
		GroupBy: xAxis,
		Colors:  s.getColors(xAxis),
	}
	for _, r := range rows {
		name := r.Name
		if name == "" {
			name = "N/A"
		}
		resp.Labels = append(resp.Labels, name)
		resp.Values = append(resp.Values, r.Cnt)
		resp.Total += r.Cnt
	}
	if resp.Labels == nil {
		resp.Labels = []string{}
		resp.Values = []int{}
	}
	return resp, nil
}

// resolveXAxis 根据 x_axis 值返回 (selectExpr, joinClause, rawColumnName)
func (s *ReportService) resolveXAxis(xAxis, interval string) (selectExpr, joinClause, column string) {
	switch xAxis {
	// ---- 分类轴：复用 buildAggregation ----
	case "state", "priority", "assignee", "type", "label", "cycle", "module", "state_group", "created_by":
		sel, join := s.buildAggregation(xAxis)
		return sel, join, ""

	// ---- 时间轴：created_at ----
	case "created_day":
		return fmt.Sprintf("COALESCE(TO_CHAR(issues.created_at, '%s'), 'N/A')", s.timeFormat("day", interval)), "", "created_at"
	case "created_week":
		return fmt.Sprintf("COALESCE(TO_CHAR(issues.created_at, '%s'), 'N/A')", s.timeFormat("week", interval)), "", "created_at"
	case "created_month":
		return fmt.Sprintf("COALESCE(TO_CHAR(issues.created_at, '%s'), 'N/A')", s.timeFormat("month", interval)), "", "created_at"

	// ---- 时间轴：completed_at ----
	case "completed_day":
		return fmt.Sprintf("COALESCE(TO_CHAR(issues.completed_at, '%s'), 'N/A')", s.timeFormat("day", interval)), "", "completed_at"
	case "completed_week":
		return fmt.Sprintf("COALESCE(TO_CHAR(issues.completed_at, '%s'), 'N/A')", s.timeFormat("week", interval)), "", "completed_at"
	case "completed_month":
		return fmt.Sprintf("COALESCE(TO_CHAR(issues.completed_at, '%s'), 'N/A')", s.timeFormat("month", interval)), "", "completed_at"

	// ---- 时间轴：updated_at ----
	case "updated_day":
		return fmt.Sprintf("COALESCE(TO_CHAR(issues.updated_at, '%s'), 'N/A')", s.timeFormat("day", interval)), "", "updated_at"
	case "updated_week":
		return fmt.Sprintf("COALESCE(TO_CHAR(issues.updated_at, '%s'), 'N/A')", s.timeFormat("week", interval)), "", "updated_at"
	case "updated_month":
		return fmt.Sprintf("COALESCE(TO_CHAR(issues.updated_at, '%s'), 'N/A')", s.timeFormat("month", interval)), "", "updated_at"

	default:
		// 默认回退到 state
		sel, join := s.buildAggregation("state")
		return sel, join, ""
	}
}

// timeFormat 根据粒度返回 PostgreSQL TO_CHAR 格式
func (s *ReportService) timeFormat(grain, interval string) string {
	// 如果调用方传了 interval 则优先使用
	if interval == "day" || interval == "week" || interval == "month" {
		grain = interval
	}
	switch grain {
	case "day":
		return "YYYY-MM-DD"
	case "month":
		return "YYYY-MM"
	default: // week
		return "IYYY-IW"
	}
}

// resolveYAxis 根据 y_axis 值返回 (ySelectExpr, extraWhere)
func (s *ReportService) resolveYAxis(yAxis string) (selectExpr, extraWhere string) {
	switch yAxis {
	case "avg_processing_time":
		return "AVG(EXTRACT(EPOCH FROM (issues.completed_at - issues.created_at)) / 86400)", "issues.completed_at IS NOT NULL"
	case "current_retention":
		return "AVG(EXTRACT(EPOCH FROM (NOW() - issues.created_at)) / 86400)", "issues.completed_at IS NULL"
	default: // count
		return "COUNT(*)", ""
	}
}

// v2CreatedVsResolved 处理 created_vs_resolved 指标：需要两个独立查询
func (s *ReportService) v2CreatedVsResolved(projectID uint64, xAxis string, req *ReportV2Request, ids []uint64) (*ReportResponse, error) {
	// created_vs_resolved 需要时间轴作为 X 轴
	// 解析 x_axis 以获取时间列和格式
	_, _, timeColumn := s.resolveXAxis(xAxis, req.Interval)
	if timeColumn == "" {
		// 非时间轴不支持 created_vs_resolved，回退到 created_at week
		timeColumn = "created_at"
	}

	// 确定 TO_CHAR 格式
	var dateFormat string
	switch {
	case strings.Contains(xAxis, "_day"):
		dateFormat = s.timeFormat("day", req.Interval)
	case strings.Contains(xAxis, "_month"):
		dateFormat = s.timeFormat("month", req.Interval)
	default:
		dateFormat = s.timeFormat("week", req.Interval)
	}

	// 根据 x_axis 选择对应的时间列
	var createdTimeCol, resolvedTimeCol string
	switch {
	case strings.HasPrefix(xAxis, "completed_"):
		createdTimeCol = "created_at"
		resolvedTimeCol = "completed_at"
	case strings.HasPrefix(xAxis, "updated_"):
		createdTimeCol = "created_at"
		resolvedTimeCol = "updated_at"
	default:
		createdTimeCol = "created_at"
		resolvedTimeCol = "completed_at"
	}

	type crRow struct {
		Period string
		Cnt    int
	}

	// ---- 构建 created 查询 ----
	createdSQL := fmt.Sprintf(
		"SELECT TO_CHAR(issues.%s, '%s') as period, COUNT(*)::int as cnt FROM issues WHERE issues.project_id = ? AND issues.archived_at IS NULL",
		createdTimeCol, dateFormat)
	createdArgs := []interface{}{projectID}

	// ---- 构建 resolved 查询 ----
	resolvedSQL := fmt.Sprintf(
		"SELECT TO_CHAR(issues.%s, '%s') as period, COUNT(*)::int as cnt FROM issues WHERE issues.project_id = ? AND issues.archived_at IS NULL AND issues.completed_at IS NOT NULL",
		resolvedTimeCol, dateFormat)
	resolvedArgs := []interface{}{projectID}

	// 追加 RQL 过滤
	if ids != nil {
		if len(ids) == 0 {
			return &ReportResponse{Type: "created_vs_resolved", GroupBy: xAxis, Labels: []string{}, Values: []int{}, Values2: []int{}}, nil
		}
		placeholders := make([]string, len(ids))
		vals := make([]interface{}, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			vals[i] = id
		}
		inClause := " AND issues.id IN (" + strings.Join(placeholders, ",") + ")"
		createdSQL += inClause
		resolvedSQL += inClause
		createdArgs = append(createdArgs, vals...)
		resolvedArgs = append(resolvedArgs, vals...)
	}

	// 追加日期过滤
	if req.DateFrom != "" {
		createdSQL += " AND issues." + createdTimeCol + " >= ?"
		createdArgs = append(createdArgs, req.DateFrom)
		resolvedSQL += " AND issues." + resolvedTimeCol + " >= ?"
		resolvedArgs = append(resolvedArgs, req.DateFrom)
	}
	if req.DateTo != "" {
		createdSQL += " AND issues." + createdTimeCol + " <= ?"
		createdArgs = append(createdArgs, req.DateTo+" 23:59:59")
		resolvedSQL += " AND issues." + resolvedTimeCol + " <= ?"
		resolvedArgs = append(resolvedArgs, req.DateTo+" 23:59:59")
	}

	createdSQL += " GROUP BY 1 ORDER BY 1"
	resolvedSQL += " GROUP BY 1 ORDER BY 1"

	// 执行两个查询
	var createdRows, resolvedRows []crRow
	if err := s.db.Raw(createdSQL, createdArgs...).Scan(&createdRows).Error; err != nil {
		return nil, common.Internal("Failed v2 created query: " + err.Error())
	}
	if err := s.db.Raw(resolvedSQL, resolvedArgs...).Scan(&resolvedRows).Error; err != nil {
		return nil, common.Internal("Failed v2 resolved query: " + err.Error())
	}

	// 合并两个数据集
	periodMap := make(map[string]*[2]int)
	for _, r := range createdRows {
		v := periodMap[r.Period]
		if v == nil {
			v = &[2]int{}
			periodMap[r.Period] = v
		}
		v[0] = r.Cnt
	}
	for _, r := range resolvedRows {
		v := periodMap[r.Period]
		if v == nil {
			v = &[2]int{}
			periodMap[r.Period] = v
		}
		v[1] = r.Cnt
	}

	// 按 period 排序
	var periods []string
	for p := range periodMap {
		periods = append(periods, p)
	}
	for i := 0; i < len(periods); i++ {
		for j := i + 1; j < len(periods); j++ {
			if periods[i] > periods[j] {
				periods[i], periods[j] = periods[j], periods[i]
			}
		}
	}

	resp := &ReportResponse{
		Type:    "created_vs_resolved",
		GroupBy: xAxis,
		Colors:  map[string]string{"Created": "#3B82F6", "Resolved": "#10B981"},
	}
	for _, p := range periods {
		v := periodMap[p]
		resp.Labels = append(resp.Labels, p)
		resp.Values = append(resp.Values, v[0])
		resp.Values2 = append(resp.Values2, v[1])
		resp.Total += v[0]
	}
	if resp.Labels == nil {
		resp.Labels = []string{}
		resp.Values = []int{}
		resp.Values2 = []int{}
	}
	return resp, nil
}

// getFilteredIssueIDs 使用 RQL 筛选得到 Issue ID 列表
func (s *ReportService) getFilteredIssueIDs(projectID uint64, rqlQuery string) ([]uint64, error) {
	if rqlQuery == "" {
		return nil, nil // nil 表示不筛选，使用全量
	}

	issues, _, err := s.rqlService.SearchIssues(s.db, projectID, rqlQuery, 1, 10000)
	if err != nil {
		return nil, err
	}
	if len(issues) == 0 {
		return nil, nil // 返回 nil 表示无匹配
	}
	ids := make([]uint64, len(issues))
	for i, issue := range issues {
		ids[i] = issue.ID
	}
	return ids, nil
}

// applyIssueFilter 应用 Issue ID 过滤
func (s *ReportService) applyIssueFilter(query *gorm.DB, ids []uint64) *gorm.DB {
	if ids == nil {
		return query // 全量
	}
	if len(ids) == 0 {
		return query.Where("1 = 0") // 无结果
	}
	return query.Where("issues.id IN ?", ids)
}

// ========================
// 1. Distribution (分组统计)
// ========================
func (s *ReportService) distribution(projectID uint64, req *ReportRequest) (*ReportResponse, error) {
	ids, err := s.getFilteredIssueIDs(projectID, req.RQL)
	if err != nil {
		return nil, common.BadRequest("Invalid RQL: " + err.Error())
	}

	type row struct {
		Name string
		Cnt  int
	}
	var rows []row

	selectExpr, joinClause := s.buildAggregation(req.GroupBy)
	sql := "SELECT " + selectExpr + ", COUNT(*) as cnt FROM issues"
	if joinClause != "" {
		sql += " " + joinClause
	}
	sql += " WHERE issues.project_id = ? AND issues.archived_at IS NULL"
	args := []interface{}{projectID}

	if ids != nil {
		// RQL 过滤
		placeholders := make([]string, len(ids))
		vals := make([]interface{}, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			vals[i] = id
		}
		sql += " AND issues.id IN (" + strings.Join(placeholders, ",") + ")"
		args = append(args, vals...)
	}

	if req.DateFrom != "" {
		sql += " AND issues.created_at >= ?"
		args = append(args, req.DateFrom)
	}
	if req.DateTo != "" {
		sql += " AND issues.created_at <= ?"
		args = append(args, req.DateTo+" 23:59:59")
	}
	sql += " GROUP BY 1 ORDER BY cnt DESC"

	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, common.Internal("Failed to generate distribution report: " + err.Error())
	}

	resp := &ReportResponse{
		Type:    "distribution",
		GroupBy: req.GroupBy,
		Colors:  s.getColors(req.GroupBy),
	}
	for _, r := range rows {
		resp.Labels = append(resp.Labels, r.Name)
		resp.Values = append(resp.Values, r.Cnt)
		resp.Total += r.Cnt
	}
	if resp.Labels == nil {
		resp.Labels = []string{}
		resp.Values = []int{}
	}
	return resp, nil
}

// ========================
// 2. Created vs Resolved (创建 vs 解决趋势)
// ========================
func (s *ReportService) createdVsResolved(projectID uint64, req *ReportRequest) (*ReportResponse, error) {
	ids, err := s.getFilteredIssueIDs(projectID, req.RQL)
	if err != nil {
		return nil, common.BadRequest("Invalid RQL: " + err.Error())
	}

	interval := req.Interval
	if interval == "" {
		interval = "week"
	}

	// DateFormat for SQL TO_CHAR
	var dateFormat string
	switch interval {
	case "month":
		dateFormat = "YYYY-MM"
	case "day":
		dateFormat = "YYYY-MM-DD"
	default:
		dateFormat = "IYYY-IW" // ISO week
	}

	// 需要在 RQL 子集中计算，如果 ids 不为空则需要在子查询中过滤
	// 简化方案：如果 RQL 为空，全量计算；如果有 RQL，先拿 ids 再手工过滤
	if ids == nil {
		createdSQL := fmt.Sprintf(
			"SELECT TO_CHAR(created_at, '%s') as period, COUNT(*) as created FROM issues WHERE project_id = ? AND archived_at IS NULL",
			dateFormat)
		createdArgs := []interface{}{projectID}
		if req.DateFrom != "" {
			createdSQL += " AND created_at >= ?"
			createdArgs = append(createdArgs, req.DateFrom)
		}
		if req.DateTo != "" {
			createdSQL += " AND created_at <= ?"
			createdArgs = append(createdArgs, req.DateTo+" 23:59:59")
		}
		createdSQL += " GROUP BY 1 ORDER BY 1"

		resolvedSQL := fmt.Sprintf(
			"SELECT TO_CHAR(completed_at, '%s') as period, COUNT(*) as resolved FROM issues WHERE project_id = ? AND archived_at IS NULL AND completed_at IS NOT NULL",
			dateFormat)
		resolvedArgs := []interface{}{projectID}
		if req.DateFrom != "" {
			resolvedSQL += " AND completed_at >= ?"
			resolvedArgs = append(resolvedArgs, req.DateFrom)
		}
		if req.DateTo != "" {
			resolvedSQL += " AND completed_at <= ?"
			resolvedArgs = append(resolvedArgs, req.DateTo+" 23:59:59")
		}
		resolvedSQL += " GROUP BY 1 ORDER BY 1"

		type crRow struct {
			Period string
			Cnt    int
		}
		var createdRows, resolvedRows []crRow
		if err := s.db.Raw(createdSQL, createdArgs...).Scan(&createdRows).Error; err != nil {
			return nil, common.Internal("Failed: " + err.Error())
		}
		if err := s.db.Raw(resolvedSQL, resolvedArgs...).Scan(&resolvedRows).Error; err != nil {
			return nil, common.Internal("Failed: " + err.Error())
		}

		// 合并两个数据集
		periodMap := make(map[string]*[2]int)
		for _, r := range createdRows {
			v := periodMap[r.Period]
			if v == nil {
				v = &[2]int{}
				periodMap[r.Period] = v
			}
			v[0] = r.Cnt
		}
		for _, r := range resolvedRows {
			v := periodMap[r.Period]
			if v == nil {
				v = &[2]int{}
				periodMap[r.Period] = v
			}
			v[1] = r.Cnt
		}

		// 按 period 排序
		var periods []string
		for p := range periodMap {
			periods = append(periods, p)
		}
		// bubble sort by period string
		for i := 0; i < len(periods); i++ {
			for j := i + 1; j < len(periods); j++ {
				if periods[i] > periods[j] {
					periods[i], periods[j] = periods[j], periods[i]
				}
			}
		}

		resp := &ReportResponse{Type: "created_vs_resolved", Colors: map[string]string{"Created": "#3B82F6", "Resolved": "#10B981"}}
		for _, p := range periods {
			v := periodMap[p]
			resp.Labels = append(resp.Labels, p)
			resp.Values = append(resp.Values, v[0])
			resp.Values2 = append(resp.Values2, v[1])
			resp.Total += v[0]
		}
		return resp, nil
	}

	// 有 RQL 的情况：只能按 dateFormat 在 Go 里手动分组
	// 查询 ids 子集中的所有 issues 的 created_at / completed_at
	issueDates, err := s.getIssueDates(projectID, ids, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, err
	}

	return s.buildCreatedVsResolvedFromDates(issueDates, interval), nil
}

type issueDates struct {
	CreatedAt   *time.Time
	CompletedAt *time.Time
}

func (s *ReportService) getIssueDates(projectID uint64, ids []uint64, dateFrom, dateTo string) ([]issueDates, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	type row struct {
		CreatedAt   *time.Time
		CompletedAt *time.Time
	}
	var rows []row
	query := s.db.Table("issues").
		Select("created_at, completed_at").
		Where("project_id = ? AND id IN ? AND archived_at IS NULL", projectID, ids)
	if dateFrom != "" {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("created_at <= ?", dateTo+" 23:59:59")
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]issueDates, len(rows))
	for i, r := range rows {
		result[i] = issueDates{CreatedAt: r.CreatedAt, CompletedAt: r.CompletedAt}
	}
	return result, nil
}

func (s *ReportService) buildCreatedVsResolvedFromDates(dates []issueDates, interval string) *ReportResponse {
	periodMap := make(map[string]*[2]int)
	for _, d := range dates {
		if d.CreatedAt == nil {
			continue
		}
		period := s.formatPeriod(*d.CreatedAt, interval)
		v := periodMap[period]
		if v == nil {
			v = &[2]int{}
			periodMap[period] = v
		}
		v[0]++
		if d.CompletedAt != nil {
			rp := s.formatPeriod(*d.CompletedAt, interval)
			rv := periodMap[rp]
			if rv == nil {
				rv = &[2]int{}
				periodMap[rp] = rv
			}
			rv[1]++
		}
	}

	var periods []string
	for p := range periodMap {
		periods = append(periods, p)
	}
	for i := 0; i < len(periods); i++ {
		for j := i + 1; j < len(periods); j++ {
			if periods[i] > periods[j] {
				periods[i], periods[j] = periods[j], periods[i]
			}
		}
	}

	resp := &ReportResponse{Type: "created_vs_resolved", Colors: map[string]string{"Created": "#3B82F6", "Resolved": "#10B981"}}
	for _, p := range periods {
		v := periodMap[p]
		resp.Labels = append(resp.Labels, p)
		resp.Values = append(resp.Values, v[0])
		resp.Values2 = append(resp.Values2, v[1])
		resp.Total += v[0]
	}
	return resp
}

func (s *ReportService) formatPeriod(t time.Time, interval string) string {
	switch interval {
	case "day":
		return t.Format("2006-01-02")
	case "month":
		return t.Format("2006-01")
	default: // week
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	}
}

// ========================
// 3. Average Age (平均处理时长)
// ========================
func (s *ReportService) avgAge(projectID uint64, req *ReportRequest) (*ReportResponse, error) {
	ids, err := s.getFilteredIssueIDs(projectID, req.RQL)
	if err != nil {
		return nil, common.BadRequest("Invalid RQL: " + err.Error())
	}

	type row struct {
		Name    string
		AvgDays float64
		Cnt     int
	}
	var rows []row

	selectExpr, joinClause := s.buildAggregation(req.GroupBy)
	sql := `SELECT ` + selectExpr + `,
		AVG(EXTRACT(EPOCH FROM (issues.completed_at - issues.created_at)) / 86400)::numeric(10,1) as avg_days,
		COUNT(*) as cnt
		FROM issues`
	if joinClause != "" {
		sql += " " + joinClause
	}
	sql += " WHERE issues.project_id = ? AND issues.archived_at IS NULL AND issues.completed_at IS NOT NULL"
	args := []interface{}{projectID}

	if ids != nil {
		if len(ids) == 0 {
			return &ReportResponse{Type: "avg_age", GroupBy: req.GroupBy, Labels: []string{}, Values: []int{}}, nil
		}
		placeholders := make([]string, len(ids))
		vals := make([]interface{}, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			vals[i] = id
		}
		sql += " AND issues.id IN (" + strings.Join(placeholders, ",") + ")"
		args = append(args, vals...)
	}
	if req.DateFrom != "" {
		sql += " AND issues.created_at >= ?"
		args = append(args, req.DateFrom)
	}
	if req.DateTo != "" {
		sql += " AND issues.created_at <= ?"
		args = append(args, req.DateTo+" 23:59:59")
	}
	sql += " GROUP BY 1 ORDER BY avg_days DESC"

	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, common.Internal("Failed: " + err.Error())
	}

	resp := &ReportResponse{Type: "avg_age", GroupBy: req.GroupBy, Colors: s.getColors(req.GroupBy)}
	var totalDays float64
	for _, r := range rows {
		resp.Labels = append(resp.Labels, fmt.Sprintf("%s (%.1fd)", r.Name, r.AvgDays))
		resp.Values = append(resp.Values, int(r.AvgDays))
		resp.Total += r.Cnt
		totalDays += r.AvgDays * float64(r.Cnt)
	}
	if resp.Labels == nil {
		resp.Labels = []string{}
		resp.Values = []int{}
	}
	if resp.Total > 0 {
		resp.Summary = &ReportSummary{AvgDays: totalDays / float64(resp.Total)}
	}
	return resp, nil
}

// ========================
// 4. Current Age (当前滞留时长)
// ========================
func (s *ReportService) currentAge(projectID uint64, req *ReportRequest) (*ReportResponse, error) {
	ids, err := s.getFilteredIssueIDs(projectID, req.RQL)
	if err != nil {
		return nil, common.BadRequest("Invalid RQL: " + err.Error())
	}

	type row struct {
		Name    string
		AvgDays float64
		Cnt     int
	}
	var rows []row

	selectExpr, joinClause := s.buildAggregation(req.GroupBy)
	sql := `SELECT ` + selectExpr + `,
		AVG(EXTRACT(EPOCH FROM (NOW() - issues.created_at)) / 86400)::numeric(10,1) as avg_days,
		COUNT(*) as cnt
		FROM issues`
	if joinClause != "" {
		sql += " " + joinClause
	}
	sql += " WHERE issues.project_id = ? AND issues.archived_at IS NULL AND issues.completed_at IS NULL"
	args := []interface{}{projectID}

	if ids != nil {
		if len(ids) == 0 {
			return &ReportResponse{Type: "current_age", GroupBy: req.GroupBy, Labels: []string{}, Values: []int{}}, nil
		}
		placeholders := make([]string, len(ids))
		vals := make([]interface{}, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			vals[i] = id
		}
		sql += " AND issues.id IN (" + strings.Join(placeholders, ",") + ")"
		args = append(args, vals...)
	}
	sql += " GROUP BY 1 ORDER BY avg_days DESC"

	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, common.Internal("Failed: " + err.Error())
	}

	resp := &ReportResponse{Type: "current_age", GroupBy: req.GroupBy, Colors: s.getColors(req.GroupBy)}
	var totalDays float64
	for _, r := range rows {
		resp.Labels = append(resp.Labels, fmt.Sprintf("%s (%.1fd)", r.Name, r.AvgDays))
		resp.Values = append(resp.Values, int(r.AvgDays))
		resp.Total += r.Cnt
		totalDays += r.AvgDays * float64(r.Cnt)
	}
	if resp.Labels == nil {
		resp.Labels = []string{}
		resp.Values = []int{}
	}
	if resp.Total > 0 {
		resp.Summary = &ReportSummary{AvgDays: totalDays / float64(resp.Total)}
	}
	return resp, nil
}

// ========================
// 5. Created Trend (创建趋势)
// ========================
func (s *ReportService) createdTrend(projectID uint64, req *ReportRequest) (*ReportResponse, error) {
	ids, err := s.getFilteredIssueIDs(projectID, req.RQL)
	if err != nil {
		return nil, common.BadRequest("Invalid RQL: " + err.Error())
	}

	interval := req.Interval
	if interval == "" {
		interval = "day"
	}

	if ids == nil {
		var dateFormat string
		switch interval {
		case "month":
			dateFormat = "YYYY-MM"
		case "week":
			dateFormat = "IYYY-IW"
		default:
			dateFormat = "YYYY-MM-DD"
		}

		sql := fmt.Sprintf(
			"SELECT TO_CHAR(created_at, '%s') as period, COUNT(*) as cnt FROM issues WHERE project_id = ? AND archived_at IS NULL",
			dateFormat)
		args := []interface{}{projectID}
		if req.DateFrom != "" {
			sql += " AND created_at >= ?"
			args = append(args, req.DateFrom)
		}
		if req.DateTo != "" {
			sql += " AND created_at <= ?"
			args = append(args, req.DateTo+" 23:59:59")
		}
		sql += " GROUP BY 1 ORDER BY 1"

		type crRow struct {
			Period string
			Cnt    int
		}
		var crs []crRow
		if err := s.db.Raw(sql, args...).Scan(&crs).Error; err != nil {
			return nil, common.Internal("Failed: " + err.Error())
		}
		resp := &ReportResponse{Type: "created_trend"}
		for _, r := range crs {
			resp.Labels = append(resp.Labels, r.Period)
			resp.Values = append(resp.Values, r.Cnt)
			resp.Total += r.Cnt
		}
		return resp, nil
	}

	// 有 RQL：查 dates 后手动分组
	dates, err := s.getIssueDates(projectID, ids, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, err
	}
	periodMap := make(map[string]int)
	for _, d := range dates {
		if d.CreatedAt == nil {
			continue
		}
		periodMap[s.formatPeriod(*d.CreatedAt, interval)]++
	}

	var periods []string
	for p := range periodMap {
		periods = append(periods, p)
	}
	for i := 0; i < len(periods); i++ {
		for j := i + 1; j < len(periods); j++ {
			if periods[i] > periods[j] {
				periods[i], periods[j] = periods[j], periods[i]
			}
		}
	}

	resp := &ReportResponse{Type: "created_trend"}
	for _, p := range periods {
		resp.Labels = append(resp.Labels, p)
		resp.Values = append(resp.Values, periodMap[p])
		resp.Total += periodMap[p]
	}
	return resp, nil
}

// ========================
// Helpers
// ========================

func (s *ReportService) buildAggregation(groupBy string) (selectExpr, join string) {
	switch groupBy {
	case "state":
		return "COALESCE(s.name, 'Unknown') as name", "LEFT JOIN states s ON issues.state_id = s.id"
	case "state_group":
		return "COALESCE(sg.\"group\", 'No Group') as name", "LEFT JOIN states sg ON issues.state_id = sg.id"
	case "priority":
		return "COALESCE(issues.priority, 'none') as name", ""
	case "assignee":
		return "COALESCE(u.display_name, 'Unassigned') as name", "LEFT JOIN issue_assignees ia ON issues.id = ia.issue_id LEFT JOIN users u ON ia.user_id = u.id"
	case "type":
		return "COALESCE(it.name, 'Untyped') as name", "LEFT JOIN issue_types it ON issues.issue_type_id = it.id"
	case "label":
		return "COALESCE(l.name, 'No Label') as name", "LEFT JOIN issue_labels il ON issues.id = il.issue_id LEFT JOIN labels l ON il.label_id = l.id"
	case "cycle":
		return "COALESCE(c.name, 'No Cycle') as name", "LEFT JOIN issue_cycles ic ON issues.id = ic.issue_id LEFT JOIN cycles c ON ic.cycle_id = c.id"
	case "module":
		return "COALESCE(m.name, 'No Module') as name", "LEFT JOIN module_issues mi ON issues.id = mi.issue_id LEFT JOIN modules m ON mi.module_id = m.id"
	case "created_by":
		return "COALESCE(cu.display_name, 'Unknown') as name", "LEFT JOIN users cu ON issues.created_by_id = cu.id"
	default:
		return "COALESCE(s.name, 'Unknown') as name", "LEFT JOIN states s ON issues.state_id = s.id"
	}
}

func (s *ReportService) getColors(groupBy string) map[string]string {
	switch groupBy {
	case "priority":
		return map[string]string{
			"urgent": "#EF4444", "high": "#F59E0B", "medium": "#3B82F6",
			"low": "#6B7280", "none": "#9CA3AF",
		}
	case "state":
		return map[string]string{
			"Backlog": "#6B7280", "Todo": "#3B82F6", "In Progress": "#F59E0B",
			"In Review": "#8B5CF6", "Done": "#10B981", "Cancelled": "#EF4444",
		}
	default:
		return presetColors
	}
}

var presetColors = map[string]string{
	"#1": "#3B82F6", "#2": "#10B981", "#3": "#F59E0B", "#4": "#EF4444",
	"#5": "#8B5CF6", "#6": "#EC4899", "#7": "#06B6D4", "#8": "#84CC16",
	"#9": "#F97316", "#10": "#6366F1",
}
