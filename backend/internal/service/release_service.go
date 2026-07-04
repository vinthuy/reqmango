package service

import (
	"fmt"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type ReleaseService struct {
	db *gorm.DB
}

func NewReleaseService(db *gorm.DB) *ReleaseService {
	return &ReleaseService{db: db}
}

func (s *ReleaseService) Create(projectID uint64, req *request.ReleaseCreateRequest) (*model.Release, error) {
	release := &model.Release{
		Name:        req.Name,
		Version:     req.Version,
		Description: req.Description,
		Status:      req.Status,
		ProjectID:   projectID,
	}

	if release.Status == "" {
		release.Status = "planned"
	}

	if req.ReleaseDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.ReleaseDate); err == nil {
			release.ReleaseDate = &t
		}
	}

	if err := s.db.Create(release).Error; err != nil {
		return nil, common.Internal("Failed to create release")
	}

	return release, nil
}

func (s *ReleaseService) List(projectID uint64) ([]model.Release, error) {
	var releases []model.Release
	if err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&releases).Error; err != nil {
		return nil, common.Internal("Failed to fetch releases")
	}
	return releases, nil
}

func (s *ReleaseService) Search(projectID uint64, query string) ([]model.Release, error) {
	var releases []model.Release
	if err := s.db.Where("project_id = ? AND (name ILIKE ? OR version ILIKE ?)", projectID, "%"+query+"%", "%"+query+"%").Order("created_at DESC").Find(&releases).Error; err != nil {
		return nil, common.Internal("Failed to search releases")
	}
	return releases, nil
}

func (s *ReleaseService) Get(projectID, releaseID uint64) (*model.Release, error) {
	var release model.Release
	if err := s.db.Where("project_id = ? AND id = ?", projectID, releaseID).First(&release).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Release not found")
		}
		return nil, common.Internal("Failed to fetch release")
	}
	return &release, nil
}

func (s *ReleaseService) Update(projectID, releaseID uint64, req *request.ReleaseUpdateRequest) (*model.Release, error) {
	var release model.Release
	if err := s.db.Where("project_id = ? AND id = ?", projectID, releaseID).First(&release).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Release not found")
		}
		return nil, common.Internal("Failed to fetch release")
	}

	if req.Name != "" {
		release.Name = req.Name
	}
	if req.Version != "" {
		release.Version = req.Version
	}
	if req.Description != "" {
		release.Description = req.Description
	}
	if req.Status != "" {
		release.Status = req.Status
	}
	if req.ReleaseDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.ReleaseDate); err == nil {
			release.ReleaseDate = &t
		}
	}

	if err := s.db.Save(&release).Error; err != nil {
		return nil, common.Internal("Failed to update release")
	}

	return &release, nil
}

func (s *ReleaseService) Delete(projectID, releaseID uint64) error {
	var release model.Release
	if err := s.db.Where("project_id = ? AND id = ?", projectID, releaseID).First(&release).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("Release not found")
		}
		return common.Internal("Failed to fetch release")
	}

	if err := s.db.Delete(&release).Error; err != nil {
		return common.Internal("Failed to delete release")
	}

	return nil
}

func (s *ReleaseService) AddIssues(projectID, releaseID uint64, issueIDs []uint64) error {
	var release model.Release
	if err := s.db.Where("project_id = ? AND id = ?", projectID, releaseID).First(&release).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("Release not found")
		}
		return common.Internal("Failed to fetch release")
	}

	for _, issueID := range issueIDs {
		var issue model.Issue
		if err := s.db.Where("project_id = ? AND id = ?", projectID, issueID).First(&issue).Error; err != nil {
			return common.NotFound(fmt.Sprintf("Issue not found: %d", issueID))
		}

		var exists model.ReleaseIssue
		if err := s.db.Where("release_id = ? AND issue_id = ?", releaseID, issueID).First(&exists).Error; err == nil {
			continue
		}

		if err := s.db.Create(&model.ReleaseIssue{ReleaseID: releaseID, IssueID: issueID}).Error; err != nil {
			return common.Internal("Failed to add issue to release")
		}
	}

	return nil
}

func (s *ReleaseService) RemoveIssues(projectID, releaseID uint64, issueIDs []uint64) error {
	var release model.Release
	if err := s.db.Where("project_id = ? AND id = ?", projectID, releaseID).First(&release).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("Release not found")
		}
		return common.Internal("Failed to fetch release")
	}

	for _, issueID := range issueIDs {
		if err := s.db.Where("release_id = ? AND issue_id = ?", releaseID, issueID).Delete(&model.ReleaseIssue{}).Error; err != nil {
			return common.Internal("Failed to remove issue from release")
		}
	}

	return nil
}

func (s *ReleaseService) GetProgress(projectID, releaseID uint64) (map[string]interface{}, error) {
	var release model.Release
	if err := s.db.Where("project_id = ? AND id = ?", projectID, releaseID).First(&release).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Release not found")
		}
		return nil, common.Internal("Failed to fetch release")
	}

	var totalIssues int64
	var doneIssues int64

	s.db.Table("release_issues").
		Where("release_id = ?", releaseID).
		Count(&totalIssues)

	s.db.Table("release_issues").
		Joins("JOIN issues ON release_issues.issue_id = issues.id").
		Joins("JOIN states ON issues.state_id = states.id").
		Where("release_issues.release_id = ? AND states.group IN ('completed','cancelled')", releaseID).
		Count(&doneIssues)

	progress := 0
	if totalIssues > 0 {
		progress = (int(doneIssues) * 100) / int(totalIssues)
	}

	return map[string]interface{}{
		"id":           releaseID,
		"name":         release.Name,
		"total_issues": totalIssues,
		"done_issues":  doneIssues,
		"progress":     progress,
	}, nil
}
