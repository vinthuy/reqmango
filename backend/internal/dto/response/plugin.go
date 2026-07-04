package response

import (
	"encoding/json"
	"time"
)

// PluginResponse is the API response for an installed plugin.
type PluginResponse struct {
	ID               uint64          `json:"id"`
	Name             string          `json:"name"`
	Slug             string          `json:"slug"`
	Description      *string         `json:"description"`
	Author           string          `json:"author"`
	Version          string          `json:"version"`
	IconURL          string          `json:"icon_url"`
	Type             string          `json:"type"`
	EntryPoint       string          `json:"entry_point"`
	ConfigSchema     json.RawMessage `json:"config_schema"`
	Config           json.RawMessage `json:"config"`
	SubscribedEvents json.RawMessage `json:"subscribed_events"`
	Enabled          bool            `json:"enabled"`
	WorkspaceID      uint64          `json:"workspace_id"`
	InstalledByID    uint64          `json:"installed_by_id"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}
