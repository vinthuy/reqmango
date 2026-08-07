package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

// DeveloperAgentHandler exposes the Developer Agent HTTP endpoints
// (PRD P4-001: 代码生成、提交仓库、创建 PR).
type DeveloperAgentHandler struct {
	svc *service.DeveloperAgentService
}

func NewDeveloperAgentHandler(svc *service.DeveloperAgentService) *DeveloperAgentHandler {
	return &DeveloperAgentHandler{svc: svc}
}

func (h *DeveloperAgentHandler) respond(c *gin.Context, err error) bool {
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

func (h *DeveloperAgentHandler) parseWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

// List handles GET /workspaces/:wsParam/developer-agent/jobs
// Query params: status (optional), limit (optional, default 50, max 200)
func (h *DeveloperAgentHandler) List(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	resp, e := h.svc.List(wid, c.Query("status"), limit)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// Get handles GET /workspaces/:wsParam/developer-agent/jobs/:jobId
func (h *DeveloperAgentHandler) Get(c *gin.Context) {
	jobID, err := strconv.ParseUint(c.Param("jobId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid job id"})
		return
	}
	resp, e := h.svc.Get(jobID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// Create handles POST /workspaces/:wsParam/developer-agent/jobs
// Starts a new Developer Agent job. The workflow runs asynchronously; the
// response returns the pending job so the client can poll or subscribe to
// SSE events (developer_job.updated) for progress.
func (h *DeveloperAgentHandler) Create(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	var req service.DeveloperJobCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	resp, e := h.svc.Create(wid, middleware.GetCurrentUser(c).ID, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(201, resp)
}

// Cancel handles POST /workspaces/:wsParam/developer-agent/jobs/:jobId/cancel
func (h *DeveloperAgentHandler) Cancel(c *gin.Context) {
	jobID, err := strconv.ParseUint(c.Param("jobId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid job id"})
		return
	}
	resp, e := h.svc.Cancel(jobID, middleware.GetCurrentUser(c).ID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// Delete handles DELETE /workspaces/:wsParam/developer-agent/jobs/:jobId
func (h *DeveloperAgentHandler) Delete(c *gin.Context) {
	jobID, err := strconv.ParseUint(c.Param("jobId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid job id"})
		return
	}
	if e := h.svc.Delete(jobID, middleware.GetCurrentUser(c).ID); h.respond(c, e) {
		return
	}
	c.JSON(204, gin.H{"message": "deleted"})
}
