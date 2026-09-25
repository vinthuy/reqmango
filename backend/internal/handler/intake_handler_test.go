package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/testutil"
)

type mockIssueCreatedNotifier struct {
	calls []*model.Issue
}

func (m *mockIssueCreatedNotifier) NotifyIssueCreated(issue *model.Issue) {
	m.calls = append(m.calls, issue)
}

func TestIntakeHandler_Submit_NotifiesIssueCreated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer func() { _ = sqlDB.Close() }()

	const projectID uint64 = 10
	const workspaceID uint64 = 1
	const stateID uint64 = 5
	const issueID uint64 = 42

	mock.ExpectQuery(`SELECT \* FROM "projects"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "deleted_at", "created_by_id", "updated_by_id",
			"name", "identifier", "workspace_id",
		}).AddRow(projectID, nil, nil, nil, nil, nil, "Demo", "DEM", workspaceID))

	mock.ExpectQuery(`SELECT \* FROM "states"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "deleted_at", "created_by_id", "updated_by_id",
			"name", "project_id", "workspace_id", "is_default",
		}).AddRow(stateID, nil, nil, nil, nil, nil, "Backlog", projectID, workspaceID, true))

	mock.ExpectQuery(`INSERT INTO "issues"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(issueID))

	notifier := &mockIssueCreatedNotifier{}
	h := NewIntakeHandler(db, notifier)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "projectId", Value: "10"}}
	c.Request, _ = http.NewRequest("POST", "/api/v1/intake/10", strings.NewReader(`{"name":"Intake bug","description":"from form"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Submit(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(issueID), body["id"])
	assert.Equal(t, "pending", body["status"])

	require.Len(t, notifier.calls, 1)
	got := notifier.calls[0]
	assert.Equal(t, issueID, got.ID)
	assert.Equal(t, "Intake bug", got.Name)
	assert.Equal(t, projectID, got.ProjectID)
	require.NotNil(t, got.IntakeStatus)
	assert.Equal(t, "pending", *got.IntakeStatus)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIntakeHandler_Submit_NilNotifierStillSucceeds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer func() { _ = sqlDB.Close() }()

	mock.ExpectQuery(`SELECT \* FROM "projects"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "deleted_at", "created_by_id", "updated_by_id",
			"name", "identifier", "workspace_id",
		}).AddRow(uint64(10), nil, nil, nil, nil, nil, "Demo", "DEM", uint64(1)))

	mock.ExpectQuery(`SELECT \* FROM "states"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "deleted_at", "created_by_id", "updated_by_id",
			"name", "project_id", "workspace_id", "is_default",
		}).AddRow(uint64(5), nil, nil, nil, nil, nil, "Backlog", uint64(10), uint64(1), true))

	mock.ExpectQuery(`INSERT INTO "issues"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(7)))

	h := NewIntakeHandler(db, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "projectId", Value: "10"}}
	c.Request, _ = http.NewRequest("POST", "/api/v1/intake/10", strings.NewReader(`{"name":"No notify"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Submit(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
