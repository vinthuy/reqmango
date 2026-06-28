package service

import (
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type ProjectUpdateService struct{ db *gorm.DB }

func NewProjectUpdateService(db *gorm.DB) *ProjectUpdateService { return &ProjectUpdateService{db} }

func (s *ProjectUpdateService) Create(projectID, authorID uint64, status, content string) (*model.ProjectUpdate, error) {
	u := model.ProjectUpdate{ProjectID: projectID, AuthorID: authorID, Status: status, Content: content}
	err := s.db.Create(&u).Error
	if err != nil {
		return nil, err
	}
	s.db.Preload("Author").First(&u, u.ID)
	return &u, nil
}

func (s *ProjectUpdateService) List(projectID uint64, limit int) ([]model.ProjectUpdate, error) {
	var updates []model.ProjectUpdate
	if limit <= 0 {
		limit = 20
	}
	err := s.db.Where("project_id = ?", projectID).Preload("Author").
		Order("created_at DESC").Limit(limit).Find(&updates).Error
	return updates, err
}
