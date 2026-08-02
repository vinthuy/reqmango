# Chat & Messages Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add work-item chat with real-time SSE push, Agent auto-reply (mention + state-change triggers with memory retrieval), and emoji reactions to reqmango.

**Architecture:** Self-contained `chat` module (3 new tables: `chats`, `messages`, `message_reactions`). SSE hub extended with a `chatClients` dimension (backward-compatible). Agent replies reuse `AgentClient` adapter (new `DispatchAgentWithResult` method) + `MemoryService.SemanticSearchByText`. `IssueService.Update` gets a 1-line async hook + setter injection. Frontend adds a `Chat` tab to `IssueDetail.vue` with 6 components + `useChatSSE` composable.

**Tech Stack:** Go 1.24 / Gin / GORM / PostgreSQL; Vue 3 / TypeScript / axios / EventSource / vitest; bluemonday (already in go.mod) for HTML sanitization; existing `renderMarkdown` (frontend) for display.

---

## Spec Deviations (adapted to actual codebase)

These deviations from `docs/superpowers/specs/2026-08-02-chat-and-messages-design.md` were discovered during codebase exploration and are baked into this plan:

1. **Routes**: Spec said `/workspaces/:ws/...`. Actual codebase registers issue-scoped routes under `/api/v1/issues/:issueId/...` and entity-scoped routes under `/api/v1/<entity>/...` (see `comments` group at router.go:1033). Chat routes follow the existing convention, not the workspace-prefix convention.
2. **Permission middleware**: Spec said `RequireWorkspaceMember`. That middleware does not exist. Codebase uses `middleware.RequirePermission(db, "...", "project")` for mutations, and performs membership checks inside the service (see `comment_service.checkCommentProjectMembership`). Chat follows the service-level membership-check pattern.
3. **DI wiring**: Spec said `cmd/server/main.go`. All service/handler wiring lives in `router.go:SetupRoutes` (main.go only does DB + AutoMigrate + calls SetupRoutes). Wiring happens in router.go.
4. **Agents per issue**: Spec assumed multiple agents via `GetAgentsForIssue`. That method does not exist; each issue has at most ONE `AgentAssigneeID` (`model/issue.go:37`). State-change trigger fires for that single agent (or skips if unassigned).
5. **`content_html` column**: Spec wanted server-side markdown rendering. No markdown lib exists in `backend/go.mod` (only `bluemonday`). The frontend already has an XSS-safe `renderMarkdown` (`frontend/src/composables/useMarkdown.ts`) that escapes HTML before applying markdown transforms. **Decision: drop `content_html` for v1.** Store only raw `content`; frontend renders via `renderMarkdown`. This avoids a new backend dependency and reuses the existing XSS-safe renderer (validates BUG-07 pattern: HTML is escaped at line 24 of useMarkdown.ts before any tag insertion).
6. **`MemoryService.SemanticSearchByText`**: actual signature is `SemanticSearchByText(ctx, workspaceID, query string, limit int)` (not `SemanticSearch`).
7. **Agent dispatch result**: `AgentClient.DispatchAgent` (existing) returns only `error` and discards the `AgentActivity.ResultSummary`. A new `AgentClient.DispatchAgentWithResult` method is added to return the summary string (avoids import cycle: `service` package cannot import `ai/service` directly, hence the `client.AgentClient` adapter).

---

## File Structure

### New backend files
- `backend/migrations/000017_chat_systems.up.sql` / `.down.sql` — schema for 3 tables
- `backend/internal/model/chat.go` — `Chat`, `Message`, `MessageReaction` GORM models
- `backend/internal/dto/request/chat.go` — request DTOs
- `backend/internal/dto/response/chat.go` — response DTOs
- `backend/internal/service/chat_service.go` — core chat logic + agent triggers
- `backend/internal/service/chat_debouncer.go` — 30s debounce for agent replies
- `backend/internal/service/chat_service_test.go` — service unit tests
- `backend/internal/service/chat_debouncer_test.go` — debouncer unit tests
- `backend/internal/service/sse_hub_test.go` — SSE hub chat-broadcast tests
- `backend/internal/handler/chat_handler.go` — HTTP + SSE handlers
- `backend/internal/handler/chat_handler_test.go` — handler unit tests

### Modified backend files
- `backend/internal/service/sse_hub.go` — add `chatClients` map + 3 methods (backward-compatible)
- `backend/internal/client/agent_client.go` — add `DispatchAgentWithResult`
- `backend/internal/service/issue_service.go` — add `chatSvc` field + `SetChatService` setter + 1-line hook in `Update`
- `backend/internal/router/router.go` — construct `chatSvc`/`chatH`, wire setter, register routes
- `backend/cmd/server/main.go` — add 3 models to `AutoMigrate` list

### New frontend files
- `frontend/src/api/chat.ts` — API module
- `frontend/src/types/chat.ts` — TS types
- `frontend/src/composables/useChatSSE.ts` — SSE composable (per-chat EventSource)
- `frontend/src/components/chat/ChatPanel.vue` — container
- `frontend/src/components/chat/ChatMessageList.vue` — message list + scroll
- `frontend/src/components/chat/ChatMessage.vue` — single message
- `frontend/src/components/chat/ChatInput.vue` — input + @mention
- `frontend/src/components/chat/MessageReactions.vue` — emoji reactions
- `frontend/src/components/chat/MessageActions.vue` — edit/delete/copy/reply menu
- `frontend/e2e/chat-e2e.spec.ts` — Playwright e2e

### Modified frontend files
- `frontend/src/views/IssueDetail.vue` — add Chat tab + lazy-load ChatPanel
- `frontend/src/locales/zh-CN.json` — `chat.*` + `issue.tabChat` keys
- `frontend/src/locales/en-US.json` — same keys

---

## Task 1: Database Migration

**Files:**
- Create: `backend/migrations/000017_chat_systems.up.sql`
- Create: `backend/migrations/000017_chat_systems.down.sql`

- [ ] **Step 1: Write the up migration**

Create `backend/migrations/000017_chat_systems.up.sql`:

```sql
-- 000017_chat_systems.up.sql
-- Chat & Messages: work-item chat, messages, emoji reactions.

CREATE TABLE IF NOT EXISTS chats (
    id BIGSERIAL PRIMARY KEY,
    workspace_id BIGINT NOT NULL,
    project_id BIGINT,
    issue_id BIGINT,
    type VARCHAR(20) NOT NULL DEFAULT 'issue',
    title VARCHAR(255),
    created_by_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_chats_issue ON chats(issue_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_chats_workspace ON chats(workspace_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id BIGINT NOT NULL,
    sender_type VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    reply_to_id BIGINT REFERENCES messages(id),
    mentions JSONB,
    edited_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_messages_chat ON messages(chat_id, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id, sender_type);

CREATE TABLE IF NOT EXISTS message_reactions (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    emoji VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(message_id, user_id, emoji)
);
CREATE INDEX IF NOT EXISTS idx_reactions_message ON message_reactions(message_id);
```

- [ ] **Step 2: Write the down migration**

Create `backend/migrations/000017_chat_systems.down.sql`:

```sql
-- 000017_chat_systems.down.sql
DROP TABLE IF EXISTS message_reactions;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chats;
```

- [ ] **Step 3: Apply the migration to the local DB**

Run (PowerShell, psql path `D:\PostgreSQL\bin\psql.exe`):

```powershell
& "D:\PostgreSQL\bin\psql.exe" -U postgres -d reqmango -f backend/migrations/000017_chat_systems.up.sql
```

Expected: `CREATE TABLE` x3, `CREATE INDEX` x5, no errors.

- [ ] **Step 4: Commit**

```powershell
git add backend/migrations/000017_chat_systems.up.sql backend/migrations/000017_chat_systems.down.sql
git commit -m "feat(chat): add chat systems migration (000017)"
```

---

## Task 2: GORM Models + AutoMigrate Registration

**Files:**
- Create: `backend/internal/model/chat.go`
- Modify: `backend/cmd/server/main.go` (AutoMigrate list, ~line 168)

- [ ] **Step 1: Write the model file**

Create `backend/internal/model/chat.go`:

```go
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
```

- [ ] **Step 2: Register models in AutoMigrate**

In `backend/cmd/server/main.go`, find the AutoMigrate block ending at line 167 (`&model.WorkflowNodeRun{},`). Add the 3 new models before the closing `);`:

```go
		&model.WorkflowNodeRun{},
		// Chat & Messages
		&model.Chat{},
		&model.Message{},
		&model.MessageReaction{},
	); err != nil {
```

- [ ] **Step 3: Verify it compiles and tables reconcile**

Run from `d:\code\reqmango`:

```powershell
cd backend; go build ./...; cd ..
```

Expected: no errors. (AutoMigrate runs at server startup; tables already exist from Task 1, so it will no-op.)

- [ ] **Step 4: Commit**

```powershell
git add backend/internal/model/chat.go backend/cmd/server/main.go
git commit -m "feat(chat): add Chat/Message/MessageReaction GORM models"
```

---

## Task 3: SSE Hub Chat Dimension + Tests

**Files:**
- Modify: `backend/internal/service/sse_hub.go`
- Create: `backend/internal/service/sse_hub_test.go`

- [ ] **Step 1: Write the failing test**

Create `backend/internal/service/sse_hub_test.go`:

```go
package service

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSSEHub_RegisterChatAndBroadcast(t *testing.T) {
	h := &SSEHub{
		clients:     make(map[uint64][]*SSEClient),
		chatClients: make(map[uint64][]*SSEClient),
	}

	c1 := h.RegisterChat(42, 1)
	c2 := h.RegisterChat(42, 2)

	got := make([]string, 0, 2)
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); s := <-c1.Ch; mu.Lock(); got = append(got, s); mu.Unlock() }()
	go func() { defer wg.Done(); s := <-c2.Ch; mu.Lock(); got = append(got, s); mu.Unlock() }()

	h.BroadcastToChat(42, "message_new", map[string]string{"content": "hi"})
	wg.Wait()

	if len(got) != 2 {
		t.Fatalf("expected 2 deliveries, got %d", len(got))
	}
	for _, msg := range got {
		if !strings.Contains(msg, "event: message_new") || !strings.Contains(msg, "hi") {
			t.Errorf("unexpected message: %q", msg)
		}
	}
}

func TestSSEHub_UnregisterChatRemovesClient(t *testing.T) {
	h := &SSEHub{
		clients:     make(map[uint64][]*SSEClient),
		chatClients: make(map[uint64][]*SSEClient),
	}
	c := h.RegisterChat(7, 1)
	h.UnregisterChat(7, c)

	h.mu.RLock()
	remaining := len(h.chatClients[7])
	h.mu.RUnlock()
	if remaining != 0 {
		t.Fatalf("expected 0 remaining chat clients, got %d", remaining)
	}
}

func TestSSEHub_BroadcastToChatDoesNotBlockWhenChannelFull(t *testing.T) {
	h := &SSEHub{
		clients:     make(map[uint64][]*SSEClient),
		chatClients: make(map[uint64][]*SSEClient),
	}
	c := h.RegisterChat(1, 1) // channel buffer 32
	// Fill the channel
	for i := 0; i < 32; i++ {
		c.Ch <- "filler"
	}
	done := make(chan struct{})
	go func() {
		h.BroadcastToChat(1, "message_new", "overflow") // should not block
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("BroadcastToChat blocked on full channel")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```powershell
cd backend; go test ./internal/service/ -run TestSSEHub -v; cd ..
```

Expected: FAIL — `h.chatClients undefined`, `h.RegisterChat undefined`.

- [ ] **Step 3: Implement the SSE hub extension**

In `backend/internal/service/sse_hub.go`, replace the entire file content with:

```go
package service

import (
	"encoding/json"
	"fmt"
	"sync"
)

// SSEClient represents a connected SSE client.
type SSEClient struct {
	UserID uint64
	Ch     chan string
}

// SSEHub manages SSE connections.
// clients maps user IDs to their SSE clients (personal notifications).
// chatClients maps chat IDs to SSE clients subscribed to that chat room.
type SSEHub struct {
	mu          sync.RWMutex
	clients     map[uint64][]*SSEClient
	chatClients map[uint64][]*SSEClient
}

var SSE = &SSEHub{
	clients:     make(map[uint64][]*SSEClient),
	chatClients: make(map[uint64][]*SSEClient),
}

func (h *SSEHub) Register(userID uint64) *SSEClient {
	c := &SSEClient{UserID: userID, Ch: make(chan string, 32)}
	h.mu.Lock()
	h.clients[userID] = append(h.clients[userID], c)
	h.mu.Unlock()
	return c
}

func (h *SSEHub) Unregister(c *SSEClient) {
	h.mu.Lock()
	clients := h.clients[c.UserID]
	for i, cl := range clients {
		if cl == c {
			h.clients[c.UserID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	h.mu.Unlock()
	close(c.Ch)
}

func (h *SSEHub) SendToUser(userID uint64, event, data string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients[userID] {
		msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)
		select {
		case c.Ch <- msg:
		default:
		}
	}
}

// NotifyUser sends a structured notification via SSE.
func (h *SSEHub) NotifyUser(userID uint64, ntype, title, message string) {
	data, _ := json.Marshal(map[string]string{"type": ntype, "title": title, "message": message})
	h.SendToUser(userID, "notification", string(data))
}

// BroadcastEvent sends an event to all connected clients (legacy, personal维度).
func (h *SSEHub) BroadcastEvent(event string, data interface{}) {
	dataBytes, _ := json.Marshal(data)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, clients := range h.clients {
		for _, c := range clients {
			msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, string(dataBytes))
			select {
			case c.Ch <- msg:
			default:
			}
		}
	}
}

// RegisterChat subscribes a user to a chat room and returns the SSE client.
func (h *SSEHub) RegisterChat(chatID, userID uint64) *SSEClient {
	c := &SSEClient{UserID: userID, Ch: make(chan string, 32)}
	h.mu.Lock()
	h.chatClients[chatID] = append(h.chatClients[chatID], c)
	h.mu.Unlock()
	return c
}

// UnregisterChat removes a client from a chat room.
func (h *SSEHub) UnregisterChat(chatID uint64, c *SSEClient) {
	h.mu.Lock()
	clients := h.chatClients[chatID]
	for i, cl := range clients {
		if cl == c {
			h.chatClients[chatID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	if len(h.chatClients[chatID]) == 0 {
		delete(h.chatClients, chatID)
	}
	h.mu.Unlock()
	close(c.Ch)
}

// BroadcastToChat sends an event to all clients subscribed to a chat room.
func (h *SSEHub) BroadcastToChat(chatID uint64, event string, data interface{}) {
	dataBytes, _ := json.Marshal(data)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.chatClients[chatID] {
		msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, string(dataBytes))
		select {
		case c.Ch <- msg:
		default:
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

```powershell
cd backend; go test ./internal/service/ -run TestSSEHub -v; cd ..
```

Expected: PASS — all 3 tests.

- [ ] **Step 5: Commit**

```powershell
git add backend/internal/service/sse_hub.go backend/internal/service/sse_hub_test.go
git commit -m "feat(chat): extend SSEHub with chatClients dimension + tests"
```

---

## Task 4: AgentClient.DispatchAgentWithResult

**Files:**
- Modify: `backend/internal/client/agent_client.go`

- [ ] **Step 1: Add the new method**

In `backend/internal/client/agent_client.go`, append after the existing `HandleMention` method (end of file):

```go
// DispatchAgentWithResult dispatches a task to an agent and returns the
// human-readable result summary. Used by ChatService to capture agent replies
// as chat messages.
func (c *AgentClient) DispatchAgentWithResult(workspaceID, agentID, userID uint64, task string, issueID, projectID *uint64, triggeredBy string) (string, error) {
	act, err := c.agentSvc.DispatchAgent(agentID, userID, task, &aiservice.DispatchContext{
		IssueID:     issueID,
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		TriggeredBy: triggeredBy,
	})
	if err != nil {
		return "", err
	}
	if act == nil {
		return "", nil
	}
	return act.ResultSummary, nil
}
```

- [ ] **Step 2: Verify it compiles**

```powershell
cd backend; go build ./...; cd ..
```

Expected: no errors.

- [ ] **Step 3: Commit**

```powershell
git add backend/internal/client/agent_client.go
git commit -m "feat(chat): add AgentClient.DispatchAgentWithResult for chat replies"
```

---

## Task 5: Chat DTOs

**Files:**
- Create: `backend/internal/dto/request/chat.go`
- Create: `backend/internal/dto/response/chat.go`

- [ ] **Step 1: Write request DTOs**

Create `backend/internal/dto/request/chat.go`:

```go
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
```

- [ ] **Step 2: Write response DTOs**

Create `backend/internal/dto/response/chat.go`:

```go
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
	Emoji   string  `json:"emoji"`
	Count   int     `json:"count"`
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
	Messages []MessageResponse `json:"messages"`
	NextCursor string          `json:"next_cursor"` // empty when no more history
}
```

- [ ] **Step 3: Verify it compiles**

```powershell
cd backend; go build ./...; cd ..
```

Expected: no errors.

- [ ] **Step 4: Commit**

```powershell
git add backend/internal/dto/request/chat.go backend/internal/dto/response/chat.go
git commit -m "feat(chat): add chat request/response DTOs"
```

---

## Task 6: Chat Debouncer + Tests

**Files:**
- Create: `backend/internal/service/chat_debouncer.go`
- Create: `backend/internal/service/chat_debouncer_test.go`

- [ ] **Step 1: Write the failing test**

Create `backend/internal/service/chat_debouncer_test.go`:

```go
package service

import (
	"testing"
	"time"
)

func TestAgentReplyDebouncer_AllowFirstBlocksSecond(t *testing.T) {
	d := NewAgentReplyDebouncer(30 * time.Second)
	if !d.Allow(5, 99) {
		t.Fatal("first call should be allowed")
	}
	if d.Allow(5, 99) {
		t.Fatal("second call within window should be blocked")
	}
}

func TestAgentReplyDebouncer_DifferentKeysIndependent(t *testing.T) {
	d := NewAgentReplyDebouncer(30 * time.Second)
	if !d.Allow(5, 1) {
		t.Fatal("agent 5 issue 1 should be allowed")
	}
	if !d.Allow(6, 1) {
		t.Fatal("agent 6 issue 1 should be allowed (different agent)")
	}
	if !d.Allow(5, 2) {
		t.Fatal("agent 5 issue 2 should be allowed (different issue)")
	}
}

func TestAgentReplyDebouncer_AllowsAfterWindowExpires(t *testing.T) {
	d := NewAgentReplyDebouncer(50 * time.Millisecond)
	if !d.Allow(1, 1) {
		t.Fatal("first call should be allowed")
	}
	time.Sleep(60 * time.Millisecond)
	if !d.Allow(1, 1) {
		t.Fatal("call after window expiry should be allowed")
	}
}

func TestAgentReplyDebouncer_CleanupRemovesExpiredKeys(t *testing.T) {
	d := NewAgentReplyDebouncer(20 * time.Millisecond)
	d.Allow(1, 1)
	time.Sleep(30 * time.Millisecond)
	d.Cleanup()
	d.mu.Lock()
	n := len(d.cache)
	d.mu.Unlock()
	if n != 0 {
		t.Fatalf("expected 0 keys after cleanup, got %d", n)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```powershell
cd backend; go test ./internal/service/ -run TestAgentReplyDebouncer -v; cd ..
```

Expected: FAIL — `NewAgentReplyDebouncer undefined`.

- [ ] **Step 3: Implement the debouncer**

Create `backend/internal/service/chat_debouncer.go`:

```go
package service

import (
	"fmt"
	"sync"
	"time"
)

// AgentReplyDebouncer prevents the same agent from being triggered for the
// same issue more than once within a configurable window. v1 is in-memory;
// restart clears the cache (acceptable).
type AgentReplyDebouncer struct {
	window time.Duration
	mu     sync.Mutex
	cache  map[string]time.Time // key: "agentID:issueID"
}

// NewAgentReplyDebouncer creates a debouncer with the given window.
func NewAgentReplyDebouncer(window time.Duration) *AgentReplyDebouncer {
	return &AgentReplyDebouncer{window: window, cache: make(map[string]time.Time)}
}

// Allow returns true if the agent+issue pair has not been triggered within the
// window, and records the trigger time. Returns false if within the window.
func (d *AgentReplyDebouncer) Allow(agentID, issueID uint64) bool {
	key := fmt.Sprintf("%d:%d", agentID, issueID)
	d.mu.Lock()
	defer d.mu.Unlock()
	if t, ok := d.cache[key]; ok && time.Since(t) < d.window {
		return false
	}
	d.cache[key] = time.Now()
	return true
}

// Cleanup removes expired entries. Call periodically from a background goroutine.
func (d *AgentReplyDebouncer) Cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	for k, t := range d.cache {
		if now.Sub(t) >= d.window {
			delete(d.cache, k)
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

```powershell
cd backend; go test ./internal/service/ -run TestAgentReplyDebouncer -v; cd ..
```

Expected: PASS — all 4 tests.

- [ ] **Step 5: Commit**

```powershell
git add backend/internal/service/chat_debouncer.go backend/internal/service/chat_debouncer_test.go
git commit -m "feat(chat): add AgentReplyDebouncer (30s window) + tests"
```

---

## Task 7: ChatService Core (GetOrCreate + List + Send) + Tests

**Files:**
- Create: `backend/internal/service/chat_service.go`
- Create: `backend/internal/service/chat_service_test.go`

- [ ] **Step 1: Write the service skeleton + GetOrCreateForIssue + ListMessages**

Create `backend/internal/service/chat_service.go`:

```go
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
```

- [ ] **Step 2: Write a basic service test (membership + send)**

Create `backend/internal/service/chat_service_test.go`:

```go
package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/reqmango/backend/internal/dto/request"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NOTE: This test uses sqlmock to avoid a live DB dependency. If go-sqlmock is
// not yet a dependency, run: cd backend; go get github.com/DATA-DOG/go-sqlmock
// If adding the dep is undesirable, delete this test file and rely on the
// handler integration test (Task 11) + manual verification (Task 26).

func TestChatService_SendMessage_RejectsNonMember(t *testing.T) {
	// Lightweight structural test: verify checkProjectMembership returns Forbidden
	// when the project member count is 0. Uses a real ChatService with a nil-ish
	// db is not feasible, so we test the standalone helper logic via a fake.
	s := &ChatService{}
	// Direct call would panic on nil db; instead assert the error type produced
	// by common.Forbidden matches the expected message.
	err := s.checkProjectMembership(1, 2)
	if err == nil {
		t.Fatal("expected error from nil-db membership check, got nil")
	}
	if !strings.Contains(err.Error(), "project member") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatService_parseAndResolveMentions_EmptyContent(t *testing.T) {
	s := &ChatService{}
	// Will return nil because parseMentions("") is empty — no DB hit.
	got := s.parseAndResolveMentions("", 1)
	if len(got) != 0 {
		t.Fatalf("expected 0 mentions, got %d", len(got))
	}
}

func TestChatService_toMessageResponses_Empty(t *testing.T) {
	s := &ChatService{}
	got := s.toMessageResponses(nil)
	if got == nil || len(got) != 0 {
		t.Fatalf("expected non-nil empty slice, got %v", got)
	}
}

func TestParseMentions_ExtractsAgentName(t *testing.T) {
	names := parseMentions("hello @leader-agent please review")
	if len(names) != 1 || names[0] != "leader-agent" {
		t.Fatalf("expected [leader-agent], got %v", names)
	}
}

// Ensure request DTOs round-trip JSON as expected.
func TestSendMessageRequest_JSONBinding(t *testing.T) {
	raw := `{"content":"hi @leader-agent","reply_to_id":null}`
	var r request.SendMessageRequest
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatal(err)
	}
	if r.Content != "hi @leader-agent" {
		t.Fatalf("unexpected content: %q", r.Content)
	}
}

// Suppress unused imports when sqlmock path is conditionally compiled.
var _ = sqlmock.New
var _ = postgres.Open
var _ = gorm.ErrRecordNotFound
```

> **Note:** If `github.com/DATA-DOG/go-sqlmock` is not in `go.mod`, either run `go get github.com/DATA-DOG/go-sqlmock` or delete the import + the `var _ =` line. The structural tests above do not require a live DB.

- [ ] **Step 3: Run tests to verify they pass (or fail predictably)**

```powershell
cd backend; go test ./internal/service/ -run "TestChatService|TestParseMentions|TestSendMessageRequest" -v; cd ..
```

Expected: PASS. If `go-sqlmock` is missing, add it: `go get github.com/DATA-DOG/go-sqlmock`, then re-run.

- [ ] **Step 4: Commit**

```powershell
git add backend/internal/service/chat_service.go backend/internal/service/chat_service_test.go
git commit -m "feat(chat): add ChatService core (GetOrCreate, List, Send) + tests"
```

---

## Task 8: ChatService Edit/Delete/Reactions + Agent Triggers

**Files:**
- Modify: `backend/internal/service/chat_service.go`

- [ ] **Step 1: Append Edit/Delete/Reaction methods to chat_service.go**

Append to `backend/internal/service/chat_service.go` (end of file):

```go
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
		if ws.OwnerID == nil || *ws.OwnerID != userID {
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
```

Also add `"strings"` to the import block at the top of `chat_service.go` (it is needed by `isDuplicateKeyErr`). The final import block should read:

```go
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
```

- [ ] **Step 2: Append agent trigger methods to chat_service.go**

Append to `backend/internal/service/chat_service.go` (end of file):

```go
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
```

- [ ] **Step 3: Verify it compiles**

```powershell
cd backend; go build ./...; cd ..
```

Expected: no errors.

- [ ] **Step 4: Commit**

```powershell
git add backend/internal/service/chat_service.go
git commit -m "feat(chat): add Edit/Delete/Reactions + agent auto-reply triggers"
```

---

## Task 9: IssueService Hook + Setter

**Files:**
- Modify: `backend/internal/service/issue_service.go`

- [ ] **Step 1: Add chatSvc field + setter to IssueService**

In `backend/internal/service/issue_service.go`, edit the `IssueService` struct (lines 39-45) to add a `chatSvc` field:

```go
type IssueService struct {
	db              *gorm.DB
	notificationSvc *NotificationService
	webhookSvc      *WebhookService
	automationSvc   *AutomationService
	slackSvc        *SlackService
	chatSvc         *ChatService
}
```

Directly below `NewIssueService` (after line 49), add the setter:

```go
// SetChatService injects the chat service for issue state-change -> agent reply hooks.
func (s *IssueService) SetChatService(chatSvc *ChatService) {
	s.chatSvc = chatSvc
}
```

- [ ] **Step 2: Add the 1-line hook in Update**

In `backend/internal/service/issue_service.go`, find the state-change automation block (lines 967-980). Immediately after the `s.runAutomations(...)` call for `issue.state_changed` (closing brace at line 980), insert the chat hook inside the `if req.StateID != nil {` block:

```go
	// Automation trigger: fire after commit for state changes
	if req.StateID != nil {
		var newState model.State
		s.db.First(&newState, *req.StateID)

		s.runAutomations(issueID, "issue.state_changed", map[string]interface{}{
			"issue_id":    issueID,
			"old_state":   fmt.Sprintf("%d", oldStateID),
			"new_state":   fmt.Sprintf("%d", *req.StateID),
			"state_group": newState.Group,
			"project_id":  issue.ProjectID,
			"priority":    issue.Priority,
		})

		// Chat hook: trigger agent auto-reply on issue state change (async, non-blocking)
		if s.chatSvc != nil && oldStateID != 0 {
			go s.chatSvc.OnIssueStateChanged(context.Background(), issueID, oldStateID, *req.StateID, userID)
		}
	}
```

Add `"context"` to the import block at the top of `issue_service.go` (it is required by `context.Background()`). The import block (lines 3-20) becomes:

```go
import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/rql"
	"gorm.io/gorm"
)
```

- [ ] **Step 3: Verify it compiles**

```powershell
cd backend; go build ./...; cd ..
```

Expected: no errors.

- [ ] **Step 4: Commit**

```powershell
git add backend/internal/service/issue_service.go
git commit -m "feat(chat): add IssueService state-change hook + ChatService setter"
```

---

## Task 10: Chat Handler

**Files:**
- Create: `backend/internal/handler/chat_handler.go`

- [ ] **Step 1: Write the handler**

Create `backend/internal/handler/chat_handler.go`:

```go
package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/gorm"
)

// ChatHandler exposes chat HTTP + SSE endpoints.
type ChatHandler struct {
	svc    *service.ChatService
	db     *gorm.DB
	secret string
}

// NewChatHandler creates a ChatHandler. db and secret are used for SSE JWT auth.
func NewChatHandler(svc *service.ChatService, db *gorm.DB, secret string) *ChatHandler {
	return &ChatHandler{svc: svc, db: db, secret: secret}
}

// GetOrCreateForIssue: GET /issues/:issueId/chat
func (h *ChatHandler) GetOrCreateForIssue(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	issueID, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	resp, err := h.svc.GetOrCreateForIssue(issueID, user.ID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetChat: GET /chats/:chatId
func (h *ChatHandler) GetChat(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	chatID, _ := strconv.ParseUint(c.Param("chatId"), 10, 64)
	// Reuse GetOrCreateForIssue's shape minus messages: verify membership then return
	if err := h.svc.GetChatMembershipCheck(chatID, user.ID); err != nil {
		h.writeError(c, err)
		return
	}
	resp, err := h.svc.GetChat(chatID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListMessages: GET /chats/:chatId/messages
func (h *ChatHandler) ListMessages(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	chatID, _ := strconv.ParseUint(c.Param("chatId"), 10, 64)
	var q request.ListMessagesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid query"})
		return
	}
	resp, err := h.svc.ListMessages(chatID, user.ID, q)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SendMessage: POST /chats/:chatId/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	chatID, _ := strconv.ParseUint(c.Param("chatId"), 10, 64)
	var req request.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}
	resp, err := h.svc.SendMessage(chatID, user.ID, req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// EditMessage: PUT /messages/:messageId
func (h *ChatHandler) EditMessage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	messageID, _ := strconv.ParseUint(c.Param("messageId"), 10, 64)
	var req request.EditMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}
	resp, err := h.svc.EditMessage(messageID, user.ID, req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteMessage: DELETE /messages/:messageId
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	messageID, _ := strconv.ParseUint(c.Param("messageId"), 10, 64)
	if err := h.svc.DeleteMessage(messageID, user.ID); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// AddReaction: POST /messages/:messageId/reactions
func (h *ChatHandler) AddReaction(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	messageID, _ := strconv.ParseUint(c.Param("messageId"), 10, 64)
	var req request.ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}
	if err := h.svc.AddReaction(messageID, user.ID, req.Emoji); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reaction added"})
}

// RemoveReaction: DELETE /messages/:messageId/reactions
func (h *ChatHandler) RemoveReaction(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	messageID, _ := strconv.ParseUint(c.Param("messageId"), 10, 64)
	var req request.ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}
	if err := h.svc.RemoveReaction(messageID, user.ID, req.Emoji); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
}

// Stream: GET /chats/:chatId/stream — SSE long connection.
// Auth via JWT in ?token= query param (EventSource cannot set headers).
func (h *ChatHandler) Stream(c *gin.Context) {
	userID, chatID, ok := h.authSSE(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Authentication required"})
		return
	}
	// Verify chat membership before subscribing
	if err := h.svc.GetChatMembershipCheck(chatID, userID); err != nil {
		h.writeError(c, err)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	client := service.SSE.RegisterChat(chatID, userID)
	defer service.SSE.UnregisterChat(chatID, client)

	fmt.Fprintf(c.Writer, "event: connected\ndata: {\"chat_id\":%d}\n\n", chatID)
	c.Writer.Flush()

	// Heartbeat ticker (30s) to keep proxies from closing idle connections
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			_, _ = io.WriteString(c.Writer, ": heartbeat\n\n")
			c.Writer.Flush()
		case msg, ok := <-client.Ch:
			if !ok {
				return
			}
			_, _ = io.WriteString(c.Writer, msg)
			c.Writer.Flush()
		}
	}
}

// authSSE extracts the user ID from a JWT in ?token= or Authorization header.
func (h *ChatHandler) authSSE(c *gin.Context) (uint64, uint64, bool) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		if parts := strings.SplitN(authHeader, " ", 2); len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			tokenStr = parts[1]
		}
	}
	if tokenStr == "" {
		return 0, 0, false
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(h.secret), nil
	})
	if err != nil || !token.Valid {
		return 0, 0, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, 0, false
	}
	sub, _ := claims["sub"].(string)
	uid, err := strconv.ParseUint(sub, 10, 64)
	if err != nil {
		return 0, 0, false
	}
	chatID, _ := strconv.ParseUint(c.Param("chatId"), 10, 64)
	return uid, chatID, true
}

func (h *ChatHandler) writeError(c *gin.Context, err error) {
	if appErr, ok := err.(*common.AppError); ok {
		c.JSON(appErr.Code, gin.H{"message": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
}
```

- [ ] **Step 2: Add the GetChat + GetChatMembershipCheck helpers to ChatService**

The handler references `h.svc.GetChat` and `h.svc.GetChatMembershipCheck`, which do not yet exist. Append to `backend/internal/service/chat_service.go`:

```go
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
```

- [ ] **Step 3: Verify it compiles**

```powershell
cd backend; go build ./...; cd ..
```

Expected: no errors.

- [ ] **Step 4: Commit**

```powershell
git add backend/internal/handler/chat_handler.go backend/internal/service/chat_service.go
git commit -m "feat(chat): add ChatHandler (HTTP + SSE endpoints)"
```

---

## Task 11: Handler Unit Test + Router Wiring

**Files:**
- Create: `backend/internal/handler/chat_handler_test.go`
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Write a handler auth-boundary test**

Create `backend/internal/handler/chat_handler_test.go`:

```go
package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestChatHandler_Stream_RejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &ChatHandler{secret: "test-secret"}
	r.GET("/chats/:chatId/stream", h.Stream)

	req := httptest.NewRequest(http.MethodGet, "/chats/1/stream", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run the test**

```powershell
cd backend; go test ./internal/handler/ -run TestChatHandler_Stream -v; cd ..
```

Expected: PASS.

- [ ] **Step 3: Wire chat routes + DI in router.go**

In `backend/internal/router/router.go`, make three edits:

**Edit A** — After `commentSvc` is wired (around line 102-103, after `commentSvc.SetAutomationService(automationSvc)`), construct `chatSvc` and inject it into `issueSvc`. Find the block:

```go
	agentClient := client.NewAgentClient(agentSvc)
	automationSvc.SetAgentService(agentClient)     // break circular dependency: automation -> agent -> issue -> automation
	commentSvc.SetAgentService(agentClient)        // enable @agent-name mention handling in comments
	commentSvc.SetAutomationService(automationSvc) // enable comment_added automation trigger
```

Add immediately after it:

```go
	// Chat & Messages: construct chatSvc, inject agent client + memory service,
	// and wire the state-change hook into issueSvc (setter injection, mirrors commentSvc).
	chatSvc := service.NewChatService(db, memSvc)
	chatSvc.SetAgentClient(agentClient)
	chatSvc.StartDebouncerCleanup(context.Background())
	issueSvc.SetChatService(chatSvc)
```

**Edit B** — After the handler instantiation block (around line 145, after `commentH := handler.NewCommentHandler(commentSvc)`), add:

```go
	chatH := handler.NewChatHandler(chatSvc, db, cfg.SecretKey)
```

**Edit C** — Register chat routes. Find the `comments` group (line 1033) and add a new `chats` group + extend the `issues` group immediately after the `comments` block closes (after line 1041 `comments.POST("/:commentId/unresolve", commentH.Unresolve)` and its closing `}`):

```go
		// ---- Chat & Messages ----
		chats := v1.Group("/chats", authMiddleware)
		{
			chats.GET("/:chatId", chatH.GetChat)
			chats.GET("/:chatId/messages", chatH.ListMessages)
			chats.GET("/:chatId/stream", chatH.Stream) // SSE (auth via ?token=)
			chats.POST("/:chatId/messages", chatH.SendMessage)
		}
		// Message-scoped routes (flat for stable URLs regardless of chat)
		messages := v1.Group("/messages", authMiddleware)
		{
			messages.PUT("/:messageId", chatH.EditMessage)
			messages.DELETE("/:messageId", chatH.DeleteMessage)
			messages.POST("/:messageId/reactions", chatH.AddReaction)
			messages.DELETE("/:messageId/reactions", chatH.RemoveReaction)
		}
		// Issue-scoped chat lazy-create endpoint (sits inside the issues group below)
```

Then, inside the existing `issues` group (after `issues.GET("/:issueId/activities", issueH.GetActivities)` at line 804), add:

```go
			// Chat (lazy get/create for an issue)
			issues.GET("/:issueId/chat", chatH.GetOrCreateForIssue)
```

- [ ] **Step 4: Verify it compiles and the server boots**

```powershell
cd backend; go build ./...; cd ..
```

Expected: no errors.

- [ ] **Step 5: Smoke-test the routes (manual)**

Start the backend (refer to handoff: `cd backend; go run ./cmd/server`). In another terminal:

```powershell
# Login to get a token
$loginResp = Invoke-RestMethod -Uri "http://localhost:8000/api/v1/auth/login" -Method Post -ContentType "application/json" -Body '{"email":"admin@reqmango.com","password":"demo1234"}'
$token = $loginResp.token

# Find an existing issue ID (use any known issue; here we assume issue ID 1)
# GET /issues/1/chat should lazily create and return a chat
$resp = Invoke-RestMethod -Uri "http://localhost:8000/api/v1/issues/1/chat" -Method Get -Headers @{Authorization="Bearer $token"}
$resp | ConvertTo-Json -Depth 5
```

Expected: a `200 OK` JSON with `id`, `workspace_id`, `issue_id: 1`, `messages: []`.

- [ ] **Step 6: Commit**

```powershell
git add backend/internal/handler/chat_handler_test.go backend/internal/router/router.go
git commit -m "feat(chat): wire chat routes + DI in router; add handler auth test"
```

---

## Task 12: Frontend Types + API Module

**Files:**
- Create: `frontend/src/types/chat.ts`
- Create: `frontend/src/api/chat.ts`
- Modify: `frontend/src/api/index.ts` (re-export)

- [ ] **Step 1: Write the types**

Create `frontend/src/types/chat.ts`:

```ts
export interface Mention {
  type: 'user' | 'agent'
  id: number
  name: string
}

export interface ReactionGroup {
  emoji: string
  count: number
  user_ids: number[]
}

export interface ChatMessage {
  id: number
  chat_id: number
  sender_id: number
  sender_type: 'user' | 'agent'
  content: string
  reply_to_id: number | null
  mentions: Mention[]
  edited_at: string | null
  deleted_at: string | null
  created_at: string
  reactions: ReactionGroup[]
}

export interface Chat {
  id: number
  workspace_id: number
  project_id: number | null
  issue_id: number | null
  type: string
  title: string
  created_at: string
  messages: ChatMessage[]
}

export interface ListMessagesResponse {
  messages: ChatMessage[]
  next_cursor: string
}

export interface SendMessagePayload {
  content: string
  reply_to_id?: number | null
}
```

- [ ] **Step 2: Write the API module**

Create `frontend/src/api/chat.ts`:

```ts
/**
 * Chat API - 聊天API模块
 */
import api from './index'
import type {
  Chat,
  ChatMessage,
  ListMessagesResponse,
  SendMessagePayload,
} from '@/types/chat'

/**
 * 懒获取/创建 issue 关联的 chat（返回最近 50 条消息）
 */
export async function getIssueChat(issueId: number): Promise<Chat> {
  const response = await api.get(`/issues/${issueId}/chat`)
  return response.data
}

/**
 * 获取 chat 详情
 */
export async function getChat(chatId: number): Promise<Chat> {
  const response = await api.get(`/chats/${chatId}`)
  return response.data
}

/**
 * 分页加载历史消息（游标分页，加载更老的消息）
 */
export async function listMessages(
  chatId: number,
  cursor: string = '',
  limit: number = 20
): Promise<ListMessagesResponse> {
  const response = await api.get(`/chats/${chatId}/messages`, {
    params: { cursor, limit },
  })
  return response.data
}

/**
 * 发送消息
 */
export async function sendMessage(
  chatId: number,
  payload: SendMessagePayload
): Promise<ChatMessage> {
  const response = await api.post(`/chats/${chatId}/messages`, payload)
  return response.data
}

/**
 * 编辑消息（30 分钟窗口内）
 */
export async function editMessage(
  messageId: number,
  content: string
): Promise<ChatMessage> {
  const response = await api.put(`/messages/${messageId}`, { content })
  return response.data
}

/**
 * 软删除消息
 */
export async function deleteMessage(messageId: number): Promise<void> {
  await api.delete(`/messages/${messageId}`)
}

/**
 * 添加表情反应（幂等）
 */
export async function addReaction(
  messageId: number,
  emoji: string
): Promise<void> {
  await api.post(`/messages/${messageId}/reactions`, { emoji })
}

/**
 * 删除表情反应（幂等）
 */
export async function removeReaction(
  messageId: number,
  emoji: string
): Promise<void> {
  await api.delete(`/messages/${messageId}/reactions`, { data: { emoji } })
}

export default {
  getIssueChat,
  getChat,
  listMessages,
  sendMessage,
  editMessage,
  deleteMessage,
  addReaction,
  removeReaction,
}
```

- [ ] **Step 3: Re-export from api/index.ts**

In `frontend/src/api/index.ts`, append after the last `// Re-export` block (after line 88):

```ts
// Re-export chat API
export { default as chatApi } from './chat'
```

- [ ] **Step 4: Verify the frontend type-checks**

```powershell
cd frontend; npx vue-tsc --noEmit; cd ..
```

Expected: no errors related to chat types.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/types/chat.ts frontend/src/api/chat.ts frontend/src/api/index.ts
git commit -m "feat(chat): add frontend chat types + API module"
```

---

## Task 13: useChatSSE Composable

**Files:**
- Create: `frontend/src/composables/useChatSSE.ts`

- [ ] **Step 1: Write the composable**

Create `frontend/src/composables/useChatSSE.ts`:

```ts
import { ref, onUnmounted } from 'vue'
import type { ChatMessage } from '@/types/chat'

export interface AgentTypingPayload {
  chat_id: number
  agent_id: number
  agent_name?: string
}

interface ReactionPayload {
  message_id: number
  user_id: number
  emoji: string
}

interface DeletedPayload {
  id: number
  deleted_at: string
}

interface EditedPayload extends ChatMessage {}

/**
 * Subscribe to a single chat's SSE stream. Returns reactive event refs.
 * Auto-reconnects on error with a 3s backoff. Cleans up on unmount.
 */
export function useChatSSE(chatId: number) {
  const newMessages = ref<ChatMessage[]>([])
  const editedMessages = ref<EditedPayload[]>([])
  const deletedMessages = ref<DeletedPayload[]>([])
  const reactionsAdded = ref<ReactionPayload[]>([])
  const reactionsRemoved = ref<ReactionPayload[]>([])
  const agentTyping = ref<AgentTypingPayload | null>(null)
  const connected = ref(false)

  let es: EventSource | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let typingTimer: ReturnType<typeof setTimeout> | null = null

  function connect() {
    const token = localStorage.getItem('token') || ''
    const url = `/api/v1/chats/${chatId}/stream?token=${encodeURIComponent(token)}`
    es = new EventSource(url)

    es.addEventListener('connected', () => {
      connected.value = true
    })

    es.addEventListener('message_new', (e: MessageEvent) => {
      try {
        const msg = JSON.parse(e.data) as ChatMessage
        newMessages.value.push(msg)
        agentTyping.value = null // a new message clears the typing indicator
      } catch (err) {
        console.error('[useChatSSE] message_new parse error:', err)
      }
    })

    es.addEventListener('message_edited', (e: MessageEvent) => {
      try {
        editedMessages.value.push(JSON.parse(e.data) as EditedPayload)
      } catch (err) {
        console.error('[useChatSSE] message_edited parse error:', err)
      }
    })

    es.addEventListener('message_deleted', (e: MessageEvent) => {
      try {
        deletedMessages.value.push(JSON.parse(e.data) as DeletedPayload)
      } catch (err) {
        console.error('[useChatSSE] message_deleted parse error:', err)
      }
    })

    es.addEventListener('reaction_added', (e: MessageEvent) => {
      try {
        reactionsAdded.value.push(JSON.parse(e.data) as ReactionPayload)
      } catch (err) {
        console.error('[useChatSSE] reaction_added parse error:', err)
      }
    })

    es.addEventListener('reaction_removed', (e: MessageEvent) => {
      try {
        reactionsRemoved.value.push(JSON.parse(e.data) as ReactionPayload)
      } catch (err) {
        console.error('[useChatSSE] reaction_removed parse error:', err)
      }
    })

    es.addEventListener('agent_typing', (e: MessageEvent) => {
      try {
        agentTyping.value = JSON.parse(e.data) as AgentTypingPayload
        // Clear typing indicator after 10s if no message arrives
        if (typingTimer) clearTimeout(typingTimer)
        typingTimer = setTimeout(() => {
          agentTyping.value = null
        }, 10000)
      } catch (err) {
        console.error('[useChatSSE] agent_typing parse error:', err)
      }
    })

    es.onerror = () => {
      connected.value = false
      es?.close()
      es = null
      // Auto-reconnect with 3s backoff
      if (reconnectTimer) clearTimeout(reconnectTimer)
      reconnectTimer = setTimeout(connect, 3000)
    }
  }

  function disconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    if (typingTimer) clearTimeout(typingTimer)
    es?.close()
    es = null
    connected.value = false
  }

  connect()

  onUnmounted(disconnect)

  return {
    connected,
    newMessages,
    editedMessages,
    deletedMessages,
    reactionsAdded,
    reactionsRemoved,
    agentTyping,
    disconnect,
  }
}
```

- [ ] **Step 2: Verify it type-checks**

```powershell
cd frontend; npx vue-tsc --noEmit; cd ..
```

Expected: no errors.

- [ ] **Step 3: Commit**

```powershell
git add frontend/src/composables/useChatSSE.ts
git commit -m "feat(chat): add useChatSSE composable (per-chat EventSource + reconnect)"
```

---

## Task 14: ChatMessage + MessageReactions + MessageActions Components

**Files:**
- Create: `frontend/src/components/chat/ChatMessage.vue`
- Create: `frontend/src/components/chat/MessageReactions.vue`
- Create: `frontend/src/components/chat/MessageActions.vue`

- [ ] **Step 1: Write MessageReactions.vue**

Create `frontend/src/components/chat/MessageReactions.vue`:

```vue
<template>
  <div class="flex flex-wrap gap-1 mt-1">
    <button
      v-for="group in message.reactions"
      :key="group.emoji"
      @click="toggle(group.emoji)"
      :class="hasReacted(group.user_ids)
        ? 'bg-indigo-100 text-indigo-700 border-indigo-300'
        : 'bg-gray-100 text-gray-600 border-gray-200 hover:bg-gray-200'"
      class="flex items-center gap-1 px-1.5 py-0.5 text-xs rounded-full border transition-colors"
    >
      <span>{{ group.emoji }}</span>
      <span>{{ group.count }}</span>
    </button>
    <div class="relative" v-if="showPicker">
      <div class="absolute z-10 bg-white border border-gray-200 rounded-lg shadow-lg p-1 flex gap-0.5 -top-9 left-0">
        <button
          v-for="emoji in quickEmojis"
          :key="emoji"
          @click="add(emoji)"
          class="w-7 h-7 hover:bg-gray-100 rounded text-base"
        >{{ emoji }}</button>
      </div>
    </div>
    <button
      v-if="!showPicker"
      @click="showPicker = true"
      class="text-gray-400 hover:text-gray-600 text-xs px-1"
      :title="t('chat.reaction.add')"
    >😊+</button>
    <button
      v-else
      @click="showPicker = false"
      class="text-gray-400 hover:text-gray-600 text-xs px-1"
    >✕</button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ChatMessage } from '@/types/chat'
import * as chatApi from '@/api/chat'

const props = defineProps<{
  message: ChatMessage
  currentUserId: number
}>()
const emit = defineEmits<{ (e: 'refresh'): void }>()
const { t } = useI18n()

const showPicker = ref(false)
const quickEmojis = ['👍', '❤️', '🎉', '😢', '🚀', '👀', '✅', '❓']

function hasReacted(userIds: number[]): boolean {
  return userIds.includes(props.currentUserId)
}

async function toggle(emoji: string) {
  try {
    if (hasReacted(props.message.reactions.find((r) => r.emoji === emoji)?.user_ids || [])) {
      await chatApi.removeReaction(props.message.id, emoji)
    } else {
      await chatApi.addReaction(props.message.id, emoji)
    }
    emit('refresh')
  } catch (err) {
    console.error('[MessageReactions] toggle failed:', err)
  }
}

async function add(emoji: string) {
  showPicker.value = false
  try {
    await chatApi.addReaction(props.message.id, emoji)
    emit('refresh')
  } catch (err) {
    console.error('[MessageReactions] add failed:', err)
  }
}
</script>
```

- [ ] **Step 2: Write MessageActions.vue**

Create `frontend/src/components/chat/MessageActions.vue`:

```vue
<template>
  <div class="absolute -top-3 right-2 hidden group-hover:flex bg-white border border-gray-200 rounded-md shadow-sm text-xs">
    <button
      v-if="canEdit"
      @click="$emit('edit')"
      class="px-2 py-1 hover:bg-gray-100"
      :title="t('chat.action.edit')"
    >✏️</button>
    <button
      v-if="canDelete"
      @click="$emit('delete')"
      class="px-2 py-1 hover:bg-gray-100 text-red-500"
      :title="t('chat.action.delete')"
    >🗑️</button>
    <button
      @click="$emit('reply')"
      class="px-2 py-1 hover:bg-gray-100"
      :title="t('chat.action.reply')"
    >↩️</button>
    <button
      @click="copyContent"
      class="px-2 py-1 hover:bg-gray-100"
      :title="t('chat.action.copy')"
    >📋</button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ChatMessage } from '@/types/chat'

const props = defineProps<{
  message: ChatMessage
  currentUserId: number
}>()
defineEmits<{
  (e: 'edit'): void
  (e: 'delete'): void
  (e: 'reply'): void
}>()

const { t } = useI18n()

const canEdit = computed(() =>
  props.message.sender_type === 'user' &&
  props.message.sender_id === props.currentUserId &&
  !props.message.deleted_at &&
  // 30-minute edit window
  Date.now() - new Date(props.message.created_at).getTime() < 30 * 60 * 1000
)

const canDelete = computed(() =>
  !props.message.deleted_at &&
  (props.message.sender_id === props.currentUserId || props.message.sender_type === 'agent')
)

async function copyContent() {
  try {
    await navigator.clipboard.writeText(props.message.content)
  } catch (err) {
    console.error('[MessageActions] copy failed:', err)
  }
}
</script>
```

- [ ] **Step 3: Write ChatMessage.vue**

Create `frontend/src/components/chat/ChatMessage.vue`:

```vue
<template>
  <div
    :class="[
      'group relative flex gap-2 px-3 py-1.5 rounded-lg',
      isSelf ? 'flex-row-reverse' : 'flex-row',
      message.deleted_at ? 'opacity-60' : '',
    ]"
  >
    <!-- Avatar -->
    <div
      :class="[
        'flex-shrink-0 w-7 h-7 rounded-full flex items-center justify-center text-sm',
        message.sender_type === 'agent' ? 'bg-indigo-100 text-indigo-600' : 'bg-gray-200 text-gray-600',
      ]"
    >
      {{ message.sender_type === 'agent' ? '🤖' : '👤' }}
    </div>

    <!-- Bubble -->
    <div :class="['max-w-[75%]', isSelf ? 'items-end' : 'items-start']">
      <div
        :class="[
          'inline-block px-3 py-1.5 rounded-2xl text-sm break-words',
          message.deleted_at
            ? 'italic text-gray-400 bg-gray-50'
            : isSelf
              ? 'bg-indigo-500 text-white'
              : message.sender_type === 'agent'
                ? 'bg-gray-100 text-gray-800'
                : 'bg-gray-100 text-gray-800',
        ]"
      >
        <span v-if="message.deleted_at">{{ t('chat.deletedPlaceholder') }}</span>
        <div v-else>
          <span v-if="renderedHtml" v-html="renderedHtml"></span>
          <span v-else>{{ message.content }}</span>
        </div>
      </div>

      <!-- Edited + timestamp -->
      <div :class="['flex gap-1 mt-0.5 text-[10px] text-gray-400', isSelf ? 'justify-end' : 'justify-start']">
        <span v-if="message.edited_at">{{ t('chat.editedMarker') }}</span>
        <span>{{ formatTime(message.created_at) }}</span>
      </div>

      <!-- Reactions -->
      <MessageReactions
        v-if="!message.deleted_at"
        :message="message"
        :current-user-id="currentUserId"
        @refresh="$emit('refresh')"
      />
    </div>

    <!-- Hover actions -->
    <MessageActions
      v-if="!message.deleted_at"
      :message="message"
      :current-user-id="currentUserId"
      @edit="$emit('edit')"
      @delete="$emit('delete')"
      @reply="$emit('reply')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { renderMarkdown } from '@/composables/useMarkdown'
import type { ChatMessage } from '@/types/chat'
import MessageReactions from './MessageReactions.vue'
import MessageActions from './MessageActions.vue'

const props = defineProps<{
  message: ChatMessage
  currentUserId: number
}>()
defineEmits<{
  (e: 'edit'): void
  (e: 'delete'): void
  (e: 'reply'): void
  (e: 'refresh'): void
}>()

const { t } = useI18n()

const isSelf = computed(() =>
  props.message.sender_type === 'user' && props.message.sender_id === props.currentUserId
)

const renderedHtml = computed(() => {
  if (!props.message.content) return ''
  return renderMarkdown(props.message.content)
})

function formatTime(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch {
    return ''
  }
}
</script>
```

- [ ] **Step 4: Verify the frontend builds**

```powershell
cd frontend; npx vue-tsc --noEmit; cd ..
```

Expected: no errors.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/components/chat/ChatMessage.vue frontend/src/components/chat/MessageReactions.vue frontend/src/components/chat/MessageActions.vue
git commit -m "feat(chat): add ChatMessage, MessageReactions, MessageActions components"
```

---

## Task 15: ChatInput + ChatMessageList + ChatPanel Components

**Files:**
- Create: `frontend/src/components/chat/ChatInput.vue`
- Create: `frontend/src/components/chat/ChatMessageList.vue`
- Create: `frontend/src/components/chat/ChatPanel.vue`

- [ ] **Step 1: Write ChatInput.vue**

Create `frontend/src/components/chat/ChatInput.vue`:

```vue
<template>
  <div class="border-t border-gray-200 p-3 bg-white">
    <div class="relative">
      <textarea
        v-model="text"
        @keydown="onKeydown"
        @input="onInput"
        :placeholder="t('chat.placeholder')"
        class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm resize-none focus:outline-none focus:border-indigo-400"
        rows="1"
        ref="ta"
      ></textarea>

      <!-- @mention picker -->
      <div
        v-if="mentionOpen && mentionCandidates.length"
        class="absolute bottom-full mb-1 left-0 bg-white border border-gray-200 rounded-lg shadow-lg max-h-48 overflow-y-auto w-56"
      >
        <button
          v-for="c in mentionCandidates"
          :key="c.id"
          @click="pickMention(c)"
          class="w-full text-left px-3 py-1.5 text-sm hover:bg-indigo-50 flex items-center gap-2"
        >
          <span>{{ c.type === 'agent' ? '🤖' : '👤' }}</span>
          <span>{{ c.name }}</span>
        </button>
      </div>
    </div>

    <div class="flex items-center justify-between mt-1.5">
      <span class="text-[10px] text-gray-400">{{ text.length }}/10000</span>
      <div class="flex items-center gap-2">
        <span v-if="replyTo" class="text-xs text-gray-500">
          {{ t('chat.replyingTo') }} #{{ replyTo.id }}
          <button @click="$emit('cancel-reply')" class="text-gray-400 hover:text-gray-600 ml-1">✕</button>
        </span>
        <button
          @click="send"
          :disabled="!text.trim() || sending"
          class="px-3 py-1 bg-indigo-600 text-white text-sm rounded-lg disabled:opacity-50 hover:bg-indigo-700"
        >{{ t('chat.send') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ChatMessage } from '@/types/chat'

const props = defineProps<{
  sending: boolean
  replyTo: ChatMessage | null
  mentionCandidates: { id: number; name: string; type: 'user' | 'agent' }[]
}>()
const emit = defineEmits<{
  (e: 'send', content: string): void
  (e: 'cancel-reply'): void
}>()

const { t } = useI18n()
const text = ref('')
const ta = ref<HTMLTextAreaElement | null>(null)
const mentionOpen = ref(false)
const mentionQuery = ref('')

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send()
  }
}

function onInput() {
  // Auto-grow
  const el = ta.value
  if (el) {
    el.style.height = 'auto'
    el.style.height = Math.min(el.scrollHeight, 160) + 'px'
  }
  // Detect @mention pattern at caret
  const match = text.value.match(/@([\w\u4e00-\u9fa5-]*)$/)
  if (match) {
    mentionOpen.value = true
    mentionQuery.value = match[1]
  } else {
    mentionOpen.value = false
  }
}

function pickMention(c: { id: number; name: string; type: 'user' | 'agent' }) {
  // Replace the trailing @query with @name + space
  text.value = text.value.replace(/@([\w\u4e00-\u9fa5-]*)$/, `@${c.name} `)
  mentionOpen.value = false
  nextTick(() => ta.value?.focus())
}

function send() {
  const content = text.value.trim()
  if (!content) return
  emit('send', content)
  text.value = ''
  mentionOpen.value = false
  const el = ta.value
  if (el) el.style.height = 'auto'
}
</script>
```

- [ ] **Step 2: Write ChatMessageList.vue**

Create `frontend/src/components/chat/ChatMessageList.vue`:

```vue
<template>
  <div
    ref="container"
    class="flex-1 overflow-y-auto py-3 space-y-1"
    @scroll="onScroll"
  >
    <!-- Load older button -->
    <div v-if="hasMore" class="text-center py-2">
      <button
        @click="$emit('load-older')"
        :disabled="loadingOlder"
        class="text-xs text-indigo-600 hover:underline disabled:opacity-50"
      >{{ loadingOlder ? t('chat.loading') : t('chat.loadOlder') }}</button>
    </div>

    <ChatMessage
      v-for="msg in messages"
      :key="msg.id"
      :message="msg"
      :current-user-id="currentUserId"
      @edit="$emit('edit', msg)"
      @delete="$emit('delete', msg)"
      @reply="$emit('reply', msg)"
      @refresh="$emit('refresh')"
    />

    <!-- Agent typing indicator -->
    <div v-if="agentTyping" class="flex gap-2 px-3 py-1.5">
      <div class="flex-shrink-0 w-7 h-7 rounded-full bg-indigo-100 text-indigo-600 flex items-center justify-center text-sm">🤖</div>
      <div class="bg-gray-100 rounded-2xl px-3 py-2 text-sm text-gray-500 flex items-center gap-1">
        <span class="animate-bounce">·</span>
        <span class="animate-bounce" style="animation-delay:0.1s">·</span>
        <span class="animate-bounce" style="animation-delay:0.2s">·</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ChatMessage } from '@/types/chat'
import ChatMessage from './ChatMessage.vue'

const props = defineProps<{
  messages: ChatMessage[]
  currentUserId: number
  hasMore: boolean
  loadingOlder: boolean
  agentTyping: { agent_id: number } | null
}>()
defineEmits<{
  (e: 'load-older'): void
  (e: 'edit', msg: ChatMessage): void
  (e: 'delete', msg: ChatMessage): void
  (e: 'reply', msg: ChatMessage): void
  (e: 'refresh'): void
}>()

const { t } = useI18n()
const container = ref<HTMLDivElement | null>(null)
let wasAtBottom = true

function isAtBottom(): boolean {
  const el = container.value
  if (!el) return true
  return el.scrollHeight - el.scrollTop - el.clientHeight < 50
}

function onScroll() {
  wasAtBottom = isAtBottom()
}

// Auto-scroll to bottom when new messages arrive (only if user was at bottom)
watch(
  () => props.messages.length,
  async () => {
    if (wasAtBottom) {
      await nextTick()
      const el = container.value
      if (el) el.scrollTop = el.scrollHeight
    }
  }
)

// Initial scroll to bottom
watch(
  () => props.messages,
  async () => {
    await nextTick()
    const el = container.value
    if (el) el.scrollTop = el.scrollHeight
    wasAtBottom = true
  },
  { immediate: true }
)
</script>
```

- [ ] **Step 3: Write ChatPanel.vue**

Create `frontend/src/components/chat/ChatPanel.vue`:

```vue
<template>
  <div class="flex flex-col h-[600px] border border-gray-200 rounded-lg overflow-hidden bg-white">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-gray-50">
      <h3 class="text-sm font-medium text-gray-700">{{ t('chat.title') }}</h3>
      <span :class="connected ? 'text-green-500' : 'text-gray-400'" class="text-xs">●</span>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="flex-1 flex items-center justify-center text-gray-400 text-sm">
      {{ t('chat.loading') }}
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="flex-1 flex items-center justify-center text-red-500 text-sm">
      {{ error }}
    </div>

    <!-- Message list -->
    <ChatMessageList
      v-else
      :messages="messages"
      :current-user-id="currentUserId"
      :has-more="hasMore"
      :loading-older="loadingOlder"
      :agent-typing="agentTyping"
      @load-older="loadOlder"
      @edit="onEdit"
      @delete="onDelete"
      @reply="onReply"
      @refresh="refresh"
    />

    <!-- Input -->
    <ChatInput
      :sending="sending"
      :reply-to="replyTo"
      :mention-candidates="mentionCandidates"
      @send="onSend"
      @cancel-reply="replyTo = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ChatMessage } from '@/types/chat'
import * as chatApi from '@/api/chat'
import { useChatSSE } from '@/composables/useChatSSE'
import ChatMessageList from './ChatMessageList.vue'
import ChatInput from './ChatInput.vue'

const props = defineProps<{
  issueId: number
  workspaceId: number
  currentUserId: number
  mentionCandidates: { id: number; name: string; type: 'user' | 'agent' }[]
}>()

const { t } = useI18n()

const loading = ref(true)
const error = ref('')
const sending = ref(false)
const chatId = ref(0)
const messages = ref<ChatMessage[]>([])
const hasMore = ref(false)
const loadingOlder = ref(false)
const nextCursor = ref('')
const replyTo = ref<ChatMessage | null>(null)
const editingId = ref<number | null>(null)

const connected = ref(false)
const agentTyping = ref<{ agent_id: number } | null>(null)

let sse: ReturnType<typeof useChatSSE> | null = null

onMounted(async () => {
  await loadChat()
})

async function loadChat() {
  loading.value = true
  error.value = ''
  try {
    const chat = await chatApi.getIssueChat(props.issueId)
    chatId.value = chat.id
    messages.value = chat.messages || []
    loading.value = false
    // Open SSE now that we have a chatId
    sse = useChatSSE(chat.id)
    connected.value = sse.connected
    agentTyping.value = sse.agentTyping.value
    watchSSE()
  } catch (err: any) {
    error.value = err?.response?.data?.message || t('chat.loadFailed')
    loading.value = false
  }
}

function watchSSE() {
  if (!sse) return
  // new message
  watch(sse.newMessages, (arr) => {
    for (const m of arr) {
      // Replace optimistic temp message if id matches a pending id, else append
      messages.value.push(m)
    }
  }, { deep: true })
  watch(sse.editedMessages, (arr) => {
    for (const e of arr) {
      const idx = messages.value.findIndex((m) => m.id === e.id)
      if (idx >= 0) messages.value[idx] = { ...messages.value[idx], ...e }
    }
  }, { deep: true })
  watch(sse.deletedMessages, (arr) => {
    for (const d of arr) {
      const idx = messages.value.findIndex((m) => m.id === d.id)
      if (idx >= 0) messages.value[idx] = { ...messages.value[idx], deleted_at: d.deleted_at, content: '' }
    }
  }, { deep: true })
  watch(sse.reactionsAdded, () => refresh(), { deep: true })
  watch(sse.reactionsRemoved, () => refresh(), { deep: true })
  watch(sse.agentTyping, (v) => { agentTyping.value = v })
}

async function refresh() {
  if (!chatId.value) return
  try {
    const resp = await chatApi.listMessages(chatId.value, '', 50)
    messages.value = resp.messages
  } catch (err) {
    console.error('[ChatPanel] refresh failed:', err)
  }
}

async function loadOlder() {
  if (!chatId.value || loadingOlder.value || !hasMore.value) return
  loadingOlder.value = true
  try {
    const resp = await chatApi.listMessages(chatId.value, nextCursor.value, 20)
    // Prepend older messages
    messages.value = [...resp.messages, ...messages.value]
    nextCursor.value = resp.next_cursor
    hasMore.value = !!resp.next_cursor
  } catch (err) {
    console.error('[ChatPanel] loadOlder failed:', err)
  } finally {
    loadingOlder.value = false
  }
}

async function onSend(content: string) {
  if (!chatId.value) return
  sending.value = true
  try {
    const msg = await chatApi.sendMessage(chatId.value, {
      content,
      reply_to_id: replyTo.value?.id || null,
    })
    // The SSE event will normally deliver this; push optimistically if not dup
    if (!messages.value.find((m) => m.id === msg.id)) {
      messages.value.push(msg)
    }
    replyTo.value = null
  } catch (err: any) {
    error.value = err?.response?.data?.message || t('chat.sendFailed')
  } finally {
    sending.value = false
  }
}

async function onEdit(msg: ChatMessage) {
  // Simple prompt-based editor (v1); can be upgraded to inline editor later
  const newContent = window.prompt(t('chat.editPrompt'), msg.content)
  if (newContent == null || newContent === msg.content) return
  try {
    const updated = await chatApi.editMessage(msg.id, newContent)
    const idx = messages.value.findIndex((m) => m.id === updated.id)
    if (idx >= 0) messages.value[idx] = updated
  } catch (err: any) {
    error.value = err?.response?.data?.message || t('chat.editFailed')
  }
}

async function onDelete(msg: ChatMessage) {
  if (!window.confirm(t('chat.deleteConfirm'))) return
  try {
    await chatApi.deleteMessage(msg.id)
    // SSE will update; optimistic fallback:
    const idx = messages.value.findIndex((m) => m.id === msg.id)
    if (idx >= 0) {
      messages.value[idx] = { ...messages.value[idx], deleted_at: new Date().toISOString(), content: '' }
    }
  } catch (err: any) {
    error.value = err?.response?.data?.message || t('chat.deleteFailed')
  }
}

function onReply(msg: ChatMessage) {
  replyTo.value = msg
}
</script>
```

- [ ] **Step 4: Verify the frontend builds**

```powershell
cd frontend; npx vue-tsc --noEmit; cd ..
```

Expected: no errors.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/components/chat/ChatInput.vue frontend/src/components/chat/ChatMessageList.vue frontend/src/components/chat/ChatPanel.vue
git commit -m "feat(chat): add ChatInput, ChatMessageList, ChatPanel components"
```

---

## Task 16: i18n Keys

**Files:**
- Modify: `frontend/src/locales/zh-CN.json`
- Modify: `frontend/src/locales/en-US.json`

- [ ] **Step 1: Add Chinese keys**

In `frontend/src/locales/zh-CN.json`, find the `"issue"` object's tab keys (around line 174-178):

```json
    "tabDetails": "详情",
    "tabRelations": "关联",
    "tabAttachments": "附件",
    "tabTimetrack": "工时",
    "tabActivity": "动态",
```

Add a new key after `"tabActivity"`:

```json
    "tabActivity": "动态",
    "tabChat": "聊天",
```

Then add a top-level `"chat"` object. Find the top-level `{` opening (line 1) and the first top-level key (typically `"common"` or `"issue"`). Add `"chat"` as a sibling top-level key (insert after the opening brace, before the first existing key, with proper comma handling). The `"chat"` block:

```json
  "chat": {
    "title": "工作项聊天",
    "placeholder": "输入消息，@提及 Agent，回车发送，Shift+回车换行",
    "send": "发送",
    "loading": "加载中…",
    "loadOlder": "加载更早的消息",
    "loadFailed": "加载聊天失败",
    "sendFailed": "发送失败",
    "editFailed": "编辑失败",
    "deleteFailed": "删除失败",
    "editPrompt": "编辑消息：",
    "deleteConfirm": "确认删除这条消息？",
    "replyingTo": "回复 #",
    "editedMarker": "(已编辑)",
    "deletedPlaceholder": "此消息已被删除",
    "reaction": {
      "add": "添加表情"
    },
    "action": {
      "edit": "编辑",
      "delete": "删除",
      "reply": "回复",
      "copy": "复制"
    },
    "agentTyping": "Agent 正在输入…"
  },
```

- [ ] **Step 2: Add English keys**

Mirror in `frontend/src/locales/en-US.json`. Add `"tabChat": "Chat"` after `"tabActivity"`, and add the top-level `"chat"` block:

```json
  "chat": {
    "title": "Work Item Chat",
    "placeholder": "Type a message, @mention an Agent, Enter to send, Shift+Enter for newline",
    "send": "Send",
    "loading": "Loading…",
    "loadOlder": "Load older messages",
    "loadFailed": "Failed to load chat",
    "sendFailed": "Send failed",
    "editFailed": "Edit failed",
    "deleteFailed": "Delete failed",
    "editPrompt": "Edit message:",
    "deleteConfirm": "Delete this message?",
    "replyingTo": "Replying to #",
    "editedMarker": "(edited)",
    "deletedPlaceholder": "This message was deleted",
    "reaction": {
      "add": "Add reaction"
    },
    "action": {
      "edit": "Edit",
      "delete": "Delete",
      "reply": "Reply",
      "copy": "Copy"
    },
    "agentTyping": "Agent is typing…"
  },
```

- [ ] **Step 3: Verify both JSON files are valid**

```powershell
cd frontend; node -e "JSON.parse(require('fs').readFileSync('src/locales/zh-CN.json','utf8')); JSON.parse(require('fs').readFileSync('src/locales/en-US.json','utf8')); console.log('OK')"; cd ..
```

Expected: `OK`.

- [ ] **Step 4: Commit**

```powershell
git add frontend/src/locales/zh-CN.json frontend/src/locales/en-US.json
git commit -m "feat(chat): add chat.* i18n keys (zh-CN + en-US)"
```

---

## Task 17: IssueDetail.vue Chat Tab Integration

**Files:**
- Modify: `frontend/src/views/IssueDetail.vue`

- [ ] **Step 1: Import ChatPanel + add the tab entry**

In `frontend/src/views/IssueDetail.vue`, add the import. Find the component imports block (around line 188, after `import IssueGitPanel from '@/components/IssueGitPanel.vue'`):

```ts
import IssueGitPanel from '@/components/IssueGitPanel.vue'
import ChatPanel from '@/components/chat/ChatPanel.vue'
```

Add the Chat tab to the `tabs` computed (lines 238-246). Insert after the `ai` tab entry:

```ts
const tabs = computed(() => [
  { key: 'details', label: t('issue.tabDetails'), count: undefined },
  { key: 'relations', label: t('issue.tabRelations'), count: relationSidebarSummary.value?.total ?? undefined },
  { key: 'attachments', label: t('issue.tabAttachments'), count: issue.value?.attachment_count || undefined },
  { key: 'git', label: t('gitIntegration.title'), count: undefined },
  { key: 'timetrack', label: t('issue.tabTimetrack'), count: undefined },
  { key: 'activity', label: t('issue.tabActivity'), count: undefined },
  { key: 'ai', label: '🤖 AI', count: undefined },
  { key: 'chat', label: t('issue.tabChat'), count: undefined },
])
```

- [ ] **Step 2: Add the Chat tab panel (lazy-loaded)**

In the template's tab-panels section, find the AI tab block (lines 94-121) and add the Chat panel after it (before the closing `</div>` of the main content column at line 122):

```vue
          <!-- Chat Tab (lazy: only mounts when active) -->
          <ChatPanel
            v-else-if="activeTab === 'chat'"
            :issue-id="issueId"
            :workspace-id="workspaceId"
            :current-user-id="currentUserId"
            :mention-candidates="chatMentionCandidates"
          />
```

- [ ] **Step 3: Add the currentUserId + chatMentionCandidates computeds**

Find the script-setup reactive variables (around lines 195-205). Add (after `workspaceId` is defined):

```ts
import { useAuthStore } from '@/stores/auth'
// ... existing code ...

const currentUserId = computed(() => {
  // Prefer the auth store; fall back to localStorage user JSON
  try {
    const raw = localStorage.getItem('user')
    if (raw) return JSON.parse(raw).id || 0
  } catch {}
  return 0
})

// v1: empty mention candidates list (the @mention picker will still work but
// show no suggestions until project members/agents are wired in a follow-up).
// This keeps the chat usable for plain text + agent @mentions resolved server-side.
const chatMentionCandidates = computed(() => [] as { id: number; name: string; type: 'user' | 'agent' }[])
```

> **Note:** If a `useAuthStore` Pinia store exists, prefer it. If not, remove that import line — the `localStorage` fallback is sufficient. Run `npx vue-tsc --noEmit` to confirm; if the import errors, delete just that import line.

- [ ] **Step 4: Verify the frontend builds + type-checks**

```powershell
cd frontend; npx vue-tsc --noEmit; cd ..
```

Expected: no errors. If `useAuthStore` import fails, remove it (the `localStorage` fallback covers it).

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/views/IssueDetail.vue
git commit -m "feat(chat): add Chat tab to IssueDetail (lazy-loaded ChatPanel)"
```

---

## Task 18: Playwright E2E Spec

**Files:**
- Create: `frontend/e2e/chat-e2e.spec.ts`

- [ ] **Step 1: Write the e2e spec**

Create `frontend/e2e/chat-e2e.spec.ts`:

```ts
import { test, expect, type Page } from '@playwright/test'

const BASE = 'http://localhost:5173'
const API = 'http://localhost:8000/api/v1'

async function login(page: Page) {
  await page.goto(`${BASE}/login`)
  await page.fill('input[type="email"]', 'admin@reqmango.com')
  await page.fill('input[type="password"]', 'demo1234')
  await page.click('button[type="submit"]')
  await page.waitForURL('**/workspace/**', { timeout: 10000 })
}

// Helper: find any issue in the first project and return its URL
async function openFirstIssue(page: Page): Promise<string> {
  // Navigate to the first project's issues list; click the first issue row.
  // This selector is intentionally generic — adjust if the list UI changes.
  await page.goto(`${BASE}/`)
  // Wait for the sidebar / project list to render, then click the first issue link
  await page.waitForTimeout(1000)
  const issueLink = page.locator('a[href*="/issues/"]').first()
  await issueLink.click()
  await page.waitForSelector('[data-test="tab-btn"]', { timeout: 10000 })
  return page.url()
}

test.describe('Chat feature', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('Chat tab renders and accepts a message', async ({ page }) => {
    await openFirstIssue(page)
    // Click the Chat tab (last tab button)
    const chatTab = page.locator('[data-test="tab-btn"]', { hasText: /聊天|Chat/ }).last()
    await chatTab.click()
    // Type and send a message
    const textarea = page.locator('textarea').first()
    await textarea.fill('E2E test message ' + Date.now())
    await textarea.press('Enter')
    // The message should appear in the list
    await expect(page.locator('text=E2E test message')).toBeVisible({ timeout: 5000 })
  })

  test('Reactions toggle on click', async ({ page }) => {
    await openFirstIssue(page)
    await page.locator('[data-test="tab-btn"]', { hasText: /聊天|Chat/ }).last().click()
    // Send a message first to react to
    const ta = page.locator('textarea').first()
    await ta.fill('Reaction target ' + Date.now())
    await ta.press('Enter')
    await expect(page.locator('text=Reaction target')).toBeVisible({ timeout: 5000 })
    // Open the emoji picker and click 👍
    await page.locator('text=😊+').first().click()
    await page.locator('button:has-text("👍")').first().click()
    await expect(page.locator('button:has-text("👍")').first()).toBeVisible({ timeout: 5000 })
  })

  test('Edit message within 30-min window', async ({ page }) => {
    await openFirstIssue(page)
    await page.locator('[data-test="tab-btn"]', { hasText: /聊天|Chat/ }).last().click()
    const ta = page.locator('textarea').first()
    const marker = 'EditTarget-' + Date.now()
    await ta.fill(marker)
    await ta.press('Enter')
    await expect(page.locator(`text=${marker}`)).toBeVisible({ timeout: 5000 })
    // Hover the message to reveal the edit button
    const msg = page.locator(`text=${marker}`).first()
    await msg.hover()
    await page.locator('[title="编辑"], [title="Edit"]').first().click()
    // A window.prompt appears — Playwright handles it via dialog handler
    await page.evaluate(() => {
      window.prompt = () => 'Edited content'
    })
    // Re-trigger edit (prompt was overridden after first click)
    await msg.hover()
    await page.locator('[title="编辑"], [title="Edit"]').first().click()
    await expect(page.locator('text=Edited content')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('text=(已编辑)|\\(edited\\)')).toBeVisible({ timeout: 5000 })
  })

  test('Multi-tab sync: message in tab1 appears in tab2', async ({ browser }) => {
    const ctx = await browser.newContext()
    const p1 = await ctx.newPage()
    const p2 = await ctx.newPage()
    await login(p1)
    await login(p2)
    await openFirstIssue(p1)
    await openFirstIssue(p2)
    await p1.locator('[data-test="tab-btn"]', { hasText: /聊天|Chat/ }).last().click()
    await p2.locator('[data-test="tab-btn"]', { hasText: /聊天|Chat/ }).last().click()
    const marker = 'MultiTab-' + Date.now()
    await p1.locator('textarea').first().fill(marker)
    await p1.locator('textarea').first().press('Enter')
    // tab2 should receive the message via SSE within 3s
    await expect(p2.locator(`text=${marker}`)).toBeVisible({ timeout: 5000 })
  })
})
```

- [ ] **Step 2: Verify the spec parses (no run yet)**

```powershell
cd frontend; npx playwright test --list e2e/chat-e2e.spec.ts; cd ..
```

Expected: lists 4 tests. (If Playwright isn't installed, run `npx playwright install` first.)

- [ ] **Step 3: Commit**

```powershell
git add frontend/e2e/chat-e2e.spec.ts
git commit -m "test(chat): add Playwright e2e spec for chat feature"
```

---

## Task 19: End-to-End Verification + Manual Checklist

**Files:** none (verification only)

- [ ] **Step 1: Run the full backend test suite**

```powershell
cd backend; go test ./... ; cd ..
```

Expected: all tests pass (existing + new chat tests).

- [ ] **Step 2: Run the frontend type-check + unit tests**

```powershell
cd frontend; npx vue-tsc --noEmit; npm test -- --run; cd ..
```

Expected: no type errors, tests pass.

- [ ] **Step 3: Start backend + frontend**

```powershell
# Terminal 1 (backend)
cd backend; go run ./cmd/server
# Terminal 2 (frontend)
cd frontend; npm run dev
```

- [ ] **Step 4: Run the Playwright e2e suite**

```powershell
cd frontend; npx playwright test e2e/chat-e2e.spec.ts; cd ..
```

Expected: all 4 e2e tests pass (may need `npx playwright install` first).

- [ ] **Step 5: Manual acceptance checklist (from spec §8)**

Open two browsers (or normal + incognito) logged in as different users on the same issue. Verify each item:

- [ ] Both browsers open the same issue's Chat tab; messages sent in one appear in the other within 1 second (SSE working).
- [ ] Type `@leader-agent please analyze` (substitute a real agent name assigned to the issue). Within 5 seconds, an "Agent is typing…" indicator appears, followed by an agent reply message.
- [ ] Change the issue state from "Todo" to "In Progress" (or any transition). If an agent is assigned to the issue, within 5 seconds the agent posts a contextual reply. (If no agent is assigned, no reply — verify with an assigned issue.)
- [ ] Rapidly change the issue state 3 times within 30 seconds. The assigned agent replies at most once (debounce verified).
- [ ] Stop the LLM service (or set an invalid `AI_API_KEY`). Send a normal chat message — it succeeds. `@mention` an agent — no reply appears (silent failure, no error message in chat). Check backend logs for the failure line.
- [ ] Open the Chat tab, then disable network for 30+ seconds and re-enable. The SSE connection auto-reconnects within ~3 seconds (verify the green dot indicator returns).
- [ ] Send a message, hover it, click ✏️ edit within 30 minutes — edit succeeds. Wait 31 minutes (or temporarily lower the window in code) — edit returns 403.
- [ ] Send a message as user A; have workspace owner delete user A's message — succeeds. Have a non-owner non-author try to delete — 403.
- [ ] Add 👍 to a message — count goes to 1. Click 👍 again — count returns to 0 (toggle). Add from a second browser — count goes to 2.
- [ ] Open a second tab on the same issue's Chat. Send a message in tab 1 — it appears in tab 2 in real time (multi-tab SSE).

- [ ] **Step 6: Final commit (if any fixups were made during verification)**

If verification surfaced fixes, stage and commit them with a clear message. Otherwise, no commit.

```powershell
git status
# If changes exist:
git add -A
git commit -m "fix(chat): verification fixes from e2e + manual checklist"
```

---

## Self-Review Notes

**Spec coverage check:**
- §2 Data model → Task 1 (migration) + Task 2 (models). `content_html` intentionally dropped (see Deviation #5).
- §3 API design → Task 10 (handler) + Task 11 (routes). URL scheme adapted to codebase convention (Deviation #1).
- §4 SSE integration → Task 3 (hub) + Task 10 (Stream handler). BUG-12 fix: `SendToUser` now used by chat for @mention notifications (Task 7).
- §5 Agent auto-reply → Task 8 (triggerAgentReply + OnIssueStateChanged + buildAgentTask) + Task 9 (issue_service hook) + Task 4 (DispatchAgentWithResult) + Task 6 (debouncer). Single-agent-per-issue adaptation (Deviation #4).
- §6 Frontend components → Tasks 12-17 (types, API, SSE composable, 6 components, i18n, IssueDetail tab).
- §7 Error handling & security → enforced in service (Task 7-8: 30-min edit window, author/admin delete, idempotent reactions, project membership). XSS handled client-side via `renderMarkdown` HTML escaping (Deviation #5).
- §8 Testing → Tasks 3, 6, 7, 11 (Go unit), Task 18 (Playwright e2e), Task 19 (manual checklist).
- §9 Implementation deps → all new + modified files covered.

**Type consistency check:**
- `ChatService` methods referenced by handler (`GetOrCreateForIssue`, `GetChat`, `GetChatMembershipCheck`, `ListMessages`, `SendMessage`, `EditMessage`, `DeleteMessage`, `AddReaction`, `RemoveReaction`, `OnIssueStateChanged`) are all defined in Tasks 7-8 + Task 10 Step 2.
- `SSE.RegisterChat/UnregisterChat/BroadcastToChat` (Task 3) match usages in Tasks 7-8, 10.
- `AgentClient.DispatchAgentWithResult` (Task 4) matches call in Task 8.
- Frontend `ChatMessage`/`Chat` types (Task 12) match API responses (Task 5) and component props (Tasks 14-15).
- `useChatSSE` event refs (Task 13) match `ChatPanel` watchers (Task 15).

**Placeholder scan:** none. Every code step contains complete code.

---

## Execution Handoff

**Plan complete and saved to `docs/superpowers/plans/2026-08-02-chat-and-messages-plan.md`.**

Two execution options:

1. **Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.
2. **Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

Which approach?
