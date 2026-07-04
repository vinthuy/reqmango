package security_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/config"
	"github.com/reqmango/backend/internal/handler"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 初始化测试数据库并创建两个隔离的 Workspace
func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	
	// 自动迁移表结构
	err = db.AutoMigrate(&model.User{}, &model.Workspace{}, &model.WorkspaceMember{}, &model.Project{}, &model.Issue{})
	if err != nil {
		return nil, err
	}
	
	// 创建用户 A 和 B
	userA := model.User{Email: "a@test.com", Username: "user_a", PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"} // password: secret
	userB := model.User{Email: "b@test.com", Username: "user_b", PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"}
	db.Create(&userA)
	db.Create(&userB)

	// 创建工作空间 1 和 2
	ws1 := model.Workspace{Name: "WS 1", Slug: "ws1", OwnerID: userA.ID}
	ws2 := model.Workspace{Name: "WS 2", Slug: "ws2", OwnerID: userB.ID}
	db.Create(&ws1)
	db.Create(&ws2)

	// 建立成员关系：A 在 WS1，B 在 WS2
	db.Create(&model.WorkspaceMember{WorkspaceID: ws1.ID, UserID: userA.ID, Role: 1, IsActive: true})
	db.Create(&model.WorkspaceMember{WorkspaceID: ws2.ID, UserID: userB.ID, Role: 1, IsActive: true})

	return db, nil
}

func TestWorkspaceIsolation(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	cfg := config.Load()
	authSvc := service.NewAuthService(db, cfg)
	
	// 生成 UserA 的 Token
	tokenA, _, _ := authSvc.GenerateTokenForTest(1)
	
	// 初始化服务
	issueSvc := service.NewIssueService(db, nil, nil, nil, nil)
	projectSvc := service.NewProjectService(db)

	// 初始化处理器
	issueHandler := handler.NewIssueHandler(issueSvc)
	projectHandler := handler.NewProjectHandler(projectSvc, nil)

	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	t.Run("UserA cannot list projects in UserB's workspace", func(t *testing.T) {
		// 模拟请求：UserA 尝试列出 WS2 (UserB 的空间) 的项目
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/projects?workspace_id=2", nil)
		c.Request.Header.Set("Authorization", "Bearer "+tokenA)
		
		// 模拟 JWT 中间件注入 UserA (ID=1)
		c.Set("user", model.User{BaseModel: model.BaseModel{ID: 1}})

		projectHandler.List(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("UserA cannot view issues in UserB's workspace", func(t *testing.T) {
		// 创建一个属于 WS2 的问题
		var ws2 model.Workspace
		db.Where("slug = ?", "ws2").First(&ws2)
		
		var proj model.Project
		db.Where("workspace_id = ?", ws2.ID).First(&proj)
		if proj.ID == 0 {
			proj = model.Project{Name: "Test Proj", WorkspaceID: ws2.ID}
			db.Create(&proj)
		}

		issue := model.Issue{Name: "Secret Issue", ProjectID: proj.ID, WorkspaceID: ws2.ID}
		db.Create(&issue)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "issueId", Value: strconv.FormatUint(issue.ID, 10)}} 
		c.Request, _ = http.NewRequest("GET", "/issues/"+strconv.FormatUint(issue.ID, 10), nil)
		c.Request.Header.Set("Authorization", "Bearer "+tokenA)
		
		// 模拟 JWT 中间件注入 UserA (ID: 1)
		c.Set("user", model.User{BaseModel: model.BaseModel{ID: 1}})

		issueHandler.Get(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
