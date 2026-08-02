package request

// SendMessageRequest is the body for POST /chats/:chatId/messages.
type SendMessageRequest struct {
	Content   string  `json:"content" binding:"required,max=10000"`
	ReplyToID *uint64 `json:"reply_to_id"`
}

// EditMessageRequest is the body for PUT /messages/:messageId.
type EditMessageRequest struct {
	Content string `json:"content" binding:"required,max=10000"`
}

// ReactionRequest is the body for POST/DELETE /messages/:messageId/reactions.
type ReactionRequest struct {
	Emoji string `json:"emoji" binding:"required,max=50"`
}

// ListMessagesQuery is the query for GET /chats/:chatId/messages.
type ListMessagesQuery struct {
	Cursor string `form:"cursor"` // ISO8601 created_at of the oldest already-loaded message
	Limit  int    `form:"limit"`
}
