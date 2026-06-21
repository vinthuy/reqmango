package service

import (
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

type CycleService struct {
	db *gorm.DB
}

func NewCycleService(db *gorm.DB) *CycleService {
	return &CycleService{db: db}
}

func (s *CycleService) ListByProject(projectID uint64) ([]model.Cycle, error) {
	var cycles []model.Cycle
	err := s.db.Where("project_id = ?", projectID).
		Order("start_date DESC").
		Find(&cycles).Error
	if err != nil {
		return nil, err
	}
	if cycles == nil {
		cycles = []model.Cycle{}
	}
	return cycles, nil
}
