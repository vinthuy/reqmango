package model

import (
	"encoding/json"
	"time"
)

// MemoryType defines the scope of memory
type MemoryType string

const (
	MemoryShortTerm  MemoryType = "short_term"   // Session-level, expires quickly
	MemoryMediumTerm MemoryType = "medium_term"  // Project-level, persists across sessions
	MemoryLongTerm   MemoryType = "long_term"    // Workspace-level, persists indefinitely
)

// MemoryScope defines where memory is applicable
type MemoryScope string

const (
	ScopeWorkspace MemoryScope = "workspace"
	ScopeProject   MemoryScope = "project"
	ScopeIssue     MemoryScope = "issue"
	ScopeAgent     MemoryScope = "agent"
)

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	BaseModel

	WorkspaceID uint64          `gorm:"not null;index" json:"workspace_id"`
	ProjectID   *uint64         `gorm:"index" json:"project_id,omitempty"`
	IssueID     *uint64         `gorm:"index" json:"issue_id,omitempty"`
	AgentID     *uint64         `gorm:"index" json:"agent_id,omitempty"`

	// Memory classification
	MemoryType MemoryType `gorm:"size:20;not null;index" json:"memory_type"`
	Scope      MemoryScope `gorm:"size:20;not null" json:"scope"`

	// Content
	Content     string          `gorm:"type:text;not null" json:"content"`
	Summary     string          `gorm:"-" json:"summary,omitempty"` // Generated summary for search results, not stored in DB
	Embedding   json.RawMessage `gorm:"type:jsonb" json:"embedding,omitempty"` // Vector embedding for semantic search
	Metadata    json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	Tags        json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"tags"` // Array of strings

	// Context
	ContextKey  string `gorm:"size:255;index" json:"context_key"`   // Unique key for context grouping
	ContextName string `gorm:"size:255" json:"context_name"`        // Human-readable context name

	// Relevance scoring
	RelevanceScore float64 `gorm:"type:decimal(5,4);default:0" json:"relevance_score"`

	// Expiry
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // For short-term memory

	// Relationships
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Project   *Project  `gorm:"foreignKey:ProjectID" json:"-"`
	Issue     *Issue    `gorm:"foreignKey:IssueID" json:"-"`
	Agent     *Agent    `gorm:"foreignKey:AgentID" json:"-"`
}

func (MemoryEntry) TableName() string { return "memory_entries" }

// MemorySession tracks the memory usage during a session
type MemorySession struct {
	ID          string    `gorm:"primaryKey;size:64" json:"id"`
	WorkspaceID uint64    `gorm:"not null;index" json:"workspace_id"`
	SessionType string    `gorm:"size:50;not null" json:"session_type"` // "chat", "task", "analysis", "agent"
	ContextID   *string   `gorm:"size:64" json:"context_id,omitempty"`
	StartedAt   time.Time `gorm:"autoCreateTime" json:"started_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
	MemoryCount int       `gorm:"default:0" json:"memory_count"`
}

func (MemorySession) TableName() string { return "memory_sessions" }

// MemoryVectorIndex provides inverted index for fast semantic search
type MemoryVectorIndex struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	WorkspaceID  uint64    `gorm:"not null;index" json:"workspace_id"`
	MemoryEntryID uint64   `gorm:"not null;uniqueIndex" json:"memory_entry_id"`
	VectorData   json.RawMessage `gorm:"type:jsonb" json:"vector_data"` // Normalized vector
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (MemoryVectorIndex) TableName() string { return "memory_vector_indexes" }