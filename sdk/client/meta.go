package client

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// State is a workflow state.
type State struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Group       string `json:"group"`
	Sequence    int    `json:"sequence"`
	IsDefault   bool   `json:"is_default"`
	IsActive    bool   `json:"is_active"`
	ProjectID   uint64 `json:"project_id"`
	WorkspaceID uint64 `json:"workspace_id"`
}

// ListStates returns project workflow states (GET /projects/:id/settings/states).
func (c *Client) ListStates(ctx context.Context, projectID uint64) ([]State, error) {
	var out []State
	_, err := c.GetJSON(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/settings/states", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Label is a project label.
type Label struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
	ProjectID   uint64 `json:"project_id"`
	WorkspaceID uint64 `json:"workspace_id"`
}

// ListLabels returns project labels (GET /projects/:id/settings/labels).
func (c *Client) ListLabels(ctx context.Context, projectID uint64) ([]Label, error) {
	var out []Label
	_, err := c.GetJSON(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/settings/labels", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IssueType is an issue type definition.
type IssueType struct {
	ID           uint64  `json:"id"`
	Name         string  `json:"name"`
	Color        string  `json:"color"`
	Icon         string  `json:"icon"`
	Description  string  `json:"description"`
	Level        string  `json:"level"`
	ParentTypeID *uint64 `json:"parent_type_id"`
	IsDefault    bool    `json:"is_default"`
	Sequence     int     `json:"sequence"`
	IsActive     bool    `json:"is_active"`
	ProjectID    uint64  `json:"project_id"`
	WorkspaceID  uint64  `json:"workspace_id"`
}

// ListIssueTypes returns issue types (GET /issue-types?workspace_id=..., workspace_id required).
func (c *Client) ListIssueTypes(ctx context.Context, workspaceID, projectID uint64) ([]IssueType, error) {
	q := url.Values{}
	q.Set("workspace_id", strconv.FormatUint(workspaceID, 10))
	if projectID != 0 {
		q.Set("project_id", strconv.FormatUint(projectID, 10))
	}
	var out []IssueType
	if _, err := c.GetJSON(ctx, "/issue-types", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Page is a project document.
type Page struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Published   bool      `json:"published"`
	Sequence    int       `json:"sequence"`
	ParentID    *uint64   `json:"parent_id"`
	Depth       int       `json:"depth"`
	ProjectID   uint64    `json:"project_id"`
	WorkspaceID uint64    `json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListPages returns project documents (GET /projects/:id/pages).
func (c *Client) ListPages(ctx context.Context, projectID uint64) ([]Page, error) {
	var out []Page
	_, err := c.GetJSON(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/pages", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetPage fetches one page (GET /projects/:pid/pages/:pageId).
func (c *Client) GetPage(ctx context.Context, projectID, pageID uint64) (*Page, error) {
	path := "/projects/" + strconv.FormatUint(projectID, 10) + "/pages/" + strconv.FormatUint(pageID, 10)
	var out Page
	if _, err := c.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Notification is a user notification.
type Notification struct {
	ID        uint64     `json:"id"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	Type      string     `json:"type"`
	Priority  string     `json:"priority"`
	IsRead    bool       `json:"is_read"`
	ReadAt    *time.Time `json:"read_at"`
	ActionURL string     `json:"action_url"`
	IssueID   *uint64    `json:"issue_id"`
	CreatedAt time.Time  `json:"created_at"`
}

// ListNotifications returns the user's notifications (GET /notifications).
func (c *Client) ListNotifications(ctx context.Context, unreadOnly bool, limit, offset int) ([]Notification, error) {
	q := url.Values{}
	if unreadOnly {
		q.Set("unread_only", "true")
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	var out []Notification
	if _, err := c.GetJSON(ctx, "/notifications", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}
