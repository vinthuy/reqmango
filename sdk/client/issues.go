package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Issue is the response shape of GET /issues/:id (subset of IssueResponse).
type Issue struct {
	ID              uint64     `json:"id"`
	Name            string     `json:"name"`
	DescriptionHTML string     `json:"description_html"`
	Priority        string     `json:"priority"`
	SequenceID      int        `json:"sequence_id"`
	SortOrder       float64    `json:"sort_order"`
	StartDate       *time.Time `json:"start_date"`
	TargetDate      *time.Time `json:"target_date"`
	CompletedAt     *time.Time `json:"completed_at"`
	IsDraft         bool       `json:"is_draft"`
	ArchivedAt      *time.Time `json:"archived_at"`
	ProjectID       uint64     `json:"project_id"`
	WorkspaceID     uint64     `json:"workspace_id"`
	StateID         uint64     `json:"state_id"`
	StateName       string     `json:"state_name"`
	StateGroup      string     `json:"state_group"`
	ParentID        *uint64    `json:"parent_id"`
	Depth           int        `json:"depth"`
	Assignees       []UserLite `json:"assignees"`
	Labels          []uint64   `json:"labels"`
	SubIssuesCount  int64      `json:"sub_issues_count"`
	LinkCount       int        `json:"link_count"`
	EstimatePointID *uint64    `json:"estimate_point_id"`
	CycleID         *uint64    `json:"cycle_id"`
	ModuleIDs       []uint64   `json:"module_ids"`
	ReleaseID       *uint64    `json:"release_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// IssueListOptions maps to the GET /issues query parameters.
type IssueListOptions struct {
	ProjectID   uint64
	WorkspaceID uint64
	StateID     uint64
	Priority    string
	AssigneeID  uint64
	ParentID    uint64
	CycleID     uint64
	IssueTypeID uint64
	ModuleID    uint64
	Search      string
	RQL         string
	SortBy      string
	SortDir     string
	IsDraft     *bool
	Limit       int
	Offset      int
}

// Query encodes non-zero fields into url.Values.
func (o IssueListOptions) Query() url.Values {
	q := url.Values{}
	setU := func(k string, v uint64) {
		if v != 0 {
			q.Set(k, strconv.FormatUint(v, 10))
		}
	}
	setS := func(k, v string) {
		if v != "" {
			q.Set(k, v)
		}
	}
	setU("project_id", o.ProjectID)
	setU("workspace_id", o.WorkspaceID)
	setU("state_id", o.StateID)
	setS("priority", o.Priority)
	setU("assignee_id", o.AssigneeID)
	setU("parent_id", o.ParentID)
	setU("cycle_id", o.CycleID)
	setU("issue_type_id", o.IssueTypeID)
	setU("module_id", o.ModuleID)
	setS("search", o.Search)
	setS("rql", o.RQL)
	setS("sort_by", o.SortBy)
	setS("sort_dir", o.SortDir)
	if o.IsDraft != nil {
		q.Set("is_draft", strconv.FormatBool(*o.IsDraft))
	}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Offset > 0 {
		q.Set("offset", strconv.Itoa(o.Offset))
	}
	return q
}

// IssueListResult pairs the bare-array body with the X-Total-Count header.
type IssueListResult struct {
	Items []Issue
	Total int
}

// ListIssues lists issues. Total comes from the X-Total-Count response header.
func (c *Client) ListIssues(ctx context.Context, opts IssueListOptions) (*IssueListResult, error) {
	var items []Issue
	hdr, err := c.GetJSON(ctx, "/issues", opts.Query(), &items)
	if err != nil {
		return nil, err
	}
	total, _ := strconv.Atoi(hdr.Get("X-Total-Count"))
	return &IssueListResult{Items: items, Total: total}, nil
}

// CreateIssueRequest is the body for POST /issues (project_id & workspace_id go in query).
type CreateIssueRequest struct {
	Name            string   `json:"name"`
	DescriptionHTML string   `json:"description_html,omitempty"`
	Priority        string   `json:"priority,omitempty"`
	StateID         *uint64  `json:"state_id,omitempty"`
	AssigneeIDs     []uint64 `json:"assignee_ids,omitempty"`
	LabelIDs        []uint64 `json:"label_ids,omitempty"`
	StartDate       *string  `json:"start_date,omitempty"`
	TargetDate      *string  `json:"target_date,omitempty"`
	ParentID        *uint64  `json:"parent_id,omitempty"`
	TypeID          *uint64  `json:"type_id,omitempty"`
	CycleID         *uint64  `json:"cycle_id,omitempty"`
}

// CreateIssue creates an issue. projectID and workspaceID are required query params.
func (c *Client) CreateIssue(ctx context.Context, projectID, workspaceID uint64, req *CreateIssueRequest) (*Issue, error) {
	q := url.Values{}
	q.Set("project_id", strconv.FormatUint(projectID, 10))
	q.Set("workspace_id", strconv.FormatUint(workspaceID, 10))
	var out Issue
	if _, err := c.PostJSON(ctx, "/issues", q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIssue fetches one issue.
func (c *Client) GetIssue(ctx context.Context, id uint64) (*Issue, error) {
	var out Issue
	if _, err := c.GetJSON(ctx, "/issues/"+strconv.FormatUint(id, 10), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateIssueRequest mirrors IssueUpdateRequest (pointer fields = partial update;
// slices have replace-all semantics).
type UpdateIssueRequest struct {
	Name            *string  `json:"name,omitempty"`
	DescriptionHTML *string  `json:"description_html,omitempty"`
	Priority        *string  `json:"priority,omitempty"`
	StateID         *uint64  `json:"state_id,omitempty"`
	AssigneeIDs     []uint64 `json:"assignee_ids,omitempty"`
	LabelIDs        []uint64 `json:"label_ids,omitempty"`
	TargetDate      *string  `json:"target_date,omitempty"`
	ParentID        *uint64  `json:"parent_id,omitempty"`
	TypeID          *uint64  `json:"type_id,omitempty"`
	CycleID         *uint64  `json:"cycle_id,omitempty"`
}

// UpdateIssue updates an issue. A 409 (approval flow) surfaces as *APIError
// whose Body carries transition_id/workflow fields.
func (c *Client) UpdateIssue(ctx context.Context, id uint64, req *UpdateIssueRequest) (*Issue, error) {
	var out Issue
	if _, err := c.PutJSON(ctx, "/issues/"+strconv.FormatUint(id, 10), nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// IssueSearchResult is the shape of GET /issues/search.
type IssueSearchResult struct {
	ID                uint64 `json:"id"`
	Name              string `json:"name"`
	SequenceID        int    `json:"sequence_id"`
	ProjectIdentifier string `json:"project_identifier"`
	ProjectID         uint64 `json:"project_id"`
	WorkspaceSlug     string `json:"workspace_slug"`
}

// SearchIssues performs full-text search (workspace_id & query required).
func (c *Client) SearchIssues(ctx context.Context, workspaceID uint64, query string, projectID *uint64, limit int) ([]IssueSearchResult, error) {
	q := url.Values{}
	q.Set("workspace_id", strconv.FormatUint(workspaceID, 10))
	q.Set("query", query)
	if projectID != nil {
		q.Set("project_id", strconv.FormatUint(*projectID, 10))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	var out []IssueSearchResult
	if _, err := c.GetJSON(ctx, "/issues/search", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ResolveIssueCode resolves "DEMO-42" to a numeric issue ID:
// find the project by identifier in the workspace, then match sequence_id.
func (c *Client) ResolveIssueCode(ctx context.Context, workspaceID uint64, code string) (uint64, error) {
	parts := strings.SplitN(code, "-", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid issue code %q (expected IDENTIFIER-NUMBER)", code)
	}
	identifier, seqStr := parts[0], parts[1]
	seq, err := strconv.Atoi(seqStr)
	if err != nil {
		return 0, fmt.Errorf("invalid issue code %q: %w", code, err)
	}

	projects, err := c.ListProjects(ctx, workspaceID)
	if err != nil {
		return 0, err
	}
	var projectID uint64
	for _, p := range projects {
		if strings.EqualFold(p.Identifier, identifier) {
			projectID = p.ID
			break
		}
	}
	if projectID == 0 {
		return 0, fmt.Errorf("project with identifier %q not found", identifier)
	}

	res, err := c.ListIssues(ctx, IssueListOptions{ProjectID: projectID, Search: seqStr, Limit: 100})
	if err != nil {
		return 0, err
	}
	for _, it := range res.Items {
		if it.SequenceID == seq {
			return it.ID, nil
		}
	}
	return 0, fmt.Errorf("issue %s not found", code)
}

// Comment is the shape of POST /comments.
type Comment struct {
	ID         uint64     `json:"id"`
	IssueID    uint64     `json:"issue_id"`
	AuthorID   uint64     `json:"author_id"`
	Body       string     `json:"body"`
	IsResolved bool       `json:"is_resolved"`
	ParentID   *uint64    `json:"parent_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// AddComment creates a comment on an issue (POST /comments).
func (c *Client) AddComment(ctx context.Context, issueID uint64, body string, parentID *uint64) (*Comment, error) {
	var out Comment
	req := map[string]any{"issue_id": issueID, "body": body}
	if parentID != nil {
		req["parent_id"] = *parentID
	}
	if _, err := c.PostJSON(ctx, "/comments", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListComments lists comments (GET /comments/issue/:id → {comments,total,...}).
func (c *Client) ListComments(ctx context.Context, issueID uint64, page, pageSize int) ([]Comment, int, error) {
	q := url.Values{}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		q.Set("page_size", strconv.Itoa(pageSize))
	}
	var out struct {
		Comments []Comment `json:"comments"`
		Total    int       `json:"total"`
	}
	if _, err := c.GetJSON(ctx, "/comments/issue/"+strconv.FormatUint(issueID, 10), q, &out); err != nil {
		return nil, 0, err
	}
	return out.Comments, out.Total, nil
}
