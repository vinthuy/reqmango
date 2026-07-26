package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/ai/common"
	"github.com/reqmango/backend/internal/ai/harness"
	"github.com/reqmango/backend/internal/ai/registry"
	"github.com/reqmango/backend/internal/ai/service"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type AgentPipelineHandler struct {
	db       *gorm.DB
	registry *registry.Registry
	agentSvc *service.AgentService
}

func NewAgentPipelineHandler(db *gorm.DB, reg *registry.Registry, agentSvc *service.AgentService) *AgentPipelineHandler {
	return &AgentPipelineHandler{db: db, registry: reg, agentSvc: agentSvc}
}

func (h *AgentPipelineHandler) getWSAndUser(c *gin.Context) (uint64, uint64, error) {
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

func (h *AgentPipelineHandler) Create(c *gin.Context) {
	wsID, userID, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	var req struct {
		Name        string          `json:"name" binding:"required"`
		Description string          `json:"description"`
		PipelineDef json.RawMessage `json:"pipeline_def" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	p := &model.Pipeline{
		WorkspaceID: wsID,
		Name:        req.Name,
		PipelineDef: req.PipelineDef,
		CreatedByID: &userID,
		Status:      "active",
	}
	if req.Description != "" {
		p.Description = &req.Description
	}
	if err := h.db.Create(p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *AgentPipelineHandler) List(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	var pipelines []model.Pipeline
	h.db.Where("workspace_id = ? AND status != ?", wsID, "archived").Order("created_at DESC").Find(&pipelines)
	c.JSON(http.StatusOK, pipelines)
}

func (h *AgentPipelineHandler) Get(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var p model.Pipeline
	if err := h.db.Where("id = ? AND workspace_id = ?", id, wsID).First(&p).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "pipeline not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *AgentPipelineHandler) Update(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name        *string         `json:"name"`
		Description *string         `json:"description"`
		PipelineDef json.RawMessage `json:"pipeline_def"`
		Status      *string         `json:"status"`
	}
	c.ShouldBindJSON(&req)
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.PipelineDef != nil {
		updates["pipeline_def"] = req.PipelineDef
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	h.db.Model(&model.Pipeline{}).Where("id = ? AND workspace_id = ?", id, wsID).Updates(updates)
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AgentPipelineHandler) Delete(c *gin.Context) {
	wsID, _, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	h.db.Where("id = ? AND workspace_id = ?", id, wsID).Delete(&model.Pipeline{})
	c.JSON(http.StatusNoContent, nil)
}

func (h *AgentPipelineHandler) Run(c *gin.Context) {
	wsID, userID, err := h.getWSAndUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var p model.Pipeline
	if err := h.db.Where("id = ? AND workspace_id = ?", id, wsID).First(&p).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "pipeline not found"})
		return
	}

	// Parse YAML pipeline def
	dsl, err := harness.ParsePipelineDSL(p.PipelineDef)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid pipeline definition: " + err.Error()})
		return
	}

	config, err := dsl.ToPipelineConfig()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "config error: " + err.Error()})
		return
	}

	// Build AgentCaller adapter using real AgentService
	caller := &agentServiceCaller{agentSvc: h.agentSvc, workspaceID: wsID, userID: userID, db: h.db}
	runner := harness.NewPipelineRunner(caller)

	// Create run record
	run := &model.PipelineRun{
		PipelineID:  p.ID,
		TriggerType: "manual",
		Status:      "running",
	}
	if err := h.db.Create(run).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Execute in background
	go func() {
		ctx := context.Background()
		results, runErr := runner.Run(ctx, *config, map[string]interface{}{
			"workspace_id": wsID,
			"user_id":      userID,
		})

		// Update run record
		resultsJSON, _ := json.Marshal(results)
		run.StagesResult = resultsJSON
		totalTokens := 0
		totalCost := 0.0
		for _, r := range results {
			totalTokens += r.TokensUsed
			totalCost += r.Cost
		}
		run.TokensUsed = totalTokens
		run.CostUSD = totalCost

		if runErr != nil {
			run.Status = "failed"
			msg := runErr.Error()
			run.ErrorMessage = &msg
		} else {
			run.Status = "completed"
		}
		h.db.Save(run)
	}()

	c.JSON(http.StatusOK, run)
}

func (h *AgentPipelineHandler) GetRuns(c *gin.Context) {
	_, _, _ = h.getWSAndUser(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	var runs []model.PipelineRun
	h.db.Where("pipeline_id = ?", id).Order("created_at DESC").Limit(limit).Find(&runs)
	c.JSON(http.StatusOK, runs)
}

func (h *AgentPipelineHandler) GetRun(c *gin.Context) {
	_, _, _ = h.getWSAndUser(c)
	runID, _ := strconv.ParseUint(c.Param("runId"), 10, 64)
	var run model.PipelineRun
	if err := h.db.First(&run, runID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "run not found"})
		return
	}
	c.JSON(http.StatusOK, run)
}

// agentServiceCaller adapts the AgentService to the harness.AgentCaller interface.
type agentServiceCaller struct {
	agentSvc    *service.AgentService
	workspaceID uint64
	userID      uint64
	db          *gorm.DB
}

func (c *agentServiceCaller) CallAgent(ctx context.Context, agentID uint64, modelName string, systemPrompt string, userMessage string, contextMap map[string]interface{}) (string, int, float64, error) {
	// Check if agent exists
	var agent model.Agent
	if err := c.db.Where("id = ? AND workspace_id = ?", agentID, c.workspaceID).First(&agent).Error; err != nil {
		return "", 0, 0, common.NotFound("Agent not found")
	}

	// Build dispatch context from contextMap
	var issueIDPtr, projectIDPtr *uint64
	if id, ok := contextMap["issue_id"].(uint64); ok {
		issueIDPtr = &id
	}
	if id, ok := contextMap["project_id"].(uint64); ok {
		projectIDPtr = &id
	}

	result, err := c.agentSvc.DispatchAgent(agentID, c.userID, userMessage, &service.DispatchContext{
		IssueID:     issueIDPtr,
		ProjectID:   projectIDPtr,
		WorkspaceID: c.workspaceID,
		TriggeredBy: "pipeline",
	})
	if err != nil {
		return "", 0, 0, err
	}

	// Estimate tokens and cost
	tokens := 500
	cost := float64(tokens) * 0.000002

	return result.ResultSummary, tokens, cost, nil
}
