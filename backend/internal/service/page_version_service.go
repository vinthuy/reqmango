package service

import (
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// PageVersionService handles page version history.
type PageVersionService struct {
	db *gorm.DB
}

// NewPageVersionService creates a new PageVersionService.
func NewPageVersionService(db *gorm.DB) *PageVersionService {
	return &PageVersionService{db: db}
}

// List returns all versions for a page.
func (s *PageVersionService) List(pageID uint64) ([]response.PageVersionResponse, error) {
	var versions []model.PageVersion
	if err := s.db.Where("page_id = ?", pageID).
		Preload("CreatedBy").
		Order("version_number DESC").
		Find(&versions).Error; err != nil {
		return nil, common.Internal("Failed to fetch page versions")
	}

	resps := make([]response.PageVersionResponse, len(versions))
	for i, v := range versions {
		resps[i] = toVersionResponse(&v)
	}
	return resps, nil
}

// GetByVersion returns a specific version of a page.
func (s *PageVersionService) GetByVersion(pageID uint64, versionNumber int) (*response.PageVersionResponse, error) {
	var v model.PageVersion
	if err := s.db.Where("page_id = ? AND version_number = ?", pageID, versionNumber).
		Preload("CreatedBy").
		First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Page version not found")
		}
		return nil, common.Internal("Failed to fetch page version")
	}
	resp := toVersionResponse(&v)
	return &resp, nil
}

// Restore restores a page to a specific version.
func (s *PageVersionService) Restore(pageID uint64, versionNumber int, userID uint64) error {
	var v model.PageVersion
	if err := s.db.Where("page_id = ? AND version_number = ?", pageID, versionNumber).First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("Page version not found")
		}
		return common.Internal("Failed to fetch page version")
	}

	// Update the page with version content
	updates := map[string]interface{}{
		"title":       v.Title,
		"content":     v.Content,
		"content_json": v.ContentJSON,
		"updated_by_id": userID,
	}
	if err := s.db.Model(&model.Page{}).Where("id = ?", pageID).Updates(updates).Error; err != nil {
		return common.Internal("Failed to restore page version")
	}

	// Create a new version snapshot after restore
	s.createVersion(pageID, v.Title, v.Content, v.ContentJSON, userID)

	return nil
}

// createVersion creates a new version snapshot (called internally).
func (s *PageVersionService) createVersion(pageID uint64, title, content string, contentJSON *string, userID uint64) {
	var maxVersion int
	s.db.Model(&model.PageVersion{}).
		Where("page_id = ?", pageID).
		Select("COALESCE(MAX(version_number), 0)").
		Scan(&maxVersion)

	version := &model.PageVersion{
		PageID:        pageID,
		Title:         title,
		Content:       content,
		ContentJSON:   contentJSON,
		VersionNumber: maxVersion + 1,
	}
	version.CreatedByID = &userID
	s.db.Create(version)

	// Limit to 50 versions per page (delete oldest)
	var count int64
	s.db.Model(&model.PageVersion{}).Where("page_id = ?", pageID).Count(&count)
	if count > 50 {
		// Keep only the 50 most recent versions
		var keepIDs []uint64
		s.db.Model(&model.PageVersion{}).
			Where("page_id = ?", pageID).
			Order("version_number DESC").
			Limit(50).
			Pluck("id", &keepIDs)

		s.db.Where("page_id = ? AND id NOT IN ?", pageID, keepIDs).Delete(&model.PageVersion{})
	}
}

// SaveVersion is called externally to snapshot a page version.
func (s *PageVersionService) SaveVersion(pageID uint64, title, content string, contentJSON *string, userID uint64) {
	s.createVersion(pageID, title, content, contentJSON, userID)
}

func toVersionResponse(v *model.PageVersion) response.PageVersionResponse {
	resp := response.PageVersionResponse{
		ID:            v.ID,
		PageID:        v.PageID,
		Title:         v.Title,
		Content:       v.Content,
		ContentJSON:   v.ContentJSON,
		VersionNumber: v.VersionNumber,
		ChangeSummary: v.ChangeSummary,
		CreatedAt:     v.CreatedAt,
		CreatedByID:   v.CreatedByID,
	}
	if v.CreatedByID != nil {
		resp.CreatedByID = v.CreatedByID
	}
	if v.CreatedBy != nil {
		resp.CreatedByName = v.CreatedBy.DisplayName
	}
	return resp
}
