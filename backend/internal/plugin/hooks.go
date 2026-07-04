package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// EventType constants for system events that plugins can subscribe to.
const (
	EventIssueCreated  = "issue.created"
	EventIssueUpdated  = "issue.updated"
	EventIssueDeleted  = "issue.deleted"
	EventCommentCreated = "comment.created"
	EventCycleStarted  = "cycle.started"
	EventCycleEnded    = "cycle.ended"
)

// EventPayload carries data about a system event for plugin consumption.
type EventPayload struct {
	EventType  string          `json:"event_type"`
	WorkspaceID uint64         `json:"workspace_id"`
	ProjectID  *uint64         `json:"project_id,omitempty"`
	ActorID    uint64          `json:"actor_id"`
	Data       json.RawMessage `json:"data"`
	Timestamp  time.Time       `json:"timestamp"`
}

// HookManager dispatches events to enabled plugins.
type HookManager struct {
	db *gorm.DB
}

// NewHookManager creates a new HookManager.
func NewHookManager(db *gorm.DB) *HookManager {
	return &HookManager{db: db}
}

// Dispatch sends an event to all enabled plugins in the workspace that subscribe to it.
func (m *HookManager) Dispatch(payload EventPayload) {
	var plugins []model.Plugin
	if err := m.db.Where("workspace_id = ? AND enabled = ?", payload.WorkspaceID, true).Find(&plugins).Error; err != nil {
		log.Printf("[PluginHook] failed to query plugins for workspace %d: %v", payload.WorkspaceID, err)
		return
	}

	for _, p := range plugins {
		if !m.isSubscribed(p, payload.EventType) {
			continue
		}
		go m.executePlugin(p, payload)
	}
}

// isSubscribed checks if the plugin subscribes to the given event type.
func (m *HookManager) isSubscribed(p model.Plugin, eventType string) bool {
	if p.SubscribedEvents == nil {
		return false
	}
	var events []string
	if err := json.Unmarshal(p.SubscribedEvents, &events); err != nil {
		return false
	}
	for _, e := range events {
		if e == eventType {
			return true
		}
	}
	return false
}

// executePlugin runs a single plugin against an event payload.
func (m *HookManager) executePlugin(p model.Plugin, payload EventPayload) {
	bodyBytes, _ := json.Marshal(payload)
	start := time.Now()

	var status string
	var respBody string
	var statusCode int

	switch p.Type {
	case "webhook", "notification", "automation":
		status, respBody, statusCode = m.executeHTTP(p, bodyBytes)
	default:
		status = "skipped"
		respBody = fmt.Sprintf("unsupported plugin type: %s", p.Type)
		statusCode = 0
	}

	duration := time.Since(start).Milliseconds()

	// Log execution
	logEntry := model.PluginEventLog{
		PluginID:     p.ID,
		EventType:    payload.EventType,
		Status:       status,
		RequestBody:  string(bodyBytes),
		ResponseBody: respBody,
		StatusCode:   statusCode,
		DurationMs:   duration,
	}
	if err := m.db.Create(&logEntry).Error; err != nil {
		log.Printf("[PluginHook] failed to log execution for plugin %d: %v", p.ID, err)
	}
}

// executeHTTP calls the plugin's entrypoint URL with the event payload.
func (m *HookManager) executeHTTP(p model.Plugin, body []byte) (status, respBody string, statusCode int) {
	// Parse config once
	var cfg map[string]interface{}
	if p.Config != nil {
		json.Unmarshal(p.Config, &cfg)
	}

	entrypoint := p.EntryPoint
	if entrypoint == "" && cfg != nil {
		if url, ok := cfg["url"].(string); ok {
			entrypoint = url
		}
	}
	if entrypoint == "" {
		return "error", "no entrypoint configured", 0
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", entrypoint, http.NoBody)
	if err != nil {
		return "error", err.Error(), 0
	}

	// Override method from config
	if cfg != nil {
		if method, ok := cfg["method"].(string); ok && method != "" {
			req.Method = method
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Plugin-Event", "true")

	// Add custom headers from config
	if cfg != nil {
		if headers, ok := cfg["headers"].(map[string]interface{}); ok {
			for k, v := range headers {
				if vs, ok := v.(string); ok {
					req.Header.Set(k, vs)
				}
			}
		}
	}

	// Set the payload body
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))

	resp, err := client.Do(req)
	if err != nil {
		return "error", err.Error(), 0
	}
	defer resp.Body.Close()

	// Read response body (limited size)
	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	respBody = string(buf[:n])

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status = "success"
	} else {
		status = "error"
	}
	statusCode = resp.StatusCode
	return
}
