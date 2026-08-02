package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/reqmango/backend/internal/ai/llm"
	"github.com/reqmango/backend/internal/model"
)

var ErrMergeRequiresTwoMemories = errors.New("at least 2 memories are required to merge")

// MemoryService provides memory management capabilities
type MemoryService struct {
	db  *gorm.DB
	llm *llm.LLMClient
}

// NewMemoryService creates a new MemoryService. The llm client is optional and
// enables automatic embedding generation when configured.
func NewMemoryService(db *gorm.DB, llmClient *llm.LLMClient) *MemoryService {
	return &MemoryService{db: db, llm: llmClient}
}

// SetLLMClient injects an LLM client for embedding generation.
func (s *MemoryService) SetLLMClient(c *llm.LLMClient) {
	s.llm = c
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

	// Auto-generate embedding if not provided and an LLM client is configured.
	// Failures are non-fatal: we still persist the memory without an embedding
	// so that memory creation never breaks when the AI backend is unavailable.
	if len(entry.Embedding) == 0 && s.llm != nil && s.llm.SupportsEmbedding() {
		if emb, err := s.llm.GenerateEmbedding(ctx, entry.Content); err == nil && len(emb) > 0 {
			if embJSON, mErr := json.Marshal(emb); mErr == nil {
				entry.Embedding = embJSON
			}
		}
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
func (s *MemoryService) ListMemories(ctx context.Context, workspaceID uint64, filters interface{}) ([]*model.MemoryEntry, error) {
	query := s.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)

	switch f := filters.(type) {
	case MemoryListFilters:
		if f.ProjectID != nil {
			query = query.Where("project_id = ?", *f.ProjectID)
		}
		if f.IssueID != nil {
			query = query.Where("issue_id = ?", *f.IssueID)
		}
		if f.AgentID != nil {
			query = query.Where("agent_id = ?", *f.AgentID)
		}
		if f.MemoryType != "" {
			query = query.Where("memory_type = ?", f.MemoryType)
		}
		if f.Scope != "" {
			query = query.Where("scope = ?", f.Scope)
		}
		if f.ContextKey != "" {
			query = query.Where("context_key = ?", f.ContextKey)
		}
		if f.Tag != "" {
			query = query.Where("tags @> ?", fmt.Sprintf("[\"%s\"]", f.Tag))
		}
		if f.SearchQuery != "" {
			keywords := strings.Fields(strings.ToLower(f.SearchQuery))
			var conditions []string
			var args []interface{}
			for _, kw := range keywords {
				conditions = append(conditions, "LOWER(content) LIKE ?")
				args = append(args, "%"+kw+"%")
			}
			query = query.Where(strings.Join(conditions, " OR "), args...)
		}
		if f.MinRelevance > 0 {
			query = query.Where("relevance_score >= ?", f.MinRelevance)
		}
		if f.StartDate != nil {
			query = query.Where("created_at >= ?", *f.StartDate)
		}
		if f.EndDate != nil {
			query = query.Where("created_at <= ?", *f.EndDate)
		}
		if f.Limit > 0 {
			query = query.Limit(f.Limit)
		}
		if f.Offset > 0 {
			query = query.Offset(f.Offset)
		}
	case map[string]interface{}:
		if v, ok := f["project_id"]; ok {
			query = query.Where("project_id = ?", v)
		}
		if v, ok := f["issue_id"]; ok {
			query = query.Where("issue_id = ?", v)
		}
		if v, ok := f["agent_id"]; ok {
			query = query.Where("agent_id = ?", v)
		}
		if v, ok := f["memory_type"]; ok {
			query = query.Where("memory_type = ?", v)
		}
		if v, ok := f["scope"]; ok {
			query = query.Where("scope = ?", v)
		}
		if v, ok := f["limit"]; ok {
			query = query.Limit(v.(int))
		}
	}

	// Exclude expired short-term memories
	query = query.Where("expires_at IS NULL OR expires_at > NOW()")

	switch f := filters.(type) {
	case MemoryListFilters:
		sortField := "relevance_score"
		if f.SortBy == "created_at" || f.SortBy == "updated_at" {
			sortField = f.SortBy
		}
		sortOrder := "DESC"
		if f.SortOrder == "asc" {
			sortOrder = "ASC"
		}
		query = query.Order(fmt.Sprintf("%s %s", sortField, sortOrder))
	default:
		query = query.Order("relevance_score DESC, created_at DESC")
	}

	var entries []*model.MemoryEntry
	if err := query.Find(&entries).Error; err != nil {
		return nil, err
	}

	if f, ok := filters.(MemoryListFilters); ok && f.SearchQuery != "" {
		for _, entry := range entries {
			entry.Summary = extractSummary(entry.Content, f.SearchQuery)
		}
	}

	return entries, nil
}

// MemoryListFilters defines filters for listing memories
type MemoryListFilters struct {
	ProjectID    *uint64
	IssueID      *uint64
	AgentID      *uint64
	MemoryType   model.MemoryType
	Scope        model.MemoryScope
	ContextKey   string
	Tag          string
	Limit        int
	Offset       int
	SearchQuery  string
	MinRelevance float64
	StartDate    *time.Time
	EndDate      *time.Time
	SortBy       string
	SortOrder    string
}

// UpdateMemory updates an existing memory entry
func (s *MemoryService) UpdateMemory(ctx context.Context, id, workspaceID uint64, updates map[string]interface{}) (*model.MemoryEntry, error) {
	var entry model.MemoryEntry
	if err := s.db.WithContext(ctx).
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		First(&entry).Error; err != nil {
		return nil, err
	}

	// If content is being updated and embedding support is configured,
	// regenerate the embedding so semantic search stays accurate.
	if newContent, ok := updates["content"].(string); ok && newContent != "" && s.llm != nil && s.llm.SupportsEmbedding() {
		if emb, err := s.llm.GenerateEmbedding(ctx, newContent); err == nil && len(emb) > 0 {
			if embJSON, mErr := json.Marshal(emb); mErr == nil {
				updates["embedding"] = embJSON
			}
		}
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

// SemanticSearch performs semantic similarity search using embedding vectors.
// Entries are fetched from the DB, their cosine similarity to the query
// embedding is computed in Go, and the result is re-sorted by that score
// (descending) so callers receive truly most-similar entries first.
func (s *MemoryService) SemanticSearch(ctx context.Context, workspaceID uint64, queryEmbedding []float64, limit int) ([]*model.MemoryEntry, error) {
	if len(queryEmbedding) == 0 {
		return nil, errors.New("query embedding is required")
	}

	// Fetch candidate entries that have embeddings. We over-fetch relative to
	// the requested limit because the DB-side relevance_score is unrelated to
	// the query embedding similarity and would otherwise prune good matches.
	fetchLimit := limit * 5
	if fetchLimit < 50 {
		fetchLimit = 50
	}

	var entries []*model.MemoryEntry
	if err := s.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Where("embedding IS NOT NULL").
		Where("expires_at IS NULL OR expires_at > NOW()").
		Order("created_at DESC").
		Limit(fetchLimit).
		Find(&entries).Error; err != nil {
		return nil, err
	}

	// Compute cosine similarity against the query embedding for every entry.
	scored := make([]*model.MemoryEntry, 0, len(entries))
	for _, entry := range entries {
		if len(entry.Embedding) == 0 {
			continue
		}
		var embedding []float64
		if err := json.Unmarshal(entry.Embedding, &embedding); err != nil {
			continue
		}
		sim := cosineSimilarity(queryEmbedding, embedding)
		entry.RelevanceScore = sim
		scored = append(scored, entry)
	}

	// Re-sort by computed similarity (descending) — the DB-side order by
	// relevance_score is meaningless for an arbitrary query embedding.
	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].RelevanceScore > scored[j].RelevanceScore
	})

	if limit > 0 && len(scored) > limit {
		scored = scored[:limit]
	}

	return scored, nil
}

// SemanticSearchByText generates an embedding for the query text and then
// performs a semantic search. Returns an error if no LLM client is configured
// or the provider does not support embeddings.
func (s *MemoryService) SemanticSearchByText(ctx context.Context, workspaceID uint64, query string, limit int) ([]*model.MemoryEntry, error) {
	if query == "" {
		return nil, errors.New("query is required")
	}
	if s.llm == nil || !s.llm.SupportsEmbedding() {
		return nil, errors.New("semantic search requires an LLM client with embedding support")
	}
	if limit <= 0 {
		limit = 10
	}
	emb, err := s.llm.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("generate query embedding: %w", err)
	}
	return s.SemanticSearch(ctx, workspaceID, emb, limit)
}

// extractSummary extracts a summary from content that contains search keywords
// Returns a summary of ~100-150 characters, prioritizing context around keywords
func extractSummary(content, query string) string {
	contentLen := len(content)
	if contentLen <= 150 {
		return content
	}

	summaryLen := 150
	keywords := strings.Fields(strings.ToLower(query))
	lowerContent := strings.ToLower(content)

	for _, kw := range keywords {
		idx := strings.Index(lowerContent, kw)
		if idx != -1 {
			start := idx - 40
			if start < 0 {
				start = 0
			}
			end := idx + len(kw) + 110
			if end > contentLen {
				end = contentLen
			}

			result := content[start:end]
			if start > 0 {
				result = "..." + result
			}
			if end < contentLen {
				result = result + "..."
			}
			return result
		}
	}

	return content[:summaryLen] + "..."
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
	sim := dot / (math.Sqrt(magA) * math.Sqrt(magB))
	if math.IsNaN(sim) || math.IsInf(sim, 0) {
		return 0
	}
	return sim
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
func (s *MemoryService) PruneExpiredMemories(ctx context.Context) (int64, error) {
	result := s.db.WithContext(ctx).
		Where("memory_type = ? AND expires_at IS NOT NULL AND expires_at < NOW()", model.MemoryShortTerm).
		Delete(&model.MemoryEntry{})
	return result.RowsAffected, result.Error
}

// PruneLowRelevanceMemories removes old memories with low relevance score
func (s *MemoryService) PruneLowRelevanceMemories(ctx context.Context, maxDays int, minScore float64) (int64, error) {
	// Build the interval with MAKE_INTERVAL so the integer parameter is
	// substituted by the driver (a literal '? days' string would not be
	// interpolated inside single quotes).
	result := s.db.WithContext(ctx).
		Where("created_at < NOW() - MAKE_INTERVAL(days => ?) AND relevance_score < ?", maxDays, minScore).
		Delete(&model.MemoryEntry{})
	return result.RowsAffected, result.Error
}

// GetMemoryStats returns memory statistics by type and scope
func (s *MemoryService) GetMemoryStats(ctx context.Context, workspaceID uint64) (map[string]int64, error) {
	result := map[string]int64{
		"total":           0,
		"short_term":      0,
		"medium_term":     0,
		"long_term":       0,
		"scope_workspace": 0,
		"scope_project":   0,
		"scope_issue":     0,
		"scope_agent":     0,
		"expired":         0,
	}

	var total int64
	if err := s.db.WithContext(ctx).Model(&model.MemoryEntry{}).
		Where("workspace_id = ?", workspaceID).Count(&total).Error; err != nil {
		return nil, err
	}
	result["total"] = total

	var typeCounts []struct {
		MemoryType string `gorm:"column:memory_type"`
		Count      int64  `gorm:"column:count"`
	}
	if err := s.db.WithContext(ctx).Model(&model.MemoryEntry{}).
		Select("memory_type, COUNT(*) as count").
		Where("workspace_id = ?", workspaceID).
		Group("memory_type").
		Find(&typeCounts).Error; err != nil {
		return nil, err
	}
	for _, tc := range typeCounts {
		switch tc.MemoryType {
		case string(model.MemoryShortTerm):
			result["short_term"] = tc.Count
		case string(model.MemoryMediumTerm):
			result["medium_term"] = tc.Count
		case string(model.MemoryLongTerm):
			result["long_term"] = tc.Count
		}
	}

	var scopeCounts []struct {
		Scope string `gorm:"column:scope"`
		Count int64  `gorm:"column:count"`
	}
	if err := s.db.WithContext(ctx).Model(&model.MemoryEntry{}).
		Select("scope, COUNT(*) as count").
		Where("workspace_id = ?", workspaceID).
		Group("scope").
		Find(&scopeCounts).Error; err != nil {
		return nil, err
	}
	for _, sc := range scopeCounts {
		switch sc.Scope {
		case string(model.ScopeWorkspace):
			result["scope_workspace"] = sc.Count
		case string(model.ScopeProject):
			result["scope_project"] = sc.Count
		case string(model.ScopeIssue):
			result["scope_issue"] = sc.Count
		case string(model.ScopeAgent):
			result["scope_agent"] = sc.Count
		}
	}

	var expired int64
	if err := s.db.WithContext(ctx).Model(&model.MemoryEntry{}).
		Where("workspace_id = ? AND expires_at IS NOT NULL AND expires_at < NOW()", workspaceID).Count(&expired).Error; err != nil {
		return nil, err
	}
	result["expired"] = expired

	return result, nil
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

// tokenize splits text into words (tokens), supports both Chinese and English
func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	
	// Split by whitespace first
	words := strings.Fields(text)
	
	for _, w := range words {
		var cleaned string
		var hasChinese bool
		
		for _, c := range w {
			// Check if it's a Chinese character
			if c >= '\u4e00' && c <= '\u9fff' {
				// Chinese: add each character as separate token
				tokens = append(tokens, string(c))
				hasChinese = true
			} else if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
				// English/numbers: accumulate into word
				cleaned += string(c)
			}
		}
		
		if !hasChinese && cleaned != "" {
			tokens = append(tokens, cleaned)
		}
	}
	
	return tokens
}

// calculateWordFrequency creates a frequency map for words in text
func calculateWordFrequency(text string) map[string]int {
	tokens := tokenize(text)
	freq := make(map[string]int)
	for _, t := range tokens {
		freq[t]++
	}
	return freq
}

// CalculateTextSimilarity calculates cosine similarity between two texts based on word frequency
func (s *MemoryService) CalculateTextSimilarity(text1, text2 string) float64 {
	if text1 == "" || text2 == "" {
		return 0.0
	}

	freq1 := calculateWordFrequency(text1)
	freq2 := calculateWordFrequency(text2)

	var dot, mag1, mag2 float64
	words := make(map[string]bool)
	for w := range freq1 {
		words[w] = true
	}
	for w := range freq2 {
		words[w] = true
	}

	for w := range words {
		f1 := float64(freq1[w])
		f2 := float64(freq2[w])
		dot += f1 * f2
		mag1 += f1 * f1
		mag2 += f2 * f2
	}

	if mag1 == 0 || mag2 == 0 {
		return 0.0
	}

	return dot / (math.Sqrt(mag1) * math.Sqrt(mag2))
}

// SimilarMemoryResult represents a memory entry with its similarity score
type SimilarMemoryResult struct {
	MemoryEntry *model.MemoryEntry `json:"memory_entry"`
	Similarity  float64            `json:"similarity"`
}

// FindSimilarMemories finds memories similar to the given content
// Limits search to the most recent 100 non-expired memories
func (s *MemoryService) FindSimilarMemories(ctx context.Context, workspaceID uint64, content string, threshold float64, limit int) ([]SimilarMemoryResult, error) {
	if content == "" {
		return nil, errors.New("content is required")
	}

	var entries []*model.MemoryEntry
	if err := s.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Where("expires_at IS NULL OR expires_at > NOW()").
		Order("created_at DESC").
		Limit(100).
		Find(&entries).Error; err != nil {
		return nil, err
	}

	results := make([]SimilarMemoryResult, 0)
	for _, entry := range entries {
		sim := s.CalculateTextSimilarity(content, entry.Content)
		if sim >= threshold {
			results = append(results, SimilarMemoryResult{
				MemoryEntry: entry,
				Similarity:  sim,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// MergeMemories merges multiple memories into one, marking original memories with merged_into metadata
// Requires at least 2 memories to merge
func (s *MemoryService) MergeMemories(ctx context.Context, memoryIDs []uint64, workspaceID uint64) (*model.MemoryEntry, error) {
	if len(memoryIDs) < 2 {
		return nil, ErrMergeRequiresTwoMemories
	}

	var mergedEntry *model.MemoryEntry

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var entries []*model.MemoryEntry
		if err := tx.
			Where("id IN ? AND workspace_id = ?", memoryIDs, workspaceID).
			Find(&entries).Error; err != nil {
			return err
		}

		if len(entries) != len(memoryIDs) {
			return errors.New("some memory entries not found")
		}

		var mergedContent strings.Builder
		var allTags []string
		tagSet := make(map[string]bool)
		var allMetadata map[string]interface{} = make(map[string]interface{})
		var projectID *uint64
		var issueID *uint64
		var agentID *uint64
		var memoryType model.MemoryType = model.MemoryMediumTerm
		var scope model.MemoryScope = model.ScopeWorkspace
		var contextKey string
		var contextName string

		for i, entry := range entries {
			if i > 0 {
				mergedContent.WriteString("\n\n---\n\n")
			}
			mergedContent.WriteString(entry.Content)

			if entry.Tags != nil {
				var tags []string
				if err := json.Unmarshal(entry.Tags, &tags); err == nil {
					for _, t := range tags {
						if !tagSet[t] {
							tagSet[t] = true
							allTags = append(allTags, t)
						}
					}
				}
			}

			if entry.Metadata != nil {
				var meta map[string]interface{}
				if err := json.Unmarshal(entry.Metadata, &meta); err == nil {
					for k, v := range meta {
						if _, exists := allMetadata[k]; !exists {
							allMetadata[k] = v
						}
					}
				}
			}

			if projectID == nil && entry.ProjectID != nil {
				projectID = entry.ProjectID
			}
			if issueID == nil && entry.IssueID != nil {
				issueID = entry.IssueID
			}
			if agentID == nil && entry.AgentID != nil {
				agentID = entry.AgentID
			}
			if entry.MemoryType == model.MemoryLongTerm {
				memoryType = model.MemoryLongTerm
			}
			if entry.Scope == model.ScopeProject || entry.Scope == model.ScopeIssue || entry.Scope == model.ScopeAgent {
				scope = entry.Scope
			}
			if contextKey == "" {
				contextKey = entry.ContextKey
			}
			if contextName == "" {
				contextName = entry.ContextName
			}
		}

		mergedEntry = &model.MemoryEntry{
			WorkspaceID:    workspaceID,
			ProjectID:      projectID,
			IssueID:        issueID,
			AgentID:        agentID,
			MemoryType:     memoryType,
			Scope:          scope,
			Content:        mergedContent.String(),
			ContextKey:     contextKey,
			ContextName:    contextName,
			RelevanceScore: 0.9,
		}

		if len(allTags) > 0 {
			tagsJSON, err := json.Marshal(allTags)
			if err != nil {
				return err
			}
			mergedEntry.Tags = tagsJSON
		}

		if len(allMetadata) > 0 {
			metaJSON, err := json.Marshal(allMetadata)
			if err != nil {
				return err
			}
			mergedEntry.Metadata = metaJSON
		}

		if err := tx.Create(mergedEntry).Error; err != nil {
			return err
		}

		for _, entry := range entries {
			var meta map[string]interface{}
			if entry.Metadata != nil {
				if err := json.Unmarshal(entry.Metadata, &meta); err != nil {
					return err
				}
			} else {
				meta = make(map[string]interface{})
			}
			meta["merged_into"] = mergedEntry.ID
			metaJSON, err := json.Marshal(meta)
			if err != nil {
				return err
			}
			if err := tx.
				Model(&model.MemoryEntry{}).
				Where("id = ?", entry.ID).
				Update("metadata", metaJSON).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return mergedEntry, nil
}