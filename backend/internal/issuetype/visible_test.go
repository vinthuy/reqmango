package issuetype

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.IssueType{}, &model.IssueTypeImport{}))
	return db
}

func ptr(v uint64) *uint64 { return &v }

func mkType(t *testing.T, db *gorm.DB, id, workspaceID uint64, name string, projectID *uint64) {
	t.Helper()
	require.NoError(t, db.Create(&model.IssueType{
		BaseModel:   model.BaseModel{ID: id},
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Name:        name,
		IsActive:    true,
		Sequence:    1,
	}).Error)
}

func names(types []model.IssueType) []string {
	out := make([]string, len(types))
	for i, tp := range types {
		out[i] = tp.Name
	}
	return out
}

func TestVisible_WorkspaceScopeReturnsWorkspaceTypesOnly(t *testing.T) {
	db := setupDB(t)
	mkType(t, db, 1, 1, "Bug", nil)
	mkType(t, db, 2, 1, "Feature", nil)
	mkType(t, db, 3, 1, "Project Bug", ptr(9))

	got, err := Visible(db, 1, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"Bug", "Feature"}, names(got))
}

func TestVisible_ProjectScopePrefersItsOwnTypes(t *testing.T) {
	// The regression: a project whose picker offers its own Bug (id 12369) must
	// not be told the workspace Bug (id 3) is usable — adopting that id fails.
	db := setupDB(t)
	mkType(t, db, 3, 1, "Bug", nil)
	mkType(t, db, 12369, 1, "Bug", ptr(1))
	mkType(t, db, 12370, 1, "Task", ptr(1))

	got, err := Visible(db, 1, ptr(1))
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, []string{"Bug", "Task"}, names(got))
	assert.Equal(t, uint64(12369), got[0].ID)
}

func TestVisible_ProjectScopeAddsImportedWorkspaceTypes(t *testing.T) {
	// Importing is opt-in: the workspace type becomes usable, its un-imported
	// siblings do not.
	db := setupDB(t)
	mkType(t, db, 3, 1, "Bug", nil)
	mkType(t, db, 2, 1, "Feature", nil)
	mkType(t, db, 12369, 1, "Project Bug", ptr(1))
	require.NoError(t, db.Create(&model.IssueTypeImport{
		ProjectID: 1, WorkspaceTypeID: 3, WorkspaceID: 1,
	}).Error)

	got, err := Visible(db, 1, ptr(1))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Project Bug", "Bug"}, names(got))
}

func TestVisible_ProjectWithoutOwnTypesFallsBackToWorkspaceTypes(t *testing.T) {
	db := setupDB(t)
	mkType(t, db, 3, 1, "Bug", nil)
	mkType(t, db, 4, 1, "Task", nil)

	got, err := Visible(db, 1, ptr(1))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Bug", "Task"}, names(got))
}

func TestVisible_FallbackDoesNotLeakOtherProjectsTypes(t *testing.T) {
	// Matching on workspace_id alone would pull project 2's private types into
	// project 1's list, and the agent would propose ids project 1 cannot use.
	db := setupDB(t)
	mkType(t, db, 3, 1, "Bug", nil)
	mkType(t, db, 500, 1, "Other Project Only", ptr(2))

	got, err := Visible(db, 1, ptr(1))
	require.NoError(t, err)
	assert.Equal(t, []string{"Bug"}, names(got))
}

func TestVisible_IgnoresOtherWorkspaces(t *testing.T) {
	db := setupDB(t)
	mkType(t, db, 3, 1, "Bug", nil)
	mkType(t, db, 9, 2, "Bug", nil)

	got, err := Visible(db, 1, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"Bug"}, names(got))
	assert.Equal(t, uint64(3), got[0].ID)
}
