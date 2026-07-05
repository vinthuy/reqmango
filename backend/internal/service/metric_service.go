package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type MetricService struct {
	db *gorm.DB
}

func NewMetricService(db *gorm.DB) *MetricService {
	return &MetricService{db: db}
}

type MetricTemplate struct {
	ID             string                 `json:"id"`
	Category       string                 `json:"category"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	ChartType      string                 `json:"chart_type"`
	DefaultXAxis   string                 `json:"default_x_axis"`
	DefaultYAxis   string                 `json:"default_y_axis"`
	DefaultFilters map[string]interface{} `json:"default_filters"`
	DefaultConfig  map[string]interface{} `json:"default_config"`
	Icon           string                 `json:"icon"`
}

type TemplateCategory struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Templates []MetricTemplate `json:"templates"`
}

func (s *MetricService) ListTemplates() []TemplateCategory {
	return []TemplateCategory{
		{
			ID: "agile", Name: "敏捷效能",
			Templates: []MetricTemplate{
				{ID: "agile_burndown", Category: "agile", Name: "迭代燃尽图", Description: "展示迭代内剩余工作量趋势", ChartType: "line", DefaultXAxis: "created_week", DefaultYAxis: "count", DefaultConfig: map[string]interface{}{"reference_lines": []map[string]interface{}{{"type": "average", "label": "平均值"}}, "show_labels": true}, Icon: "flame"},
				{ID: "agile_velocity", Category: "agile", Name: "速率趋势", Description: "展示每个迭代的完成速率", ChartType: "bar", DefaultXAxis: "created_week", DefaultYAxis: "count", DefaultConfig: map[string]interface{}{"show_labels": true}, Icon: "trending-up"},
				{ID: "agile_cfd", Category: "agile", Name: "累积流图", Description: "展示各状态Issue数量累积变化", ChartType: "area", DefaultXAxis: "created_week", DefaultYAxis: "count", DefaultConfig: map[string]interface{}{"stack_mode": "stack"}, Icon: "layers"},
				{ID: "agile_cycle_time", Category: "agile", Name: "周期时间分布", Description: "展示Issue从开始到完成的时间分布", ChartType: "bar", DefaultXAxis: "completed_week", DefaultYAxis: "avg_processing_time", Icon: "clock"},
				{ID: "agile_lead_time", Category: "agile", Name: "前置时间趋势", Description: "展示需求提出到交付的平均前置时间", ChartType: "line", DefaultXAxis: "created_week", DefaultYAxis: "avg_processing_time", Icon: "timer"},
				{ID: "agile_wip", Category: "agile", Name: "WIP限制", Description: "展示各状态的在制品数量", ChartType: "bar", DefaultXAxis: "state", DefaultYAxis: "count", Icon: "columns"},
			},
		},
		{
			ID: "project", Name: "项目管理",
			Templates: []MetricTemplate{
				{ID: "pm_throughput", Category: "project", Name: "需求吞吐量", Description: "展示单位时间内完成的需求数量", ChartType: "bar", DefaultXAxis: "completed_week", DefaultYAxis: "count", Icon: "package"},
				{ID: "pm_milestone", Category: "project", Name: "里程碑达成率", Description: "展示里程碑按时完成百分比", ChartType: "doughnut", DefaultXAxis: "state", DefaultYAxis: "count", Icon: "target"},
				{ID: "pm_defect_trend", Category: "project", Name: "缺陷趋势", Description: "展示缺陷新建与解决趋势", ChartType: "line", DefaultXAxis: "created_week", DefaultYAxis: "count", DefaultFilters: map[string]interface{}{"type_filter": "bug"}, Icon: "alert-triangle"},
				{ID: "pm_priority", Category: "project", Name: "优先级分布", Description: "展示各优先级Issue占比", ChartType: "pie", DefaultXAxis: "priority", DefaultYAxis: "count", Icon: "bar-chart-2"},
				{ID: "pm_workload", Category: "project", Name: "资源负载", Description: "展示每个成员的工作负载", ChartType: "bar", DefaultXAxis: "assignee", DefaultYAxis: "count", Icon: "users"},
			},
		},
		{
			ID: "quality", Name: "质量分析",
			Templates: []MetricTemplate{
				{ID: "qa_density", Category: "quality", Name: "缺陷密度", Description: "展示各模块的缺陷密度", ChartType: "bar", DefaultXAxis: "module", DefaultYAxis: "count", DefaultFilters: map[string]interface{}{"type_filter": "bug"}, Icon: "search"},
				{ID: "qa_escape", Category: "quality", Name: "缺陷逃逸率", Description: "展示线上缺陷占总缺陷比例", ChartType: "doughnut", DefaultXAxis: "state", DefaultYAxis: "count", DefaultFilters: map[string]interface{}{"type_filter": "bug"}, Icon: "shield"},
				{ID: "qa_review_pass", Category: "quality", Name: "评审一次通过率", Description: "展示评审一次通过比例", ChartType: "doughnut", DefaultXAxis: "state", DefaultYAxis: "count", Icon: "check-circle"},
				{ID: "qa_state_dwell", Category: "quality", Name: "状态停留时间", Description: "展示Issue在各状态的平均停留时间", ChartType: "bar", DefaultXAxis: "state", DefaultYAxis: "current_retention", Icon: "pause-circle"},
			},
		},
	}
}

type ChartConfig struct {
	StackMode      string          `json:"stack_mode,omitempty"`
	ReferenceLines []ReferenceLine `json:"reference_lines,omitempty"`
	ShowLabels     bool            `json:"show_labels"`
	DualYAxis      bool            `json:"dual_y_axis"`
	LegendPosition string          `json:"legend_position,omitempty"`
}

type ReferenceLine struct {
	Type  string  `json:"type"`
	Value float64 `json:"value,omitempty"`
	Label string  `json:"label,omitempty"`
	Color string  `json:"color,omitempty"`
	Style string  `json:"style,omitempty"`
}

func (s *MetricService) ListCharts(projectID uint64) ([]model.MetricChart, error) {
	var charts []model.MetricChart
	err := s.db.Where("project_id = ? AND is_visible = ?", projectID, true).
		Order("sort_order ASC, id ASC").
		Find(&charts).Error
	return charts, err
}

func (s *MetricService) GetChart(projectID, chartID uint64) (*model.MetricChart, error) {
	var chart model.MetricChart
	err := s.db.Where("project_id = ? AND id = ?", projectID, chartID).First(&chart).Error
	if err != nil {
		return nil, common.NotFound("Chart not found")
	}
	return &chart, nil
}

func (s *MetricService) CreateChart(projectID, creatorID uint64, req *CreateChartRequest) (*model.MetricChart, error) {
	filtersJSON, _ := json.Marshal(req.Filters)
	configJSON, _ := json.Marshal(req.Config)

	chart := model.MetricChart{
		ProjectID:  projectID,
		CreatorID:  creatorID,
		Name:       req.Name,
		TemplateID: req.TemplateID,
		ChartType:  req.ChartType,
		XAxis:      req.XAxis,
		YAxis:      req.YAxis,
		Filters:    string(filtersJSON),
		Config:     string(configJSON),
		SortOrder:  req.SortOrder,
		IsVisible:  true,
	}
	if err := s.db.Create(&chart).Error; err != nil {
		return nil, common.Internal("Failed to create chart: " + err.Error())
	}
	return &chart, nil
}

func (s *MetricService) UpdateChart(projectID, chartID uint64, req *UpdateChartRequest) (*model.MetricChart, error) {
	chart, err := s.GetChart(projectID, chartID)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		chart.Name = req.Name
	}
	if req.ChartType != "" {
		chart.ChartType = req.ChartType
	}
	if req.XAxis != "" {
		chart.XAxis = req.XAxis
	}
	if req.YAxis != "" {
		chart.YAxis = req.YAxis
	}
	if req.Filters != nil {
		filtersJSON, _ := json.Marshal(req.Filters)
		chart.Filters = string(filtersJSON)
	}
	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		chart.Config = string(configJSON)
	}
	if err := s.db.Save(chart).Error; err != nil {
		return nil, common.Internal("Failed to update chart: " + err.Error())
	}
	return chart, nil
}

func (s *MetricService) DeleteChart(projectID, chartID uint64) error {
	result := s.db.Where("project_id = ? AND id = ?", projectID, chartID).Delete(&model.MetricChart{})
	if result.RowsAffected == 0 {
		return common.NotFound("Chart not found")
	}
	return result.Error
}

func (s *MetricService) ReorderCharts(projectID uint64, chartIDs []uint64) error {
	for i, id := range chartIDs {
		s.db.Model(&model.MetricChart{}).
			Where("project_id = ? AND id = ?", projectID, id).
			Update("sort_order", i)
	}
	return nil
}

type CustomFieldInfo struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	FieldType string `json:"field_type"`
}

func (s *MetricService) GetCustomFields(projectID uint64) ([]CustomFieldInfo, error) {
	var fields []CustomFieldInfo
	err := s.db.Raw(
		`SELECT id, name, field_type FROM custom_fields WHERE (project_id = ? OR project_id IS NULL) AND is_active = true ORDER BY name`,
		projectID,
	).Scan(&fields).Error
	if err != nil {
		return nil, common.Internal("Failed to query custom fields: " + err.Error())
	}
	return fields, nil
}

func (s *MetricService) RenderChart(projectID, chartID uint64) (*RenderResponse, error) {
	chart, err := s.GetChart(projectID, chartID)
	if err != nil {
		return nil, err
	}

	reportReq := &ReportV2Request{
		XAxis: chart.XAxis,
		YAxis: chart.YAxis,
	}

	var filtersMap map[string]interface{}
	if chart.Filters != "" && chart.Filters != "{}" {
		json.Unmarshal([]byte(chart.Filters), &filtersMap)
		if rql, ok := filtersMap["rql"].(string); ok && rql != "" {
			reportReq.RQL = rql
		} else if conds, ok := filtersMap["conditions"].([]interface{}); ok && len(conds) > 0 {
			reportReq.RQL = conditionsToRQL(conds)
		}
	}

	reportSvc := NewReportService(s.db)
	reportResp, err := reportSvc.GenerateV2(projectID, reportReq)
	if err != nil {
		return nil, err
	}

	var config ChartConfig
	if chart.Config != "" && chart.Config != "{}" {
		json.Unmarshal([]byte(chart.Config), &config)
	}

	var refLines []ReferenceLineData
	for _, rl := range config.ReferenceLines {
		refLine := ReferenceLineData{Type: rl.Type, Label: rl.Label, Color: rl.Color, Style: rl.Style}
		switch rl.Type {
		case "constant":
			refLine.Value = rl.Value
		case "average":
			if len(reportResp.Values) > 0 {
				sum := 0
				for _, v := range reportResp.Values {
					sum += v
				}
				refLine.Value = float64(sum) / float64(len(reportResp.Values))
			}
		case "max":
			if len(reportResp.Values) > 0 {
				max := reportResp.Values[0]
				for _, v := range reportResp.Values[1:] {
					if v > max {
						max = v
					}
				}
				refLine.Value = float64(max)
			}
		case "min":
			if len(reportResp.Values) > 0 {
				min := reportResp.Values[0]
				for _, v := range reportResp.Values[1:] {
					if v < min {
						min = v
					}
				}
				refLine.Value = float64(min)
			}
		case "sum":
			sum := 0
			for _, v := range reportResp.Values {
				sum += v
			}
			refLine.Value = float64(sum)
		}
		refLines = append(refLines, refLine)
	}

	return &RenderResponse{
		Labels: reportResp.Labels, Values: reportResp.Values, Total: reportResp.Total,
		Colors: reportResp.Colors, ReferenceLines: refLines, ChartType: chart.ChartType, Config: config,
	}, nil
}

// RenderChartData renders chart data directly from request params (for live preview).
func (s *MetricService) RenderChartData(projectID uint64, req *CreateChartRequest) (*RenderResponse, error) {
	reportReq := &ReportV2Request{
		XAxis: req.XAxis,
		YAxis: req.YAxis,
	}
	if req.Filters != nil {
		// Prefer explicit rql string if provided
		if rql, ok := req.Filters["rql"].(string); ok && rql != "" {
			reportReq.RQL = rql
		} else if conds, ok := req.Filters["conditions"].([]interface{}); ok && len(conds) > 0 {
			// Convert structured conditions array to RQL string
			reportReq.RQL = conditionsToRQL(conds)
		}
	}

	reportSvc := NewReportService(s.db)
	reportResp, err := reportSvc.GenerateV2(projectID, reportReq)
	if err != nil {
		return nil, err
	}

	var config ChartConfig
	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		json.Unmarshal(configJSON, &config)
	}

	chartType := req.ChartType
	if chartType == "" {
		chartType = "bar"
	}

	return &RenderResponse{
		Labels: reportResp.Labels, Values: reportResp.Values, Total: reportResp.Total,
		Colors: reportResp.Colors, ReferenceLines: nil, ChartType: chartType, Config: config,
	}, nil
}

// FilterValues holds distinct values for each filter field in a project.
type FilterValues struct {
	State        []string            `json:"state"`
	Priority     []string            `json:"priority"`
	Assignee     []string            `json:"assignee"`
	Type         []string            `json:"type"`
	Label        []string            `json:"label"`
	Module       []string            `json:"module"`
	CreatedBy    []string            `json:"created_by"`
	CustomFields map[string][]string `json:"custom_fields"` // field_id -> values
}

// GetFilterValues returns distinct non-empty values for filter fields in a project.
func (s *MetricService) GetFilterValues(projectID uint64) (*FilterValues, error) {
	fv := &FilterValues{}

	// State names via join with states table
	var states []string
	err := s.db.Raw(
		`SELECT DISTINCT s.name FROM issues i JOIN states s ON i.state_id = s.id WHERE i.project_id = ? AND i.deleted_at IS NULL`,
		projectID,
	).Scan(&states).Error
	if err != nil {
		return nil, common.Internal("Failed to query state values: " + err.Error())
	}
	fv.State = states

	// Priority values
	var priorities []string
	err = s.db.Raw(
		`SELECT DISTINCT priority FROM issues WHERE project_id = ? AND deleted_at IS NULL AND priority != '' AND priority IS NOT NULL`,
		projectID,
	).Scan(&priorities).Error
	if err != nil {
		return nil, common.Internal("Failed to query priority values: " + err.Error())
	}
	fv.Priority = priorities

	// Assignee display names via join with issue_assignees and users
	var assignees []string
	err = s.db.Raw(
		`SELECT DISTINCT COALESCE(NULLIF(TRIM(u.display_name), ''), u.username) FROM issues i JOIN issue_assignees ia ON ia.issue_id = i.id JOIN users u ON ia.user_id = u.id WHERE i.project_id = ? AND i.deleted_at IS NULL`,
		projectID,
	).Scan(&assignees).Error
	if err != nil {
		return nil, common.Internal("Failed to query assignee values: " + err.Error())
	}
	fv.Assignee = assignees

	// Issue type names via join with issue_types
	var types []string
	err = s.db.Raw(
		`SELECT DISTINCT t.name FROM issues i JOIN issue_types t ON i.issue_type_id = t.id WHERE i.project_id = ? AND i.deleted_at IS NULL AND i.issue_type_id IS NOT NULL`,
		projectID,
	).Scan(&types).Error
	if err != nil {
		return nil, common.Internal("Failed to query type values: " + err.Error())
	}
	fv.Type = types

	// Label names via join with issue_labels and labels
	var labels []string
	err = s.db.Raw(
		`SELECT DISTINCT l.name FROM issues i JOIN issue_labels il ON il.issue_id = i.id JOIN labels l ON il.label_id = l.id WHERE i.project_id = ? AND i.deleted_at IS NULL`,
		projectID,
	).Scan(&labels).Error
	if err != nil {
		return nil, common.Internal("Failed to query label values: " + err.Error())
	}
	fv.Label = labels

	// Module names via join with module_issues and modules
	var modules []string
	err = s.db.Raw(
		`SELECT DISTINCT m.name FROM issues i JOIN module_issues mi ON mi.issue_id = i.id JOIN modules m ON mi.module_id = m.id WHERE i.project_id = ? AND i.deleted_at IS NULL`,
		projectID,
	).Scan(&modules).Error
	if err != nil {
		return nil, common.Internal("Failed to query module values: " + err.Error())
	}
	fv.Module = modules

	// Created by (issue creator names)
	var creators []string
	err = s.db.Raw(
		`SELECT DISTINCT COALESCE(NULLIF(TRIM(u.display_name), ''), u.username) FROM issues i JOIN users u ON i.created_by = u.id WHERE i.project_id = ? AND i.deleted_at IS NULL`,
		projectID,
	).Scan(&creators).Error
	if err != nil {
		return nil, common.Internal("Failed to query creator values: " + err.Error())
	}
	fv.CreatedBy = creators

	// Custom field values
	cfValues := make(map[string][]string)
	var cfFields []struct {
		ID   uint64
		Name string
		Type string
	}
	s.db.Raw(`SELECT id, name, field_type FROM custom_fields WHERE (project_id = ? OR project_id IS NULL) AND is_active = true`, projectID).Scan(&cfFields)
	for _, cf := range cfFields {
		var vals []string
		s.db.Raw(
			`SELECT DISTINCT value FROM issue_custom_field_values WHERE field_id = ? AND value IS NOT NULL AND value != '' AND issue_id IN (SELECT id FROM issues WHERE project_id = ? AND deleted_at IS NULL)`,
			cf.ID, projectID,
		).Scan(&vals)
		if len(vals) > 0 {
			cfValues[fmt.Sprintf("%d", cf.ID)] = vals
		}
	}
	fv.CustomFields = cfValues

	return fv, nil
}

type CreateChartRequest struct {
	Name       string                 `json:"name" binding:"required"`
	TemplateID string                 `json:"template_id"`
	ChartType  string                 `json:"chart_type" binding:"required"`
	XAxis      string                 `json:"x_axis" binding:"required"`
	YAxis      string                 `json:"y_axis" binding:"required"`
	Filters    map[string]interface{} `json:"filters"`
	Config     map[string]interface{} `json:"config"`
	SortOrder  int                    `json:"sort_order"`
}

type UpdateChartRequest struct {
	Name      string                 `json:"name"`
	ChartType string                 `json:"chart_type"`
	XAxis     string                 `json:"x_axis"`
	YAxis     string                 `json:"y_axis"`
	Filters   map[string]interface{} `json:"filters"`
	Config    map[string]interface{} `json:"config"`
}

type RenderResponse struct {
	Labels         []string            `json:"labels"`
	Values         []int               `json:"values"`
	Total          int                 `json:"total"`
	Colors         map[string]string   `json:"colors"`
	ReferenceLines []ReferenceLineData `json:"reference_lines"`
	ChartType      string              `json:"chart_type"`
	Config         ChartConfig         `json:"config"`
}

type ReferenceLineData struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Label string  `json:"label"`
	Color string  `json:"color"`
	Style string  `json:"style"`
}

// conditionsToRQL converts a structured conditions array (from the frontend filter panel)
// into an RQL query string. Each condition is {"field": "...", "operator": "...", "values": [...]}.
// Conditions are joined with AND.
func conditionsToRQL(conditions []interface{}) string {
	var parts []string
	for _, c := range conditions {
		cond, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		field, _ := cond["field"].(string)
		operator, _ := cond["operator"].(string)
		valuesRaw, _ := cond["values"].([]interface{})

		if field == "" || operator == "" {
			continue
		}

		// Custom fields use "custom_field:N" format — map to "cf_N" for the executor
		escapedField := field
		if strings.HasPrefix(field, "custom_field:") {
			escapedField = "cf_" + strings.TrimPrefix(field, "custom_field:")
		}

		// Handle empty/not_empty operators (no values needed)
		if operator == "empty" {
			parts = append(parts, fmt.Sprintf("%s IS NULL", escapedField))
			continue
		}
		if operator == "not_empty" {
			parts = append(parts, fmt.Sprintf("%s IS NOT NULL", escapedField))
			continue
		}

		// Fallback: if values is empty but value (single string) is present, use it
		if len(valuesRaw) == 0 {
			if v, ok := cond["value"].(string); ok && v != "" {
				valuesRaw = []interface{}{v}
			}
		}

		if len(valuesRaw) == 0 {
			continue
		}

		switch operator {
		case "=", "!=":
			if len(valuesRaw) == 1 {
				val := escapeRQLString(fmt.Sprintf("%v", valuesRaw[0]))
				parts = append(parts, fmt.Sprintf("%s %s %s", escapedField, operator, val))
			} else {
				// Multiple values → IN / NOT IN
				inOp := "IN"
				if operator == "!=" {
					inOp = "NOT IN"
				}
				quoted := make([]string, len(valuesRaw))
				for i, v := range valuesRaw {
					quoted[i] = escapeRQLString(fmt.Sprintf("%v", v))
				}
				parts = append(parts, fmt.Sprintf("%s %s (%s)", escapedField, inOp, strings.Join(quoted, ", ")))
			}
		case "IN":
			quoted := make([]string, len(valuesRaw))
			for i, v := range valuesRaw {
				quoted[i] = escapeRQLString(fmt.Sprintf("%v", v))
			}
			parts = append(parts, fmt.Sprintf("%s IN (%s)", escapedField, strings.Join(quoted, ", ")))
		case "NOT IN":
			quoted := make([]string, len(valuesRaw))
			for i, v := range valuesRaw {
				quoted[i] = escapeRQLString(fmt.Sprintf("%v", v))
			}
			parts = append(parts, fmt.Sprintf("%s NOT IN (%s)", escapedField, strings.Join(quoted, ", ")))
		case "LIKE", "NOT LIKE":
			if len(valuesRaw) > 0 {
				val := escapeRQLString(fmt.Sprintf("%%%s%%", valuesRaw[0]))
				parts = append(parts, fmt.Sprintf("%s %s %s", escapedField, operator, val))
			}
		case "contains":
			if len(valuesRaw) > 0 {
				val := escapeRQLString(fmt.Sprintf("%%%s%%", valuesRaw[0]))
				parts = append(parts, fmt.Sprintf("%s LIKE %s", escapedField, val))
			}
		case "not_contains":
			if len(valuesRaw) > 0 {
				val := escapeRQLString(fmt.Sprintf("%%%s%%", valuesRaw[0]))
				parts = append(parts, fmt.Sprintf("%s NOT LIKE %s", escapedField, val))
			}
		case "IS NULL", "IS NOT NULL":
			parts = append(parts, fmt.Sprintf("%s %s", escapedField, operator))
		default:
			// Generic comparison: >, <, >=, <=
			if len(valuesRaw) > 0 {
				val := escapeRQLString(fmt.Sprintf("%v", valuesRaw[0]))
				parts = append(parts, fmt.Sprintf("%s %s %s", escapedField, operator, val))
			}
		}
	}
	return strings.Join(parts, " AND ")
}

// escapeRQLString wraps a value in double quotes for RQL, escaping internal quotes.
func escapeRQLString(val string) string {
	// If the value is numeric, don't quote it
	if _, err := strconv.ParseFloat(val, 64); err == nil {
		return val
	}
	// If it looks like a date (YYYY-MM-DD), don't quote it
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}`, val); matched {
		return val
	}
	// Otherwise, wrap in double quotes with escaped internal quotes
	escaped := strings.ReplaceAll(val, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}
