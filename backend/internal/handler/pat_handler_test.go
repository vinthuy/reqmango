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
	"github.com/reqmango/backend/internal/service"
	"github.com/reqmango/backend/internal/testutil"
)

func TestPATHandler_List(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery(`SELECT \* FROM "personal_access_tokens" WHERE`).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "token_prefix", "token_hash", "scopes"}).
			AddRow(1, 1, "cli", "reqmango_pat_ab3d", "h", ""))

	h := NewPATHandler(service.NewPATService(db))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/auth/tokens", nil)
	c.Set("currentUser", &model.User{BaseModel: model.BaseModel{ID: 1}})

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var body []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Len(t, body, 1)
	assert.Equal(t, "cli", body[0]["name"])
	assert.Equal(t, "reqmango_pat_ab3d", body[0]["token_prefix"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPATHandler_Create(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery(`INSERT INTO "personal_access_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	h := NewPATHandler(service.NewPATService(db))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/auth/tokens", strings.NewReader(`{"name":"cli-macbook"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("currentUser", &model.User{BaseModel: model.BaseModel{ID: 1}})

	h.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.True(t, strings.HasPrefix(body["token"].(string), service.PATPrefix))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPATHandler_Revoke(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	mock.ExpectExec(`UPDATE "personal_access_tokens" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	h := NewPATHandler(service.NewPATService(db))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("DELETE", "/auth/tokens/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("currentUser", &model.User{BaseModel: model.BaseModel{ID: 1}})

	h.Revoke(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
