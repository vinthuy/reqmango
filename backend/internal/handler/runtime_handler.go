package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/service"
)

type RuntimeHandler struct{ svc *service.RuntimeService }

func NewRuntimeHandler(svc *service.RuntimeService) *RuntimeHandler {
	return &RuntimeHandler{svc: svc}
}

func (h *RuntimeHandler) respond(c *gin.Context, err error) bool {
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

func (h *RuntimeHandler) parseWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

func (h *RuntimeHandler) CreateRuntime(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	var req request.RuntimeCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Create(wid, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(201, resp)
}

func (h *RuntimeHandler) GetRuntime(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("runtimeId"), 10, 64)
	resp, e := h.svc.Get(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *RuntimeHandler) ListRuntimes(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	resp, e := h.svc.List(wid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *RuntimeHandler) UpdateRuntime(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("runtimeId"), 10, 64)
	var req request.RuntimeUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Update(id, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *RuntimeHandler) DeleteRuntime(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("runtimeId"), 10, 64)
	if h.respond(c, h.svc.Delete(id)) {
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}

func (h *RuntimeHandler) RegisterRuntime(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	var req request.RuntimeCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Register(wid, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *RuntimeHandler) Heartbeat(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("runtimeId"), 10, 64)
	var req request.RuntimeHeartbeat
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Heartbeat(id, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *RuntimeHandler) FindAvailableRuntime(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	resp, e := h.svc.FindAvailable(wid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// HealthCheck performs health check for all runtimes in a workspace.
func (h *RuntimeHandler) HealthCheck(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	if h.respond(c, h.svc.PerformHealthCheck(wid)) {
		return
	}
	c.JSON(200, gin.H{"message": "Health check completed"})
}

// GlobalHealthCheck performs health check for all runtimes across all workspaces.
func (h *RuntimeHandler) GlobalHealthCheck(c *gin.Context) {
	if h.respond(c, h.svc.PerformGlobalHealthCheck()) {
		return
	}
	c.JSON(200, gin.H{"message": "Global health check completed"})
}

// ScheduleTask finds an available runtime and schedules a task.
func (h *RuntimeHandler) ScheduleTask(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	runtimeID, e := h.svc.ScheduleTask(wid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, gin.H{"runtime_id": runtimeID})
}

// ReleaseTask releases a runtime from a completed task.
func (h *RuntimeHandler) ReleaseTask(c *gin.Context) {
	runtimeID, _ := strconv.ParseUint(c.Param("runtimeId"), 10, 64)
	if h.respond(c, h.svc.ReleaseTask(runtimeID)) {
		return
	}
	c.JSON(200, gin.H{"message": "Task released from runtime"})
}

// GetRuntimeStats returns runtime statistics for a workspace.
func (h *RuntimeHandler) GetRuntimeStats(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	stats, e := h.svc.GetRuntimeStats(wid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, stats)
}
