package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/service"
)

// AgentPerformanceHandler exposes agent task performance analytics endpoints
// (PRD P4-010: 执行效率统计、成功率分析、报表).
type AgentPerformanceHandler struct{ svc *service.AgentPerformanceService }

func NewAgentPerformanceHandler(svc *service.AgentPerformanceService) *AgentPerformanceHandler {
	return &AgentPerformanceHandler{svc: svc}
}

func (h *AgentPerformanceHandler) respond(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*common.AppError); ok {
		c.JSON(ae.Code, gin.H{"message": ae.Message})
		return true
	}
	c.JSON(500, gin.H{"message": "Internal server error"})
	return true
}

func (h *AgentPerformanceHandler) parseWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

// Overview returns aggregated performance metrics for the workspace.
// Query params: from, to (RFC3339) — optional period bounds.
func (h *AgentPerformanceHandler) Overview(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	resp, e := h.svc.Overview(wid, c.Query("from"), c.Query("to"))
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// ByTemplate returns per-template performance breakdown.
// Query params: from, to (RFC3339) — optional period bounds.
func (h *AgentPerformanceHandler) ByTemplate(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	resp, e := h.svc.ByTemplate(wid, c.Query("from"), c.Query("to"))
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// Timeline returns time-bucketed metrics for trend charts.
// Query params: bucket (day|week|month, default day), from, to (RFC3339).
func (h *AgentPerformanceHandler) Timeline(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	resp, e := h.svc.Timeline(wid, c.Query("bucket"), c.Query("from"), c.Query("to"))
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// FailureBreakdown returns failure_reason tallies for failed tasks.
// Query params: from, to (RFC3339) — optional period bounds.
func (h *AgentPerformanceHandler) FailureBreakdown(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	resp, e := h.svc.FailureBreakdown(wid, c.Query("from"), c.Query("to"))
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}
