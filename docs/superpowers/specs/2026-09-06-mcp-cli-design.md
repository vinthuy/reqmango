# MCP Server 重写 + CLI 设计

- 日期：2026-09-06
- 分支：feat/mcp-cli
- 状态：设计已确认，待实施

## 1. 背景与目标

reqmango 目前有一个 6 月底实现的独立 `mcp-server/`（Go、stdio、14 tools、静态 JWT token），此后后端 API 经历了 workflow 重写、RQL 落地、agent 平台合并、chat/messages 等大量变更，该 MCP server 已过时。同时没有面向日常用户的 CLI。

**目标**：

1. 重写独立 MCP server：适配当前 API，扩充工具覆盖（核心 CRUD + AI 能力），支持 stdio + HTTP 双传输，供 Claude Code / Claude Desktop / Cursor 使用。
2. 新建 `reqmango` CLI：面向日常用户的操作用 CLI（类似 gh），覆盖 issue/cycle/project 的增删查改、RQL 查询、agent 触发等。
3. 后端新增 PAT（Personal Access Token）系统，为 CLI 与 MCP 提供长期有效的认证凭据（原 catchup roadmap 中"MCP Server P0.2 PAT Token 认证模式"的落地）。

**非目标**（YAGNI）：

- PAT scopes（首版全权限，表结构预留字段）
- MCP OAuth 授权流程（PAT 覆盖 CI/CD 场景）
- 对外发布 Go SDK（client 包先留在 tools module 内，将来拆分是机械操作）
- Python / npm SDK（catchup roadmap P0.3，独立立项）
- MCP 工具覆盖 SavedView（对 AI 来说直接给 RQL 更灵活）
- 后端 MCP 客户端（MCP Lite，已于 2026-08-07 完成，本次不涉及）

## 2. 架构总览

```
                    ┌───────────────┐
  Claude Code ─────▶│  reqmango-mcp │──┐ (stdio)
  Claude Desktop ──▶│  (mcp-go)     │  │
  HTTP 客户端 ──────▶│  :8080 /mcp   │  │
                    └───────────────┘  │
 终端用户 ─────────▶┌───────────────┐  │
                    │   reqmango    │──┤ (cobra)
                    │   CLI         │  │
                    └───────────────┘  │
                                       ▼
                    ┌──────────────────────────┐
                    │ sdk/client  (共享 API 客户端) │
                    └──────────────────────────┘
                                       │ HTTP + PAT
                                       ▼
                    ┌──────────────────────────┐
                    │ backend  (Gin, 新增 PAT 认证) │
                    └──────────────────────────┘
```

CLI 与 MCP server 是两个薄壳，全部业务逻辑在 `client` 包；`client` 仅依赖标准库（http/json），认证、分页、错误映射、类型定义各一份。

**关键决策**（已确认）：

| 决策点 | 选择 | 理由 |
|---|---|---|
| 仓库结构 | 单一新 Go module，三包两二进制 | 无 replace 指令，最简单；发布 SDK 时再拆 |
| MCP 协议实现 | mark3labs/mcp-go | stdio/SSE/streamable HTTP 开箱即用，免维护 JSON-RPC 细节 |
| CLI 框架 | cobra | Go 标准选择，嵌套命令 + flag 组合自然 |
| 认证 | 后端新增 PAT 系统 | 长期有效可吊销，MCP 配置一次不用管；符合原 roadmap |

## 3. 模块结构

```
sdk/                          ← 新 Go module：github.com/reqmango/tools
  go.mod
  client/                     ← API 客户端 + 全部 DTO 类型 + 错误映射
    client.go                 ← 核心 do() 函数、认证头、超时、分页
    errors.go                 ← APIError 类型化错误
    auth.go                   ← login / PAT 管理
    workspaces.go projects.go
    issues.go cycles.go
    chat.go agents.go meta.go
  mcp/                        ← MCP server 组装
    server.go                 ← mcp-go 集成、stdio/HTTP 双传输
    tools_core.go tools_ai.go ← 24 个 tools 定义 + handler
  cli/                        ← cobra 命令树
    root.go                   ← 全局 flag、config 加载
    config.go                 ← ~/.reqmango/config.json
    output.go                 ← table / json 格式化
    cmd_auth.go cmd_issue.go cmd_project.go cmd_cycle.go
    cmd_agent.go cmd_meta.go cmd_ask.go
  cmd/reqmango-mcp/main.go    ← 二进制 1
  cmd/reqmango/main.go        ← 二进制 2
```

- 旧 `mcp-server/` 目录整体删除（git 历史可查），其 README 内容吸收进 sdk。
- 根目录 `Makefile` 增加 `make tools`（构建两个二进制到 `bin/`）、`make test-tools`。
- 环境变量统一：`REQMANGO_API_URL`（默认 `http://localhost:8000/api/v1`）、`REQMANGO_PAT`。

## 4. 后端 PAT 系统

### 4.1 数据模型

新表 `personal_access_tokens`：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | |
| user_id | BIGINT FK | 归属用户 |
| name | varchar(100) | 展示名（如 "cli-macbook"） |
| token_prefix | varchar(20) | 展示用前缀 `reqmango_pat_ab3d` |
| token_hash | varchar(64) | SHA-256 hex，明文 token 仅创建时返回一次 |
| scopes | text | 预留，首版空（全权限） |
| last_used_at | timestamptz null | 每次认证刷新 |
| expires_at | timestamptz null | null = 永不过期 |
| revoked_at | timestamptz null | 吊销时间 |
| created_at | timestamptz | |

### 4.2 端点与中间件

- `GET  /api/v1/auth/tokens` —— 当前用户 PAT 列表（不含 hash、不含明文）
- `POST /api/v1/auth/tokens` —— 创建（body: name, expires_at?），响应含一次性明文 token
- `DELETE /api/v1/auth/tokens/:id` —— 吊销
- `AuthMiddleware`：识别 `Authorization: Bearer reqmango_pat_*` → 查 token_hash → 未吊销且未过期则按对应用户放行并刷新 last_used_at；其余走现有 JWT 逻辑。权限校验链路（RequirePermission 等）与 JWT 用户完全一致。

CLI 登录流：`reqmango login`（用户名+密码 → 换 JWT → 用 JWT 创建 PAT → 存 `~/.reqmango/config.json`，JWT 即弃）。

## 5. MCP Server

### 5.1 传输

- **stdio**（默认）：Claude Desktop / Claude Code / Cursor 本地场景
- **streamable HTTP**：`reqmango-mcp --http :8080` 监听 `/mcp`，`Authorization: Bearer <PAT>` 认证，供远程/CI 场景

### 5.2 工具集（24 个）

**核心 CRUD（19 个）**——旧 14 个适配新 API + 新增 5 个：

| 工具 | 说明 |
|---|---|
| `list_workspaces` / `list_projects` / `get_project` | 入口工具，AI 先发现 ID |
| `create_issue` | 含 state/priority/type/assignee/labels/parent 等当前 Issue 模型字段 |
| `list_issues` | `rql` 参数 + 常用 filters + 分页 |
| `get_issue` | 含评论、relations |
| `update_issue` | 状态流转走现有 API 校验（复用后端规则） |
| `search_issues` | 全文搜索（`/issues/search`） |
| `list_cycles` / `get_cycle_progress` / `add_issue_to_cycle` | cycle 三件套 |
| `list_members` / `get_states` / `get_labels` / `list_issue_types` | 元数据 |
| `add_comment` | issue 评论 |
| `list_notifications` | 当前用户通知 |
| `list_pages` / `get_page` | 项目文档 |

**AI 能力（5 个）**：

| 工具 | 说明 |
|---|---|
| `ai_search` | AI 语义搜索（`POST /projects/:id/ai/search`） |
| `ai_chat` | 单轮对话：无 chatId 则按 issue（`GET /issues/:id/chat`）或 `POST /projects/:id/ai/chat`，SSE 流在 server 侧聚合后一次性返回 |
| `list_agents` / `dispatch_agent` | 列出 workspace 下 agent、按 agentId+task 触发（`POST /agents/:id/dispatch`） |
| `get_agent_task` | 查询触发后的任务状态/日志 |

## 6. CLI

```
reqmango auth login|logout|status|revoke
reqmango workspace list|switch
reqmango project list|show <id|identifier>|create
reqmango issue list   [--project] [--assignee me] [--state] [--type] [--rql "..."] [--limit] [--json]
reqmango issue show   <id|code> [--comments]
reqmango issue create --project X --title "..." [--desc] [--type] [--priority] [--state] [--assignee] [--labels a,b] [--parent]
reqmango issue update <id|code> [--state|--assignee|--priority|--title|--labels ...]
reqmango issue search <query>
reqmango cycle list|progress|burndown
reqmango meta states|labels|issue-types
reqmango agent list|dispatch <agentId> "task..."|task <taskId>
reqmango ask "..." [--issue <id>]
```

- 全局 flag：`--workspace`、`--project`（记忆在 config，可覆盖）、`--output table|json`
- 表格输出对齐列宽；`--json` 输出原始 API 结构（管道友好）
- issue 支持 `DEMO-42` 格式定位（client 侧解析：project identifier + issue sequence_id 匹配），也支持数字 ID
- 配置：`~/.reqmango/config.json` 存 api_url + PAT + 当前 workspace/project

## 7. 数据流与错误处理

**client 包统一错误模型**：

- 所有请求走一个 `do(ctx, method, path, query, body, &out)` 核心函数
- HTTP 4xx/5xx → 解析后端错误体（扁平 `{"message": ...}`，见后端错误契约）→ 映射为类型化错误 `*APIError{StatusCode, Message}`（完整响应体保留在 `Body`）
- 实现说明：原设计稿的 `{error: {code, message}}` 形状以实际后端契约为准做了修正（§7 修订于 2026-09-06 全分支终审）
- **401 特殊处理**：CLI 打印"PAT 已失效，请运行 `reqmango auth login`"；MCP 将 401 作为工具结果返回（`isError: true`）并附带同样提示
- 超时：client 默认 30s；`ai_chat`/`dispatch_agent` 等长操作允许 `context.WithTimeout` 延长到 5min
- MCP 工具执行失败**不**中断协议会话——错误以结构化 tool result 返回给模型，让 AI 自行纠正（如换正确的 project_id）

## 8. 测试策略

| 层 | 测试 | 工具 |
|---|---|---|
| backend PAT | AuthMiddleware 接受 PAT / 拒绝已吊销 / 拒绝过期；`/auth/tokens` CRUD + 用户隔离 | Go 单测（沿用 `internal/testutil`） |
| sdk/client | `httptest.Server` 模拟后端路由，覆盖分页、错误映射、401 | Go 单测 |
| sdk/mcp | fake client 测每个 tool 的参数校验与输出结构；stdio 一次握手往返 | Go 单测 |
| sdk/cli | 命令解析 + output 格式化 | Go 单测 |
| e2e 冒烟 | 脚本：起 backend → `reqmango login` → issue 全生命周期 → `reqmango-mcp` 走 `tools/list` + `create_issue` | bash 脚本（放 `scripts/`） |

## 9. 交付顺序

1. **backend PAT**（模型 + 迁移 + 端点 + 中间件 + 测试）
2. **sdk/client**（核心 do() + 各资源客户端 + 测试）
3. **sdk/mcp**（mcp-go 集成 + 24 tools + stdio/HTTP 双传输 + 测试）
4. **sdk/cli**（cobra 命令树 + config + 测试）
5. **e2e 冒烟 + README**（Claude Code `claude mcp add` 配置示例、Makefile、旧 mcp-server 目录删除）
