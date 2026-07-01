package service

import (
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type SavedReportService struct{ db *gorm.DB }

func NewSavedReportService(db *gorm.DB) *SavedReportService {
	return &SavedReportService{db: db}
}

func (s *SavedReportService) List(projectID uint64) ([]model.SavedReport, error) {
	var reports []model.SavedReport
	if err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&reports).Error; err != nil {
		return nil, common.Internal("Failed to list saved reports")
	}
	return reports, nil
}

func (s *SavedReportService) Create(report *model.SavedReport) error {
	if report.Name == "" || report.ReportType == "" {
		return common.BadRequest("Name and report_type are required")
	}
	return s.db.Create(report).Error
}

func (s *SavedReportService) Update(id, projectID uint64, updates map[string]interface{}) error {
	result := s.db.Model(&model.SavedReport{}).
		Where("id = ? AND project_id = ?", id, projectID).
		Updates(updates)
	if result.RowsAffected == 0 {
		return common.NotFound("Saved report not found")
	}
	return result.Error
}

func (s *SavedReportService) Delete(id, projectID uint64) error {
	result := s.db.Where("id = ? AND project_id = ?", id, projectID).Delete(&model.SavedReport{})
	if result.RowsAffected == 0 {
		return common.NotFound("Saved report not found")
	}
	return result.Error
}
