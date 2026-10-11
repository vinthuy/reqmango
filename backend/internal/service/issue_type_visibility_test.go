package service

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTypeVisibilityDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.IssueType{}, &model.IssueTypeImport{}))
	return db
}

func TestValidateTypeVisibleInProject_LegacyFallbackMatchesVisible(t *testing.T) {
	db := setupTypeVisibilityDB(t)
	require.NoError(t, db.Create(&model.IssueType{
		BaseModel: model.BaseModel{ID: 3}, WorkspaceID: 1, Name: "Task", IsActive: true, Sequence: 1,
	}).Error)

	svc := &IssueService{db: db}
	err := svc.validateTypeVisibleInProject(3, 1, 1)
	require.NoError(t, err, "workspace type must be writable under Visible's legacy fallback")
}

func TestValidateTypeVisibleInProject_RejectsUnimportedWhenProjectHasImports(t *testing.T) {
	db := setupTypeVisibilityDB(t)
	require.NoError(t, db.Create(&model.IssueType{
		BaseModel: model.BaseModel{ID: 3}, WorkspaceID: 1, Name: "Bug", IsActive: true, Sequence: 1,
	}).Error)
	require.NoError(t, db.Create(&model.IssueType{
		BaseModel: model.BaseModel{ID: 4}, WorkspaceID: 1, Name: "Task", IsActive: true, Sequence: 2,
	}).Error)
	require.NoError(t, db.Create(&model.IssueTypeImport{
		ProjectID: 1, WorkspaceTypeID: 3, WorkspaceID: 1,
	}).Error)

	svc := &IssueService{db: db}
	err := svc.validateTypeVisibleInProject(4, 1, 1)
	require.Error(t, err)
	appErr, ok := err.(*common.AppError)
	require.True(t, ok)
	assert.Equal(t, 400, appErr.Code)
}
