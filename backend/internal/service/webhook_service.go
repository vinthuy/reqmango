package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type WebhookService struct{ db *gorm.DB }

func NewWebhookService(db *gorm.DB) *WebhookService { return &WebhookService{db: db} }

func (s *WebhookService) List(projectID uint64) ([]model.Webhook, error) {
	var wh []model.Webhook
	s.db.Where("project_id = ?", projectID).Find(&wh)
	return wh, nil
}

func (s *WebhookService) Create(projectID, workspaceID uint64, req *struct {
	Name, URL, Secret, Events string
}) (*model.Webhook, error) {
	events := req.Events
	if events == "" { events = "issue_created,issue_updated,state_changed" }
	w := &model.Webhook{Name: req.Name, URL: req.URL, Secret: req.Secret, Events: events, IsActive: true, ProjectID: projectID, WorkspaceID: workspaceID}
	if err := s.db.Create(w).Error; err != nil { return nil, common.Internal("Failed to create webhook") }
	return w, nil
}

func (s *WebhookService) Update(id uint64, req *struct {
	Name, URL, Secret, Events *string; IsActive *bool
}) (*model.Webhook, error) {
	var w model.Webhook
	if s.db.First(&w, id).Error != nil { return nil, common.NotFound("Webhook not found") }
	u := map[string]interface{}{}
	if req.Name != nil { u["name"] = *req.Name }
	if req.URL != nil { u["url"] = *req.URL }
	if req.Secret != nil && *req.Secret != "" { u["secret"] = *req.Secret }
	if req.Events != nil { u["events"] = *req.Events }
	if req.IsActive != nil { u["is_active"] = *req.IsActive }
	s.db.Model(&w).Updates(u); s.db.First(&w, id)
	return &w, nil
}

func (s *WebhookService) Delete(id uint64) error {
	if s.db.Delete(&model.Webhook{}, id).RowsAffected == 0 { return common.NotFound("Webhook not found") }
	return nil
}

// Fire sends webhook payloads to all matching webhooks for the given event.
func (s *WebhookService) Fire(projectID uint64, event string, payload map[string]interface{}) {
	var webhooks []model.Webhook
	s.db.Where("project_id = ? AND is_active = ?", projectID, true).Find(&webhooks)
	for _, w := range webhooks {
		if !strings.Contains(w.Events, event) { continue }
		go s.send(w, event, payload)
	}
}

func (s *WebhookService) send(w model.Webhook, event string, payload map[string]interface{}) {
	body := map[string]interface{}{"event": event, "timestamp": time.Now().UTC().Format(time.RFC3339), "payload": payload}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", w.URL, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ReqMan-Event", event)
	if w.Secret != "" {
		mac := hmac.New(sha256.New, []byte(w.Secret))
		mac.Write(b)
		req.Header.Set("X-ReqMan-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}

	maxRetries := 3
	backoff := 1 * time.Second
	cli := &http.Client{Timeout: 10 * time.Second}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := cli.Do(req)
		if err != nil {
			fmt.Printf("webhook %s attempt %d error: %v\n", w.Name, attempt, err)
			if attempt < maxRetries {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return
		}
		if resp.StatusCode >= 500 && attempt < maxRetries {
			fmt.Printf("webhook %s attempt %d returned %d, retrying...\n", w.Name, attempt, resp.StatusCode)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		fmt.Printf("webhook %s attempt %d returned %d\n", w.Name, attempt, resp.StatusCode)
		return
	}
}
