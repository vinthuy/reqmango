package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func importCSVTestHandler(t *testing.T) (*gin.Engine, model.Project, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.Workspace{}, &model.Project{}, &model.State{}, &model.Issue{},
		&model.User{}, &model.WorkspaceMember{}, &model.ProjectMember{},
		&model.IssueType{}, &model.IssueTypeImport{}, &model.Label{},
	))

	workspace := model.Workspace{Name: "acme", Slug: "acme"}
	require.NoError(t, db.Create(&workspace).Error)
	project := model.Project{Name: "core", Identifier: "CORE", WorkspaceID: workspace.ID}
	require.NoError(t, db.Create(&project).Error)
	require.NoError(t, db.Create(&model.State{
		Name: "Backlog", Group: "backlog", IsDefault: true, ProjectID: &project.ID, WorkspaceID: workspace.ID,
	}).Error)
	user := model.User{Email: "importer@example.com", Username: "importer", IsActive: true}
	require.NoError(t, db.Create(&user).Error)

	h := NewIssueHandler(service.NewIssueService(db, nil, nil, nil, nil))
	r := gin.New()
	r.POST("/issues/import/csv", func(c *gin.Context) {
		c.Set("currentUser", &user)
		c.Next()
	}, h.ImportCSV)
	return r, project, db
}

func postCSV(t *testing.T, r *gin.Engine, query, mapping, csv string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "issues.csv")
	require.NoError(t, err)
	_, err = part.Write([]byte(csv))
	require.NoError(t, err)
	if mapping != "" {
		require.NoError(t, w.WriteField("mapping", mapping))
	}
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, "/issues/import/csv?"+query, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func importQuery(project model.Project, dryRun bool) string {
	q := "project_id=" + strconv.FormatUint(project.ID, 10) + "&workspace_id=" + strconv.FormatUint(project.WorkspaceID, 10)
	if dryRun {
		q += "&dry_run=true"
	}
	return q
}

func TestImportCSVDryRunReturnsHeadersAndMapping(t *testing.T) {
	r, project, db := importCSVTestHandler(t)
	rec := postCSV(t, r, importQuery(project, true), `{"name":"?????"}`, "?????,?????\n?????,dev@example.com\n")

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var result response.ImportResult
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	require.True(t, result.DryRun)
	require.Equal(t, []string{"?????", "?????"}, result.Headers)
	require.Equal(t, "?????", result.Mapping["name"])
	require.Equal(t, 1, result.TotalRows)
	require.Equal(t, 1, result.SuccessCount)
	require.Empty(t, result.ImportedIDs)

	var count int64
	require.NoError(t, db.Model(&model.Issue{}).Count(&count).Error)
	require.Zero(t, count, "dry run must not persist issues")
}

func TestImportCSVRejectsMalformedMapping(t *testing.T) {
	r, project, _ := importCSVTestHandler(t)
	rec := postCSV(t, r, importQuery(project, true), "not-json", "name\nrow\n")
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestImportCSVRejectsProjectFromAnotherWorkspace(t *testing.T) {
	r, project, _ := importCSVTestHandler(t)
	query := "project_id=" + strconv.FormatUint(project.ID, 10) + "&workspace_id=" + strconv.FormatUint(project.WorkspaceID+1, 10) + "&dry_run=true"
	rec := postCSV(t, r, query, "", "name\nrow\n")
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
