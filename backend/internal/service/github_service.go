package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type GitHubService struct {
	db *gorm.DB
}

func NewGitHubService(db *gorm.DB) *GitHubService {
	return &GitHubService{db: db}
}

// checkWorkspaceAdmin verifies that the caller is an active admin-level member
// of the workspace. Guards mutations against privilege escalation.
func (s *GitHubService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage GitHub connections")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage GitHub connections")
	}
	return nil
}

// ======== Request/Response types ========

type GitHubCreateRequest struct {
	ProjectID     uint64 `json:"project_id" binding:"required"`
	RepoOwner     string `json:"repo_owner" binding:"required"`
	RepoName      string `json:"repo_name" binding:"required"`
	AccessToken   string `json:"access_token"`
	WebhookSecret string `json:"webhook_secret"`
	IsEnabled     *bool  `json:"is_enabled"`
	SyncIssues    *bool  `json:"sync_issues"`
	SyncPRs       *bool  `json:"sync_prs"`
}

type GitHubUpdateRequest struct {
	RepoOwner     *string `json:"repo_owner"`
	RepoName      *string `json:"repo_name"`
	AccessToken   *string `json:"access_token"`
	WebhookSecret *string `json:"webhook_secret"`
	IsEnabled     *bool   `json:"is_enabled"`
	SyncIssues    *bool   `json:"sync_issues"`
	SyncPRs       *bool   `json:"sync_prs"`
}

type GitHubResponse struct {
	ID          uint64  `json:"id"`
	WorkspaceID uint64  `json:"workspace_id"`
	ProjectID   uint64  `json:"project_id"`
	RepoOwner   string  `json:"repo_owner"`
	RepoName    string  `json:"repo_name"`
	IsEnabled   bool    `json:"is_enabled"`
	SyncIssues  bool    `json:"sync_issues"`
	SyncPRs     bool    `json:"sync_prs"`
	LastSyncAt  *string `json:"last_sync_at"`
	WebhookID   *uint64 `json:"webhook_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
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

func (s *GitHubService) Create(workspaceID uint64, callerID uint64, req *GitHubCreateRequest) (*GitHubResponse, error) {
	if err := s.checkWorkspaceAdmin(workspaceID, callerID); err != nil {
		return nil, err
	}
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

func (s *GitHubService) Update(id uint64, callerID uint64, req *GitHubUpdateRequest) (*GitHubResponse, error) {
	var conn model.GitHubConnection
	if err := s.db.First(&conn, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("GitHub connection not found")
		}
		return nil, common.Internal("Failed to get GitHub connection")
	}
	if err := s.checkWorkspaceAdmin(conn.WorkspaceID, callerID); err != nil {
		return nil, err
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

func (s *GitHubService) Delete(id uint64, callerID uint64) error {
	var conn model.GitHubConnection
	if err := s.db.First(&conn, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("GitHub connection not found")
		}
		return common.Internal("Failed to get GitHub connection")
	}
	if err := s.checkWorkspaceAdmin(conn.WorkspaceID, callerID); err != nil {
		return err
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
		Name:            gi.Title,
		DescriptionHTML: fmt.Sprintf("<p>%s</p>\n<p><a href=\"%s\">View on GitHub</a></p>", gi.Body, gi.HTMLURL),
		Priority:        priority,
		ProjectID:       conn.ProjectID,
		WorkspaceID:     conn.WorkspaceID,
		StateID:         stateModel.ID,
		ExternalID:      &externalID,
		ExternalSource:  &source,
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
	req.Header.Set("User-Agent", "reqmango")
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
	Action string `json:"action"`
	Issue  struct {
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		State   string `json:"state"`
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
		ID:          conn.ID,
		WorkspaceID: conn.WorkspaceID,
		ProjectID:   conn.ProjectID,
		RepoOwner:   conn.RepoOwner,
		RepoName:    conn.RepoName,
		IsEnabled:   conn.IsEnabled,
		SyncIssues:  conn.SyncIssues,
		SyncPRs:     conn.SyncPRs,
		LastSyncAt:  conn.LastSyncAt,
		WebhookID:   conn.WebhookID,
		CreatedAt:   conn.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   conn.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ======== Developer Agent support (P4-001) ========
// The methods below wrap the GitHub REST API endpoints required by the
// Developer Agent to create a working branch, commit generated files, and
// open a pull request. They return typed results so callers (the developer
// agent service) can persist them on the DeveloperJob record.

// GitHubFileInput represents a single file to commit via the GitHub API.
type GitHubFileInput struct {
	Path    string `json:"path"`           // repository-relative path, e.g. "src/auth/login.go"
	Content string `json:"content"`        // raw file content; will be base64-encoded by the API
	Mode    string `json:"mode,omitempty"` // optional file mode hint ("100644" default)
}

// GitHubBranchRef describes a branch SHA reference returned by the Git refs API.
type GitHubBranchRef struct {
	Ref    string `json:"ref"`
	NodeID string `json:"node_id"`
	URL    string `json:"url"`
	Object struct {
		Type string `json:"type"`
		SHA  string `json:"sha"`
	} `json:"object"`
}

// GitHubCommitFileResponse is the response from POST /repos/{owner}/{repo}/contents/{path}.
type GitHubCommitFileResponse struct {
	Content struct {
		Path string `json:"path"`
		SHA  string `json:"sha"`
	} `json:"content"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

// GitHubPullRequest represents a pull request opened by the Developer Agent.
type GitHubPullRequest struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
	State   string `json:"state"`
	Head    struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

// GetConnection fetches a GitHubConnection row by ID.
// Used by the developer agent to resolve credentials and repo coordinates.
func (s *GitHubService) GetConnection(connID uint64) (*model.GitHubConnection, error) {
	var conn model.GitHubConnection
	if err := s.db.First(&conn, connID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("GitHub connection not found")
		}
		return nil, common.Internal("Failed to get GitHub connection")
	}
	return &conn, nil
}

// GetBranchSHA fetches the SHA that a branch currently points to.
func (s *GitHubService) GetBranchSHA(owner, repo, branch, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/%s", owner, repo, branch)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	s.applyGitHubHeaders(req, token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 404 {
		return "", common.NotFound(fmt.Sprintf("Branch '%s' not found in %s/%s", branch, owner, repo))
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub getBranchSHA error %d: %s", resp.StatusCode, string(body))
	}

	var ref GitHubBranchRef
	if err := json.Unmarshal(body, &ref); err != nil {
		return "", err
	}
	return ref.Object.SHA, nil
}

// CreateBranch creates a new branch pointing at the supplied base SHA.
// If baseSHA is empty, the current HEAD of baseBranch is used.
func (s *GitHubService) CreateBranch(owner, repo, baseBranch, newBranch, token, baseSHA string) (*GitHubBranchRef, error) {
	if baseSHA == "" {
		var err error
		baseSHA, err = s.GetBranchSHA(owner, repo, baseBranch, token)
		if err != nil {
			return nil, err
		}
	}

	payload := map[string]interface{}{
		"ref": "refs/heads/" + newBranch,
		"sha": baseSHA,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", owner, repo)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	s.applyGitHubHeaders(req, token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 422 {
		return nil, common.BadRequest(fmt.Sprintf("Branch '%s' already exists or SHA is invalid", newBranch))
	}
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("GitHub createBranch error %d: %s", resp.StatusCode, string(respBody))
	}

	var ref GitHubBranchRef
	if err := json.Unmarshal(respBody, &ref); err != nil {
		return nil, err
	}
	return &ref, nil
}

// CommitFile creates or updates a single file on the supplied branch.
// Returns the commit SHA so callers can chain multiple commits.
func (s *GitHubService) CommitFile(owner, repo, branch, path, content, commitMessage, token, existingFileSHA string) (*GitHubCommitFileResponse, error) {
	payload := map[string]interface{}{
		"message": commitMessage,
		"content": base64.StdEncoding.EncodeToString([]byte(content)),
		"branch":  branch,
		"path":    path,
	}
	if existingFileSHA != "" {
		payload["sha"] = existingFileSHA
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	s.applyGitHubHeaders(req, token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("GitHub commitFile error %d: %s", resp.StatusCode, string(respBody))
	}

	var out GitHubCommitFileResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CommitFiles commits multiple files in sequence on the given branch.
// It returns the SHA of the last commit (head of branch) and the count of
// files successfully committed. The Developer Agent uses this to push a
// batch of generated source files in a single logical change.
func (s *GitHubService) CommitFiles(owner, repo, branch, commitMessage, token string, files []GitHubFileInput) (string, int, error) {
	if len(files) == 0 {
		return "", 0, common.BadRequest("No files to commit")
	}
	lastSHA := ""
	committed := 0
	for _, f := range files {
		resp, err := s.CommitFile(owner, repo, branch, f.Path, f.Content, commitMessage, token, "")
		if err != nil {
			if lastSHA == "" {
				return "", committed, err
			}
			// Continue partial: caller may inspect committed count + error.
			return lastSHA, committed, err
		}
		lastSHA = resp.Commit.SHA
		committed++
	}
	return lastSHA, committed, nil
}

// CreatePullRequest opens a PR from headBranch into baseBranch.
func (s *GitHubService) CreatePullRequest(owner, repo, baseBranch, headBranch, title, body, token string) (*GitHubPullRequest, error) {
	payload := map[string]interface{}{
		"title": title,
		"body":  body,
		"head":  headBranch,
		"base":  baseBranch,
	}
	bodyBytes, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	s.applyGitHubHeaders(req, token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("GitHub createPullRequest error %d: %s", resp.StatusCode, string(respBody))
	}

	var pr GitHubPullRequest
	if err := json.Unmarshal(respBody, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// applyGitHubHeaders sets the standard headers required by the GitHub REST API.
func (s *GitHubService) applyGitHubHeaders(req *http.Request, token string) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "reqmango")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}
