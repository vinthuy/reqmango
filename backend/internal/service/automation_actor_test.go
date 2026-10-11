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

func TestAutomationActor_IsTheRuleCreatorNotUserOne(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Workspace{}))
	ws := model.Workspace{Name: "w", Slug: "w", OwnerID: 42}
	require.NoError(t, db.Create(&ws).Error)

	creator := uint64(7)
	rule := model.AutomationRule{WorkspaceID: ws.ID}
	rule.ID = 3
	rule.CreatedByID = &creator

	base := map[string]interface{}{"issue_id": uint64(5)}
	ctx := withRuleActor(base, rule)
	assert.NotContains(t, base, ruleActorKey, "the shared event context is not mutated across rules")
	assert.EqualValues(t, 5, ctx["issue_id"])

	actor, err := automationActor(ctx, db, ws.ID)
	require.NoError(t, err)
	assert.EqualValues(t, 7, actor)

	legacy := model.AutomationRule{WorkspaceID: ws.ID}
	actor, err = automationActor(withRuleActor(base, legacy), db, ws.ID)
	require.NoError(t, err)
	assert.EqualValues(t, 42, actor, "rules without a creator run as the workspace owner")

	_, err = automationActor(withRuleActor(base, legacy), db, 999)
	assert.Error(t, err, "no creator and no owner: refuse rather than borrow an arbitrary account")
}
