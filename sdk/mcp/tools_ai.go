package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/reqmango/tools/client"
)

type aiSearchArgs struct {
	ProjectID int64  `json:"project_id"`
	Query     string `json:"query"`
}

type aiChatArgs struct {
	ProjectID int64  `json:"project_id"`
	Message   string `json:"message"`
	ThreadID  int64  `json:"thread_id"`
}

type dispatchAgentArgs struct {
	WorkspaceID int64  `json:"workspace_id"`
	AgentID     int64  `json:"agent_id"`
	Task        string `json:"task"`
	IssueID     int64  `json:"issue_id"`
}

type agentTaskArgs struct {
	WorkspaceID int64 `json:"workspace_id"`
	TaskID      int64 `json:"task_id"`
}

// registerAITools adds the 5 AI capability tools.
func registerAITools(s *server.MCPServer, cli *client.Client) {
	s.AddTool(mcp.NewTool("ai_search",
		mcp.WithDescription("Convert a natural-language question into an issue search (returns RQL + matching issues)"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID to search in")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Natural-language question, e.g. 'high priority bugs from last week'"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a aiSearchArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.AISearch(ctx, uint64(a.ProjectID), a.Query)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("ai_chat",
		mcp.WithDescription("Send one message to the project AI assistant. The SSE stream is aggregated server-side into a single reply. Long-running (up to 5 min)."),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID")),
		mcp.WithString("message", mcp.Required(), mcp.Description("Message to send")),
		mcp.WithInteger("thread_id", mcp.Description("Thread ID from a previous ai_chat reply to continue the conversation"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a aiChatArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var tid *uint64
			if a.ThreadID != 0 {
				tid = uintPtr64(uint64(a.ThreadID))
			}
			reply, err := cli.AIChat(ctx, uint64(a.ProjectID), a.Message, tid)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(map[string]any{
				"text":       reply.Text,
				"thread_id":  reply.ThreadID,
				"tool_calls": reply.ToolCalls,
			}), nil
		})

	s.AddTool(mcp.NewTool("list_agents",
		mcp.WithDescription("List AI agents available in a workspace"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a workspaceIDArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListAgents(ctx, uint64(a.WorkspaceID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("dispatch_agent",
		mcp.WithDescription("Trigger an AI agent to run a task. Long-running (up to 5 min); returns the activity record."),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID")),
		mcp.WithInteger("agent_id", mcp.Required(), mcp.Description("Agent ID from list_agents")),
		mcp.WithString("task", mcp.Required(), mcp.Description("Task description for the agent")),
		mcp.WithInteger("issue_id", mcp.Description("Related issue ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a dispatchAgentArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var issueID *uint64
			if a.IssueID != 0 {
				issueID = uintPtr64(uint64(a.IssueID))
			}
			out, err := cli.DispatchAgent(ctx, uint64(a.WorkspaceID), uint64(a.AgentID), a.Task, issueID, nil)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_agent_task",
		mcp.WithDescription("Get an agent task's status, progress and recent logs"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID")),
		mcp.WithInteger("task_id", mcp.Required(), mcp.Description("Task ID (from agent-tasks listing)"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a agentTaskArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			task, err := cli.GetAgentTask(ctx, uint64(a.WorkspaceID), uint64(a.TaskID))
			if err != nil {
				return toolAPIError(err), nil
			}
			logs, _ := cli.GetAgentTaskLogs(ctx, uint64(a.WorkspaceID), uint64(a.TaskID))
			return toolResultJSON(map[string]any{"task": task, "logs": logs}), nil
		})
}
