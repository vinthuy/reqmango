package service

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFieldPermissionListRequiresTenant(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Workspace{}, &model.FieldPermission{}))
	require.NoError(t, db.Create(&model.Workspace{BaseModel: model.BaseModel{ID: 10}, Name: "a", Slug: "a"}).Error)
	require.NoError(t, db.Create(&model.Workspace{BaseModel: model.BaseModel{ID: 20}, Name: "b", Slug: "b"}).Error)
	mk := func(ws, project uint64, field string) {
		require.NoError(t, db.Create(&model.FieldPermission{
			Resource: "issue", FieldName: field, RoleID: 1, WorkspaceID: ws, ProjectID: project,
		}).Error)
	}
	mk(10, 0, "title")
	mk(10, 4, "priority")
	mk(20, 0, "secret")

	svc := NewFieldPermissionService(db)
	_, err = svc.List(0, 0)
	require.Error(t, err)

	own, err := svc.List(0, 10)
	require.NoError(t, err)
	require.Len(t, own, 2)
	for _, p := range own {
		assert.Equal(t, uint64(10), p.WorkspaceID)
	}

	byProject, err := svc.List(4, 0)
	require.NoError(t, err)
	require.Len(t, byProject, 1)
	assert.Equal(t, "priority", byProject[0].FieldName)

	id, err := svc.ResolveWorkspaceID("a")
	require.NoError(t, err)
	assert.Equal(t, uint64(10), id)
}
