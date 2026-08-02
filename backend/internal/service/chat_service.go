package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
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

// --- Agent auto-reply ---

// triggerAgentReply is called asynchronously when a user @mentions an agent or
// when an issue state changes. It builds a context-aware prompt, dispatches the
// agent, and creates a sender_type=agent message with the result summary.
// Failures are silent (logged) and never write a message — see spec §5.
func (s *ChatService) triggerAgentReply(chatID, agentID, userID uint64, trigger, triggerContent string) error {
	if s.agentClient == nil {
		return nil // agent auto-reply disabled
	}
	var chat model.Chat
	if err := s.db.First(&chat, chatID).Error; err != nil {
		return err
	}
	if chat.IssueID == nil {
		return nil
	}
	issueID := *chat.IssueID

	// Debounce per agent+issue
	if !s.debouncer.Allow(agentID, issueID) {
		return nil
	}

	// Signal "agent is typing"
	SSE.BroadcastToChat(chatID, "agent_typing", map[string]interface{}{
		"chat_id": chatID, "agent_id": agentID,
	})

	task, err := s.buildAgentTask(chatID, agentID, issueID, trigger, triggerContent)
	if err != nil {
		return err
	}

	summary, err := s.agentClient.DispatchAgentWithResult(
		chat.WorkspaceID, agentID, userID, task, &issueID, chat.ProjectID, "chat:"+trigger,
	)
	if err != nil {
		return err
	}
	if strings.TrimSpace(summary) == "" {
		return nil // empty reply -> don't pollute chat
	}

	m := model.Message{
		ChatID:     chatID,
		SenderID:   agentID,
		SenderType: "agent",
		Content:    summary,
	}
	if err := s.db.Create(&m).Error; err != nil {
		return err
	}
	resp := s.toMessageResponses([]model.Message{m})[0]
	SSE.BroadcastToChat(chatID, "message_new", resp)
	return nil
}

// OnIssueStateChanged is invoked (asynchronously, via a goroutine) by
// IssueService.Update after a successful state transition. It triggers an
// agent reply for the issue's assigned agent (if any). No-op if the issue has
// no chat or no assigned agent.
func (s *ChatService) OnIssueStateChanged(ctx context.Context, issueID, oldStateID, newStateID, userID uint64) {
	if s.agentClient == nil {
		return
	}
	var chat model.Chat
	err := s.db.Where("issue_id = ? AND deleted_at IS NULL", issueID).First(&chat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return // no chat -> don't force-create
	} else if err != nil {
		log.Printf("[ChatService] OnIssueStateChanged: load chat failed: %v", err)
		return
	}

	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return
	}
	if issue.AgentAssigneeID == nil {
		return // no agent assigned -> skip
	}
	agentID := *issue.AgentAssigneeID

	var oldState, newState model.State
	s.db.First(&oldState, oldStateID)
	s.db.First(&newState, newStateID)

	triggerContent := fmt.Sprintf("状态从 %s 变为 %s", oldState.Name, newState.Name)
	// Use the issue context, not a user message, as the trigger content
	if err := s.triggerAgentReply(chat.ID, agentID, userID, "state_change", triggerContent); err != nil {
		log.Printf("[ChatService] OnIssueStateChanged: agent reply failed (agent=%d): %v", agentID, err)
	}
}

// buildAgentTask constructs the LLM prompt for an agent reply. It gathers:
//   - issue context (title, type, priority, description)
//   - state transition context (if trigger == "state_change")
//   - relevant memories via MemoryService.SemanticSearchByText (degrades gracefully)
//   - the 10 most recent chat messages
func (s *ChatService) buildAgentTask(chatID, agentID, issueID uint64, trigger, triggerContent string) (string, error) {
	var issue model.Issue
	if err := s.db.Preload("Project").First(&issue, issueID).Error; err != nil {
		return "", err
	}
	descStripped := ""
	if issue.DescriptionStripped != nil {
		descStripped = *issue.DescriptionStripped
	}
	if len(descStripped) > 500 {
		descStripped = descStripped[:500]
	}

	var agent model.Agent
	if err := s.db.First(&agent, agentID).Error; err != nil {
		return "", err
	}

	// Recent messages (10, newest first, exclude nothing)
	var recent []model.Message
	s.db.Where("chat_id = ? AND deleted_at IS NULL", chatID).
		Order("created_at DESC").Limit(10).Find(&recent)
	// Reverse to chronological
	for i, j := 0, len(recent)-1; i < j; i, j = i+1, j-1 {
		recent[i], recent[j] = recent[j], recent[i]
	}

	// Memory retrieval (degrades gracefully on failure)
	var memories []string
	if s.memorySvc != nil {
		query := triggerContent
		if trigger == "mention" {
			query = triggerContent
		}
		entries, err := s.memorySvc.SemanticSearchByText(context.Background(), issue.WorkspaceID, query, 5)
		if err == nil {
			for _, e := range entries {
				if e.Content != "" {
					memories = append(memories, e.Content)
				}
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("你是被分配到工作项 #%d 的 %s。\n\n", issue.SequenceID, agent.Name))
	sb.WriteString("[工作项上下文]\n")
	sb.WriteString(fmt.Sprintf("- 标题: %s\n", issue.Name))
	sb.WriteString(fmt.Sprintf("- 优先级: %s\n", issue.Priority))
	sb.WriteString(fmt.Sprintf("- 描述: %s\n", descStripped))
	if trigger == "state_change" {
		sb.WriteString(fmt.Sprintf("[触发] %s\n", triggerContent))
	} else {
		sb.WriteString(fmt.Sprintf("[触发] 用户消息: %s\n", triggerContent))
	}

	if len(memories) > 0 {
		sb.WriteString("\n[相关记忆]\n")
		for i, m := range memories {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, m))
		}
	}

	if len(recent) > 0 {
		sb.WriteString("\n[最近对话]\n")
		for _, m := range recent {
			role := "用户"
			if m.SenderType == "agent" {
				role = "Agent"
			}
			sb.WriteString(fmt.Sprintf("[%s] %s\n", role, m.Content))
		}
	}

	sb.WriteString("\n[任务]\n基于上下文，提供 1-3 句简明建议。不要重复已知信息。\n回复:")
	return sb.String(), nil
}

// --- Edit / Delete / Reactions ---

// EditMessage updates a message's content. Only the author may edit, and only
// within 30 minutes of creation. EditedAt is stamped.
func (s *ChatService) EditMessage(messageID, userID uint64, req request.EditMessageRequest) (*response.MessageResponse, error) {
	var m model.Message
	if err := s.db.First(&m, messageID).Error; err != nil {
		return nil, common.NotFound("Message not found")
	}
	if m.DeletedAt != nil {
		return nil, common.NotFound("Message not found")
	}
	if m.SenderType != "user" || m.SenderID != userID {
		return nil, common.Forbidden("You can only edit your own messages")
	}
	if time.Since(m.CreatedAt) > 30*time.Minute {
		return nil, common.Forbidden("Edit window (30 minutes) has expired")
	}
	m.Content = req.Content
	m.Mentions = mustJSON(s.parseAndResolveMentions(req.Content, m.ChatID))
	now := time.Now()
	m.EditedAt = &now
	if err := s.db.Save(&m).Error; err != nil {
		return nil, common.Internal("Failed to edit message")
	}
	resp := s.toMessageResponses([]model.Message{m})[0]
	SSE.BroadcastToChat(m.ChatID, "message_edited", resp)
	return &resp, nil
}

// DeleteMessage soft-deletes a message. The author or a project admin may delete.
// Content is cleared to avoid leaking; the message row is retained for context.
func (s *ChatService) DeleteMessage(messageID, userID uint64) error {
	var m model.Message
	if err := s.db.First(&m, messageID).Error; err != nil {
		return common.NotFound("Message not found")
	}
	if m.DeletedAt != nil {
		return nil // idempotent
	}
	isAuthor := m.SenderType == "user" && m.SenderID == userID
	if !isAuthor {
		// Allow project admins (workspace owner) to delete any message
		if err := s.checkChatMembership(m.ChatID, userID); err != nil {
			return err
		}
		// Only the issue's project admin (workspace owner) may delete others' messages
		var chat model.Chat
		if err := s.db.First(&chat, m.ChatID).Error; err != nil {
			return common.Internal("Failed to load chat")
		}
		if chat.IssueID == nil {
			return common.Forbidden("Forbidden")
		}
		var issue model.Issue
		if err := s.db.Preload("Project").First(&issue, *chat.IssueID).Error; err != nil {
			return common.Internal("Failed to load issue")
		}
		var ws model.Workspace
		if err := s.db.First(&ws, issue.WorkspaceID).Error; err != nil {
			return common.Internal("Failed to load workspace")
		}
		if ws.OwnerID != userID {
			return common.Forbidden("Only the author or a workspace owner may delete messages")
		}
	}
	now := time.Now()
	m.DeletedAt = &now
	m.Content = ""
	if err := s.db.Save(&m).Error; err != nil {
		return common.Internal("Failed to delete message")
	}
	SSE.BroadcastToChat(m.ChatID, "message_deleted", map[string]interface{}{
		"id": m.ID, "deleted_at": m.DeletedAt,
	})
	return nil
}

// AddReaction adds an emoji reaction (idempotent via DB UNIQUE constraint).
func (s *ChatService) AddReaction(messageID, userID uint64, emoji string) error {
	if err := s.checkMessageMembership(messageID, userID); err != nil {
		return err
	}
	r := model.MessageReaction{MessageID: messageID, UserID: userID, Emoji: emoji}
	if err := s.db.Create(&r).Error; err != nil {
		// UNIQUE violation -> already exists, treat as success (idempotent)
		// GORM returns error; we ignore the duplicate-key case heuristically.
		// PostgreSQL error code 23505 would be ideal, but string match is fine.
		if !isDuplicateKeyErr(err) {
			return common.Internal("Failed to add reaction")
		}
	}
	SSE.BroadcastToChat(s.messageChatID(messageID), "reaction_added", map[string]interface{}{
		"message_id": messageID, "user_id": userID, "emoji": emoji,
	})
	return nil
}

// RemoveReaction removes an emoji reaction (idempotent).
func (s *ChatService) RemoveReaction(messageID, userID uint64, emoji string) error {
	if err := s.checkMessageMembership(messageID, userID); err != nil {
		return err
	}
	s.db.Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Delete(&model.MessageReaction{})
	SSE.BroadcastToChat(s.messageChatID(messageID), "reaction_removed", map[string]interface{}{
		"message_id": messageID, "user_id": userID, "emoji": emoji,
	})
	return nil
}

func (s *ChatService) checkMessageMembership(messageID, userID uint64) error {
	var m model.Message
	if err := s.db.Select("chat_id").First(&m, messageID).Error; err != nil {
		return common.NotFound("Message not found")
	}
	return s.checkChatMembership(m.ChatID, userID)
}

func (s *ChatService) messageChatID(messageID uint64) uint64 {
	var m model.Message
	s.db.Select("chat_id").First(&m, messageID)
	return m.ChatID
}

func isDuplicateKeyErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}

func mustJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	if b == nil {
		return json.RawMessage("[]")
	}
	return b
}

// GetChat returns a chat by ID (without messages). Caller must have already
// verified membership via GetChatMembershipCheck.
func (s *ChatService) GetChat(chatID uint64) (*response.ChatResponse, error) {
	var chat model.Chat
	if err := s.db.First(&chat, chatID).Error; err != nil {
		return nil, common.NotFound("Chat not found")
	}
	return &response.ChatResponse{
		ID:          chat.ID,
		WorkspaceID: chat.WorkspaceID,
		ProjectID:   chat.ProjectID,
		IssueID:     chat.IssueID,
		Type:        chat.Type,
		Title:       chat.Title,
		CreatedAt:   chat.CreatedAt,
		Messages:    []response.MessageResponse{},
	}, nil
}

// GetChatMembershipCheck verifies the user is a project member of the chat's
// issue. Used by SSE subscription + GetChat.
func (s *ChatService) GetChatMembershipCheck(chatID, userID uint64) error {
	return s.checkChatMembership(chatID, userID)
}
