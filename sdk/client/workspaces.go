package client

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// Workspace is the list shape of GET /workspaces.
type Workspace struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListWorkspaces returns the workspaces visible to the user.
func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	var out []Workspace
	if _, err := c.GetJSON(ctx, "/workspaces", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// WorkspaceCreateRequest is the payload for POST /workspaces.
type WorkspaceCreateRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// CreateWorkspace creates a new workspace and returns it.
func (c *Client) CreateWorkspace(ctx context.Context, req WorkspaceCreateRequest) (*Workspace, error) {
	var out Workspace
	if _, err := c.PostJSON(ctx, "/workspaces", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UserLite is the embedded user shape in member/issue responses.
type UserLite struct {
	ID          uint64 `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
}

// Member is a workspace member.
type Member struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspace_id"`
	UserID      uint64    `json:"user_id"`
	Role        int       `json:"role"`
	IsActive    bool      `json:"is_active"`
	User        UserLite  `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListMembers returns workspace members.
func (c *Client) ListMembers(ctx context.Context, workspaceID uint64) ([]Member, error) {
	var out []Member
	_, err := c.GetJSON(ctx, "/workspaces/"+strconv.FormatUint(workspaceID, 10)+"/members", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Project is the list shape of GET /projects.
type Project struct {
	ID           uint64     `json:"id"`
	Name         string     `json:"name"`
	Identifier   string     `json:"identifier"`
	Description  string     `json:"description"`
	WorkspaceID  uint64     `json:"workspace_id"`
	ArchivedAt   *time.Time `json:"archived_at"`
	TotalIssues  int64      `json:"total_issues"`
	TotalMembers int64      `json:"total_members"`
	IsFavorite   bool       `json:"is_favorite"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ListProjects returns projects in a workspace (workspace_id query is required).
func (c *Client) ListProjects(ctx context.Context, workspaceID uint64) ([]Project, error) {
	q := url.Values{}
	q.Set("workspace_id", strconv.FormatUint(workspaceID, 10))
	var out []Project
	if _, err := c.GetJSON(ctx, "/projects", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProject fetches one project by numeric ID.
func (c *Client) GetProject(ctx context.Context, id uint64) (*Project, error) {
	var out Project
	if _, err := c.GetJSON(ctx, "/projects/"+strconv.FormatUint(id, 10), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ProjectCreateRequest is the payload for POST /projects.
type ProjectCreateRequest struct {
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description,omitempty"`
}

// CreateProject creates a project in the given workspace.
func (c *Client) CreateProject(ctx context.Context, workspaceID uint64, req *ProjectCreateRequest) (*Project, error) {
	q := url.Values{}
	q.Set("workspace_id", strconv.FormatUint(workspaceID, 10))
	var out Project
	if _, err := c.PostJSON(ctx, "/projects", q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
