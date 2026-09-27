package service

import (
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func triageDef(t *testing.T) PMAgentDef {
	t.Helper()
	for _, def := range PMAgentDefs() {
		if def.Name == "请求分诊" {
			return def
		}
	}
	t.Fatal("triage agent definition missing")
	return PMAgentDef{}
}

func newAgentDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Agent{}))
	return db
}

func storedCaps(t *testing.T, db *gorm.DB, id uint64) []string {
	t.Helper()
	var a model.Agent
	require.NoError(t, db.First(&a, id).Error)
	var caps []string
	require.NoError(t, json.Unmarshal(a.Capabilities, &caps))
	return caps
}

func TestUpgradeLegacyCapabilities_UpgradesAnUntouchedBuiltinAgent(t *testing.T) {
	db := newAgentDB(t)
	def := triageDef(t)
	legacy, _ := json.Marshal(def.LegacyCapabilities)
	agent := model.Agent{Name: def.Name, WorkspaceID: 1, Capabilities: legacy, InvocationTargets: json.RawMessage(`[]`)}
	require.NoError(t, db.Create(&agent).Error)

	upgraded, err := (&AgentService{db: db}).upgradeLegacyCapabilities(&agent, def)

	require.NoError(t, err)
	assert.True(t, upgraded)
	assert.Equal(t, def.Capabilities, storedCaps(t, db, agent.ID))
}

func TestUpgradeLegacyCapabilities_LeavesAUserEditedAgentAlone(t *testing.T) {
	db := newAgentDB(t)
	def := triageDef(t)
	custom, _ := json.Marshal([]string{"search", "comment"})
	agent := model.Agent{Name: def.Name, WorkspaceID: 1, Capabilities: custom, InvocationTargets: json.RawMessage(`[]`)}
	require.NoError(t, db.Create(&agent).Error)

	upgraded, err := (&AgentService{db: db}).upgradeLegacyCapabilities(&agent, def)

	require.NoError(t, err)
	assert.False(t, upgraded)
	assert.Equal(t, []string{"search", "comment"}, storedCaps(t, db, agent.ID))
}

func TestPMAgentDefs_HasFour(t *testing.T) {
	assert.Len(t, PMAgentDefs(), 4)
	names := map[string]bool{}
	for _, d := range PMAgentDefs() {
		names[d.Name] = true
		assert.NotEmpty(t, d.SystemPrompt)
	}
	assert.True(t, names["请求分诊"])
	assert.True(t, names["交付风险"])
	assert.True(t, names["Sprint 总结"])
	assert.True(t, names["Spec 草稿"])
}
