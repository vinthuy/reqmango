package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type GitService struct {
	db *gorm.DB
}

func NewGitService(db *gorm.DB) *GitService {
	return &GitService{db: db}
}

func (s *GitService) CreateIntegration(projectID uint64, provider, repoURL, repoName, accessToken, webhookSecret string) (*model.GitIntegration, error) {
	integration := &model.GitIntegration{
		ProjectID:     projectID,
		Provider:      provider,
		RepoURL:       repoURL,
		RepoName:      repoName,
		AccessToken:   accessToken,
		WebhookSecret: webhookSecret,
		Active:        true,
		SyncPRs:       true,
		SyncCommits:   true,
		SyncBranches:  false,
	}

	if err := s.db.Create(integration).Error; err != nil {
		return nil, common.Internal("Failed to create git integration")
	}

	return integration, nil
}

func (s *GitService) GetIntegration(projectID uint64) (*model.GitIntegration, error) {
	var integration model.GitIntegration
	if err := s.db.Where("project_id = ?", projectID).First(&integration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Git integration not found")
		}
		return nil, common.Internal("Failed to fetch git integration")
	}
	return &integration, nil
}

func (s *GitService) UpdateIntegration(projectID uint64, updates map[string]interface{}) (*model.GitIntegration, error) {
	var integration model.GitIntegration
	if err := s.db.Where("project_id = ?", projectID).First(&integration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Git integration not found")
		}
		return nil, common.Internal("Failed to fetch git integration")
	}

	if err := s.db.Model(&integration).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to update git integration")
	}

	return &integration, nil
}

func (s *GitService) DeleteIntegration(projectID uint64) error {
	result := s.db.Where("project_id = ?", projectID).Delete(&model.GitIntegration{})
	if result.RowsAffected == 0 {
		return common.NotFound("Git integration not found")
	}
	return result.Error
}

var issueKeyRegex = regexp.MustCompile(`([A-Z]+-\d+)`)

func (s *GitService) ParseIssueKey(commitMessage string) []string {
	return issueKeyRegex.FindAllString(commitMessage, -1)
}

var smartCommitRegex = regexp.MustCompile(`(fixes|closes|resolve|ref)\s+([A-Z]+-\d+)`)

func (s *GitService) ParseSmartCommit(commitMessage string) []map[string]string {
	matches := smartCommitRegex.FindAllStringSubmatch(strings.ToLower(commitMessage), -1)
	result := make([]map[string]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, map[string]string{
			"action": strings.ToLower(match[1]),
			"key":    strings.ToUpper(match[2]),
		})
	}
	return result
}

func (s *GitService) LinkIssueToGit(issueID uint64, gitType, gitID, gitURL, gitTitle, gitState, gitAuthor, gitBranch string, integrationID uint64) error {
	var existing model.GitIssueLink
	if err := s.db.Where("issue_id = ? AND git_type = ? AND git_id = ?", issueID, gitType, gitID).First(&existing).Error; err == nil {
		return s.db.Model(&existing).Updates(map[string]interface{}{
			"git_url":  gitURL,
			"git_title": gitTitle,
			"git_state": gitState,
			"git_author": gitAuthor,
			"git_branch": gitBranch,
		}).Error
	}

	return s.db.Create(&model.GitIssueLink{
		IssueID:       issueID,
		GitType:       gitType,
		GitID:         gitID,
		GitURL:        gitURL,
		GitTitle:      gitTitle,
		GitState:      gitState,
		GitAuthor:     gitAuthor,
		GitBranch:     gitBranch,
		IntegrationID: integrationID,
	}).Error
}

func (s *GitService) GetIssueGitLinks(issueID uint64) ([]model.GitIssueLink, error) {
	var links []model.GitIssueLink
	if err := s.db.Where("issue_id = ?", issueID).Order("updated_at DESC").Find(&links).Error; err != nil {
		return nil, common.Internal("Failed to fetch git links")
	}
	return links, nil
}

func (s *GitService) HandlePushEvent(projectID uint64, commits []map[string]interface{}) error {
	integration, err := s.GetIntegration(projectID)
	if err != nil {
		return err
	}

	if !integration.SyncCommits {
		return nil
	}

	for _, commit := range commits {
		message := fmt.Sprintf("%v", commit["message"])
		smartCommits := s.ParseSmartCommit(message)

		for _, sc := range smartCommits {
			var issue model.Issue
			if err := s.db.Where("sequence_id = ? AND project_id = ?", parseSequenceID(sc["key"]), projectID).First(&issue).Error; err != nil {
				continue
			}

			commitURL := fmt.Sprintf("%v", commit["url"])
			author := ""
			if authorMap, ok := commit["author"].(map[string]interface{}); ok {
				author = fmt.Sprintf("%v", authorMap["name"])
			}

			s.LinkIssueToGit(issue.ID, "commit", commitURL, commitURL, message, "pushed", author, "", integration.ID)

			if sc["action"] == "fixes" || sc["action"] == "closes" {
				s.db.Model(&issue).Update("state_id", s.getCompletedStateID(projectID))
			}
		}
	}

	return nil
}

func (s *GitService) HandlePullRequestEvent(projectID uint64, pr map[string]interface{}) error {
	integration, err := s.GetIntegration(projectID)
	if err != nil {
		return err
	}

	if !integration.SyncPRs {
		return nil
	}

	prID := fmt.Sprintf("%v", pr["id"])
	prURL := fmt.Sprintf("%v", pr["html_url"])
	prTitle := fmt.Sprintf("%v", pr["title"])
	prState := fmt.Sprintf("%v", pr["state"])

	prAuthor := ""
	if userMap, ok := pr["user"].(map[string]interface{}); ok {
		prAuthor = fmt.Sprintf("%v", userMap["login"])
	}

	prBranch := ""
	if headMap, ok := pr["head"].(map[string]interface{}); ok {
		prBranch = fmt.Sprintf("%v", headMap["ref"])
	}

	title := fmt.Sprintf("%v", pr["title"])
	issueKeys := s.ParseIssueKey(title)

	for _, key := range issueKeys {
		var issue model.Issue
		if err := s.db.Where("sequence_id = ? AND project_id = ?", parseSequenceID(key), projectID).First(&issue).Error; err != nil {
			continue
		}

		s.LinkIssueToGit(issue.ID, "pull_request", prID, prURL, prTitle, prState, prAuthor, prBranch, integration.ID)

		if prState == "closed" && pr["merged"] == true {
			s.db.Model(&issue).Update("state_id", s.getCompletedStateID(projectID))
		}
	}

	return nil
}

func parseSequenceID(key string) uint64 {
	parts := strings.Split(key, "-")
	if len(parts) != 2 {
		return 0
	}
	var seqID uint64
	fmt.Sscanf(parts[1], "%d", &seqID)
	return seqID
}

func (s *GitService) getCompletedStateID(projectID uint64) uint64 {
	var state model.State
	s.db.Joins("JOIN workflows ON workflows.id = states.workflow_id").
		Where("workflows.project_id = ? AND states.group = ?", projectID, common.StateGroupCompleted).
		First(&state)
	return state.ID
}