package service

import (
	"errors"
	"time"

	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

type InitiativeService struct{ db *gorm.DB }

func NewInitiativeService(db *gorm.DB) *InitiativeService { return &InitiativeService{db} }

func (s *InitiativeService) Create(workspaceID uint64, req request.CreateInitiativeReq) (*model.Initiative, error) {
	status := req.Status
	if status == "" {
		status = "active"
	}
	initiative := model.Initiative{
		WorkspaceID: workspaceID, Name: req.Name, Description: req.Description,
		Color: req.Color, Status: status, CreatedByID: 1,
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
			tx.Create(&model.InitiativeProject{InitiativeID: initiative.ID, ProjectID: pid})
		}
	}
	tx.Commit()
	s.db.Preload("Projects").First(&initiative, initiative.ID)
	return &initiative, nil
}

func (s *InitiativeService) List(workspaceID uint64) ([]model.Initiative, error) {
	var initiatives []model.Initiative
	err := s.db.Where("workspace_id = ?", workspaceID).
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

func (s *InitiativeService) Update(id uint64, req request.UpdateInitiativeReq) (*model.Initiative, error) {
	var i model.Initiative
	if err := s.db.First(&i, id).Error; err != nil {
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

func (s *InitiativeService) Delete(id uint64) error {
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
		"total_issues":    totalIssues,
		"completed_issues": completedIssues,
		"progress":        progress,
		"project_count":   len(initiative.Projects),
	}, nil
}
