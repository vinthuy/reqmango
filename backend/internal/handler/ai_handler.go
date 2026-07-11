package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/gorm"
)

// AIHandler handles AI endpoints.
type AIHandler struct {
	svc *service.AIService
	db  *gorm.DB
}

// NewAIHandler creates an AIHandler.
func NewAIHandler(svc *service.AIService, db *gorm.DB) *AIHandler {
	return &AIHandler{svc: svc, db: db}
}

func (h *AIHandler) getUserID(c *gin.Context) uint64 {
	user, exists := c.Get("currentUser")
	if !exists {
		return 0
	}
	if u, ok := user.(*model.User); ok {
		return u.ID
	}
	return 0
}

// buildContext builds the AI context from request parameters.
func (h *AIHandler) buildContext(c *gin.Context) *service.AIContext {
	actx := &service.AIContext{Mode: "ask"}

	actx.UserID = h.getUserID(c)

	if pid, err := strconv.ParseUint(c.Param("projectId"), 10, 64); err == nil {
		actx.ProjectID = pid
	}
	if wid, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64); err == nil {
		actx.WorkspaceID = wid
	}

	// Try to load workspace/project names from DB via the service
	// These would ideally come from a quick DB lookup; for now use IDs
	if actx.ProjectID > 0 {
		if projectSvc := h.getProjectInfo(actx.ProjectID); projectSvc != nil {
			if name, ok := projectSvc["name"].(string); ok {
				actx.ProjectName = name
			}
			if identifier, ok := projectSvc["identifier"].(string); ok {
				actx.ProjectIdentifier = identifier
			}
		}
	}

	pageIDStr := c.Query("page_id")
	if pageIDStr != "" {
		if pid, err := strconv.ParseUint(pageIDStr, 10, 64); err == nil && pid > 0 {
			var page model.Page
			if h.db.First(&page, pid).Error == nil {
				actx.PageTitle = page.Title
			}
		}
	}

	issueIDStr := c.Query("issue_id")
	if issueIDStr != "" {
		if iid, err := strconv.ParseUint(issueIDStr, 10, 64); err == nil && iid > 0 {
			actx.IssueID = iid
		}
	}

	return actx
}

func (h *AIHandler) getProjectInfo(projectID uint64) map[string]interface{} {
	var project model.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		return nil
	}
	return map[string]interface{}{
		"name":       project.Name,
		"identifier": project.Identifier,
	}
}

// ==================== Chat (SSE Streaming) ====================

// Chat handles POST /projects/:projectId/ai/chat (SSE streaming).
func (h *AIHandler) Chat(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	var req service.AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	actx := h.buildContext(c)
	actx.ProjectID = projectID
	if req.Mode != "" {
		actx.Mode = req.Mode
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	streamCh, err := h.svc.Chat(c.Request.Context(), &req, actx)
	if err != nil {
		data, _ := json.Marshal(service.StreamEvent{Type: "error", Error: err.Error()})
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		c.Writer.Flush()
		return
	}

	c.Stream(func(w io.Writer) bool {
		evt, ok := <-streamCh
		if !ok {
			return false
		}
		data, _ := json.Marshal(evt)
		fmt.Fprintf(w, "data: %s\n\n", data)
		return true
	})
}

// ==================== Search ====================

// Search handles POST /projects/:projectId/ai/search.
func (h *AIHandler) Search(c *gin.Context) {
	var req service.AISearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	actx := h.buildContext(c)
	result, err := h.svc.Search(c.Request.Context(), &req, actx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ==================== Create Preview ====================

// Analyze handles POST /projects/:projectId/ai/analyze.
func (h *AIHandler) Analyze(c *gin.Context) {
	actx := h.buildContext(c)
	result, err := h.svc.Analyze(c.Request.Context(), actx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// CreatePreview handles POST /projects/:projectId/ai/create.
func (h *AIHandler) CreatePreview(c *gin.Context) {
	var req service.AICreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	actx := h.buildContext(c)
	result, err := h.svc.CreatePreview(c.Request.Context(), &req, actx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ==================== Chart Generation ====================

// Chart handles POST /projects/:projectId/ai/chart.
func (h *AIHandler) Chart(c *gin.Context) {
	var req service.AIChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "query is required"})
		return
	}

	actx := h.buildContext(c)
	result, err := h.svc.GenerateChart(c.Request.Context(), &req, actx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ==================== Page AI ====================

// PageAI handles POST /pages/:pageId/ai.
func (h *AIHandler) PageAI(c *gin.Context) {
	var req service.PageAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "content is required"})
		return
	}
	result, err := h.svc.PageAI(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// AssistComment handles POST /issues/:issueId/ai/comment.
func (h *AIHandler) AssistComment(c *gin.Context) {
	var req service.AICommentRequest
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":err.Error()}); return }
	r, err := h.svc.AssistComment(c.Request.Context(), &req)
	if err != nil { c.JSON(500, gin.H{"message":err.Error()}); return }
	c.JSON(200, r)
}

// SprintPlan handles POST /projects/:projectId/ai/sprint-plan.
func (h *AIHandler) SprintPlan(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	r, err := h.svc.SprintPlan(c.Request.Context(), pid)
	if err != nil { c.JSON(500, gin.H{"message":err.Error()}); return }
	c.JSON(200, r)
}

// SuggestLabels handles POST /projects/:projectId/ai/suggest-labels.
func (h *AIHandler) SuggestLabels(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":err.Error()}); return }
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	r, err := h.svc.SuggestLabels(c.Request.Context(), pid, req.Name, req.Description)
	if err != nil { c.JSON(500, gin.H{"message":err.Error()}); return }
	c.JSON(200, r)
}

// TriageAnalyze handles POST /projects/:projectId/intake/:issueId/ai-analyze.
func (h *AIHandler) TriageAnalyze(c *gin.Context) {
	issueID, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	result, err := h.svc.TriageAnalyze(c.Request.Context(), issueID, projectID)
	if err != nil { c.JSON(500, gin.H{"message":err.Error()}); return }
	c.JSON(200, result)
}

// ==================== AI Config (workspace-level) ====================

// GetAIConfig handles GET /workspaces/:wsParam/ai-config.
func (h *AIHandler) GetAIConfig(c *gin.Context) {
	wsID, _ := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if wsID == 0 { c.JSON(400, gin.H{"message":"Invalid workspace"}); return }

	var cfg model.AIConfig
	if err := h.db.Where("workspace_id = ?", wsID).First(&cfg).Error; err != nil {
		c.JSON(200, gin.H{"provider":"deepseek","model":"deepseek-chat","max_tokens":4096,"is_active":true,"configured":false})
		return
	}
	c.JSON(200, gin.H{"id":cfg.ID,"provider":cfg.Provider,"model":cfg.Model,"max_tokens":cfg.MaxTokens,"is_active":cfg.IsActive,"configured":true})
}

// UpdateAIConfig handles PUT /workspaces/:wsParam/ai-config.
func (h *AIHandler) UpdateAIConfig(c *gin.Context) {
	wsID, _ := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if wsID == 0 { c.JSON(400, gin.H{"message":"Invalid workspace"}); return }

	var req struct {
		Provider  *string `json:"provider"`
		Model     *string `json:"model"`
		APIKey    *string `json:"api_key"`
		MaxTokens *int    `json:"max_tokens"`
		IsActive  *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":err.Error()}); return }

	var cfg model.AIConfig
	if err := h.db.Where("workspace_id = ?", wsID).First(&cfg).Error; err != nil {
		cfg = model.AIConfig{WorkspaceID: wsID}
	}
	if req.Provider != nil { cfg.Provider = *req.Provider }
	if req.Model != nil { cfg.Model = *req.Model }
	if req.APIKey != nil && *req.APIKey != "" { cfg.APIKey = *req.APIKey }
	if req.MaxTokens != nil { cfg.MaxTokens = *req.MaxTokens }
	if req.IsActive != nil { cfg.IsActive = *req.IsActive }

	if cfg.ID == 0 { h.db.Create(&cfg) } else { h.db.Save(&cfg) }

	c.JSON(200, gin.H{"id":cfg.ID,"provider":cfg.Provider,"model":cfg.Model,"max_tokens":cfg.MaxTokens,"is_active":cfg.IsActive,"configured":true})
}
