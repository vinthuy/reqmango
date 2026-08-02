# Chat & Messages Design (P3-009~012)

> Date: 2026-08-02
> Status: Approved (pending user review)
> Owner: reqmango team
> PRD Reference: [AI_AGENT_PRD.md](../../../AI_AGENT_PRD.md) §11, P3-009~012

## 1. Overview

为 reqmango 增加工作项聊天与消息能力，使 Agent 能在工作项内通过实时对话与团队成员协作。本特性是 PRD Phase 4 SDLC 流程编排的前置基础设施。

### 范围

| 子能力 | PRD 任务 | 状态 |
|--------|---------|------|
| 工作项聊天（表+API+前端） | P3-009/010/011 | 新建 |
| Agent 自动回复（混合触发 + 记忆检索） | P3-012 | 新建 |
| 实时推送（SSE，复用 sse_hub 顺便修复 BUG-12） | - | 扩展 |
| 表情反应（emoji reactions） | PRD §11.1 | 新建 |

### 关键设计决策

| 决策点 | 选择 | 理由 |
|-------|------|------|
| 传输方式 | SSE（复用现有 sse_hub.go）| 单向推送够用、JWT 鉴权简单、修复 BUG-12 |
| Agent 触发 | 混合（@mention + 状态变化）| 兼顾主动询问与主动建议 |
| 上下文内容 | 记忆检索 + issue 上下文 | 智能且成本可控 |
| 架构方案 | 自洽独立模块（A）| 模块边界清晰、与评论语义分离 |

## 2. 数据模型

新增 migration `000017_chat_systems.up.sql`，3 张表：

```sql
-- 聊天会话
CREATE TABLE chats (
    id BIGSERIAL PRIMARY KEY,
    workspace_id BIGINT NOT NULL,
    project_id BIGINT,
    issue_id BIGINT,                    -- 当前版本仅支持 issue 关联
    type VARCHAR(20) NOT NULL DEFAULT 'issue',  -- issue/group/dm (预留)
    title VARCHAR(255),
    created_by_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_chats_issue ON chats(issue_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chats_workspace ON chats(workspace_id) WHERE deleted_at IS NULL;

-- 消息
CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id BIGINT NOT NULL,
    sender_type VARCHAR(20) NOT NULL,   -- user / agent
    content TEXT NOT NULL,
    content_html TEXT,                  -- markdown 渲染后 HTML (bluemonday 清理)
    reply_to_id BIGINT REFERENCES messages(id),
    mentions JSONB,                     -- [{"type":"user|agent","id":123,"name":"..."}]
    edited_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_messages_chat ON messages(chat_id, created_at);
CREATE INDEX idx_messages_sender ON messages(sender_id, sender_type);

-- 表情反应
CREATE TABLE message_reactions (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    emoji VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(message_id, user_id, emoji)
);
CREATE INDEX idx_reactions_message ON message_reactions(message_id);
```

### 设计要点

- **懒创建**：Issue 创建时不自动建 chat，首次访问 `GET /issues/:id/chat` 时懒创建（避免空 chat 噪音）
- **mentions JSONB**：存储解析后的 @mention 结果，便于通知和审计
- **content_html 必须清理**：用 bluemonday 白名单清理（顺便验证修复 BUG-07 的 XSS 模式）
- **GORM 模型**：`backend/internal/model/chat.go`（Chat, Message, MessageReaction struct）

## 3. API 设计

路由注册在 `backend/internal/router/router.go` 的 `workspaces` 组下，复用 `authMiddleware` + `RequireWorkspaceMember`。

### 聊天会话
| Method | Path | 说明 |
|--------|------|------|
| GET | `/workspaces/:ws/issues/:issueId/chat` | 懒获取/创建 issue 关联 chat（返回 chat + 最近 50 条消息） |
| GET | `/workspaces/:ws/chats/:chatId` | 获取 chat 详情 |
| GET | `/workspaces/:ws/chats/:chatId/messages?cursor=&limit=20` | 分页加载历史消息（游标分页，基于 `created_at`） |

### 消息操作
| Method | Path | 说明 |
|--------|------|------|
| POST | `/workspaces/:ws/chats/:chatId/messages` | 发送消息（body: `{content, reply_to_id?}`）|
| PUT | `/workspaces/:ws/messages/:messageId` | 编辑消息（仅作者，30 分钟窗口内）|
| DELETE | `/workspaces/:ws/messages/:messageId` | 软删除消息（作者或项目管理员）|

### 表情反应
| Method | Path | 说明 |
|--------|------|------|
| POST | `/workspaces/:ws/messages/:messageId/reactions` | 添加表情（body: `{emoji}`）|
| DELETE | `/workspaces/:ws/messages/:messageId/reactions` | 删除表情（body: `{emoji}`）|

### SSE 订阅
| Method | Path | 说明 |
|--------|------|------|
| GET | `/workspaces/:ws/chats/:chatId/stream` | SSE 长连接，订阅 chat 内事件 |

### SSE 事件类型
```
event: message_new
data: {"id":123,"chat_id":1,"sender_id":1,"sender_type":"user","content":"...","created_at":"..."}

event: message_edited
data: {"id":123,"content":"...","edited_at":"..."}

event: message_deleted
data: {"id":123,"deleted_at":"..."}

event: reaction_added
data: {"message_id":123,"user_id":1,"emoji":"👍"}

event: reaction_removed
data: {"message_id":123,"user_id":1,"emoji":"👍"}

event: agent_typing
data: {"chat_id":1,"agent_id":5,"agent_name":"Leader"}
```

### 权限模型
- **读 chat**：必须是 project 成员（issue 关联项目）
- **发消息**：必须是 project 成员
- **编辑/删除**：作者本人（30 分钟窗口内）；项目管理员可删除任意
- **表情**：仅 project 成员可加
- **SSE 订阅**：JWT 验证 + project 成员校验

### 请求/响应示例
```json
// POST /chats/:chatId/messages
Request: {"content": "请帮我分析这个 issue @leader-agent", "reply_to_id": null}
Response: {
  "id": 124,
  "chat_id": 1,
  "sender_id": 1,
  "sender_type": "user",
  "content": "请帮我分析这个 issue @leader-agent",
  "content_html": "<p>请帮我分析这个 issue @leader-agent</p>",
  "mentions": [{"type":"agent","id":5,"name":"leader-agent"}],
  "created_at": "2026-08-02T15:00:00Z"
}
```

## 4. SSE 集成（修复 BUG-12）

现有 `backend/internal/service/sse_hub.go` 是按 `userID` 分组的全局单例，缺"chat room"概念。本设计扩展而不重写。

### SSEHub 扩展（保持向后兼容）

```go
type SSEHub struct {
    mu          sync.RWMutex
    clients     map[uint64][]*SSEClient   // 按用户（保持原有，用于个人通知）
    chatClients map[uint64][]*SSEClient   // 按 chat_id（新增，用于聊天广播）
}

// 新增方法
func (h *SSEHub) RegisterChat(chatID, userID uint64) *SSEClient
func (h *SSEHub) UnregisterChat(chatID uint64, c *SSEClient)
func (h *SSEHub) BroadcastToChat(chatID uint64, event string, data interface{})
```

### SSE 端点实现要点

- **Headers**：`Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`, `X-Accel-Buffering: no`（nginx 不缓冲）
- **心跳**：每 30 秒发 `: heartbeat\n\n` 注释行，防止代理超时（默认 60s）
- **权限校验**：订阅时校验 project 成员资格（不依赖广播时校验）
- **连接关闭**：`c.Request.Context().Done()` 触发 UnregisterChat
- **多 tab**：同一用户多 tab 多客户端，各自独立 SSE 连接，全部接收事件
- **重连**：客户端 EventSource 自动重连（浏览器原生），重连后用 HTTP 拉取最近消息补偿（v1 不实现 Last-Event-ID）

### 广播触发点（在 chat_service 内）

```go
func (s *ChatService) SendMessage(ctx, chatID, senderID, senderType, content, replyTo) (*Message, error) {
    // 1. 解析 @mention
    // 2. 渲染 markdown → HTML（bluemonday 清理）
    // 3. 入库
    // 4. 触发 Agent 自动回复（异步 goroutine）
    // 5. SSE 广播 message_new 事件
    s.hub.BroadcastToChat(chatID, "message_new", message)

    // 6. 通知被 @mention 的用户（推送到个人 SSE）
    for _, m := range message.Mentions {
        if m.Type == "user" {
            s.hub.SendToUser(m.ID, "mention", ...)
        }
    }
    return message, nil
}
```

### BUG-12 修复说明
- 现有 `SendToUser`/`NotifyUser`/`BroadcastEvent` 保留不动（向后兼容）
- 本特性落地后 `SendToUser` 将被 `chat_service` 用于 @mention 通知，BUG-12 自动消除
- 新增 `chatClients` 维度，不影响原有 `clients` 逻辑

## 5. Agent 自动回复（混合触发 + 记忆检索）

### 触发路径 A：@mention（用户主动）

```
用户发消息 "@leader-agent 帮我分析"
  → chat_service.SendMessage 解析 mentions=[{type:agent, id:5}]
  → 异步 goroutine: chat_service.triggerAgentReply(chatID, agentID, userID, trigger="mention")
  → 发送 agent_typing SSE 事件（前端显示"Agent 正在输入..."）
  → 调用 agent_service.DispatchAgent (已有方法)
      → 上下文: agent_template.instructions + 最近 10 条消息 + issue 上下文
      → memory_service.SemanticSearch(issue 工作空间, query=用户消息内容, limit=5)
      → LLM 生成回复
  → 创建 sender_type=agent 的 message
  → SSE 广播 message_new
```

### 触发路径 B：状态变化（Agent 主动）

在 `backend/internal/service/issue_service.go` UpdateIssue 状态转换成功后插入 hook：

```go
// issue_service.go UpdateIssue 内, 状态转换成功后
if oldStateID != newStateID {
    // 异步触发，不阻塞状态转换
    go s.chatSvc.OnIssueStateChanged(ctx, issue.ID, oldStateID, newStateID, userID)
}
```

`ChatService.OnIssueStateChanged` 逻辑：
1. 查找 issue 关联的 chat（无则跳过，不强制创建）
2. 查找被分配给该 issue 的 agent（`issue_agent_service.GetAgentsForIssue`）
3. 对每个 agent（并发 + 防抖）：
   a. 检查防抖缓存：同一 agent+issue 在 30 秒内是否已触发？是→跳过
   b. 写入防抖缓存
   c. 发送 agent_typing SSE 事件
   d. 构造状态变化上下文：
      - issue 标题/描述/类型/优先级
      - 旧状态名 → 新状态名
      - `memory_service.SemanticSearch(query="状态变化 " + newState.Name, limit=5)`
   e. 调用 LLM 生成回复
   f. 创建 agent message + SSE 广播
   g. 失败兜底：静默失败 + 不写入失败消息（不污染聊天）

### LLM Prompt 模板

```
[系统指令] (来自 agent_template.instructions)
你是被分配到工作项 #{issue.sequence_id} 的 {agent_template.name}。

[工作项上下文]
- 标题: {issue.name}
- 类型: {issue_type.name}
- 优先级: {issue.priority}
- 当前状态: {newState.name} (从 {oldState.name} 转入)
- 描述: {issue.description_stripped[:500]}

[相关记忆] (来自 memory_service)
1. {memory[0].content}
2. {memory[1].content}
...

[最近对话] (最近 10 条消息, 不含当前触发)
[10:00] 用户 张三: 这个 issue 怎么处理?
[10:01] Agent Leader: 我建议先...

[任务]
基于状态变化和上下文，提供 1-3 句简明建议。不要重复已知信息。
回复:
```

### 防抖设计

```go
type AgentReplyDebouncer struct {
    mu    sync.Mutex
    cache map[string]time.Time  // key: "agentID:issueID"
}

func (d *AgentReplyDebouncer) Allow(agentID, issueID uint64) bool {
    key := fmt.Sprintf("%d:%d", agentID, issueID)
    d.mu.Lock()
    defer d.mu.Unlock()
    if t, ok := d.cache[key]; ok && time.Since(t) < 30*time.Second {
        return false
    }
    d.cache[key] = time.Now()
    return true
}
```

- 防抖窗口 30 秒（避免状态快速变化时轰炸）
- 内存缓存（v1 不持久化，重启清空可接受）
- 后台 goroutine 每 5 分钟清理过期 key

### 失败兜底

| 失败场景 | 处理 |
|---------|------|
| LLM 调用超时（>30s） | 取消 goroutine，不写消息，记日志 |
| LLM 返回空 | 不写消息 |
| Memory 服务失败 | 跳过记忆，仅用 issue 上下文（降级）|
| Agent 不存在 | 跳过 |
| Chat 不存在（issue 无 chat） | 跳过（不强制创建）|
| 发送 SSE 时连接已断 | 静默（客户端重连后拉历史）|

### issue_service 侵入点

最小化改动：仅在 `issue_service.go` UpdateIssue 方法的状态转换分支增加 1 行 `go s.chatSvc.OnIssueStateChanged(...)`。`chatSvc` 通过 setter 注入（参照已有 `commentSvc.SetAgentService` 模式）。

## 6. 前端组件

### 组件结构

```
frontend/src/components/chat/
├── ChatPanel.vue              # 主容器（嵌入 IssueDetail Tab 或侧栏）
├── ChatMessageList.vue        # 消息列表（虚拟滚动 + 自动滚动到底部）
├── ChatMessage.vue            # 单条消息（user/agent 样式区分）
├── ChatInput.vue              # 输入框（@mention picker + 回车发送 + Shift+回车换行）
├── MessageReactions.vue       # 表情反应组件（👍❤️🎉😢🚀 + emoji picker）
├── MessageActions.vue         # 消息操作菜单（编辑/删除/复制/回复）
└── useChatSSE.ts              # SSE composable（连接管理 + 自动重连）
```

### 嵌入位置：IssueDetail 新增 "Chat" Tab

参照现有 `frontend/src/views/IssueDetail.vue` 的 Tab 结构（Details/Activity/Relations/Attachments/TimeTracking），新增 `Chat` Tab：

```vue
<a-tab-pane key="chat" :tab="t('issue.tabs.chat') + unreadBadge">
  <ChatPanel :issue-id="issue.id" :workspace-id="workspaceId" />
</a-tab-pane>
```

- 未读消息数显示在 Tab 标题（红点徽章）
- 默认不展开（避免干扰现有详情页性能）
- 首次点击 Tab 时懒加载 ChatPanel

### ChatPanel.vue 核心逻辑

- SSE 自动接管实时更新
- 历史消息：onMounted 时 GET `/issues/:id/chat` 拉取最近 50 条
- 滚动到顶：游标加载更老消息
- 滚动到底：新消息自动滚动（如果用户在底部）

### useChatSSE composable

```ts
export function useChatSSE(chatId: number) {
  const events = ref<ChatEvent[]>([])
  let es: EventSource | null = null

  function connect() {
    es = new EventSource(`/api/v1/workspaces/${wsId}/chats/${chatId}/stream`)
    es.addEventListener('message_new', e => events.value.push(JSON.parse(e.data)))
    es.addEventListener('agent_typing', e => typing.value = JSON.parse(e.data))
    // ... 其他事件
    es.onerror = () => { es?.close(); setTimeout(connect, 3000) }  // 3s 重连
  }

  onMounted(connect)
  onUnmounted(() => es?.close())

  return { events, typing }
}
```

### 消息样式区分

| 类型 | 样式 |
|------|------|
| user 消息 | 右对齐，主题色背景（蓝） |
| agent 消息 | 左对齐，灰色背景，agent 头像 + 角色徽章 |
| agent typing | 灰色气泡 + 三点动画 |
| @mention 高亮 | 蓝色 chip，可点击跳转 |
| 已编辑标记 | 灰色小字 "(已编辑)" |
| 软删除消息 | 灰色斜体 "此消息已被删除" |

### @mention picker

复用现有 `frontend/src/components/AgentSelector.vue` 的逻辑，输入 `@` 时弹出 Agent 列表（仅显示被分配到当前 issue 的 agent + workspace 内所有 agent 模板）。

### i18n
- 所有 UI 文本走 i18n（参照项目约束 "All UI text must be properly internationalized"）
- 新增 key：`chat.placeholder`, `chat.edit_window_expired`, `chat.agent_typing`, `chat.reaction.add` 等
- 同步 zh-CN.json 和 en-US.json（顺便修复 BUG-28 部分）

## 7. 错误处理与安全

### 错误处理与边界

| 场景 | 处理 |
|------|------|
| 消息内容超 10000 字符 | 422 返回，前端输入时实时计数 |
| Markdown 渲染失败 | 降级显示纯文本 content |
| 编辑窗口已过（>30min） | 403 Forbidden |
| 删除他人消息且非管理员 | 403 Forbidden |
| 重复添加同表情 | DB UNIQUE 约束拦截，返回 200（幂等）|
| Agent LLM 调用失败 | 静默失败 + 记日志，不污染聊天 |
| SSE 写入阻塞（channel 满）| `select default` 丢弃（已有逻辑，避免阻塞广播）|
| 并发编辑同一消息 | 乐观锁：`edited_at` 字段 + If-Match（v1 简化为最后写入胜出）|
| Chat 不存在 | 404 |
| 非项目成员访问 | 403 |

### 安全要点

| 风险 | 防护 |
|------|------|
| XSS（消息 HTML）| `content_html` 入库前用 bluemonday 清理（白名单 `<p><a><code><strong><em><br>`），顺便验证修复 BUG-07 模式 |
| 越权读 chat | SSE 订阅 + HTTP GET 都校验 project 成员资格 |
| 越权发消息 | POST 时校验 project 成员资格 |
| Agent 越权回复 | Agent 只能回复被分配的 issue 的 chat（issue_agent_service 校验）|
| SSE 连接劫持 | JWT 验证（已有 middleware）|
| 消息注入（恶意 markdown）| bluemonday + markdown 渲染库默认转义 |
| 大量表情刷屏 | DB UNIQUE 约束 + 每用户每消息最多 1 个表情（自然限制）|

### 性能边界（参照 PRD 17.1）

| 指标 | 目标 | 设计保证 |
|------|------|---------|
| SSE 消息延迟 | < 100ms | 内存广播，无 DB 查询 |
| 单 chat 并发连接 | 100+ | goroutine + channel，每连接 1 goroutine |
| 历史消息分页 | < 300ms | 游标分页 + `(chat_id, created_at)` 索引 |
| Agent 回复延迟 | < 5s | LLM 调用 + memory 检索，异步不阻塞 |
| 单 issue 消息数 | 无上限 | 软删除 + 分页 |

## 8. 测试策略

### 单元测试（Go test）
- `service/chat_service_test.go`: SendMessage/Edit/Delete/Reaction CRUD，@mention 解析
- `service/chat_debouncer_test.go`: 防抖逻辑（同 agent+issue 30s 内只触发 1 次）
- `service/sse_hub_test.go`: BroadcastToChat 多客户端分发，Unregister 清理
- `handler/chat_handler_test.go`: 权限校验（非成员 403），编辑窗口（>30min 403）
- `service/issue_service_chat_hook_test.go`: 状态变化触发 chatSvc.OnIssueStateChanged 调用

### 集成测试
- SSE 端到端：发送消息 → SSE 收到 message_new 事件
- Agent 自动回复链路：@mention → agent_typing → message_new（agent 消息）
- 状态变化触发：issue 状态转换 → agent 自动回复消息出现

### 前端测试
- `components/chat/ChatPanel.spec.ts`: 消息列表渲染，发送消息，编辑/删除
- `components/chat/ChatInput.spec.ts`: @mention picker，回车发送，Shift+回车换行
- `components/chat/MessageReactions.spec.ts`: 添加/删除表情

### E2E（Playwright）
`e2e/chat-e2e.spec.ts`:
1. 打开 issue → 切换 Chat Tab → 发送消息 → 实时显示
2. @agent → 出现 typing → 收到 agent 回复
3. 添加表情 → 表情计数更新
4. 编辑消息（30min 内）→ 显示"已编辑"
5. 多 tab 同步：tab1 发消息 → tab2 实时显示

### 手动验收清单
- [ ] 2 个浏览器同时打开同一 issue chat，互相发消息实时同步
- [ ] @agent-name 触发 agent 回复（< 5s）
- [ ] issue 状态从 "Todo" 转为 "In Progress"，被分配 agent 主动提供上下文
- [ ] 状态快速切换 3 次（< 30s），agent 只回复 1 次（防抖生效）
- [ ] LLM 服务挂掉时，发消息正常工作，agent 不回复（静默失败）
- [ ] SSE 断网 30s 后恢复，客户端自动重连

## 9. 实施依赖

### 新建文件
- `backend/migrations/000017_chat_systems.up.sql` / `.down.sql`
- `backend/internal/model/chat.go`
- `backend/internal/handler/chat_handler.go` + `chat_handler_test.go`
- `backend/internal/service/chat_service.go` + `chat_service_test.go`
- `backend/internal/service/chat_debouncer.go` + `chat_debouncer_test.go`
- `backend/internal/dto/request/chat.go`
- `backend/internal/dto/response/chat.go`
- `frontend/src/api/chat.ts`
- `frontend/src/components/chat/` 目录下 7 个组件
- `frontend/src/composables/useChatSSE.ts`
- `frontend/e2e/chat-e2e.spec.ts`

### 修改文件（最小化）
- `backend/internal/router/router.go`: 注册 chat 路由
- `backend/internal/service/sse_hub.go`: 扩展 chatClients 维度（向后兼容）
- `backend/internal/service/issue_service.go`: 增加状态变化 hook（1 行）+ chatSvc setter
- `backend/cmd/server/main.go`: 初始化 chatSvc 并注入 issue_service
- `frontend/src/views/IssueDetail.vue`: 新增 Chat Tab
- `frontend/src/locales/zh-CN.json` + `en-US.json`: 新增 chat.* i18n keys
- `frontend/src/router/index.ts`: 无需修改（chat 嵌入 IssueDetail，无独立路由）

### 复用现有
- `agent_service.DispatchAgent` - Agent 调用 LLM 执行任务
- `memory_service.SemanticSearch` - 记忆检索
- `issue_agent_service.GetAgentsForIssue` - 获取 issue 分配的 agent
- `comment_service` 的 @mention 解析逻辑（参考实现）
- `middleware.RequireWorkspaceMember` - 权限校验
- `service.SSE` 全局单例 - SSE 推送

## 10. 不在范围内（v1 不实现）

- 群聊（group chat）和工作空间 DM
- 消息附件上传（已有 attachment_service，可后续接入）
- 消息搜索（已有 RQL，可后续接入）
- 消息已读回执
- Last-Event-ID SSE 补偿
- 富文本编辑器高级功能（表格、代码块语法高亮等，v1 用基础 markdown）
- 移动端原生适配（响应式 CSS 即可）
- 消息线程（threaded replies，v1 仅单层 reply_to）
- 跨 issue 聊天搜索
- 消息导出
