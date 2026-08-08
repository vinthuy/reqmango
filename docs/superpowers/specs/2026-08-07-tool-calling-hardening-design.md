# 方向 A：Tool Calling 引擎加固 设计文档

> **日期**: 2026-08-07  
> **状态**: 待用户审核  
> **对应 PRD**: §6 Tool Calling 引擎

---

## 1. 背景与目标

### 1.1 现有缺口（基线）
基于代码审查，Tool Calling 引擎存在以下 5 个关键缺口：

| # | 缺口 | 风险 | 对应 PRD |
|---|------|------|----------|
| 1 | Tool 权限未校验（Call 方法无 workspace/permission 检查） | 越权调用，违反 §17.2 安全要求 | §6.2 Tool 结构 permission 字段 |
| 2 | checkRateLimit 实现过于简单（无 caller 维度，无真正滑动窗口） | 单 Agent 耗尽配额，多实例下计数失真 | §17.1 性能 + §6.1 RateLimit |
| 3 | MCP Server 连接/同步/工具执行链路未打通 | 无法复用 MCP 生态，PRD P2 §6.2 落空 | §6 Tool Calling / §9 MCP |
| 4 | Built-in Function 只有 3 个 issue 相关，缺代码审查工具 | Code Reviewer / Tester Agent 无法执行 | §3.8 代码审查阶段 |
| 5 | 无 Tool 前端控制台，用户看不到配置/日志/权限 | 功能不可操作，无法验收闭环 | §2.3 Tool 管理页面 |

### 1.2 加固目标
- **安全**：任何 Tool.Call 必须经过「成员身份 → 高危操作 Admin → ToolPermission 白/黑」三层校验
- **公平**：双维度滑动窗口限流，防止单 Agent 独占配额
- **扩展**：MCP Lite 打通工具发现（Sync）→ 注册 → 调用链路
- **可用**：代码审查内置函数扩展 4 个，Reviewer Agent 可独立完成 PR 审批
- **闭环**：前端 ToolManager.vue 四 Tab 控制台 + Dashboard 入口

---

## 2. 架构设计

```
┌───────────────────────────────────────────────────────────────────────┐
│                    Tool Calling Engine (加固后)                        │
├───────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  Tool.Call(wid, req)                                                  │
│       │                                                               │
│       ▼                                                               │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌──────────┐  │
│  │ A2 权限层    │ → │ A3 限流层    │ → │ A4 执行层    │ → │ A4 审计层│  │
│  │ • workspace  │   │ • 全局滑动   │   │ • api       │   │ • 写日志 │  │
│  │ • admin高危  │   │ • caller 20%│   │ • function  │   │ • SSE广播│  │
│  │ • whitelist  │   │ • 429错误    │   │ • mcp lite  │   │          │  │
│  └─────────────┘   └─────────────┘   └─────────────┘   └──────────┘  │
│                                                                       │
└───────────────────────────────────────────────────────────────────────┘
```

---

## 3. A2 权限层设计（最小权限模型）

### 3.1 三步校验流程

```
CallToolRequest {
    tool_id, input_params,
    + agent_template_id?   // 用于 ToolPermission 匹配
    + caller_user_id       // 用于 workspace 成员校验
    + agent_task_id?       // 审计日志关联任务
}
        │
        ▼
  Step 1: Workspace 成员身份
    └─ SELECT * FROM workspace_members
       WHERE workspace_id = ? AND user_id = ? AND is_active = true
    └─ 不存在 → 403 Forbidden("Workspace member required")
        │
        ▼
  Step 2: 高危操作 Admin 校验（可选）
    └─ tool.Category == "dangerous"
    └─ member.Role < RoleAdmin → 403 Forbidden("Admin required for dangerous tools")
        │
        ▼
  Step 3: ToolPermission 白/黑名单
    └─ SELECT * FROM tool_permissions
       WHERE tool_id = ? AND (agent_template_id = ? OR agent_template_id IS NULL)
       ORDER BY agent_template_id DESC NULLS LAST  (先精确匹配，再通配)
    └─ 命中 allowed=false → 403 Forbidden("Tool permission denied")
    └─ 无记录 → fallback: 允许（兼容旧数据，避免破坏性变更）
```

### 3.2 新增字段

| 模型 | 字段 | 类型 | 说明 |
|------|------|------|------|
| request.CallToolRequest | agent_template_id | `*uint64` | 调用方 Agent 模板（或 nil 为匿名用户调用） |
| request.CallToolRequest | caller_user_id | `uint64` | 实际发起用户（审计+权限用） |
| request.CallToolRequest | agent_task_id | `*uint64` | 关联任务 ID（补全日志） |

---

## 4. A3 限流层设计（滑动窗口 + 双维度）

### 4.1 数据结构

```go
type rateLimiter struct {
    mu    sync.RWMutex
    store sync.Map  // key → *rateLimitEntry
}

type rateLimitEntry struct {
    mu    sync.Mutex
    calls []time.Time  // 仅保留最近 60s
}

var globalRateLimiter = &rateLimiter{}
```

Key 命名约定：
- `"g:<toolID>"`  → 全局限流
- `"c:<toolID>:<callerID>"` → 单调用者限流（callerID 格式：`u:<uid>` 或 `a:<agentTplID>`）

### 4.2 算法

```
checkRateLimit(tool, callerID):
  1. 如果 tool.RateLimit == 0，返回 nil（无限额）
  2. 窗口大小：60 * time.Second
  3. 对两个 key 分别执行：
     a) entry.calls = filter(entry.calls, t → now - t <= 窗口)  // 懒清理
     b) if len(entry.calls) >= threshold:
          return 429 + log.RateLimited = true
  4. threshold 计算：
     - 全局 key: tool.RateLimit
     - 调用者 key: ceil(tool.RateLimit / 5)  // 单调用者最多 20%
  5. 通过: entry.calls = append(calls, now)
  6. 广播 SSE: tool_call.rate_limited （如超限）
```

### 4.3 性能考虑
- 懒清理 O(N)，N = RateLimit（通常 ≤ 600，对 Go slice 完全可接受）
- 单 entry mutex，不是全局锁，并发友好

---

## 5. A4 执行层 + 审计层设计

### 5.1 MCP Lite

**数据模型新增字段**：

| 模型 | 字段 | 类型 | 说明 |
|------|------|------|------|
| model.Tool | MCPConfigID | `*uint64` FK | 所属 MCP Server（MCP 类型 tool 必填） |

**新增方法**：

```go
// mcp_service.go
func (s *MCPService) SyncTools(workspaceID, configID, callerID uint64) (int, error)
  → 调用 MCP Server JSON-RPC tools/list
  → 循环响应，按 (mcp_config_id, name) 唯一键 upsert 到 tools 表
  → 返回 new + updated 数量

func (s *MCPService) ExecuteTool(workspaceID, configID uint64, toolName string, args map[string]any) (json.RawMessage, error)
  → HTTP POST JSON-RPC:
    { "jsonrpc": "2.0", "method": "tools/call",
      "params": {"name": toolName, "arguments": args}, "id": 1 }

// tool_service.go
func (s *ToolService) executeMCPTool(tool model.Tool, params json.RawMessage) (json.RawMessage, error)
  → 校验 tool.MCPConfigID != nil
  → 调 MCPService.ExecuteTool
```

**新增路由**（router.go）：
```
POST /api/v1/workspaces/:wid/mcp-configs/:id/sync → SyncToolsHandler
```

### 5.2 Built-in Function 扩展（代码审查）

| 函数名 | 参数 | 调用的 GitHub API |
|--------|------|-------------------|
| `get_pr_diff` | `{repo_owner, repo_name, pr_number}` | `GET /repos/{owner}/{repo}/pulls/{n}` + `Accept: application/vnd.github.diff` |
| `add_review_comment` | `{repo_owner, repo_name, pr_number, body, commit_id?, path?, line?}` | `POST /repos/{owner}/{repo}/pulls/{n}/comments` |
| `list_pr_commits` | `{repo_owner, repo_name, pr_number}` | `GET /repos/{owner}/{repo}/pulls/{n}/commits` |
| `create_pr_review` | `{repo_owner, repo_name, pr_number, event: APPROVE\|REQUEST_CHANGES\|COMMENT, body, comments?}` | `POST /repos/{owner}/{repo}/pulls/{n}/reviews` |

均复用 `github_service.go` 的 token 管理逻辑。

### 5.3 审计层增强

**SSE 广播（Call 完成后）**：
| 事件名 | payload | 触发时机 |
|--------|---------|----------|
| `tool_call.completed` | `{id, tool_id, tool_name, status, duration_ms}` | 成功/失败后统一 |
| `tool_call.failed` | `{id, tool_id, error_message}` | status=failed |
| `tool_call.rate_limited` | `{tool_id, retry_after_sec: 60}` | 限流拦截 |

**GetCallLogs 接口增强**：
- query 支持：`status`, `tool_id`, `agent_id`, `from_time`, `to_time`
- 复用 `common.PaginationParams` 分页
- Response 不变

---

## 6. A5 前端 Tool 控制台（ToolManager.vue）

### 6.1 页面结构

```
views/agents/ToolManager.vue
├── TopBar (标题 + 创建按钮 + MCP Sync 下拉)
├── Tabs (4)
│   ├── Tab Tools
│   │   ├── FilterBar: Category / Status / ToolType
│   │   ├── DataTable (列: Name, Type badge, Category, Status badge,
│   │   │               RateLimit, Timeout, Actions)
│   │   └── Modal: CRUD 表单 (JSON Schema 编辑器可简化为 textarea)
│   │
│   ├── Tab Logs
│   │   ├── StatCards: TotalCalls / SuccessRate / P95Duration / RateLimited
│   │   ├── FilterBar: Tool / Status / Caller / DateRange
│   │   └── Timeline 列表，展开查看 Input/Output JSON
│   │
│   ├── Tab Permissions
│   │   ├── Left: Tool 列表 (选中高亮)
│   │   └── Right: AgentTemplate 列表，每行 switch=Allowed
│   │
│   └── Tab MCP Servers
│       ├── Card: Name, URL, Status(连接状态), ToolsCount, LastSyncAt
│       └── Actions: [同步工具] (loading 动画 + Toast 结果)
│
└── AgentDashboard.vue 新增 🛠️ Tools 卡片
    └── → /agents/tools 路由
```

### 6.2 新增文件清单

| 路径 | 说明 |
|------|------|
| `frontend/src/views/agents/ToolManager.vue` | 主页面（4 Tab） |
| `frontend/src/api/tool.ts` | 扩展：SyncMCP, GetCallLogs(filtered), GetPermissions, SetPermissions |
| `frontend/src/composables/useSSE.ts` | + tool_call.* 4 个事件监听 |
| `frontend/src/locales/zh-CN.json` | + ai.tools.* (≈25 keys) |
| `frontend/src/locales/en-US.json` | + ai.tools.* 对应英文 |
| `frontend/src/router/index.ts` | + 路由 `/agents/tools` & mirror |
| `frontend/src/views/agents/AgentDashboard.vue` | + Tools 卡片 |

### 6.3 设计复用参考
- 表格+Modal 模式 → SquadList.vue §L13-L116
- Tab+Cards 模式 → CICDManager.vue（Config/Builds 切换）
- Dashboard 卡片 → AgentDashboard.vue §L175-L188 Performance 卡片风格

---

## 7. 数据库变更（Migration）

```sql
-- === 新建 migration: 000022_tool_hardening.up.sql ===

-- 1. Tool 新增 MCPConfigID 外键
ALTER TABLE tools ADD COLUMN mcp_config_id BIGINT REFERENCES mcp_configs(id) ON DELETE SET NULL;
CREATE INDEX idx_tools_mcp_config_id ON tools(mcp_config_id);
CREATE UNIQUE INDEX idx_tools_mcp_uniq ON tools(mcp_config_id, name) WHERE mcp_config_id IS NOT NULL;

-- 2. tools.category 扩展 dangerous 枚举 (CHECK 可选，默认用应用层约束)
ALTER TABLE tools ALTER COLUMN category SET DEFAULT 'general';

-- 3. tool_call_logs 补全 caller 维度索引
CREATE INDEX idx_tool_call_logs_agent_id ON tool_call_logs(agent_id);
CREATE INDEX idx_tool_call_logs_created_at ON tool_call_logs(created_at DESC);
```

---

## 8. 单元测试要求

| 测试文件 | 测试目标 | 用例数 (估) |
|----------|----------|--------------|
| `tool_service_test.go` (新增) | rateLimiter 滑动窗口 + 双维度限流 | 8+ |
| 同上 | 权限层三步校验（成员/Admin/白名单） | 6+ |
| 同上 | MCP SyncTools upsert 幂等性 | 3+ |
| 同上 | Built-in 4 个审查函数参数校验 | 4+ |
| `mcp_service_test.go` (补) | ExecuteTool JSON-RPC 请求构造 | 2+ |

---

## 9. 验收标准

### 9.1 后端
- `go build ./...` 通过
- `go vet ./...` 无告警
- `go test ./internal/service/ -run "Tool|MCP"` 全绿
- Smoke 脚本验证：
  - 非 workspace 成员 Call → 403
  - Admin Call dangerous tool → 200；非 Admin → 403
  - 连续 10 次 Call rate_limit=5 → 第 6 次 429
  - MCP Sync → tools 表新增 N 条记录，MCPConfigID 正确
  - Call Type=mcp Tool → JSON-RPC 请求正确发出

### 9.2 前端
- `vue-tsc --noEmit` 无 ToolManager 相关错误
- 浏览器访问 `/agents/tools` → 4 Tab 渲染正常
- Logs Tab 点时间 → 展开 Input/Output JSON
- Permissions Tab 切 switch → 白名单生效（Call 403）
- Dashboard Tools 卡片 → 跳转正常

### 9.3 安全回归
- AgentTask 正常 workflow（enqueue→claim→call→complete）不受影响
- 现有 Built-in (create_issue 等) 行为未变
- 无 SQL 注入、无越权（测试脚本交叉 workspace_id 验证）
