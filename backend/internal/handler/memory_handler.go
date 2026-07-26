package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
)

// MemoryHandler handles memory-related endpoints
type MemoryHandler struct {
	svc *service.MemoryService
}

// NewMemoryHandler creates a new MemoryHandler
func NewMemoryHandler(db *gorm.DB) *MemoryHandler {
	return &MemoryHandler{
		svc: service.NewMemoryService(db),
	}
}

// parseWorkspaceID parses workspace ID from route parameter
func (h *MemoryHandler) parseWorkspaceID(c *gin.Context) uint64 {
	id, _ := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	return id
}

// respond handles error responses
func (h *MemoryHandler) respond(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	return true
}

// ListMemories handles GET /workspaces/:wsParam/memories
func (h *MemoryHandler) ListMemories(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	var filters service.MemoryListFilters

	if pidStr := c.Query("project_id"); pidStr != "" {
		if pid, err := strconv.ParseUint(pidStr, 10, 64); err == nil {
			filters.ProjectID = &pid
		}
	}
	if iidStr := c.Query("issue_id"); iidStr != "" {
		if iid, err := strconv.ParseUint(iidStr, 10, 64); err == nil {
			filters.IssueID = &iid
		}
	}
	if aidStr := c.Query("agent_id"); aidStr != "" {
		if aid, err := strconv.ParseUint(aidStr, 10, 64); err == nil {
			filters.AgentID = &aid
		}
	}
	if mt := c.Query("memory_type"); mt != "" {
		filters.MemoryType = model.MemoryType(mt)
	}
	if sc := c.Query("scope"); sc != "" {
		filters.Scope = model.MemoryScope(sc)
	}
	if ck := c.Query("context_key"); ck != "" {
		filters.ContextKey = ck
	}
	if tag := c.Query("tag"); tag != "" {
		filters.Tag = tag
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filters.Offset = offset
		}
	}
	if q := c.Query("q"); q != "" {
		filters.SearchQuery = q
	}
	if minRelevanceStr := c.Query("min_relevance"); minRelevanceStr != "" {
		if minRelevance, err := strconv.ParseFloat(minRelevanceStr, 64); err == nil {
			filters.MinRelevance = minRelevance
		}
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters.EndDate = &endDate
		}
	}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}
	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filters.SortOrder = sortOrder
	}

	entries, err := h.svc.ListMemories(c.Request.Context(), wid, filters)
	if h.respond(c, err) {
		return
	}
	c.JSON(200, entries)
}

// GetMemory handles GET /workspaces/:wsParam/memories/:memoryId
func (h *MemoryHandler) GetMemory(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	id, err := strconv.ParseUint(c.Param("memoryId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid memory ID"})
		return
	}

	entry, err := h.svc.GetMemoryByID(c.Request.Context(), id, wid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"message": "Memory not found"})
			return
		}
		h.respond(c, err)
		return
	}
	c.JSON(200, entry)
}

// CreateMemory handles POST /workspaces/:wsParam/memories
func (h *MemoryHandler) CreateMemory(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	var req struct {
		ProjectID   *uint64              `json:"project_id"`
		IssueID     *uint64              `json:"issue_id"`
		AgentID     *uint64              `json:"agent_id"`
		MemoryType  *string              `json:"memory_type"`
		Scope       *string              `json:"scope"`
		Content     string               `json:"content"`
		Embedding   []float64            `json:"embedding"`
		Metadata    map[string]interface{} `json:"metadata"`
		Tags        []string             `json:"tags"`
		ContextKey  string               `json:"context_key"`
		ContextName string               `json:"context_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	if req.Content == "" {
		c.JSON(400, gin.H{"message": "content is required"})
		return
	}

	entry := &model.MemoryEntry{
		WorkspaceID: wid,
		ProjectID:   req.ProjectID,
		IssueID:     req.IssueID,
		AgentID:     req.AgentID,
		Content:     req.Content,
		ContextKey:  req.ContextKey,
		ContextName: req.ContextName,
	}

	if req.MemoryType != nil {
		entry.MemoryType = model.MemoryType(*req.MemoryType)
	}
	if req.Scope != nil {
		entry.Scope = model.MemoryScope(*req.Scope)
	}
	if len(req.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Tags)
		entry.Tags = tagsJSON
	}
	if len(req.Metadata) > 0 {
		metaJSON, _ := json.Marshal(req.Metadata)
		entry.Metadata = metaJSON
	}
	if len(req.Embedding) > 0 {
		embeddingJSON, _ := json.Marshal(req.Embedding)
		entry.Embedding = embeddingJSON
	}

	result, err := h.svc.CreateMemory(c.Request.Context(), entry)
	if h.respond(c, err) {
		return
	}
	c.JSON(201, result)
}

// UpdateMemory handles PUT /workspaces/:wsParam/memories/:memoryId
func (h *MemoryHandler) UpdateMemory(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	id, err := strconv.ParseUint(c.Param("memoryId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid memory ID"})
		return
	}

	var req struct {
		Content     *string              `json:"content"`
		MemoryType  *string              `json:"memory_type"`
		Scope       *string              `json:"scope"`
		Tags        *[]string            `json:"tags"`
		Metadata    *map[string]interface{} `json:"metadata"`
		Embedding   *[]float64           `json:"embedding"`
		ContextKey  *string              `json:"context_key"`
		ContextName *string              `json:"context_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.MemoryType != nil {
		updates["memory_type"] = *req.MemoryType
	}
	if req.Scope != nil {
		updates["scope"] = *req.Scope
	}
	if req.ContextKey != nil {
		updates["context_key"] = *req.ContextKey
	}
	if req.ContextName != nil {
		updates["context_name"] = *req.ContextName
	}
	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(*req.Tags)
		updates["tags"] = tagsJSON
	}
	if req.Metadata != nil {
		metaJSON, _ := json.Marshal(*req.Metadata)
		updates["metadata"] = metaJSON
	}
	if req.Embedding != nil {
		embeddingJSON, _ := json.Marshal(*req.Embedding)
		updates["embedding"] = embeddingJSON
	}

	result, err := h.svc.UpdateMemory(c.Request.Context(), id, wid, updates)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"message": "Memory not found"})
			return
		}
		h.respond(c, err)
		return
	}
	c.JSON(200, result)
}

// DeleteMemory handles DELETE /workspaces/:wsParam/memories/:memoryId
func (h *MemoryHandler) DeleteMemory(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	id, err := strconv.ParseUint(c.Param("memoryId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid memory ID"})
		return
	}

	if err := h.svc.DeleteMemory(c.Request.Context(), id, wid); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"message": "Memory not found"})
			return
		}
		h.respond(c, err)
		return
	}
	c.JSON(204, nil)
}

// SearchMemories handles POST /workspaces/:wsParam/memories/search
func (h *MemoryHandler) SearchMemories(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	var req struct {
		Query string `json:"query" binding:"required"`
		Limit int    `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	entries, err := h.svc.SearchMemories(c.Request.Context(), wid, req.Query, req.Limit)
	if h.respond(c, err) {
		return
	}
	c.JSON(200, entries)
}

// SemanticSearch handles POST /workspaces/:wsParam/memories/semantic-search
func (h *MemoryHandler) SemanticSearch(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	var req struct {
		Embedding []float64 `json:"embedding" binding:"required"`
		Limit     int       `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	entries, err := h.svc.SemanticSearch(c.Request.Context(), wid, req.Embedding, req.Limit)
	if h.respond(c, err) {
		return
	}
	c.JSON(200, entries)
}

// CreateMemorySession handles POST /workspaces/:wsParam/memory-sessions
func (h *MemoryHandler) CreateMemorySession(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	var req struct {
		SessionType string `json:"session_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	session, err := h.svc.CreateMemorySession(c.Request.Context(), wid, req.SessionType)
	if h.respond(c, err) {
		return
	}
	c.JSON(201, session)
}

// GetMemorySession handles GET /workspaces/:wsParam/memory-sessions/:sessionId
func (h *MemoryHandler) GetMemorySession(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(400, gin.H{"message": "Invalid session ID"})
		return
	}

	session, err := h.svc.GetMemorySession(c.Request.Context(), sessionID, wid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"message": "Session not found"})
			return
		}
		h.respond(c, err)
		return
	}
	c.JSON(200, session)
}

// CloseMemorySession handles PUT /workspaces/:wsParam/memory-sessions/:sessionId/close
func (h *MemoryHandler) CloseMemorySession(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(400, gin.H{"message": "Invalid session ID"})
		return
	}

	if err := h.svc.CloseMemorySession(c.Request.Context(), sessionID, wid); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"message": "Session not found"})
			return
		}
		h.respond(c, err)
		return
	}
	c.JSON(200, gin.H{"status": "closed"})
}

// GetContextMemories handles GET /workspaces/:wsParam/memories/context/:contextKey
func (h *MemoryHandler) GetContextMemories(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	contextKey := c.Param("contextKey")
	if contextKey == "" {
		c.JSON(400, gin.H{"message": "context_key is required"})
		return
	}

	entries, err := h.svc.GetContextMemories(c.Request.Context(), wid, contextKey)
	if h.respond(c, err) {
		return
	}
	c.JSON(200, entries)
}

// PruneMemories handles POST /workspaces/:wsParam/memories/prune
func (h *MemoryHandler) PruneMemories(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	var req struct {
		MaxDays   *int     `json:"max_days"`
		MinScore  *float64 `json:"min_score"`
		Expired   *bool    `json:"expired"`
		LowRelevance *bool `json:"low_relevance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	result := gin.H{}

	if req.Expired != nil && *req.Expired {
		expiredCount, err := h.svc.PruneExpiredMemories(c.Request.Context())
		if h.respond(c, err) {
			return
		}
		result["expired_pruned"] = expiredCount
	}

	if req.LowRelevance != nil && *req.LowRelevance {
		maxDays := 30
		if req.MaxDays != nil && *req.MaxDays > 0 {
			maxDays = *req.MaxDays
		}
		minScore := 0.3
		if req.MinScore != nil && *req.MinScore >= 0 {
			minScore = *req.MinScore
		}
		lowRelevanceCount, err := h.svc.PruneLowRelevanceMemories(c.Request.Context(), maxDays, minScore)
		if h.respond(c, err) {
			return
		}
		result["low_relevance_pruned"] = lowRelevanceCount
	}

	c.JSON(200, result)
}

// GetMemoryStats handles GET /workspaces/:wsParam/memories/stats
func (h *MemoryHandler) GetMemoryStats(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	stats, err := h.svc.GetMemoryStats(c.Request.Context(), wid)
	if h.respond(c, err) {
		return
	}
	c.JSON(200, stats)
}

// FindSimilarMemories handles POST /workspaces/:wsParam/memories/find-similar
func (h *MemoryHandler) FindSimilarMemories(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	var req struct {
		Content   string  `json:"content" binding:"required"`
		Threshold float64 `json:"threshold"`
		Limit     int     `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	if req.Threshold <= 0 {
		req.Threshold = 0.5
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	results, err := h.svc.FindSimilarMemories(c.Request.Context(), wid, req.Content, req.Threshold, req.Limit)
	if h.respond(c, err) {
		return
	}
	c.JSON(200, results)
}

// MergeMemories handles POST /workspaces/:wsParam/memories/merge
func (h *MemoryHandler) MergeMemories(c *gin.Context) {
	wid := h.parseWorkspaceID(c)
	if wid == 0 {
		c.JSON(400, gin.H{"message": "Invalid workspace"})
		return
	}

	var req struct {
		MemoryIDs []uint64 `json:"memory_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	merged, err := h.svc.MergeMemories(c.Request.Context(), req.MemoryIDs, wid)
	if err != nil {
		if errors.Is(err, service.ErrMergeRequiresTwoMemories) {
			c.JSON(400, gin.H{"message": err.Error()})
			return
		}
		if err.Error() == "some memory entries not found" {
			c.JSON(404, gin.H{"message": err.Error()})
			return
		}
		h.respond(c, err)
		return
	}
	c.JSON(200, merged)
}