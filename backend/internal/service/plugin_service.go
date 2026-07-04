package service

import (
	"encoding/json"
	"log"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/plugin"
	"gorm.io/gorm"
)

// PluginService handles plugin business logic.
type PluginService struct {
	db *gorm.DB
}

// NewPluginService creates a new PluginService.
func NewPluginService(db *gorm.DB) *PluginService {
	return &PluginService{db: db}
}

// DB returns the underlying database for use by the hook manager.
func (s *PluginService) DB() *gorm.DB {
	return s.db
}

// ListInstalled returns all installed plugins for a workspace.
func (s *PluginService) ListInstalled(workspaceID uint64) ([]response.PluginResponse, error) {
	var plugins []model.Plugin
	if err := s.db.Where("workspace_id = ?", workspaceID).Order("created_at DESC").Find(&plugins).Error; err != nil {
		return nil, common.Internal("Failed to fetch plugins")
	}
	resps := make([]response.PluginResponse, len(plugins))
	for i, p := range plugins {
		resps[i] = pluginToResponse(&p)
	}
	return resps, nil
}

// Get returns a single installed plugin.
func (s *PluginService) Get(id, workspaceID uint64) (*response.PluginResponse, error) {
	var p model.Plugin
	if err := s.db.Where("id = ? AND workspace_id = ?", id, workspaceID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Plugin not found")
		}
		return nil, common.Internal("Failed to fetch plugin")
	}
	resp := pluginToResponse(&p)
	return &resp, nil
}

// Install installs a plugin from the catalog into a workspace.
func (s *PluginService) Install(workspaceID, userID uint64, req *request.PluginInstallRequest) (*response.PluginResponse, error) {
	info := plugin.GetBuiltin(req.Slug)
	if info == nil {
		return nil, common.NotFound("Plugin not found in catalog")
	}

	// Check for duplicate (same slug in same workspace)
	var count int64
	s.db.Model(&model.Plugin{}).Where("slug = ? AND workspace_id = ?", req.Slug, workspaceID).Count(&count)
	if count > 0 {
		return nil, common.Conflict("Plugin already installed in this workspace")
	}

	// Build subscribed events
	events := info.SubscribedEvents
	if req.SubscribedEvents != nil {
		events = req.SubscribedEvents
	}
	eventsJSON, _ := json.Marshal(events)

	desc := info.Description
	p := model.Plugin{
		Name:             info.Name,
		Slug:             info.Slug,
		Description:      &desc,
		Author:           info.Author,
		Version:          info.Version,
		IconURL:          info.IconURL,
		Type:             info.Type,
		EntryPoint:       info.EntryPoint,
		ConfigSchema:     info.ConfigSchema,
		Config:           normalizeJSON(req.Config),
		SubscribedEvents: eventsJSON,
		Enabled:          false,
		WorkspaceID:      workspaceID,
		InstalledByID:    userID,
	}

	if err := s.db.Create(&p).Error; err != nil {
		log.Printf("[Plugin] failed to install plugin %s: %v", req.Slug, err)
		return nil, common.Internal("Failed to install plugin")
	}

	resp := pluginToResponse(&p)
	return &resp, nil
}

// Update updates an installed plugin's configuration.
func (s *PluginService) Update(id, workspaceID uint64, req *request.PluginUpdateRequest) (*response.PluginResponse, error) {
	var p model.Plugin
	if err := s.db.Where("id = ? AND workspace_id = ?", id, workspaceID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Plugin not found")
		}
		return nil, common.Internal("Failed to fetch plugin")
	}

	updates := map[string]interface{}{}
	if req.Config != nil {
		updates["config"] = normalizeJSON(*req.Config)
	}
	if req.SubscribedEvents != nil {
		eventsJSON, _ := json.Marshal(*req.SubscribedEvents)
		updates["subscribed_events"] = eventsJSON
	}

	if len(updates) > 0 {
		if err := s.db.Model(&p).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update plugin")
		}
		s.db.First(&p, p.ID)
	}

	resp := pluginToResponse(&p)
	return &resp, nil
}

// Uninstall removes a plugin from a workspace.
func (s *PluginService) Uninstall(id, workspaceID uint64) error {
	result := s.db.Where("id = ? AND workspace_id = ?", id, workspaceID).Delete(&model.Plugin{})
	if result.Error != nil {
		return common.Internal("Failed to uninstall plugin")
	}
	if result.RowsAffected == 0 {
		return common.NotFound("Plugin not found")
	}

	// Also clean up event logs
	s.db.Where("plugin_id = ?", id).Delete(&model.PluginEventLog{})
	return nil
}

// Enable enables a plugin.
func (s *PluginService) Enable(id, workspaceID uint64) (*response.PluginResponse, error) {
	var p model.Plugin
	if err := s.db.Where("id = ? AND workspace_id = ?", id, workspaceID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Plugin not found")
		}
		return nil, common.Internal("Failed to fetch plugin")
	}

	if err := s.db.Model(&p).Update("enabled", true).Error; err != nil {
		return nil, common.Internal("Failed to enable plugin")
	}
	p.Enabled = true
	resp := pluginToResponse(&p)
	return &resp, nil
}

// Disable disables a plugin.
func (s *PluginService) Disable(id, workspaceID uint64) (*response.PluginResponse, error) {
	var p model.Plugin
	if err := s.db.Where("id = ? AND workspace_id = ?", id, workspaceID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Plugin not found")
		}
		return nil, common.Internal("Failed to fetch plugin")
	}

	if err := s.db.Model(&p).Update("enabled", false).Error; err != nil {
		return nil, common.Internal("Failed to disable plugin")
	}
	p.Enabled = false
	resp := pluginToResponse(&p)
	return &resp, nil
}

// GetEventLogs returns execution logs for a plugin instance.
func (s *PluginService) GetEventLogs(pluginID, workspaceID uint64, limit int) ([]model.PluginEventLog, error) {
	// Verify plugin belongs to workspace
	var p model.Plugin
	if err := s.db.Where("id = ? AND workspace_id = ?", pluginID, workspaceID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Plugin not found")
		}
		return nil, common.Internal("Failed to fetch plugin")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var logs []model.PluginEventLog
	if err := s.db.Where("plugin_id = ?", pluginID).Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, common.Internal("Failed to fetch event logs")
	}
	return logs, nil
}

// ==================== Helpers ====================

func normalizeJSON(raw json.RawMessage) json.RawMessage {
	if raw == nil || string(raw) == "" || string(raw) == "null" {
		return json.RawMessage("{}")
	}
	return raw
}

func pluginToResponse(p *model.Plugin) response.PluginResponse {
	return response.PluginResponse{
		ID:               p.ID,
		Name:             p.Name,
		Slug:             p.Slug,
		Description:      p.Description,
		Author:           p.Author,
		Version:          p.Version,
		IconURL:          p.IconURL,
		Type:             p.Type,
		EntryPoint:       p.EntryPoint,
		ConfigSchema:     p.ConfigSchema,
		Config:           p.Config,
		SubscribedEvents: p.SubscribedEvents,
		Enabled:          p.Enabled,
		WorkspaceID:      p.WorkspaceID,
		InstalledByID:    p.InstalledByID,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}
