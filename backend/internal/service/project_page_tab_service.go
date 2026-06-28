package service

import (
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type ProjectPageTabService struct {
	db *gorm.DB
}

func NewProjectPageTabService(db *gorm.DB) *ProjectPageTabService {
	return &ProjectPageTabService{db: db}
}

// List returns all page tabs for a project owned by the user.
func (s *ProjectPageTabService) List(projectID, userID uint64) ([]model.ProjectPageTab, error) {
	var tabs []model.ProjectPageTab
	err := s.db.Where("project_id = ? AND owner_id = ?", projectID, userID).
		Order("sort_order ASC").Find(&tabs).Error
	return tabs, err
}

// Create adds a new page tab.
func (s *ProjectPageTabService) Create(tab *model.ProjectPageTab) error {
	return s.db.Create(tab).Error
}

// Update modifies an existing page tab. Only owner can update.
func (s *ProjectPageTabService) Update(tabID, userID uint64, updates map[string]interface{}) error {
	return s.db.Model(&model.ProjectPageTab{}).
		Where("id = ? AND owner_id = ?", tabID, userID).
		Updates(updates).Error
}

// Delete removes a page tab. Only owner can delete.
func (s *ProjectPageTabService) Delete(tabID, userID uint64) error {
	return s.db.Where("id = ? AND owner_id = ?", tabID, userID).
		Delete(&model.ProjectPageTab{}).Error
}

// Get returns a single tab by ID.
func (s *ProjectPageTabService) Get(tabID, userID uint64) (*model.ProjectPageTab, error) {
	var tab model.ProjectPageTab
	err := s.db.Where("id = ? AND owner_id = ?", tabID, userID).First(&tab).Error
	if err != nil {
		return nil, err
	}
	return &tab, nil
}

// BatchSave replaces all tabs for a project+user in one transaction.
func (s *ProjectPageTabService) BatchSave(projectID, userID uint64, tabs []model.ProjectPageTab) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ? AND owner_id = ?", projectID, userID).
			Delete(&model.ProjectPageTab{}).Error; err != nil {
			return err
		}
		for i := range tabs {
			tabs[i].ID = 0
			tabs[i].ProjectID = projectID
			tabs[i].OwnerID = userID
			tabs[i].SortOrder = i
			if err := tx.Create(&tabs[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Reorder updates the sort_order for all tabs in bulk.
func (s *ProjectPageTabService) Reorder(projectID, userID uint64, orderedIDs []uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range orderedIDs {
			if err := tx.Model(&model.ProjectPageTab{}).
				Where("id = ? AND owner_id = ?", id, userID).
				Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
