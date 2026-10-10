package service

import (
	"testing"

	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/require"
)

func TestCustomFieldService_CreateWorkspaceFieldVisibleInProjects(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&model.CustomField{},
		&model.CustomFieldOption{},
		&model.IssueCustomFieldValue{},
		&model.ProjectCustomFieldEnrollment{},
		&model.IssueTypeField{},
		&model.IssueTypeImport{},
	))
	wsID, projID, userID := seedTestData(t, db)
	other := &model.Project{Name: "Other", Identifier: "OTH", WorkspaceID: wsID}
	require.NoError(t, db.Create(other).Error)

	svc := NewCustomFieldService(db)
	created, err := svc.Create(wsID, userID, request.CustomFieldCreate{Name: "Fresh", FieldType: "text"})
	require.NoError(t, err)

	for _, p := range []uint64{projID, other.ID} {
		fields, err := svc.List(wsID, &p, nil)
		require.NoError(t, err)
		var names []string
		for _, f := range fields {
			names = append(names, f.Name)
		}
		require.Contains(t, names, "Fresh", "project %d should list the new workspace field", p)
	}

	// A project-private field must not leak into other projects.
	private, err := svc.Create(wsID, userID, request.CustomFieldCreate{Name: "Private", FieldType: "text", ProjectID: &projID})
	require.NoError(t, err)
	fields, err := svc.List(wsID, &other.ID, nil)
	require.NoError(t, err)
	for _, f := range fields {
		require.NotEqual(t, private.ID, f.ID)
	}
	require.NotZero(t, created.ID)
}
