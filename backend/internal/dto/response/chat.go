package response

import (
	"encoding/json"
	"time"
)

// Mention is a parsed @mention target.
type Mention struct {
	Type string `json:"type"` // user | agent
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// MessageResponse is the API shape for a single chat message.
type MessageResponse struct {
	ID         uint64          `json:"id"`
	ChatID     uint64          `json:"chat_id"`
	SenderID   uint64          `json:"sender_id"`
	SenderType string          `json:"sender_type"` // user | agent
	Content    string          `json:"content"`
	ReplyToID  *uint64         `json:"reply_to_id"`
	Mentions   json.RawMessage `json:"mentions"`
	EditedAt   *time.Time      `json:"edited_at"`
	DeletedAt  *time.Time      `json:"deleted_at"`
	CreatedAt  time.Time       `json:"created_at"`
	Reactions  []ReactionGroup `json:"reactions"`
}

// ReactionGroup aggregates reactions per emoji on a message.
type ReactionGroup struct {
	Emoji   string   `json:"emoji"`
	Count   int      `json:"count"`
	UserIDs []uint64 `json:"user_ids"`
}

// ChatResponse is the API shape for a chat session.
type ChatResponse struct {
	ID          uint64            `json:"id"`
	WorkspaceID uint64            `json:"workspace_id"`
	ProjectID   *uint64           `json:"project_id"`
	IssueID     *uint64           `json:"issue_id"`
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	CreatedAt   time.Time         `json:"created_at"`
	Messages    []MessageResponse `json:"messages"` // only populated for GetOrCreateForIssue
}

// ListMessagesResponse is the paginated history response.
type ListMessagesResponse struct {
	Messages   []MessageResponse `json:"messages"`
	NextCursor string            `json:"next_cursor"` // empty when no more history
}
