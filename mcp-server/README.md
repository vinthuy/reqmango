# ReqManPy MCP Server

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server that enables AI assistants like **Claude**, **Cursor**, and **VS Code** to interact with your [ReqManPy](https://github.com/your-org/reqmanpy) project management instance.

## Quick Start

### 1. Build

```bash
cd mcp-server
go build -o reqmanpy-mcp .
```

### 2. Configure

Set these environment variables:

```bash
export REQMANPY_API_URL="http://localhost:8000/api/v1"  # Your ReqManPy instance
export REQMANPY_API_TOKEN="your-jwt-token-here"         # JWT Bearer token from ReqManPy
```

To get an API token, log in to your ReqManPy instance and use the token from the auth login response.

### 3. Test

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./reqmanpy-mcp
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | ./reqmanpy-mcp
```

---

## Claude Desktop Setup

Add this to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "reqmanpy": {
      "command": "/path/to/reqmanpy-mcp",
      "env": {
        "REQMANPY_API_URL": "http://localhost:8000/api/v1",
        "REQMANPY_API_TOKEN": "your-jwt-token-here"
      }
    }
  }
}
```

Restart Claude Desktop. You should see 🔨 14 tools when you look at connected MCP servers.

---

## Claude Code (CLI) Setup

```bash
claude mcp add reqmanpy \
  --command /path/to/reqmanpy-mcp \
  --env REQMANPY_API_URL=http://localhost:8000/api/v1 \
  --env REQMANPY_API_TOKEN=your-jwt-token-here
```

---

## VS Code / Cursor Setup

Add to your `.vscode/mcp.json` or Cursor MCP config:

```json
{
  "servers": {
    "reqmanpy": {
      "command": "/path/to/reqmanpy-mcp",
      "env": {
        "REQMANPY_API_URL": "http://localhost:8000/api/v1",
        "REQMANPY_API_TOKEN": "your-jwt-token-here"
      }
    }
  }
}
```

---

## Available Tools (14)

| Tool | Description |
|------|-------------|
| `list_projects` | List all projects in a workspace |
| `get_project` | Get project details by ID |
| `create_issue` | Create a new work item |
| `list_issues` | List issues with filters (state, priority, assignee) |
| `get_issue` | Get issue details by ID |
| `update_issue` | Update issue properties |
| `search_issues` | Full-text search across workspace issues |
| `list_cycles` | List cycles/sprints in a project |
| `get_cycle_progress` | Get cycle progress and burndown stats |
| `add_issue_to_cycle` | Add an issue to a sprint |
| `list_members` | List project members |
| `get_states` | Get workflow states defined for a project |
| `get_labels` | Get labels defined for a project |
| `list_issue_types` | List issue types in a workspace |

---

## Example Conversations

Once connected, you can ask your AI assistant things like:

- *"Show me all projects in workspace 1"* → calls `list_projects`
- *"Create a high-priority bug: Login button broken on mobile"* → calls `create_issue`
- *"What are all the urgent issues in project 1?"* → calls `list_issues`
- *"Move issue 42 to 'In Progress' and assign it to user 3"* → calls `update_issue`
- *"Show me progress for the current sprint"* → calls `get_cycle_progress`
- *"Search for all issues related to authentication"* → calls `search_issues`

---

## Troubleshooting

**"Parse error" on startup**: Make sure you're sending valid JSON (one line per message, newline-delimited).

**API errors**: Verify `REQMANPY_API_URL` is correct and the ReqManPy instance is running.

**401 Unauthorized**: Your `REQMANPY_API_TOKEN` may be expired. Generate a new one by logging in.

**Tool returns empty results**: Verify the IDs you're using exist in your ReqManPy instance (workspace_id, project_id, etc.).

---

## License

MIT — same as ReqManPy.
