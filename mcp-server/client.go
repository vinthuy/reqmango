package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client is an HTTP client for the ReqManPy REST API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// -------- Lightweight API response types --------

// Project is a lightweight project representation from the API.
type Project struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description"`
	WorkspaceID uint64 `json:"workspace_id"`
	LeadID      *uint64 `json:"lead_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Issue is a lightweight issue representation from the API.
type Issue struct {
	ID          uint64      `json:"id"`
	Name        string      `json:"name"`
	SequenceID  int         `json:"sequence_id"`
	Description string      `json:"description"`
	Priority    string      `json:"priority"`
	StateID     uint64      `json:"state_id"`
	IssueTypeID *uint64     `json:"issue_type_id"`
	ProjectID   uint64      `json:"project_id"`
	ParentID    *uint64     `json:"parent_id"`
	StartDate   *string     `json:"start_date"`
	TargetDate  *string     `json:"target_date"`
	IsDraft     bool        `json:"is_draft"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

// Cycle represents a sprint/cycle from the API.
type Cycle struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Status      string  `json:"status"`
	ProjectID   uint64  `json:"project_id"`
	CreatedAt   string  `json:"created_at"`
}

// CycleProgress holds cycle progress statistics.
type CycleProgress struct {
	CycleID     uint64         `json:"cycle_id"`
	CycleName   string         `json:"cycle_name"`
	Status      string         `json:"status"`
	TotalIssues int            `json:"total_issues"`
	Completed   int            `json:"completed"`
	InProgress  int            `json:"in_progress"`
	NotStarted  int            `json:"not_started"`
	Cancelled   int            `json:"cancelled"`
	Progress    float64        `json:"progress"`
	StateBreakdown map[string]int `json:"state_breakdown,omitempty"`
}

// Member represents a project/workspace member.
type Member struct {
	ID       uint64 `json:"id"`
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     int    `json:"role"`
}

// State represents a workflow state.
type State struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	GroupName string `json:"group_name"`
	Sequence  int    `json:"sequence"`
	IsDefault bool   `json:"is_default"`
}

// Label represents a project label.
type Label struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	ProjectID uint64 `json:"project_id"`
}

// IssueType represents an issue type.
type IssueType struct {
	ID        uint64  `json:"id"`
	Name      string  `json:"name"`
	Icon      string  `json:"icon"`
	Color     string  `json:"color"`
	Level     int     `json:"level"`
	ProjectID *uint64 `json:"project_id"`
}

// IssueCreatePayload is sent to the create issue endpoint.
type IssueCreatePayload struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	StateID     *uint64  `json:"state_id,omitempty"`
	IssueTypeID *uint64  `json:"issue_type_id,omitempty"`
	ParentID    *uint64  `json:"parent_id,omitempty"`
	StartDate   *string  `json:"start_date,omitempty"`
	TargetDate  *string  `json:"target_date,omitempty"`
	AssigneeIDs []uint64 `json:"assignee_ids,omitempty"`
	LabelIDs    []uint64 `json:"label_ids,omitempty"`
}

// IssueUpdatePayload is sent to the update issue endpoint.
type IssueUpdatePayload struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Priority    *string  `json:"priority,omitempty"`
	StateID     *uint64  `json:"state_id,omitempty"`
	IssueTypeID *uint64  `json:"issue_type_id,omitempty"`
	StartDate   *string  `json:"start_date,omitempty"`
	TargetDate  *string  `json:"target_date,omitempty"`
	AssigneeIDs []uint64 `json:"assignee_ids,omitempty"`
	LabelIDs    []uint64 `json:"label_ids,omitempty"`
}

// SearchResult wraps a search response.
type SearchResult struct {
	Issues []Issue `json:"issues"`
	Total  int64   `json:"total"`
}

// -------- API methods --------

// doRequest performs an HTTP request and unmarshals the JSON response.
func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	fullURL := c.baseURL + path
	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respData))
	}

	if result != nil {
		if err := json.Unmarshal(respData, result); err != nil {
			return fmt.Errorf("unmarshal response: %w (body: %s)", err, string(respData))
		}
	}
	return nil
}

// ListProjects returns all projects in the given workspace.
func (c *Client) ListProjects(workspaceID uint64) ([]Project, error) {
	path := fmt.Sprintf("/projects?workspace_id=%d", workspaceID)
	var projects []Project
	if err := c.doRequest("GET", path, nil, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProject returns a single project by ID.
func (c *Client) GetProject(projectID uint64) (*Project, error) {
	path := fmt.Sprintf("/projects/%d", projectID)
	var project Project
	if err := c.doRequest("GET", path, nil, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateIssue creates a new issue.
func (c *Client) CreateIssue(workspaceID, projectID uint64, payload *IssueCreatePayload) (*Issue, error) {
	path := fmt.Sprintf("/issues?workspace_id=%d&project_id=%d", workspaceID, projectID)
	var issue Issue
	if err := c.doRequest("POST", path, payload, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// ListIssues returns issues for a project with optional filters.
func (c *Client) ListIssues(projectID uint64, filters map[string]string) ([]Issue, error) {
	u, _ := url.Parse(fmt.Sprintf("/issues"))
	q := u.Query()
	q.Set("project_id", strconv.FormatUint(projectID, 10))
	for k, v := range filters {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	path := u.String()

	var issues []Issue
	if err := c.doRequest("GET", path, nil, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

// GetIssue returns a single issue by ID.
func (c *Client) GetIssue(issueID uint64) (*Issue, error) {
	path := fmt.Sprintf("/issues/%d", issueID)
	var issue Issue
	if err := c.doRequest("GET", path, nil, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// UpdateIssue updates an existing issue.
func (c *Client) UpdateIssue(issueID uint64, payload *IssueUpdatePayload) (*Issue, error) {
	path := fmt.Sprintf("/issues/%d", issueID)
	var issue Issue
	if err := c.doRequest("PUT", path, payload, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// SearchIssues searches across the workspace by text query.
func (c *Client) SearchIssues(workspaceID uint64, query string) (*SearchResult, error) {
	path := fmt.Sprintf("/issues/search?workspace_id=%d&query=%s", workspaceID, url.QueryEscape(query))
	var result SearchResult
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListCycles returns cycles/sprints for a project.
func (c *Client) ListCycles(projectID uint64) ([]Cycle, error) {
	path := fmt.Sprintf("/projects/%d/cycles", projectID)
	var cycles []Cycle
	if err := c.doRequest("GET", path, nil, &cycles); err != nil {
		return nil, err
	}
	return cycles, nil
}

// GetCycleProgress returns progress stats for a cycle.
func (c *Client) GetCycleProgress(cycleID uint64) (*CycleProgress, error) {
	path := fmt.Sprintf("/cycles/%d/progress", cycleID)
	var progress CycleProgress
	if err := c.doRequest("GET", path, nil, &progress); err != nil {
		return nil, err
	}
	return &progress, nil
}

// AddIssueToCycle adds an issue to a cycle.
func (c *Client) AddIssueToCycle(cycleID, issueID uint64) error {
	path := fmt.Sprintf("/cycles/%d/issues?issue_id=%d", cycleID, issueID)
	return c.doRequest("POST", path, nil, nil)
}

// ListMembers returns project members.
func (c *Client) ListMembers(projectID uint64) ([]Member, error) {
	path := fmt.Sprintf("/projects/%d/members", projectID)
	var members []Member
	if err := c.doRequest("GET", path, nil, &members); err != nil {
		return nil, err
	}
	return members, nil
}

// GetStates returns workflow states for a project.
func (c *Client) GetStates(projectID uint64) ([]State, error) {
	path := fmt.Sprintf("/projects/%d/settings/states", projectID)
	var states []State
	if err := c.doRequest("GET", path, nil, &states); err != nil {
		return nil, err
	}
	return states, nil
}

// GetLabels returns labels for a project.
func (c *Client) GetLabels(projectID uint64) ([]Label, error) {
	path := fmt.Sprintf("/projects/%d/settings/labels", projectID)
	var labels []Label
	if err := c.doRequest("GET", path, nil, &labels); err != nil {
		return nil, err
	}
	return labels, nil
}

// ListIssueTypes returns issue types for a workspace.
func (c *Client) ListIssueTypes(workspaceID uint64) ([]IssueType, error) {
	path := fmt.Sprintf("/issue-types?workspace_id=%d", workspaceID)
	var types []IssueType
	if err := c.doRequest("GET", path, nil, &types); err != nil {
		return nil, err
	}
	return types, nil
}
