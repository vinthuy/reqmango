package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/reqmango/agent-service/internal/registry"
	"gorm.io/gorm"
)

// AgentPipelineHandler handles HTTP requests for agent pipelines.
type AgentPipelineHandler struct {
	db  *gorm.DB
	reg *registry.Registry
}

// NewAgentPipelineHandler creates a new AgentPipelineHandler.
func NewAgentPipelineHandler(db *gorm.DB, reg *registry.Registry) *AgentPipelineHandler {
	return &AgentPipelineHandler{db: db, reg: reg}
}

func (h *AgentPipelineHandler) List(c *gin.Context)   { c.JSON(200, gin.H{"data": []interface{}{}}) }
func (h *AgentPipelineHandler) Create(c *gin.Context) { c.JSON(201, gin.H{"message": "created"}) }
func (h *AgentPipelineHandler) Get(c *gin.Context)    { c.JSON(200, gin.H{"message": "not implemented"}) }
func (h *AgentPipelineHandler) Update(c *gin.Context) { c.JSON(200, gin.H{"message": "updated"}) }
func (h *AgentPipelineHandler) Delete(c *gin.Context) { c.JSON(200, gin.H{"message": "deleted"}) }
func (h *AgentPipelineHandler) Run(c *gin.Context)    { c.JSON(200, gin.H{"message": "run started"}) }
func (h *AgentPipelineHandler) GetRuns(c *gin.Context)  { c.JSON(200, gin.H{"data": []interface{}{}}) }
func (h *AgentPipelineHandler) GetRun(c *gin.Context)   { c.JSON(200, gin.H{"message": "not implemented"}) }
