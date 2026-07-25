package service

import (
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type AgentConfigService struct{ db *gorm.DB }

func NewAgentConfigService(db *gorm.DB) *AgentConfigService {
	return &AgentConfigService{db: db}
}

func (s *AgentConfigService) Create(wid uint64, req request.AgentConfigCreate) (*response.AgentConfigResponse, error) {
	config := model.AgentConfig{
		Name:           req.Name,
		Description:    req.Description,
		Provider:       req.Provider,
		Model:          req.Model,
		APIKey:         req.APIKey,
		APIEndpoint:    req.APIEndpoint,
		InferenceLevel: req.InferenceLevel,
		ServiceLevel:   req.ServiceLevel,
		MaxTokens:      req.MaxTokens,
		Temperature:    req.Temperature,
		TopP:           req.TopP,
		IsDefault:      req.IsDefault,
		IsActive:       true,
		WorkspaceID:    wid,
	}

	if err := s.db.Create(&config).Error; err != nil {
		return nil, common.Internal("Failed to create agent config")
	}

	return s.Get(config.ID)
}

func (s *AgentConfigService) Get(id uint64) (*response.AgentConfigResponse, error) {
	var config model.AgentConfig
	if err := s.db.First(&config, id).Error; err != nil {
		return nil, common.NotFound("Agent config not found")
	}

	return s.toResponse(&config), nil
}

func (s *AgentConfigService) List(wid uint64) ([]response.AgentConfigResponse, error) {
	var configs []model.AgentConfig
	s.db.Where("workspace_id = ?", wid).Find(&configs)

	res := make([]response.AgentConfigResponse, 0, len(configs))
	for _, c := range configs {
		res = append(res, *s.toResponse(&c))
	}

	return res, nil
}

func (s *AgentConfigService) Update(id uint64, req request.AgentConfigUpdate) (*response.AgentConfigResponse, error) {
	var config model.AgentConfig
	if err := s.db.First(&config, id).Error; err != nil {
		return nil, common.NotFound("Agent config not found")
	}

	if req.Name != nil {
		config.Name = *req.Name
	}
	if req.Description != nil {
		config.Description = req.Description
	}
	if req.Provider != nil {
		config.Provider = *req.Provider
	}
	if req.Model != nil {
		config.Model = *req.Model
	}
	if req.APIKey != nil {
		config.APIKey = *req.APIKey
	}
	if req.APIEndpoint != nil {
		config.APIEndpoint = req.APIEndpoint
	}
	if req.InferenceLevel != nil {
		config.InferenceLevel = *req.InferenceLevel
	}
	if req.ServiceLevel != nil {
		config.ServiceLevel = *req.ServiceLevel
	}
	if req.MaxTokens != nil {
		config.MaxTokens = *req.MaxTokens
	}
	if req.Temperature != nil {
		config.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		config.TopP = *req.TopP
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			s.db.Model(&model.AgentConfig{}).Where("workspace_id = ?", config.WorkspaceID).Update("is_default", false)
		}
		config.IsDefault = *req.IsDefault
	}
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}

	if err := s.db.Save(&config).Error; err != nil {
		return nil, common.Internal("Failed to update agent config")
	}

	return s.Get(id)
}

func (s *AgentConfigService) Delete(id uint64) error {
	var config model.AgentConfig
	if err := s.db.First(&config, id).Error; err != nil {
		return common.NotFound("Agent config not found")
	}

	return s.db.Delete(&config).Error
}

func (s *AgentConfigService) GetDefault(wid uint64) (*response.AgentConfigResponse, error) {
	var config model.AgentConfig
	if err := s.db.Where("workspace_id = ? AND is_default = ?", wid, true).First(&config).Error; err != nil {
		return nil, common.NotFound("No default agent config found")
	}

	return s.toResponse(&config), nil
}

func (s *AgentConfigService) toResponse(c *model.AgentConfig) *response.AgentConfigResponse {
	return &response.AgentConfigResponse{
		ID:             c.ID,
		Name:           c.Name,
		Description:    c.Description,
		Provider:       c.Provider,
		Model:          c.Model,
		APIEndpoint:    c.APIEndpoint,
		InferenceLevel: c.InferenceLevel,
		ServiceLevel:   c.ServiceLevel,
		MaxTokens:      c.MaxTokens,
		Temperature:    c.Temperature,
		TopP:           c.TopP,
		IsDefault:      c.IsDefault,
		IsActive:       c.IsActive,
		WorkspaceID:    c.WorkspaceID,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
