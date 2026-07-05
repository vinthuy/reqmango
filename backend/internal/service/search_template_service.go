package service

import (
	"encoding/json"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type SearchTemplateService struct {
	db *gorm.DB
}

func NewSearchTemplateService(db *gorm.DB) *SearchTemplateService {
	return &SearchTemplateService{db: db}
}

var builtInTemplates = []struct {
	Name        string
	Description string
	Icon        string
	RQLTemplate string
	ViewType    string
	SortConfig  string
	GroupBy     *string
}{
	{
		Name:        "我的待办",
		Description: "分配给我的所有待处理任务",
		Icon:        "📝",
		RQLTemplate: "state_group IN ('backlog', 'unstarted', 'started') AND assignee_id IN ($CURRENT_USER)",
		ViewType:    "list",
		SortConfig:  `[{"field":"priority","dir":"desc"},{"field":"target_date","dir":"asc"}]`,
	},
	{
		Name:        "本周到期",
		Description: "本周即将到期的任务",
		Icon:        "⏰",
		RQLTemplate: "target_date <= $END_OF_WEEK AND state_group != 'completed'",
		ViewType:    "list",
		SortConfig:  `[{"field":"target_date","dir":"asc"}]`,
	},
	{
		Name:        "高优先级",
		Description: "所有高优先级和紧急任务",
		Icon:        "🚨",
		RQLTemplate: "priority IN ('high', 'urgent') AND state_group != 'completed'",
		ViewType:    "list",
		SortConfig:  `[{"field":"priority","dir":"desc"},{"field":"created_at","dir":"desc"}]`,
	},
	{
		Name:        "我的已完成",
		Description: "我完成的任务",
		Icon:        "✅",
		RQLTemplate: "state_group = 'completed' AND assignee_id IN ($CURRENT_USER)",
		ViewType:    "list",
		SortConfig:  `[{"field":"completed_at","dir":"desc"}]`,
	},
	{
		Name:        "未分配任务",
		Description: "还没有分配人的任务",
		Icon:        "👤",
		RQLTemplate: "assignee_id IS NULL AND state_group != 'completed'",
		ViewType:    "list",
		SortConfig:  `[{"field":"created_at","dir":"asc"}]`,
	},
	{
		Name:        "需要关注",
		Description: "超过一周未更新的任务",
		Icon:        "🔔",
		RQLTemplate: "updated_at <= $ONE_WEEK_AGO AND state_group != 'completed'",
		ViewType:    "list",
		SortConfig:  `[{"field":"updated_at","dir":"asc"}]`,
	},
	{
		Name:        "待审核",
		Description: "所有进行中和评审中的任务",
		Icon:        "🔍",
		RQLTemplate: "state_group = 'started'",
		ViewType:    "list",
		SortConfig:  `[{"field":"updated_at","dir":"desc"}]`,
	},
	{
		Name:        "看板视图",
		Description: "按状态分组的看板视图",
		Icon:        "📋",
		RQLTemplate: "",
		ViewType:    "kanban",
		SortConfig:  `[{"field":"priority","dir":"desc"}]`,
	},
}

func (s *SearchTemplateService) List(projectID, userID uint64) ([]response.SearchTemplateResponse, error) {
	var templates []model.SearchTemplate
	if err := s.db.Where("project_id = ? AND (is_built_in = ? OR is_public = ? OR owner_id = ?)", projectID, true, true, userID).
		Order("is_built_in DESC, created_at ASC").Find(&templates).Error; err != nil {
		return nil, common.Internal("Failed to fetch search templates")
	}

	resps := make([]response.SearchTemplateResponse, len(templates))
	for i, t := range templates {
		resps[i] = searchTemplateToResponse(&t)
	}
	return resps, nil
}

func (s *SearchTemplateService) Get(id, projectID, userID uint64) (*response.SearchTemplateResponse, error) {
	var t model.SearchTemplate
	if err := s.db.Where("id = ? AND project_id = ? AND (is_built_in = ? OR is_public = ? OR owner_id = ?)", id, projectID, true, true, userID).
		First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Search template not found")
		}
		return nil, common.Internal("Failed to fetch search template")
	}
	resp := searchTemplateToResponse(&t)
	return &resp, nil
}

func (s *SearchTemplateService) Create(projectID, userID uint64, name, description, icon, rqlTemplate, viewType string, sortConfig, columns json.RawMessage, groupBy *string) (*response.SearchTemplateResponse, error) {
	t := &model.SearchTemplate{
		Name:        name,
		Icon:        icon,
		RQLTemplate: rqlTemplate,
		ViewType:    viewType,
		SortConfig:  normalizeJSON(sortConfig),
		Columns:     normalizeJSON(columns),
		GroupBy:     groupBy,
		IsBuiltIn:   false,
		IsPublic:    true,
		OwnerID:     &userID,
		ProjectID:   projectID,
	}
	if description != "" {
		t.Description = &description
	}
	if t.ViewType == "" {
		t.ViewType = "list"
	}

	if err := s.db.Create(t).Error; err != nil {
		return nil, common.Internal("Failed to create search template")
	}
	resp := searchTemplateToResponse(t)
	return &resp, nil
}

func (s *SearchTemplateService) Delete(id, projectID, userID uint64) error {
	result := s.db.Where("id = ? AND project_id = ? AND owner_id = ? AND is_built_in = ?", id, projectID, userID, false).Delete(&model.SearchTemplate{})
	if result.Error != nil {
		return common.Internal("Failed to delete search template")
	}
	if result.RowsAffected == 0 {
		return common.NotFound("Search template not found or access denied")
	}
	return nil
}

func (s *SearchTemplateService) ApplyTemplate(id, projectID, userID uint64) (*response.SearchTemplateResponse, error) {
	return s.Get(id, projectID, userID)
}

func (s *SearchTemplateService) InitializeBuiltInTemplates(projectID uint64) error {
	for _, bt := range builtInTemplates {
		var existing model.SearchTemplate
		err := s.db.Where("project_id = ? AND name = ? AND is_built_in = ?", projectID, bt.Name, true).First(&existing).Error
		if err == nil {
			// Template exists — update RQL, description, and sort config in case they changed
			desc := bt.Description
			updates := map[string]interface{}{
				"rql_template": bt.RQLTemplate,
				"description":  &desc,
				"sort_config":  json.RawMessage(bt.SortConfig),
				"view_type":    bt.ViewType,
				"icon":         bt.Icon,
				"group_by":     bt.GroupBy,
			}
			if err := s.db.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
			continue
		}

		t := &model.SearchTemplate{
			Name:        bt.Name,
			Icon:        bt.Icon,
			RQLTemplate: bt.RQLTemplate,
			ViewType:    bt.ViewType,
			SortConfig:  json.RawMessage(bt.SortConfig),
			GroupBy:     bt.GroupBy,
			IsBuiltIn:   true,
			IsPublic:    true,
			OwnerID:     nil,
			ProjectID:   projectID,
		}
		desc := bt.Description
		t.Description = &desc

		if err := s.db.Create(t).Error; err != nil {
			return err
		}
	}
	return nil
}

func searchTemplateToResponse(t *model.SearchTemplate) response.SearchTemplateResponse {
	return response.SearchTemplateResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Icon:        t.Icon,
		RQLTemplate: t.RQLTemplate,
		ViewType:    t.ViewType,
		SortConfig:  t.SortConfig,
		GroupBy:     t.GroupBy,
		Columns:     t.Columns,
		IsBuiltIn:   t.IsBuiltIn,
		IsPublic:    t.IsPublic,
		OwnerID:     t.OwnerID,
		ProjectID:   t.ProjectID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}