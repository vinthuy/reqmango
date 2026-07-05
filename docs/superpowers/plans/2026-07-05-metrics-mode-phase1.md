# Metrics Mode Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Metrics Mode MVP — a new metrics view with 16 pre-built templates, custom chart CRUD, drag-to-reorder, and enhanced chart config (stack mode, data labels), replacing the existing ReportBuilder.

**Architecture:** Backend adds a `metric_charts` table + CRUD/render API reusing the existing V2 report engine. Frontend builds a new MetricsView component with template gallery and chart grid, replacing ReportBuilder in Project.vue.

**Tech Stack:** Vue 3 Composition API, Chart.js, Go/Gin, GORM, PostgreSQL JSONB

---

## File Structure

### Backend (new files)
| File | Responsibility |
|------|---------------|
| `backend/internal/model/metric_chart.go` | MetricChart model struct |
| `backend/internal/handler/metric_handler.go` | HTTP handlers for metrics API |
| `backend/internal/service/metric_service.go` | Template engine, chart CRUD, render logic |

### Backend (modify files)
| File | Change |
|------|--------|
| `backend/internal/router/router.go:251-256` | Add metrics routes |
| `backend/internal/model/migrate.go` | Add MetricChart to auto-migrate |

### Frontend (new files)
| File | Responsibility |
|------|---------------|
| `frontend/src/views/MetricsView.vue` | Main metrics view page |
| `frontend/src/components/metrics/MetricsTemplateGallery.vue` | Pre-built template cards by category |
| `frontend/src/components/metrics/MetricsChartCard.vue` | Single chart card with toolbar |
| `frontend/src/components/metrics/MetricsChartConfig.vue` | Chart configuration panel/dialog |
| `frontend/src/api/metrics.ts` | Metrics API client |
| `frontend/src/types/metrics.ts` | TypeScript types for metrics |

### Frontend (modify files)
| File | Change |
|------|--------|
| `frontend/src/views/Project.vue:184-186` | Replace ReportBuilder with MetricsView |
| `frontend/src/views/Project.vue:309` | Update import |
| `frontend/src/views/Project.vue:524` | Rename tab label |
| `frontend/src/locales/en-US.json` | Add metrics i18n keys |
| `frontend/src/locales/zh-CN.json` | Add metrics i18n keys |

---

## Task 1: Backend — MetricChart Model + Migration

**Files:**
- Create: `backend/internal/model/metric_chart.go`
- Modify: `backend/internal/model/migrate.go`

- [ ] **Step 1: Create MetricChart model**

```go
// backend/internal/model/metric_chart.go
package model

import "time"

// MetricChart 度量图表配置
type MetricChart struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID  uint64    `gorm:"not null;index" json:"project_id"`
	CreatorID  uint64    `gorm:"not null" json:"creator_id"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	TemplateID string    `gorm:"size:100" json:"template_id"`
	ChartType  string    `gorm:"size:50;not null" json:"chart_type"`
	XAxis      string    `gorm:"size:100;not null" json:"x_axis"`
	YAxis      string    `gorm:"size:100;not null" json:"y_axis"`
	Filters    string    `gorm:"type:jsonb;default:'{}'" json:"filters"`
	Config     string    `gorm:"type:jsonb;default:'{}'" json:"config"`
	SortOrder  int       `gorm:"default:0" json:"sort_order"`
	IsVisible  bool      `gorm:"default:true" json:"is_visible"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (MetricChart) TableName() string {
	return "metric_charts"
}
```

- [ ] **Step 2: Add to auto-migrate**

Read `backend/internal/model/migrate.go` to find the AutoMigrate call, then add `MetricChart{}` to it.

- [ ] **Step 3: Run backend build to verify**

Run: `cd backend/cmd/server && go build .`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add backend/internal/model/metric_chart.go backend/internal/model/migrate.go
git commit -m "feat(metrics): add MetricChart model and migration"
```

---

## Task 2: Backend — Metrics Service

**Files:**
- Create: `backend/internal/service/metric_service.go`

- [ ] **Step 1: Create MetricService with template definitions**

```go
// backend/internal/service/metric_service.go
package service

import (
	"encoding/json"
	"sort"

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

// Template represents a pre-built metrics template
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

// ListTemplates returns all pre-built template categories
func (s *MetricService) ListTemplates() []TemplateCategory {
	return []TemplateCategory{
		{
			ID:   "agile",
			Name: "敏捷效能",
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
			ID:   "project",
			Name: "项目管理",
			Templates: []MetricTemplate{
				{ID: "pm_throughput", Category: "project", Name: "需求吞吐量", Description: "展示单位时间内完成的需求数量", ChartType: "bar", DefaultXAxis: "completed_week", DefaultYAxis: "count", Icon: "package"},
				{ID: "pm_milestone", Category: "project", Name: "里程碑达成率", Description: "展示里程碑按时完成百分比", ChartType: "doughnut", DefaultXAxis: "state", DefaultYAxis: "count", Icon: "target"},
				{ID: "pm_defect_trend", Category: "project", Name: "缺陷趋势", Description: "展示缺陷新建与解决趋势", ChartType: "line", DefaultXAxis: "created_week", DefaultYAxis: "count", DefaultFilters: map[string]interface{}{"type_filter": "bug"}, Icon: "alert-triangle"},
				{ID: "pm_priority", Category: "project", Name: "优先级分布", Description: "展示各优先级Issue占比", ChartType: "pie", DefaultXAxis: "priority", DefaultYAxis: "count", Icon: "bar-chart-2"},
				{ID: "pm_workload", Category: "project", Name: "资源负载", Description: "展示每个成员的工作负载", ChartType: "bar", DefaultXAxis: "assignee", DefaultYAxis: "count", Icon: "users"},
			},
		},
		{
			ID:   "quality",
			Name: "质量分析",
			Templates: []MetricTemplate{
				{ID: "qa_density", Category: "quality", Name: "缺陷密度", Description: "展示各模块的缺陷密度", ChartType: "bar", DefaultXAxis: "module", DefaultYAxis: "count", DefaultFilters: map[string]interface{}{"type_filter": "bug"}, Icon: "search"},
				{ID: "qa_escape", Category: "quality", Name: "缺陷逃逸率", Description: "展示线上缺陷占总缺陷比例", ChartType: "doughnut", DefaultXAxis: "state", DefaultYAxis: "count", DefaultFilters: map[string]interface{}{"type_filter": "bug"}, Icon: "shield"},
				{ID: "qa_review_pass", Category: "quality", Name: "评审一次通过率", Description: "展示评审一次通过比例", ChartType: "doughnut", DefaultXAxis: "state", DefaultYAxis: "count", Icon: "check-circle"},
				{ID: "qa_state_dwell", Category: "quality", Name: "状态停留时间", Description: "展示Issue在各状态的平均停留时间", ChartType: "bar", DefaultXAxis: "state", DefaultYAxis: "current_retention", Icon: "pause-circle"},
			},
		},
	}
}

// ChartConfig represents the JSON config for a metric chart
type ChartConfig struct {
	StackMode     string               `json:"stack_mode,omitempty"`     // none | stack | percent_stack
	ReferenceLines []ReferenceLine     `json:"reference_lines,omitempty"`
	ShowLabels    bool                 `json:"show_labels"`
	DualYAxis     bool                 `json:"dual_y_axis"`
	LegendPosition string              `json:"legend_position,omitempty"` // top | bottom | left | right
}

type ReferenceLine struct {
	Type  string `json:"type"`  // constant | max | min | average | median | sum
	Value float64 `json:"value,omitempty"`
	Label string `json:"label,omitempty"`
	Color string `json:"color,omitempty"`
	Style string `json:"style,omitempty"` // solid | dashed | dotted
}

// CRUD operations

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

// RenderChart generates chart data for a given chart config
func (s *MetricService) RenderChart(projectID, chartID uint64) (*RenderResponse, error) {
	chart, err := s.GetChart(projectID, chartID)
	if err != nil {
		return nil, err
	}

	// Delegate to existing V2 report engine
	reportReq := &ReportV2Request{
		XAxis: chart.XAxis,
		YAxis: chart.YAxis,
	}

	// Parse filters for RQL
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

	// Parse config for reference lines
	var config ChartConfig
	if chart.Config != "" && chart.Config != "{}" {
		json.Unmarshal([]byte(chart.Config), &config)
	}

	// Calculate reference lines
	var refLines []ReferenceLineData
	for _, rl := range config.ReferenceLines {
		refLine := ReferenceLineData{
			Type:  rl.Type,
			Label: rl.Label,
			Color: rl.Color,
			Style: rl.Style,
		}
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
					if v > max { max = v }
				}
				refLine.Value = float64(max)
			}
		case "min":
			if len(reportResp.Values) > 0 {
				min := reportResp.Values[0]
				for _, v := range reportResp.Values[1:] {
					if v < min { min = v }
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
		Labels:         reportResp.Labels,
		Values:         reportResp.Values,
		Total:          reportResp.Total,
		Colors:         reportResp.Colors,
		ReferenceLines: refLines,
		ChartType:      chart.ChartType,
		Config:         config,
	}, nil
}

// Request/Response types

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
	Labels         []string           `json:"labels"`
	Values         []int              `json:"values"`
	Total          int                `json:"total"`
	Colors         map[string]string  `json:"colors"`
	ReferenceLines []ReferenceLineData `json:"reference_lines"`
	ChartType      string             `json:"chart_type"`
	Config         ChartConfig        `json:"config"`
}

type ReferenceLineData struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Label string  `json:"label"`
	Color string  `json:"color"`
	Style string  `json:"style"`
}
```

- [ ] **Step 2: Verify build**

Run: `cd backend/cmd/server && go build .`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/metric_service.go
git commit -m "feat(metrics): add MetricService with templates and CRUD"
```

---

## Task 3: Backend — Metrics Handler + Router

**Files:**
- Create: `backend/internal/handler/metric_handler.go`
- Modify: `backend/internal/router/router.go:251-256`

- [ ] **Step 1: Create MetricHandler**

```go
// backend/internal/handler/metric_handler.go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

type MetricHandler struct {
	svc *service.MetricService
}

func NewMetricHandler(svc *service.MetricService) *MetricHandler {
	return &MetricHandler{svc: svc}
}

func (h *MetricHandler) ListTemplates(c *gin.Context) {
	categories := h.svc.ListTemplates()
	common.RespondOK(c, gin.H{"categories": categories})
}

func (h *MetricHandler) ListCharts(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	charts, err := h.svc.ListCharts(pid)
	if err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, gin.H{"charts": charts, "total": len(charts)})
}

func (h *MetricHandler) GetChart(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	cid, _ := strconv.ParseUint(c.Param("chartId"), 10, 64)
	chart, err := h.svc.GetChart(pid, cid)
	if err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, chart)
}

func (h *MetricHandler) CreateChart(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	user := middleware.GetCurrentUser(c)
	var req service.CreateChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest(err.Error()))
		return
	}
	chart, err := h.svc.CreateChart(pid, user.ID, &req)
	if err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, chart)
}

func (h *MetricHandler) UpdateChart(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	cid, _ := strconv.ParseUint(c.Param("chartId"), 10, 64)
	var req service.UpdateChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest(err.Error()))
		return
	}
	chart, err := h.svc.UpdateChart(pid, cid, &req)
	if err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, chart)
}

func (h *MetricHandler) DeleteChart(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	cid, _ := strconv.ParseUint(c.Param("chartId"), 10, 64)
	if err := h.svc.DeleteChart(pid, cid); err != nil {
		common.RespondError(c, err); return
	}
	common.RespondOK(c, gin.H{"message": "deleted"})
}

func (h *MetricHandler) RenderChart(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	cid, _ := strconv.ParseUint(c.Param("chartId"), 10, 64)
	resp, err := h.svc.RenderChart(pid, cid)
	if err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, resp)
}

func (h *MetricHandler) ReorderCharts(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	var req struct {
		ChartIDs []uint64 `json:"chart_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest(err.Error()))
		return
	}
	if err := h.svc.ReorderCharts(pid, req.ChartIDs); err != nil {
		common.RespondError(c, err); return
	}
	common.RespondOK(c, gin.H{"message": "reordered"})
}
```

- [ ] **Step 2: Register routes in router.go**

Add after line 256 (after saved-reports routes) in `backend/internal/router/router.go`:

```go
// Metrics
metricH := handler.NewMetricHandler(metricSvc)
projects.GET("/:projectId/metrics/templates", metricH.ListTemplates)
projects.GET("/:projectId/metrics/charts", metricH.ListCharts)
projects.GET("/:projectId/metrics/charts/:chartId", metricH.GetChart)
projects.POST("/:projectId/metrics/charts", metricH.CreateChart)
projects.PUT("/:projectId/metrics/charts/:chartId", metricH.UpdateChart)
projects.DELETE("/:projectId/metrics/charts/:chartId", metricH.DeleteChart)
projects.POST("/:projectId/metrics/charts/:chartId/render", metricH.RenderChart)
projects.POST("/:projectId/metrics/charts/reorder", metricH.ReorderCharts)
```

Also ensure `metricSvc` is initialized in the service creation block (near where `reportSvc` is created).

- [ ] **Step 3: Verify build**

Run: `cd backend/cmd/server && go build .`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/metric_handler.go backend/internal/router/router.go
git commit -m "feat(metrics): add MetricHandler and register metrics API routes"
```

---

## Task 4: Frontend — Types + API Client

**Files:**
- Create: `frontend/src/types/metrics.ts`
- Create: `frontend/src/api/metrics.ts`

- [ ] **Step 1: Create TypeScript types**

```typescript
// frontend/src/types/metrics.ts

export interface MetricTemplate {
  id: string
  category: string
  name: string
  description: string
  chart_type: string
  default_x_axis: string
  default_y_axis: string
  default_filters?: Record<string, any>
  default_config?: Record<string, any>
  icon: string
}

export interface TemplateCategory {
  id: string
  name: string
  templates: MetricTemplate[]
}

export interface ReferenceLine {
  type: 'constant' | 'max' | 'min' | 'average' | 'median' | 'sum'
  value?: number
  label?: string
  color?: string
  style?: string
}

export interface MetricChartConfig {
  stack_mode?: string
  reference_lines?: ReferenceLine[]
  show_labels?: boolean
  dual_y_axis?: boolean
  legend_position?: string
}

export interface MetricChart {
  id: number
  project_id: number
  creator_id: number
  name: string
  template_id: string
  chart_type: string
  x_axis: string
  y_axis: string
  filters: string
  config: string
  sort_order: number
  is_visible: boolean
  created_at: string
  updated_at: string
}

export interface RenderResult {
  labels: string[]
  values: number[]
  total: number
  colors: Record<string, string>
  reference_lines: Array<{
    type: string
    value: number
    label: string
    color: string
    style: string
  }>
  chart_type: string
  config: MetricChartConfig
}

export interface CreateChartPayload {
  name: string
  template_id?: string
  chart_type: string
  x_axis: string
  y_axis: string
  filters?: Record<string, any>
  config?: MetricChartConfig
  sort_order?: number
}
```

- [ ] **Step 2: Create API client**

```typescript
// frontend/src/api/metrics.ts
import api from '@/api'

export const metricsApi = {
  listTemplates: async (projectId: number) => {
    const res = await api.get(`/projects/${projectId}/metrics/templates`)
    return res.data
  },

  listCharts: async (projectId: number) => {
    const res = await api.get(`/projects/${projectId}/metrics/charts`)
    return res.data
  },

  getChart: async (projectId: number, chartId: number) => {
    const res = await api.get(`/projects/${projectId}/metrics/charts/${chartId}`)
    return res.data
  },

  createChart: async (projectId: number, data: any) => {
    const res = await api.post(`/projects/${projectId}/metrics/charts`, data)
    return res.data
  },

  updateChart: async (projectId: number, chartId: number, data: any) => {
    const res = await api.put(`/projects/${projectId}/metrics/charts/${chartId}`, data)
    return res.data
  },

  deleteChart: async (projectId: number, chartId: number) => {
    const res = await api.delete(`/projects/${projectId}/metrics/charts/${chartId}`)
    return res.data
  },

  renderChart: async (projectId: number, chartId: number) => {
    const res = await api.post(`/projects/${projectId}/metrics/charts/${chartId}/render`)
    return res.data
  },

  reorderCharts: async (projectId: number, chartIds: number[]) => {
    const res = await api.post(`/projects/${projectId}/metrics/charts/reorder`, { chart_ids: chartIds })
    return res.data
  },
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/types/metrics.ts frontend/src/api/metrics.ts
git commit -m "feat(metrics): add TypeScript types and API client for metrics"
```

---

## Task 5: Frontend — MetricsChartConfig Component

**Files:**
- Create: `frontend/src/components/metrics/MetricsChartConfig.vue`

- [ ] **Step 1: Create chart configuration panel**

This is a dialog/panel component that lets users configure:
- Chart type selector (bar/line/pie/doughnut/area/table)
- X axis dimension selector
- Y axis metric selector
- Filter conditions
- Stack mode (none/stack/percent)
- Show labels toggle
- Reference lines (basic: average, constant)

The component receives props: `projectId`, `template?` (pre-filled values), `chart?` (existing chart for edit mode), and emits `save` and `cancel`.

Key implementation notes:
- Use the existing filter options loading pattern from `loadFilterOptions()` in ReportBuilder.vue (lines 554-565)
- X axis options: the same list from the current ReportBuilder V2 (state, priority, assignee, type, label, cycle, module, state_group, created_by, created_day/week/month, completed_day/week/month, updated_day/week/month)
- Y axis options: count, avg_processing_time, current_retention, created_vs_resolved
- Include a live preview area that renders the chart using Chart.js

- [ ] **Step 2: Verify build**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 | head -20`
Expected: No type errors related to new files

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/metrics/MetricsChartConfig.vue
git commit -m "feat(metrics): add MetricsChartConfig panel component"
```

---

## Task 6: Frontend — MetricsChartCard + MetricsTemplateGallery

**Files:**
- Create: `frontend/src/components/metrics/MetricsChartCard.vue`
- Create: `frontend/src/components/metrics/MetricsTemplateGallery.vue`

- [ ] **Step 1: Create MetricsChartCard**

A card component that:
- Shows chart title, type switcher (bar/line/pie), edit/delete/fullscreen buttons in toolbar
- Renders the chart using Chart.js (follow existing pattern from useReportChart.ts)
- Shows loading state while fetching data
- Emits `edit`, `delete`, `fullscreen`, `type-change` events

Props: `chart: MetricChart`, `projectId: number`
Data: fetches render data on mount via `metricsApi.renderChart()`

- [ ] **Step 2: Create MetricsTemplateGallery**

A grid of template cards grouped by category:
- Tab-like category selector (敏捷效能 / 项目管理 / 质量分析)
- Grid of template cards, each showing icon, name, description
- Click a template → emits `use-template` event with the template data

Props: `categories: TemplateCategory[]`
Emits: `use-template(template: MetricTemplate)`

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/metrics/MetricsChartCard.vue frontend/src/components/metrics/MetricsTemplateGallery.vue
git commit -m "feat(metrics): add MetricsChartCard and MetricsTemplateGallery components"
```

---

## Task 7: Frontend — MetricsView (Main Page)

**Files:**
- Create: `frontend/src/views/MetricsView.vue`

- [ ] **Step 1: Create MetricsView**

The main metrics page that composes all sub-components:

```
┌──────────────────────────────────────────────┐
│  Tabs: [预置模板] [我的图表]                    │
├──────────────────────────────────────────────┤
│                                              │
│  (预置模板 tab)                               │
│  MetricsTemplateGallery                      │
│                                              │
│  (我的图表 tab)                               │
│  [+ 新建图表]                                 │
│  ┌────────────┐ ┌────────────┐              │
│  │ ChartCard  │ │ ChartCard  │ ...          │
│  └────────────┘ └────────────┘              │
│                                              │
└──────────────────────────────────────────────┘

MetricsChartConfig dialog (shown when creating/editing)
```

Key logic:
- On mount: fetch templates (`metricsApi.listTemplates`) and charts (`metricsApi.listCharts`)
- Template click → open MetricsChartConfig with template defaults
- "新建图表" → open MetricsChartConfig with empty config
- Chart card edit → open MetricsChartConfig with chart data
- Save → create or update chart, refresh list
- Delete → confirm → delete chart, refresh list
- Drag reorder → update sort_order via `metricsApi.reorderCharts()`

- [ ] **Step 2: Verify full build**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 | head -20`
Expected: No type errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/MetricsView.vue
git commit -m "feat(metrics): add MetricsView main page"
```

---

## Task 8: Integrate MetricsView into Project.vue

**Files:**
- Modify: `frontend/src/views/Project.vue`

- [ ] **Step 1: Replace ReportBuilder with MetricsView**

In `Project.vue`:

1. Change import (line 309):
```typescript
// OLD:
import ReportBuilder from '@/components/ReportBuilder.vue'
// NEW:
import MetricsView from '@/views/MetricsView.vue'
```

2. Change template (line 184-186):
```vue
<!-- OLD: -->
<div v-if="activeTab === 'reports'">
  <ReportBuilder :project-id="projectId" />
</div>
<!-- NEW: -->
<div v-if="activeTab === 'reports'">
  <MetricsView :project-id="projectId" />
</div>
```

3. Update tab label in computedTabs (line 524) to show "度量" instead of "报表":
```typescript
{ id: 'reports', name: t('project.tab.metrics') },
```

- [ ] **Step 2: Add i18n keys**

In `en-US.json`, add under `project.tab`:
```json
"metrics": "Metrics"
```

In `zh-CN.json`, add under `project.tab`:
```json
"metrics": "度量"
```

Also add the `metrics` section with all template names, chart config labels, etc.

- [ ] **Step 3: Verify build**

Run: `cd frontend && npm run build 2>&1 | tail -5`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/Project.vue frontend/src/locales/en-US.json frontend/src/locales/zh-CN.json
git commit -m "feat(metrics): integrate MetricsView into Project, replace ReportBuilder"
```

---

## Task 9: End-to-End Smoke Test

- [ ] **Step 1: Start backend**

Run: `cd backend/cmd/server && go run .`
Expected: Server starts on port 8000

- [ ] **Step 2: Start frontend**

Run: `cd frontend && npm run dev`
Expected: Dev server starts on port 5173

- [ ] **Step 3: Manual smoke test**

1. Navigate to a project → click "度量" tab
2. Verify template gallery shows 3 categories with 16 templates
3. Click a template → config panel opens with pre-filled values
4. Save as custom chart → appears in "我的图表"
5. Edit chart → change type → verify update
6. Delete chart → verify removal
7. Drag reorder → verify sort persists after refresh

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat(metrics): Phase 1 MVP complete — metrics mode with 16 templates and custom charts"
```
