package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/reqmango/backend/internal/client"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// ChatService handles chat sessions, messages, reactions, and agent auto-reply.
type ChatService struct {
	db          *gorm.DB
	agentClient *client.AgentClient
	memorySvc   *MemoryService
	debouncer   *AgentReplyDebouncer
}

// NewChatService creates a ChatService. agentClient and memorySvc are optional
// (injected via setters) — when nil, agent auto-reply is disabled gracefully.
func NewChatService(db *gorm.DB, memorySvc *MemoryService) *ChatService {
	return &ChatService{
		db:        db,
		memorySvc: memorySvc,
		debouncer: NewAgentReplyDebouncer(30 * time.Second),
	}
}

// SetAgentClient injects the agent client for auto-reply. Optional.
func (s *ChatService) SetAgentClient(c *client.AgentClient) { s.agentClient = c }

// StartDebouncerCleanup launches a background goroutine that purges expired
// debounce entries every 5 minutes. Call once at startup.
func (s *ChatService) StartDebouncerCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.debouncer.Cleanup()
			}
		}
	}()
}

// GetOrCreateForIssue returns the chat for an issue, lazily creating it.
// Also returns the most recent 50 messages (oldest→newest).
func (s *ChatService) GetOrCreateForIssue(issueID, userID uint64) (*response.ChatResponse, error) {
	var issue model.Issue
	if err := s.db.Preload("Project").First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}
	if err := s.checkProjectMembership(issue.ProjectID, userID); err != nil {
		return nil, err
	}

	var chat model.Chat
	err := s.db.Where("issue_id = ? AND deleted_at IS NULL", issueID).First(&chat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		chat = model.Chat{
			WorkspaceID: issue.WorkspaceID,
			ProjectID:   &issue.ProjectID,
			IssueID:     &issueID,
			Type:        "issue",
			Title:       fmt.Sprintf("Issue #%d", issue.SequenceID),
			BaseModel:   model.BaseModel{CreatedByID: &userID},
		}
		if err := s.db.Create(&chat).Error; err != nil {
			return nil, common.Internal("Failed to create chat")
		}
	} else if err != nil {
		return nil, common.Internal("Failed to load chat")
	}

	msgs, err := s.listMessages(chat.ID, 50, "")
	if err != nil {
		return nil, err
	}

	return &response.ChatResponse{
		ID:          chat.ID,
		WorkspaceID: chat.WorkspaceID,
		ProjectID:   chat.ProjectID,
		IssueID:     chat.IssueID,
		Type:        chat.Type,
		Title:       chat.Title,
		CreatedAt:   chat.CreatedAt,
		Messages:    msgs,
	}, nil
}

// ListMessages returns messages older than the cursor (exclusive), newest→oldest,
// up to limit. The response is reversed to oldest→newest for display. next_cursor
// is the created_at of the oldest returned message (or "" if no more history).
func (s *ChatService) ListMessages(chatID, userID uint64, q request.ListMessagesQuery) (*response.ListMessagesResponse, error) {
	if err := s.checkChatMembership(chatID, userID); err != nil {
		return nil, err
	}
	limit := q.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	msgs, err := s.listMessages(chatID, limit, q.Cursor)
	if err != nil {
		return nil, err
	}
	nextCursor := ""
	if len(msgs) == limit {
		nextCursor = msgs[0].CreatedAt.Format(time.RFC3339Nano)
	}
	return &response.ListMessagesResponse{Messages: msgs, NextCursor: nextCursor}, nil
}

func (s *ChatService) listMessages(chatID uint64, limit int, cursor string) ([]response.MessageResponse, error) {
	q := s.db.Model(&model.Message{}).Where("chat_id = ? AND deleted_at IS NULL", chatID)
	if cursor != "" {
		t, err := time.Parse(time.RFC3339Nano, cursor)
		if err == nil {
			q = q.Where("created_at < ?", t)
		}
	}
	var msgs []model.Message
	if err := q.Order("created_at DESC").Limit(limit).Find(&msgs).Error; err != nil {
		return nil, common.Internal("Failed to load messages")
	}
	// Reverse to oldest→newest
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return s.toMessageResponses(msgs), nil
}

// SendMessage creates a user message, broadcasts it via SSE, notifies @mentioned
// users, and asynchronously triggers agent replies for @mentioned agents.
func (s *ChatService) SendMessage(chatID, userID uint64, req request.SendMessageRequest) (*response.MessageResponse, error) {
	if err := s.checkChatMembership(chatID, userID); err != nil {
		return nil, err
	}
	mentions := s.parseAndResolveMentions(req.Content, chatID)
	mentionsJSON, _ := json.Marshal(mentions)

	m := model.Message{
		ChatID:     chatID,
		SenderID:   userID,
		SenderType: "user",
		Content:    req.Content,
		ReplyToID:  req.ReplyToID,
		Mentions:   mentionsJSON,
	}
	if err := s.db.Create(&m).Error; err != nil {
		return nil, common.Internal("Failed to send message")
	}

	resp := s.toMessageResponses([]model.Message{m})[0]
	SSE.BroadcastToChat(chatID, "message_new", resp)

	// Notify @mentioned users via personal SSE
	for _, mn := range mentions {
		if mn.Type == "user" {
			SSE.NotifyUser(mn.ID, "mention", "你在聊天中被提及", req.Content)
		}
	}

	// Trigger agent replies asynchronously
	for _, mn := range mentions {
		if mn.Type == "agent" {
			agentID := mn.ID
			go func(aid uint64, name string) {
				if err := s.triggerAgentReply(chatID, aid, userID, "mention", req.Content); err != nil {
					log.Printf("[ChatService] agent mention reply failed (agent=%d): %v", aid, err)
				}
			}(agentID, mn.Name)
		}
	}

	return &resp, nil
}

// --- helpers ---

func (s *ChatService) checkProjectMembership(projectID, userID uint64) error {
	var count int64
	s.db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, userID, true).
		Count(&count)
	if count == 0 {
		return common.Forbidden("You must be a project member to access this chat")
	}
	return nil
}

// checkChatMembership verifies the user is a member of the project that owns
// the issue the chat is attached to.
func (s *ChatService) checkChatMembership(chatID, userID uint64) error {
	var chat model.Chat
	if err := s.db.First(&chat, chatID).Error; err != nil {
		return common.NotFound("Chat not found")
	}
	if chat.IssueID == nil {
		return common.NotFound("Chat not found")
	}
	var issue model.Issue
	if err := s.db.Select("project_id").First(&issue, *chat.IssueID).Error; err != nil {
		return common.Internal("Failed to resolve chat project")
	}
	return s.checkProjectMembership(issue.ProjectID, userID)
}

// MentionTarget is the internal shape of a parsed @mention.
type MentionTarget struct {
	Type string `json:"type"`
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// parseAndResolveMentions reuses comment_service.parseMentions to extract names,
// then resolves them to user/agent IDs scoped to the chat's workspace.
func (s *ChatService) parseAndResolveMentions(content string, chatID uint64) []response.Mention {
	names := parseMentions(content)
	if len(names) == 0 {
		return nil
	}
	var chat model.Chat
	if err := s.db.First(&chat, chatID).Error; err != nil {
		return nil
	}
	out := make([]response.Mention, 0, len(names))
	// Resolve users
	var users []model.User
	s.db.Where("username IN ?", names).Find(&users)
	for _, u := range users {
		out = append(out, response.Mention{Type: "user", ID: u.ID, Name: u.Username})
	}
	// Resolve agents (active agents in this workspace matching the names)
	var agents []model.Agent
	s.db.Where("workspace_id = ? AND name IN ? AND status = 'active'", chat.WorkspaceID, names).Find(&agents)
	for _, a := range agents {
		out = append(out, response.Mention{Type: "agent", ID: a.ID, Name: a.Name})
	}
	return out
}

func (s *ChatService) toMessageResponses(msgs []model.Message) []response.MessageResponse {
	if len(msgs) == 0 {
		return []response.MessageResponse{}
	}
	ids := make([]uint64, 0, len(msgs))
	for _, m := range msgs {
		ids = append(ids, m.ID)
	}
	var reactions []model.MessageReaction
	s.db.Where("message_id IN ?", ids).Find(&reactions)
	groups := make(map[uint64][]response.ReactionGroup)
	for _, r := range reactions {
		found := false
		for i, g := range groups[r.MessageID] {
			if g.Emoji == r.Emoji {
				groups[r.MessageID][i].Count++
				groups[r.MessageID][i].UserIDs = append(groups[r.MessageID][i].UserIDs, r.UserID)
				found = true
				break
			}
		}
		if !found {
			groups[r.MessageID] = append(groups[r.MessageID], response.ReactionGroup{
				Emoji: r.Emoji, Count: 1, UserIDs: []uint64{r.UserID},
			})
		}
	}
	out := make([]response.MessageResponse, len(msgs))
	for i, m := range msgs {
		var mn json.RawMessage
		if len(m.Mentions) > 0 {
			mn = m.Mentions
		} else {
			mn = json.RawMessage("[]")
		}
		out[i] = response.MessageResponse{
			ID: m.ID, ChatID: m.ChatID, SenderID: m.SenderID, SenderType: m.SenderType,
			Content: m.Content, ReplyToID: m.ReplyToID, Mentions: mn,
			EditedAt: m.EditedAt, DeletedAt: m.DeletedAt, CreatedAt: m.CreatedAt,
			Reactions: groups[m.ID],
		}
		if out[i].Reactions == nil {
			out[i].Reactions = []response.ReactionGroup{}
		}
	}
	return out
}

// triggerAgentReply is a stub; full implementation arrives in Task 8.
func (s *ChatService) triggerAgentReply(chatID, agentID, userID uint64, trigger, content string) error {
	return nil
}
