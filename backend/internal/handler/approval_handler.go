package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

type ApprovalHandler struct{ svc *service.ApprovalService }

func NewApprovalHandler(svc *service.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{svc: svc}
}

func (h *ApprovalHandler) respond(c *gin.Context, err error) bool {
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

// Create: POST /api/v1/issues/:issueId/approvals
func (h *ApprovalHandler) Create(c *gin.Context) {
	issueID, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	var req request.ApprovalCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	userID := middleware.GetUserID(c)
	resp, e := h.svc.Create(issueID, userID, req.TransitionID, req.RequestNote)
	if h.respond(c, e) {
		return
	}
	c.JSON(201, resp)
}

// ListByWorkspace: GET /api/v1/workspaces/:wsParam/approvals
func (h *ApprovalHandler) ListByWorkspace(c *gin.Context) {
	wid, _ := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	var q request.ApprovalListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(400, gin.H{"message": "Invalid query"})
		return
	}
	list, e := h.svc.List(q, wid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, list)
}

// ListByProject: GET /api/v1/projects/:projectId/approvals?workspace_id=
func (h *ApprovalHandler) ListByProject(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	wid, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	var q request.ApprovalListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(400, gin.H{"message": "Invalid query"})
		return
	}
	q.ProjectID = pid
	list, e := h.svc.List(q, wid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, list)
}

// Get: GET /api/v1/approvals/:id
func (h *ApprovalHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	resp, e := h.svc.Get(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// Decide: POST /api/v1/approvals/:id/decide
func (h *ApprovalHandler) Decide(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req request.ApprovalDecision
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	userID := middleware.GetUserID(c)
	resp, e := h.svc.Decide(id, userID, req.Decision, req.Note)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// Cancel: POST /api/v1/approvals/:id/cancel
func (h *ApprovalHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := middleware.GetUserID(c)
	resp, e := h.svc.Cancel(id, userID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// CountPending: GET /api/v1/workspaces/:wsParam/approvals/count
func (h *ApprovalHandler) CountPending(c *gin.Context) {
	wid, _ := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	userID := middleware.GetUserID(c)
	count, e := h.svc.CountPending(wid, userID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, gin.H{"pending_count": count})
}
