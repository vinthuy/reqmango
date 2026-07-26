package security_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/handler"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupMockDB creates a mock database for testing
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{Conn: sqlDB})
	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
		require.NoError(t, err)

	return db, mock, sqlDB
}

func TestWorkspaceIsolation(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("UserA cannot list projects in UserB's workspace", func(t *testing.T) {
		db, _, sqlDB := setupMockDB(t)
		defer sqlDB.Close()

		// Initialize services
		projectSvc := service.NewProjectService(db)
		projectHandler := handler.NewProjectHandler(projectSvc, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/projects?workspace_id=2", nil)
		c.Request.Header.Set("Authorization", "Bearer test-token")
		
		// Mock JWT middleware injection - UserA (ID=1)
		c.Set("currentUser", &model.User{BaseModel: model.BaseModel{ID: 1}})

		projectHandler.List(c)
	
// Should return 403 Forbidden because UserA is not in workspace 2
		assert.Equal(t, http.StatusForbidden, w.Code)
		})

	t.Run("UserA cannot view issues in UserB's workspace", func(t *testing.T) {
		db, mock, sqlDB := setupMockDB(t)
		defer sqlDB.Close()

		// Mock: Issue belongs to workspace 2 (UserB's)
		mock.ExpectQuery(`SELECT`).
			WillReturnRows(sqlmock.NewRows([]string{"workspace_id"}).
				AddRow(2))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "issueId", Value: "1"}}
		c.Request, _ = http.NewRequest("GET", "/issues/1", nil)
		c.Request.Header.Set("Authorization", "Bearer test-token")
		
		// Mock JWT middleware injection - UserA (ID=1)
		c.Set("currentUser", &model.User{BaseModel: model.BaseModel{ID: 1}})

// Create issue handler
		issueSvc := service.NewIssueService(db, nil, nil, nil, nil)
		issueHandler := handler.NewIssueHandler(issueSvc)
		
		issueHandler.Get(c)

		// Should return 403 Forbidden because UserA is not in workspace 2
assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
