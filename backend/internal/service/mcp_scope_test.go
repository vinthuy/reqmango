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

func TestMCPInWorkspace(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPConfig{}))
	cfg := model.MCPConfig{WorkspaceID: 10, Name: "m", ServerURL: "http://x", APIKey: "secret"}
	require.NoError(t, db.Create(&cfg).Error)

	s := &MCPService{db: db}
	assert.NoError(t, s.InWorkspace(cfg.ID, 10))
	assert.Error(t, s.InWorkspace(cfg.ID, 20), "another workspace cannot use this server and its key")
}
