package service

import (
	"errors"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type InitiativeService struct{ db *gorm.DB }

func NewInitiativeService(db *gorm.DB) *InitiativeService { return &InitiativeService{db} }

// checkWorkspaceAdmin verifies that the caller is an active admin-level member
// of the workspace. Guards mutations against privilege escalation.
func (s *InitiativeService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage initiatives")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage initiatives")
	}
	return nil
}

func (s *InitiativeService) Create(workspaceID uint64, req request.CreateInitiativeReq, userID uint64) (*model.Initiative, error) {
	if err := s.checkWorkspaceAdmin(workspaceID, userID); err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	initiative := model.Initiative{
		WorkspaceID: workspaceID, Name: req.Name, Description: req.Description,
		Color: req.Color, Status: status, CreatedByID: userID,
	}
	if req.TargetDate != "" {
		t, _ := time.Parse("2006-01-02", req.TargetDate)
		initiative.TargetDate = &t
	}
	if req.StartDate != "" {
		t, _ := time.Parse("2006-01-02", req.StartDate)
		initiative.StartDate = &t
	}

	tx := s.db.Begin()
	if err := tx.Create(&initiative).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if len(req.ProjectIDs) > 0 {
		for _, pid := range req.ProjectIDs {
			if err := tx.Create(&model.InitiativeProject{InitiativeID: initiative.ID, ProjectID: pid}).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	s.db.Preload("Projects").First(&initiative, initiative.ID)
	return &initiative, nil
}

func (s *InitiativeService) List(workspaceID uint64) ([]model.Initiative, error) {
	var initiatives []model.Initiative
	err := s.db.Where("workspace_id = ?", workspaceID).
		Preload("Projects").Order("sort_order ASC, created_at DESC").Find(&initiatives).Error
	return initiatives, err
}

func (s *InitiativeService) Search(workspaceID uint64, query string) ([]model.Initiative, error) {
	var initiatives []model.Initiative
	err := s.db.Where("workspace_id = ? AND name ILIKE ?", workspaceID, "%"+query+"%").
		Preload("Projects").Order("sort_order ASC, created_at DESC").Find(&initiatives).Error
	return initiatives, err
}

func (s *InitiativeService) Get(id uint64) (*model.Initiative, error) {
	var i model.Initiative
	err := s.db.Preload("Projects").First(&i, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound("Initiative not found")
	}
	return &i, err
}

func (s *InitiativeService) Update(id, callerID uint64, req request.UpdateInitiativeReq) (*model.Initiative, error) {
	var i model.Initiative
	if err := s.db.First(&i, id).Error; err != nil {
		return nil, err
	}
	if err := s.checkWorkspaceAdmin(i.WorkspaceID, callerID); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.TargetDate != nil {
		if *req.TargetDate == "" {
			updates["target_date"] = nil
		} else {
			t, _ := time.Parse("2006-01-02", *req.TargetDate)
			updates["target_date"] = &t
		}
	}
	if req.StartDate != nil {
		if *req.StartDate == "" {
			updates["start_date"] = nil
		} else {
			t, _ := time.Parse("2006-01-02", *req.StartDate)
			updates["start_date"] = &t
		}
	}
	s.db.Model(&i).Updates(updates)
	if req.ProjectIDs != nil {
		s.db.Where("initiative_id = ?", id).Delete(&model.InitiativeProject{})
		for _, pid := range req.ProjectIDs {
			s.db.Create(&model.InitiativeProject{InitiativeID: id, ProjectID: pid})
		}
	}
	s.db.Preload("Projects").First(&i, id)
	return &i, nil
}

func (s *InitiativeService) Delete(id, callerID uint64) error {
	var i model.Initiative
	if err := s.db.First(&i, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("Initiative not found")
		}
		return err
	}
	if err := s.checkWorkspaceAdmin(i.WorkspaceID, callerID); err != nil {
		return err
	}
	s.db.Where("initiative_id = ?", id).Delete(&model.InitiativeProject{})
	return s.db.Delete(&model.Initiative{}, id).Error
}

// GetProgress returns stats for projects linked to this initiative
func (s *InitiativeService) GetProgress(id uint64) (map[string]interface{}, error) {
	var initiative model.Initiative
	if err := s.db.Preload("Projects").First(&initiative, id).Error; err != nil {
		return nil, err
	}

	projectIDs := make([]uint64, len(initiative.Projects))
	for i, p := range initiative.Projects {
		projectIDs[i] = p.ID
	}

	var totalIssues, completedIssues int64
	if len(projectIDs) > 0 {
		s.db.Model(&model.Issue{}).Where("project_id IN ? AND archived_at IS NULL", projectIDs).Count(&totalIssues)
		s.db.Model(&model.Issue{}).Where("project_id IN ? AND archived_at IS NULL", projectIDs).
			Joins("JOIN states ON states.id = issues.state_id AND states.group IN ('completed','cancelled')").Count(&completedIssues)
	}

	progress := float64(0)
	if totalIssues > 0 {
		progress = float64(completedIssues) / float64(totalIssues) * 100
	}

	return map[string]interface{}{
		"total_issues":     totalIssues,
		"completed_issues": completedIssues,
		"progress":         progress,
		"project_count":    len(initiative.Projects),
	}, nil
}
