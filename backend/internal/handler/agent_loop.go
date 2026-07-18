package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

type AgentLoopHandler struct {
	svc *service.LoopService
}

func NewAgentLoopHandler(svc *service.LoopService) *AgentLoopHandler {
	return &AgentLoopHandler{svc: svc}
}

func (h *AgentLoopHandler) getWSAndUser(c *gin.Context) (uint64, uint64, error) {
	wsID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	user := middleware.GetCurrentUser(c)
	if user == nil {
		return 0, 0, common.NotFound("user not found")
	}
	return wsID, user.ID, nil
}

func (h *AgentLoopHandler) Create(c *gin.Context) {
	wsID, userID, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}

	var req struct {
		Name        string          `json:"name" binding:"required"`
		Description string          `json:"description"`
		LoopDef     json.RawMessage `json:"loop_def" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	loop, err := h.svc.CreateLoop(wsID, userID, req.Name, req.Description, req.LoopDef)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, loop)
}

func (h *AgentLoopHandler) List(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loops, err := h.svc.ListLoops(wsID)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loops)
}

func (h *AgentLoopHandler) Get(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid loop id"})
		return
	}
	loop, err := h.svc.GetLoop(wsID, loopID)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loop)
}

func (h *AgentLoopHandler) Update(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid loop id"})
		return
	}

	var req struct {
		Name        *string         `json:"name"`
		Description *string         `json:"description"`
		LoopDef     json.RawMessage `json:"loop_def"`
		Status      *string         `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	loop, err := h.svc.UpdateLoop(wsID, loopID, req.Name, req.Description, req.LoopDef, req.Status)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loop)
}

func (h *AgentLoopHandler) Delete(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid loop id"})
		return
	}
	if err := h.svc.DeleteLoop(wsID, loopID); err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *AgentLoopHandler) Start(c *gin.Context) {
	wsID, userID, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid loop id"})
		return
	}
	run, err := h.svc.StartLoop(wsID, loopID, userID)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *AgentLoopHandler) Stop(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	runID, err := strconv.ParseUint(c.Param("runId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid run id"})
		return
	}
	if err := h.svc.StopLoop(wsID, runID); err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "loop stopped"})
}

func (h *AgentLoopHandler) GetRuns(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	loopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid loop id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	runs, err := h.svc.GetLoopRuns(wsID, loopID, limit)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, runs)
}

func (h *AgentLoopHandler) GetRun(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	runID, err := strconv.ParseUint(c.Param("runId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid run id"})
		return
	}
	run, iterations, err := h.svc.GetLoopRun(wsID, runID)
	if err != nil {
		if ae, ok := err.(*common.AppError); ok {
			c.JSON(ae.Code, gin.H{"message": ae.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"run": run, "iterations": iterations})
}
