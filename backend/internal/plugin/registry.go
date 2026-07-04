package plugin

import "encoding/json"

// Info describes a plugin available in the catalog (not yet installed).
type Info struct {
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Author      string          `json:"author"`
	Version     string          `json:"version"`
	Type        string          `json:"type"`
	IconURL     string          `json:"icon_url"`
	EntryPoint  string          `json:"entry_point"`
	ConfigSchema json.RawMessage `json:"config_schema"`
	SubscribedEvents []string    `json:"subscribed_events"`
}

// BuiltinCatalog returns all available built-in plugins.
func BuiltinCatalog() []Info {
	return []Info{
		{
			Slug:        "outgoing-webhook",
			Name:        "Outgoing Webhook",
			Description: "Send HTTP webhook notifications to external services when events occur in your workspace.",
			Author:      "ReqMango",
			Version:     "1.0.0",
			Type:        "webhook",
			IconURL:     "",
			ConfigSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"url": {"type": "string", "description": "Webhook target URL"},
					"secret": {"type": "string", "description": "HMAC secret for signature verification"},
					"method": {"type": "string", "enum": ["POST", "PUT"], "default": "POST"},
					"headers": {"type": "object", "description": "Custom HTTP headers"}
				},
				"required": ["url"]
			}`),
			SubscribedEvents: []string{"issue.created", "issue.updated", "comment.created", "cycle.started", "cycle.ended"},
		},
		{
			Slug:        "incoming-webhook",
			Name:        "Incoming Webhook",
			Description: "Receive webhooks from external services to automatically create or update issues.",
			Author:      "ReqMango",
			Version:     "1.0.0",
			Type:        "webhook",
			IconURL:     "",
			ConfigSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"secret": {"type": "string", "description": "HMAC secret for payload verification"},
					"default_project_id": {"type": "number", "description": "Default project for created issues"},
					"field_mapping": {"type": "object", "description": "JSON path mapping from payload to issue fields"}
				},
				"required": ["default_project_id"]
			}`),
		},
		{
			Slug:        "slack-notifier",
			Name:        "Slack Notifier",
			Description: "Send issue activity notifications to Slack channels.",
			Author:      "ReqMango",
			Version:     "1.0.0",
			Type:        "notification",
			IconURL:     "",
			ConfigSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"webhook_url": {"type": "string", "description": "Slack Incoming Webhook URL"},
					"channel": {"type": "string", "description": "Default channel (optional)"},
					"events": {"type": "array", "items": {"type": "string"}, "description": "Event types to notify on"}
				},
				"required": ["webhook_url"]
			}`),
			SubscribedEvents: []string{"issue.created", "issue.updated", "comment.created"},
		},
		{
			Slug:        "email-notifier",
			Name:        "Email Notifier",
			Description: "Send email notifications for issue events to team members.",
			Author:      "ReqMango",
			Version:     "1.0.0",
			Type:        "notification",
			IconURL:     "",
			ConfigSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"smtp_host": {"type": "string", "description": "SMTP server host"},
					"smtp_port": {"type": "number", "default": 587},
					"smtp_user": {"type": "string"},
					"smtp_password": {"type": "string"},
					"from_address": {"type": "string", "format": "email"},
					"events": {"type": "array", "items": {"type": "string"}, "description": "Event types to notify on"}
				},
				"required": ["smtp_host", "smtp_user", "smtp_password", "from_address"]
			}`),
			SubscribedEvents: []string{"issue.created", "issue.updated", "issue.deleted", "comment.created"},
		},
		{
			Slug:        "csv-importer",
			Name:        "CSV Importer",
			Description: "Import issues from CSV files with field mapping support.",
			Author:      "ReqMango",
			Version:     "1.0.0",
			Type:        "importer",
			IconURL:     "",
			ConfigSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"default_project_id": {"type": "number", "description": "Default project for imported issues"},
					"field_mapping": {"type": "object", "description": "CSV column to issue field mapping"},
					"skip_header": {"type": "boolean", "default": true}
				},
				"required": ["default_project_id"]
			}`),
		},
		{
			Slug:        "custom-script",
			Name:        "Custom Script",
			Description: "Execute custom automation scripts or API calls on workspace events.",
			Author:      "ReqMango",
			Version:     "1.0.0",
			Type:        "automation",
			IconURL:     "",
			ConfigSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"script_type": {"type": "string", "enum": ["http", "shell"], "default": "http"},
					"command": {"type": "string", "description": "HTTP URL or shell command to execute"},
					"timeout_seconds": {"type": "number", "default": 30},
					"env": {"type": "object", "description": "Environment variables or extra headers"},
					"events": {"type": "array", "items": {"type": "string"}, "description": "Event types to trigger on"}
				},
				"required": ["command"]
			}`),
			SubscribedEvents: []string{"issue.created", "issue.updated", "issue.deleted", "comment.created", "cycle.started", "cycle.ended"},
		},
	}
}

// GetBuiltin returns a built-in plugin by slug, or nil if not found.
func GetBuiltin(slug string) *Info {
	for _, p := range BuiltinCatalog() {
		if p.Slug == slug {
			return &p
		}
	}
	return nil
}
