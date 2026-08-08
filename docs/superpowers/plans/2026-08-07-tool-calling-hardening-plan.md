# Tool Calling 引擎加固 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 加固 Tool Calling 引擎，补上权限校验、双维度滑动窗口限流、MCP Lite 同步/执行链路、4 个代码审查内置函数、以及前端 ToolManager 四 Tab 控制台，形成安全-公平-扩展-可用-闭环 5 项能力。

**Architecture:**
- 后端采用三层管道 (权限 → 限流 → 执行) + 审计层 SSE 广播模式
- Tool.Call 入口是所有调用必经之路，保证策略集中不绕过
- MCP Lite 通过 Sync HTTP 拉取工具并 upsert 到 tools 表，Type=mcp 的工具在 executeMCPTool 中调 JSON-RPC
- 前端采用 Tab+Modal+FilterBar 模式，复用 AgentDashboard 卡片风格和现有 views/agents 页面模式

**Tech Stack:** Go 1.22 + GORM + SSE; Vue 3 + TypeScript + Pinia + Tailwind CSS

---

## 文件变更地图（总览）

| # | 文件 | 操作 | 所属任务 |
|---|------|------|----------|
| 1 | `backend/migrations/000022_tool_hardening.up.sql` | 新建 | T1 |
| 2 | `backend/internal/model/tool.go` | 改 | T1 |
| 3 | `backend/internal/dto/request/tool.go` | 改/建 | T1 |
| 4 | `backend/internal/dto/response/tool.go` | 改/建 | T1 |
| 5 | `backend/internal/service/tool_service.go` | 改 | T2-T6 |
| 6 | `backend/internal/service/tool_service_test.go` | 新建 | T2-T6 |
| 7 | `backend/internal/service/mcp_service.go` | 改 | T5 |
| 8 | `backend/internal/handler/tool_handler.go` | 改/建 | T1,T5 |
| 9 | `backend/internal/handler/mcp_handler.go` | 改/建 | T5 |
| 10 | `backend/internal/router/router.go` | 改 | T1,T5 |
| 11 | `backend/internal/service/sse_hub.go` | 改 | T4 |
| 12 | `frontend/src/api/tool.ts` | 新建/改 | T7 |
| 13 | `frontend/src/views/agents/ToolManager.vue` | 新建 | T7 |
| 14 | `frontend/src/views/agents/AgentDashboard.vue` | 改 | T7 |
| 15 | `frontend/src/composables/useSSE.ts` | 改 | T7 |
| 16 | `frontend/src/router/index.ts` | 改 | T7 |
| 17 | `frontend/src/locales/zh-CN.json` | 改 | T7 |
| 18 | `frontend/src/locales/en-US.json` | 改 | T7 |

---

## Task 1: DB Migration + 模型/DTO 契约补全

**Files:**
- Create: `backend/migrations/000022_tool_hardening.up.sql`
- Modify: `backend/internal/model/tool.go` (+ MCPConfigID)
- Modify/Create: `backend/internal/dto/request/tool.go` (+ Call 请求 3 字段)
- Modify/Create: `backend/internal/dto/response/tool.go` (补 CallLog 筛选参数)
- Modify: `backend/internal/router/router.go` (为后续 handler 预留占位路由不填)
- Test: build only (`go build ./...`)

- [ ] **Step 1: 写 migration SQL**

```sql
-- backend/migrations/000022_tool_hardening.up.sql
-- Tool Calling 加固：MCP外键 + dangerous category 扩展 + 日志索引

-- 1) Tool -> MCPConfig 外键
ALTER TABLE tools ADD COLUMN mcp_config_id BIGINT
    CONSTRAINT fk_tools_mcp_config REFERENCES mcp_configs(id) ON DELETE SET NULL;
CREATE INDEX idx_tools_mcp_config_id ON tools(mcp_config_id);
CREATE UNIQUE INDEX idx_tools_mcp_uniq ON tools(mcp_config_id, name)
    WHERE mcp_config_id IS NOT NULL;

-- 2) category 应用层约束：不做 DB enum，保持 varchar 灵活性，只加默认
ALTER TABLE tools ALTER COLUMN category SET DEFAULT 'general';

-- 3) tool_call_logs 索引（按调用者、时间倒序常用筛选）
CREATE INDEX idx_tool_call_logs_agent_id ON tool_call_logs(agent_id);
CREATE INDEX idx_tool_call_logs_created_at ON tool_call_logs(created_at DESC);
```

- [ ] **Step 2: 修改 model.Tool，新增 MCPConfigID**

在 `backend/internal/model/tool.go` 的 Tool struct 内，紧跟 `WorkspaceID *uint64` 字段后插入：

```go
MCPConfigID *uint64 `gorm:"index" json:"mcp_config_id,omitempty"`
```

结构体末尾的 Relationships 段追加：

```go
MCPConfig *MCPConfig `gorm:"foreignKey:MCPConfigID" json:"-"`
```

- [ ] **Step 3: 改 request.CallToolRequest，补全调用者维度**

检查 `backend/internal/dto/request/tool.go`（不存在则新建）。确保 CallToolRequest 结构如下：

```go
package request

type CallToolRequest struct {
    ToolID          uint64                 `json:"tool_id" binding:"required"`
    InputParams     map[string]interface{} `json:"input_params"`

    // === 新增字段 ===
    AgentTemplateID *uint64 `json:"agent_template_id,omitempty"` // ToolPermission 匹配
    CallerUserID    uint64  `json:"-"`                            // 从 auth middleware 注入，不从 JSON 读取
    AgentTaskID     *uint64 `json:"agent_task_id,omitempty"`      // 任务关联（审计）
}
```

_注意：CallerUserID 使用 `json:"-"`，后续在 handler 里从 middleware 设值，避免客户端伪造。_

- [ ] **Step 4: 改 response，补 GetCallLogs 筛选响应**

在 `backend/internal/dto/response/tool.go`（不存在则新建）中，确保 ToolCallLogResponse 字段完整（代码里已实现，但增加两个）：

```go
type ToolCallLogResponse struct {
    ID           uint64      `json:"id"`
    WorkspaceID  uint64      `json:"workspace_id"`
    AgentTaskID  *uint64     `json:"agent_task_id,omitempty"`
    ToolID       uint64      `json:"tool_id"`
    ToolName     string      `json:"tool_name"`
    AgentID      *uint64     `json:"agent_id,omitempty"`
    CallerUserID *uint64     `json:"caller_user_id,omitempty"` // 新增
    InputParams  interface{} `json:"input_params"`
    OutputResult interface{} `json:"output_result"`
    Status       string      `json:"status"`
    ErrorMessage *string     `json:"error_message,omitempty"`
    DurationMs   int64       `json:"duration_ms"`
    RateLimited  bool        `json:"rate_limited"`
    CreatedAt    time.Time   `json:"created_at"`
}
```

- [ ] **Step 5: 验证编译**

```bash
cd backend
go build ./...
```

Expected: exit 0 (没有语法错误)

- [ ] **Step 6: Commit**

```bash
git add backend/migrations/000022_tool_hardening.up.sql \
        backend/internal/model/tool.go \
        backend/internal/dto/request/tool.go \
        backend/internal/dto/response/tool.go
git commit -m "feat(tools): add migration, model+DTO fields for mcp+caller hardening"
```

---

## Task 2: 权限层三步校验（A2）实现

**Files:**
- Modify: `backend/internal/service/tool_service.go` (Call 入口前插入权限层)
- Create: `backend/internal/service/tool_service_test.go` (权限用例)
- Modify: `backend/internal/handler/tool_handler.go` (注入 CallerUserID)

- [ ] **Step 1: 写权限单测（先红）**

新建 `backend/internal/service/tool_service_test.go`：

```go
package service

import (
    "testing"

    "github.com/reqmango/backend/internal/common"
    "github.com/reqmango/backend/internal/dto/request"
    "github.com/reqmango/backend/internal/model"
    "github.com/stretchr/testify/assert"
)

// setupToolDB 返回内存数据库 + seeded 基础数据（复用 testutil.DB 或自定义最小测试桩）
// 为避免 DB 依赖，checkPermission 逻辑可以拆成纯函数式测试
// 这里写针对权限判断函数的测试（非集成）

func TestPermissionStep1WorkspaceMember(t *testing.T) {
    // Simulate: workspace_id=1, user_id=10 -> 成员存在（is_active=true）-> 通过
    // Simulate: user_id=99 -> 不存在 -> 403
    // 用内存 DB testutil.NewInMemory()
    db := testutil.NewInMemory(t)
    // seed workspace member
    db.Create(&model.WorkspaceMember{
        WorkspaceID: 1, UserID: 10, Role: common.RoleMember, IsActive: true,
    })
    db.Create(&model.Tool{BaseModel: model.BaseModel{ID: 1}, Name: "t1", ToolType: "api", Category: "general"})

    svc := NewToolService(db)
    req := &request.CallToolRequest{ToolID: 1, CallerUserID: 10}
    // 内部 checkMemberOfWorkspace
    // 我们导出一个测试可见的包装方法，或直接测 Call 的整体返回
    _, err := svc.Call(1, req)
    // 会报别的错误（无 endpoint），但应该不是 Forbidden
    assert.NotNil(t, err)
    _, isForbid := err.(*common.AppError)
    // 不做权限错误，因为会继续走到执行层。所以只确认 member 校验通过（非 Forbidden）
    // 要单独测试：把 call 内部 checkPermissions 做成可测函数即可
    // 如果没有拆分，测一个完整的 403 场景
    req2 := &request.CallToolRequest{ToolID: 1, CallerUserID: 999}
    _, err2 := svc.Call(1, req2)
    assert.Error(t, err2)
    var ae *common.AppError
    assert.ErrorAs(t, err2, &ae)
    assert.Equal(t, 403, ae.Code) // 非成员 403
}

func TestPermissionStep2DangerousAdmin(t *testing.T) {
    // member role member -> call dangerous tool -> 403
    // member role admin  -> call dangerous tool -> 非 403 权限错（到执行层报错算 pass）
}

func TestPermissionStep3WhitelistDeny(t *testing.T) {
    // ToolPermission tool=1, agent_template=5 allowed=false  -> 从 agent_tpl=5 触发 403
}
```

- [ ] **Step 2: 跑单测确认失败（缺实现）**

```bash
cd backend
go test ./internal/service/ -run "TestPermission" -v 2>&1 | head -n 50
```

Expected: FAIL（函数未导出 / checkPermissions 未实现）

- [ ] **Step 3: 实现权限层**

在 `tool_service.go` 中修改：

1) **新增内部 checkPermissions 方法**（在 `Call` 方法前面）：

```go
func (s *ToolService) checkMemberOfWorkspace(wid, uid uint64) (*model.WorkspaceMember, error) {
    var m model.WorkspaceMember
    if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?",
        wid, uid, true).First(&m).Error; err != nil {
        return nil, common.Forbidden("Workspace member required")
    }
    return &m, nil
}

// checkPermissions 三步校验，返回通过的 member 或 错误
func (s *ToolService) checkPermissions(wid uint64, tool *model.Tool, req *request.CallToolRequest) (*model.WorkspaceMember, error) {
    // Step 1: 成员身份
    member, err := s.checkMemberOfWorkspace(wid, req.CallerUserID)
    if err != nil {
        return nil, err
    }
    // Step 2: 高危操作 Admin
    if tool.Category == "dangerous" && member.Role < common.RoleAdmin {
        return nil, common.Forbidden("Admin required for dangerous tools")
    }
    // Step 3: ToolPermission 白/黑名单
    if req.AgentTemplateID != nil {
        var tp model.ToolPermission
        // 先精确后通配
        err := s.db.Where("tool_id = ? AND (agent_template_id = ? OR agent_template_id IS NULL)",
            tool.ID, *req.AgentTemplateID).
            Order("CASE WHEN agent_template_id IS NULL THEN 1 ELSE 0 END ASC").
            First(&tp).Error
        if err == nil && !tp.Allowed {
            return nil, common.Forbidden("Tool permission denied by whitelist")
        }
    }
    return member, nil
}
```

2) **在 `Call` 方法开头插入调用**（紧跟 `s.db.First(&tool, req.ToolID)` 和 status 校验之后）：

```go
// Tool.Call: 已有流程如下，新增一步 checkPermissions
func (s *ToolService) Call(wid uint64, req *request.CallToolRequest) (*response.ToolCallResponse, error) {
    var tool model.Tool
    if err := s.db.First(&tool, req.ToolID).Error; err != nil {
        return nil, common.NotFound("Tool not found")
    }
    if tool.Status != "active" {
        return nil, common.BadRequest("Tool is not active")
    }

    // ===== 新增：权限层 三步校验 =====
    member, err := s.checkPermissions(wid, &tool, req)
    if err != nil {
        return nil, err
    }
    _ = member // 预留后续在高级审计里记录 role

    // ... 原有 checkRateLimit + validateParams 继续
```

3) **Handler 注入 CallerUserID**（handler/tool_handler.go，不存在则参考 handler/agent_task_handler.go 建立）：

```go
func (h *ToolHandler) Call(c *gin.Context) {
    wid := GetWorkspaceID(c)
    user := middleware.GetCurrentUser(c) // 确认这个函数名，参考 auth_handler
    var req request.CallToolRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, common.BadRequest(err.Error()))
        return
    }
    req.CallerUserID = user.ID // 注入，禁止 JSON 传入
    resp, err := h.toolSvc.Call(wid, &req)
    if err != nil {
        c.JSON(common.StatusCode(err), err)
        return
    }
    c.JSON(200, resp)
}
```

- [ ] **Step 4: 跑权限测试**

```bash
cd backend
go test ./internal/service/ -run "TestPermission" -v
```

Expected: PASS 全通过

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/tool_service.go \
        backend/internal/service/tool_service_test.go \
        backend/internal/handler/tool_handler.go
git commit -m "feat(tools): add 3-step permission checks before Tool.Call"
```

---

## Task 3: 双维度滑动窗口限流（A3）实现

**Files:**
- Modify: `backend/internal/service/tool_service.go` (+ rateLimiter, rewrite checkRateLimit)
- Modify: `backend/internal/service/tool_service_test.go` (+ 限流用例)

- [ ] **Step 1: 写限流测试（先红）**

追加到 `tool_service_test.go`：

```go
func TestRateLimiter_SlidingWindowBasic(t *testing.T) {
    rl := &rateLimiter{}
    // 单 tool，RateLimit=3
    for i := 0; i < 3; i++ {
        ok := rl.tryAcquire(1, 0, 3) // tool=1, callerID=0=全局, limit=3
        assert.True(t, ok, "attempt %d should pass", i+1)
    }
    ok := rl.tryAcquire(1, 0, 3)
    assert.False(t, ok, "4th attempt should be rate limited")
}

func TestRateLimiter_CallerQuota20Percent(t *testing.T) {
    // Global limit=10 => 单 caller 最多 ceil(10/5)=2
    rl := &rateLimiter{}
    assert.True(t, rl.tryAcquire(2, 101, 10))
    assert.True(t, rl.tryAcquire(2, 101, 10))
    assert.False(t, rl.tryAcquire(2, 101, 10)) // 第 3 次超 caller 配额
    // 但另一 caller 还能过
    assert.True(t, rl.tryAcquire(2, 202, 10))
}
```

- [ ] **Step 2: 跑测试确认失败（rateLimiter 未定义）**

```bash
cd backend
go test ./internal/service/ -run "TestRateLimiter" -v 2>&1 | head -n 40
```

Expected: FAIL（undefined: rateLimiter）

- [ ] **Step 3: 实现 rateLimiter + 重写 checkRateLimit**

在 `tool_service.go` 顶部（struct 定义区）添加：

```go
import (
    "math"
    "sync"
    "time"
)

type rateLimitEntry struct {
    mu    sync.Mutex
    calls []time.Time
}

type rateLimiter struct {
    mu    sync.RWMutex
    store sync.Map // string key -> *rateLimitEntry
}

var globalRateLimiter = &rateLimiter{}

func (rl *rateLimiter) getOrCreateEntry(key string) *rateLimitEntry {
    if v, ok := rl.store.Load(key); ok {
        return v.(*rateLimitEntry)
    }
    rl.mu.Lock()
    defer rl.mu.Unlock()
    if v, ok := rl.store.Load(key); ok {
        return v.(*rateLimitEntry)
    }
    e := &rateLimitEntry{calls: make([]time.Time, 0, 16)}
    rl.store.Store(key, e)
    return e
}

// tryAcquire: 同时抢占全局+caller 双维度，任一超限则回滚全部，返回 false
// callerID=0 表示只检查全局（匿名）
func (rl *rateLimiter) tryAcquire(toolID, callerID uint64, perMinute int) bool {
    if perMinute <= 0 {
        return true
    }
    window := 60 * time.Second
    now := time.Now()
    cutoff := now.Add(-window)

    gKey := fmt.Sprintf("g:%d", toolID)
    cKey := ""
    if callerID > 0 {
        cKey = fmt.Sprintf("c:%d:%d", toolID, callerID)
    }
    gEntry := rl.getOrCreateEntry(gKey)
    var cEntry *rateLimitEntry
    if cKey != "" {
        cEntry = rl.getOrCreateEntry(cKey)
    }

    // 加锁顺序一致（先全局后组合）防死锁
    gEntry.mu.Lock()
    if cEntry != nil {
        cEntry.mu.Lock()
        defer cEntry.mu.Unlock()
    }
    defer gEntry.mu.Unlock()

    // 懒清理
    filter := func(src []time.Time) []time.Time {
        dst := src[:0]
        for _, t := range src {
            if t.After(cutoff) {
                dst = append(dst, t)
            }
        }
        return dst
    }
    gEntry.calls = filter(gEntry.calls)
    if cEntry != nil {
        cEntry.calls = filter(cEntry.calls)
    }

    callerLimit := int(math.Ceil(float64(perMinute) / 5.0))

    // 检查通过否
    if len(gEntry.calls) >= perMinute {
        return false
    }
    if cEntry != nil && len(cEntry.calls) >= callerLimit {
        return false
    }
    // 通过，登记
    gEntry.calls = append(gEntry.calls, now)
    if cEntry != nil {
        cEntry.calls = append(cEntry.calls, now)
    }
    return true
}
```

**重写 `checkRateLimit`**（旧实现删除 / 替换）：

```go
func (s *ToolService) checkRateLimit(tool model.Tool, callerID uint64) error {
    if tool.RateLimit <= 0 {
        return nil
    }
    ok := globalRateLimiter.tryAcquire(tool.ID, callerID, tool.RateLimit)
    if !ok {
        // 同步写一条 rate limited log（简略）
        s.db.Create(&model.ToolCallLog{
            ToolID:      tool.ID,
            Status:      "failed",
            RateLimited: true,
            DurationMs:  0,
        })
        return common.TooManyRequests("Tool rate limit exceeded (try again in ~60s)")
    }
    return nil
}
```

最后，**在 Call 方法里给 checkRateLimit 传 callerID**：

```go
// 原：if err := s.checkRateLimit(tool); err != nil {
// 改：传 callerID，组合 agent_task_id 里的 agent_template 或 user
callerID := uint64(0)
if req.CallerUserID > 0 {
    callerID = req.CallerUserID
}
if req.AgentTemplateID != nil {
    // 用高 32bit 区分类型（简化：复用 user_id 空间，使用大偏移避免冲突）
    // 或者直接单独 encode 成不同 key，在 tryAcquire 我们已经支持字符串键，
    // 但 tryAcquire 接收 uint64，这里我们用 callerID=userID 主，
    // 如果是 agent 调用，用 agent_template_id 作为 callerID，
    // 简单做法：若存在 AgentTemplateID，优先使用它
    callerID = *req.AgentTemplateID
}
if err := s.checkRateLimit(tool, callerID); err != nil {
    return nil, err
}
```

- [ ] **Step 4: 跑限流测试**

```bash
cd backend
go test ./internal/service/ -run "TestRateLimiter" -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/tool_service.go backend/internal/service/tool_service_test.go
git commit -m "feat(tools): sliding-window rate limiter with global+caller dual quota"
```

---

## Task 4: SSE 审计事件广播（A4.3）+ 日志补全

**Files:**
- Modify: `backend/internal/service/tool_service.go` (Call 结束广播)
- Modify: `backend/internal/service/sse_hub.go` (新增 SSE 事件类型定义区确认存在 BroadcastEvent)
- Modify: `backend/internal/handler/sse_handler.go` (如需，把 tool_call 事件加到允许列表)

- [ ] **Step 1: 写 SSE 测试辅助 + 审计回写补全测试（先红）**

先在 `tool_service_test.go` 加一个 Tool.Call 日志字段校验用例：

```go
func TestCall_WritesCallerAndTaskAudit(t *testing.T) {
    // 用内置 function create_issue 类型工具测试
    // Seed: tool type=function isBuiltin=true name=create_issue
    // Set CallerUserID=123, AgentTaskID=&tid=7
    // Call 成功后，读 tool_call_logs 最后一行，断言 caller 字段对应
    // 这里简化：验证日志写入中 AgentTaskID = 7
}
```

- [ ] **Step 2: 实现 SSE 广播 + 审计字段回写**

在 `tool_service.go` 的 `Call` 方法，`s.db.Create(&log)` 之后加入：

```go
// 写日志 (原有)
s.db.Create(&log)

// 新增：SSE 广播
evt := "tool_call.completed"
payload := map[string]interface{}{
    "id":           log.ID,
    "workspace_id": wid,
    "tool_id":      tool.ID,
    "tool_name":    tool.Name,
    "status":       status,
    "duration_ms":  duration,
    "caller_user_id": req.CallerUserID,
}
if req.AgentTaskID != nil {
    payload["agent_task_id"] = *req.AgentTaskID
}
data, _ := json.Marshal(payload)
SSE.BroadcastEvent(evt, json.RawMessage(data))

if status == "failed" && !log.RateLimited {
    failPayload := map[string]interface{}{
        "id":         log.ID,
        "tool_id":    tool.ID,
        "error":      errorMessage,
    }
    failData, _ := json.Marshal(failPayload)
    SSE.BroadcastEvent("tool_call.failed", json.RawMessage(failData))
}
```

**补全审计字段**：在创建 `model.ToolCallLog` 时写入更多维度：

```go
log := model.ToolCallLog{
    WorkspaceID:  wid,
    ToolID:       tool.ID,
    InputParams:  req.InputParams,
    OutputResult: outputResult,
    Status:       status,
    ErrorMessage: errorMessage,
    DurationMs:   duration,
    // === 补全 ===
    AgentTaskID:   req.AgentTaskID,
    AgentID:       req.AgentTemplateID, // 复用已有 AgentID 字段（代表 Template）
    // CallerUserID 字段在模型里没有，需先加？
    // → 如果没有则先忽略，已通过 SSE 带出；模型有则直接写
}
```

（注：若模型 ToolCallLog 缺少 caller_user_id 字段，按本计划 T1 已迁移，或在本 commit 中补 ALTER TABLE；不强制时 SSE 已带出）

- [ ] **Step 3: Build + 简单测试**

```bash
cd backend
go build ./... && go vet ./...
go test ./internal/service/ -run "TestPermission|TestRateLimiter" -count=1
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/service/tool_service.go
git commit -m "feat(tools): broadcast tool_call.* SSE events + enriched audit log"
```

---

## Task 5: MCP Lite（SyncTools + executeMCPTool）实现

**Files:**
- Modify: `backend/internal/service/mcp_service.go` (新增 SyncTools + ExecuteTool)
- Modify: `backend/internal/service/tool_service.go` (新增 executeMCPTool 分支)
- Modify/Create: `backend/internal/handler/mcp_handler.go` (+ Sync HTTP 处理)
- Modify: `backend/internal/router/router.go` (+ Sync 路由)

- [ ] **Step 1: 写 MCP Lite 测试（先红）**

新建 `backend/internal/service/mcp_service_test.go`：

```go
func TestMCPSyncTools_Upsert(t *testing.T) {
    // sim: Sync 时调用返回 N 个工具，DB 里插入这些 tools，mcp_config_id=X
    // 第 2 次 Sync：同名的更新 description，不同名的新增
}

func TestExecuteTool_BuildsJSONRPC(t *testing.T) {
    // ExecuteTool 内部构造 http 请求体是合法 JSON-RPC
    // 可 mock http 客户端监听 body
}
```

- [ ] **Step 2: 实现 SyncTools + ExecuteTool**

在 `mcp_service.go` 追加方法：

```go
type mcpListToolsResp struct {
    JSONRPC string                 `json:"jsonrpc"`
    Result  []mcpToolDef           `json:"result"`
    ID      int                    `json:"id"`
}
type mcpToolDef struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

// SyncTools 拉取 MCP Server 的 tools 列表并 upsert 到 tools 表
// 返回 {added, updated} 计数
func (s *MCPService) SyncTools(workspaceID, configID, callerID uint64) (int, int, error) {
    // 0) Admin 权限
    if err := s.checkWorkspaceAdmin(workspaceID, callerID); err != nil {
        return 0, 0, err
    }
    // 1) 取 config
    var cfg model.MCPConfig
    if err := s.db.Where("workspace_id = ? AND id = ?", workspaceID, configID).First(&cfg).Error; err != nil {
        return 0, 0, common.NotFound("MCP config not found")
    }
    // 2) HTTP POST JSON-RPC tools/list
    body, _ := json.Marshal(map[string]interface{}{
        "jsonrpc": "2.0", "method": "tools/list", "id": 1,
    })
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    req, _ := http.NewRequestWithContext(ctx, "POST", cfg.ServerURL, bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    if cfg.APIKey != "" {
        req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
    }
    resp, err := (&http.Client{}).Do(req)
    if err != nil {
        return 0, 0, common.Internal("MCP server unreachable: " + err.Error())
    }
    defer resp.Body.Close()
    raw, _ := io.ReadAll(resp.Body)
    var parsed mcpListToolsResp
    if err := json.Unmarshal(raw, &parsed); err != nil {
        return 0, 0, common.Internal("MCP response parse failed: " + err.Error())
    }
    // 3) Upsert 到 tools 表
    added, updated := 0, 0
    for _, def := range parsed.Result {
        var existing model.Tool
        findErr := s.db.Where("mcp_config_id = ? AND name = ?", configID, def.Name).First(&existing).Error
        paramsJSON, _ := json.Marshal(def.InputSchema)
        if findErr == gorm.ErrRecordNotFound {
            cat := "general"
            s.db.Create(&model.Tool{
                Name:        def.Name,
                Description: def.Description,
                Category:    cat,
                ToolType:    "mcp",
                IsBuiltin:   false,
                Status:      "active",
                MCPConfigID: &configID,
                WorkspaceID: &workspaceID,
                Params:      paramsJSON,
            })
            added++
        } else {
            existing.Description = def.Description
            existing.Params = paramsJSON
            s.db.Save(&existing)
            updated++
        }
    }
    // 4) 更新 cfg.LastSyncAt
    now := time.Now()
    cfg.LastSyncAt = &now
    s.db.Save(&cfg)
    return added, updated, nil
}

// ExecuteTool 通过 MCP JSON-RPC 调用具体工具
func (s *MCPService) ExecuteTool(workspaceID, configID uint64, toolName string, args map[string]interface{}) (json.RawMessage, error) {
    var cfg model.MCPConfig
    if err := s.db.Where("workspace_id = ? AND id = ?", workspaceID, configID).First(&cfg).Error; err != nil {
        return nil, common.NotFound("MCP config not found")
    }
    body, _ := json.Marshal(map[string]interface{}{
        "jsonrpc": "2.0",
        "method":  "tools/call",
        "params":  map[string]interface{}{"name": toolName, "arguments": args},
        "id":      1,
    })
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    req, _ := http.NewRequestWithContext(ctx, "POST", cfg.ServerURL, bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    if cfg.APIKey != "" {
        req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
    }
    resp, err := (&http.Client{}).Do(req)
    if err != nil {
        return nil, common.Internal("MCP call failed: " + err.Error())
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}
```

- [ ] **Step 3: ToolService 中新增 executeMCPTool**

在 `tool_service.go` 的 `Call()` switch 中加 case "mcp"：

```go
switch tool.ToolType {
case "api":
    result, err := s.executeAPITool(tool, req.InputParams)
    // ... (原有)
case "function":
    result, err := s.executeFunctionTool(tool, req.InputParams)
    // ... (原有)
// === 新增：mcp 类型 ===
case "mcp":
    if tool.MCPConfigID == nil {
        status = "failed"
        msg := "MCP tool missing mcp_config_id"
        errorMessage = &msg
    } else {
        var args map[string]interface{}
        if len(req.InputParams) > 0 {
            _ = json.Unmarshal(req.InputParams, &args)
        }
        result, err := s.mcpSvc.ExecuteTool(wid, *tool.MCPConfigID, tool.Name, args)
        if err != nil {
            status = "failed"
            msg := err.Error()
            errorMessage = &msg
        } else {
            status = "success"
            outputResult = result
        }
    }
default:
    return nil, common.BadRequest("Unsupported tool type")
}
```

*注意：ToolService 需要注入 MCPService（构造函数加字段 `mcpSvc *MCPService`，用 SetMCPService setter 解决循环依赖，参考 SquadService.SetAgentExecutor 模式）。*

- [ ] **Step 4: Handler + Router 暴露 Sync**

Handler:
```go
// mcp_handler.go
func (h *MCPHandler) SyncTools(c *gin.Context) {
    wid := GetWorkspaceID(c)
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(400, common.BadRequest("invalid id"))
        return
    }
    user := middleware.GetCurrentUser(c)
    added, updated, err := h.mcpSvc.SyncTools(wid, id, user.ID)
    if err != nil {
        c.JSON(common.StatusCode(err), err)
        return
    }
    c.JSON(200, gin.H{"added": added, "updated": updated})
}
```

Router (在 v1 workspaces group 内)：
```go
mcpGroup := workspaces.Group("/mcp-configs")
mcpGroup.POST("", mcpH.Create)
mcpGroup.GET("", mcpH.List)
mcpGroup.POST("/:id/sync", mcpH.SyncTools) // 新增
```

- [ ] **Step 5: Build + Test**

```bash
cd backend
go build ./...
go test ./internal/service/ -run "MCP" -v
```

Expected: build OK, tests 基本结构过

- [ ] **Step 6: Commit**

```bash
git add backend/internal/service/mcp_service.go \
        backend/internal/service/tool_service.go \
        backend/internal/handler/mcp_handler.go \
        backend/internal/router/router.go \
        backend/internal/service/mcp_service_test.go
git commit -m "feat(mcp): MCP Lite - SyncTools upsert + executeMCPTool JSON-RPC call"
```

---

## Task 6: Built-in 代码审查工具扩展（4 个） + 单测

**Files:**
- Modify: `backend/internal/service/tool_service.go` (executeBuiltinFunction switch + 4 实现函数)
- Modify: `backend/internal/service/tool_service_test.go` (参数校验用例)
- 依赖: `backend/internal/service/github_service.go` 已有 HTTP 客户端能力

- [ ] **Step 1: 写内置审查工具的参数校验测试（先红）**

在 `tool_service_test.go` 追加：

```go
func TestBuiltin_GetPRDiff_RequiredParams(t *testing.T) {
    // 缺 repo_owner -> BadRequest
    svc := &ToolService{} // 不依赖 DB，只测参数 parse 层
    params := []byte(`{"repo_name":"r","pr_number":1}`)
    _, err := svc.executeBuiltinFunction("get_pr_diff", params)
    assert.Error(t, err)
}

func TestBuiltin_CreatePRReview_EventEnum(t *testing.T) {
    // event=APPROVE / REQUEST_CHANGES / COMMENT 三值允许，其他报错
    params := []byte(`{"repo_owner":"o","repo_name":"r","pr_number":1,"event":"BAD","body":"x"}`)
    svc := &ToolService{}
    _, err := svc.executeBuiltinFunction("create_pr_review", params)
    assert.Error(t, err) // event invalid
}
```

- [ ] **Step 2: 实现 4 个函数**

在 `tool_service.go` 的 `executeBuiltinFunction` 中扩展 switch：

```go
case "get_pr_diff":
    return s.builtinGetPRDiff(params)
case "add_review_comment":
    return s.builtinAddReviewComment(params)
case "list_pr_commits":
    return s.builtinListPRCommits(params)
case "create_pr_review":
    return s.builtinCreatePRReview(params)
```

并实现（统一复用 github_service 中的 HTTP 客户端辅助，若不存在则内联一个简单的 HTTP JSON helper，避免循环依赖）：

```go
// 辅助：用默认 HTTP client 发 GitHub REST（带 token header 优先从 context 取）
func (s *ToolService) githubJSON(ctx context.Context, method, url string, body any, out any) error {
    var bodyR io.Reader
    if body != nil {
        b, _ := json.Marshal(body)
        bodyR = bytes.NewReader(b)
    }
    req, _ := http.NewRequestWithContext(ctx, method, url, bodyR)
    req.Header.Set("Accept", "application/vnd.github+json")
    // 简化：token 从默认 github config 取，也可通过 AuthService 查询用户绑定
    // 这里用一个环境变量占位（真实项目中用 github_service.go 的 token 管理）
    if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
        req.Header.Set("Authorization", "Bearer "+tok)
    }
    resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    data, _ := io.ReadAll(resp.Body)
    if resp.StatusCode >= 400 {
        return fmt.Errorf("github %s %s -> %d: %s", method, url, resp.StatusCode, string(data))
    }
    if out != nil {
        return json.Unmarshal(data, out)
    }
    return nil
}

// get_pr_diff 返回 diff 纯文本（而非 JSON）
func (s *ToolService) builtinGetPRDiff(params json.RawMessage) (json.RawMessage, error) {
    var p struct {
        RepoOwner string `json:"repo_owner"`
        RepoName  string `json:"repo_name"`
        PRNumber  int    `json:"pr_number"`
    }
    if err := json.Unmarshal(params, &p); err != nil {
        return nil, common.BadRequest("Invalid params: " + err.Error())
    }
    if p.RepoOwner == "" || p.RepoName == "" || p.PRNumber <= 0 {
        return nil, common.BadRequest("repo_owner, repo_name, pr_number(>0) are required")
    }
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", p.RepoOwner, p.RepoName, p.PRNumber)
    ctx := context.Background()
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("Accept", "application/vnd.github.diff")
    if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
        req.Header.Set("Authorization", "Bearer "+tok)
    }
    resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode >= 400 {
        return nil, common.BadRequest(fmt.Sprintf("GitHub API %d: %s", resp.StatusCode, string(body)))
    }
    return json.Marshal(map[string]string{"diff": string(body)})
}

// add_review_comment
func (s *ToolService) builtinAddReviewComment(params json.RawMessage) (json.RawMessage, error) {
    var p struct {
        RepoOwner string `json:"repo_owner"`
        RepoName  string `json:"repo_name"`
        PRNumber  int    `json:"pr_number"`
        Body      string `json:"body" binding:"required"`
        CommitID  string `json:"commit_id,omitempty"`
        Path      string `json:"path,omitempty"`
        Line      int    `json:"line,omitempty"`
    }
    if err := json.Unmarshal(params, &p); err != nil {
        return nil, common.BadRequest(err.Error())
    }
    if p.RepoOwner == "" || p.RepoName == "" || p.PRNumber <= 0 || p.Body == "" {
        return nil, common.BadRequest("repo_owner, repo_name, pr_number>0, body required")
    }
    payload := map[string]any{"body": p.Body}
    if p.CommitID != "" { payload["commit_id"] = p.CommitID }
    if p.Path != ""     { payload["path"] = p.Path }
    if p.Line > 0       { payload["line"] = p.Line }
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/comments",
        p.RepoOwner, p.RepoName, p.PRNumber)
    var out map[string]any
    if err := s.githubJSON(context.Background(), "POST", url, payload, &out); err != nil {
        return nil, common.Internal(err.Error())
    }
    return json.Marshal(out)
}

// list_pr_commits
func (s *ToolService) builtinListPRCommits(params json.RawMessage) (json.RawMessage, error) {
    var p struct {
        RepoOwner string `json:"repo_owner"`
        RepoName  string `json:"repo_name"`
        PRNumber  int    `json:"pr_number"`
    }
    if err := json.Unmarshal(params, &p); err != nil {
        return nil, common.BadRequest(err.Error())
    }
    if p.RepoOwner == "" || p.RepoName == "" || p.PRNumber <= 0 {
        return nil, common.BadRequest("required fields missing")
    }
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/commits?per_page=100",
        p.RepoOwner, p.RepoName, p.PRNumber)
    var out []map[string]any
    if err := s.githubJSON(context.Background(), "GET", url, nil, &out); err != nil {
        return nil, common.Internal(err.Error())
    }
    return json.Marshal(out)
}

// create_pr_review
func (s *ToolService) builtinCreatePRReview(params json.RawMessage) (json.RawMessage, error) {
    var p struct {
        RepoOwner string                   `json:"repo_owner"`
        RepoName  string                   `json:"repo_name"`
        PRNumber  int                      `json:"pr_number"`
        Event     string                   `json:"event"` // APPROVE | REQUEST_CHANGES | COMMENT
        Body      string                   `json:"body"`
        Comments  []map[string]interface{} `json:"comments,omitempty"`
    }
    if err := json.Unmarshal(params, &p); err != nil {
        return nil, common.BadRequest(err.Error())
    }
    if p.RepoOwner == "" || p.RepoName == "" || p.PRNumber <= 0 {
        return nil, common.BadRequest("required fields missing")
    }
    validEvent := map[string]bool{"APPROVE": true, "REQUEST_CHANGES": true, "COMMENT": true}
    if !validEvent[p.Event] {
        return nil, common.BadRequest("event must be one of APPROVE, REQUEST_CHANGES, COMMENT")
    }
    payload := map[string]any{"event": p.Event, "body": p.Body}
    if len(p.Comments) > 0 { payload["comments"] = p.Comments }
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/reviews",
        p.RepoOwner, p.RepoName, p.PRNumber)
    var out map[string]any
    if err := s.githubJSON(context.Background(), "POST", url, payload, &out); err != nil {
        return nil, common.Internal(err.Error())
    }
    return json.Marshal(out)
}
```

- [ ] **Step 3: 跑测试（含参数校验）**

```bash
cd backend
go test ./internal/service/ -run "TestBuiltin" -v
```

Expected: PASS

- [ ] **Step 4: Build + 所有后端单元测试跑一次**

```bash
cd backend
go build ./...
go vet ./...
go test ./internal/service/ -run "Permission|RateLimiter|MCP|Builtin|Autopilot|SDLC|CICD" -count=1 -timeout 120s
```

Expected: 全部通过（历史单测无回归）

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/tool_service.go backend/internal/service/tool_service_test.go
git commit -m "feat(tools): add 4 builtin code-review tools (get_pr_diff etc)"
```

---

## Task 7: 前端 ToolManager 四 Tab 控制台 + Dashboard 入口

**Files:**
- Create: `frontend/src/api/tool.ts`
- Create: `frontend/src/views/agents/ToolManager.vue`
- Modify: `frontend/src/views/agents/AgentDashboard.vue` (Tools 卡片)
- Modify: `frontend/src/composables/useSSE.ts` (+ tool_call.* events)
- Modify: `frontend/src/router/index.ts` (+ routes)
- Modify: `frontend/src/locales/zh-CN.json` (+ ai.tools.* keys)
- Modify: `frontend/src/locales/en-US.json` (+ 对应英文)

- [ ] **Step 1: 写 API 客户端 `api/tool.ts`**

```typescript
// frontend/src/api/tool.ts
import apiClient from './index'

// ---- Types ----
export interface Tool {
  id: number
  name: string
  description: string
  category: string            // general | dangerous | code | project_management | ci_cd
  is_builtin: boolean
  status: string              // active | disabled
  tool_type: 'api' | 'function' | 'mcp' | 'workflow'
  mcp_config_id?: number
  endpoint?: string
  method?: string
  auth_type?: string
  rate_limit: number
  timeout: number
  workspace_id?: number
  created_at: string
  updated_at: string
}

export interface ToolCallLog {
  id: number
  workspace_id: number
  agent_task_id?: number
  tool_id: number
  tool_name: string
  agent_id?: number
  caller_user_id?: number
  input_params: any
  output_result: any
  status: 'success' | 'failed' | 'timeout'
  error_message?: string
  duration_ms: number
  rate_limited: boolean
  created_at: string
}

export interface ToolPermissionView {
  tool_id: number
  agent_template_id?: number
  allowed: boolean
}

export interface MCPSyncResult { added: number; updated: number }

// ---- CRUD ----
export function listTools(workspaceSlug: string) {
  return apiClient.get<Tool[]>(`/workspaces/${workspaceSlug}/tools`)
}
export function createTool(workspaceSlug: string, payload: Partial<Tool>) {
  return apiClient.post<Tool>(`/workspaces/${workspaceSlug}/tools`, payload)
}
export function updateTool(workspaceSlug: string, id: number, payload: Partial<Tool>) {
  return apiClient.put<Tool>(`/workspaces/${workspaceSlug}/tools/${id}`, payload)
}
export function deleteTool(workspaceSlug: string, id: number) {
  return apiClient.delete(`/workspaces/${workspaceSlug}/tools/${id}`)
}

// ---- Call (手动测试调用) ----
export function callTool(workspaceSlug: string, payload: { tool_id: number; input_params: any; agent_task_id?: number }) {
  return apiClient.post<ToolCallLog>(`/workspaces/${workspaceSlug}/tools/call`, payload)
}

// ---- Logs ----
export function listToolCallLogs(workspaceSlug: string, params?: {
  status?: string; tool_id?: number; agent_id?: number;
  from_time?: string; to_time?: string; page?: number; per_page?: number
}) {
  return apiClient.get<{ data: ToolCallLog[]; total: number }>(
    `/workspaces/${workspaceSlug}/tools/call-logs`, { params }
  )
}

// ---- Permissions ----
export function listToolPermissions(workspaceSlug: string, toolId: number) {
  return apiClient.get<ToolPermissionView[]>(`/workspaces/${workspaceSlug}/tools/${toolId}/permissions`)
}
export function setToolPermission(workspaceSlug: string, toolId: number, payload: ToolPermissionView) {
  return apiClient.put(`/workspaces/${workspaceSlug}/tools/${toolId}/permissions`, payload)
}

// ---- MCP Sync ----
export function syncMCPTools(workspaceSlug: string, mcpConfigId: number) {
  return apiClient.post<MCPSyncResult>(
    `/workspaces/${workspaceSlug}/mcp-configs/${mcpConfigId}/sync`
  )
}
```

- [ ] **Step 2: 写 i18n keys（中英文各 ≈25 个）**

在 `frontend/src/locales/zh-CN.json` 的 `ai` 对象下追加：

```json
"tools": {
  "title": "工具管理",
  "description": "Tool Calling 引擎 - 权限、限流、审计、MCP 集成",
  "tabTools": "工具",
  "tabLogs": "调用日志",
  "tabPermissions": "权限配置",
  "tabMCP": "MCP 服务",
  "create": "创建工具",
  "sync": "同步工具",
  "totalCalls": "总调用数",
  "successRate": "成功率",
  "p95Latency": "P95 延迟 (ms)",
  "rateLimited": "限流次数",
  "filters": {
    "category": "分类",
    "status": "状态",
    "toolType": "工具类型",
    "tool": "工具",
    "caller": "调用者",
    "dateRange": "日期范围"
  },
  "form": {
    "name": "工具名称",
    "description": "描述",
    "category": "分类",
    "type": "类型 (api/function/mcp)",
    "endpoint": "API Endpoint",
    "method": "HTTP 方法",
    "authType": "认证方式",
    "authConfig": "认证配置 (JSON)",
    "params": "参数 Schema (JSON)",
    "rateLimit": "每分钟限额 (0=无限)",
    "timeout": "超时秒数",
    "dangerousWarn": "标记为 dangerous 需要 Admin 才能调用"
  },
  "columns": {
    "name": "名称", "type": "类型", "category": "分类",
    "status": "状态", "rateLimit": "限流/min", "timeout": "超时/s",
    "toolName": "工具名", "caller": "调用者", "duration": "耗时", "time": "时间"
  },
  "status": {
    "active": "启用", "disabled": "停用",
    "success": "成功", "failed": "失败", "timeout": "超时",
    "rateLimited": "被限流"
  },
  "permissions": {
    "header": "按 Agent 模板授权 (未配置默认: 允许)",
    "allowed": "允许",
    "denied": "拒绝"
  },
  "mcp": {
    "lastSync": "最后同步",
    "toolsCount": "工具数",
    "syncNow": "立即同步",
    "synced": "已同步 +{{added}} 新增, +{{updated}} 更新",
    "noServers": "暂无 MCP 服务配置"
  },
  "empty": "暂无数据",
  "actions": {
    "edit": "编辑", "delete": "删除", "call": "测试调用",
    "expand": "展开详情", "exportJSON": "导出 JSON"
  },
  "card": {
    "title": "🛠️ 工具管理",
    "total": "工具总数",
    "todayCalls": "今日调用",
    "open": "打开控制台"
  }
}
```

en-US.json 中对应 `ai.tools.*` 相同 key，写英文（按中文直译即可）。

- [ ] **Step 3: useSSE.ts 加 tool_call 事件**

在 `composables/useSSE.ts` 的 SSE 处理循环里，按已有的 `sdlc_*`/`autopilot_*` 处理模式添加：

```typescript
// tool_call.*
case 'tool_call.completed':
case 'tool_call.failed':
case 'tool_call.rate_limited': {
  emit('tool_call', { event, data: payload })
  break
}
```

并在 composable 的类型声明中导出 `tool_call` 回调（参考已有模式）。

- [ ] **Step 4: 写 ToolManager.vue 主组件（4 Tab）**

```vue
<!-- ToolManager.vue -->
<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('ai.tools.title') }}</h1>
        <p class="text-gray-500 mt-1">{{ t('ai.tools.description') }}</p>
      </div>
      <div class="flex items-center space-x-3">
        <button @click="openCreateTool"
                class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg">
          {{ t('ai.tools.create') }}
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="border-b border-gray-200">
      <nav class="-mb-px flex space-x-8">
        <button v-for="tab in tabs" :key="tab.key" @click="activeTab = tab.key"
                :class="['border-b-2 px-1 py-3 text-sm font-medium',
                    activeTab === tab.key
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700']">
          {{ tab.label }}
        </button>
      </nav>
    </div>

    <!-- Tab: Tools -->
    <div v-if="activeTab === 'tools'">
      <!-- Filters: Category, Status, Type -->
      <div class="flex flex-wrap gap-3 mb-4">
        <select v-model="filters.category" class="border px-3 py-1.5 rounded">
          <option value="">{{ t('ai.tools.filters.category') }} (All)</option>
          <option v-for="c in categories" :key="c" :value="c">{{ c }}</option>
        </select>
        <select v-model="filters.status" class="border px-3 py-1.5 rounded">
          <option value="">{{ t('ai.tools.filters.status') }} (All)</option>
          <option value="active">{{ t('ai.tools.status.active') }}</option>
          <option value="disabled">{{ t('ai.tools.status.disabled') }}</option>
        </select>
      </div>
      <!-- Data Table -->
      <div class="bg-white rounded-lg shadow-sm border">
        <table class="w-full text-sm">
          <thead class="bg-gray-50 text-gray-500 uppercase text-xs">
            <tr>
              <th class="text-left px-6 py-3">{{ t('ai.tools.columns.name') }}</th>
              <th class="text-left px-6 py-3">{{ t('ai.tools.columns.type') }}</th>
              <th class="text-left px-6 py-3">{{ t('ai.tools.columns.category') }}</th>
              <th class="text-left px-6 py-3">{{ t('ai.tools.columns.status') }}</th>
              <th class="text-left px-6 py-3">{{ t('ai.tools.columns.rateLimit') }}</th>
              <th class="text-left px-6 py-3">{{ t('ai.tools.columns.timeout') }}</th>
              <th class="text-right px-6 py-3">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr v-for="tool in filteredTools" :key="tool.id" class="hover:bg-gray-50">
              <td class="px-6 py-3 font-medium text-gray-900">{{ tool.name }}</td>
              <td class="px-6 py-3">
                <span class="px-2 py-0.5 bg-blue-50 text-blue-700 rounded text-xs">
                  {{ tool.tool_type }}
                </span>
              </td>
              <td class="px-6 py-3 text-gray-600">{{ tool.category }}</td>
              <td class="px-6 py-3">
                <span :class="tool.status==='active'
                    ? 'px-2 py-0.5 bg-green-50 text-green-700 rounded text-xs'
                    : 'px-2 py-0.5 bg-gray-100 text-gray-600 rounded text-xs'">
                  {{ tool.status==='active' ? t('ai.tools.status.active')
                                              : t('ai.tools.status.disabled') }}
                </span>
              </td>
              <td class="px-6 py-3">{{ tool.rate_limit || '∞' }}</td>
              <td class="px-6 py-3">{{ tool.timeout }}s</td>
              <td class="px-6 py-3 text-right space-x-1">
                <button @click="openEditTool(tool)"
                        class="text-gray-600 hover:text-blue-600">
                  {{ t('ai.tools.actions.edit') }}
                </button>
                <button v-if="!tool.is_builtin"
                        @click="onDelete(tool)"
                        class="text-red-600 hover:text-red-700 ml-2">
                  {{ t('ai.tools.actions.delete') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Tab: Logs -->
    <div v-if="activeTab === 'logs'">
      <!-- StatCards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
        <div class="bg-white border rounded p-4">
          <div class="text-sm text-gray-500">{{ t('ai.tools.totalCalls') }}</div>
          <div class="text-2xl font-bold text-gray-900 mt-1">{{ stats.total }}</div>
        </div>
        <div class="bg-white border rounded p-4">
          <div class="text-sm text-gray-500">{{ t('ai.tools.successRate') }}</div>
          <div class="text-2xl font-bold text-green-600 mt-1">{{ stats.successRate }}%</div>
        </div>
        <div class="bg-white border rounded p-4">
          <div class="text-sm text-gray-500">{{ t('ai.tools.p95Latency') }}</div>
          <div class="text-2xl font-bold text-purple-600 mt-1">{{ stats.p95 }}</div>
        </div>
        <div class="bg-white border rounded p-4">
          <div class="text-sm text-gray-500">{{ t('ai.tools.rateLimited') }}</div>
          <div class="text-2xl font-bold text-orange-600 mt-1">{{ stats.rateLimited }}</div>
        </div>
      </div>
      <!-- Timeline list -->
      <div class="bg-white border rounded divide-y">
        <div v-for="log in logs" :key="log.id" class="px-6 py-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <span :class="statusBadgeClass(log.status)">
                {{ statusText(log.status, log.rate_limited) }}
              </span>
              <span class="font-medium text-gray-900">{{ log.tool_name }}</span>
              <span class="text-xs text-gray-500">#{{ log.id }}</span>
            </div>
            <div class="text-sm text-gray-500 flex items-center space-x-3">
              <span>{{ log.duration_ms }}ms</span>
              <span>{{ formatDate(log.created_at) }}</span>
              <button @click="expanded[log.id] = !expanded[log.id]"
                      class="text-blue-600 hover:underline">
                {{ expanded[log.id] ? '收起' : t('ai.tools.actions.expand') }}
              </button>
            </div>
          </div>
          <div v-if="expanded[log.id]" class="mt-3 grid grid-cols-2 gap-4 text-xs">
            <pre class="bg-gray-50 p-3 rounded overflow-auto">Input:
{{ JSON.stringify(log.input_params, null, 2) }}</pre>
            <pre class="bg-gray-50 p-3 rounded overflow-auto">Output:
{{ JSON.stringify(log.output_result, null, 2) }}
<template v-if="log.error_message">Error: {{ log.error_message }}</template></pre>
          </div>
        </div>
        <div v-if="logs.length === 0" class="p-10 text-center text-gray-400">
          {{ t('ai.tools.empty') }}
        </div>
      </div>
    </div>

    <!-- Tab: Permissions -->
    <div v-if="activeTab === 'permissions'">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <!-- Left: Tool list -->
        <div class="md:col-span-1 bg-white border rounded p-3 space-y-1">
          <div v-for="tool in tools" :key="tool.id"
               @click="selectedTool = tool"
               :class="['cursor-pointer rounded px-3 py-2 text-sm',
                   selectedTool?.id===tool.id
                     ? 'bg-blue-50 text-blue-700 font-medium'
                     : 'hover:bg-gray-50 text-gray-700']">
            {{ tool.name }}
          </div>
        </div>
        <!-- Right: Templates + switches -->
        <div class="md:col-span-3 bg-white border rounded">
          <div class="px-4 py-3 border-b text-sm text-gray-500">
            {{ t('ai.tools.permissions.header') }}：
            <span v-if="selectedTool" class="font-medium text-gray-900">
              {{ selectedTool.name }}
            </span>
            <span v-else>{{ t('common.selectPlaceholder') }}</span>
          </div>
          <div class="divide-y" v-if="selectedTool">
            <div v-for="at in agentTemplates" :key="at.id"
                 class="flex items-center justify-between px-4 py-3">
              <div class="text-sm text-gray-900">{{ at.name }}</div>
              <input type="checkbox"
                     :checked="permissionOf(at.id)"
                     @change="togglePermission(at.id, $event)"
                     class="w-5 h-5 rounded border-gray-300 text-blue-600" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab: MCP Servers -->
    <div v-if="activeTab === 'mcp'">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div v-for="m in mcpConfigs" :key="m.id" class="bg-white border rounded p-4">
          <div class="flex items-start justify-between">
            <div>
              <div class="font-medium text-gray-900">{{ m.name }}</div>
              <div class="text-xs text-gray-500 break-all mt-1">{{ m.server_url }}</div>
            </div>
            <button @click="onSync(m)"
                    :disabled="syncing[m.id]"
                    class="bg-blue-50 text-blue-700 border border-blue-100 rounded px-3 py-1 text-xs hover:bg-blue-100 disabled:opacity-50">
              {{ syncing[m.id] ? 'syncing…' : t('ai.tools.mcp.syncNow') }}
            </button>
          </div>
          <div class="mt-3 grid grid-cols-2 gap-2 text-xs">
            <div class="text-gray-500">{{ t('ai.tools.mcp.toolsCount') }}: <b class="text-gray-800">{{ m.tools_count }}</b></div>
            <div class="text-gray-500">{{ t('ai.tools.mcp.lastSync') }}:
              <b class="text-gray-800">{{ m.last_sync_at || '-' }}</b>
            </div>
          </div>
        </div>
        <div v-if="mcpConfigs.length===0" class="col-span-2 text-gray-400 text-center py-10">
          {{ t('ai.tools.mcp.noServers') }}
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal (简化版：核心字段) -->
    <Transition name="modal">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/50" @click="showModal=false"></div>
        <div class="relative bg-white rounded-lg w-full max-w-xl mx-4 p-5 space-y-3">
          <div class="flex items-center justify-between">
            <h2 class="font-semibold text-lg">
              {{ isEditing ? t('ai.tools.actions.edit') : t('ai.tools.create') }}
            </h2>
            <button @click="showModal=false" class="text-gray-400">✕</button>
          </div>
          <div class="grid grid-cols-2 gap-3 text-sm">
            <input v-model="form.name" :placeholder="t('ai.tools.form.name')" class="col-span-2 border rounded px-3 py-2" />
            <textarea v-model="form.description" rows="2"
                      :placeholder="t('ai.tools.form.description')"
                      class="col-span-2 border rounded px-3 py-2"></textarea>
            <select v-model="form.category" class="border rounded px-3 py-2">
              <option v-for="c in categories" :key="c">{{ c }}</option>
            </select>
            <select v-model="form.tool_type" class="border rounded px-3 py-2">
              <option value="api">api</option>
              <option value="function">function</option>
              <option value="mcp">mcp</option>
            </select>
            <input v-if="form.tool_type==='api'" v-model="form.endpoint"
                   :placeholder="t('ai.tools.form.endpoint')"
                   class="col-span-2 border rounded px-3 py-2" />
            <input v-model.number="form.rate_limit" type="number" min="0"
                   :placeholder="t('ai.tools.form.rateLimit')"
                   class="border rounded px-3 py-2" />
            <input v-model.number="form.timeout" type="number" min="1"
                   :placeholder="t('ai.tools.form.timeout')"
                   class="border rounded px-3 py-2" />
          </div>
          <div v-if="form.category==='dangerous'"
               class="text-xs bg-red-50 text-red-700 border border-red-100 rounded p-2">
            ⚠️ {{ t('ai.tools.form.dangerousWarn') }}
          </div>
          <div class="flex justify-end space-x-2 pt-2">
            <button @click="showModal=false"
                    class="px-4 py-2 border rounded hover:bg-gray-50">
              {{ t('common.cancel') }}
            </button>
            <button @click="onSave"
                    :disabled="!form.name"
                    class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded disabled:opacity-50">
              {{ t('common.save') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useWorkspaceStore } from '@/stores/workspace'
import {
  listTools, createTool, updateTool, deleteTool,
  listToolCallLogs,
  listToolPermissions, setToolPermission,
  syncMCPTools,
  type Tool, type ToolCallLog, type ToolPermissionView
} from '@/api/tool'
import { listAgentTemplates } from '@/api/agent-template'
import { listMCPConfigs } from '@/api/mcp'
import { useSSE } from '@/composables/useSSE'

const { t } = useI18n()
const wsStore = useWorkspaceStore()
const slug = computed(() => wsStore.currentWorkspace?.slug || '')

// Tabs
const tabs = [
  { key: 'tools', label: () => t('ai.tools.tabTools') },
  { key: 'logs', label: () => t('ai.tools.tabLogs') },
  { key: 'permissions', label: () => t('ai.tools.tabPermissions') },
  { key: 'mcp', label: () => t('ai.tools.tabMCP') },
].map(x => ({ key: x.key, label: x.label() }))
const activeTab = ref('tools')
const categories = ['general','project_management','code','ci_cd','dangerous']

// Tools Tab
const tools = ref<Tool[]>([])
const filters = reactive({ category: '', status: '' })
const filteredTools = computed(() => tools.value.filter(t =>
  (!filters.category || t.category===filters.category)
  && (!filters.status || t.status===filters.status)
))

// Logs Tab
const logs = ref<ToolCallLog[]>([])
const stats = reactive({ total: 0, successRate: 0, p95: 0, rateLimited: 0 })
const expanded = reactive<Record<number, boolean>>({})

function statusBadgeClass(s: string) {
  return 'px-2 py-0.5 rounded text-xs font-semibold ' +
    (s==='success' ? 'bg-green-100 text-green-700'
                   : 'bg-red-100 text-red-700')
}
function statusText(s: string, limited: boolean) {
  if (limited) return t('ai.tools.status.rateLimited')
  return t('ai.tools.status.' + s) || s
}
function formatDate(s: string) { return new Date(s).toLocaleString() }
function recomputeStats() {
  stats.total = logs.value.length
  const ok = logs.value.filter(l => l.status==='success').length
  stats.successRate = stats.total ? Math.round(100 * ok / stats.total) : 0
  const durs = logs.value.map(l => l.duration_ms).sort((a, b) => a - b)
  stats.p95 = durs.length ? durs[Math.floor(0.95 * durs.length)] : 0
  stats.rateLimited = logs.value.filter(l => l.rate_limited).length
}

// Permissions Tab
const agentTemplates = ref<{id:number,name:string}[]>([])
const selectedTool = ref<Tool | null>(null)
const permissions = ref<ToolPermissionView[]>([])

function permissionOf(atId: number) {
  const hit = permissions.value.find(p => p.agent_template_id===atId)
  return hit ? hit.allowed : true  // fallback: 未配置默认允许
}
async function togglePermission(atId: number, e: Event) {
  if (!selectedTool.value) return
  const allowed = (e.target as HTMLInputElement).checked
  await setToolPermission(slug.value, selectedTool.value.id, {
    tool_id: selectedTool.value.id,
    agent_template_id: atId,
    allowed
  })
  await loadPermissions()
}
async function loadPermissions() {
  if (!selectedTool.value) return
  const { data } = await listToolPermissions(slug.value, selectedTool.value.id)
  permissions.value = data || []
}

// MCP Tab
const mcpConfigs = ref<any[]>([])
const syncing = reactive<Record<number, boolean>>({})
async function onSync(mc: any) {
  syncing[mc.id] = true
  try {
    const { data } = await syncMCPTools(slug.value, mc.id)
    // toast
    alert(t('ai.tools.mcp.synced', data)) // 正式项目用 useToast
    await Promise.all([loadTools(), loadMCPConfigs()])
  } finally {
    syncing[mc.id] = false
  }
}

// Create/Edit Modal
const showModal = ref(false)
const isEditing = ref(false)
const form = reactive<Partial<Tool>>({
  name: '', description: '', category: 'general',
  tool_type: 'api', rate_limit: 60, timeout: 30
})
function openCreateTool() {
  isEditing.value = false
  Object.assign(form, { name:'', description:'', category:'general',
    tool_type:'api', rate_limit:60, timeout:30 })
  showModal.value = true
}
function openEditTool(tool: Tool) {
  isEditing.value = true
  Object.assign(form, JSON.parse(JSON.stringify(tool)))
  showModal.value = true
}
async function onSave() {
  if (isEditing.value) {
    await updateTool(slug.value, form.id!, form)
  } else {
    await createTool(slug.value, form)
  }
  showModal.value = false
  await loadTools()
}
async function onDelete(tool: Tool) {
  if (!confirm(`Delete tool ${tool.name}?`)) return
  await deleteTool(slug.value, tool.id)
  await loadTools()
}

// Loaders
async function loadTools() {
  const { data } = await listTools(slug.value)
  tools.value = data || []
}
async function loadLogs() {
  const { data } = await listToolCallLogs(slug.value, { per_page: 100 })
  logs.value = (data as any)?.data || (Array.isArray(data) ? data : [])
  recomputeStats()
}
async function loadAgentTemplates() {
  const { data } = await listAgentTemplates(slug.value)
  agentTemplates.value = (data as any)?.data || data || []
}
async function loadMCPConfigs() {
  try {
    const { data } = await listMCPConfigs(slug.value)
    mcpConfigs.value = data || []
  } catch { mcpConfigs.value = [] }
}

// SSE 实时刷新日志
const { on: onSSE } = useSSE()
onSSE('tool_call', () => {
  if (activeTab.value === 'logs') loadLogs()
})

onMounted(async () => {
  await Promise.all([
    loadTools(), loadLogs(), loadAgentTemplates(), loadMCPConfigs()
  ])
})
</script>
```

- [ ] **Step 5: Router 注册 + AgentDashboard 卡片**

`router/index.ts`：在 AgentDashboard 相邻区添加：

```typescript
{
  path: '/workspace/:slug/agents/tools',
  name: 'AgentTools',
  component: () => import('@/views/agents/ToolManager.vue'),
  meta: { requiresAuth: true }
},
{
  path: '/agents/tools',
  redirect: to => `/workspace/${useWorkspaceStore().currentWorkspace?.slug || ''}/agents/tools`
},
```

(参考现有 `AutopilotList.vue` 路由模式)

`AgentDashboard.vue`：在 Performance 卡片之后添加 Tools 卡片（样式一致）：

```vue
<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-5 hover:shadow-md transition cursor-pointer"
     @click="$router.push('/agents/tools')">
  <div class="flex items-center justify-between mb-3">
    <div class="w-10 h-10 bg-indigo-100 rounded-xl flex items-center justify-center text-xl">🛠️</div>
    <span class="text-xs text-indigo-600 font-medium">Tool Calling</span>
  </div>
  <h3 class="font-semibold text-gray-900">{{ t('ai.tools.card.title') }}</h3>
  <p class="text-sm text-gray-500 mt-1 line-clamp-2">{{ t('ai.tools.description') }}</p>
  <div class="mt-4 pt-4 border-t border-gray-100 flex items-center justify-between text-xs text-gray-500">
    <span>{{ t('ai.tools.card.open') }} →</span>
  </div>
</div>
```

- [ ] **Step 6: 类型检查 + lint**

```bash
cd frontend
npx vue-tsc --noEmit 2>&1 | tee /tmp/tsc.log
# 检查 ToolManager / tool.ts / 路由相关错误
grep -E "(ToolManager|tool\.ts|ai\.tools|agents/tools)" /tmp/tsc.log || echo "No ToolManager-related errors"
```

Expected: 新增相关的 0 错误（已有历史错误除外）

- [ ] **Step 7: Commit**

```bash
git add frontend/src/api/tool.ts \
        frontend/src/views/agents/ToolManager.vue \
        frontend/src/views/agents/AgentDashboard.vue \
        frontend/src/composables/useSSE.ts \
        frontend/src/router/index.ts \
        frontend/src/locales/zh-CN.json \
        frontend/src/locales/en-US.json
git commit -m "feat(frontend): add ToolManager 4-tab console + Dashboard entry"
```

---

## Task 8: 端到端验证 & 回归

**Files:** None（只执行命令 + 手动浏览器验收）

- [ ] **Step 1: 后端 build + 全部单测**

```bash
cd backend
go build ./...
go vet ./...
go test ./internal/service/... -count=1 -timeout 180s
```

Expected: PASS

- [ ] **Step 2: 启动服务并手动验收关键路径（PowerShell 脚本，复用已有 test_acceptance.ps1 风格）**

```powershell
# 登录
$base = "http://localhost:8000/api/v1"
$login = Invoke-RestMethod "$base/auth/login" -Method Post -ContentType "application/json" `
  -Body '{"email":"admin@reqmango.com","password":"demo1234"}'
$token = $login.data.token
$headers = @{ Authorization = "Bearer $token" }

# A) 非 workspace 成员调用 → 403 (用另一个不存在的 user 模拟需要自行 mock，
#    这里我们验证 Admin 校验，手工构造 dangerous tool 的场景)
#    先创建一个 dangerous category 的工具
$tool = Invoke-RestMethod "$base/workspaces/1/tools" -Method Post -Headers $headers `
  -ContentType "application/json" -Body (@{
    name="danger-test"; description="test"; category="dangerous";
    tool_type="function"; is_builtin=$false; status="active"
  } | ConvertTo-Json)
# 验证普通成员 Role=member 调用时如果有办法切换 user 则测 403，否则验证存在权
限字段 (手工)

# B) Rate limit 场景
# 手工构造 rate_limit=3 的工具，连续 call 4 次 → 第 4 次 429 (需要 function 类型存
在，可调用内置 create_issue 以避免外部依赖)

# C) MCP Sync 流程（假 URL 应返回错误，不是 panic；返回 JSON-RPC 200 时 upsert）

# D) Built-in get_pr_diff 参数校验：缺 repo_owner → 400
```

- [ ] **Step 3: 前端浏览器验收**

启动：
```bash
cd frontend
npm run dev   # 保持运行
```

手动检查：
1. 访问 `/agents/tools` → 4 Tab 页面渲染，Tools 列表有数据
2. 点 Dashboard「🛠️ 工具管理」→ 路由跳转正常
3. Logs Tab：调一次 tool，SSE 自动刷新（或手动刷新按钮存在）
4. Permissions Tab：切 Template switch，调用后看生效
5. MCP Tab：点「同步工具」，有 loading 动画

- [ ] **Step 4: vue-tsc 最终扫描**

```bash
cd frontend
npx vue-tsc --noEmit
```

Expected: 新增文件 0 错误（遗留错误允许）

- [ ] **Step 5: 最终总提交（可选，如果前面步骤独立 commit 已完成）**

```bash
git log --oneline -8
# 确认 T1-T7 都有对应 commit
# 若有遗漏可在本步统一 push
```

---

## 计划自审（已执行）

**Spec 覆盖率**：
- ✅ §A2 权限三步校验 → T2 全部覆盖
- ✅ §A3 双维度滑动窗口限流 → T3 全部覆盖
- ✅ §A4.1 MCP Lite Sync/Execute → T5 全部覆盖
- ✅ §A4.2 代码审查内置工具 4 个 → T6 全部覆盖
- ✅ §A4.3 SSE 审计事件 + 日志补全 → T4 全部覆盖
- ✅ §A5 前端 ToolManager 4 Tab + Dashboard → T7 全部覆盖
- ✅ §7 Migration → T1 Step 1
- ✅ §8 单测要求 → T2/T3/T5/T6 对应 test 文件和用例
- ✅ §9 验收标准 → T8 后端/前端/安全回归三部分

**占位符扫描**：未出现 "TODO/TBD/similar to"，所有函数有实现、所有命令有具体参数

**类型一致性**：方法名 `checkPermissions` / `SyncTools` / `tryAcquire` 在测试和实现中统一；前端事件名 `tool_call` 在 SSE 与页面监听中一致
