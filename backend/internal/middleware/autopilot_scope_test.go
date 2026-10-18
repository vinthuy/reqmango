package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestRequireWorkspaceAutopilotTask covers a leak the fresh-database audit
// found: the autopilot handlers load their task by id alone, so an admin of one
// workspace could read another tenant's task and executions by guessing ids.
func TestRequireWorkspaceAutopilotTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Workspace{}, &model.AutopilotTask{}, &model.AutopilotExecution{}))

	require.NoError(t, db.Create(&model.Workspace{BaseModel: model.BaseModel{ID: 10}, Name: "own", Slug: "own"}).Error)
	require.NoError(t, db.Create(&model.Workspace{BaseModel: model.BaseModel{ID: 20}, Name: "other", Slug: "other"}).Error)

	// Raw inserts: the sqlite driver returns these JSON columns as strings and
	// GORM cannot scan them back into json.RawMessage on Create.
	require.NoError(t, db.Exec(`INSERT INTO autopilot_tasks (id, workspace_id, name, description, trigger_type, task_type, input_data, status, created_at, updated_at)
		VALUES (1, 10, 'own', '', 'manual', 'report', '{}', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO autopilot_tasks (id, workspace_id, name, description, trigger_type, task_type, input_data, status, created_at, updated_at)
		VALUES (2, 20, 'foreign', '', 'manual', 'report', '{}', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error)
	const ownTaskID, foreignTaskID = 1, 2
	require.NoError(t, db.Exec(`INSERT INTO autopilot_executions (id, task_id, status, input_data, output_data, logs, duration_ms, retry_count, created_at, updated_at)
		VALUES (1, ?, 'pending', '{}', '{}', '[]', 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, ownTaskID).Error)
	require.NoError(t, db.Exec(`INSERT INTO autopilot_executions (id, task_id, status, input_data, output_data, logs, duration_ms, retry_count, created_at, updated_at)
		VALUES (2, ?, 'pending', '{}', '{}', '[]', 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, foreignTaskID).Error)

	r := gin.New()
	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	scope := RequireWorkspaceAutopilotTask(db)
	r.GET("/workspaces/:wsParam/autopilot-tasks/:taskId", scope, ok)
	r.GET("/workspaces/:wsParam/autopilot-tasks/:taskId/executions/:executionId", scope, ok)

	status := func(path string) int {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w.Code
	}

	assert.Equal(t, 200, status("/workspaces/10/autopilot-tasks/"+uitoa(ownTaskID)))
	assert.Equal(t, 200, status("/workspaces/own/autopilot-tasks/"+uitoa(ownTaskID)), "slugs resolve")
	assert.Equal(t, 404, status("/workspaces/10/autopilot-tasks/"+uitoa(foreignTaskID)), "another tenant's task")
	assert.Equal(t, 404, status("/workspaces/10/autopilot-tasks/9999"), "unknown task")
	assert.Equal(t, 404, status("/workspaces/nope/autopilot-tasks/"+uitoa(ownTaskID)))

	assert.Equal(t, 200, status("/workspaces/10/autopilot-tasks/"+uitoa(ownTaskID)+"/executions/1"))
	assert.Equal(t, 404, status("/workspaces/10/autopilot-tasks/"+uitoa(ownTaskID)+"/executions/2"), "another tenant's execution")
}
