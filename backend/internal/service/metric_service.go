package service

import (
	"encoding/json"

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
