package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AgentSessionHandler handles HTTP requests for agent sessions.
type AgentSessionHandler struct {
	db *gorm.DB
}

// NewAgentSessionHandler creates a new AgentSessionHandler.
func NewAgentSessionHandler(db *gorm.DB) *AgentSessionHandler {
	return &AgentSessionHandler{db: db}
}

func (h *AgentSessionHandler) List(c *gin.Context) { c.JSON(200, gin.H{"data": []interface{}{}}) }
func (h *AgentSessionHandler) Get(c *gin.Context)  { c.JSON(200, gin.H{"message": "not implemented"}) }
