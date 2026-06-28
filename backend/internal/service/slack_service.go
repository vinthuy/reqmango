package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type SlackService struct {
	db *gorm.DB
}

func NewSlackService(db *gorm.DB) *SlackService {
	return &SlackService{db: db}
}

// ======== Request/Response types ========

type SlackCreateRequest struct {
	ProjectID        uint64 `json:"project_id" binding:"required"`
	ChannelName      string `json:"channel_name" binding:"required"`
	WebhookURL       string `json:"webhook_url" binding:"required"`
	BotToken         string `json:"bot_token"`
	IsEnabled        *bool  `json:"is_enabled"`
	NotifyOnCreate   *bool  `json:"notify_on_create"`
	NotifyOnUpdate   *bool  `json:"notify_on_update"`
	NotifyOnComment  *bool  `json:"notify_on_comment"`
	NotifyOnComplete *bool  `json:"notify_on_complete"`
}

type SlackUpdateRequest struct {
	ChannelName      *string `json:"channel_name"`
	WebhookURL       *string `json:"webhook_url"`
	BotToken         *string `json:"bot_token"`
	IsEnabled        *bool   `json:"is_enabled"`
	NotifyOnCreate   *bool   `json:"notify_on_create"`
	NotifyOnUpdate   *bool   `json:"notify_on_update"`
	NotifyOnComment  *bool   `json:"notify_on_comment"`
	NotifyOnComplete *bool   `json:"notify_on_complete"`
}

type SlackResponse struct {
	ID               uint64 `json:"id"`
	WorkspaceID      uint64 `json:"workspace_id"`
	ProjectID        uint64 `json:"project_id"`
	ChannelName      string `json:"channel_name"`
	IsEnabled        bool   `json:"is_enabled"`
	NotifyOnCreate   bool   `json:"notify_on_create"`
	NotifyOnUpdate   bool   `json:"notify_on_update"`
	NotifyOnComment  bool   `json:"notify_on_comment"`
	NotifyOnComplete bool   `json:"notify_on_complete"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type SlackNotification struct {
	IssueID   uint64 `json:"issue_id"`
	IssueName string `json:"issue_name"`
	Event     string `json:"event"` // "created", "updated", "commented", "completed"
	User      string `json:"user"`
	URL       string `json:"url"`
}

// ======== CRUD ========

func (s *SlackService) List(workspaceID uint64) ([]SlackResponse, error) {
	var conns []model.SlackConnection
	if err := s.db.Where("workspace_id = ?", workspaceID).Find(&conns).Error; err != nil {
		return nil, common.Internal("Failed to list Slack connections")
	}
	res := make([]SlackResponse, len(conns))
	for i, c := range conns {
		res[i] = s.toResponse(&c)
	}
	if res == nil {
		res = []SlackResponse{}
	}
	return res, nil
}

func (s *SlackService) Get(id uint64) (*SlackResponse, error) {
	var conn model.SlackConnection
	if err := s.db.First(&conn, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Slack connection not found")
		}
		return nil, common.Internal("Failed to get Slack connection")
	}
	r := s.toResponse(&conn)
	return &r, nil
}

func (s *SlackService) Create(workspaceID uint64, req *SlackCreateRequest) (*SlackResponse, error) {
	enabled := true
	if req.IsEnabled != nil {
		enabled = *req.IsEnabled
	}
	onCreate := true
	if req.NotifyOnCreate != nil {
		onCreate = *req.NotifyOnCreate
	}
	onUpdate := true
	if req.NotifyOnUpdate != nil {
		onUpdate = *req.NotifyOnUpdate
	}
	onComment := false
	if req.NotifyOnComment != nil {
		onComment = *req.NotifyOnComment
	}
	onComplete := true
	if req.NotifyOnComplete != nil {
		onComplete = *req.NotifyOnComplete
	}

	conn := model.SlackConnection{
		WorkspaceID:      workspaceID,
		ProjectID:        req.ProjectID,
		ChannelName:      req.ChannelName,
		WebhookURL:       req.WebhookURL,
		BotToken:         req.BotToken,
		IsEnabled:        enabled,
		NotifyOnCreate:   onCreate,
		NotifyOnUpdate:   onUpdate,
		NotifyOnComment:  onComment,
		NotifyOnComplete: onComplete,
	}

	if err := s.db.Create(&conn).Error; err != nil {
		return nil, common.Internal("Failed to create Slack connection")
	}
	r := s.toResponse(&conn)
	return &r, nil
}

func (s *SlackService) Update(id uint64, req *SlackUpdateRequest) (*SlackResponse, error) {
	var conn model.SlackConnection
	if err := s.db.First(&conn, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Slack connection not found")
		}
		return nil, common.Internal("Failed to get Slack connection")
	}

	updates := map[string]interface{}{}
	if req.ChannelName != nil {
		updates["channel_name"] = *req.ChannelName
	}
	if req.WebhookURL != nil {
		updates["webhook_url"] = *req.WebhookURL
	}
	if req.BotToken != nil {
		updates["bot_token"] = *req.BotToken
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.NotifyOnCreate != nil {
		updates["notify_on_create"] = *req.NotifyOnCreate
	}
	if req.NotifyOnUpdate != nil {
		updates["notify_on_update"] = *req.NotifyOnUpdate
	}
	if req.NotifyOnComment != nil {
		updates["notify_on_comment"] = *req.NotifyOnComment
	}
	if req.NotifyOnComplete != nil {
		updates["notify_on_complete"] = *req.NotifyOnComplete
	}

	if err := s.db.Model(&conn).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to update Slack connection")
	}

	s.db.First(&conn, id)
	r := s.toResponse(&conn)
	return &r, nil
}

func (s *SlackService) Delete(id uint64) error {
	var conn model.SlackConnection
	if err := s.db.First(&conn, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("Slack connection not found")
		}
		return common.Internal("Failed to get Slack connection")
	}
	if err := s.db.Delete(&conn).Error; err != nil {
		return common.Internal("Failed to delete Slack connection")
	}
	return nil
}

// ======== Notification ========

// SendNotification sends a formatted Slack message to the channel.
func (s *SlackService) SendNotification(id uint64, notif *SlackNotification) error {
	var conn model.SlackConnection
	if err := s.db.First(&conn, id).Error; err != nil {
		return common.NotFound("Slack connection not found")
	}

	if !conn.IsEnabled {
		return nil
	}

	// Check notification rules
	switch notif.Event {
	case "created":
		if !conn.NotifyOnCreate {
			return nil
		}
	case "updated":
		if !conn.NotifyOnUpdate {
			return nil
		}
	case "commented":
		if !conn.NotifyOnComment {
			return nil
		}
	case "completed":
		if !conn.NotifyOnComplete {
			return nil
		}
	}

	return s.sendToSlack(conn.WebhookURL, notif)
}

// TestNotification sends a test message to verify the Slack connection.
func (s *SlackService) TestNotification(id uint64) (map[string]interface{}, error) {
	var conn model.SlackConnection
	if err := s.db.First(&conn, id).Error; err != nil {
		return nil, common.NotFound("Slack connection not found")
	}

	testNotif := &SlackNotification{
		IssueID:   0,
		IssueName: "Test Notification",
		Event:     "created",
		User:      "ReqMan",
		URL:       "",
	}

	err := s.sendToSlack(conn.WebhookURL, testNotif)
	if err != nil {
		return nil, common.Internal(fmt.Sprintf("Failed to send test: %v", err))
	}
	return map[string]interface{}{"status": "sent", "channel": conn.ChannelName}, nil
}

func (s *SlackService) sendToSlack(webhookURL string, notif *SlackNotification) error {
	color := "#36a64f" // green
	switch notif.Event {
	case "updated":
		color = "#439FE0" // blue
	case "commented":
		color = "#FFA500" // orange
	case "completed":
		color = "#9370DB" // purple
	}

	emoji := map[string]string{
		"created":   ":new:",
		"updated":   ":pencil:",
		"commented": ":speech_balloon:",
		"completed": ":white_check_mark:",
	}[notif.Event]

	title := fmt.Sprintf("%s Issue %s: %s", emoji, notif.Event, notif.IssueName)

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  title,
				"fields": []map[string]interface{}{
					{"title": "Issue", "value": notif.IssueName, "short": true},
					{"title": "Event", "value": notif.Event, "short": true},
					{"title": "User", "value": notif.User, "short": true},
				},
				"footer":     "ReqMan",
				"ts":         time.Now().Unix(),
			},
		},
	}

	if notif.URL != "" {
		payload["text"] = fmt.Sprintf("<%s|View in ReqMan>", notif.URL)
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Slack webhook returned %d", resp.StatusCode)
	}
	return nil
}

// ======== Helpers ========

func (s *SlackService) toResponse(conn *model.SlackConnection) SlackResponse {
	return SlackResponse{
		ID:               conn.ID,
		WorkspaceID:      conn.WorkspaceID,
		ProjectID:        conn.ProjectID,
		ChannelName:      conn.ChannelName,
		IsEnabled:        conn.IsEnabled,
		NotifyOnCreate:   conn.NotifyOnCreate,
		NotifyOnUpdate:   conn.NotifyOnUpdate,
		NotifyOnComment:  conn.NotifyOnComment,
		NotifyOnComplete: conn.NotifyOnComplete,
		CreatedAt:        conn.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        conn.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
