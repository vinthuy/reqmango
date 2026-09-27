// Package issuetype answers which issue types are usable in a given scope.
//
// The answer is not simply "the types in the workspace": a project draws from
// its own project-level types plus the workspace-level types it explicitly
// imported, and anything outside that set is rejected when a work item is
// written. Callers that re-derive the list themselves end up offering ids that
// cannot be used — the type picker and the agent tool layer both resolve it
// here so they can never disagree.
package issuetype

import (
	"gorm.io/gorm"

	"github.com/reqmango/backend/internal/model"
)

// Visible returns the issue types usable in the given scope, ordered the way
// the type picker lists them.
//
// Workspace scope (projectID == nil) uses the workspace-level types only. A
// project scope adds the project's own types plus the workspace-level types the
// project imported; a project with neither falls back to the workspace-level
// set, so a legacy project is never left with an empty list.
//
// Like the project type picker, inactive types are included: the write path
// accepts them, so hiding them here would only make the two lists disagree.
func Visible(db *gorm.DB, workspaceID uint64, projectID *uint64) ([]model.IssueType, error) {
	query := db.Where("workspace_id = ?", workspaceID)

	if projectID == nil {
		query = query.Where("project_id IS NULL")
	} else {
		importedIDs, err := importedTypeIDs(db, *projectID)
		if err != nil {
			return nil, err
		}

		switch {
		case len(importedIDs) > 0:
			query = query.Where("project_id = ? OR (project_id IS NULL AND id IN ?)", *projectID, importedIDs)
		default:
			// Nothing imported: a project that owns types uses those.
			var ownCount int64
			if err := db.Model(&model.IssueType{}).
				Where("workspace_id = ? AND project_id = ?", workspaceID, *projectID).
				Count(&ownCount).Error; err != nil {
				return nil, err
			}
			if ownCount > 0 {
				query = query.Where("project_id = ?", *projectID)
			} else {
				// Legacy project with no types of its own: fall back to the
				// workspace-level set. Scoping to project_id IS NULL matters —
				// matching on workspace_id alone would leak other projects'
				// private types into this project's picker.
				query = query.Where("project_id IS NULL")
			}
		}
	}

	var types []model.IssueType
	if err := query.Order("sequence, created_at").Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

// importedTypeIDs returns the workspace-level type IDs the project has
// explicitly imported via the Import model.
func importedTypeIDs(db *gorm.DB, projectID uint64) ([]uint64, error) {
	var ids []uint64
	if err := db.Model(&model.IssueTypeImport{}).
		Where("project_id = ?", projectID).
		Pluck("workspace_type_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
