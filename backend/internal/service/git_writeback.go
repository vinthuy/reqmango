package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// PRContext carries the top-level webhook fields that are not part of the
// pull_request object itself.
type PRContext struct {
	Action string
	Number int
	Repo   string // owner/name
}

// PRLinkResult describes what a pull-request event did to one issue.
type PRLinkResult struct {
	IssueID   uint64 `json:"issue_id"`
	Key       string `json:"key"`
	NewLink   bool   `json:"new_link"`
	Completed bool   `json:"completed"`
}

// resolveProjectIssues maps issue keys (e.g. MOBILE-12) onto issues of the
// project, ignoring keys whose prefix is not the project's identifier.
func (s *GitService) resolveProjectIssues(projectID uint64, keys []string) ([]model.Issue, []string) {
	var project model.Project
	if err := s.db.Select("id", "identifier").First(&project, projectID).Error; err != nil {
		return nil, nil
	}
	seen := map[uint64]bool{}
	var issues []model.Issue
	var matched []string
	for _, key := range keys {
		i := strings.LastIndex(key, "-")
		if i <= 0 || !strings.EqualFold(key[:i], project.Identifier) {
			continue
		}
		seq, err := strconv.ParseUint(key[i+1:], 10, 64)
		if err != nil || seq == 0 {
			continue
		}
		var issue model.Issue
		if err := s.db.Where("project_id = ? AND sequence_id = ?", projectID, seq).First(&issue).Error; err != nil {
			continue
		}
		if seen[issue.ID] {
			continue
		}
		seen[issue.ID] = true
		issues = append(issues, issue)
		matched = append(matched, strings.ToUpper(project.Identifier)+"-"+strconv.FormatUint(seq, 10))
	}
	return issues, matched
}

// completeIssue moves an issue into the project's first completed state and
// records the transition. It is a no-op for issues already completed or cancelled.
func (s *GitService) completeIssue(tx *gorm.DB, issue *model.Issue, reason string) (bool, error) {
	var current model.State
	if err := tx.First(&current, issue.StateID).Error; err == nil {
		if current.Group == common.StateGroupCompleted || current.Group == common.StateGroupCancelled {
			return false, nil
		}
	}
	var done model.State
	if err := tx.Where("project_id = ? AND \"group\" = ?", issue.ProjectID, common.StateGroupCompleted).
		Order("sequence ASC, id ASC").First(&done).Error; err != nil {
		return false, nil
	}
	now := time.Now()
	if err := tx.Model(&model.Issue{}).Where("id = ?", issue.ID).
		Updates(map[string]interface{}{"state_id": done.ID, "completed_at": now}).Error; err != nil {
		return false, err
	}
	field := "state"
	ov, nv := current.Name, done.Name
	act := &model.IssueActivity{IssueID: &issue.ID, Verb: "updated", Field: &field, OldValue: &ov, NewValue: &nv, Comment: &reason}
	if err := tx.Create(act).Error; err != nil {
		return false, err
	}
	issue.StateID = done.ID
	return true, nil
}

func gitActivity(tx *gorm.DB, issueID uint64, verb, url, comment string, actorID *uint64) error {
	field := "git_pr"
	act := &model.IssueActivity{IssueID: &issueID, Verb: verb, Field: &field, NewValue: &url, Comment: &comment, ActorID: actorID}
	if actorID != nil {
		act.CreatedByID = actorID
	}
	return tx.Create(act).Error
}

// issueURL builds a deep link to the issue when APP_BASE_URL is configured.
func (s *GitService) issueURL(issue *model.Issue) string {
	base := strings.TrimRight(os.Getenv("APP_BASE_URL"), "/")
	if base == "" {
		return ""
	}
	var ws model.Workspace
	if err := s.db.Select("workspaces.id", "workspaces.slug").Joins("JOIN projects ON projects.workspace_id = workspaces.id").
		Where("projects.id = ?", issue.ProjectID).First(&ws).Error; err != nil {
		return ""
	}
	return fmt.Sprintf("%s/workspace/%s/project/%d/issues/%d", base, ws.Slug, issue.ProjectID, issue.ID)
}

func githubAPIBase() string {
	if v := strings.TrimRight(os.Getenv("GITHUB_API_URL"), "/"); v != "" {
		return v
	}
	return "https://api.github.com"
}

var repoFromURLRe = regexp.MustCompile(`github\.com[/:]([^/\s]+)/([^/\s]+?)(?:\.git)?/?$`)

func repoFullName(integration *model.GitIntegration, ctxRepo string) string {
	if ctxRepo != "" {
		return ctxRepo
	}
	if m := repoFromURLRe.FindStringSubmatch(integration.RepoURL); m != nil {
		return m[1] + "/" + m[2]
	}
	if strings.Count(integration.RepoName, "/") == 1 {
		return integration.RepoName
	}
	return ""
}

// postPRComment writes a comment on the GitHub pull request. Failures are
// logged; they never fail the webhook.
func postPRComment(integration *model.GitIntegration, repo string, number int, body string) {
	if integration.Provider != "" && !strings.EqualFold(integration.Provider, "github") {
		return
	}
	if integration.AccessToken == "" || repo == "" || number <= 0 {
		return
	}
	payload, _ := json.Marshal(map[string]string{"body": body})
	url := fmt.Sprintf("%s/repos/%s/issues/%d/comments", githubAPIBase(), repo, number)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+integration.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[git] PR comment on %s#%d failed: %v", repo, number, err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("[git] PR comment on %s#%d returned %d", repo, number, resp.StatusCode)
	}
}

func (s *GitService) issueRef(issue *model.Issue, key string) string {
	if u := s.issueURL(issue); u != "" {
		return fmt.Sprintf("[%s %s](%s)", key, issue.Name, u)
	}
	return fmt.Sprintf("%s %s", key, issue.Name)
}

var githubPRURLRe = regexp.MustCompile(`^https?://[^/]+/([^/]+)/([^/]+)/pull/(\d+)`)

// LinkPullRequest links a PR URL to an issue on behalf of a project member.
func (s *GitService) LinkPullRequest(issueID uint64, user *model.User, prURL, title string) (*model.GitIssueLink, error) {
	prURL = strings.TrimSpace(prURL)
	m := githubPRURLRe.FindStringSubmatch(prURL)
	if m == nil {
		return nil, common.BadRequest("url must be a pull request URL like https://github.com/owner/repo/pull/123")
	}
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}
	if !user.IsSuperuser {
		var member model.ProjectMember
		if err := s.db.Where("project_id = ? AND user_id = ? AND is_active = ?", issue.ProjectID, user.ID, true).First(&member).Error; err != nil || member.Role < common.RoleMember {
			return nil, common.Forbidden("Access denied: insufficient project role")
		}
	}
	gitID := fmt.Sprintf("%s/%s#%s", m[1], m[2], m[3])
	title = strings.TrimSpace(title)
	if title == "" {
		title = fmt.Sprintf("%s/%s#%s", m[1], m[2], m[3])
	}
	var integrationID uint64
	if integration, err := s.GetIntegration(issue.ProjectID); err == nil {
		integrationID = integration.ID
	}
	var existing model.GitIssueLink
	isNew := s.db.Where("issue_id = ? AND git_type = ? AND git_id = ?", issue.ID, "pull_request", gitID).First(&existing).Error != nil
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if isNew {
			existing = model.GitIssueLink{IssueID: issue.ID, GitType: "pull_request", GitID: gitID, GitURL: prURL, GitTitle: title, GitState: "open", GitAuthor: user.DisplayName, IntegrationID: integrationID}
			if err := tx.Create(&existing).Error; err != nil {
				return err
			}
			return gitActivity(tx, issue.ID, "git_linked", prURL, title, &user.ID)
		}
		return tx.Model(&existing).Updates(map[string]interface{}{"git_url": prURL, "git_title": title}).Error
	})
	if err != nil {
		return nil, common.Internal("Failed to link pull request")
	}
	return &existing, nil
}
