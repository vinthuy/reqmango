package client

import (
	"context"
	"strconv"
	"time"
)

// Message is a chat message.
type Message struct {
	ID         uint64     `json:"id"`
	ChatID     uint64     `json:"chat_id"`
	SenderID   uint64     `json:"sender_id"`
	SenderType string     `json:"sender_type"` // "user" | "agent"
	Content    string     `json:"content"`
	ReplyToID  *uint64    `json:"reply_to_id"`
	EditedAt   *time.Time `json:"edited_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Chat is the GET /issues/:id/chat response.
type Chat struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspace_id"`
	ProjectID   *uint64   `json:"project_id"`
	IssueID     *uint64   `json:"issue_id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"created_at"`
	Messages    []Message `json:"messages"`
}

// GetIssueChat returns (or creates) the chat for an issue.
func (c *Client) GetIssueChat(ctx context.Context, issueID uint64) (*Chat, error) {
	var out Chat
	if _, err := c.GetJSON(ctx, "/issues/"+strconv.FormatUint(issueID, 10)+"/chat", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendMessage posts a user message to a chat (POST /chats/:id/messages).
func (c *Client) SendMessage(ctx context.Context, chatID uint64, content string) (*Message, error) {
	var out Message
	_, err := c.PostJSON(ctx, "/chats/"+strconv.FormatUint(chatID, 10)+"/messages", nil,
		map[string]string{"content": content}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
