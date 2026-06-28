package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/reqmanpy/backend/internal/common"
	"github.com/reqmanpy/backend/internal/model"
	"gorm.io/gorm"
)

type GitHubService struct {
	db *gorm.DB
}

func NewGitHubService(db *gorm.DB) *GitHubService {
	return &GitHubService{db: db}
}

// ======== Request/Response types ========

type GitHubCreateRequest struct {
	ProjectID      uint64 `json:"project_id" binding:"required"`
	RepoOwner      string `json:"repo_owner" binding:"required"`
	RepoName       string `json:"repo_name" binding:"required"`
	AccessToken    string `json:"access_token"`
	WebhookSecret  string `json:"webhook_secret"`
	IsEnabled      *bool  `json:"is_enabled"`
	SyncIssues     *bool  `json:"sync_issues"`
	SyncPRs        *bool  `json:"sync_prs"`
}

type GitHubUpdateRequest struct {
	RepoOwner      *string `json:"repo_owner"`
	RepoName       *string `json:"repo_name"`
	AccessToken    *string `json:"access_token"`
	WebhookSecret  *string `json:"webhook_secret"`
	IsEnabled      *bool   `json:"is_enabled"`
	SyncIssues     *bool   `json:"sync_issues"`
	SyncPRs        *bool   `json:"sync_prs"`
}

type GitHubResponse struct {
	ID            uint64  `json:"id"`
	WorkspaceID   uint64  `json:"workspace_id"`
	ProjectID     uint64  `json:"project_id"`
	RepoOwner     string  `json:"repo_owner"`
	RepoName      string  `json:"repo_name"`
	IsEnabled     bool    `json:"is_enabled"`
	SyncIssues    bool    `json:"sync_issues"`
	SyncPRs       bool    `json:"sync_prs"`
	LastSyncAt    *string `json:"last_sync_at"`
	WebhookID     *uint64 `json:"webhook_id"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type GitHubIssue struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	State     string   `json:"state"`
	Labels    []string `json:"labels"`
	HTMLURL   string   `json:"html_url"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// ======== CRUD ========

func (s *GitHubService) List(workspaceID uint64) ([]GitHubResponse, error) {
	var conns []model.GitHubConnection
	if err := s.db.Where("workspace_id = ?", workspaceID).Find(&conns).Error; err != nil {
		return nil, common.Internal("Failed to list GitHub connections")
	}
	res := make([]GitHubResponse, len(conns))
	for i, c := range conns {
		res[i] = s.toResponse(&c)
	}
	if res == nil {
		res = []GitHubResponse{}
	}
	return res, nil
}

func (s *GitHubService) Get(id uint64) (*GitHubResponse, error) {
	var conn model.GitHubConnection
	if err := s.db.First(&conn, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("GitHub connection not found")
		}
		return nil, common.Internal("Failed to get GitHub connection")
	}
	r := s.toResponse(&conn)
	return &r, nil
}

func (s *GitHubService) Create(workspaceID uint64, req *GitHubCreateRequest) (*GitHubResponse, error) {
	enabled := true
	if req.IsEnabled != nil {
		enabled = *req.IsEnabled
	}
	syncIssues := true
	if req.SyncIssues != nil {
		syncIssues = *req.SyncIssues
	}
	syncPRs := true
	if req.SyncPRs != nil {
		syncPRs = *req.SyncPRs
	}

	conn := model.GitHubConnection{
		WorkspaceID:   workspaceID,
		ProjectID:     req.ProjectID,
		RepoOwner:     req.RepoOwner,
		RepoName:      req.RepoName,
		AccessToken:   req.AccessToken,
		WebhookSecret: req.WebhookSecret,
		IsEnabled:     enabled,
		SyncIssues:    syncIssues,
		SyncPRs:       syncPRs,
	}

	if err := s.db.Create(&conn).Error; err != nil {
		return nil, common.Internal("Failed to create GitHub connection")
	}
	r := s.toResponse(&conn)
	return &r, nil
}

func (s *GitHubService) Update(id uint64, req *GitHubUpdateRequest) (*GitHubResponse, error) {
	var conn model.GitHubConnection
	if err := s.db.First(&conn, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("GitHub connection not found")
		}
		return nil, common.Internal("Failed to get GitHub connection")
	}

	updates := map[string]interface{}{}
	if req.RepoOwner != nil {
		updates["repo_owner"] = *req.RepoOwner
	}
	if req.RepoName != nil {
		updates["repo_name"] = *req.RepoName
	}
	if req.AccessToken != nil {
		updates["access_token"] = *req.AccessToken
	}
	if req.WebhookSecret != nil {
		updates["webhook_secret"] = *req.WebhookSecret
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.SyncIssues != nil {
		updates["sync_issues"] = *req.SyncIssues
	}
	if req.SyncPRs != nil {
		updates["sync_prs"] = *req.SyncPRs
	}

	if err := s.db.Model(&conn).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to update GitHub connection")
	}

	s.db.First(&conn, id)
	r := s.toResponse(&conn)
	return &r, nil
}

func (s *GitHubService) Delete(id uint64) error {
	var conn model.GitHubConnection
	if err := s.db.First(&conn, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("GitHub connection not found")
		}
		return common.Internal("Failed to get GitHub connection")
	}
	if err := s.db.Delete(&conn).Error; err != nil {
		return common.Internal("Failed to delete GitHub connection")
	}
	return nil
}

// ======== Issue Sync ========

func (s *GitHubService) SyncIssues(connID uint64) ([]GitHubIssue, error) {
	var conn model.GitHubConnection
	if err := s.db.First(&conn, connID).Error; err != nil {
		return nil, common.NotFound("GitHub connection not found")
	}

	issues, err := s.fetchIssuesFromGitHub(conn.RepoOwner, conn.RepoName, conn.AccessToken)
	if err != nil {
		return nil, common.Internal(fmt.Sprintf("Failed to sync issues: %v", err))
	}

	// Import issues into the project
	for _, gi := range issues {
		s.importIssue(&conn, &gi)
	}

	// Update last sync time
	now := time.Now().Format("2006-01-02T15:04:05Z")
	s.db.Model(&conn).Update("last_sync_at", now)

	return issues, nil
}

func (s *GitHubService) importIssue(conn *model.GitHubConnection, gi *GitHubIssue) {
	// Check if already imported by external_id
	var existing model.Issue
	result := s.db.Where("external_id = ? AND project_id = ?",
		fmt.Sprintf("github:%d", gi.Number), conn.ProjectID).First(&existing)
	if result.Error == nil {
		return // Already synced
	}

	// Map GitHub state to project state
	stateMap := map[string]string{
		"open":   "open",
		"closed": "closed",
	}
	state := stateMap[gi.State]
	if state == "" {
		state = "open"
	}

	// Find state ID
	var stateModel model.State
	if err := s.db.Where("project_id = ? AND name = ?", conn.ProjectID, state).First(&stateModel).Error; err != nil {
		return // Can't find matching state, skip
	}

	priority := "medium"
	for _, label := range gi.Labels {
		l := strings.ToLower(label)
		if l == "bug" || l == "urgent" || l == "critical" {
			priority = "urgent"
			break
		} else if l == "enhancement" || l == "feature" {
			priority = "high"
		}
	}

	externalID := fmt.Sprintf("github:%d", gi.Number)
	source := "github"

	issue := model.Issue{
		Name:           gi.Title,
		DescriptionHTML: fmt.Sprintf("<p>%s</p>\n<p><a href=\"%s\">View on GitHub</a></p>", gi.Body, gi.HTMLURL),
		Priority:       priority,
		ProjectID:      conn.ProjectID,
		WorkspaceID:    conn.WorkspaceID,
		StateID:        stateModel.ID,
		ExternalID:     &externalID,
		ExternalSource: &source,
	}
	s.db.Create(&issue)
}

// fetchIssuesFromGitHub calls GitHub REST API to get open issues.
func (s *GitHubService) fetchIssuesFromGitHub(owner, repo, token string) ([]GitHubIssue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=all&per_page=100", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "reqmanpy")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}

	var rawIssues []map[string]interface{}
	if err := json.Unmarshal(body, &rawIssues); err != nil {
		return nil, err
	}

	var issues []GitHubIssue
	for _, raw := range rawIssues {
		// Skip pull requests (they also appear in issues endpoint)
		if _, hasPR := raw["pull_request"]; hasPR {
			continue
		}

		labels := []string{}
		if labelArr, ok := raw["labels"].([]interface{}); ok {
			for _, l := range labelArr {
				if lm, ok := l.(map[string]interface{}); ok {
					if name, ok := lm["name"].(string); ok {
						labels = append(labels, name)
					}
				}
			}
		}

		number := int(raw["number"].(float64))
		title, _ := raw["title"].(string)
		bodyStr, _ := raw["body"].(string)
		state, _ := raw["state"].(string)
		htmlURL, _ := raw["html_url"].(string)
		createdAt, _ := raw["created_at"].(string)
		updatedAt, _ := raw["updated_at"].(string)

		issues = append(issues, GitHubIssue{
			Number:    number,
			Title:     title,
			Body:      bodyStr,
			State:     state,
			Labels:    labels,
			HTMLURL:   htmlURL,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}
	return issues, nil
}

// ======== Webhook Handler ========

// WebhookEvent represents a GitHub webhook payload.
type WebhookEvent struct {
	Action     string `json:"action"`
	Issue      struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		State  string `json:"state"`
		HTMLURL string `json:"html_url"`
	} `json:"issue"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// HandleWebhook processes incoming GitHub webhook events.
func (s *GitHubService) HandleWebhook(connID uint64, rawBody []byte, eventType string) (map[string]interface{}, error) {
	var conn model.GitHubConnection
	if err := s.db.First(&conn, connID).Error; err != nil {
		return nil, common.NotFound("GitHub connection not found")
	}

	var event WebhookEvent
	if err := json.Unmarshal(rawBody, &event); err != nil {
		return map[string]interface{}{"status": "received", "message": "unparseable payload"}, nil
	}

	switch eventType {
	case "issues":
		return s.handleIssueWebhook(&conn, &event)
	case "ping":
		return map[string]interface{}{"status": "pong"}, nil
	default:
		return map[string]interface{}{"status": "received", "event": eventType}, nil
	}
}

func (s *GitHubService) handleIssueWebhook(conn *model.GitHubConnection, event *WebhookEvent) (map[string]interface{}, error) {
	externalID := fmt.Sprintf("github:%d", event.Issue.Number)

	switch event.Action {
	case "opened", "reopened":
		issue := &GitHubIssue{
			Number:  event.Issue.Number,
			Title:   event.Issue.Title,
			Body:    event.Issue.Body,
			State:   "open",
			HTMLURL: event.Issue.HTMLURL,
		}
		s.importIssue(conn, issue)
		return map[string]interface{}{"status": "synced", "action": event.Action}, nil

	case "closed":
		// Update state to closed
		var issue model.Issue
		if err := s.db.Where("external_id = ? AND project_id = ?", externalID, conn.ProjectID).First(&issue).Error; err != nil {
			return map[string]interface{}{"status": "not_found"}, nil
		}
		var closedState model.State
		if err := s.db.Where("project_id = ? AND name = ?", conn.ProjectID, "closed").First(&closedState).Error; err == nil {
			s.db.Model(&issue).Update("state_id", closedState.ID)
		}
		return map[string]interface{}{"status": "updated", "action": event.Action}, nil

	default:
		return map[string]interface{}{"status": "ignored", "action": event.Action}, nil
	}
}

// ======== Helpers ========

func (s *GitHubService) toResponse(conn *model.GitHubConnection) GitHubResponse {
	return GitHubResponse{
		ID:            conn.ID,
		WorkspaceID:   conn.WorkspaceID,
		ProjectID:     conn.ProjectID,
		RepoOwner:     conn.RepoOwner,
		RepoName:      conn.RepoName,
		IsEnabled:     conn.IsEnabled,
		SyncIssues:    conn.SyncIssues,
		SyncPRs:       conn.SyncPRs,
		LastSyncAt:    conn.LastSyncAt,
		WebhookID:     conn.WebhookID,
		CreatedAt:     conn.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     conn.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
