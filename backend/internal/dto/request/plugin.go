package request

import "encoding/json"

// PluginInstallRequest represents a request to install a plugin from the catalog.
type PluginInstallRequest struct {
	Slug        string          `json:"slug" binding:"required"`
	Config      json.RawMessage `json:"config"`
	SubscribedEvents []string   `json:"subscribed_events"`
}

// PluginUpdateRequest represents a request to update an installed plugin.
type PluginUpdateRequest struct {
	Config           *json.RawMessage `json:"config"`
	SubscribedEvents *[]string        `json:"subscribed_events"`
}
