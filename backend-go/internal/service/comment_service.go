package service

import (
	"fmt"

	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

type CommentService struct {
	db             *gorm.DB
	notificationSvc *NotificationService
}

func NewCommentService(db *gorm.DB, notificationSvc *NotificationService) *CommentService {
	return &CommentService{db: db, notificationSvc: notificationSvc}
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
			}
		}
	}

	return &c, nil
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

func (s *CommentService) Update(id uint64, body string) (*model.Comment, error) {
	var c model.Comment
	if err := s.db.First(&c, id).Error; err != nil {
		return nil, common.NotFound("Comment not found")
	}
	c.Body = body
	s.db.Save(&c)
	s.db.Preload("Author").First(&c, id)
	return &c, nil
}

func (s *CommentService) Delete(id uint64) error {
	if err := s.db.First(&model.Comment{}, id).Error; err != nil {
		return common.NotFound("Comment not found")
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
