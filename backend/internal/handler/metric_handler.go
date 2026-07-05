package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

type MetricHandler struct{ svc *service.MetricService }

func NewMetricHandler(svc *service.MetricService) *MetricHandler {
	return &MetricHandler{svc: svc}
}

// ListTemplates returns all pre-built template categories.
func (h *MetricHandler) ListTemplates(c *gin.Context) {
	categories := h.svc.ListTemplates()
	common.RespondOK(c, categories)
}

// ListCharts lists charts for a project.
func (h *MetricHandler) ListCharts(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	charts, err := h.svc.ListCharts(pid)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, charts)
}

// GetChart gets a single chart.
func (h *MetricHandler) GetChart(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	cid, _ := strconv.ParseUint(c.Param("chartId"), 10, 64)
	chart, err := h.svc.GetChart(pid, cid)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, chart)
}

// CreateChart creates a new chart (needs auth user).
func (h *MetricHandler) CreateChart(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	user := middleware.GetCurrentUser(c)
	var req service.CreateChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest(err.Error()))
		return
	}
	chart, err := h.svc.CreateChart(pid, user.ID, &req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, chart)
}

// UpdateChart updates an existing chart.
func (h *MetricHandler) UpdateChart(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	cid, _ := strconv.ParseUint(c.Param("chartId"), 10, 64)
	var req service.UpdateChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest(err.Error()))
		return
	}
	chart, err := h.svc.UpdateChart(pid, cid, &req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, chart)
}

// DeleteChart deletes a chart.
func (h *MetricHandler) DeleteChart(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	cid, _ := strconv.ParseUint(c.Param("chartId"), 10, 64)
	if err := h.svc.DeleteChart(pid, cid); err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, gin.H{"message": "Deleted"})
}

// RenderChart renders chart data.
func (h *MetricHandler) RenderChart(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	cid, _ := strconv.ParseUint(c.Param("chartId"), 10, 64)
	result, err := h.svc.RenderChart(pid, cid)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, result)
}

// ReorderCharts reorders charts.
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
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, gin.H{"message": "Charts reordered"})
}
