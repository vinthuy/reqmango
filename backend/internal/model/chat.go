package model

import (
	"encoding/json"
	"time"
)

// Chat is a conversation scoped to an issue (v1).
type Chat struct {
	BaseModel

	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	IssueID     *uint64 `gorm:"index" json:"issue_id"`
	Type        string  `gorm:"size:20;default:issue" json:"type"` // issue | group | dm (reserved)
	Title       string  `gorm:"size:255" json:"title"`

	// Relationships
	Issue    *Issue    `gorm:"foreignKey:IssueID" json:"-"`
	Messages []Message `gorm:"foreignKey:ChatID" json:"-"`
}

func (Chat) TableName() string { return "chats" }

// Message is a single chat message from a user or agent.
type Message struct {
	ID         uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatID     uint64          `gorm:"not null;index" json:"chat_id"`
	SenderID   uint64          `gorm:"not null" json:"sender_id"`
	SenderType string          `gorm:"size:20;not null" json:"sender_type"` // user | agent
	Content    string          `gorm:"type:text;not null" json:"content"`
	ReplyToID  *uint64         `gorm:"index" json:"reply_to_id"`
	Mentions   json.RawMessage `gorm:"type:jsonb" json:"mentions"` // [{"type":"user|agent","id":1,"name":"..."}]
	EditedAt   *time.Time      `json:"edited_at"`
	DeletedAt  *time.Time      `gorm:"index" json:"deleted_at"`
	CreatedAt  time.Time       `json:"created_at"`

	// Relationships
	Chat       *Message           `gorm:"foreignKey:ReplyToID" json:"-"`
	Reactions  []MessageReaction  `gorm:"foreignKey:MessageID" json:"reactions"`
}

func (Message) TableName() string { return "messages" }

// MessageReaction is an emoji reaction on a message.
type MessageReaction struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID uint64    `gorm:"not null;index" json:"message_id"`
	UserID    uint64    `gorm:"not null" json:"user_id"`
	Emoji     string    `gorm:"size:50;not null" json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

func (MessageReaction) TableName() string { return "message_reactions" }
