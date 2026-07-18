package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/agent-service/internal/common"
	"github.com/reqmango/agent-service/internal/middleware"
	"github.com/reqmango/agent-service/internal/service"
)

type AgentLoopHandler struct {
	svc *service.LoopService
}

func NewAgentLoopHandler(svc *service.LoopService) *AgentLoopHandler {
	return &AgentLoopHandler{svc: svc}
}

func (h *AgentLoopHandler) getWSAndToken(c *gin.Context) (uint64, uint64, string, error) {
	wsID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if err != nil {
		return 0, 0, "", err
	}
	user := middleware.GetCurrentUser(c)
	if user == nil {
		return 0, 0, "", common.NotFound("user not found")
	}
	token := c.GetHeader("Authorization")
	if len(token) > 7 {
		token = token[7:] // strip "Bearer "
	}
	return wsID, user.ID, token, nil
}

func (h *AgentLoopHandler) Create(c *gin.Context) {
	wsID, userID, _, err := h.getWSAndToken(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid workspace"})
		return
	}
	var req struct {
		Name        string          `json:"name" binding:"required"`
		Description string          `json:"description"`
		LoopDef     json.RawMessage `json:"loop_def" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	loop, err := h.svc.CreateLoop(wsID, userID, req.Name, req.Description, req.LoopDef)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	c.JSON(201, loop)
}

func (h *AgentLoopHandler) List(c *gin.Context) {
	wsID, _, _, err := h.getWSAndToken(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid workspace"})
		return
	}
	loops, err := h.svc.ListLoops(wsID)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, loops)
}

func (h *AgentLoopHandler) Get(c *gin.Context) {
	wsID, _, _, err := h.getWSAndToken(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid workspace"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	loop, err := h.svc.GetLoop(wsID, id)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(404, gin.H{"message": "loop not found"})
		return
	}
	c.JSON(200, loop)
}

func (h *AgentLoopHandler) Update(c *gin.Context) {
	wsID, _, _, err := h.getWSAndToken(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid workspace"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name        *string         `json:"name"`
		Description *string         `json:"description"`
		LoopDef     json.RawMessage `json:"loop_def"`
		Status      *string         `json:"status"`
	}
	c.ShouldBindJSON(&req)
	loop, err := h.svc.UpdateLoop(wsID, id, req.Name, req.Description, req.LoopDef, req.Status)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, loop)
}

func (h *AgentLoopHandler) Delete(c *gin.Context) {
	wsID, _, _, err := h.getWSAndToken(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid workspace"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteLoop(wsID, id); err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(404, gin.H{"message": err.Error()})
		return
	}
	c.JSON(204, nil)
}

func (h *AgentLoopHandler) Start(c *gin.Context) {
	wsID, userID, token, err := h.getWSAndToken(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid workspace"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	run, err := h.svc.StartLoop(wsID, id, userID, token)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, run)
}

func (h *AgentLoopHandler) Stop(c *gin.Context) {
	wsID, _, _, err := h.getWSAndToken(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid workspace"})
		return
	}
	runID, _ := strconv.ParseUint(c.Param("runId"), 10, 64)
	if err := h.svc.StopLoop(wsID, runID); err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(404, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "loop stopped"})
}

func (h *AgentLoopHandler) GetRuns(c *gin.Context) {
	wsID, _, _, err := h.getWSAndToken(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid workspace"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	runs, err := h.svc.GetLoopRuns(wsID, id, limit)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, runs)
}

func (h *AgentLoopHandler) GetRun(c *gin.Context) {
	wsID, _, _, err := h.getWSAndToken(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid workspace"})
		return
	}
	runID, _ := strconv.ParseUint(c.Param("runId"), 10, 64)
	run, iterations, err := h.svc.GetLoopRun(wsID, runID)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(404, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"run": run, "iterations": iterations})
}
