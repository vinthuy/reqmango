package service

import (
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type IssueTypeService struct {
	db *gorm.DB
}

func NewIssueTypeService(db *gorm.DB) *IssueTypeService {
	return &IssueTypeService{db: db}
}

// buildResponse converts an IssueType model to its API response shape.
// isImported indicates the project has explicitly imported this workspace-level
// type via the Plane v3-style Import model (only meaningful in project context).
func (s *IssueTypeService) buildResponse(t model.IssueType, isImported bool) *response.IssueTypeResponse {
	return &response.IssueTypeResponse{
		ID:           t.ID,
		Name:         t.Name,
		Color:        t.Color,
		Icon:         t.Icon,
		Description:  t.Description,
		Level:        t.Level,
		ParentTypeID: t.ParentTypeID,
		IsDefault:    t.IsDefault,
		Sequence:     t.Sequence,
		IsActive:     t.IsActive,
		ProjectID:    t.ProjectID,
		WorkspaceID:  t.WorkspaceID,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
		IsInherited:  t.ProjectID == nil,
		IsImported:   isImported,
	}
}

// ensureOneDefault ensures only one default type exists in the given scope.
func (s *IssueTypeService) ensureOneDefault(workspaceID, excludeID uint64, projectID *uint64) error {
	query := s.db.Model(&model.IssueType{}).Where("workspace_id = ? AND is_default = ?", workspaceID, true)
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	} else {
		query = query.Where("project_id IS NULL")
	}
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	return query.Update("is_default", false).Error
}

// listImportedTypeIDs returns the set of workspace-level type IDs that the
// given project has explicitly imported via the Import model.
func (s *IssueTypeService) listImportedTypeIDs(projectID uint64) (map[uint64]bool, error) {
	var imports []model.IssueTypeImport
	if err := s.db.Where("project_id = ?", projectID).Find(&imports).Error; err != nil {
		return nil, err
	}
	m := make(map[uint64]bool, len(imports))
	for _, i := range imports {
		m[i.WorkspaceTypeID] = true
	}
	return m, nil
}

// ==================== CRUD ====================

func (s *IssueTypeService) Create(workspaceID, userID uint64, req request.IssueTypeCreate) (*response.IssueTypeResponse, error) {
	if req.IsDefault {
		if err := s.ensureOneDefault(workspaceID, 0, req.ProjectID); err != nil {
			return nil, common.Internal("Failed to update defaults")
		}
	}
	if req.Sequence == 0 {
		req.Sequence = 1
	}
	if req.Color == "" {
		req.Color = "#6366F1"
	}
	if req.Icon == "" {
		req.Icon = "circle"
	}

	t := model.IssueType{
		Name:         req.Name,
		Color:        req.Color,
		Icon:         req.Icon,
		Description:  req.Description,
		Level:        req.Level,
		ParentTypeID: req.ParentTypeID,
		IsDefault:    req.IsDefault,
		Sequence:     req.Sequence,
		IsActive:     true,
		ProjectID:    req.ProjectID,
		WorkspaceID:  workspaceID,
	}
	t.CreatedByID = &userID

	if err := s.db.Create(&t).Error; err != nil {
		return nil, common.Internal("Failed to create issue type")
	}

	return s.buildResponse(t, false), nil
}

// List returns issue types visible in the given scope.
//
//   - workspace scope (projectID == nil): returns all workspace-level types
//     (project_id IS NULL). IsImported is always false (no project context).
//   - project scope (projectID != nil): returns project-private types plus
//     workspace-level types that the project can see (legacy auto-inherit OR
//     query, kept for backward compatibility). The IsImported flag is true
//     for workspace-level types the project has explicitly imported via the
//     Plane v3-style Import model.
func (s *IssueTypeService) List(workspaceID uint64, projectID *uint64) ([]response.IssueTypeResponse, error) {
	if projectID == nil {
		var types []model.IssueType
		if err := s.db.Where("workspace_id = ? AND project_id IS NULL", workspaceID).
			Order("sequence, created_at").Find(&types).Error; err != nil {
			return nil, common.Internal("Failed to list issue types")
		}
		result := make([]response.IssueTypeResponse, len(types))
		for i, t := range types {
			result[i] = *s.buildResponse(t, false)
		}
		if result == nil {
			result = []response.IssueTypeResponse{}
		}
		return result, nil
	}

	// Project scope: project-private + workspace-shared (legacy auto-inherit).
	var types []model.IssueType
	if err := s.db.Where("workspace_id = ? AND (project_id = ? OR project_id IS NULL)", workspaceID, *projectID).
		Order("sequence, created_at").Find(&types).Error; err != nil {
		return nil, common.Internal("Failed to list issue types")
	}

	importedIDs, err := s.listImportedTypeIDs(*projectID)
	if err != nil {
		return nil, common.Internal("Failed to load import records")
	}

	result := make([]response.IssueTypeResponse, len(types))
	for i, t := range types {
		isImported := t.ProjectID == nil && importedIDs[t.ID]
		result[i] = *s.buildResponse(t, isImported)
	}
	if result == nil {
		result = []response.IssueTypeResponse{}
	}
	return result, nil
}

func (s *IssueTypeService) Get(typeID uint64) (*response.IssueTypeResponse, error) {
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return nil, common.NotFound("Issue type not found")
	}
	return s.buildResponse(t, false), nil
}

func (s *IssueTypeService) Update(typeID, userID uint64, req request.IssueTypeUpdate) (*response.IssueTypeResponse, error) {
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return nil, common.NotFound("Issue type not found")
	}

	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Color != nil {
		t.Color = *req.Color
	}
	if req.Icon != nil {
		t.Icon = *req.Icon
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			if err := s.ensureOneDefault(t.WorkspaceID, t.ID, t.ProjectID); err != nil {
				return nil, common.Internal("Failed to update defaults")
			}
		}
		t.IsDefault = *req.IsDefault
	}
	if req.Sequence != nil {
		t.Sequence = *req.Sequence
	}
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}
	if req.ProjectID != nil {
		t.ProjectID = req.ProjectID
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.Level != nil {
		t.Level = *req.Level
	}
	if req.ParentTypeID != nil {
		t.ParentTypeID = req.ParentTypeID
	}
	if req.AllowedChildTypeIDs != nil {
		t.AllowedChildTypeIDs = req.AllowedChildTypeIDs
	}

	t.UpdatedByID = &userID
	if err := s.db.Save(&t).Error; err != nil {
		return nil, common.Internal("Failed to update issue type")
	}

	return s.buildResponse(t, false), nil
}

func (s *IssueTypeService) Delete(typeID uint64) error {
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return common.NotFound("Issue type not found")
	}

	// Clean up join table entries
	s.db.Where("type_id = ?", typeID).Delete(&model.IssueTypeField{})

	// Clean up import records referencing this workspace type
	s.db.Where("workspace_type_id = ?", typeID).Delete(&model.IssueTypeImport{})

	return s.db.Delete(&t).Error
}

func (s *IssueTypeService) Disable(typeID uint64, isActive bool) error {
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return common.NotFound("Issue type not found")
	}
	return s.db.Model(&t).Update("is_active", isActive).Error
}

// ==================== Field Association ====================

func (s *IssueTypeService) AddField(typeID, fieldID uint64, isRequired bool, sequence int) (*response.IssueTypeFieldResponse, error) {
	// Verify type exists
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return nil, common.NotFound("Issue type not found")
	}

	// Verify field exists
	var f model.CustomField
	if err := s.db.First(&f, fieldID).Error; err != nil {
		return nil, common.NotFound("Custom field not found")
	}

	// Check for duplicate
	var count int64
	s.db.Model(&model.IssueTypeField{}).Where("type_id = ? AND field_id = ?", typeID, fieldID).Count(&count)
	if count > 0 {
		return nil, common.Conflict("Field is already associated with this issue type")
	}

	link := model.IssueTypeField{
		TypeID:     typeID,
		FieldID:    fieldID,
		IsRequired: isRequired,
		Sequence:   sequence,
	}

	if err := s.db.Create(&link).Error; err != nil {
		return nil, common.Internal("Failed to add field to issue type")
	}

	return &response.IssueTypeFieldResponse{
		FieldID:     fieldID,
		TypeID:      typeID,
		IsRequired:  isRequired,
		Sequence:    sequence,
		Name:        f.Name,
		FieldType:   f.FieldType,
		Description: f.Description,
	}, nil
}

func (s *IssueTypeService) RemoveField(typeID, fieldID uint64) error {
	result := s.db.Where("type_id = ? AND field_id = ?", typeID, fieldID).Delete(&model.IssueTypeField{})
	if result.RowsAffected == 0 {
		return common.NotFound("Field association not found")
	}
	return nil
}

func (s *IssueTypeService) UpdateField(typeID, fieldID uint64, req request.IssueTypeFieldUpdate) (*response.IssueTypeFieldResponse, error) {
	var link model.IssueTypeField
	if err := s.db.Where("type_id = ? AND field_id = ?", typeID, fieldID).First(&link).Error; err != nil {
		return nil, common.NotFound("Field association not found")
	}

	if req.IsRequired != nil {
		link.IsRequired = *req.IsRequired
	}
	if req.Sequence != nil {
		link.Sequence = *req.Sequence
	}

	if err := s.db.Save(&link).Error; err != nil {
		return nil, common.Internal("Failed to update field association")
	}

	// Load field info for response
	var f model.CustomField
	s.db.First(&f, fieldID)

	return &response.IssueTypeFieldResponse{
		FieldID:     fieldID,
		TypeID:      typeID,
		IsRequired:  link.IsRequired,
		Sequence:    link.Sequence,
		Name:        f.Name,
		FieldType:   f.FieldType,
		Description: f.Description,
	}, nil
}

// ==================== Project-scoped Operations ====================

// CopyFromWorkspace copies all workspace-level issue types to the target project.
func (s *IssueTypeService) CopyFromWorkspace(workspaceID, projectID, userID uint64) ([]response.IssueTypeResponse, error) {
	var templates []model.IssueType
	if err := s.db.Where("workspace_id = ? AND project_id IS NULL AND is_active = ?", workspaceID, true).
		Order("sequence, created_at").Find(&templates).Error; err != nil {
		return nil, common.Internal("Failed to fetch workspace issue types")
	}

	// Check if project already has types
	var existingCount int64
	s.db.Model(&model.IssueType{}).Where("project_id = ?", projectID).Count(&existingCount)
	if existingCount > 0 {
		return nil, common.Conflict("Project already has issue types configured")
	}

	types := make([]model.IssueType, len(templates))
	for i, tmpl := range templates {
		types[i] = model.IssueType{
			Name:         tmpl.Name,
			Color:        tmpl.Color,
			Icon:         tmpl.Icon,
			Description:  tmpl.Description,
			Level:        tmpl.Level,
			ParentTypeID: nil, // Reset hierarchy for project scope — will be re-linked post-creation
			IsDefault:    tmpl.IsDefault,
			Sequence:     tmpl.Sequence,
			IsActive:     true,
			ProjectID:    &projectID,
			WorkspaceID:  workspaceID,
		}
		types[i].CreatedByID = &userID
	}

	if err := s.db.Create(&types).Error; err != nil {
		return nil, common.Internal("Failed to copy issue types to project")
	}

	// Re-link parent relationships within the copied types
	// Build a map from template ID -> new type ID
	idMap := make(map[uint64]uint64)
	for i, tmpl := range templates {
		idMap[tmpl.ID] = types[i].ID
	}
	for i, tmpl := range templates {
		if tmpl.ParentTypeID != nil {
			if newParentID, ok := idMap[*tmpl.ParentTypeID]; ok {
				s.db.Model(&types[i]).Update("parent_type_id", newParentID)
			}
		}
	}

	// Also copy field associations from workspace types to project types
	// Build a map from workspace type ID -> new project type ID
	for i, tmpl := range templates {
		// Copy IssueTypeField associations from workspace type
		var workspaceFields []model.IssueTypeField
		if err := s.db.Where("type_id = ?", tmpl.ID).Find(&workspaceFields).Error; err != nil {
			continue
		}

		// Create corresponding fields for the new project type
		for _, wf := range workspaceFields {
			newField := model.IssueTypeField{
				TypeID:     types[i].ID,
				FieldID:    wf.FieldID,
				IsRequired: wf.IsRequired,
				Sequence:   wf.Sequence,
			}
			s.db.Create(&newField)
		}
	}

	result := make([]response.IssueTypeResponse, len(types))
	for i, t := range types {
		result[i] = *s.buildResponse(t, false)
	}
	return result, nil
}

// Reorder updates the sequence for issue types within a project.
func (s *IssueTypeService) Reorder(projectID uint64, typeIDs []uint64) error {
	for i, typeID := range typeIDs {
		if err := s.db.Model(&model.IssueType{}).Where("id = ? AND project_id = ?", typeID, projectID).
			Update("sequence", i+1).Error; err != nil {
			return common.Internal("Failed to reorder issue types")
		}
	}
	return nil
}

// ReorderWorkspace updates the sequence for issue types within a workspace.
func (s *IssueTypeService) ReorderWorkspace(workspaceID uint64, typeIDs []uint64) error {
	for i, typeID := range typeIDs {
		if err := s.db.Model(&model.IssueType{}).Where("id = ? AND workspace_id = ?", typeID, workspaceID).
			Update("sequence", i+1).Error; err != nil {
			return common.Internal("Failed to reorder issue types")
		}
	}
	return nil
}

// ListFields returns the custom fields attached to a type, optionally filtered
// by project enrollment.
//
// Behavior:
//   - No project context: all attached fields are returned.
//   - Project context:
//   - If the type is project-private, all its attached fields are visible.
//   - If the type is a workspace-level type imported by the project via the
//     Plane v3-style Import model, all attached fields are visible (they
//     "follow" the type) — no enrollment required.
//   - Otherwise (workspace-level type visible only via legacy auto-inherit),
//     workspace-shared fields require explicit enrollment to be visible.
func (s *IssueTypeService) ListFields(typeID uint64, projectID ...uint64) ([]response.IssueTypeFieldResponse, error) {
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return nil, common.NotFound("Issue type not found")
	}

	var links []model.IssueTypeField
	if err := s.db.Preload("Field").Where("type_id = ?", typeID).Order("sequence").Find(&links).Error; err != nil {
		return nil, common.Internal("Failed to list fields")
	}

	hasProjectContext := len(projectID) > 0 && projectID[0] > 0
	var enabledFieldIDs map[uint64]bool
	typeIsImported := false
	if hasProjectContext {
		pid := projectID[0]
		// Imported workspace type: all attached fields visible without enrollment.
		typeIsImported = s.IsImported(pid, typeID)
		if !typeIsImported && t.ProjectID == nil {
			// Legacy auto-inherited workspace type: fall back to enrollment filter.
			enabledFieldIDs = make(map[uint64]bool)
			var enrollments []model.ProjectCustomFieldEnrollment
			s.db.Where("project_id = ?", pid).Find(&enrollments)
			for _, e := range enrollments {
				enabledFieldIDs[e.FieldID] = true
			}
		}
	}

	result := make([]response.IssueTypeFieldResponse, 0)
	for _, link := range links {
		// Skip workspace-shared fields the project hasn't enrolled, but only when
		// the type is NOT imported (legacy auto-inherit path).
		if hasProjectContext && !typeIsImported && t.ProjectID == nil {
			if link.Field.ProjectID == nil && !enabledFieldIDs[link.FieldID] {
				continue
			}
		}

		fr := response.IssueTypeFieldResponse{
			FieldID:     link.FieldID,
			TypeID:      link.TypeID,
			IsRequired:  link.IsRequired,
			Sequence:    link.Sequence,
			Name:        link.Field.Name,
			FieldType:   link.Field.FieldType,
			Description: link.Field.Description,
			Options:     make([]response.CustomFieldOptionResponse, 0),
		}
		if link.Field.FieldType == "dropdown" {
			var opts []model.CustomFieldOption
			s.db.Where("field_id = ?", link.FieldID).Order("sequence").Find(&opts)
			for _, o := range opts {
				fr.Options = append(fr.Options, response.CustomFieldOptionResponse{
					ID:       o.ID,
					FieldID:  o.FieldID,
					Value:    o.Value,
					Color:    o.Color,
					Sequence: o.Sequence,
				})
			}
		}
		result = append(result, fr)
	}
	if result == nil {
		result = []response.IssueTypeFieldResponse{}
	}
	return result, nil
}

// ==================== Plane v3-style Import Model ====================

// ListImportable returns workspace-level types that the project has NOT yet
// imported. These are candidates for the Import dialog in the project UI.
func (s *IssueTypeService) ListImportable(workspaceID, projectID uint64) ([]response.IssueTypeResponse, error) {
	importedIDs, err := s.listImportedTypeIDs(projectID)
	if err != nil {
		return nil, common.Internal("Failed to load import records")
	}

	var types []model.IssueType
	if err := s.db.Where("workspace_id = ? AND project_id IS NULL AND is_active = ?", workspaceID, true).
		Order("sequence, created_at").Find(&types).Error; err != nil {
		return nil, common.Internal("Failed to list workspace issue types")
	}

	result := make([]response.IssueTypeResponse, 0, len(types))
	for _, t := range types {
		if importedIDs[t.ID] {
			continue
		}
		r := s.buildResponse(t, false)
		r.IsImported = false
		result = append(result, *r)
	}
	if result == nil {
		result = []response.IssueTypeResponse{}
	}
	return result, nil
}

// ImportType records a project's reference to a workspace-level type (Plane v3
// Import model). After import, custom fields attached to the type become
// visible in the project automatically — no separate enrollment required.
func (s *IssueTypeService) ImportType(projectID, workspaceTypeID uint64) error {
	var t model.IssueType
	if err := s.db.First(&t, workspaceTypeID).Error; err != nil {
		return common.NotFound("Issue type not found")
	}
	if t.ProjectID != nil {
		return common.BadRequest("Cannot import a project-level type")
	}

	rec := model.IssueTypeImport{
		ProjectID:       projectID,
		WorkspaceTypeID: workspaceTypeID,
		WorkspaceID:     t.WorkspaceID,
	}
	if err := s.db.Create(&rec).Error; err != nil {
		if common.IsUniqueViolation(err) {
			return common.Conflict("Type already imported by this project")
		}
		return common.Internal("Failed to import issue type")
	}
	return nil
}

// UnimportType removes a project's reference to a workspace-level type. The
// type itself is not deleted — it remains available at the workspace level and
// may still be visible to the project via the legacy auto-inherit OR query
// (backward compatibility). To fully hide a workspace type from a project,
// an exclusion mechanism would be required (not part of the v2 Import model).
//
// Hard-deletes (Unscoped) instead of soft-deletes because IssueTypeImport is a
// pure reference/join table: a soft-deleted row would still occupy the
// (project_id, workspace_type_id) unique index and block a subsequent
// re-import with a 409 Conflict. Unimport is fully reversible via ImportType.
func (s *IssueTypeService) UnimportType(projectID, workspaceTypeID uint64) error {
	result := s.db.Unscoped().
		Where("project_id = ? AND workspace_type_id = ?", projectID, workspaceTypeID).
		Delete(&model.IssueTypeImport{})
	if result.RowsAffected == 0 {
		return common.NotFound("Import record not found")
	}
	return nil
}

// IsImported reports whether the project has explicitly imported the given
// workspace-level type via the Plane v3-style Import model.
func (s *IssueTypeService) IsImported(projectID, workspaceTypeID uint64) bool {
	var count int64
	s.db.Model(&model.IssueTypeImport{}).
		Where("project_id = ? AND workspace_type_id = ?", projectID, workspaceTypeID).
		Count(&count)
	return count > 0
}
