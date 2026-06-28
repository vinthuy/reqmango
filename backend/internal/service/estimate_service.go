package service

import (
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/common"
	"gorm.io/gorm"
)

type EstimateService struct {
	db *gorm.DB
}

func NewEstimateService(db *gorm.DB) *EstimateService {
	return &EstimateService{db: db}
}

func (s *EstimateService) GetSettings(projectID uint64) (*model.ProjectEstimateSettings, error) {
	var settings model.ProjectEstimateSettings
	err := s.db.Where("project_id = ?", projectID).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		return nil, common.NotFound("Estimate settings not found")
	}
	if err != nil {
		return nil, common.Internal("Failed to fetch estimate settings")
	}
	return &settings, nil
}

func (s *EstimateService) UpdateSettings(projectID, workspaceID uint64, mode model.EstimateMode) (*model.ProjectEstimateSettings, error) {
	var settings model.ProjectEstimateSettings
	err := s.db.Where("project_id = ?", projectID).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		settings = model.ProjectEstimateSettings{
			ProjectID:   projectID,
			WorkspaceID: workspaceID,
			Mode:        mode,
			PointsEnabled:  mode == model.EstimateModePoints,
			CategoriesEnabled: mode == model.EstimateModeCategories,
			TimeEnabled:    mode == model.EstimateModeTime,
		}
		err = s.db.Create(&settings).Error
	} else {
		settings.Mode = mode
		err = s.db.Save(&settings).Error
	}
	if err != nil {
		return nil, common.Internal("Failed to update estimate settings")
	}
	return &settings, nil
}

func (s *EstimateService) ListPoints(projectID uint64) ([]model.EstimatePoint, error) {
	var points []model.EstimatePoint
	err := s.db.Where("project_id = ?", projectID).Order("sequence ASC").Find(&points).Error
	if err != nil {
		return nil, common.Internal("Failed to fetch estimate points")
	}
	return points, nil
}

func (s *EstimateService) GetPoint(projectID, pointID uint64) (*model.EstimatePoint, error) {
	var point model.EstimatePoint
	err := s.db.Where("project_id = ? AND id = ?", projectID, pointID).First(&point).Error
	if err == gorm.ErrRecordNotFound {
		return nil, common.NotFound("Estimate point not found")
	}
	if err != nil {
		return nil, common.Internal("Failed to fetch estimate point")
	}
	return &point, nil
}

func (s *EstimateService) CreatePoint(projectID, workspaceID uint64, name string, value int, isDefault bool, sequence int) (*model.EstimatePoint, error) {
	if isDefault {
		s.db.Model(&model.EstimatePoint{}).Where("project_id = ? AND is_default = ?", projectID, true).Update("is_default", false)
	}

	point := model.EstimatePoint{
		Name:        name,
		Value:       value,
		Mode:        model.EstimateModePoints,
		IsDefault:   isDefault,
		Sequence:    sequence,
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
	}

	err := s.db.Create(&point).Error
	if err != nil {
		return nil, common.Internal("Failed to create estimate point")
	}
	return &point, nil
}

func (s *EstimateService) UpdatePoint(projectID, pointID uint64, name *string, value *int, isDefault *bool, sequence *int) (*model.EstimatePoint, error) {
	var point model.EstimatePoint
	err := s.db.Where("project_id = ? AND id = ?", projectID, pointID).First(&point).Error
	if err == gorm.ErrRecordNotFound {
		return nil, common.NotFound("Estimate point not found")
	}
	if err != nil {
		return nil, common.Internal("Failed to fetch estimate point")
	}

	if name != nil {
		point.Name = *name
	}
	if value != nil {
		point.Value = *value
	}
	if isDefault != nil && *isDefault {
		s.db.Model(&model.EstimatePoint{}).Where("project_id = ? AND is_default = ?", projectID, true).Update("is_default", false)
		point.IsDefault = true
	}
	if sequence != nil {
		point.Sequence = *sequence
	}

	err = s.db.Save(&point).Error
	if err != nil {
		return nil, common.Internal("Failed to update estimate point")
	}
	return &point, nil
}

func (s *EstimateService) DeletePoint(projectID, pointID uint64) error {
	err := s.db.Where("project_id = ? AND id = ?", projectID, pointID).Delete(&model.EstimatePoint{}).Error
	if err != nil {
		return common.Internal("Failed to delete estimate point")
	}
	return nil
}

func (s *EstimateService) ReorderPoints(projectID uint64, pointIDs []uint64) error {
	for i, id := range pointIDs {
		err := s.db.Model(&model.EstimatePoint{}).Where("project_id = ? AND id = ?", projectID, id).Update("sequence", i+1).Error
		if err != nil {
			return common.Internal("Failed to reorder estimate points")
		}
	}
	return nil
}

func (s *EstimateService) CreateDefaultPoints(projectID, workspaceID uint64) ([]model.EstimatePoint, error) {
	s.db.Where("project_id = ?", projectID).Delete(&model.EstimatePoint{})

	defaultPoints := []model.EstimatePoint{
		{Name: "0 - 不需要估算", Value: 0, Mode: model.EstimateModePoints, IsDefault: true, Sequence: 1, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "1 - 很小", Value: 1, Mode: model.EstimateModePoints, IsDefault: false, Sequence: 2, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "2 - 小", Value: 2, Mode: model.EstimateModePoints, IsDefault: false, Sequence: 3, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "3 - 中等", Value: 3, Mode: model.EstimateModePoints, IsDefault: false, Sequence: 4, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "5 - 较大", Value: 5, Mode: model.EstimateModePoints, IsDefault: false, Sequence: 5, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "8 - 大", Value: 8, Mode: model.EstimateModePoints, IsDefault: false, Sequence: 6, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "13 - 很大", Value: 13, Mode: model.EstimateModePoints, IsDefault: false, Sequence: 7, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "21 - 巨大", Value: 21, Mode: model.EstimateModePoints, IsDefault: false, Sequence: 8, ProjectID: projectID, WorkspaceID: workspaceID},
	}

	err := s.db.Create(&defaultPoints).Error
	if err != nil {
		return nil, common.Internal("Failed to create default estimate points")
	}
	return defaultPoints, nil
}

func (s *EstimateService) BulkCreatePoints(projectID, workspaceID uint64, points []struct {
	Name       string `json:"name"`
	Value      int    `json:"value"`
	IsDefault  bool   `json:"is_default"`
	Sequence   int    `json:"sequence"`
}) ([]model.EstimatePoint, error) {
	var estimatePoints []model.EstimatePoint
	for _, p := range points {
		estimatePoints = append(estimatePoints, model.EstimatePoint{
			Name:        p.Name,
			Value:       p.Value,
			Mode:        model.EstimateModePoints,
			IsDefault:   p.IsDefault,
			Sequence:    p.Sequence,
			ProjectID:   projectID,
			WorkspaceID: workspaceID,
		})
	}

	err := s.db.Create(&estimatePoints).Error
	if err != nil {
		return nil, common.Internal("Failed to bulk create estimate points")
	}
	return estimatePoints, nil
}

func (s *EstimateService) ListCategories(projectID uint64) ([]model.EstimateCategory, error) {
	var categories []model.EstimateCategory
	err := s.db.Where("project_id = ?", projectID).Order("sequence ASC").Find(&categories).Error
	if err != nil {
		return nil, common.Internal("Failed to fetch estimate categories")
	}
	return categories, nil
}

func (s *EstimateService) CreateCategory(projectID, workspaceID uint64, name string, isDefault bool, sequence int) (*model.EstimateCategory, error) {
	if isDefault {
		s.db.Model(&model.EstimateCategory{}).Where("project_id = ? AND is_default = ?", projectID, true).Update("is_default", false)
	}

	category := model.EstimateCategory{
		Name:        name,
		Mode:        model.EstimateModeCategories,
		IsDefault:   isDefault,
		Sequence:    sequence,
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
	}

	err := s.db.Create(&category).Error
	if err != nil {
		return nil, common.Internal("Failed to create estimate category")
	}
	return &category, nil
}

func (s *EstimateService) CreateDefaultCategories(projectID, workspaceID uint64) ([]model.EstimateCategory, error) {
	s.db.Where("project_id = ?", projectID).Delete(&model.EstimateCategory{})

	defaultCategories := []model.EstimateCategory{
		{Name: "XS - 极小", Mode: model.EstimateModeCategories, IsDefault: false, Sequence: 1, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "S - 小", Mode: model.EstimateModeCategories, IsDefault: false, Sequence: 2, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "M - 中等", Mode: model.EstimateModeCategories, IsDefault: true, Sequence: 3, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "L - 大", Mode: model.EstimateModeCategories, IsDefault: false, Sequence: 4, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "XL - 极大", Mode: model.EstimateModeCategories, IsDefault: false, Sequence: 5, ProjectID: projectID, WorkspaceID: workspaceID},
	}

	err := s.db.Create(&defaultCategories).Error
	if err != nil {
		return nil, common.Internal("Failed to create default estimate categories")
	}
	return defaultCategories, nil
}

func (s *EstimateService) ListTime(projectID uint64) ([]model.EstimateTime, error) {
	var times []model.EstimateTime
	err := s.db.Where("project_id = ?", projectID).Order("sequence ASC").Find(&times).Error
	if err != nil {
		return nil, common.Internal("Failed to fetch estimate time")
	}
	return times, nil
}

func (s *EstimateService) CreateTime(projectID, workspaceID uint64, name string, minutes int, isDefault bool, sequence int) (*model.EstimateTime, error) {
	if isDefault {
		s.db.Model(&model.EstimateTime{}).Where("project_id = ? AND is_default = ?", projectID, true).Update("is_default", false)
	}

	etime := model.EstimateTime{
		Name:        name,
		Minutes:     minutes,
		Mode:        model.EstimateModeTime,
		IsDefault:   isDefault,
		Sequence:    sequence,
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
	}

	err := s.db.Create(&etime).Error
	if err != nil {
		return nil, common.Internal("Failed to create estimate time")
	}
	return &etime, nil
}

func (s *EstimateService) CreateDefaultTime(projectID, workspaceID uint64) ([]model.EstimateTime, error) {
	s.db.Where("project_id = ?", projectID).Delete(&model.EstimateTime{})

	defaultTime := []model.EstimateTime{
		{Name: "15 分钟", Minutes: 15, Mode: model.EstimateModeTime, IsDefault: false, Sequence: 1, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "30 分钟", Minutes: 30, Mode: model.EstimateModeTime, IsDefault: false, Sequence: 2, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "1 小时", Minutes: 60, Mode: model.EstimateModeTime, IsDefault: true, Sequence: 3, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "2 小时", Minutes: 120, Mode: model.EstimateModeTime, IsDefault: false, Sequence: 4, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "4 小时", Minutes: 240, Mode: model.EstimateModeTime, IsDefault: false, Sequence: 5, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "1 天", Minutes: 480, Mode: model.EstimateModeTime, IsDefault: false, Sequence: 6, ProjectID: projectID, WorkspaceID: workspaceID},
		{Name: "2 天", Minutes: 960, Mode: model.EstimateModeTime, IsDefault: false, Sequence: 7, ProjectID: projectID, WorkspaceID: workspaceID},
	}

	err := s.db.Create(&defaultTime).Error
	if err != nil {
		return nil, common.Internal("Failed to create default estimate time")
	}
	return defaultTime, nil
}