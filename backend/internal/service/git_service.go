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

var issueKeyRegex = regexp.MustCompile(`(?i)\b([a-z][a-z0-9]*-\d+)\b`)

// ParseIssueKey returns the distinct upper-cased issue keys found in text
// (titles, branch names such as feature/mobile-12-login, PR bodies).
func (s *GitService) ParseIssueKey(text string) []string {
	seen := map[string]bool{}
	var keys []string
	for _, k := range issueKeyRegex.FindAllString(text, -1) {
		k = strings.ToUpper(k)
		if !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	return keys
}

var smartCommitRegex = regexp.MustCompile(`(?i)\b(fix(?:es|ed)?|close[sd]?|resolve[sd]?|ref)\s+([a-z][a-z0-9]*-\d+)\b`)

func (s *GitService) ParseSmartCommit(commitMessage string) []map[string]string {
	matches := smartCommitRegex.FindAllStringSubmatch(commitMessage, -1)
	result := make([]map[string]string, 0, len(matches))
	for _, match := range matches {
		action := strings.ToLower(match[1])
		switch {
		case strings.HasPrefix(action, "fix"):
			action = "fixes"
		case strings.HasPrefix(action, "close"):
			action = "closes"
		case strings.HasPrefix(action, "resolve"):
			action = "resolves"
		}
		result = append(result, map[string]string{
			"action": action,
			"key":    strings.ToUpper(match[2]),
		})
	}
	return result
}

func (s *GitService) LinkIssueToGit(issueID uint64, gitType, gitID, gitURL, gitTitle, gitState, gitAuthor, gitBranch string, integrationID uint64) error {
	var existing model.GitIssueLink
	if err := s.db.Where("issue_id = ? AND git_type = ? AND git_id = ?", issueID, gitType, gitID).First(&existing).Error; err == nil {
		return s.db.Model(&existing).Updates(map[string]interface{}{
			"git_url":    gitURL,
			"git_title":  gitTitle,
			"git_state":  gitState,
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
		keys := make([]string, 0, len(smartCommits))
		closing := map[string]bool{}
		for _, sc := range smartCommits {
			keys = append(keys, sc["key"])
			if sc["action"] != "ref" {
				closing[sc["key"]] = true
			}
		}
		issues, matched := s.resolveProjectIssues(projectID, keys)

		commitURL := fmt.Sprintf("%v", commit["url"])
		author := ""
		if authorMap, ok := commit["author"].(map[string]interface{}); ok {
			author = fmt.Sprintf("%v", authorMap["name"])
		}
		firstLine := strings.SplitN(message, "\n", 2)[0]

		for i := range issues {
			issue := &issues[i]
			_ = s.LinkIssueToGit(issue.ID, "commit", commitURL, commitURL, message, "pushed", author, "", integration.ID)
			if closing[matched[i]] {
				_ = s.db.Transaction(func(tx *gorm.DB) error {
					_, err := s.completeIssue(tx, issue, "Commit: "+firstLine)
					return err
				})
			}
		}
	}

	return nil
}

// HandlePullRequestEvent links a PR to every issue referenced in its title,
// head branch or body, and completes those issues when the PR is merged.
func (s *GitService) HandlePullRequestEvent(projectID uint64, pr map[string]interface{}, pctx PRContext) ([]PRLinkResult, error) {
	integration, err := s.GetIntegration(projectID)
	if err != nil {
		return nil, err
	}

	if !integration.Active || !integration.SyncPRs {
		return nil, nil
	}

	str := func(v interface{}) string {
		if v == nil {
			return ""
		}
		return fmt.Sprintf("%v", v)
	}
	prID := str(pr["id"])
	prURL := str(pr["html_url"])
	prTitle := str(pr["title"])
	prState := str(pr["state"])
	merged := pr["merged"] == true
	if merged {
		prState = "merged"
	}
	number := pctx.Number
	if number == 0 {
		if n, ok := pr["number"].(float64); ok {
			number = int(n)
		}
	}

	prAuthor := ""
	if userMap, ok := pr["user"].(map[string]interface{}); ok {
		prAuthor = str(userMap["login"])
	}

	prBranch := ""
	if headMap, ok := pr["head"].(map[string]interface{}); ok {
		prBranch = str(headMap["ref"])
	}

	keys := s.ParseIssueKey(prTitle + "\n" + prBranch + "\n" + str(pr["body"]))
	issues, matched := s.resolveProjectIssues(projectID, keys)
	repo := repoFullName(integration, pctx.Repo)
	label := prTitle
	if number > 0 {
		label = fmt.Sprintf("PR #%d: %s", number, prTitle)
	}

	var results []PRLinkResult
	var linkedRefs, completedRefs []string
	for i := range issues {
		issue := &issues[i]
		var existing model.GitIssueLink
		isNew := s.db.Where("issue_id = ? AND git_type = ? AND git_id = ?", issue.ID, "pull_request", prID).First(&existing).Error != nil
		wasMerged := !isNew && existing.GitState == "merged"

		res := PRLinkResult{IssueID: issue.ID, Key: matched[i], NewLink: isNew}
		err := s.db.Transaction(func(tx *gorm.DB) error {
			if isNew {
				if err := tx.Create(&model.GitIssueLink{IssueID: issue.ID, GitType: "pull_request", GitID: prID, GitURL: prURL, GitTitle: prTitle, GitState: prState, GitAuthor: prAuthor, GitBranch: prBranch, IntegrationID: integration.ID}).Error; err != nil {
					return err
				}
				if err := gitActivity(tx, issue.ID, "git_linked", prURL, label, nil); err != nil {
					return err
				}
			} else if err := tx.Model(&existing).Updates(map[string]interface{}{"git_url": prURL, "git_title": prTitle, "git_state": prState, "git_author": prAuthor, "git_branch": prBranch}).Error; err != nil {
				return err
			}
			if merged && !wasMerged {
				if err := gitActivity(tx, issue.ID, "git_merged", prURL, label, nil); err != nil {
					return err
				}
				done, err := s.completeIssue(tx, issue, "Merged "+label)
				if err != nil {
					return err
				}
				res.Completed = done
			}
			return nil
		})
		if err != nil {
			return results, common.Internal("Failed to record pull request")
		}
		results = append(results, res)
		if isNew {
			linkedRefs = append(linkedRefs, s.issueRef(issue, matched[i]))
		}
		if res.Completed {
			completedRefs = append(completedRefs, s.issueRef(issue, matched[i]))
		}
	}

	var parts []string
	if len(linkedRefs) > 0 {
		parts = append(parts, "🔗 Linked to ReqMango: "+strings.Join(linkedRefs, ", "))
	}
	if len(completedRefs) > 0 {
		parts = append(parts, "✅ Merged — marked as done in ReqMango: "+strings.Join(completedRefs, ", "))
	}
	if len(parts) > 0 {
		body := strings.Join(parts, "\n\n")
		go postPRComment(integration, repo, number, body)
	}

	return results, nil
}
