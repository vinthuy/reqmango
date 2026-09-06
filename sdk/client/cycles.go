package client

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// Cycle is the response shape (list wrapped in {items,total,...}).
type Cycle struct {
	ID              uint64     `json:"id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description"`
	Status          string     `json:"status"` // computed: upcoming|active|completed|cancelled
	Progress        float64    `json:"progress"`
	TotalIssues     int64      `json:"total_issues"`
	CompletedIssues int64      `json:"completed_issues"`
	StartDate       string     `json:"start_date"`
	EndDate         *string    `json:"end_date"`
	ProjectID       uint64     `json:"project_id"`
	WorkspaceID     uint64     `json:"workspace_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CycleListResult is the wrapped list shape.
type CycleListResult struct {
	Items  []Cycle
	Total  int
	Limit  int
	Offset int
}

// ListCycles lists project cycles (GET /projects/:id/cycles).
func (c *Client) ListCycles(ctx context.Context, projectID uint64, status string, limit, offset int) (*CycleListResult, error) {
	q := url.Values{}
	if status != "" {
		q.Set("status", status)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	var out CycleListResult
	if _, err := c.GetJSON(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/cycles", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCycle fetches one cycle.
func (c *Client) GetCycle(ctx context.Context, id uint64) (*Cycle, error) {
	var out Cycle
	if _, err := c.GetJSON(ctx, "/cycles/"+strconv.FormatUint(id, 10), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// StateBreakdown is one row of cycle progress state stats.
type StateBreakdown struct {
	State string `json:"state"`
	Group string `json:"group"`
	Count int64  `json:"count"`
}

// CycleProgress is the GET /cycles/:id/progress response.
type CycleProgress struct {
	CycleID         uint64           `json:"cycle_id"`
	CycleName       string           `json:"cycle_name"`
	TotalIssues     int64            `json:"total_issues"`
	CompletedIssues int64            `json:"completed_issues"`
	Progress        float64          `json:"progress"`
	StateBreakdown  []StateBreakdown `json:"state_breakdown"`
}

// GetCycleProgress fetches cycle progress.
func (c *Client) GetCycleProgress(ctx context.Context, id uint64) (*CycleProgress, error) {
	var out CycleProgress
	if _, err := c.GetJSON(ctx, "/cycles/"+strconv.FormatUint(id, 10)+"/progress", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BurndownDayPoint is one day of the burndown chart.
type BurndownDayPoint struct {
	DayIndex        int     `json:"day_index"`
	Date            string  `json:"date"`
	IdealRemaining  float64 `json:"ideal_remaining"`
	ActualCompleted int64   `json:"actual_completed"`
	ActualRemaining float64 `json:"actual_remaining"`
}

// BurndownData is the GET /cycles/:id/burndown response.
type BurndownData struct {
	CycleID         uint64            `json:"cycle_id"`
	CycleName       string            `json:"cycle_name"`
	StartDate       string            `json:"start_date"`
	EndDate         string            `json:"end_date"`
	TotalIssues     int64             `json:"total_issues"`
	TotalDays       int               `json:"total_days"`
	DaysElapsed     int               `json:"days_elapsed"`
	IdealDailyBurn  float64           `json:"ideal_daily_burn"`
	IdealRemaining  float64           `json:"ideal_remaining"`
	ActualCompleted int64             `json:"actual_completed"`
	ActualRemaining float64           `json:"actual_remaining"`
	IsOnTrack       bool              `json:"is_on_track"`
	DailyPoints     []BurndownDayPoint `json:"daily_points"`
}

// GetCycleBurndown fetches burndown data.
func (c *Client) GetCycleBurndown(ctx context.Context, id uint64) (*BurndownData, error) {
	var out BurndownData
	if _, err := c.GetJSON(ctx, "/cycles/"+strconv.FormatUint(id, 10)+"/burndown", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddIssueToCycle adds an issue to a cycle (POST /cycles/:id/issues?issue_id=).
func (c *Client) AddIssueToCycle(ctx context.Context, cycleID, issueID uint64) error {
	q := url.Values{}
	q.Set("issue_id", strconv.FormatUint(issueID, 10))
	var out map[string]any
	_, err := c.PostJSON(ctx, "/cycles/"+strconv.FormatUint(cycleID, 10)+"/issues", q, nil, &out)
	return err
}
