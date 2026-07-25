package service

import (
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type SkillService struct{ db *gorm.DB }

func NewSkillService(db *gorm.DB) *SkillService {
	return &SkillService{db: db}
}

func (s *SkillService) Create(wid uint64, req request.SkillCreate) (*response.SkillResponse, error) {
	skill := model.Skill{
		Name:        req.Name,
		Description: req.Description,
		SkillType:   req.SkillType,
		Version:     "1.0",
		Status:      "active",
		SkillMD:     req.SkillMD,
		Parameters:  req.Parameters,
		Tags:        req.Tags,
		UsageCount:  0,
		IsShared:    req.IsShared,
		WorkspaceID: wid,
	}

	if err := s.db.Create(&skill).Error; err != nil {
		return nil, common.Internal("Failed to create skill")
	}

	return s.Get(skill.ID)
}

func (s *SkillService) Get(id uint64) (*response.SkillResponse, error) {
	var skill model.Skill
	if err := s.db.First(&skill, id).Error; err != nil {
		return nil, common.NotFound("Skill not found")
	}

	return s.toResponse(&skill), nil
}

func (s *SkillService) List(wid uint64) ([]response.SkillResponse, error) {
	var skills []model.Skill
	s.db.Where("workspace_id = ?", wid).Find(&skills)

	res := make([]response.SkillResponse, 0, len(skills))
	for _, sk := range skills {
		res = append(res, *s.toResponse(&sk))
	}

	return res, nil
}

func (s *SkillService) Update(id uint64, req request.SkillUpdate) (*response.SkillResponse, error) {
	var skill model.Skill
	if err := s.db.First(&skill, id).Error; err != nil {
		return nil, common.NotFound("Skill not found")
	}

	if req.Name != nil {
		skill.Name = *req.Name
	}
	if req.Description != nil {
		skill.Description = req.Description
	}
	if req.SkillType != nil {
		skill.SkillType = *req.SkillType
	}
	if req.SkillMD != nil {
		skill.SkillMD = *req.SkillMD
	}
	if req.Parameters != nil {
		skill.Parameters = *req.Parameters
	}
	if req.Tags != nil {
		skill.Tags = *req.Tags
	}
	if req.IsShared != nil {
		skill.IsShared = *req.IsShared
	}
	if req.Status != nil {
		skill.Status = *req.Status
	}

	if err := s.db.Save(&skill).Error; err != nil {
		return nil, common.Internal("Failed to update skill")
	}

	return s.Get(id)
}

func (s *SkillService) Delete(id uint64) error {
	var skill model.Skill
	if err := s.db.First(&skill, id).Error; err != nil {
		return common.NotFound("Skill not found")
	}

	return s.db.Delete(&skill).Error
}

func (s *SkillService) IncrementUsage(id uint64) error {
	return s.db.Model(&model.Skill{}).Where("id = ?", id).Update("usage_count", gorm.Expr("usage_count + 1")).Error
}

func (s *SkillService) toResponse(sk *model.Skill) *response.SkillResponse {
	return &response.SkillResponse{
		ID:          sk.ID,
		Name:        sk.Name,
		Description: sk.Description,
		SkillType:   sk.SkillType,
		Version:     sk.Version,
		Status:      sk.Status,
		SkillMD:     sk.SkillMD,
		Parameters:  sk.Parameters,
		Tags:        sk.Tags,
		UsageCount:  sk.UsageCount,
		IsShared:    sk.IsShared,
		WorkspaceID: sk.WorkspaceID,
		CreatedAt:   sk.CreatedAt,
		UpdatedAt:   sk.UpdatedAt,
	}
}
