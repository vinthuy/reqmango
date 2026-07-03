package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// PageService handles page business logic.
type PageService struct {
	db              *gorm.DB
	versionSvc      *PageVersionService
}

// NewPageService creates a new PageService.
func NewPageService(db *gorm.DB) *PageService {
	return &PageService{
		db:         db,
		versionSvc: NewPageVersionService(db),
	}
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

// Update updates a page. Auto-saves a version snapshot if content changed.
func (s *PageService) Update(pageID, projectID, userID uint64, req *request.PageUpdateRequest) (*response.PageResponse, error) {
	var p model.Page
	if err := s.db.Where("id = ? AND project_id = ?", pageID, projectID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Page not found")
		}
		return nil, common.Internal("Failed to fetch page")
	}

	// Check lock
	if p.LockedByID != nil && *p.LockedByID != userID && p.LockedAt != nil {
		lockAge := time.Since(*p.LockedAt)
		if lockAge < 30*time.Minute {
			return nil, common.Validation("Page is locked by another user")
		}
		// Auto-release lock after 30 minutes
	}

	contentChanged := false
	updates := map[string]interface{}{}
	if req.Title != nil && *req.Title != p.Title {
		updates["title"] = *req.Title
		contentChanged = true
	}
	if req.Content != nil && *req.Content != p.Content {
		updates["content"] = *req.Content
		contentChanged = true
	}
	if req.ContentJSON != nil && (p.ContentJSON == nil || *req.ContentJSON != *p.ContentJSON) {
		updates["content_json"] = req.ContentJSON
		contentChanged = true
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

		// Auto-save version if content changed
		if contentChanged {
			s.versionSvc.SaveVersion(p.ID, p.Title, p.Content, p.ContentJSON, userID)
		}
	}
	resp := pageToResponse(&p)
	return &resp, nil
}

// Lock locks a page for editing by a user.
func (s *PageService) Lock(pageID, projectID, userID uint64) (*response.PageResponse, error) {
	var p model.Page
	if err := s.db.Where("id = ? AND project_id = ?", pageID, projectID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Page not found")
		}
		return nil, common.Internal("Failed to fetch page")
	}

	// Check if already locked by someone else
	if p.LockedByID != nil && *p.LockedByID != userID && p.LockedAt != nil {
		if time.Since(*p.LockedAt) < 30*time.Minute {
			return nil, common.Validation("Page is currently locked by another user")
		}
	}

	now := time.Now()
	updates := map[string]interface{}{
		"locked_by_id": userID,
		"locked_at":    now,
	}
	if err := s.db.Model(&p).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to lock page")
	}
	s.db.First(&p, p.ID)
	resp := pageToResponse(&p)
	return &resp, nil
}

// Unlock unlocks a page.
func (s *PageService) Unlock(pageID, projectID, userID uint64) (*response.PageResponse, error) {
	var p model.Page
	if err := s.db.Where("id = ? AND project_id = ?", pageID, projectID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Page not found")
		}
		return nil, common.Internal("Failed to fetch page")
	}

	// Only the locker or anyone can unlock after timeout
	if p.LockedByID != nil && *p.LockedByID != userID && p.LockedAt != nil {
		if time.Since(*p.LockedAt) < 30*time.Minute {
			return nil, common.Validation("Page is locked by another user")
		}
	}

	updates := map[string]interface{}{
		"locked_by_id": nil,
		"locked_at":    nil,
	}
	if err := s.db.Model(&p).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to unlock page")
	}
	s.db.First(&p, p.ID)
	resp := pageToResponse(&p)
	return &resp, nil
}

// GetForExport returns page content formatted for export.
func (s *PageService) GetForExport(pageID, projectID uint64, format string) (string, string, error) {
	var p model.Page
	if err := s.db.Where("id = ? AND project_id = ?", pageID, projectID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", "", common.NotFound("Page not found")
		}
		return "", "", common.Internal("Failed to fetch page")
	}

	content := p.Content
	if content == "" && p.ContentJSON != nil {
		content = *p.ContentJSON
	}

	switch strings.ToLower(format) {
	case "md", "markdown":
		md := fmt.Sprintf("# %s\n\n%s\n\n---\n*Exported from ReqMango on %s*", p.Title, content, time.Now().Format("2006-01-02 15:04"))
		return p.Title + ".md", md, nil
	case "html":
		html := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>%s</title></head><body><h1>%s</h1>%s<hr><em>Exported from ReqMango on %s</em></body></html>`,
			p.Title, p.Title, content, time.Now().Format("2006-01-02 15:04"))
		return p.Title + ".html", html, nil
	default:
		return p.Title + ".txt", content, nil
	}
}

// ConvertToIssue converts a page into a new issue.
func (s *PageService) ConvertToIssue(pageID, projectID, userID uint64, issueTypeID *uint64) (*model.Issue, error) {
	var p model.Page
	if err := s.db.Where("id = ? AND project_id = ?", pageID, projectID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Page not found")
		}
		return nil, common.Internal("Failed to fetch page")
	}

	// Create issue from page content
	issue := &model.Issue{
		Name:          p.Title,
		DescriptionHTML: p.Content,
		ProjectID:     p.ProjectID,
		WorkspaceID:   p.WorkspaceID,
	}
	issue.CreatedByID = &userID

	if issueTypeID != nil {
		issue.IssueTypeID = issueTypeID
	}

	if err := s.db.Create(issue).Error; err != nil {
		return nil, common.Internal("Failed to create issue from page")
	}

	// Link the page to the issue
	s.db.Exec("INSERT INTO issue_pages (issue_id, page_id) VALUES (?, ?) ON CONFLICT DO NOTHING", issue.ID, p.ID)

	return issue, nil
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
	resp := response.PageResponse{
		ID:          p.ID,
		Title:       p.Title,
		Content:     p.Content,
		ContentJSON: p.ContentJSON,
		Published:   p.Published,
		ArchivedAt:  p.ArchivedAt,
		Sequence:    p.Sequence,
		ParentID:    p.ParentID,
		Depth:       p.Depth,
		LockedByID:  p.LockedByID,
		LockedAt:    p.LockedAt,
		ProjectID:   p.ProjectID,
		WorkspaceID: p.WorkspaceID,
		CreatedByID: p.CreatedByID,
		UpdatedByID: p.UpdatedByID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		Children:    nil,
	}
	if p.LockedBy != nil {
		resp.LockedByName = p.LockedBy.DisplayName
	}
	return resp
}
