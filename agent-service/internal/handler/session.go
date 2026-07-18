package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	agentmodel "github.com/reqmango/agent-service/internal/model"
	"gorm.io/gorm"
)

type AgentSessionHandler struct {
	db *gorm.DB
}

func NewAgentSessionHandler(db *gorm.DB) *AgentSessionHandler {
	return &AgentSessionHandler{db: db}
}

func (h *AgentSessionHandler) getWS(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

func (h *AgentSessionHandler) List(c *gin.Context) {
	wsID, err := h.getWS(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	agentType := c.Query("agent_type")
	status := c.Query("status")

	var sessions []agentmodel.AgentSession
	query := h.db.Where("workspace_id = ?", wsID)
	if agentType != "" {
		query = query.Where("agent_type = ?", agentType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("started_at DESC").Limit(limit).Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

func (h *AgentSessionHandler) Get(c *gin.Context) {
	wsID, err := h.getWS(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
		return
	}

	sessionID := c.Param("sessionId")
	var session agentmodel.AgentSession
	if err := h.db.Where("id = ? AND workspace_id = ?", sessionID, wsID).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session)
}
