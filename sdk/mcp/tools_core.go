package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/reqmango/tools/client"
)

// decodeArgs converts CallToolRequest.Arguments into the target struct.
func decodeArgs(req mcp.CallToolRequest, out any) error {
	raw, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}
	return json.Unmarshal(raw, out)
}

// toolResultJSON renders v as indented JSON text content.
func toolResultJSON(v any) *mcp.CallToolResult {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultErrorf("failed to serialize result: %v", err)
	}
	return mcp.NewToolResultText(string(data))
}

// toolAPIError maps an error to an isError tool result. 401s get a re-login
// hint so the model can tell the user what to fix.
func toolAPIError(err error) *mcp.CallToolResult {
	if apiErr := client.AsAPIError(err); apiErr != nil && apiErr.StatusCode == 401 {
		return mcp.NewToolResultError("authentication failed (401): the configured PAT is invalid or revoked. Create a new one with `reqmango auth login` and restart with the new REQMANGO_PAT.")
	}
	return mcp.NewToolResultError(err.Error())
}

// parseIDList splits a comma-separated ID list into uint64s.
func parseIDList(s string) []uint64 {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []uint64
	for _, part := range strings.Split(s, ",") {
		if n, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64); err == nil && n != 0 {
			out = append(out, n)
		}
	}
	return out
}

// resolveIssueArg accepts "123" or "DEMO-42" and returns the numeric issue ID.
func resolveIssueArg(ctx context.Context, cli *client.Client, workspaceID int64, arg string) (uint64, error) {
	if n, err := strconv.ParseUint(arg, 10, 64); err == nil {
		return n, nil
	}
	if workspaceID == 0 {
		return 0, fmt.Errorf("issue code %q needs workspace_id to resolve the project identifier", arg)
	}
	return cli.ResolveIssueCode(ctx, uint64(workspaceID), arg)
}

type listProjectsArgs struct {
	WorkspaceID int64 `json:"workspace_id"`
}

type getProjectArgs struct {
	ProjectID int64 `json:"project_id"`
}

type createIssueArgs struct {
	WorkspaceID     int64  `json:"workspace_id"`
	ProjectID       int64  `json:"project_id"`
	Name            string `json:"name"`
	DescriptionHTML string `json:"description_html"`
	Priority        string `json:"priority"`
	StateID         int64  `json:"state_id"`
	AssigneeIDs     string `json:"assignee_ids"` // comma-separated user ids
	LabelIDs        string `json:"label_ids"`    // comma-separated label ids
	ParentID        int64  `json:"parent_id"`
	TypeID          int64  `json:"type_id"`
	CycleID         int64  `json:"cycle_id"`
	TargetDate      string `json:"target_date"`
}

type listIssuesArgs struct {
	WorkspaceID int64  `json:"workspace_id"`
	ProjectID   int64  `json:"project_id"`
	RQL         string `json:"rql"`
	StateID     int64  `json:"state_id"`
	Priority    string `json:"priority"`
	AssigneeID  int64  `json:"assignee_id"`
	CycleID     int64  `json:"cycle_id"`
	IssueTypeID int64  `json:"issue_type_id"`
	Search      string `json:"search"`
	SortBy      string `json:"sort_by"`
	SortDir     string `json:"sort_dir"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

type getUpdateIssueArgs struct {
	Issue       string `json:"issue"` // numeric ID or "DEMO-42" (code needs workspace_id)
	WorkspaceID int64  `json:"workspace_id"`
}

type updateIssueArgs struct {
	Issue       string `json:"issue"` // numeric ID or "DEMO-42"
	WorkspaceID int64  `json:"workspace_id"`
	Name        string `json:"name"`
	Priority    string `json:"priority"`
	StateID     int64  `json:"state_id"`
	AssigneeIDs string `json:"assignee_ids"` // replace-all
	LabelIDs    string `json:"label_ids"`    // replace-all
	CycleID     int64  `json:"cycle_id"`
	TargetDate  string `json:"target_date"`
}

type searchIssuesArgs struct {
	WorkspaceID int64  `json:"workspace_id"`
	Query       string `json:"query"`
	ProjectID   int64  `json:"project_id"`
	Limit       int    `json:"limit"`
}

type addCommentArgs struct {
	IssueID int64  `json:"issue_id"`
	Body    string `json:"body"`
}

type listCyclesArgs struct {
	ProjectID int64  `json:"project_id"`
	Status    string `json:"status"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

type cycleIDArgs struct {
	CycleID int64 `json:"cycle_id"`
}

type addIssueToCycleArgs struct {
	CycleID int64 `json:"cycle_id"`
	IssueID int64 `json:"issue_id"`
}

type workspaceIDArgs struct {
	WorkspaceID int64 `json:"workspace_id"`
}

type issueTypesArgs struct {
	WorkspaceID int64 `json:"workspace_id"`
	ProjectID   int64 `json:"project_id"`
}

type notificationsArgs struct {
	UnreadOnly bool `json:"unread_only"`
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
}

type getPageArgs struct {
	ProjectID int64 `json:"project_id"`
	PageID    int64 `json:"page_id"`
}

// registerCoreTools adds the 19 core CRUD/meta tools.
func registerCoreTools(s *server.MCPServer, cli *client.Client) {
	s.AddTool(mcp.NewTool("list_workspaces",
		mcp.WithDescription("List workspaces the current user can access. Call this first to discover workspace IDs.")),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			out, err := cli.ListWorkspaces(ctx)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_projects",
		mcp.WithDescription("List projects in a workspace"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID from list_workspaces"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a listProjectsArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListProjects(ctx, uint64(a.WorkspaceID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_project",
		mcp.WithDescription("Get one project by numeric ID"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getProjectArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.GetProject(ctx, uint64(a.ProjectID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("create_issue",
		mcp.WithDescription("Create an issue in a project"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID")),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Issue title (1-255 chars)")),
		mcp.WithString("description_html", mcp.Description("Issue description (HTML)")),
		mcp.WithString("priority", mcp.Description("Priority: none|low|medium|high|urgent")),
		mcp.WithInteger("state_id", mcp.Description("Initial workflow state ID (from get_states); default state used if omitted")),
		mcp.WithString("assignee_ids", mcp.Description("Comma-separated user IDs to assign")),
		mcp.WithString("label_ids", mcp.Description("Comma-separated label IDs (from get_labels)")),
		mcp.WithInteger("parent_id", mcp.Description("Parent issue ID")),
		mcp.WithInteger("type_id", mcp.Description("Issue type ID (from list_issue_types)")),
		mcp.WithInteger("cycle_id", mcp.Description("Cycle ID to add the issue to")),
		mcp.WithString("target_date", mcp.Description("Target date (ISO 8601)"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a createIssueArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := &client.CreateIssueRequest{
				Name:            a.Name,
				DescriptionHTML: a.DescriptionHTML,
				Priority:        a.Priority,
				AssigneeIDs:     parseIDList(a.AssigneeIDs),
				LabelIDs:        parseIDList(a.LabelIDs),
			}
			if a.StateID != 0 {
				body.StateID = uintPtr64(uint64(a.StateID))
			}
			if a.ParentID != 0 {
				body.ParentID = uintPtr64(uint64(a.ParentID))
			}
			if a.TypeID != 0 {
				body.TypeID = uintPtr64(uint64(a.TypeID))
			}
			if a.CycleID != 0 {
				body.CycleID = uintPtr64(uint64(a.CycleID))
			}
			if a.TargetDate != "" {
				body.TargetDate = &a.TargetDate
			}
			out, err := cli.CreateIssue(ctx, uint64(a.ProjectID), uint64(a.WorkspaceID), body)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_issues",
		mcp.WithDescription("List issues with optional RQL query and filters. Returns items and total count."),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID")),
		mcp.WithInteger("workspace_id", mcp.Description("Workspace ID")),
		mcp.WithString("rql", mcp.Description("RQL query, e.g. `priority = \"high\" and state_group != \"completed\"`")),
		mcp.WithInteger("state_id", mcp.Description("Filter by workflow state ID")),
		mcp.WithString("priority", mcp.Description("Filter by priority")),
		mcp.WithInteger("assignee_id", mcp.Description("Filter by assignee user ID")),
		mcp.WithInteger("cycle_id", mcp.Description("Filter by cycle ID")),
		mcp.WithInteger("issue_type_id", mcp.Description("Filter by issue type ID")),
		mcp.WithString("search", mcp.Description("Free-text search")),
		mcp.WithString("sort_by", mcp.Description("Sort field")),
		mcp.WithString("sort_dir", mcp.Description("Sort direction: asc|desc")),
		mcp.WithInteger("limit", mcp.Description("Page size (default 50)")),
		mcp.WithInteger("offset", mcp.Description("Offset for pagination"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a listIssuesArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			res, err := cli.ListIssues(ctx, client.IssueListOptions{
				ProjectID:   uint64(a.ProjectID),
				WorkspaceID: uint64(a.WorkspaceID),
				RQL:         a.RQL,
				StateID:     uint64(a.StateID),
				Priority:    a.Priority,
				AssigneeID:  uint64(a.AssigneeID),
				CycleID:     uint64(a.CycleID),
				IssueTypeID: uint64(a.IssueTypeID),
				Search:      a.Search,
				SortBy:      a.SortBy,
				SortDir:     a.SortDir,
				Limit:       a.Limit,
				Offset:      a.Offset,
			})
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(map[string]any{"total": res.Total, "items": res.Items}), nil
		})

	s.AddTool(mcp.NewTool("get_issue",
		mcp.WithDescription("Get one issue by numeric ID or code like DEMO-42 (code needs workspace_id), with its comments"),
		mcp.WithString("issue", mcp.Required(), mcp.Description("Numeric issue ID or code like DEMO-42")),
		mcp.WithInteger("workspace_id", mcp.Description("Workspace ID — required when `issue` is a code"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getUpdateIssueArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			id, err := resolveIssueArg(ctx, cli, a.WorkspaceID, a.Issue)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			iss, err := cli.GetIssue(ctx, id)
			if err != nil {
				return toolAPIError(err), nil
			}
			comments, total, err := cli.ListComments(ctx, id, 1, 100)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(map[string]any{"issue": iss, "comments": comments, "comments_total": total}), nil
		})

	s.AddTool(mcp.NewTool("update_issue",
		mcp.WithDescription("Update an issue. State transitions go through the backend workflow validation and may require approval (409 with transition details)."),
		mcp.WithString("issue", mcp.Required(), mcp.Description("Numeric issue ID or code like DEMO-42")),
		mcp.WithInteger("workspace_id", mcp.Description("Workspace ID — required when `issue` is a code")),
		mcp.WithString("name", mcp.Description("New title")),
		mcp.WithString("priority", mcp.Description("New priority: none|low|medium|high|urgent")),
		mcp.WithInteger("state_id", mcp.Description("Target state ID (from get_states)")),
		mcp.WithString("assignee_ids", mcp.Description("Comma-separated user IDs — replaces all assignees")),
		mcp.WithString("label_ids", mcp.Description("Comma-separated label IDs — replaces all labels")),
		mcp.WithInteger("cycle_id", mcp.Description("Move to cycle ID")),
		mcp.WithString("target_date", mcp.Description("New target date (ISO 8601)"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a updateIssueArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			id, err := resolveIssueArg(ctx, cli, a.WorkspaceID, a.Issue)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := &client.UpdateIssueRequest{}
			if a.Name != "" {
				body.Name = &a.Name
			}
			if a.Priority != "" {
				body.Priority = &a.Priority
			}
			if a.StateID != 0 {
				body.StateID = uintPtr64(uint64(a.StateID))
			}
			if a.AssigneeIDs != "" {
				body.AssigneeIDs = parseIDList(a.AssigneeIDs)
			}
			if a.LabelIDs != "" {
				body.LabelIDs = parseIDList(a.LabelIDs)
			}
			if a.CycleID != 0 {
				body.CycleID = uintPtr64(uint64(a.CycleID))
			}
			if a.TargetDate != "" {
				body.TargetDate = &a.TargetDate
			}
			out, err := cli.UpdateIssue(ctx, id, body)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("search_issues",
		mcp.WithDescription("Full-text search across issues in a workspace"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query text")),
		mcp.WithInteger("project_id", mcp.Description("Limit search to one project")),
		mcp.WithInteger("limit", mcp.Description("Max results (default 10)"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a searchIssuesArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var pid *uint64
			if a.ProjectID != 0 {
				pid = uintPtr64(uint64(a.ProjectID))
			}
			out, err := cli.SearchIssues(ctx, uint64(a.WorkspaceID), a.Query, pid, a.Limit)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("add_comment",
		mcp.WithDescription("Add a comment to an issue"),
		mcp.WithInteger("issue_id", mcp.Required(), mcp.Description("Issue ID")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Comment text"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a addCommentArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.AddComment(ctx, uint64(a.IssueID), a.Body, nil)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_comments",
		mcp.WithDescription("List comments on an issue"),
		mcp.WithInteger("issue_id", mcp.Required(), mcp.Description("Issue ID")),
		mcp.WithInteger("page", mcp.Description("Page number (default 1)")),
		mcp.WithInteger("page_size", mcp.Description("Page size (default 20)"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a struct {
				IssueID  int64 `json:"issue_id"`
				Page     int   `json:"page"`
				PageSize int   `json:"page_size"`
			}
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			comments, total, err := cli.ListComments(ctx, uint64(a.IssueID), a.Page, a.PageSize)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(map[string]any{"comments": comments, "total": total}), nil
		})

	s.AddTool(mcp.NewTool("list_cycles",
		mcp.WithDescription("List cycles (sprints) of a project"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID")),
		mcp.WithString("status", mcp.Description("Filter: upcoming|active|completed|cancelled")),
		mcp.WithInteger("limit", mcp.Description("Page size")),
		mcp.WithInteger("offset", mcp.Description("Offset"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a listCyclesArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListCycles(ctx, uint64(a.ProjectID), a.Status, a.Limit, a.Offset)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_cycle",
		mcp.WithDescription("Get one cycle by ID"),
		mcp.WithInteger("cycle_id", mcp.Required(), mcp.Description("Cycle ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a cycleIDArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.GetCycle(ctx, uint64(a.CycleID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_cycle_burndown",
		mcp.WithDescription("Get a cycle's burndown chart data (daily progress)"),
		mcp.WithInteger("cycle_id", mcp.Required(), mcp.Description("Cycle ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a cycleIDArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.GetCycleBurndown(ctx, uint64(a.CycleID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_cycle_progress",
		mcp.WithDescription("Get a cycle's progress breakdown by state"),
		mcp.WithInteger("cycle_id", mcp.Required(), mcp.Description("Cycle ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a cycleIDArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.GetCycleProgress(ctx, uint64(a.CycleID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("add_issue_to_cycle",
		mcp.WithDescription("Add an existing issue to a cycle"),
		mcp.WithInteger("cycle_id", mcp.Required(), mcp.Description("Cycle ID")),
		mcp.WithInteger("issue_id", mcp.Required(), mcp.Description("Issue ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a addIssueToCycleArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if err := cli.AddIssueToCycle(ctx, uint64(a.CycleID), uint64(a.IssueID)); err != nil {
				return toolAPIError(err), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("added issue %d to cycle %d", a.IssueID, a.CycleID)), nil
		})

	s.AddTool(mcp.NewTool("list_members",
		mcp.WithDescription("List members of a workspace (user IDs for assignment)"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a workspaceIDArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListMembers(ctx, uint64(a.WorkspaceID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_states",
		mcp.WithDescription("List workflow states of a project (state IDs for create/update)"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getProjectArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListStates(ctx, uint64(a.ProjectID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_labels",
		mcp.WithDescription("List labels of a project (label IDs for create/update)"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getProjectArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListLabels(ctx, uint64(a.ProjectID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_issue_types",
		mcp.WithDescription("List issue types available in a workspace (optionally filtered by project)"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID")),
		mcp.WithInteger("project_id", mcp.Description("Project ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a issueTypesArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListIssueTypes(ctx, uint64(a.WorkspaceID), uint64(a.ProjectID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_notifications",
		mcp.WithDescription("List the current user's notifications"),
		mcp.WithBoolean("unread_only", mcp.Description("Only unread (default false)")),
		mcp.WithInteger("limit", mcp.Description("Page size")),
		mcp.WithInteger("offset", mcp.Description("Offset"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a notificationsArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListNotifications(ctx, a.UnreadOnly, a.Limit, a.Offset)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_pages",
		mcp.WithDescription("List document pages of a project"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getProjectArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListPages(ctx, uint64(a.ProjectID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_page",
		mcp.WithDescription("Get the content of one document page"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID")),
		mcp.WithInteger("page_id", mcp.Required(), mcp.Description("Page ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getPageArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.GetPage(ctx, uint64(a.ProjectID), uint64(a.PageID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})
}

// uintPtr64 returns a pointer to v.
func uintPtr64(v uint64) *uint64 { return &v }
