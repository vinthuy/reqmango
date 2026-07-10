package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/service"
)

type WorkflowHandler struct{ svc *service.WorkflowService }
func NewWorkflowHandler(svc *service.WorkflowService) *WorkflowHandler { return &WorkflowHandler{svc: svc} }
func (h *WorkflowHandler) respond(c *gin.Context, err error) bool {
	if err == nil { return false }
	if ae, ok := err.(*common.AppError); ok { c.JSON(ae.Code, gin.H{"message": ae.Message}); return true }
	c.JSON(500, gin.H{"message": "Internal server error"}); return true
}
func (h *WorkflowHandler) parseProjectID(c *gin.Context) (uint64, error) { return strconv.ParseUint(c.Param("projectId"), 10, 64) }

// ---- Workflow routes ----
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	pid, _ := h.parseProjectID(c)
	var req request.WorkflowCreate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.Create(pid, req)
	if h.respond(c, e) { return }; c.JSON(201, resp)
}
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	pid, _ := h.parseProjectID(c)
	resp, e := h.svc.List(pid)
	if h.respond(c, e) { return }; c.JSON(200, resp)
}
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	resp, e := h.svc.Get(id)
	if h.respond(c, e) { return }; c.JSON(200, resp)
}
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	var req request.WorkflowUpdate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.Update(id, req)
	if h.respond(c, e) { return }; c.JSON(200, resp)
}
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if h.respond(c, h.svc.Delete(id)) { return }; c.JSON(200, gin.H{"message":"Deleted"})
}

// ---- Workspace-level Workflow routes ----
func (h *WorkflowHandler) parseWorkspaceID(c *gin.Context) (uint64, error) { return strconv.ParseUint(c.Param("wsParam"), 10, 64) }
func (h *WorkflowHandler) CreateWorkspaceWorkflow(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	var req request.WorkflowCreate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.CreateWorkspace(wid, req)
	if h.respond(c, e) { return }; c.JSON(201, resp)
}
func (h *WorkflowHandler) ListWorkspaceWorkflows(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	resp, e := h.svc.ListWorkspace(wid)
	if h.respond(c, e) { return }; c.JSON(200, resp)
}
func (h *WorkflowHandler) UpdateWorkspaceWorkflow(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	var req request.WorkflowUpdate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.UpdateWorkspace(id, req)
	if h.respond(c, e) { return }; c.JSON(200, resp)
}
func (h *WorkflowHandler) DeleteWorkspaceWorkflow(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if h.respond(c, h.svc.DeleteWorkspace(id)) { return }; c.JSON(200, gin.H{"message":"Deleted"})
}

// ---- Transition routes ----
func (h *WorkflowHandler) AddTransition(c *gin.Context) {
	wid, _ := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	var req request.TransitionCreate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.AddTransition(wid, req)
	if h.respond(c, e) { return }; c.JSON(201, resp)
}
func (h *WorkflowHandler) UpdateTransition(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("transitionId"), 10, 64)
	var req request.TransitionUpdate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.UpdateTransition(id, req)
	if h.respond(c, e) { return }; c.JSON(200, resp)
}
func (h *WorkflowHandler) DeleteTransition(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("transitionId"), 10, 64)
	if h.respond(c, h.svc.DeleteTransition(id)) { return }; c.JSON(200, gin.H{"message":"Deleted"})
}

// ---- Automation routes ----
func (h *WorkflowHandler) CreateAutomation(c *gin.Context) {
	pid, _ := h.parseProjectID(c)
	var req request.AutomationCreate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.CreateAutomation(pid, req)
	if h.respond(c, e) { return }; c.JSON(201, resp)
}
func (h *WorkflowHandler) ListAutomations(c *gin.Context) {
	pid, _ := h.parseProjectID(c)
	resp, e := h.svc.ListAutomations(pid)
	if h.respond(c, e) { return }; c.JSON(200, resp)
}
func (h *WorkflowHandler) UpdateAutomation(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req request.AutomationUpdate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.UpdateAutomation(id, req)
	if h.respond(c, e) { return }; c.JSON(200, resp)
}
func (h *WorkflowHandler) DeleteAutomation(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if h.respond(c, h.svc.DeleteAutomation(id)) { return }; c.JSON(200, gin.H{"message":"Deleted"})
}
