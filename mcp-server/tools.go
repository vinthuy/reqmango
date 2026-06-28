package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// -------- MCP Tool definitions --------

// MCPServerInfo identifies this server in the MCP initialize response.
type MCPServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// MCPTool defines a tool exposed to MCP clients.
type MCPTool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	InputSchema *ToolInputSchema  `json:"inputSchema"`
}

// ToolInputSchema is a JSON Schema object describing tool parameters.
type ToolInputSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]SchemaProperty `json:"properties"`
	Required   []string                  `json:"required,omitempty"`
}

// SchemaProperty describes a single parameter.
type SchemaProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ListAllTools returns the static list of all available MCP tools.
func ListAllTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "list_projects",
			Description: "List all projects in a workspace. Returns project ID, name, identifier, and description.",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"workspace_id": {Type: "integer", Description: "The workspace ID to list projects from"},
				},
				Required: []string{"workspace_id"},
			},
		},
		{
			Name:        "get_project",
			Description: "Get details of a single project by its ID.",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"project_id": {Type: "integer", Description: "The project ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "create_issue",
			Description: "Create a new work item (issue) in a project. Use this to create bugs, features, tasks, etc.",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"workspace_id": {Type: "integer", Description: "The workspace ID"},
					"project_id":   {Type: "integer", Description: "The project ID to create the issue in"},
					"name":         {Type: "string", Description: "The issue title/name"},
					"description":  {Type: "string", Description: "Detailed description of the issue (optional)"},
					"priority":     {Type: "string", Description: "Priority level", Enum: []string{"urgent", "high", "medium", "low", "none"}},
					"state_id":     {Type: "integer", Description: "State ID for the issue (optional)"},
					"issue_type_id": {Type: "integer", Description: "Issue type ID (optional)"},
					"parent_id":    {Type: "integer", Description: "Parent issue ID for sub-issues (optional)"},
				},
				Required: []string{"workspace_id", "project_id", "name"},
			},
		},
		{
			Name:        "list_issues",
			Description: "List issues in a project. Supports optional filters for state, priority, and assignee.",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"project_id": {Type: "integer", Description: "The project ID"},
					"state_id":   {Type: "integer", Description: "Filter by state ID (optional)"},
					"priority":   {Type: "string", Description: "Filter by priority (optional)", Enum: []string{"urgent", "high", "medium", "low", "none"}},
					"assignee_id": {Type: "integer", Description: "Filter by assignee user ID (optional)"},
					"limit":      {Type: "integer", Description: "Maximum number of results (default 50)"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "get_issue",
			Description: "Get full details of a single issue by its ID.",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"issue_id": {Type: "integer", Description: "The issue ID"},
				},
				Required: []string{"issue_id"},
			},
		},
		{
			Name:        "update_issue",
			Description: "Update an existing issue's properties. Only the fields you provide will be updated.",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"issue_id":     {Type: "integer", Description: "The issue ID to update"},
					"name":         {Type: "string", Description: "New title (optional)"},
					"description":  {Type: "string", Description: "New description (optional)"},
					"priority":     {Type: "string", Description: "New priority", Enum: []string{"urgent", "high", "medium", "low", "none"}},
					"state_id":     {Type: "integer", Description: "New state ID (optional)"},
					"issue_type_id": {Type: "integer", Description: "New issue type ID (optional)"},
				},
				Required: []string{"issue_id"},
			},
		},
		{
			Name:        "search_issues",
			Description: "Search issues across the workspace by text query (matches name and description).",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"workspace_id": {Type: "integer", Description: "The workspace ID to search in"},
					"query":        {Type: "string", Description: "Search query string"},
				},
				Required: []string{"workspace_id", "query"},
			},
		},
		{
			Name:        "list_cycles",
			Description: "List all cycles (sprints) in a project.",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"project_id": {Type: "integer", Description: "The project ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "get_cycle_progress",
			Description: "Get detailed progress statistics for a cycle, including burndown data.",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"cycle_id": {Type: "integer", Description: "The cycle ID"},
				},
				Required: []string{"cycle_id"},
			},
		},
		{
			Name:        "add_issue_to_cycle",
			Description: "Add an existing issue to a cycle (sprint).",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"cycle_id": {Type: "integer", Description: "The cycle ID"},
					"issue_id": {Type: "integer", Description: "The issue ID to add"},
				},
				Required: []string{"cycle_id", "issue_id"},
			},
		},
		{
			Name:        "list_members",
			Description: "List all members of a project with their roles.",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"project_id": {Type: "integer", Description: "The project ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "get_states",
			Description: "Get all workflow states defined for a project (backlog, todo, in progress, done, etc.).",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"project_id": {Type: "integer", Description: "The project ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "get_labels",
			Description: "Get all labels defined for a project with their colors.",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"project_id": {Type: "integer", Description: "The project ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "list_issue_types",
			Description: "List all available issue types (Bug, Feature, Epic, Task, etc.) for a workspace.",
			InputSchema: &ToolInputSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"workspace_id": {Type: "integer", Description: "The workspace ID"},
				},
				Required: []string{"workspace_id"},
			},
		},
	}
}

// -------- Tool content types for MCP responses --------

// MCPTextContent represents a text content block in a tool result.
type MCPTextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// MCPToolResult is the result returned from a tool call.
type MCPToolResult struct {
	Content []MCPTextContent `json:"content"`
	IsError bool             `json:"isError,omitempty"`
}

// -------- Tool execution dispatch --------

// ExecuteTool dispatches a tool name with arguments to the appropriate client method.
func ExecuteTool(name string, args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	switch name {
	case "list_projects":
		return execListProjects(args, client)
	case "get_project":
		return execGetProject(args, client)
	case "create_issue":
		return execCreateIssue(args, client)
	case "list_issues":
		return execListIssues(args, client)
	case "get_issue":
		return execGetIssue(args, client)
	case "update_issue":
		return execUpdateIssue(args, client)
	case "search_issues":
		return execSearchIssues(args, client)
	case "list_cycles":
		return execListCycles(args, client)
	case "get_cycle_progress":
		return execGetCycleProgress(args, client)
	case "add_issue_to_cycle":
		return execAddIssueToCycle(args, client)
	case "list_members":
		return execListMembers(args, client)
	case "get_states":
		return execGetStates(args, client)
	case "get_labels":
		return execGetLabels(args, client)
	case "list_issue_types":
		return execListIssueTypes(args, client)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// -------- Helper functions --------

func getUint64(args map[string]interface{}, key string) (uint64, error) {
	v, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("missing required argument: %s", key)
	}
	switch n := v.(type) {
	case float64:
		return uint64(n), nil
	case int:
		return uint64(n), nil
	case int64:
		return uint64(n), nil
	case uint64:
		return n, nil
	case string:
		parsed, err := strconv.ParseUint(n, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %w", key, err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("invalid type for %s: %T", key, v)
	}
}

func getString(args map[string]interface{}, key string) (string, bool) {
	v, ok := args[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func toJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(data)
}

// -------- Tool executors --------

func execListProjects(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	wsID, err := getUint64(args, "workspace_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	projects, err := client.ListProjects(wsID)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(projects)}}}, nil
}

func execGetProject(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	projectID, err := getUint64(args, "project_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	project, err := client.GetProject(projectID)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(project)}}}, nil
}

func execCreateIssue(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	wsID, err := getUint64(args, "workspace_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	projectID, err := getUint64(args, "project_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	name, _ := getString(args, "name")
	if name == "" {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: "missing required argument: name"}}, IsError: true}, nil
	}

	payload := &IssueCreatePayload{Name: name}
	if desc, ok := getString(args, "description"); ok {
		payload.Description = desc
	}
	if pri, ok := getString(args, "priority"); ok {
		payload.Priority = pri
	}
	if stateID, err := getUint64(args, "state_id"); err == nil {
		payload.StateID = &stateID
	}
	if typeID, err := getUint64(args, "issue_type_id"); err == nil {
		payload.IssueTypeID = &typeID
	}
	if parentID, err := getUint64(args, "parent_id"); err == nil {
		payload.ParentID = &parentID
	}

	issue, err := client.CreateIssue(wsID, projectID, payload)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(issue)}}}, nil
}

func execListIssues(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	projectID, err := getUint64(args, "project_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	filters := make(map[string]string)
	if stateID, err := getUint64(args, "state_id"); err == nil {
		filters["state_id"] = strconv.FormatUint(stateID, 10)
	}
	if pri, ok := getString(args, "priority"); ok {
		filters["priority"] = pri
	}
	if assigneeID, err := getUint64(args, "assignee_id"); err == nil {
		filters["assignee_id"] = strconv.FormatUint(assigneeID, 10)
	}
	if limit, err := getUint64(args, "limit"); err == nil {
		filters["limit"] = strconv.FormatUint(limit, 10)
	}

	issues, err := client.ListIssues(projectID, filters)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(issues)}}}, nil
}

func execGetIssue(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	issueID, err := getUint64(args, "issue_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	issue, err := client.GetIssue(issueID)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(issue)}}}, nil
}

func execUpdateIssue(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	issueID, err := getUint64(args, "issue_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}

	payload := &IssueUpdatePayload{}
	if name, ok := getString(args, "name"); ok {
		payload.Name = &name
	}
	if desc, ok := getString(args, "description"); ok {
		payload.Description = &desc
	}
	if pri, ok := getString(args, "priority"); ok {
		payload.Priority = &pri
	}
	if stateID, err := getUint64(args, "state_id"); err == nil {
		payload.StateID = &stateID
	}
	if typeID, err := getUint64(args, "issue_type_id"); err == nil {
		payload.IssueTypeID = &typeID
	}

	issue, err := client.UpdateIssue(issueID, payload)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(issue)}}}, nil
}

func execSearchIssues(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	wsID, err := getUint64(args, "workspace_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	query, _ := getString(args, "query")
	if query == "" {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: "missing required argument: query"}}, IsError: true}, nil
	}

	result, err := client.SearchIssues(wsID, query)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(result)}}}, nil
}

func execListCycles(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	projectID, err := getUint64(args, "project_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	cycles, err := client.ListCycles(projectID)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(cycles)}}}, nil
}

func execGetCycleProgress(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	cycleID, err := getUint64(args, "cycle_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	progress, err := client.GetCycleProgress(cycleID)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(progress)}}}, nil
}

func execAddIssueToCycle(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	cycleID, err := getUint64(args, "cycle_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	issueID, err := getUint64(args, "issue_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	if err := client.AddIssueToCycle(cycleID, issueID); err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Successfully added issue %d to cycle %d", issueID, cycleID)}}}, nil
}

func execListMembers(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	projectID, err := getUint64(args, "project_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	members, err := client.ListMembers(projectID)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(members)}}}, nil
}

func execGetStates(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	projectID, err := getUint64(args, "project_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	states, err := client.GetStates(projectID)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(states)}}}, nil
}

func execGetLabels(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	projectID, err := getUint64(args, "project_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	labels, err := client.GetLabels(projectID)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(labels)}}}, nil
}

func execListIssueTypes(args map[string]interface{}, client *Client) (*MCPToolResult, error) {
	wsID, err := getUint64(args, "workspace_id")
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
	}
	types, err := client.ListIssueTypes(wsID)
	if err != nil {
		return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}, nil
	}
	return &MCPToolResult{Content: []MCPTextContent{{Type: "text", Text: toJSON(types)}}}, nil
}
