package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/reqmango/backend/internal/model"
)

// MemoryService provides memory management capabilities
type MemoryService struct {
	db *gorm.DB
}

// NewMemoryService creates a new MemoryService
func NewMemoryService(db *gorm.DB) *MemoryService {
	return &MemoryService{db: db}
}

// CreateMemory creates a new memory entry
func (s *MemoryService) CreateMemory(ctx context.Context, entry *model.MemoryEntry) (*model.MemoryEntry, error) {
	if entry.Content == "" {
		return nil, errors.New("content is required")
	}

	// Set defaults
	if entry.ContextKey == "" {
		entry.ContextKey = uuid.New().String()[:8]
	}
	if entry.MemoryType == "" {
		entry.MemoryType = model.MemoryMediumTerm
	}
	if entry.Scope == "" {
		entry.Scope = model.ScopeWorkspace
	}

	// Set expiry for short-term memory
	if entry.MemoryType == model.MemoryShortTerm && entry.ExpiresAt == nil {
		expiry := time.Now().Add(24 * time.Hour)
		entry.ExpiresAt = &expiry
	}

	if err := s.db.WithContext(ctx).Create(entry).Error; err != nil {
		return nil, err
	}

	return entry, nil
}

// GetMemoryByID retrieves a memory entry by ID
func (s *MemoryService) GetMemoryByID(ctx context.Context, id uint64, workspaceID uint64) (*model.MemoryEntry, error) {
	var entry model.MemoryEntry
	if err := s.db.WithContext(ctx).
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// ListMemories retrieves memory entries with filters
func (s *MemoryService) ListMemories(ctx context.Context, workspaceID uint64, filters MemoryListFilters) ([]*model.MemoryEntry, error) {
	query := s.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)

	if filters.ProjectID != nil {
		query = query.Where("project_id = ?", *filters.ProjectID)
	}
	if filters.IssueID != nil {
		query = query.Where("issue_id = ?", *filters.IssueID)
	}
	if filters.AgentID != nil {
		query = query.Where("agent_id = ?", *filters.AgentID)
	}
	if filters.MemoryType != "" {
		query = query.Where("memory_type = ?", filters.MemoryType)
	}
	if filters.Scope != "" {
		query = query.Where("scope = ?", filters.Scope)
	}
	if filters.ContextKey != "" {
		query = query.Where("context_key = ?", filters.ContextKey)
	}
	if filters.Tag != "" {
		query = query.Where("tags @> ?", fmt.Sprintf("[\"%s\"]", filters.Tag))
	}

	// Exclude expired short-term memories
	query = query.Where("expires_at IS NULL OR expires_at > NOW()")

	query = query.Order("relevance_score DESC, created_at DESC")

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var entries []*model.MemoryEntry
	if err := query.Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

// MemoryListFilters defines filters for listing memories
type MemoryListFilters struct {
	ProjectID   *uint64
	IssueID     *uint64
	AgentID     *uint64
	MemoryType  model.MemoryType
	Scope       model.MemoryScope
	ContextKey  string
	Tag         string
	Limit       int
	Offset      int
}

// UpdateMemory updates an existing memory entry
func (s *MemoryService) UpdateMemory(ctx context.Context, id, workspaceID uint64, updates map[string]interface{}) (*model.MemoryEntry, error) {
	var entry model.MemoryEntry
	if err := s.db.WithContext(ctx).
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		First(&entry).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&entry).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &entry, nil
}

// DeleteMemory deletes a memory entry
func (s *MemoryService) DeleteMemory(ctx context.Context, id, workspaceID uint64) error {
	return s.db.WithContext(ctx).
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		Delete(&model.MemoryEntry{}).Error
}

// SearchMemories searches memories by content
func (s *MemoryService) SearchMemories(ctx context.Context, workspaceID uint64, query string, limit int) ([]*model.MemoryEntry, error) {
	if query == "" {
		return nil, errors.New("query is required")
	}

	// Simple keyword search using ILIKE
	keywords := strings.Fields(strings.ToLower(query))
	var conditions []string
	var args []interface{}

	for _, kw := range keywords {
		conditions = append(conditions, "LOWER(content) LIKE ?")
		args = append(args, "%"+kw+"%")
	}

	args = append([]interface{}{workspaceID}, args...)

	queryStr := fmt.Sprintf("workspace_id = ? AND (%s)", strings.Join(conditions, " OR "))

	var entries []*model.MemoryEntry
	if err := s.db.WithContext(ctx).
		Where(queryStr, args...).
		Where("expires_at IS NULL OR expires_at > NOW()").
		Order("relevance_score DESC").
		Limit(limit).
		Find(&entries).Error; err != nil {
		return nil, err
	}

	return entries, nil
}

// SemanticSearch performs semantic similarity search using embedding vectors
func (s *MemoryService) SemanticSearch(ctx context.Context, workspaceID uint64, queryEmbedding []float64, limit int) ([]*model.MemoryEntry, error) {
	if len(queryEmbedding) == 0 {
		return nil, errors.New("query embedding is required")
	}

	// Simple cosine similarity search using PostgreSQL's vector operations
	// This requires pgvector extension or we use a simplified approach

	var entries []*model.MemoryEntry
	if err := s.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Where("embedding IS NOT NULL").
		Where("expires_at IS NULL OR expires_at > NOW()").
		Order("relevance_score DESC").
		Limit(limit).
		Find(&entries).Error; err != nil {
		return nil, err
	}

	// Calculate cosine similarity with query embedding
	for _, entry := range entries {
		if entry.Embedding != nil {
			var embedding []float64
			if err := json.Unmarshal(entry.Embedding, &embedding); err == nil {
				sim := cosineSimilarity(queryEmbedding, embedding)
				entry.RelevanceScore = sim
			}
		}
	}

	return entries, nil
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}

// CreateMemorySession creates a new memory session
func (s *MemoryService) CreateMemorySession(ctx context.Context, workspaceID uint64, sessionType string) (*model.MemorySession, error) {
	if sessionType == "" {
		sessionType = "chat"
	}

	session := &model.MemorySession{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		SessionType: sessionType,
	}

	if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, err
	}

	return session, nil
}

// GetMemorySession retrieves a memory session by ID
func (s *MemoryService) GetMemorySession(ctx context.Context, id string, workspaceID uint64) (*model.MemorySession, error) {
	var session model.MemorySession
	if err := s.db.WithContext(ctx).
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// CloseMemorySession closes a memory session
func (s *MemoryService) CloseMemorySession(ctx context.Context, id string, workspaceID uint64) error {
	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&model.MemorySession{}).
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		Update("closed_at", now).Error
}

// AddMemoryToSession adds a memory entry to a session
func (s *MemoryService) AddMemoryToSession(ctx context.Context, sessionID string, memoryID uint64) error {
	return s.db.WithContext(ctx).
		Model(&model.MemorySession{}).
		Where("id = ?", sessionID).
		Update("memory_count", gorm.Expr("memory_count + 1")).Error
}

// PruneExpiredMemories removes expired short-term memories
func (s *MemoryService) PruneExpiredMemories(ctx context.Context) error {
	return s.db.WithContext(ctx).
		Where("memory_type = ? AND expires_at IS NOT NULL AND expires_at < NOW()", model.MemoryShortTerm).
		Delete(&model.MemoryEntry{}).Error
}

// GetContextMemories retrieves memories for a specific context
func (s *MemoryService) GetContextMemories(ctx context.Context, workspaceID uint64, contextKey string) ([]*model.MemoryEntry, error) {
	return s.ListMemories(ctx, workspaceID, MemoryListFilters{
		ContextKey: contextKey,
		Limit:      100,
	})
}

// UpdateRelevance updates the relevance score for a memory entry
func (s *MemoryService) UpdateRelevance(ctx context.Context, id, workspaceID uint64, score float64) error {
	return s.db.WithContext(ctx).
		Model(&model.MemoryEntry{}).
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		Update("relevance_score", score).Error
}