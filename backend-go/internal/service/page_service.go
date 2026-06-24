package service

import (
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/dto/response"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

// PageService handles page business logic.
type PageService struct {
	db *gorm.DB
}

// NewPageService creates a new PageService.
func NewPageService(db *gorm.DB) *PageService {
	return &PageService{db: db}
}

// List returns all pages for a project (flat list, not tree).
func (s *PageService) List(projectID uint64, includeArchived bool) ([]response.PageResponse, error) {
	var pages []model.Page
	q := s.db.Where("project_id = ?", projectID)
	if !includeArchived {
		q = q.Where("archived_at IS NULL")
	}
	if err := q.Order("depth, sequence, created_at").Find(&pages).Error; err != nil {
		return nil, common.Internal("Failed to fetch pages")
	}

	resps := make([]response.PageResponse, len(pages))
	for i, p := range pages {
		resps[i] = pageToResponse(&p)
	}
	return resps, nil
}

// GetTree returns the page hierarchy as a tree.
func (s *PageService) GetTree(projectID uint64) ([]response.PageResponse, error) {
	var pages []model.Page
	if err := s.db.Where("project_id = ? AND archived_at IS NULL", projectID).
		Order("depth, sequence, created_at").Find(&pages).Error; err != nil {
		return nil, common.Internal("Failed to fetch pages")
	}

	// Build tree from flat list
	pageMap := make(map[uint64]*response.PageResponse)
	roots := make([]response.PageResponse, 0)

	for i := range pages {
		resp := pageToResponse(&pages[i])
		resp.Children = make([]response.PageResponse, 0)
		pageMap[pages[i].ID] = &resp
	}

	for _, p := range pages {
		resp := pageMap[p.ID]
		if p.ParentID != nil {
			if parent, ok := pageMap[*p.ParentID]; ok {
				parent.Children = append(parent.Children, *resp)
			} else {
				roots = append(roots, *resp)
			}
		} else {
			roots = append(roots, *resp)
		}
	}

	return roots, nil
}

// Get returns a single page by ID.
func (s *PageService) Get(pageID, projectID uint64) (*response.PageResponse, error) {
	var p model.Page
	if err := s.db.Where("id = ? AND project_id = ?", pageID, projectID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Page not found")
		}
		return nil, common.Internal("Failed to fetch page")
	}
	resp := pageToResponse(&p)
	return &resp, nil
}

// Create creates a new page.
func (s *PageService) Create(req *request.PageCreateRequest, projectID, workspaceID, userID uint64) (*response.PageResponse, error) {
	depth := 0
	if req.ParentID != nil {
		var parent model.Page
		if err := s.db.First(&parent, *req.ParentID).Error; err != nil {
			return nil, common.NotFound("Parent page not found")
		}
		depth = parent.Depth + 1
		if depth > 5 {
			return nil, common.Validation("Maximum page nesting depth is 5")
		}
	}

	seq := req.Sequence
	if seq == 0 {
		seq = 1
	}

	p := &model.Page{
		Title:       req.Title,
		Content:     req.Content,
		ContentJSON: req.ContentJSON,
		Published:   true,
		Sequence:    seq,
		ParentID:    req.ParentID,
		Depth:       depth,
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
	}
	p.CreatedByID = &userID

	if err := s.db.Create(p).Error; err != nil {
		return nil, common.Internal("Failed to create page")
	}
	resp := pageToResponse(p)
	return &resp, nil
}

// Update updates a page.
func (s *PageService) Update(pageID, projectID, userID uint64, req *request.PageUpdateRequest) (*response.PageResponse, error) {
	var p model.Page
	if err := s.db.Where("id = ? AND project_id = ?", pageID, projectID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Page not found")
		}
		return nil, common.Internal("Failed to fetch page")
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.ContentJSON != nil {
		updates["content_json"] = req.ContentJSON
	}
	if req.Published != nil {
		updates["published"] = *req.Published
	}
	if req.Sequence != nil {
		updates["sequence"] = *req.Sequence
	}
	updates["updated_by_id"] = userID

	if len(updates) > 0 {
		if err := s.db.Model(&p).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update page")
		}
		s.db.First(&p, p.ID)
	}
	resp := pageToResponse(&p)
	return &resp, nil
}

// Delete soft-deletes a page.
func (s *PageService) Delete(pageID, projectID uint64) error {
	result := s.db.Where("id = ? AND project_id = ?", pageID, projectID).Delete(&model.Page{})
	if result.Error != nil {
		return common.Internal("Failed to delete page")
	}
	if result.RowsAffected == 0 {
		return common.NotFound("Page not found")
	}
	return nil
}

// Archive archives a page.
func (s *PageService) Archive(pageID, projectID uint64) error {
	result := s.db.Model(&model.Page{}).Where("id = ? AND project_id = ?", pageID, projectID).
		Update("archived_at", gorm.Expr("NOW()"))
	if result.RowsAffected == 0 {
		return common.NotFound("Page not found")
	}
	return result.Error
}

// Restore restores an archived page.
func (s *PageService) Restore(pageID, projectID uint64) error {
	result := s.db.Model(&model.Page{}).Where("id = ? AND project_id = ?", pageID, projectID).
		Update("archived_at", nil)
	if result.RowsAffected == 0 {
		return common.NotFound("Page not found")
	}
	return result.Error
}

// ListChildren returns direct children of a page.
func (s *PageService) ListChildren(pageID, projectID uint64) ([]response.PageResponse, error) {
	var pages []model.Page
	if err := s.db.Where("parent_id = ? AND project_id = ? AND archived_at IS NULL", pageID, projectID).
		Order("sequence, created_at").Find(&pages).Error; err != nil {
		return nil, common.Internal("Failed to fetch child pages")
	}

	resps := make([]response.PageResponse, len(pages))
	for i, p := range pages {
		resps[i] = pageToResponse(&p)
	}
	return resps, nil
}

// Move moves a page to a new parent and/or position.
func (s *PageService) Move(pageID, projectID uint64, req *request.PageMoveRequest) (*response.PageResponse, error) {
	var p model.Page
	if err := s.db.Where("id = ? AND project_id = ?", pageID, projectID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Page not found")
		}
		return nil, common.Internal("Failed to fetch page")
	}

	updates := map[string]interface{}{
		"sequence": req.Sequence,
	}

	if req.ParentID != nil {
		if *req.ParentID == pageID {
			return nil, common.Validation("A page cannot be its own parent")
		}
		var parent model.Page
		if err := s.db.First(&parent, *req.ParentID).Error; err != nil {
			return nil, common.NotFound("Parent page not found")
		}
		if parent.Depth+1 > 5 {
			return nil, common.Validation("Maximum page nesting depth is 5")
		}
		updates["parent_id"] = *req.ParentID
		updates["depth"] = parent.Depth + 1
	} else {
		updates["parent_id"] = nil
		updates["depth"] = 0
	}

	if err := s.db.Model(&p).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to move page")
	}
	s.db.First(&p, p.ID)
	resp := pageToResponse(&p)
	return &resp, nil
}

// ==================== Helpers ====================

func pageToResponse(p *model.Page) response.PageResponse {
	return response.PageResponse{
		ID:          p.ID,
		Title:       p.Title,
		Content:     p.Content,
		ContentJSON: p.ContentJSON,
		Published:   p.Published,
		ArchivedAt:  p.ArchivedAt,
		Sequence:    p.Sequence,
		ParentID:    p.ParentID,
		Depth:       p.Depth,
		ProjectID:   p.ProjectID,
		WorkspaceID: p.WorkspaceID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		Children:    nil,
	}
}
