package registry

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// AgentEntry represents a registered agent in the registry.
type AgentEntry struct {
	ID           uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkspaceID  *uint64         `gorm:"index" json:"workspace_id,omitempty"`
	Name         string          `gorm:"size:255;not null;uniqueIndex" json:"name"`
	DisplayName  string          `gorm:"size:255" json:"display_name"`
	Description  *string         `gorm:"type:text" json:"description,omitempty"`
	Capabilities json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"capabilities"`
	AgentDef     json.RawMessage `gorm:"type:jsonb;not null" json:"agent_def"`
	Version      string          `gorm:"size:50;default:1.0.0" json:"version"`
	Author       string          `gorm:"size:255" json:"author,omitempty"`
	IsVerified   bool            `gorm:"default:false" json:"is_verified"`
	Installs     int             `gorm:"default:0" json:"installs"`
	Rating       *float64        `gorm:"type:decimal(3,2)" json:"rating,omitempty"`
	Source       string          `gorm:"size:50;default:local" json:"source"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (AgentEntry) TableName() string { return "agent_registry" }

// Registry manages agent registration and discovery.
type Registry struct {
	db *gorm.DB
}

func NewRegistry(db *gorm.DB) *Registry {
	return &Registry{db: db}
}

// Register adds a new agent to the registry.
func (r *Registry) Register(entry *AgentEntry) error {
	return r.db.Create(entry).Error
}

// Find discovers agents by required capabilities.
func (r *Registry) Find(workspaceID *uint64, capabilities ...string) ([]AgentEntry, error) {
	var entries []AgentEntry
	query := r.db.Where("deleted_at IS NULL")

	if workspaceID != nil {
		query = query.Where("workspace_id = ? OR workspace_id IS NULL", *workspaceID)
	}

	if err := query.Order("is_verified DESC, installs DESC").Find(&entries).Error; err != nil {
		return nil, err
	}

	if len(capabilities) > 0 {
		var filtered []AgentEntry
		for _, entry := range entries {
			var caps []string
			if err := json.Unmarshal(entry.Capabilities, &caps); err != nil {
				continue
			}
			if hasAllCapabilities(caps, capabilities) {
				filtered = append(filtered, entry)
			}
		}
		return filtered, nil
	}

	return entries, nil
}

// GetByID retrieves an agent by ID.
func (r *Registry) GetByID(id uint64) (*AgentEntry, error) {
	var entry AgentEntry
	if err := r.db.First(&entry, id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// ListByWorkspace returns all agents for a workspace.
func (r *Registry) ListByWorkspace(wsID uint64) ([]AgentEntry, error) {
	var entries []AgentEntry
	if err := r.db.Where("workspace_id = ? OR workspace_id IS NULL", wsID).
		Order("created_at DESC").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

// Update modifies an agent entry.
func (r *Registry) Update(id uint64, updates map[string]interface{}) error {
	return r.db.Model(&AgentEntry{}).Where("id = ?", id).Updates(updates).Error
}

// Delete soft-deletes an agent.
func (r *Registry) Delete(id uint64) error {
	return r.db.Delete(&AgentEntry{}, id).Error
}

// IncrementInstalls increments the install counter.
func (r *Registry) IncrementInstalls(id uint64) error {
	return r.db.Model(&AgentEntry{}).Where("id = ?", id).
		UpdateColumn("installs", gorm.Expr("installs + 1")).Error
}

// SetRating updates the agent rating.
func (r *Registry) SetRating(id uint64, rating float64) error {
	return r.db.Model(&AgentEntry{}).Where("id = ?", id).
		Update("rating", rating).Error
}

// SeedDefaults creates built-in agents if they don't exist.
func (r *Registry) SeedDefaults(workspaceID *uint64) error {
	defaults := []AgentEntry{
		{
			WorkspaceID:  workspaceID,
			Name:         "sprint-analyzer",
			DisplayName:  "Sprint Analyzer",
			Capabilities: toJSON([]string{"planning", "analysis", "sprint"}),
			AgentDef:     toJSON(map[string]interface{}{"role": "planner", "model": "claude-opus-4-8"}),
			Version:      "1.0.0",
			Source:       "builtin",
			IsVerified:   true,
		},
		{
			WorkspaceID:  workspaceID,
			Name:         "triage-agent",
			DisplayName:  "Triage Agent",
			Capabilities: toJSON([]string{"triage", "classification", "auto_assign"}),
			AgentDef:     toJSON(map[string]interface{}{"role": "executor", "model": "claude-sonnet-4-6"}),
			Version:      "1.0.0",
			Source:       "builtin",
			IsVerified:   true,
		},
		{
			WorkspaceID:  workspaceID,
			Name:         "code-reviewer",
			DisplayName:  "Code Reviewer",
			Capabilities: toJSON([]string{"review", "code_review", "adversarial"}),
			AgentDef:     toJSON(map[string]interface{}{"role": "reviewer", "model": "claude-sonnet-4-6", "mode": "adversarial"}),
			Version:      "1.0.0",
			Source:       "builtin",
			IsVerified:   true,
		},
	}

	for _, entry := range defaults {
		var count int64
		r.db.Model(&AgentEntry{}).Where("name = ? AND (workspace_id = ? OR workspace_id IS NULL)", entry.Name, workspaceID).Count(&count)
		if count == 0 {
			if err := r.db.Create(&entry).Error; err != nil {
				return fmt.Errorf("failed to seed agent %s: %w", entry.Name, err)
			}
		}
	}
	return nil
}

func toJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func hasAllCapabilities(have, need []string) bool {
	capSet := make(map[string]bool)
	for _, c := range have {
		capSet[c] = true
	}
	for _, c := range need {
		if !capSet[c] {
			return false
		}
	}
	return true
}
