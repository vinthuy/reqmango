# reqmango tools

reqmango 的 MCP server 与 CLI，共享同一个 API 客户端（`client/`）。

```
sdk/
  client/   ← 共享 API 客户端（仅标准库）+ 全部 DTO + 错误映射
  mcp/      ← MCP server（mark3labs/mcp-go）：24 个工具
  cli/      ← reqmango CLI（cobra）
  cmd/reqmango-mcp   ← MCP server 二进制（stdio + streamable HTTP）
  cmd/reqmango       ← CLI 二进制
```

## 构建

```bash
cd sdk
go build -o ../bin/reqmango ./cmd/reqmango
go build -o ../bin/reqmango-mcp ./cmd/reqmango-mcp
```

或 `make tools`（仓库根目录）。

## 认证（PAT）

两种工具都使用 PAT（Personal Access Token）认证：

```bash
# 1. 登录（交换 JWT → 创建 PAT → 存 ~/.reqmango/config.json，JWT 即弃）
reqmango auth login
# 2. 查看 / 吊销
reqmango auth status
reqmango auth revoke --list
reqmango auth revoke <pat-id>
```

环境变量：`REQMANGO_API_URL`（默认 `http://localhost:8000/api/v1`）、`REQMANGO_PAT`。

## MCP server

### Claude Code

```bash
claude mcp add reqmango \
  --env REQMANGO_PAT=reqmango_pat_xxx \
  --env REQMANGO_API_URL=http://localhost:8000/api/v1 \
  -- /path/to/bin/reqmango-mcp
```

### Claude Desktop

`claude_desktop_config.json`：

```json
{
  "mcpServers": {
    "reqmango": {
      "command": "/path/to/bin/reqmango-mcp",
      "env": {
        "REQMANGO_PAT": "reqmango_pat_xxx",
        "REQMANGO_API_URL": "http://localhost:8000/api/v1"
      }
    }
  }
}
```

### Cursor

Cursor Settings → MCP → Add new MCP server：

- Type: `command`
- Command: `/path/to/bin/reqmango-mcp`
- Env: `REQMANGO_PAT=reqmango_pat_xxx`

### HTTP 模式（远程 / CI）

```bash
reqmango-mcp --http :8080
# 端点 POST http://host:8080/mcp（streamable HTTP，先 initialize 拿 Mcp-Session-Id）
# 每个请求必须带 Authorization: Bearer <PAT>，否则 401
```

⚠️ 安全：HTTP 模式每个请求都在明文携带全权限 PAT。生产/远程部署必须在前面架 TLS 终结反向代理（如 Nginx），或将地址绑定到回环（127.0.0.1）仅本机使用。绑定非回环地址时二进制会打印警告。

## 工具清单（24 个）

**核心（19）**：`list_workspaces` `list_projects` `get_project` `create_issue` `list_issues` `get_issue` `update_issue` `search_issues` `add_comment` `list_cycles` `get_cycle_progress` `add_issue_to_cycle` `list_members` `get_states` `get_labels` `list_issue_types` `list_notifications` `list_pages` `get_page`

**AI（5）**：`ai_search` `ai_chat` `list_agents` `dispatch_agent` `get_agent_task`

工具错误以 `isError: true` 的结构化结果返回，不会中断 MCP 会话；401 会附带 `reqmango auth login` 的修复提示。

## CLI

```bash
reqmango auth login|logout|status|revoke
reqmango workspace list|switch <id>
reqmango project list|show <id|identifier>|create
reqmango issue list|show <id|code>|create|update <id|code>|search <query>
reqmango cycle list|progress <id>|burndown <id>
reqmango meta states|labels|issue-types
reqmango agent list|dispatch <agentId> "task..."|task <taskId>
reqmango ask "..." [--issue <id|code>]
```

- 全局 flag：`--workspace` `--project`（记忆在 config，可覆盖）、`--output table|json`
- issue 支持 `DEMO-42` 格式定位（project identifier + sequence_id），也支持数字 ID
- `--assignee me` 表示当前用户

## 测试

```bash
cd sdk && go test ./...          # 单元测试
bash ../scripts/e2e_tools.sh     # e2e 冒烟（需后端在跑）
```
