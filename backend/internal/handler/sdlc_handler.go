package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

// SDLCHandler exposes the SDLC orchestration HTTP endpoints
// (PRD P4-006: 完整 SDLC 流程编排引擎).
//
// Routes are grouped under /workspaces/:wsParam/sdlc:
//
//	GET    /workflows                          list workflows (optional ?status=, ?limit=)
//	POST   /workflows                          create + start a workflow
//	GET    /workflows/:workflowId              get workflow (with stages)
//	POST   /workflows/:workflowId/cancel       cancel a running/pending workflow
//	DELETE /workflows/:workflowId              delete a workflow
//	POST   /workflows/:workflowId/retry        resume from a failed stage (body: stage_id)
//	GET    /workflows/:workflowId/stages       list stages of a workflow
//	GET    /workflows/:workflowId/stages/:stageId  get a single stage
type SDLCHandler struct {
	svc *service.SDLCService
}

func NewSDLCHandler(svc *service.SDLCService) *SDLCHandler {
	return &SDLCHandler{svc: svc}
}

func (h *SDLCHandler) respond(c *gin.Context, err error) bool {
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

func (h *SDLCHandler) parseWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

// ======== Workflow endpoints ========

// ListWorkflows handles GET /workspaces/:wsParam/sdlc/workflows
// Query params: status (optional), limit (optional)
func (h *SDLCHandler) ListWorkflows(c *gin.Context) {
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

// GetWorkflow handles GET /workspaces/:wsParam/sdlc/workflows/:workflowId
func (h *SDLCHandler) GetWorkflow(c *gin.Context) {
	wfID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workflow id"})
		return
	}
	resp, e := h.svc.Get(wfID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// CreateWorkflow handles POST /workspaces/:wsParam/sdlc/workflows
// Starts a new SDLC pipeline run. The orchestration runs asynchronously;
// the response returns the pending workflow so the client can poll or
// subscribe to SSE events (sdlc_workflow.updated / sdlc_stage.updated).
func (h *SDLCHandler) CreateWorkflow(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	var req service.SDLCWorkflowCreate
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

// CancelWorkflow handles POST /workspaces/:wsParam/sdlc/workflows/:workflowId/cancel
func (h *SDLCHandler) CancelWorkflow(c *gin.Context) {
	wfID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workflow id"})
		return
	}
	resp, e := h.svc.Cancel(wfID, middleware.GetCurrentUser(c).ID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// DeleteWorkflow handles DELETE /workspaces/:wsParam/sdlc/workflows/:workflowId
func (h *SDLCHandler) DeleteWorkflow(c *gin.Context) {
	wfID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workflow id"})
		return
	}
	if e := h.svc.Delete(wfID, middleware.GetCurrentUser(c).ID); h.respond(c, e) {
		return
	}
	c.JSON(204, gin.H{"message": "deleted"})
}

// RetryWorkflow handles POST /workspaces/:wsParam/sdlc/workflows/:workflowId/retry
// Body: {"stage_id": <uint64>} — resumes the pipeline from the given stage.
func (h *SDLCHandler) RetryWorkflow(c *gin.Context) {
	wfID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workflow id"})
		return
	}
	var body struct {
		StageID uint64 `json:"stage_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"message": "stage_id is required"})
		return
	}
	if body.StageID == 0 {
		c.JSON(400, gin.H{"message": "stage_id is required"})
		return
	}
	resp, e := h.svc.RetryFromStage(wfID, body.StageID, middleware.GetCurrentUser(c).ID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// ======== Stage endpoints ========

// ListStages handles GET /workspaces/:wsParam/sdlc/workflows/:workflowId/stages
func (h *SDLCHandler) ListStages(c *gin.Context) {
	wfID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workflow id"})
		return
	}
	resp, e := h.svc.ListStages(wfID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// GetStage handles GET /workspaces/:wsParam/sdlc/workflows/:workflowId/stages/:stageId
func (h *SDLCHandler) GetStage(c *gin.Context) {
	wfID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workflow id"})
		return
	}
	stageID, err := strconv.ParseUint(c.Param("stageId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid stage id"})
		return
	}
	resp, e := h.svc.GetStage(wfID, stageID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}
