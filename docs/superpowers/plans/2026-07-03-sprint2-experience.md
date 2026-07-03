# Sprint 2: 体验 + 安全 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** UI 一致性达标、Onboarding 可完整走通、安全加固完成、实时 Wiki 后端就绪可 demo

**Architecture:** vinthuy 专注 Wiki WebSocket 后端开发（最大单块工作），A 做 UI 组件库统一+Onboarding+暗色/键盘+编辑器，B 做安全加固+性能优化+Swagger 文档+结构化日志。B 的安全加固需在 W4 完成以便后续任务使用加固后的基础设施。

**Prerequisites:** Sprint 1 完成（make ci 通过、测试框架就绪）

**Design Spec:** `docs/superpowers/specs/2026-07-03-productization-design.md`

---

## Track 1: vinthuy — Wiki 后端 + 分诊增强

### Task T2.1: 实时协作 Wiki 后端 (8 天，最大单块工作)

**Files:**
- Create: `backend/internal/ws/hub.go`
- Create: `backend/internal/ws/client.go`
- Create: `backend/internal/ws/handler.go`
- Create: `backend/internal/model/page_version.go`
- Create: `backend/internal/service/page_collab_service.go`
- Create: `backend/internal/handler/ws_handler.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/model/page.go`

#### Phase 2.1a: WebSocket 基础设施 (W4 Days 1-2)

- [ ] **Step 1: 添加 gorilla/websocket 依赖**

```bash
cd backend
go get github.com/gorilla/websocket@latest
go mod tidy
```

- [ ] **Step 2: 创建 Hub 和 Client**

```go
// backend/internal/ws/hub.go
package ws

import (
    "encoding/json"
    "log"
    "sync"
)

// Hub manages WebSocket connections in rooms (one room per page)
type Hub struct {
    rooms map[uint]*Room // pageID -> Room
    mu    sync.RWMutex
}

type Room struct {
    ID      uint
    Clients map[*Client]bool
    mu      sync.RWMutex
}

type Client struct {
    Conn   *websocket.Conn
    UserID uint
    Name   string
    Color  string // cursor color
    Send   chan []byte
}

func NewHub() *Hub {
    return &Hub{rooms: make(map[uint]*Room)}
}

func (h *Hub) GetOrCreateRoom(pageID uint) *Room {
    h.mu.Lock()
    defer h.mu.Unlock()
    if room, ok := h.rooms[pageID]; ok {
        return room
    }
    room := &Room{ID: pageID, Clients: make(map[*Client]bool)}
    h.rooms[pageID] = room
    return room
}

func (r *Room) Broadcast(sender *Client, msg []byte) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    for client := range r.Clients {
        if client != sender {
            select {
            case client.Send <- msg:
            default:
                // slow client, drop message
            }
        }
    }
}

type WSMessage struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data"`
}

// Message types:
// "join" — user joined, broadcast user list
// "leave" — user left
// "operation" — OT operation (insert/delete)
// "cursor" — cursor position update
// "sync" — full document sync (on reconnect)
```

```go
// backend/internal/ws/client.go
package ws

import (
    "encoding/json"
    "log"
    "time"

    "github.com/gorilla/websocket"
)

const (
    writeWait      = 10 * time.Second
    pongWait       = 60 * time.Second
    pingPeriod     = (pongWait * 9) / 10
    maxMessageSize = 65536
)

func (c *Client) ReadPump(room *Room, hub *Hub, pageID uint) {
    defer func() {
        room.mu.Lock()
        delete(room.Clients, c)
        room.mu.Unlock()
        c.Conn.Close()
        // Broadcast user leave
        leaveMsg, _ := json.Marshal(WSMessage{
            Type: "leave",
            Data: json.RawMessage(`{"userId":` + fmt.Sprint(c.UserID) + `}`),
        })
        room.Broadcast(c, leaveMsg)
    }()

    c.Conn.SetReadLimit(maxMessageSize)
    c.Conn.SetReadDeadline(time.Now().Add(pongWait))
    c.Conn.SetPongHandler(func(string) error {
        c.Conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, msg, err := c.Conn.ReadMessage()
        if err != nil {
            break
        }
        room.Broadcast(c, msg)
    }
}

func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.Conn.Close()
    }()

    for {
        select {
        case msg, ok := <-c.Send:
            if !ok { return }
            c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
                return
            }
        case <-ticker.C:
            c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

- [ ] **Step 3: WebSocket upgrade handler**

```go
// backend/internal/handler/ws_handler.go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/reqmango/backend/internal/ws"
)

type WSHandler struct {
    hub      *ws.Hub
    upgrader websocket.Upgrader
}

func NewWSHandler(hub *ws.Hub) *WSHandler {
    return &WSHandler{
        hub: hub,
        upgrader: websocket.Upgrader{
            ReadBufferSize:  1024,
            WriteBufferSize: 1024,
            CheckOrigin:     func(r *http.Request) bool { return true },
        },
    }
}

func (h *WSHandler) HandlePageWS(c *gin.Context) {
    pageIDStr := c.Param("page_id")
    pageID, err := strconv.ParseUint(pageIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_id"})
        return
    }

    userID := c.GetUint("userID")
    userName := c.GetString("userName") // from auth middleware

    conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }

    client := &ws.Client{
        Conn:   conn,
        UserID: userID,
        Name:   userName,
        Color:  userColors[userID%uint(len(userColors))],
        Send:   make(chan []byte, 256),
    }

    room := h.hub.GetOrCreateRoom(uint(pageID))
    room.mu.Lock()
    room.Clients[client] = true
    room.mu.Unlock()

    // Broadcast join
    joinMsg, _ := json.Marshal(ws.WSMessage{
        Type: "join",
        Data: json.RawMessage(fmt.Sprintf(
            `{"userId":%d,"name":"%s","color":"%s"}`, userID, userName, client.Color,
        )),
    })
    room.Broadcast(client, joinMsg)

    go client.WritePump()
    go client.ReadPump(room, h.hub, uint(pageID))
}

var userColors = []string{"#f43f5e", "#8b5cf6", "#3b82f6", "#06b6d4", "#10b981", "#f59e0b"}
```

- [ ] **Step 4: Register WebSocket route**

```go
// backend/internal/router/router.go — 添加
hub := ws.NewHub()
wsHandler := handler.NewWSHandler(hub)

// WebSocket 路由
r.GET("/ws/page/:page_id", middleware.AuthRequired(), wsHandler.HandlePageWS)
```

#### Phase 2.1b: PageVersion 模型 + 版本历史 (W5 Days 1-3)

- [ ] **Step 5: PageVersion 模型**

```go
// backend/internal/model/page_version.go
package model

import "time"

type PageVersion struct {
    ID             uint      `gorm:"primaryKey" json:"id"`
    PageID         uint      `gorm:"index;not null" json:"page_id"`
    Version        int       `gorm:"not null" json:"version"`
    ContentSnapshot string   `gorm:"type:text" json:"content_snapshot"`
    Operations     string    `gorm:"type:jsonb" json:"operations"` // JSON array of OT ops
    CreatedBy      uint      `gorm:"not null" json:"created_by"`
    CreatedAt      time.Time `json:"created_at"`
}
```

- [ ] **Step 6: Page 模型扩展**

```go
// backend/internal/model/page.go — 添加字段
type Page struct {
    // ... existing fields
    Version     int  `gorm:"default:1" json:"version"`
    IsCollaborative bool `gorm:"default:false" json:"is_collaborative"`
}
```

- [ ] **Step 7: PageCollabService — 版本管理**

```go
// backend/internal/service/page_collab_service.go
package service

type PageCollabService struct {
    db *gorm.DB
}

func (s *PageCollabService) SaveVersion(pageID uint, content string, ops []OTOperation) error {
    page := &model.Page{}
    if err := s.db.First(page, pageID).Error; err != nil {
        return err
    }

    opsJSON, _ := json.Marshal(ops)
    version := &model.PageVersion{
        PageID:          pageID,
        Version:         page.Version + 1,
        ContentSnapshot: content,
        Operations:      string(opsJSON),
        CreatedBy:       0, // from context
    }

    return s.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(version).Error; err != nil {
            return err
        }
        return tx.Model(page).Updates(map[string]interface{}{
            "version": version.Version,
        }).Error
    })
}

func (s *PageCollabService) GetVersions(pageID uint) ([]model.PageVersion, error) {
    var versions []model.PageVersion
    err := s.db.Where("page_id = ?", pageID).
        Order("version DESC").Limit(50).Find(&versions).Error
    return versions, err
}

func (s *PageCollabService) GetDiff(pageID uint, v1, v2 int) (string, error) {
    // Compute unified diff between two versions
    // Use a simple line-based diff (e.g., github.com/sergi/go-diff)
    var ver1, ver2 model.PageVersion
    s.db.Where("page_id = ? AND version = ?", pageID, v1).First(&ver1)
    s.db.Where("page_id = ? AND version = ?", pageID, v2).First(&ver2)
    return computeUnifiedDiff(ver1.ContentSnapshot, ver2.ContentSnapshot), nil
}

func (s *PageCollabService) RestoreVersion(pageID uint, versionID uint) error {
    var version model.PageVersion
    if err := s.db.First(&version, versionID).Error; err != nil {
        return err
    }
    return s.db.Model(&model.Page{}).Where("id = ?", pageID).Updates(map[string]interface{}{
        "content": version.ContentSnapshot,
        "version": version.Version,
    }).Error
}
```

#### Phase 2.1c: OT 算法最小实现 (W5 Day 3 - W6 Day 2)

- [ ] **Step 8: 实现基础 OT**

```go
// backend/internal/ws/ot.go
package ws

// OTOperation represents a single operational transform operation
type OTOperation struct {
    Type     string `json:"type"`     // "insert" | "delete" | "retain"
    Position int    `json:"position"` // character offset
    Content  string `json:"content"`  // for insert: the text; for delete: length as string ("5")
    Version  int    `json:"version"`  // document version this op is based on
}

// Transform applies operation B against operation A, returning B'
// This is the core OT function: when two ops conflict, transform one against the other
func Transform(a, b *OTOperation) *OTOperation {
    if a.Type == "insert" && b.Type == "insert" {
        // Both insert: the one with higher position shifts right
        if b.Position >= a.Position {
            b.Position += len(a.Content)
        }
        return b
    }
    if a.Type == "delete" && b.Type == "insert" {
        // A deletes, B inserts: adjust B's position
        deleteLen, _ := strconv.Atoi(a.Content)
        if b.Position > a.Position {
            b.Position -= deleteLen
            if b.Position < a.Position {
                b.Position = a.Position
            }
        }
        return b
    }
    if a.Type == "insert" && b.Type == "delete" {
        // A inserts, B deletes: extend B's delete range
        deleteLen, _ := strconv.Atoi(b.Content)
        if a.Position <= b.Position {
            b.Position += len(a.Content)
        } else if a.Position < b.Position+deleteLen {
            b.Content = strconv.Itoa(deleteLen + len(a.Content))
        }
        return b
    }
    // delete vs delete, retain vs anything — simplified
    return b
}

// ApplyOperation applies an OT op to a string
func ApplyOperation(content string, op *OTOperation) string {
    switch op.Type {
    case "insert":
        return content[:op.Position] + op.Content + content[op.Position:]
    case "delete":
        deleteLen, _ := strconv.Atoi(op.Content)
        return content[:op.Position] + content[op.Position+deleteLen:]
    }
    return content
}
```

- [ ] **Step 9: WebSocket 消息处理集成 OT**

在 `client.go` 的 ReadPump 中集成 OT：

```go
// 在 ReadPump 中处理 operation 类型消息
var msg WSMessage
json.Unmarshal(msgBytes, &msg)

switch msg.Type {
case "operation":
    var op OTOperation
    json.Unmarshal(msg.Data, &op)
    // Transform against pending ops, apply, broadcast
    transformedOp := room.TransformAgainstPending(&op)
    newContent := ApplyOperation(room.CurrentContent, transformedOp)
    room.CurrentContent = newContent
    room.Broadcast(c, msgBytes)
case "cursor":
    // Just broadcast cursor position to others
    room.Broadcast(c, msgBytes)
}
```

- [ ] **Step 10: 版本定期快照**

```go
// 在 Room 中，每 50 个 operation 自动创建版本快照
func (r *Room) maybeSaveVersion() {
    r.opCount++
    if r.opCount >= 50 {
        r.collabService.SaveVersion(r.ID, r.CurrentContent, r.RecentOps)
        r.RecentOps = nil
        r.opCount = 0
    }
}
```

- [ ] **Step 11: 编写 WebSocket 测试**

```go
// backend/internal/ws/hub_test.go
func TestHub_CreateRoom(t *testing.T) {
    hub := NewHub()
    room := hub.GetOrCreateRoom(1)
    assert.NotNil(t, room)
    assert.Equal(t, uint(1), room.ID)
}

func TestRoom_Broadcast(t *testing.T) {
    // Test that messages are sent to all clients except sender
}

func TestOT_Transform_InsertInsert(t *testing.T) {
    a := &OTOperation{Type: "insert", Position: 5, Content: "hello"}
    b := &OTOperation{Type: "insert", Position: 10, Content: "world"}
    result := Transform(a, b)
    assert.Equal(t, 15, result.Position) // 10 + len("hello") = 15
}

func TestOT_ApplyOperation(t *testing.T) {
    content := "Hello world"
    op := &OTOperation{Type: "insert", Position: 6, Content: "beautiful "}
    result := ApplyOperation(content, op)
    assert.Equal(t, "Hello beautiful world", result)
}
```

- [ ] **Step 12: Run tests and commit**

```bash
cd backend && go test ./internal/ws/... -v
git add backend/internal/ws/ backend/internal/model/page_version.go backend/internal/service/page_collab_service.go backend/internal/handler/ws_handler.go
git commit -m "feat: real-time collaborative Wiki backend — WebSocket + OT + version history"
```

---

### Task T2.2: Code Review + Sprint 2 分配 (2 天)

- [ ] **Step 1: Sprint 1 产出验收**

检查项：
- `make ci` 是否通过
- Service 测试覆盖率是否 ≥30%（`go tool cover -func=coverage.out | grep total`）
- UX 审计报告是否完整（P0/P1/P2 分级）
- CI badge 是否 green

- [ ] **Step 2: Sprint 2 任务分配**

在 reqmango 中创建 Sprint 2 Cycle，为 T2.1-T2.3, A2.1-A2.5, B2.1-B2.6 创建 Issue。

---

### Task T2.3: AI 分诊自动路由 (2 天)

**Files:**
- Modify: `backend/internal/service/ai_service.go`
- Modify: `frontend/src/views/TriageView.vue`

- [ ] **Step 1: 增强分诊 — 一键执行**

```go
// backend/internal/service/ai_service.go — 新增函数
func (s *AIService) AutoTriage(issueID uint) (*TriageResult, error) {
    issue, _ := s.issueService.GetByID(issueID)

    // 调用 LLM 获取分诊建议
    prompt := buildTriagePrompt(issue)
    response := s.llm.Chat(prompt)

    result := parseTriageResponse(response)

    // 自动执行（如果置信度足够高）
    if result.Confidence >= 0.8 {
        updates := map[string]interface{}{}
        if result.SuggestedType != "" {
            updates["type"] = result.SuggestedType
        }
        if result.SuggestedPriority != "" {
            updates["priority"] = result.SuggestedPriority
        }
        if result.SuggestedAssignee != 0 {
            updates["assignee_id"] = result.SuggestedAssignee
        }
        s.issueService.Update(issueID, updates)

        // 记录分诊活动
        s.logTriageActivity(issueID, result)
    }

    return result, nil
}

type TriageResult struct {
    SuggestedType     string  `json:"suggested_type"`
    SuggestedPriority string  `json:"suggested_priority"`
    SuggestedAssignee uint    `json:"suggested_assignee"`
    SuggestedLabels   []string `json:"suggested_labels"`
    IsDuplicate       bool    `json:"is_duplicate"`
    DuplicateOf       uint    `json:"duplicate_of,omitempty"`
    Confidence        float64 `json:"confidence"`
    Reasoning         string  `json:"reasoning"`
}
```

- [ ] **Step 2: 前端增加一键执行按钮**

```vue
<!-- 在 TriageView 中 -->
<template>
  <div v-if="triageResult" class="p-4 border rounded-lg">
    <h3>AI 分诊建议 (置信度: {{ (triageResult.confidence * 100).toFixed(0) }}%)</h3>
    <ul>
      <li>类型: {{ triageResult.suggested_type }}</li>
      <li>优先级: {{ triageResult.suggested_priority }}</li>
      <li>指派: {{ triageResult.suggested_assignee }}</li>
    </ul>
    <p class="text-sm text-gray-500">{{ triageResult.reasoning }}</p>
    <button @click="applyTriage" class="btn-primary">一键执行</button>
    <button @click="undoTriage" v-if="applied" class="btn-ghost">撤销</button>
  </div>
</template>
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/ai_service.go frontend/src/views/TriageView.vue
git commit -m "feat: AI auto-triage with one-click apply and undo"
```

---

## Track 2: A — 产品体验

### Task A2.1: UI 组件库一致性 (5 天)

**Files:**
- Create: `frontend/src/components/ui/AppButton.vue`
- Create: `frontend/src/components/ui/AppInput.vue`
- Create: `frontend/src/components/ui/AppSelect.vue`
- Create: `frontend/src/components/ui/AppModal.vue`
- Create: `frontend/src/components/ui/AppCard.vue`
- Create: `frontend/src/components/ui/AppBadge.vue`
- Create: `frontend/src/components/ui/AppTable.vue`
- Create: `frontend/src/components/ui/AppDropdown.vue`
- Create: `frontend/src/components/ui/AppDatePicker.vue`
- Create: `frontend/src/components/ui/AppDrawer.vue`
- Modify: `frontend/tailwind.config.js` — Design Token 统一
- Create: `frontend/src/components/ui/Showcase.vue` — 组件展示页

- [ ] **Step 1: 统一 Design Token**

```javascript
// frontend/tailwind.config.js — 扩展 theme
module.exports = {
  theme: {
    extend: {
      colors: {
        primary: {
          50:  '#eef2ff', 100: '#e0e7ff', 200: '#c7d2fe',
          300: '#a5b4fc', 400: '#818cf8', 500: '#6366f1',
          600: '#4f46e5', 700: '#4338ca', 800: '#3730a3', 900: '#312e81',
        },
        danger:  { 500: '#ef4444', 600: '#dc2626' },
        success: { 500: '#22c55e', 600: '#16a34a' },
        warning: { 500: '#f59e0b', 600: '#d97706' },
      },
      borderRadius: {
        sm: '6px', DEFAULT: '8px', md: '10px', lg: '12px', xl: '16px',
      },
      fontSize: {
        xs:   ['12px', '16px'],
        sm:   ['13px', '20px'],
        base: ['14px', '22px'],
        lg:   ['16px', '24px'],
        xl:   ['20px', '28px'],
        '2xl': ['24px', '32px'],
      },
      boxShadow: {
        sm: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
        DEFAULT: '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
        md: '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
        lg: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
      },
    }
  }
}
```

- [ ] **Step 2: TDD — 先写 AppButton 测试**

```typescript
// frontend/src/components/ui/__tests__/AppButton.spec.ts
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import AppButton from '../AppButton.vue'

describe('AppButton', () => {
  it.each(['primary', 'secondary', 'ghost', 'danger'] as const)(
    'renders %s variant', (variant) => {
      const wrapper = mount(AppButton, { props: { variant }, slots: { default: 'Click' } })
      expect(wrapper.text()).toBe('Click')
      expect(wrapper.classes()).toContain('btn-' + variant)
    }
  )

  it.each(['sm', 'md', 'lg'] as const)('renders %s size', (size) => {
    const wrapper = mount(AppButton, { props: { size }, slots: { default: 'Click' } })
    expect(wrapper.classes()).toContain('btn-' + size)
  })

  it('shows loading spinner and disables click when loading', async () => {
    const wrapper = mount(AppButton, { props: { loading: true }, slots: { default: 'Submit' } })
    expect(wrapper.find('.spinner').exists()).toBe(true)
    expect(wrapper.attributes('disabled')).toBeDefined()
  })

  it('disables button when disabled prop is true', () => {
    const wrapper = mount(AppButton, { props: { disabled: true }, slots: { default: 'Click' } })
    expect(wrapper.attributes('disabled')).toBeDefined()
  })

  it('emits click event when clicked', async () => {
    const wrapper = mount(AppButton, { slots: { default: 'Click' } })
    await wrapper.trigger('click')
    expect(wrapper.emitted('click')).toBeTruthy()
  })
})
```

- [ ] **Step 3: 实现 AppButton**

```vue
<!-- frontend/src/components/ui/AppButton.vue -->
<template>
  <button
    :class="classes"
    :disabled="disabled || loading"
    @click="$emit('click', $event)"
  >
    <svg v-if="loading" class="spinner animate-spin h-4 w-4 mr-2" viewBox="0 0 24 24">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
    <slot />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
  disabled?: boolean
  block?: boolean
}>(), {
  variant: 'primary',
  size: 'md',
})

defineEmits<{ click: [e: MouseEvent] }>()

const variantClasses = {
  primary: 'bg-primary-600 text-white hover:bg-primary-700 active:bg-primary-800',
  secondary: 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 hover:bg-gray-200 dark:hover:bg-gray-700',
  ghost: 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800',
  danger: 'bg-danger-500 text-white hover:bg-danger-600',
}

const sizeClasses = {
  sm: 'px-2.5 py-1.5 text-xs rounded-sm',
  md: 'px-3.5 py-2 text-sm rounded',
  lg: 'px-5 py-2.5 text-base rounded-md',
}

const classes = computed(() => [
  'btn-' + props.variant, 'btn-' + props.size,
  'inline-flex items-center justify-center font-medium transition-colors duration-150',
  'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-1',
  'disabled:opacity-50 disabled:cursor-not-allowed',
  variantClasses[props.variant],
  sizeClasses[props.size],
  props.block ? 'w-full' : '',
])
</script>
```

- [ ] **Step 4-10: 同样 TDD pattern 实现其余 9 个组件**

每个组件遵循：test → fail → implement → pass → commit

- AppInput: variant (default/error) + size + placeholder + v-model
- AppSelect: options + placeholder + searchable + v-model
- AppModal: title + content slot + footer slot + close + size (sm/md/lg/xl)
- AppCard: header slot + default slot + hoverable + padding variants
- AppBadge: variant (default/success/warning/danger) + size + dot mode
- AppTable: columns + data + sortable + selectable + pagination
- AppDropdown: trigger slot + items + placement + close on click-outside
- AppDrawer: open + placement (left/right) + title + close
- AppDatePicker: v-model + min/max + format + clearable

- [ ] **Step 11: 创建组件展示页**

```vue
<!-- frontend/src/components/ui/Showcase.vue -->
<!-- 开发环境可见，展示所有 UI 组件的 variant/size/state 矩阵 -->
```

- [ ] **Step 12: Commit**

```bash
git add frontend/src/components/ui/ frontend/tailwind.config.js
git commit -m "feat: unified UI component library (10 standard components with Design Tokens)"
```

---

### Task A2.2: Onboarding 流程 (4 天)

**Files:**
- Create: `frontend/src/components/onboarding/OnboardingWizard.vue`
- Create: `frontend/src/components/onboarding/StepWorkspace.vue`
- Create: `frontend/src/components/onboarding/StepProject.vue`
- Create: `frontend/src/components/onboarding/StepFirstIssue.vue`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/views/DashboardView.vue`

- [ ] **Step 1: 创建 OnboardingWizard**

```vue
<!-- frontend/src/components/onboarding/OnboardingWizard.vue -->
<template>
  <div class="max-w-2xl mx-auto py-12 px-4">
    <!-- Progress bar -->
    <div class="flex items-center gap-2 mb-10">
      <div v-for="i in 3" :key="i"
        :class="['h-1.5 flex-1 rounded-full transition-colors', i <= currentStep ? 'bg-primary-600' : 'bg-gray-200 dark:bg-gray-700']"
      />
    </div>

    <!-- Step content -->
    <KeepAlive>
      <StepWorkspace  v-if="currentStep === 1" @next="handleWorkspace" />
      <StepProject    v-if="currentStep === 2" :workspace="state.workspace" @next="handleProject" @back="currentStep--" />
      <StepFirstIssue v-if="currentStep === 3" :project="state.project" @done="handleDone" @back="currentStep--" />
    </KeepAlive>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const currentStep = ref(1)
const state = reactive({ workspace: null, project: null })

function handleWorkspace(ws: any) { state.workspace = ws; currentStep.value++ }
function handleProject(prj: any) { state.project = prj; currentStep.value++ }
function handleDone() {
  localStorage.setItem('onboarding_completed', 'true')
  router.push('/')
}
</script>
```

- [ ] **Step 2-4: 实现三个步骤组件**

StepWorkspace: name + slug 输入 → 调用 createWorkspace API
StepProject: name + identifier + 选择模板 → 调用 createProject API
StepFirstIssue: title + description + type → 调用 createIssue API

- [ ] **Step 5: 路由守卫 + 空状态引导**

```typescript
// frontend/src/router/index.ts
router.beforeEach((to, from) => {
  const onboarded = localStorage.getItem('onboarding_completed')
  if (!onboarded && to.path !== '/onboarding' && authStore.isLoggedIn) {
    return '/onboarding'
  }
})
```

```vue
<!-- 空状态引导示例 — 工作区空态 -->
<EmptyState
  title="创建你的第一个工作区"
  description="工作区是你团队协作的空间，可以包含多个项目"
  icon="building"
>
  <template #action>
    <AppButton @click="showCreateWorkspace = true">创建工作区</AppButton>
  </template>
</EmptyState>
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/onboarding/ frontend/src/router/index.ts frontend/src/views/DashboardView.vue
git commit -m "feat: onboarding wizard (3-step) + empty state guides"
```

---

### Task A2.3: 暗色模式 + 键盘增强 (3 天)

**Files:**
- Modify: `frontend/src/composables/useDarkMode.ts` (检查是否已存在)
- Modify: `frontend/src/components/CmdKPalette.vue`
- Modify: `frontend/src/components/ShortcutPanel.vue`

- [ ] **Step 1: 暗色模式全站审查**

逐一检查以下场景的暗色表现：
- Charts (ECharts/Chart.js) — 需要动态切换主题色
- Tiptap editor — 暗色背景 + 亮色文字
- Notification dropdown — 暗色面板
- Modal/Drawer — 暗色遮罩
- 代码块 — 暗色语法高亮
- DatePicker — 暗色日历面板

修复每个场景的 CSS 变量或 Tailwind dark: 类。

- [ ] **Step 2: Cmd+K 命令面板扩展**

```typescript
// 在 CmdKPalette 中新增命令
const commands = [
  // 导航
  { id: 'nav-issues', label: '查看 Issues', shortcut: 'G I', action: () => router.push('/issues') },
  { id: 'nav-board', label: '查看看板', shortcut: 'G B', action: () => router.push('/board') },
  { id: 'nav-cycles', label: '查看周期', shortcut: 'G C', action: () => router.push('/cycles') },
  { id: 'nav-pages', label: '查看页面', shortcut: 'G P', action: () => router.push('/pages') },

  // 创建
  { id: 'create-issue', label: '创建 Issue', shortcut: 'C I', action: openCreateIssue },
  { id: 'create-page', label: '创建页面', shortcut: 'C P', action: openCreatePage },
  { id: 'create-cycle', label: '创建周期', shortcut: 'C C', action: openCreateCycle },
  { id: 'create-release', label: '创建发布', shortcut: 'C R', action: openCreateRelease },

  // AI
  { id: 'ai-chat', label: 'AI 聊天', shortcut: 'C J', action: openAIChat },
  { id: 'ai-search', label: 'AI 搜索', shortcut: 'C F', action: focusAISearch },

  // 视图切换
  { id: 'view-list', label: '列表视图', shortcut: 'V L', action: () => switchView('list') },
  { id: 'view-board', label: '看板视图', shortcut: 'V B', action: () => switchView('board') },
  { id: 'view-tree', label: '树形视图', shortcut: 'V T', action: () => switchView('tree') },

  // 主题
  { id: 'toggle-dark', label: '切换暗色模式', shortcut: 'T D', action: toggleDarkMode },
  { id: 'toggle-lang', label: '切换语言', shortcut: 'T L', action: toggleLanguage },
]
```

- [ ] **Step 3: `?` 快捷键面板完善**

```vue
<!-- ShortcutPanel — 分类展示快捷键 -->
<div class="grid grid-cols-2 gap-6">
  <div>
    <h4>导航</h4>
    <kbd>G</kbd> + <kbd>I</kbd> → Issues<br>
    <kbd>G</kbd> + <kbd>B</kbd> → Board<br>
    ...
  </div>
  <div>
    <h4>创建</h4>
    <kbd>C</kbd> + <kbd>I</kbd> → New Issue<br>
    ...
  </div>
</div>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/composables/useDarkMode.ts frontend/src/components/CmdKPalette.vue frontend/src/components/ShortcutPanel.vue
git commit -m "feat: dark mode polish + expanded Cmd+K (15 commands) + shortcut panel"
```

---

### Task A2.4: 富文本编辑器体验优化 (3 天)

**Files:**
- Modify: `frontend/src/components/editor/TiptapEditor.vue`
- Create: `frontend/src/components/editor/extensions/SlashCommands.ts`

- [ ] **Step 1: 添加 Slash Commands 扩展**

```typescript
// frontend/src/components/editor/extensions/SlashCommands.ts
import { Extension } from '@tiptap/core'
import Suggestion from '@tiptap/suggestion'

export const SlashCommands = Extension.create({
  name: 'slashCommands',
  addOptions() {
    return {
      suggestion: {
        char: '/',
        command: ({ editor, range, props }) => {
          props.command({ editor, range })
        },
      },
    }
  },
  addProseMirrorPlugins() {
    return [
      Suggestion({
        editor: this.editor,
        ...this.options.suggestion,
      }),
    ]
  },
})

// Slash command items
export const slashCommands = [
  {
    title: 'Heading 1',
    command: ({ editor, range }) => {
      editor.chain().focus().deleteRange(range).toggleHeading({ level: 1 }).run()
    },
  },
  {
    title: 'Heading 2',
    command: ({ editor, range }) => {
      editor.chain().focus().deleteRange(range).toggleHeading({ level: 2 }).run()
    },
  },
  {
    title: 'Image',
    command: ({ editor, range }) => {
      const url = window.prompt('Image URL:')
      if (url) editor.chain().focus().deleteRange(range).setImage({ src: url }).run()
    },
  },
  {
    title: 'Table',
    command: ({ editor, range }) => {
      editor.chain().focus().deleteRange(range).insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()
    },
  },
  {
    title: 'Code Block',
    command: ({ editor, range }) => {
      editor.chain().focus().deleteRange(range).toggleCodeBlock().run()
    },
  },
  { title: 'Quote', command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleBlockquote().run() },
  { title: 'Divider', command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setHorizontalRule().run() },
  { title: 'Bullet List', command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleBulletList().run() },
  { title: 'Numbered List', command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleOrderedList().run() },
]
```

- [ ] **Step 2: Markdown 粘贴自动转换**

```typescript
// 在 TiptapEditor 中配置 paste 处理
editor.on('paste', (event) => {
  const text = event.clipboardData?.getData('text/plain')
  if (text && looksLikeMarkdown(text)) {
    event.preventDefault()
    // 依赖 Tiptap 的 Markdown 扩展自动转换
    // 或手动: editor.commands.setContent(mdToHTML(text))
  }
})

function looksLikeMarkdown(text: string): boolean {
  return /^#{1,6}\s|^\*\s|^-\s|^```|^\[.+\]\(.+\)/.test(text)
}
```

- [ ] **Step 3: 图片拖拽/粘贴上传**

```typescript
// 在 editor 配置中添加 image upload handler
editor.on('drop', async (event) => {
  const files = event.dataTransfer?.files
  if (files?.length) {
    event.preventDefault()
    for (const file of files) {
      if (file.type.startsWith('image/')) {
        const url = await uploadImage(file) // 调用附件上传 API
        editor.chain().focus().setImage({ src: url }).run()
      }
    }
  }
})
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/editor/
git commit -m "feat: Tiptap slash commands, markdown paste, drag-drop image upload"
```

---

### Task A2.5: 通知中心体验 (1 天)

**Files:**
- Modify: `frontend/src/components/NotificationCenter.vue`

- [ ] **Step 1: 通知分组 + 已读/未读 + 设置**

```vue
<!-- NotificationCenter 关键改进 -->
<template>
  <div>
    <div class="flex justify-between items-center px-4 py-2 border-b">
      <h3>通知</h3>
      <button @click="markAllRead" class="text-xs text-primary-600">全部已读</button>
    </div>

    <!-- 分组标签 -->
    <div class="flex gap-2 px-4 py-2">
      <button v-for="group in groups" :key="group.key"
        :class="['px-2 py-1 text-xs rounded-full', activeGroup === group.key ? 'bg-primary-100 text-primary-700' : 'text-gray-500']"
        @click="activeGroup = group.key"
      >{{ group.label }} ({{ group.count }})</button>
    </div>

    <!-- 通知列表 -->
    <div class="divide-y max-h-96 overflow-y-auto">
      <div v-for="notif in filteredNotifications" :key="notif.id"
        :class="['px-4 py-3 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer', !notif.read && 'bg-primary-50 dark:bg-primary-900/20']"
        @click="goToTarget(notif)"
      >
        <div class="flex items-start gap-3">
          <component :is="iconMap[notif.type]" class="w-5 h-5 mt-0.5" />
          <div>
            <p class="text-sm">{{ notif.message }}</p>
            <span class="text-xs text-gray-400">{{ timeAgo(notif.created_at) }}</span>
          </div>
          <div v-if="!notif.read" class="w-2 h-2 rounded-full bg-primary-500 mt-2 ml-auto" />
        </div>
      </div>
    </div>
  </div>
</template>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/NotificationCenter.vue
git commit -m "feat: notification grouping, read/unread, quick actions"
```

---

## Track 3: B — 工程质量

### Task B2.1: 安全加固 (4 天)

**Files:**
- Review: `backend/internal/service/*.go` (SQL 注入审计)
- Review: `frontend/src/**/*.vue` (v-html 审计)
- Modify: `backend/internal/middleware/cors.go`, `csrf.go` (或创建)
- Modify: `backend/internal/service/auth_service.go` (JWT 安全)
- Create: `docs/dev/security-audit-2026-07.md`

- [ ] **Step 1: SQL 注入扫描**

```bash
cd backend
# 查找所有 Raw SQL / 动态拼接
grep -rn "Raw(" internal/ --include="*.go"
grep -rn "Exec(" internal/ --include="*.go"
grep -rn "fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT\|fmt.Sprintf.*UPDATE\|fmt.Sprintf.*DELETE" internal/ --include="*.go"
grep -rn "Order(" internal/ --include="*.go" | grep -v "\.Order("
```

对每个匹配项审查是否使用参数化查询。修复所有字符串拼接：

```go
// ❌ Before
db.Raw("SELECT * FROM issues WHERE title LIKE '%" + keyword + "%'")

// ✅ After
db.Raw("SELECT * FROM issues WHERE title LIKE ?", "%"+keyword+"%")

// ❌ Before
db.Order(fmt.Sprintf("%s %s", sortField, sortDir))

// ✅ After — 白名单校验
var allowedFields = map[string]bool{"id": true, "title": true, "created_at": true, "priority": true}
var allowedDirs = map[string]bool{"asc": true, "desc": true}
if !allowedFields[sortField] || !allowedDirs[sortDir] {
    return errors.New("invalid sort parameter")
}
db.Order(sortField + " " + sortDir)
```

- [ ] **Step 2: XSS 防护**

```bash
cd frontend
# 查找 v-html 使用
grep -rn "v-html" src/ --include="*.vue"
```

对每个 v-html 使用确认已做 sanitize（使用 DOMPurify）：

```typescript
import DOMPurify from 'dompurify'

// ❌ Before
<div v-html="userInput" />

// ✅ After
<div v-html="DOMPurify.sanitize(userInput)" />
```

添加 CSP Header 中间件：

```go
// backend/internal/middleware/security.go
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Next()
    }
}
```

- [ ] **Step 3: CSRF 保护**

```go
// backend/internal/middleware/csrf.go
func CSRFToken() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
            c.Next()
            return
        }
        // For state-changing requests, verify CSRF token
        token := c.GetHeader("X-CSRF-Token")
        cookie, _ := c.Cookie("csrf_token")
        if token == "" || token != cookie {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token invalid"})
            return
        }
        c.Next()
    }
}
```

- [ ] **Step 4: JWT 安全配置审计**

```go
// backend/internal/service/auth_service.go
// 检查以下配置：
// 1. Access token 过期时间 ≤ 24 小时 ✅
// 2. Refresh token 实现且支持轮换 ✅
// 3. 签发者 (iss) 验证 ✅
// 4. 签名算法检查（拒绝 none 算法）✅

func (s *AuthService) GenerateToken(user *model.User) (string, error) {
    claims := jwt.MapClaims{
        "sub":   user.ID,
        "email": user.Email,
        "iss":   "reqmango",
        "iat":   time.Now().Unix(),
        "exp":   time.Now().Add(24 * time.Hour).Unix(), // 24h max
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.jwtSecret))
}
```

- [ ] **Step 5: API Key 存储加密**

```go
// 检查 model.ApiKey 的 secret 是否加密存储
// 若明文，使用 AES-256-GCM 加密
func (s *AuthService) StoreAPIKey(key *model.APIKey) error {
    encrypted, err := encrypt(key.Secret, s.encryptionKey)
    if err != nil {
        return err
    }
    key.Secret = base64.StdEncoding.EncodeToString(encrypted)
    return s.db.Create(key).Error
}
```

- [ ] **Step 6: 密码策略增强**

```go
// backend/internal/service/auth_service.go
func validatePassword(password string) error {
    if len(password) < 8 { return errors.New("password must be at least 8 characters") }
    if !regexp.MustCompile(`[A-Z]`).MatchString(password) { return errors.New("must contain uppercase") }
    if !regexp.MustCompile(`[a-z]`).MatchString(password) { return errors.New("must contain lowercase") }
    if !regexp.MustCompile(`[0-9]`).MatchString(password) { return errors.New("must contain digit") }
    return nil
}

// bcrypt cost 设为 12
func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(bytes), err
}
```

- [ ] **Step 7: 输出安全审计报告**

```markdown
# 安全审计报告 2026-07

## SQL 注入
- 发现 X 处动态 SQL，已修复 Y 处
- 风险：0 处高危剩余

## XSS
- 发现 X 处 v-html 使用
- 已添加 DOMPurify sanitize
- CSP Header 已配置

## CSRF
- 已添加 CSRF token 中间件

## JWT
- 过期时间：24h ✅
- Refresh token 轮换：✅
- 签名算法：HS256 ✅

## API Key
- 存储加密：AES-256-GCM ✅
```

- [ ] **Step 8: Commit**

```bash
git add backend/internal/middleware/security.go backend/internal/middleware/csrf.go backend/internal/service/auth_service.go docs/dev/security-audit-2026-07.md
git commit -m "security: SQL injection fix, XSS sanitize, CSRF, JWT audit, API key encryption, password policy"
```

---

### Task B2.2: 性能优化 (3 天)

**Files:**
- Modify: `backend/internal/service/*.go` (N+1 修复)
- Modify: `frontend/vite.config.ts` (code splitting)

- [ ] **Step 1: 数据库索引优化**

```bash
# 在 PostgreSQL 中查找慢查询
docker compose exec db psql -U reqmango -d reqmango -c "
SELECT query, calls, mean_exec_time, total_exec_time
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 20;
"
```

为高频查询列添加索引：

```sql
-- 预期的索引（根据常见查询模式）
CREATE INDEX IF NOT EXISTS idx_issues_project_id ON issues(project_id);
CREATE INDEX IF NOT EXISTS idx_issues_assignee_id ON issues(assignee_id);
CREATE INDEX IF NOT EXISTS idx_issues_state ON issues(state);
CREATE INDEX IF NOT EXISTS idx_issues_priority ON issues(priority);
CREATE INDEX IF NOT EXISTS idx_issues_created_at ON issues(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_issues_project_state ON issues(project_id, state);
CREATE INDEX IF NOT EXISTS idx_comments_issue_id ON comments(issue_id);
CREATE INDEX IF NOT EXISTS idx_activities_issue_id ON issue_activities(issue_id);
CREATE INDEX IF NOT EXISTS idx_pages_project_id ON pages(project_id);
```

- [ ] **Step 2: N+1 查询消除**

审查 service 层，使用 GORM Preload 替代循环查询：

```go
// ❌ Before (N+1)
issues, _ := svc.List(projectID)
for i, issue := range issues {
    var assignee model.User
    svc.db.First(&assignee, issue.AssigneeID)
    issues[i].Assignee = &assignee
}

// ✅ After
issues, _ := svc.db.Preload("Assignee").Preload("Labels").Where("project_id = ?", projectID).Find(&issues)
```

- [ ] **Step 3: 前端打包优化**

```typescript
// frontend/vite.config.ts
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor': ['vue', 'pinia', 'vue-router'],
          'editor': ['@tiptap/vue-3', '@tiptap/starter-kit', '@tiptap/extension-image'],
          'charts': ['chart.js', 'echarts'],
          'ui': ['@headlessui/vue', '@heroicons/vue'],
        },
      },
    },
    chunkSizeWarningLimit: 500,
  },
})
```

- [ ] **Step 4: Commit**

```bash
git add backend/ frontend/vite.config.ts
git commit -m "perf: DB indexes, N+1 fixes, frontend code splitting"
```

---

### Task B2.3: API 文档 (Swagger) (3 天)

**Files:**
- Modify: `backend/internal/handler/*.go` (32 files — 添加 swaggo 注释)
- Modify: `backend/cmd/server/main.go` (挂载 Swagger UI)
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: 安装 swaggo + 添加 main.go 注释**

```go
// backend/cmd/server/main.go
// @title           reqmango API
// @version         1.0
// @description     项目管理平台 REST API
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main
```

- [ ] **Step 2: 为核心 handler 添加注释（优先 10 个最常用的）**

```go
// backend/internal/handler/issue_handler.go

// @Summary      获取 Issue 列表
// @Description  按项目分页获取 Issue，支持 RQL 筛选
// @Tags         Issues
// @Accept       json
// @Produce      json
// @Param        project_id  path      int     true  "项目 ID"
// @Param        page        query     int     false "页码" default(1)
// @Param        limit       query     int     false "每页条数" default(20)
// @Param        rql         query     string  false "RQL 查询语句"
// @Success      200  {object}  dto.PaginatedResponse{data=[]model.Issue}
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      403  {object}  dto.ErrorResponse
// @Security     BearerAuth
// @Router       /projects/{project_id}/issues [get]
func (h *IssueHandler) List(c *gin.Context) { /* ... */ }

// @Summary      创建 Issue
// @Description  在指定项目中创建新的工作项
// @Tags         Issues
// @Accept       json
// @Produce      json
// @Param        project_id  path      int                   true  "项目 ID"
// @Param        body        body      request.CreateIssueRequest true "Issue 信息"
// @Success      201  {object}  model.Issue
// @Failure      400  {object}  dto.ErrorResponse
// @Security     BearerAuth
// @Router       /projects/{project_id}/issues [post]
func (h *IssueHandler) Create(c *gin.Context) { /* ... */ }
```

优先标注的 handler：
1. `issue_handler.go` — List, Create, GetByID, Update, Delete
2. `project_handler.go` — List, Create, GetByID
3. `auth_handler.go` — Register, Login, RefreshToken
4. `workspace_handler.go` — List, Create
5. `cycle_handler.go` — List, Create, GetBurndown
6. `page_handler.go` — List, Create, GetByID, Update
7. `notification_handler.go` — List, MarkRead
8. `ai_handler.go` — Chat, Search
9. `workflow_handler.go` — GetTransitions
10. `webhook_handler.go` — List, Create

- [ ] **Step 3: 挂载 Swagger UI**

```go
// backend/internal/router/router.go
import swaggerFiles "github.com/swaggo/files"
import ginSwagger "github.com/swaggo/gin-swagger"
import _ "github.com/reqmango/backend/docs" // swag init 生成的 docs

// 在路由中添加
r.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

- [ ] **Step 4: 生成 docs + CI 检查**

```bash
cd backend
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --output docs
```

```yaml
# .github/workflows/ci.yml — 添加 swagger check step
- name: Check Swagger docs
  run: |
    cd backend
    swag init -g cmd/server/main.go --parseDependency --output docs
    if ! git diff --quiet docs/; then
      echo "Swagger docs are outdated. Run 'swag init' and commit."
      exit 1
    fi
```

- [ ] **Step 5: Commit**

```bash
git add backend/docs/ backend/internal/handler/ backend/cmd/server/main.go backend/internal/router/router.go
git commit -m "docs: add Swagger/OpenAPI annotations for 10 core handlers (50+ endpoints)"
```

---

### Task B2.4: 结构化日志 (2 天)

**Files:**
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/internal/middleware/logger.go`
- Create: `backend/internal/common/logger.go`
- Modify: `backend/internal/service/*.go` (逐步替换 log.Printf)

- [ ] **Step 1: 创建统一日志包**

```go
// backend/internal/common/logger.go
package common

import (
    "os"
    "time"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func InitLogger(level string) {
    lvl, err := zerolog.ParseLevel(level)
    if err != nil {
        lvl = zerolog.InfoLevel
    }

    output := zerolog.ConsoleWriter{
        Out:        os.Stdout,
        TimeFormat: time.RFC3339,
    }

    log.Logger = zerolog.New(output).
        Level(lvl).
        With().
        Timestamp().
        Caller().
        Logger()
}
```

- [ ] **Step 2: 请求日志中间件替换**

```go
// backend/internal/middleware/logger.go
func StructuredLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        traceID := c.GetString("traceID")
        if traceID == "" {
            traceID = uuid.New().String()
            c.Set("traceID", traceID)
        }

        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        log.Info().
            Str("trace_id", traceID).
            Str("method", c.Request.Method).
            Str("path", path).
            Str("query", query).
            Int("status", status).
            Dur("latency", latency).
            Str("client_ip", c.ClientIP()).
            Int("body_size", c.Writer.Size()).
            Msg("request")
    }
}

// TraceID 中间件
func TraceID() gin.HandlerFunc {
    return func(c *gin.Context) {
        traceID := c.GetHeader("X-Trace-ID")
        if traceID == "" {
            traceID = uuid.New().String()
        }
        c.Set("traceID", traceID)
        c.Header("X-Trace-ID", traceID)
        c.Next()
    }
}
```

- [ ] **Step 3: 逐步替换 service 层日志**

```go
// ❌ Before
log.Printf("failed to create issue: %v", err)

// ✅ After
import "github.com/rs/zerolog/log"

log.Error().Err(err).
    Uint("project_id", projectID).
    Uint("user_id", userID).
    Msg("failed to create issue")
```

优先替换这 5 个 service：
1. `issue_service.go`
2. `ai_service.go`
3. `auth_service.go`
4. `workflow_service.go`
5. `notification_service.go`

- [ ] **Step 4: 敏感信息脱敏**

```go
// 在日志中自动脱敏敏感字段
func sanitize(v interface{}) interface{} {
    // 如果字段名包含 "password", "token", "secret", "key"
    // 替换值为 "***REDACTED***"
    return v
}
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/common/logger.go backend/internal/middleware/logger.go backend/internal/service/
git commit -m "feat: structured logging with zerolog, traceID, request logging"
```

---

### Task B2.5: 前端组件测试 (2 天)

**Files:**
- Create: `frontend/src/components/__tests__/FilterBar.spec.ts`
- Create: `frontend/src/components/__tests__/IssueTable.spec.ts`
- Create: `frontend/src/components/__tests__/IssueBoard.spec.ts`
- Create: `frontend/src/components/__tests__/IssueForm.spec.ts`
- Create: `frontend/src/components/__tests__/CmdKPalette.spec.ts`
- Create: `frontend/src/components/__tests__/NotificationCenter.spec.ts`

- [ ] **Step 1-6: 为每个组件编写测试**

参考 A1.3 中 QuickFilterChips 的测试模式，为每个组件编写：
- 渲染测试（props 正确渲染）
- 交互测试（事件触发正确）
- 边界测试（空数据、极值）

关键测试示例 — FilterBar:

```typescript
// frontend/src/components/__tests__/FilterBar.spec.ts
describe('FilterBar', () => {
  it('renders filter input', () => { /* ... */ })
  it('adds filter condition on selection', async () => { /* ... */ })
  it('removes filter condition on close click', async () => { /* ... */ })
  it('clears all filters', async () => { /* ... */ })
  it('emits rql update on filter change', async () => { /* ... */ })
  it('restores filters from URL query', async () => { /* ... */ })
})
```

- [ ] **Step 7: Run all frontend tests**

```bash
cd frontend && npx vitest run --coverage
# Expected: component coverage ≥40%
```

- [ ] **Step 8: Commit**

```bash
git add frontend/src/components/__tests__/
git commit -m "test: add component tests for FilterBar, IssueTable, IssueBoard, IssueForm, CmdK, NotificationCenter"
```

---

### Task B2.6: Service 层测试 — 第 2 批 (2 天)

**Files:**
- Create: `backend/internal/service/ai_service_test.go`
- Create: `backend/internal/service/agent_service_test.go`
- Create: `backend/internal/service/page_service_test.go`
- Create: `backend/internal/service/automation_service_test.go`
- Create: `backend/internal/service/webhook_service_test.go`

按 Sprint 1 Task B1.5 的 pattern (sqlmock + table-driven) 编写。

- [ ] **Step 1: ai_service_test.go — Chat/Search/Create 测试**

Mock LLM 响应：

```go
type mockLLMClient struct {
    response string
    err      error
}

func (m *mockLLMClient) Chat(prompt string) (string, error) {
    return m.response, m.err
}

func TestAIService_Chat(t *testing.T) {
    mockLLM := &mockLLMClient{response: `{"message": "Hello!"}`}
    svc := NewAIService(db, mockLLM)

    resp, err := svc.Chat(1, "Hello")
    require.NoError(t, err)
    assert.Contains(t, resp, "Hello!")
}
```

- [ ] **Step 2-5: 其余 4 个 service 测试**

- [ ] **Step 6: Run and verify coverage**

```bash
cd backend && go test ./internal/service/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
# Expected: total coverage ≥50%
```

- [ ] **Step 7: Commit**

```bash
git add backend/internal/service/*_test.go
git commit -m "test: service tests batch 2 (ai, agent, page, automation, webhook), coverage >=50%"
```

---

## Sprint 2 结束检查点 (W6 末)

### 交叉 Review

- [ ] **A review B's output:**
  - 安全加固是否影响用户体验（CSP 是否破坏了内联资源？）
  - 结构化日志是否记录了足够的产品数据（用于用户行为分析）

- [ ] **B review A's output:**
  - UI 组件库是否可访问（a11y：contrast、keyboard nav、ARIA labels）
  - Onboarding 流程是否有 XSS/CSRF 隐患

- [ ] **vinthuy final review:**
  - Wiki 协作 demo 验证：两人同时编辑同一页面
  - Swagger UI 是否可访问
  - 安全扫描零高危
  - Service 覆盖率 ≥50%

### Sprint 2 交付物验证

```bash
# Wiki 验证
# 1. 启动服务器
# 2. 打开两个浏览器 tab 访问同一页面
# 3. 在一个 tab 编辑，观察另一个 tab 实时更新
# 4. 检查版本历史 API: GET /api/pages/:id/versions

# 质量验证
make ci && cd backend && go tool cover -func=coverage.out | grep total  # ≥50%
curl http://localhost:8080/api/docs/index.html  # Swagger UI 可访问
```

**Sprint 2 结束 → 产出功能预览版：Wiki 实时协作 demo、Onboarding 可走通、Swagger 在线。**
