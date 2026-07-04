package service

import (
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
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

func (s *SavedReportService) Get(id, projectID uint64) (*model.SavedReport, error) {
	var report model.SavedReport
	if err := s.db.Where("id = ? AND project_id = ?", id, projectID).First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Saved report not found")
		}
		return nil, common.Internal("Failed to fetch saved report")
	}
	return &report, nil
}

func (s *SavedReportService) Create(projectID uint64, req *request.SavedReportCreateRequest) (*model.SavedReport, error) {
	if req.Name == "" || req.ReportType == "" {
		return nil, common.BadRequest("Name and report_type are required")
	}
	report := &model.SavedReport{
		Name:       req.Name,
		ReportType: req.ReportType,
		GroupBy:    req.GroupBy,
		ChartType:  req.ChartType,
		RQL:        req.RQL,
		Interval:   req.Interval,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
		ProjectID:  projectID,
	}
	if err := s.db.Create(report).Error; err != nil {
		return nil, common.Internal("Failed to create saved report")
	}
	return report, nil
}

func (s *SavedReportService) Update(id, projectID uint64, req *request.SavedReportUpdateRequest) (*model.SavedReport, error) {
	var report model.SavedReport
	if err := s.db.Where("id = ? AND project_id = ?", id, projectID).First(&report).Error; err != nil {
		return nil, common.NotFound("Saved report not found")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
		report.Name = *req.Name
	}
	if req.ReportType != nil {
		updates["report_type"] = *req.ReportType
		report.ReportType = *req.ReportType
	}
	if req.GroupBy != nil {
		updates["group_by"] = *req.GroupBy
		report.GroupBy = *req.GroupBy
	}
	if req.ChartType != nil {
		updates["chart_type"] = *req.ChartType
		report.ChartType = *req.ChartType
	}
	if req.RQL != nil {
		updates["rql"] = *req.RQL
		report.RQL = *req.RQL
	}
	if req.Interval != nil {
		updates["interval"] = *req.Interval
		report.Interval = *req.Interval
	}
	if req.DateFrom != nil {
		updates["date_from"] = *req.DateFrom
		report.DateFrom = *req.DateFrom
	}
	if req.DateTo != nil {
		updates["date_to"] = *req.DateTo
		report.DateTo = *req.DateTo
	}

	if len(updates) > 0 {
		if err := s.db.Model(&report).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update saved report")
		}
	}

	return &report, nil
}

func (s *SavedReportService) Delete(id, projectID uint64) error {
	result := s.db.Where("id = ? AND project_id = ?", id, projectID).Delete(&model.SavedReport{})
	if result.RowsAffected == 0 {
		return common.NotFound("Saved report not found")
	}
	return result.Error
}
