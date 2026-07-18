package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/reqmango/agent-service/internal/service"
)

// AgentLoopHandler handles HTTP requests for agent loops.
type AgentLoopHandler struct {
	svc *service.LoopService
}

// NewAgentLoopHandler creates a new AgentLoopHandler.
func NewAgentLoopHandler(svc *service.LoopService) *AgentLoopHandler {
	return &AgentLoopHandler{svc: svc}
}

func (h *AgentLoopHandler) List(c *gin.Context)   { c.JSON(200, gin.H{"data": []interface{}{}}) }
func (h *AgentLoopHandler) Create(c *gin.Context) { c.JSON(201, gin.H{"message": "created"}) }
func (h *AgentLoopHandler) Get(c *gin.Context)    { c.JSON(200, gin.H{"message": "not implemented"}) }
func (h *AgentLoopHandler) Update(c *gin.Context) { c.JSON(200, gin.H{"message": "updated"}) }
func (h *AgentLoopHandler) Delete(c *gin.Context) { c.JSON(200, gin.H{"message": "deleted"}) }
func (h *AgentLoopHandler) Start(c *gin.Context)  { c.JSON(200, gin.H{"message": "started"}) }
func (h *AgentLoopHandler) GetRuns(c *gin.Context)  { c.JSON(200, gin.H{"data": []interface{}{}}) }
func (h *AgentLoopHandler) Stop(c *gin.Context)     { c.JSON(200, gin.H{"message": "stopped"}) }
func (h *AgentLoopHandler) GetRun(c *gin.Context)   { c.JSON(200, gin.H{"message": "not implemented"}) }
