package service

import (
	"fmt"
	"time"

	"github.com/reqmango/backend/internal/client"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type CommentService struct {
	db              *gorm.DB
	notificationSvc *NotificationService
	agentClient     *client.AgentClient
	automationSvc   *AutomationService
}

func NewCommentService(db *gorm.DB, notificationSvc *NotificationService) *CommentService {
	return &CommentService{db: db, notificationSvc: notificationSvc}
}

// SetAgentService sets the agent client for @agent-name mention handling.
func (s *CommentService) SetAgentService(agentClient *client.AgentClient) {
	s.agentClient = agentClient
}

// SetAutomationService sets the automation service for comment_added triggers.
func (s *CommentService) SetAutomationService(automationSvc *AutomationService) {
	s.automationSvc = automationSvc
}

func (s *CommentService) Create(issueID, authorID uint64, body string, parentID *uint64) (*model.Comment, error) {
	var issue model.Issue
	if err := s.db.Preload("Project").First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}
	c := model.Comment{IssueID: issueID, AuthorID: &authorID, Body: body, ParentID: parentID}
	if err := s.db.Create(&c).Error; err != nil {
		return nil, common.Internal("Failed to create comment")
	}
	s.db.Preload("Author").First(&c, c.ID)

	// Trigger notification for comment
	if s.notificationSvc != nil {
		var assignees []model.IssueAssignee
		s.db.Where("issue_id = ?", issueID).Find(&assignees)
		if len(assignees) > 0 {
			recipientIDs := make([]uint64, 0, len(assignees))
			for _, a := range assignees {
				if a.UserID != authorID {
					recipientIDs = append(recipientIDs, a.UserID)
				}
			}
			if len(recipientIDs) > 0 {
				title := fmt.Sprintf("新评论: %s", issue.Name)
				message := fmt.Sprintf("工作项 #%d 收到新评论", issue.SequenceID)
				issueIDPtr := issueID
				projectIDPtr := issue.ProjectID
				s.notificationSvc.TriggerNotificationsBulk(s.db, "issue_commented", title, message, recipientIDs, &authorID, &projectIDPtr, &issueIDPtr)
				for _, rid := range recipientIDs {
					SSE.NotifyUser(rid, "issue_commented", title, message)
				}
			}
		}

		mentioned := parseMentions(body)
		if len(mentioned) > 0 {
			var users []model.User
			s.db.Where("username IN ?", mentioned).Find(&users)
			mentionIDs := make([]uint64, 0, len(users))
			for _, u := range users {
				if u.ID != authorID {
					mentionIDs = append(mentionIDs, u.ID)
				}
			}
			if len(mentionIDs) > 0 {
				title := fmt.Sprintf("@提及: %s", issue.Name)
				msg := fmt.Sprintf("你在工作项 #%d 的评论中被 @%s 提及", issue.SequenceID, c.Author.Username)
				issueIDPtr := issueID
				projectIDPtr := issue.ProjectID
				s.notificationSvc.TriggerNotificationsBulk(s.db, "issue_mentioned", title, msg, mentionIDs, &authorID, &projectIDPtr, &issueIDPtr)
			}

			if s.agentClient != nil {
				var agents []model.Agent
				s.db.Where("workspace_id = ? AND name IN ? AND status = 'active'", issue.Project.WorkspaceID, mentioned).Find(&agents)
				for _, agent := range agents {
					go func(a model.Agent) {
						s.agentClient.HandleMention(a.WorkspaceID, a.ID, c.ID, authorID, body, issue.Name, &issueID)
					}(agent)
				}
			}
		}
	}

	if s.automationSvc != nil {
		event := Event{
			Type:      "comment.added",
			ProjectID: issue.ProjectID,
			IssueID:   issueID,
			Context: map[string]interface{}{
				"issue_id":   issueID,
				"project_id": issue.ProjectID,
				"author_id":  authorID,
				"comment":    body,
			},
			Timestamp: time.Now(),
		}
		go func() {
			if err := s.automationSvc.PublishEvent(event); err != nil {
				fmt.Printf("[CommentService] Failed to publish automation event: %v\n", err)
			}
		}()
	}

	return &c, nil
}

func parseMentions(text string) []string {
	result := make([]string, 0)
	seen := make(map[string]bool)
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		if runes[i] == '@' && (i == 0 || runes[i-1] == ' ' || runes[i-1] == '\n') {
			start := i + 1
			j := start
			for j < len(runes) {
				c := runes[j]
				isValid := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' ||
					(c >= 0x4E00 && c <= 0x9FFF) || (c >= 0x3400 && c <= 0x4DBF) || (c >= 0xAC00 && c <= 0xD7AF)
				if !isValid {
					break
				}
				j++
			}
			if j > start {
				name := string(runes[start:j])
				if !seen[name] {
					seen[name] = true
					result = append(result, name)
				}
			}
			i = j
		} else {
			i++
		}
	}
	return result
}

func (s *CommentService) ListByIssue(issueID uint64, page, pageSize int) ([]model.Comment, int64, error) {
	var total int64
	s.db.Model(&model.Comment{}).Where("issue_id = ?", issueID).Count(&total)
	var comments []model.Comment
	offset := (page - 1) * pageSize
	if err := s.db.Preload("Author").Where("issue_id = ?", issueID).
		Order("created_at ASC").Limit(pageSize).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, common.Internal("Failed to list comments")
	}
	if comments == nil { comments = []model.Comment{} }
	return comments, total, nil
}

func (s *CommentService) Get(id uint64) (*model.Comment, error) {
	var c model.Comment
	if err := s.db.Preload("Author").First(&c, id).Error; err != nil {
		return nil, common.NotFound("Comment not found")
	}
	return &c, nil
}

func (s *CommentService) Update(id, userID uint64, body string) (*model.Comment, error) {
	var c model.Comment
	if err := s.db.Preload("Author").First(&c, id).Error; err != nil {
		return nil, common.NotFound("Comment not found")
	}
	if c.AuthorID == nil || *c.AuthorID != userID {
		return nil, common.Forbidden("Forbidden")
	}

	c.Body = body
	s.db.Save(&c)

	if s.notificationSvc != nil {
		var issue model.Issue
		if err := s.db.Preload("Project").First(&issue, c.IssueID).Error; err == nil {
			mentioned := parseMentions(body)
			if len(mentioned) > 0 {
				var users []model.User
				s.db.Where("username IN ?", mentioned).Find(&users)
				mentionIDs := make([]uint64, 0, len(users))
				for _, u := range users {
					if u.ID != userID {
						mentionIDs = append(mentionIDs, u.ID)
					}
				}
				if len(mentionIDs) > 0 {
					title := fmt.Sprintf("@提及: %s", issue.Name)
					msg := fmt.Sprintf("你在工作项 #%d 的评论中被 @%s 提及", issue.SequenceID, c.Author.Username)
					issueIDPtr := c.IssueID
					projectIDPtr := issue.ProjectID
					s.notificationSvc.TriggerNotificationsBulk(s.db, "issue_mentioned", title, msg, mentionIDs, &userID, &projectIDPtr, &issueIDPtr)
				}

				if s.agentClient != nil {
					var agents []model.Agent
					s.db.Where("workspace_id = ? AND name IN ? AND status = 'active'", issue.Project.WorkspaceID, mentioned).Find(&agents)
					for _, agent := range agents {
						go func(a model.Agent) {
							s.agentClient.HandleMention(a.WorkspaceID, a.ID, c.ID, userID, body, issue.Name, &c.IssueID)
						}(agent)
					}
				}
			}
		}
	}

	return &c, nil
}

func (s *CommentService) Delete(id, userID uint64) error {
	var c model.Comment
	if err := s.db.First(&c, id).Error; err != nil {
		return common.NotFound("Comment not found")
	}
	if c.AuthorID == nil || *c.AuthorID != userID {
		return common.Forbidden("Forbidden")
	}
	return s.db.Delete(&model.Comment{}, id).Error
}

func (s *CommentService) Resolve(id uint64) (*model.Comment, error) {
	var c model.Comment
	if err := s.db.First(&c, id).Error; err != nil {
		return nil, common.NotFound("Comment not found")
	}
	c.IsResolved = true
	s.db.Save(&c)
	return &c, nil
}

func (s *CommentService) Unresolve(id uint64) (*model.Comment, error) {
	var c model.Comment
	if err := s.db.First(&c, id).Error; err != nil {
		return nil, common.NotFound("Comment not found")
	}
	c.IsResolved = false
	s.db.Save(&c)
	return &c, nil
}
