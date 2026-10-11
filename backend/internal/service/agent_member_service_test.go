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

func TestAgentMemberAdd_RejectsAgentsFromOtherWorkspaces(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Project{}, &model.Agent{}, &ProjectAgentMember{}))
	project := model.Project{Name: "p", Identifier: "P", WorkspaceID: 10}
	require.NoError(t, db.Create(&project).Error)
	own := model.Agent{WorkspaceID: 10, Name: "own", Capabilities: []byte(`[]`), InvocationTargets: []byte(`[]`)}
	foreign := model.Agent{WorkspaceID: 20, Name: "foreign", Capabilities: []byte(`[]`), InvocationTargets: []byte(`[]`)}
	require.NoError(t, db.Create(&own).Error)
	require.NoError(t, db.Create(&foreign).Error)

	s := NewAgentMemberService(db)
	_, err = s.Add(project.ID, own.ID, "")
	assert.NoError(t, err)
	_, err = s.Add(project.ID, foreign.ID, "")
	assert.Error(t, err)
	_, err = s.Add(project.ID, 999, "")
	assert.Error(t, err)
}
