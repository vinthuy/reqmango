# MCP Server 重写 + CLI 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重写 reqmango MCP server（24 tools、stdio + HTTP 双传输）并新建 `reqmango` 日常操作 CLI，两者共享 sdk/client，认证基于后端新增的 PAT 系统。

**Architecture:** 后端新增 `personal_access_tokens` 表 + `/auth/tokens` 端点 + AuthMiddleware 兼容 PAT；新建 `sdk/` Go module（`github.com/reqmango/tools`，go 1.25.5）含 `client`（共享 API 客户端）、`mcp`（mark3labs/mcp-go v1.0.0）、`cli`（cobra v1.10.2）三个包与两个二进制；最后 e2e 冒烟 + 文档 + 删除旧 mcp-server 目录。

**Tech Stack:** Go 1.25.5（sdk）/ Go 1.24（backend）、Gin、GORM、golang-migrate、go-sqlmock、mark3labs/mcp-go v1.0.0、cobra v1.10.2、golang.org/x/term。

**Spec:** `docs/superpowers/specs/2026-09-06-mcp-cli-design.md` — 本计划从 spec 展开，执行者需同时阅读两者。

## Global Constraints

- 工作分支 `feat/mcp-cli`，所有提交落在此分支（禁止 master）。
- sdk 模块路径 `github.com/reqmango/tools`，`go.mod` 中 `go 1.25.5`（mcp-go v1.0.0 硬性要求）。
- 环境变量：`REQMANGO_API_URL`（默认 `http://localhost:8000/api/v1`）、`REQMANGO_PAT`（MCP 必填）。
- PAT 明文 token 前缀 `reqmango_pat_`；数据库只存 SHA-256 哈希；明文仅创建响应返回一次。
- 后端错误响应体格式统一为 `{"message": "..."}`（gin.H，不输出 error_code）；成功响应按端点分三种：裸对象、裸数组 + `X-Total-Count` 头（issues）、包裹 `{items,total,limit,offset}`（cycles）。
- 后端测试沿用 sqlmock（`testutil.NewMockDB`）+ `c.Set("currentUser", ...)` 模式；sdk 测试用 `httptest.Server` 模拟后端。
- MCP 工具失败返回 `mcp.NewToolResultError(...)`（isError: true 的 tool result），不返回 Go error 中断协议会话。
- 每次提交必须通过对应测试：backend `cd backend && go test ./internal/...`；sdk `cd sdk && go test ./...`。

---

## 文件结构总览

**后端（Phase A）**
- Create: `backend/internal/model/pat.go` — PAT GORM 模型
- Create: `backend/migrations/000020_personal_access_tokens.up.sql` / `.down.sql`
- Create: `backend/internal/service/pat_service.go` — token 生成/哈希/CRUD/Authenticate
- Create: `backend/internal/service/pat_service_test.go`
- Create: `backend/internal/handler/pat_handler.go` — /auth/tokens 三端点
- Create: `backend/internal/handler/pat_handler_test.go`
- Modify: `backend/internal/router/router.go` — 注册 tokens 路由组
- Modify: `backend/internal/middleware/auth.go` — PAT 前缀分支
- Create: `backend/internal/middleware/auth_pat_test.go`

**sdk 模块（Phase B/C/D）**
- Create: `sdk/go.mod` `sdk/go.sum`
- Create: `sdk/client/client.go` `errors.go` — 核心 do()/APIError
- Create: `sdk/client/client_test.go`
- Create: `sdk/client/auth.go` `workspaces.go` `meta.go` + `_test.go`
- Create: `sdk/client/issues.go` + `issues_test.go`
- Create: `sdk/client/cycles.go` + `cycles_test.go`
- Create: `sdk/client/ai.go` `chat.go` `agents.go` + `_test.go`（SSE 聚合）
- Create: `sdk/mcp/server.go` `tools_core.go` `tools_ai.go` + `_test.go`
- Create: `sdk/cmd/reqmango-mcp/main.go`
- Create: `sdk/cli/config.go` `root.go` `output.go` `cmd_auth.go` `cmd_issue.go` `cmd_project.go` `cmd_cycle.go` `cmd_meta.go` `cmd_agent.go` `cmd_ask.go` + `_test.go`
- Create: `sdk/cmd/reqmango/main.go`

**收尾（Phase E）**
- Create: `sdk/README.md`、`scripts/e2e_tools.sh`
- Modify: `Makefile` — `tools`/`test-tools` 目标
- Delete: `mcp-server/`（整个目录）

---

## Phase A：后端 PAT 系统

### Task 1: PAT 模型 + 迁移

**Files:**
- Create: `backend/internal/model/pat.go`
- Create: `backend/migrations/000020_personal_access_tokens.up.sql`
- Create: `backend/migrations/000020_personal_access_tokens.down.sql`

**Interfaces:**
- Consumes: 无
- Produces: `model.PersonalAccessToken`（被 Task 2 的 service 引用，字段名见下）

- [ ] **Step 1: 写模型**

`backend/internal/model/pat.go`：

```go
package model

import "time"

// PersonalAccessToken is a long-lived API credential issued to a user.
// The plaintext token is shown only once at creation; only the SHA-256
// hash is stored.
type PersonalAccessToken struct {
	BaseModel

	UserID      uint64     `gorm:"not null;index" json:"user_id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	TokenPrefix string     `gorm:"size:20;not null;index" json:"token_prefix"` // 展示用前缀
	TokenHash   string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	Scopes      string     `gorm:"type:text;default:''" json:"scopes"` // 预留，首版空
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"` // null = 永不过期
	RevokedAt   *time.Time `json:"revoked_at"`
}

func (PersonalAccessToken) TableName() string { return "personal_access_tokens" }
```

- [ ] **Step 2: 写迁移**

`backend/migrations/000020_personal_access_tokens.up.sql`：

```sql
-- Personal Access Tokens for CLI / MCP / CI authentication
CREATE TABLE personal_access_tokens (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by_id BIGINT,
    updated_by_id BIGINT,

    user_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    token_prefix VARCHAR(20) NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    scopes TEXT DEFAULT '',
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_personal_access_tokens_token_hash ON personal_access_tokens(token_hash);
CREATE INDEX idx_personal_access_tokens_user_id ON personal_access_tokens(user_id);
```

`backend/migrations/000020_personal_access_tokens.down.sql`：

```sql
DROP TABLE IF EXISTS personal_access_tokens;
```

- [ ] **Step 3: 编译 + 迁移验证**

Run: `cd backend && go build ./...`
Expected: PASS

Run（本地 dev 数据库已起时）: `cd backend && go run ./cmd/migrate up` 然后 `go run ./cmd/migrate down`（回滚一步）再 `go run ./cmd/migrate up`
Expected: 无报错；`docker-compose exec db psql -U reqmango -d reqmango -c '\d personal_access_tokens'` 能看到表。

- [ ] **Step 4: Commit**

```bash
git add backend/internal/model/pat.go backend/migrations/000020_personal_access_tokens.up.sql backend/migrations/000020_personal_access_tokens.down.sql
git commit -m "feat(pat): add PersonalAccessToken model and migration 000020"
```

### Task 2: PAT Service

**Files:**
- Create: `backend/internal/service/pat_service.go`
- Create: `backend/internal/service/pat_service_test.go`

**Interfaces:**
- Consumes: `model.PersonalAccessToken`（Task 1）、`common.AppError`、`testutil.NewMockDB`
- Produces（Task 3/4 依赖）:
  - `const PATPrefix = "reqmango_pat_"`
  - `func NewPATService(db *gorm.DB) *PATService`
  - `func (s *PATService) Create(userID uint64, name string, expiresAt *time.Time) (string, *model.PersonalAccessToken, error)`
  - `func (s *PATService) List(userID uint64) ([]model.PersonalAccessToken, error)`
  - `func (s *PATService) Revoke(userID, id uint64) error`
  - `func (s *PATService) Authenticate(tokenStr string) (*model.User, error)`
  - `func HashToken(token string) string`

- [ ] **Step 1: 写失败测试**

`backend/internal/service/pat_service_test.go`：

```go
package service

import (
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/testutil"
)

func TestPATService_Create(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewPATService(db)
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "personal_access_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	token, pat, err := svc.Create(1, "cli-macbook", nil)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(token, PATPrefix))
	assert.Equal(t, "cli-macbook", pat.Name)
	assert.Equal(t, HashToken(token), pat.TokenHash)
	assert.Equal(t, token[:len(PATPrefix)+4], pat.TokenPrefix)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPATService_List(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewPATService(db)
	mock.ExpectQuery(`SELECT \* FROM "personal_access_tokens" WHERE`).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "cli-macbook"))

	pats, err := svc.List(1)
	require.NoError(t, err)
	assert.Len(t, pats, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPATService_Revoke_NotFound(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewPATService(db)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "personal_access_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := svc.Revoke(1, 999)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPATService_Authenticate_Valid(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewPATService(db)
	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM "personal_access_tokens" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at"}).
			AddRow(1, 2, HashToken("reqmango_pat_abc"), nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "is_active"}).AddRow(2, true))
	mock.ExpectExec(`UPDATE "personal_access_tokens" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	user, err := svc.Authenticate("reqmango_pat_abc")
	require.NoError(t, err)
	assert.Equal(t, uint64(2), user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPATService_Authenticate_Revoked(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewPATService(db)
	revoked := time.Now().Add(-time.Hour)
	mock.ExpectQuery(`SELECT \* FROM "personal_access_tokens" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at"}).
			AddRow(1, 2, HashToken("reqmango_pat_abc"), nil, &revoked))

	_, err := svc.Authenticate("reqmango_pat_abc")
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./internal/service/ -run TestPATService -v`
Expected: FAIL（`pat_service.go` 不存在 / `NewPATService` undefined）

- [ ] **Step 3: 实现 service**

`backend/internal/service/pat_service.go`：

```go
package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
)

// PATPrefix marks a personal access token in the Authorization header.
const PATPrefix = "reqmango_pat_"

// PATService manages personal access tokens.
type PATService struct {
	db *gorm.DB
}

func NewPATService(db *gorm.DB) *PATService { return &PATService{db: db} }

// HashToken returns the hex SHA-256 hash of a plaintext token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Create issues a new PAT for the user. The plaintext token is returned
// exactly once and is never stored.
func (s *PATService) Create(userID uint64, name string, expiresAt *time.Time) (string, *model.PersonalAccessToken, error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, common.Internal("failed to generate token")
	}
	token := PATPrefix + hex.EncodeToString(raw)

	pat := &model.PersonalAccessToken{
		UserID:      userID,
		Name:        name,
		TokenPrefix: token[:len(PATPrefix)+4],
		TokenHash:   HashToken(token),
		ExpiresAt:   expiresAt,
	}
	if err := s.db.Create(pat).Error; err != nil {
		return "", nil, common.Internal("failed to save token")
	}
	return token, pat, nil
}

// List returns the user's PATs. TokenHash is already hidden by its json:"-" tag.
func (s *PATService) List(userID uint64) ([]model.PersonalAccessToken, error) {
	var pats []model.PersonalAccessToken
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&pats).Error; err != nil {
		return nil, common.Internal("failed to list tokens")
	}
	return pats, nil
}

// Revoke marks a token revoked. Returns ErrNotFound if it is not owned by the user.
func (s *PATService) Revoke(userID, id uint64) error {
	res := s.db.Model(&model.PersonalAccessToken{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("revoked_at", time.Now())
	if res.Error != nil {
		return common.Internal("failed to revoke token")
	}
	if res.RowsAffected == 0 {
		return common.NotFound("token not found")
	}
	return nil
}

// Authenticate resolves a plaintext PAT to its user, checking revocation,
// expiry and account status.
func (s *PATService) Authenticate(tokenStr string) (*model.User, error) {
	if !strings.HasPrefix(tokenStr, PATPrefix) {
		return nil, common.Unauthorized("invalid token")
	}

	var pat model.PersonalAccessToken
	if err := s.db.Where("token_hash = ? AND revoked_at IS NULL", HashToken(tokenStr)).First(&pat).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.Unauthorized("invalid token")
		}
		return nil, common.Internal("failed to look up token")
	}
	if pat.ExpiresAt != nil && time.Now().After(*pat.ExpiresAt) {
		return nil, common.Unauthorized("token expired")
	}

	var user model.User
	if err := s.db.First(&user, pat.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.Unauthorized("user not found")
		}
		return nil, common.Internal("failed to load user")
	}
	if !user.IsActive || user.DeletedAt.Valid {
		return nil, common.Unauthorized("account is disabled")
	}

	// Fire the last_used_at refresh (blocking keeps sqlmock tests simple).
	if err := s.db.Model(&pat).Update("last_used_at", time.Now()).Error; err != nil {
		return nil, common.Internal("failed to refresh token")
	}
	return &user, nil
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./internal/service/ -run TestPATService -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/pat_service.go backend/internal/service/pat_service_test.go
git commit -m "feat(pat): add PATService with create/list/revoke/authenticate"
```

### Task 3: PAT Handler + 路由

**Files:**
- Create: `backend/internal/handler/pat_handler.go`
- Create: `backend/internal/handler/pat_handler_test.go`
- Create: `backend/internal/dto/response/pat.go`
- Modify: `backend/internal/router/router.go`（auth 组区域，约 L213-219 之后）

**Interfaces:**
- Consumes: Task 2 的 `PATService`、`middleware.GetCurrentUser`
- Produces（Task 15 e2e 依赖）:
  - `GET /api/v1/auth/tokens` → 200 `[]PATResponse`
  - `POST /api/v1/auth/tokens` body `{"name": "...", "expires_at": "..."|null}` → 201 `{"token": "...", "id": 1, "name": "...", "token_prefix": "...", "expires_at": null, "created_at": "..."}`
  - `DELETE /api/v1/auth/tokens/:id` → 200 `{"message": "revoked"}`

- [ ] **Step 1: 写响应 DTO + 失败测试**

`backend/internal/dto/response/pat.go`：

```go
package response

import "time"

// PATResponse is the safe (hash-less) view of a personal access token.
type PATResponse struct {
	ID          uint64     `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	Scopes      string     `json:"scopes"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	RevokedAt   *time.Time `json:"revoked_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PATCreateResponse is returned once when a token is created.
type PATCreateResponse struct {
	Token string `json:"token"`
	PATResponse
}
```

`backend/internal/handler/pat_handler_test.go`：

```go
package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "personal_access_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

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

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "personal_access_tokens" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

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
```

（注：测试需 `import "strings"`。）

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./internal/handler/ -run TestPATHandler -v`
Expected: FAIL（pat_handler.go 不存在）

- [ ] **Step 3: 实现 handler**

`backend/internal/handler/pat_handler.go`：

```go
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
)

// PATHandler serves /auth/tokens endpoints.
type PATHandler struct {
	svc *service.PATService
}

func NewPATHandler(svc *service.PATService) *PATHandler { return &PATHandler{svc: svc} }

func toPATResponse(p model.PersonalAccessToken) response.PATResponse {
	return response.PATResponse{
		ID:          p.ID,
		Name:        p.Name,
		TokenPrefix: p.TokenPrefix,
		Scopes:      p.Scopes,
		LastUsedAt:  p.LastUsedAt,
		ExpiresAt:   p.ExpiresAt,
		RevokedAt:   p.RevokedAt,
		CreatedAt:   p.CreatedAt,
	}
}

// List handles GET /auth/tokens.
func (h *PATHandler) List(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	pats, err := h.svc.List(user.ID)
	if err != nil {
		writeAppError(c, err)
		return
	}
	out := make([]response.PATResponse, 0, len(pats))
	for _, p := range pats {
		out = append(out, toPATResponse(p))
	}
	c.JSON(http.StatusOK, out)
}

// Create handles POST /auth/tokens.
func (h *PATHandler) Create(c *gin.Context) {
	var req struct {
		Name      string     `json:"name" binding:"required,min=1,max=100"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	user := middleware.GetCurrentUser(c)
	token, pat, err := h.svc.Create(user.ID, req.Name, req.ExpiresAt)
	if err != nil {
		writeAppError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.PATCreateResponse{Token: token, PATResponse: toPATResponse(*pat)})
}

// Revoke handles DELETE /auth/tokens/:id.
func (h *PATHandler) Revoke(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid token id"})
		return
	}
	user := middleware.GetCurrentUser(c)
	if err := h.svc.Revoke(user.ID, id); err != nil {
		writeAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "revoked"})
}

// writeAppError maps common.AppError to the project's {"message": ...} format.
func writeAppError(c *gin.Context, err error) {
	if appErr, ok := err.(*common.AppError); ok {
		c.JSON(appErr.Code, gin.H{"message": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
}
```

- [ ] **Step 4: 注册路由**

在 `backend/internal/router/router.go` 的 auth 组（约 L213-219）之后添加：

```go
		// ---- Personal Access Tokens ----
		patSvc := service.NewPATService(db)
		patH := handler.NewPATHandler(patSvc)
		tokens := v1.Group("/auth/tokens", authMiddleware)
		{
			tokens.GET("", patH.List)
			tokens.POST("", patH.Create)
			tokens.DELETE("/:id", patH.Revoke)
		}
```

- [ ] **Step 5: 运行测试确认通过**

Run: `cd backend && go test ./internal/handler/ -run TestPATHandler -v && go build ./...`
Expected: PASS（注意：若 `writeAppError` 与其他 handler 文件重名会编译失败，若出现则改名为 `writePATAppError` 并同步测试无需改动）

- [ ] **Step 6: Commit**

```bash
git add backend/internal/handler/pat_handler.go backend/internal/handler/pat_handler_test.go backend/internal/dto/response/pat.go backend/internal/router/router.go
git commit -m "feat(pat): add /auth/tokens endpoints and routes"
```

### Task 4: AuthMiddleware 支持 PAT

**Files:**
- Modify: `backend/internal/middleware/auth.go`
- Create: `backend/internal/middleware/auth_pat_test.go`

**Interfaces:**
- Consumes: Task 2 的 `service.PATPrefix` / `PATService.Authenticate`
- Produces: `Authorization: Bearer reqmango_pat_*` 请求可认证（Task 15 e2e 依赖）

- [ ] **Step 1: 写失败测试**

`backend/internal/middleware/auth_pat_test.go`：

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"github.com/reqmango/backend/internal/testutil"
)

func TestAuthMiddleware_PAT_Valid(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ping", AuthMiddleware(db, "secret"), func(c *gin.Context) {
		u := GetCurrentUser(c)
		c.JSON(http.StatusOK, gin.H{"uid": u.ID})
	})

	mock.ExpectQuery(`SELECT \* FROM "personal_access_tokens" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at"}).
			AddRow(1, 7, service.HashToken("reqmango_pat_testtoken"), nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "is_active"}).AddRow(7, true))
	mock.ExpectExec(`UPDATE "personal_access_tokens" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer reqmango_pat_testtoken")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"uid":7`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthMiddleware_PAT_Invalid(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ping", AuthMiddleware(db, "secret"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	mock.ExpectQuery(`SELECT \* FROM "personal_access_tokens" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at"}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer reqmango_pat_bogus")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

var _ = model.User{} // keep model import for future tests
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./internal/middleware/ -run TestAuthMiddleware_PAT -v`
Expected: FAIL（PAT 分支不存在，`reqmango_pat_testtoken` 会被当作 JWT 解析失败 → 401）

- [ ] **Step 3: 修改 AuthMiddleware**

在 `backend/internal/middleware/auth.go` 的 `tokenStr := parts[1]` 之后、JWT 解析之前插入：

```go
		// Personal access tokens take precedence: long-lived credentials
		// stored as SHA-256 hashes (used by CLI / MCP / CI).
		if strings.HasPrefix(tokenStr, service.PATPrefix) {
			patSvc := service.NewPATService(db)
			user, err := patSvc.Authenticate(tokenStr)
			if err != nil {
				if appErr, ok := err.(*common.AppError); ok {
					c.JSON(appErr.Code, gin.H{"message": appErr.Message})
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"message": msg(c, "unauthorized", "Invalid token")})
				}
				c.Abort()
				return
			}
			c.Set("currentUser", user)
			c.Next()
			return
		}
```

并确认 import 区新增 `"github.com/reqmango/backend/internal/service"`（`common` 已 import 则无需再加；若无 `common` 则补上）。

- [ ] **Step 4: 运行全部中间件 + service 测试**

Run: `cd backend && go test ./internal/middleware/ ./internal/service/ -v`
Expected: PASS（既有 JWT 测试不受影响——签名未变，PAT 分支仅前缀触发）

- [ ] **Step 5: Commit**

```bash
git add backend/internal/middleware/auth.go backend/internal/middleware/auth_pat_test.go
git commit -m "feat(pat): accept PAT bearer tokens in AuthMiddleware"
```

---

## Phase B：sdk 共享客户端

### Task 5: sdk 模块骨架 + client 核心

**Files:**
- Create: `sdk/go.mod`、`sdk/client/client.go`、`sdk/client/errors.go`、`sdk/client/client_test.go`

**Interfaces:**
- Consumes: 无
- Produces（所有后续 sdk 任务依赖）:
  - `client.DefaultBaseURL = "http://localhost:8000/api/v1"`
  - `func New(baseURL, token string) *Client`（baseURL 为空 → DefaultBaseURL）
  - `func (c *Client) GetJSON(ctx context.Context, path string, query url.Values, out any) (http.Header, error)`
  - `func (c *Client) PostJSON(ctx context.Context, path string, query url.Values, body, out any) (http.Header, error)`
  - `func (c *Client) PutJSON(ctx context.Context, path string, query url.Values, body, out any) (http.Header, error)`
  - `func (c *Client) DeleteJSON(ctx context.Context, path string, query url.Values, out any) error`
  - `type APIError struct { StatusCode int; Message string; Body map[string]any }`，实现 `Error() string`
  - `func AsAPIError(err error) *APIError`（errors.As 包装）

- [ ] **Step 1: 初始化模块**

```bash
mkdir -p sdk/client sdk/mcp sdk/cli sdk/cmd/reqmango sdk/cmd/reqmango-mcp
cd sdk && go mod init github.com/reqmango/tools
```
将 `go.mod` 中 `go` 指令改为 `go 1.25.5`（mcp-go v1.0.0 要求）。

- [ ] **Step 2: 写失败测试**

`sdk/client/client_test.go`：

```go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type pingResp struct {
	Pong string `json:"pong"`
}

func TestDo_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer reqmango_pat_abc" {
			t.Errorf("missing auth header, got %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/api/v1/ping" || r.URL.Query().Get("a") != "1" {
			t.Errorf("unexpected request %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("X-Total-Count", "42")
		json.NewEncoder(w).Encode(pingResp{Pong: "ok"})
	}))
	defer srv.Close()

	c := New(srv.URL, "reqmango_pat_abc")
	var out pingResp
	hdr, err := c.GetJSON(context.Background(), "/ping", map[string]string{"a": "1"}, &out)
	if err != nil {
		t.Fatalf("GetJSON: %v", err)
	}
	if out.Pong != "ok" || hdr.Get("X-Total-Count") != "42" {
		t.Fatalf("unexpected out=%+v hdr=%v", out, hdr)
	}
}

func TestDo_APIError_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "token expired"})
	}))
	defer srv.Close()

	c := New(srv.URL, "reqmango_pat_abc")
	var out pingResp
	_, err := c.GetJSON(context.Background(), "/ping", nil, &out)
	apiErr := AsAPIError(err)
	if apiErr == nil {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized || apiErr.Message != "token expired" {
		t.Fatalf("unexpected APIError: %+v", apiErr)
	}
}

func TestDo_APIError_409_ExtraFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{
			"message":     "approval_required",
			"transition_id": float64(9),
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "reqmango_pat_abc")
	_, err := c.GetJSON(context.Background(), "/ping", nil, &pingResp{})
	apiErr := AsAPIError(err)
	if apiErr == nil || apiErr.Body["transition_id"] != float64(9) {
		t.Fatalf("expected Body to carry extra fields, got %+v", apiErr)
	}
}

func TestNew_DefaultBaseURL(t *testing.T) {
	c := New("", "t")
	if c.baseURL != DefaultBaseURL {
		t.Fatalf("expected default baseURL, got %q", c.baseURL)
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `cd sdk && go test ./client/ -v`
Expected: FAIL（client.go 不存在）

- [ ] **Step 4: 实现核心**

`sdk/client/errors.go`：

```go
package client

import (
	"errors"
	"fmt"
)

// APIError is a typed error carrying the backend's {"message": ...} body.
type APIError struct {
	StatusCode int
	Message    string
	Body       map[string]any // full parsed error body (e.g. 409 approval payload)
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error %d: %s", e.StatusCode, e.Message)
}

// AsAPIError unwraps err to *APIError, or nil.
func AsAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}
```

`sdk/client/client.go`：

```go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL points at a local reqmango backend.
const DefaultBaseURL = "http://localhost:8000/api/v1"

// Client is the shared HTTP client for the reqmango REST API.
// It only depends on the standard library.
type Client struct {
	baseURL string // no trailing slash
	token   string
	hc      *http.Client
}

// New creates a client. Empty baseURL uses DefaultBaseURL.
func New(baseURL, token string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		hc:      &http.Client{Timeout: 30 * time.Second},
	}
}

// BaseURL returns the configured base URL (for config display).
func (c *Client) BaseURL() string { return c.baseURL }

// do performs one request. On 2xx with JSON body and non-nil out it decodes
// into out and returns response headers. On failure it returns *APIError.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) (http.Header, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode, Body: map[string]any{}}
		if len(data) > 0 {
			var m map[string]any
			if json.Unmarshal(data, &m) == nil {
				apiErr.Body = m
				if msg, ok := m["message"].(string); ok {
					apiErr.Message = msg
				}
			}
		}
		if apiErr.Message == "" {
			apiErr.Message = strings.TrimSpace(string(data))
		}
		return nil, apiErr
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
	}
	return resp.Header, nil
}

// GetJSON performs GET and decodes the JSON body into out.
func (c *Client) GetJSON(ctx context.Context, path string, query url.Values, out any) (http.Header, error) {
	return c.do(ctx, http.MethodGet, path, query, nil, out)
}

// PostJSON performs POST with a JSON body and decodes the response into out.
func (c *Client) PostJSON(ctx context.Context, path string, query url.Values, body, out any) (http.Header, error) {
	return c.do(ctx, http.MethodPost, path, query, body, out)
}

// PutJSON performs PUT with a JSON body and decodes the response into out.
func (c *Client) PutJSON(ctx context.Context, path string, query url.Values, body, out any) (http.Header, error) {
	return c.do(ctx, http.MethodPut, path, query, body, out)
}

// DeleteJSON performs DELETE. out may be nil.
func (c *Client) DeleteJSON(ctx context.Context, path string, query url.Values, out any) error {
	_, err := c.do(ctx, http.MethodDelete, path, query, nil, out)
	return err
}
```

- [ ] **Step 5: 运行测试确认通过**

Run: `cd sdk && go test ./client/ -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add sdk/go.mod sdk/client/client.go sdk/client/errors.go sdk/client/client_test.go
git commit -m "feat(sdk): scaffold tools module with shared HTTP client core"
```

### Task 6: auth / workspace / project / meta 客户端

**Files:**
- Create: `sdk/client/auth.go`、`sdk/client/workspaces.go`、`sdk/client/meta.go`
- Create: `sdk/client/auth_test.go`、`sdk/client/workspaces_test.go`、`sdk/client/meta_test.go`

**Interfaces:**
- Consumes: Task 5 的 `Client.GetJSON/PostJSON/DeleteJSON`、`APIError`
- Produces（Task 13-15 依赖）:
  - auth: `Login(ctx, email, password) (*TokenResponse, error)`、`Me(ctx) (*User, error)`、`CreatePAT(ctx, req CreatePATRequest) (*CreatePATResponse, error)`、`ListPATs(ctx) ([]PATResponse, error)`、`RevokePAT(ctx, id uint64) error`
  - workspaces: `ListWorkspaces(ctx) ([]Workspace, error)`、`ListMembers(ctx, workspaceID uint64) ([]Member, error)`、`ListProjects(ctx, workspaceID uint64) ([]Project, error)`、`GetProject(ctx, id uint64) (*Project, error)`
  - meta: `ListStates(ctx, projectID uint64) ([]State, error)`、`ListLabels(ctx, projectID uint64) ([]Label, error)`、`ListIssueTypes(ctx, workspaceID, projectID uint64) ([]IssueType, error)`、`ListPages(ctx, projectID uint64) ([]Page, error)`、`GetPage(ctx, pageID uint64) (*Page, error)`、`ListNotifications(ctx, unreadOnly bool, limit, offset int) ([]Notification, error)`

- [ ] **Step 1: 写失败测试（每个文件一个表驱动测试）**

`sdk/client/auth_test.go`（其余测试文件结构相同，见下方实现后自行对应）：

```go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/login" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["email"] != "a@b.c" || body["password"] != "pw" {
			t.Errorf("unexpected body %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"access_token": "jwt", "token_type": "Bearer", "expires_at": "2026-09-13T00:00:00Z"})
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	tok, err := c.Login(context.Background(), "a@b.c", "pw")
	if err != nil || tok.AccessToken != "jwt" {
		t.Fatalf("Login: %v %+v", err, tok)
	}
}

func TestCreatePAT(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/tokens" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer jwt" {
			t.Errorf("expected JWT auth, got %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"token": "reqmango_pat_x", "id": 3, "name": "cli",
			"token_prefix": "reqmango_pat_ab3d", "created_at": "2026-09-06T00:00:00Z",
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "jwt")
	resp, err := c.CreatePAT(context.Background(), CreatePATRequest{Name: "cli"})
	if err != nil || resp.Token != "reqmango_pat_x" || resp.ID != 3 {
		t.Fatalf("CreatePAT: %v %+v", err, resp)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd sdk && go test ./client/ -run "TestLogin|TestCreatePAT" -v`
Expected: FAIL

- [ ] **Step 3: 实现三个文件**

`sdk/client/auth.go`：

```go
package client

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// TokenResponse is the backend login response.
type TokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Login exchanges credentials for a JWT (used once to mint a PAT).
func (c *Client) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	var out TokenResponse
	_, err := c.PostJSON(ctx, "/auth/login", nil,
		map[string]string{"email": email, "password": password}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// User is the current-user shape returned by /auth/me.
type User struct {
	ID          uint64 `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// Me returns the user behind the configured token.
func (c *Client) Me(ctx context.Context) (*User, error) {
	var out User
	if _, err := c.GetJSON(ctx, "/auth/me", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreatePATRequest is the body for POST /auth/tokens.
type CreatePATRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// PATResponse is the safe view of a token.
type PATResponse struct {
	ID          uint64     `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	Scopes      string     `json:"scopes"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	RevokedAt   *time.Time `json:"revoked_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreatePATResponse carries the one-time plaintext token.
type CreatePATResponse struct {
	Token string `json:"token"`
	PATResponse
}

// CreatePAT mints a PAT using the current (JWT) token.
func (c *Client) CreatePAT(ctx context.Context, req CreatePATRequest) (*CreatePATResponse, error) {
	var out CreatePATResponse
	_, err := c.PostJSON(ctx, "/auth/tokens", nil, req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPATs returns the user's tokens.
func (c *Client) ListPATs(ctx context.Context) ([]PATResponse, error) {
	var out []PATResponse
	if _, err := c.GetJSON(ctx, "/auth/tokens", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RevokePAT revokes a token by ID.
func (c *Client) RevokePAT(ctx context.Context, id uint64) error {
	return c.DeleteJSON(ctx, "/auth/tokens/"+strconv.FormatUint(id, 10), nil, nil)
}

var _ = url.Values{} // silence unused import if trimmed
```

`sdk/client/workspaces.go`：

```go
package client

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// Workspace is the list shape of GET /workspaces.
type Workspace struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListWorkspaces returns the workspaces visible to the user.
func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	var out []Workspace
	if _, err := c.GetJSON(ctx, "/workspaces", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UserLite is the embedded user shape in member/issue responses.
type UserLite struct {
	ID          uint64 `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
}

// Member is a workspace member.
type Member struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspace_id"`
	UserID      uint64    `json:"user_id"`
	Role        int       `json:"role"`
	IsActive    bool      `json:"is_active"`
	User        UserLite  `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListMembers returns workspace members.
func (c *Client) ListMembers(ctx context.Context, workspaceID uint64) ([]Member, error) {
	var out []Member
	_, err := c.GetJSON(ctx, "/workspaces/"+strconv.FormatUint(workspaceID, 10)+"/members", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Project is the list shape of GET /projects.
type Project struct {
	ID          uint64     `json:"id"`
	Name        string     `json:"name"`
	Identifier  string     `json:"identifier"`
	Description string     `json:"description"`
	WorkspaceID uint64     `json:"workspace_id"`
	ArchivedAt  *time.Time `json:"archived_at"`
	TotalIssues int64      `json:"total_issues"`
	TotalMembers int64     `json:"total_members"`
	IsFavorite  bool       `json:"is_favorite"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ListProjects returns projects in a workspace (workspace_id query is required).
func (c *Client) ListProjects(ctx context.Context, workspaceID uint64) ([]Project, error) {
	q := url.Values{}
	q.Set("workspace_id", strconv.FormatUint(workspaceID, 10))
	var out []Project
	if _, err := c.GetJSON(ctx, "/projects", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProject fetches one project by numeric ID.
// (Route GET /projects/:projectId — confirm in router.go near L596-640 if missing.)
func (c *Client) GetProject(ctx context.Context, id uint64) (*Project, error) {
	var out Project
	if _, err := c.GetJSON(ctx, "/projects/"+strconv.FormatUint(id, 10), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
```

`sdk/client/meta.go`：

```go
package client

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// State is a workflow state.
type State struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Group       string `json:"group"`
	Sequence    int    `json:"sequence"`
	IsDefault   bool   `json:"is_default"`
	IsActive    bool   `json:"is_active"`
	ProjectID   uint64 `json:"project_id"`
	WorkspaceID uint64 `json:"workspace_id"`
}

// ListStates returns project workflow states (GET /projects/:id/settings/states).
func (c *Client) ListStates(ctx context.Context, projectID uint64) ([]State, error) {
	var out []State
	_, err := c.GetJSON(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/settings/states", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Label is a project label.
type Label struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
	ProjectID   uint64 `json:"project_id"`
	WorkspaceID uint64 `json:"workspace_id"`
}

// ListLabels returns project labels (GET /projects/:id/settings/labels).
func (c *Client) ListLabels(ctx context.Context, projectID uint64) ([]Label, error) {
	var out []Label
	_, err := c.GetJSON(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/settings/labels", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IssueType is an issue type definition.
type IssueType struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Color       string  `json:"color"`
	Icon        string  `json:"icon"`
	Description string  `json:"description"`
	Level       string  `json:"level"`
	ParentTypeID *uint64 `json:"parent_type_id"`
	IsDefault   bool    `json:"is_default"`
	Sequence    int     `json:"sequence"`
	IsActive    bool    `json:"is_active"`
	ProjectID   uint64  `json:"project_id"`
	WorkspaceID uint64  `json:"workspace_id"`
}

// ListIssueTypes returns issue types (GET /issue-types?workspace_id=..., workspace_id required).
func (c *Client) ListIssueTypes(ctx context.Context, workspaceID, projectID uint64) ([]IssueType, error) {
	q := url.Values{}
	q.Set("workspace_id", strconv.FormatUint(workspaceID, 10))
	if projectID != 0 {
		q.Set("project_id", strconv.FormatUint(projectID, 10))
	}
	var out []IssueType
	if _, err := c.GetJSON(ctx, "/issue-types", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Page is a project document.
type Page struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Published   bool      `json:"published"`
	Sequence    int       `json:"sequence"`
	ParentID    *uint64   `json:"parent_id"`
	Depth       int       `json:"depth"`
	ProjectID   uint64    `json:"project_id"`
	WorkspaceID uint64    `json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListPages returns project documents (GET /projects/:id/pages).
func (c *Client) ListPages(ctx context.Context, projectID uint64) ([]Page, error) {
	var out []Page
	_, err := c.GetJSON(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/pages", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetPage fetches one page (GET /projects/:pid/pages/:pageId).
func (c *Client) GetPage(ctx context.Context, projectID, pageID uint64) (*Page, error) {
	path := "/projects/" + strconv.FormatUint(projectID, 10) + "/pages/" + strconv.FormatUint(pageID, 10)
	var out Page
	if _, err := c.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Notification is a user notification.
type Notification struct {
	ID        uint64     `json:"id"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	Type      string     `json:"type"`
	Priority  string     `json:"priority"`
	IsRead    bool       `json:"is_read"`
	ReadAt    *time.Time `json:"read_at"`
	ActionURL string     `json:"action_url"`
	IssueID   *uint64    `json:"issue_id"`
	CreatedAt time.Time  `json:"created_at"`
}

// ListNotifications returns the user's notifications (GET /notifications).
func (c *Client) ListNotifications(ctx context.Context, unreadOnly bool, limit, offset int) ([]Notification, error) {
	q := url.Values{}
	if unreadOnly {
		q.Set("unread_only", "true")
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	var out []Notification
	if _, err := c.GetJSON(ctx, "/notifications", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}
```

补测试：为 `workspaces_test.go` 与 `meta_test.go` 各写 2 个 httptest 表驱动用例（一个验证 path+query 拼装，一个验证解码），模式与 `auth_test.go` 相同（Step 1 已给出模板）。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd sdk && go test ./client/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add sdk/client/auth.go sdk/client/workspaces.go sdk/client/meta.go sdk/client/auth_test.go sdk/client/workspaces_test.go sdk/client/meta_test.go
git commit -m "feat(sdk): add auth/workspace/project/meta client methods"
```

### Task 7: issues 客户端

**Files:**
- Create: `sdk/client/issues.go`、`sdk/client/issues_test.go`

**Interfaces:**
- Consumes: Task 5 核心、Task 6 `UserLite`
- Produces（Task 11/14/15 依赖）:
  - `func (o IssueListOptions) Query() url.Values`
  - `func (c *Client) ListIssues(ctx, opts IssueListOptions) (*IssueListResult, error)`（读 `X-Total-Count`）
  - `func (c *Client) CreateIssue(ctx, projectID, workspaceID uint64, req *CreateIssueRequest) (*Issue, error)`
  - `func (c *Client) GetIssue(ctx, id uint64) (*Issue, error)`
  - `func (c *Client) UpdateIssue(ctx, id uint64, req *UpdateIssueRequest) (*Issue, error)`
  - `func (c *Client) SearchIssues(ctx, workspaceID uint64, query string, projectID *uint64, limit int) ([]IssueSearchResult, error)`
  - `func (c *Client) ResolveIssueCode(ctx, workspaceID uint64, code string) (uint64, error)`（"DEMO-42" → ID）
  - `func (c *Client) AddComment(ctx, issueID uint64, body string, parentID *uint64) (*Comment, error)`
  - `func (c *Client) ListComments(ctx, issueID uint64, page, pageSize int) ([]Comment, int, error)`

- [ ] **Step 1: 写失败测试**

`sdk/client/issues_test.go`（节选，全部用例按此模式写全）：

```go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListIssues_ReadsTotalHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issues" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("rql") != `priority = "high"` {
			t.Errorf("rql not forwarded: %s", r.URL.RawQuery)
		}
		w.Header().Set("X-Total-Count", "7")
		json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "name": "bug"}})
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	res, err := c.ListIssues(context.Background(), IssueListOptions{ProjectID: 1, RQL: `priority = "high"`})
	if err != nil || res.Total != 7 || len(res.Items) != 1 {
		t.Fatalf("ListIssues: %v %+v", err, res)
	}
}

func TestCreateIssue_PathAndBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issues" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("project_id") != "5" || q.Get("workspace_id") != "2" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Login broken" || body["type_id"] != float64(9) {
			t.Errorf("unexpected body %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": 11, "name": "Login broken", "sequence_id": 42})
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	iss, err := c.CreateIssue(context.Background(), 5, 2, &CreateIssueRequest{
		Name: "Login broken", TypeID: uintPtr(9),
	})
	if err != nil || iss.ID != 11 || iss.SequenceID != 42 {
		t.Fatalf("CreateIssue: %v %+v", err, iss)
	}
}

func TestResolveIssueCode(t *testing.T) {
	var step int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/projects":
			json.NewEncoder(w).Encode([]map[string]any{{"id": 5, "identifier": "DEMO", "workspace_id": 2}})
		case r.URL.Path == "/api/v1/issues":
			if r.URL.Query().Get("project_id") != "5" || r.URL.Query().Get("search") != "42" {
				t.Errorf("unexpected issue query %s", r.URL.RawQuery)
			}
			step++
			w.Header().Set("X-Total-Count", "1")
			json.NewEncoder(w).Encode([]map[string]any{{"id": 11, "sequence_id": 42}})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	id, err := c.ResolveIssueCode(context.Background(), 2, "DEMO-42")
	if err != nil || id != 11 {
		t.Fatalf("ResolveIssueCode: %v id=%d", err, id)
	}
}

func uintPtr(v uint64) *uint64 { return &v }
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd sdk && go test ./client/ -run "TestListIssues|TestCreateIssue|TestResolveIssueCode" -v`
Expected: FAIL

- [ ] **Step 3: 实现 issues.go**

`sdk/client/issues.go`：

```go
package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Issue is the response shape of GET /issues/:id (subset of IssueResponse).
type Issue struct {
	ID              uint64     `json:"id"`
	Name            string     `json:"name"`
	DescriptionHTML string     `json:"description_html"`
	Priority        string     `json:"priority"`
	SequenceID      int        `json:"sequence_id"`
	SortOrder       float64    `json:"sort_order"`
	StartDate       *time.Time `json:"start_date"`
	TargetDate      *time.Time `json:"target_date"`
	CompletedAt     *time.Time `json:"completed_at"`
	IsDraft         bool       `json:"is_draft"`
	ArchivedAt      *time.Time `json:"archived_at"`
	ProjectID       uint64     `json:"project_id"`
	WorkspaceID     uint64     `json:"workspace_id"`
	StateID         uint64     `json:"state_id"`
	StateName       string     `json:"state_name"`
	StateGroup      string     `json:"state_group"`
	ParentID        *uint64    `json:"parent_id"`
	Depth           int        `json:"depth"`
	Assignees       []UserLite `json:"assignees"`
	Labels          []uint64   `json:"labels"`
	SubIssuesCount  int64      `json:"sub_issues_count"`
	LinkCount       int        `json:"link_count"`
	EstimatePointID *uint64    `json:"estimate_point_id"`
	CycleID         *uint64    `json:"cycle_id"`
	ModuleIDs       []uint64   `json:"module_ids"`
	ReleaseID       *uint64    `json:"release_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// IssueListOptions maps to the GET /issues query parameters.
type IssueListOptions struct {
	ProjectID   uint64
	WorkspaceID uint64
	StateID     uint64
	Priority    string
	AssigneeID  uint64
	ParentID    uint64
	CycleID     uint64
	IssueTypeID uint64
	ModuleID    uint64
	Search      string
	RQL         string
	SortBy      string
	SortDir     string
	IsDraft     *bool
	Limit       int
	Offset      int
}

// Query encodes non-zero fields into url.Values.
func (o IssueListOptions) Query() url.Values {
	q := url.Values{}
	setU := func(k string, v uint64) {
		if v != 0 {
			q.Set(k, strconv.FormatUint(v, 10))
		}
	}
	setS := func(k, v string) {
		if v != "" {
			q.Set(k, v)
		}
	}
	setU("project_id", o.ProjectID)
	setU("workspace_id", o.WorkspaceID)
	setU("state_id", o.StateID)
	setS("priority", o.Priority)
	setU("assignee_id", o.AssigneeID)
	setU("parent_id", o.ParentID)
	setU("cycle_id", o.CycleID)
	setU("issue_type_id", o.IssueTypeID)
	setU("module_id", o.ModuleID)
	setS("search", o.Search)
	setS("rql", o.RQL)
	setS("sort_by", o.SortBy)
	setS("sort_dir", o.SortDir)
	if o.IsDraft != nil {
		q.Set("is_draft", strconv.FormatBool(*o.IsDraft))
	}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Offset > 0 {
		q.Set("offset", strconv.Itoa(o.Offset))
	}
	return q
}

// IssueListResult pairs the bare-array body with the X-Total-Count header.
type IssueListResult struct {
	Items []Issue
	Total int
}

// ListIssues lists issues. Total comes from the X-Total-Count response header.
func (c *Client) ListIssues(ctx context.Context, opts IssueListOptions) (*IssueListResult, error) {
	var items []Issue
	hdr, err := c.GetJSON(ctx, "/issues", opts.Query(), &items)
	if err != nil {
		return nil, err
	}
	total, _ := strconv.Atoi(hdr.Get("X-Total-Count"))
	return &IssueListResult{Items: items, Total: total}, nil
}

// CreateIssueRequest is the body for POST /issues (project_id & workspace_id go in query).
type CreateIssueRequest struct {
	Name            string   `json:"name"`
	DescriptionHTML string   `json:"description_html,omitempty"`
	Priority        string   `json:"priority,omitempty"`
	StateID         *uint64  `json:"state_id,omitempty"`
	AssigneeIDs     []uint64 `json:"assignee_ids,omitempty"`
	LabelIDs        []uint64 `json:"label_ids,omitempty"`
	StartDate       *string  `json:"start_date,omitempty"`
	TargetDate      *string  `json:"target_date,omitempty"`
	ParentID        *uint64  `json:"parent_id,omitempty"`
	TypeID          *uint64  `json:"type_id,omitempty"`
	CycleID         *uint64  `json:"cycle_id,omitempty"`
}

// CreateIssue creates an issue. projectID and workspaceID are required query params.
func (c *Client) CreateIssue(ctx context.Context, projectID, workspaceID uint64, req *CreateIssueRequest) (*Issue, error) {
	q := url.Values{}
	q.Set("project_id", strconv.FormatUint(projectID, 10))
	q.Set("workspace_id", strconv.FormatUint(workspaceID, 10))
	var out Issue
	if _, err := c.PostJSON(ctx, "/issues", q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIssue fetches one issue.
func (c *Client) GetIssue(ctx context.Context, id uint64) (*Issue, error) {
	var out Issue
	if _, err := c.GetJSON(ctx, "/issues/"+strconv.FormatUint(id, 10), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateIssueRequest mirrors IssueUpdateRequest (pointer fields = partial update;
// slices have replace-all semantics).
type UpdateIssueRequest struct {
	Name            *string  `json:"name,omitempty"`
	DescriptionHTML *string  `json:"description_html,omitempty"`
	Priority        *string  `json:"priority,omitempty"`
	StateID         *uint64  `json:"state_id,omitempty"`
	AssigneeIDs     []uint64 `json:"assignee_ids,omitempty"`
	LabelIDs        []uint64 `json:"label_ids,omitempty"`
	TargetDate      *string  `json:"target_date,omitempty"`
	ParentID        *uint64  `json:"parent_id,omitempty"`
	TypeID          *uint64  `json:"type_id,omitempty"`
	CycleID         *uint64  `json:"cycle_id,omitempty"`
}

// UpdateIssue updates an issue. A 409 (approval flow) surfaces as *APIError
// whose Body carries transition_id/workflow fields.
func (c *Client) UpdateIssue(ctx context.Context, id uint64, req *UpdateIssueRequest) (*Issue, error) {
	var out Issue
	if _, err := c.PutJSON(ctx, "/issues/"+strconv.FormatUint(id, 10), nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// IssueSearchResult is the shape of GET /issues/search.
type IssueSearchResult struct {
	ID                uint64 `json:"id"`
	Name              string `json:"name"`
	SequenceID        int    `json:"sequence_id"`
	ProjectIdentifier string `json:"project_identifier"`
	ProjectID         uint64 `json:"project_id"`
	WorkspaceSlug     string `json:"workspace_slug"`
}

// SearchIssues performs full-text search (workspace_id & query required).
func (c *Client) SearchIssues(ctx context.Context, workspaceID uint64, query string, projectID *uint64, limit int) ([]IssueSearchResult, error) {
	q := url.Values{}
	q.Set("workspace_id", strconv.FormatUint(workspaceID, 10))
	q.Set("query", query)
	if projectID != nil {
		q.Set("project_id", strconv.FormatUint(*projectID, 10))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	var out []IssueSearchResult
	if _, err := c.GetJSON(ctx, "/issues/search", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ResolveIssueCode resolves "DEMO-42" to a numeric issue ID:
// find the project by identifier in the workspace, then match sequence_id.
func (c *Client) ResolveIssueCode(ctx context.Context, workspaceID uint64, code string) (uint64, error) {
	parts := strings.SplitN(code, "-", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid issue code %q (expected IDENTIFIER-NUMBER)", code)
	}
	identifier, seqStr := parts[0], parts[1]
	seq, err := strconv.Atoi(seqStr)
	if err != nil {
		return 0, fmt.Errorf("invalid issue code %q: %w", code, err)
	}

	projects, err := c.ListProjects(ctx, workspaceID)
	if err != nil {
		return 0, err
	}
	var projectID uint64
	for _, p := range projects {
		if strings.EqualFold(p.Identifier, identifier) {
			projectID = p.ID
			break
		}
	}
	if projectID == 0 {
		return 0, fmt.Errorf("project with identifier %q not found", identifier)
	}

	res, err := c.ListIssues(ctx, IssueListOptions{ProjectID: projectID, Search: seqStr, Limit: 100})
	if err != nil {
		return 0, err
	}
	for _, it := range res.Items {
		if it.SequenceID == seq {
			return it.ID, nil
		}
	}
	return 0, fmt.Errorf("issue %s not found", code)
}

// Comment is the shape of POST /comments.
type Comment struct {
	ID         uint64     `json:"id"`
	IssueID    uint64     `json:"issue_id"`
	AuthorID   uint64     `json:"author_id"`
	Body       string     `json:"body"`
	IsResolved bool       `json:"is_resolved"`
	ParentID   *uint64    `json:"parent_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// AddComment creates a comment on an issue (POST /comments).
func (c *Client) AddComment(ctx context.Context, issueID uint64, body string, parentID *uint64) (*Comment, error) {
	var out Comment
	req := map[string]any{"issue_id": issueID, "body": body}
	if parentID != nil {
		req["parent_id"] = *parentID
	}
	if _, err := c.PostJSON(ctx, "/comments", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListComments lists comments (GET /comments/issue/:id → {comments,total,...}).
func (c *Client) ListComments(ctx context.Context, issueID uint64, page, pageSize int) ([]Comment, int, error) {
	q := url.Values{}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		q.Set("page_size", strconv.Itoa(pageSize))
	}
	var out struct {
		Comments []Comment `json:"comments"`
		Total    int       `json:"total"`
	}
	if _, err := c.GetJSON(ctx, "/comments/issue/"+strconv.FormatUint(issueID, 10), q, &out); err != nil {
		return nil, 0, err
	}
	return out.Comments, out.Total, nil
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd sdk && go test ./client/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add sdk/client/issues.go sdk/client/issues_test.go
git commit -m "feat(sdk): add issues client with code resolution and comments"
```

### Task 8: cycles 客户端

**Files:**
- Create: `sdk/client/cycles.go`、`sdk/client/cycles_test.go`

**Interfaces:**
- Consumes: Task 5 核心
- Produces（Task 11/14 依赖）:
  - `func (c *Client) ListCycles(ctx, projectID uint64, status string, limit, offset int) (*CycleListResult, error)`
  - `func (c *Client) GetCycle(ctx, id uint64) (*Cycle, error)`
  - `func (c *Client) GetCycleProgress(ctx, id uint64) (*CycleProgress, error)`
  - `func (c *Client) GetCycleBurndown(ctx, id uint64) (*BurndownData, error)`
  - `func (c *Client) AddIssueToCycle(ctx, cycleID, issueID uint64) error`

- [ ] **Step 1: 写失败测试**

`sdk/client/cycles_test.go`：

```go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListCycles_WrappedShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/5/cycles" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items":  []map[string]any{{"id": 1, "name": "Sprint 1", "status": "active", "progress": 40.5}},
			"total":  1, "limit": 50, "offset": 0,
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	res, err := c.ListCycles(context.Background(), 5, "active", 50, 0)
	if err != nil || res.Total != 1 || len(res.Items) != 1 || res.Items[0].Progress != 40.5 {
		t.Fatalf("ListCycles: %v %+v", err, res)
	}
}

func TestGetCycleBurndown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cycles/3/burndown" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"cycle_id": 3, "cycle_name": "S1", "total_issues": 10, "is_on_track": true,
			"daily_points": []map[string]any{{"day_index": 0, "actual_remaining": 9.0}},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	b, err := c.GetCycleBurndown(context.Background(), 3)
	if err != nil || b.IsOnTrack != true || len(b.DailyPoints) != 1 {
		t.Fatalf("GetCycleBurndown: %v %+v", err, b)
	}
}

func TestAddIssueToCycle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cycles/3/issues" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("issue_id") != "11" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		json.NewEncoder(w).Encode(map[string]any{"cycle_id": 3, "issue_id": 11, "action": "added"})
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	if err := c.AddIssueToCycle(context.Background(), 3, 11); err != nil {
		t.Fatalf("AddIssueToCycle: %v", err)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd sdk && go test ./client/ -run "TestListCycles|TestGetCycleBurndown|TestAddIssueToCycle" -v`
Expected: FAIL

- [ ] **Step 3: 实现 cycles.go**

`sdk/client/cycles.go`：

```go
package client

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// Cycle is the response shape (list wrapped in {items,total,...}).
type Cycle struct {
	ID              uint64     `json:"id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description"`
	Status          string     `json:"status"` // computed: upcoming|active|completed|cancelled
	Progress        float64    `json:"progress"`
	TotalIssues     int64      `json:"total_issues"`
	CompletedIssues int64      `json:"completed_issues"`
	StartDate       string     `json:"start_date"`
	EndDate         *string    `json:"end_date"`
	ProjectID       uint64     `json:"project_id"`
	WorkspaceID     uint64     `json:"workspace_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CycleListResult is the wrapped list shape.
type CycleListResult struct {
	Items  []Cycle
	Total  int
	Limit  int
	Offset int
}

// ListCycles lists project cycles (GET /projects/:id/cycles).
func (c *Client) ListCycles(ctx context.Context, projectID uint64, status string, limit, offset int) (*CycleListResult, error) {
	q := url.Values{}
	if status != "" {
		q.Set("status", status)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	var out CycleListResult
	if _, err := c.GetJSON(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/cycles", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCycle fetches one cycle.
func (c *Client) GetCycle(ctx context.Context, id uint64) (*Cycle, error) {
	var out Cycle
	if _, err := c.GetJSON(ctx, "/cycles/"+strconv.FormatUint(id, 10), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// StateBreakdown is one row of cycle progress state stats.
type StateBreakdown struct {
	State string `json:"state"`
	Group string `json:"group"`
	Count int64  `json:"count"`
}

// CycleProgress is the GET /cycles/:id/progress response.
type CycleProgress struct {
	CycleID        uint64           `json:"cycle_id"`
	CycleName      string           `json:"cycle_name"`
	TotalIssues    int64            `json:"total_issues"`
	CompletedIssues int64           `json:"completed_issues"`
	Progress       float64          `json:"progress"`
	StateBreakdown []StateBreakdown `json:"state_breakdown"`
}

// GetCycleProgress fetches cycle progress.
func (c *Client) GetCycleProgress(ctx context.Context, id uint64) (*CycleProgress, error) {
	var out CycleProgress
	if _, err := c.GetJSON(ctx, "/cycles/"+strconv.FormatUint(id, 10)+"/progress", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BurndownDayPoint is one day of the burndown chart.
type BurndownDayPoint struct {
	DayIndex        int     `json:"day_index"`
	Date            string  `json:"date"`
	IdealRemaining  float64 `json:"ideal_remaining"`
	ActualCompleted int64   `json:"actual_completed"`
	ActualRemaining float64 `json:"actual_remaining"`
}

// BurndownData is the GET /cycles/:id/burndown response.
type BurndownData struct {
	CycleID        uint64            `json:"cycle_id"`
	CycleName      string            `json:"cycle_name"`
	StartDate      string            `json:"start_date"`
	EndDate        string            `json:"end_date"`
	TotalIssues    int64             `json:"total_issues"`
	TotalDays      int               `json:"total_days"`
	DaysElapsed    int               `json:"days_elapsed"`
	IdealDailyBurn float64           `json:"ideal_daily_burn"`
	IdealRemaining float64           `json:"ideal_remaining"`
	ActualCompleted int64            `json:"actual_completed"`
	ActualRemaining float64          `json:"actual_remaining"`
	IsOnTrack      bool              `json:"is_on_track"`
	DailyPoints    []BurndownDayPoint `json:"daily_points"`
}

// GetCycleBurndown fetches burndown data.
func (c *Client) GetCycleBurndown(ctx context.Context, id uint64) (*BurndownData, error) {
	var out BurndownData
	if _, err := c.GetJSON(ctx, "/cycles/"+strconv.FormatUint(id, 10)+"/burndown", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddIssueToCycle adds an issue to a cycle (POST /cycles/:id/issues?issue_id=).
func (c *Client) AddIssueToCycle(ctx context.Context, cycleID, issueID uint64) error {
	q := url.Values{}
	q.Set("issue_id", strconv.FormatUint(issueID, 10))
	var out map[string]any
	_, err := c.PostJSON(ctx, "/cycles/"+strconv.FormatUint(cycleID, 10)+"/issues", q, nil, &out)
	return err
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd sdk && go test ./client/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add sdk/client/cycles.go sdk/client/cycles_test.go
git commit -m "feat(sdk): add cycles client with progress/burndown"
```

### Task 9: AI / chat / agents 客户端（含 SSE 聚合）

**Files:**
- Create: `sdk/client/ai.go`、`sdk/client/chat.go`、`sdk/client/agents.go`
- Create: `sdk/client/ai_test.go`、`sdk/client/agents_test.go`

**Interfaces:**
- Consumes: Task 5 核心
- Produces（Task 12/14 依赖）:
  - `func (c *Client) postSSE(ctx context.Context, path string, query url.Values, body any, fn func(sseEvent) error) error`（包内共享）
  - `func (c *Client) AISearch(ctx, projectID uint64, query string) (*AISearchResponse, error)`
  - `func (c *Client) AIChat(ctx, projectID uint64, message string, threadID *uint64) (*AIChatReply, error)`
  - `func (c *Client) GetIssueChat(ctx, issueID uint64) (*Chat, error)`
  - `func (c *Client) SendMessage(ctx, chatID uint64, content string) (*Message, error)`
  - `func (c *Client) ListAgents(ctx, workspaceID uint64) ([]Agent, error)`
  - `func (c *Client) DispatchAgent(ctx, workspaceID, agentID uint64, task string, issueID, projectID *uint64) (*AgentActivity, error)`
  - `func (c *Client) GetAgentTask(ctx, workspaceID, taskID uint64) (*AgentTask, error)`
  - `func (c *Client) GetAgentTaskLogs(ctx, workspaceID, taskID uint64) ([]TaskLog, error)`

- [ ] **Step 1: 写失败测试（含 SSE 流）**

`sdk/client/ai_test.go`：

```go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAISearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/5/ai/search" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["query"] != "auth bugs" {
			t.Errorf("unexpected body %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"rql": `search "auth"`, "explanation": "matched", "issues": []map[string]any{{"id": 1}},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	res, err := c.AISearch(context.Background(), 5, "auth bugs")
	if err != nil || res.RQL == "" || len(res.Issues) != 1 {
		t.Fatalf("AISearch: %v %+v", err, res)
	}
}

func TestAIChat_AggregatesSSE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		// backend ai/chat SSE: data-only lines, no event: field
		chunks := []map[string]any{
			{"type": "text", "content": "Hel"},
			{"type": "text", "content": "lo"},
			{"type": "tool_call", "tool_call": map[string]any{"name": "get_issue", "arguments": `{"id":1}`}},
			{"type": "done", "thread_id": 99},
		}
		for _, ch := range chunks {
			b, _ := json.Marshal(ch)
			fmt.Fprintf(w, "data: %s\n\n", b)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	reply, err := c.AIChat(context.Background(), 5, "hi", nil)
	if err != nil {
		t.Fatalf("AIChat: %v", err)
	}
	if reply.Text != "Hello" {
		t.Fatalf("expected aggregated text %q", reply.Text)
	}
	if reply.ThreadID != 99 || len(reply.ToolCalls) != 1 {
		t.Fatalf("unexpected reply %+v", reply)
	}
}

func TestAIChat_StreamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {\"type\":\"error\",\"error\":\"llm down\"}\n\n")
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	_, err := c.AIChat(context.Background(), 5, "hi", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
```

`sdk/client/agents_test.go`（节选）：

```go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDispatchAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/2/agents/7/dispatch" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["task"] != "triage new bugs" {
			t.Errorf("unexpected body %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 1, "agent_id": 7, "action": "dispatch", "agent_name": "Triage"})
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	act, err := c.DispatchAgent(context.Background(), 2, 7, "triage new bugs", nil, nil)
	if err != nil || act.Action != "dispatch" || act.AgentName != "Triage" {
		t.Fatalf("DispatchAgent: %v %+v", err, act)
	}
}

func TestGetAgentTask(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/2/agent-tasks/15" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 15, "title": "T1", "status": "running", "progress": 50})
	}))
	defer srv.Close()

	c := New(srv.URL, "t")
	task, err := c.GetAgentTask(context.Background(), 2, 15)
	if err != nil || task.Status != "running" || task.Progress != 50 {
		t.Fatalf("GetAgentTask: %v %+v", err, task)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd sdk && go test ./client/ -run "TestAISearch|TestAIChat|TestDispatchAgent|TestGetAgentTask" -v`
Expected: FAIL

- [ ] **Step 3: 实现三个文件**

`sdk/client/ai.go`：

```go
package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// sseEvent is one parsed SSE frame. Backend ai/chat emits data-only lines.
type sseEvent struct {
	Event string
	Data  []byte
}

// postSSE POSTs and streams SSE frames to fn until the response ends or ctx
// is done. Handles both `data:`-only streams (ai/chat) and `event:`-tagged
// streams (chats/:id/stream).
func (c *Client) postSSE(ctx context.Context, path string, query url.Values, body any, fn func(sseEvent) error) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("request %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		apiErr := &APIError{StatusCode: resp.StatusCode, Body: map[string]any{}}
		var m map[string]any
		if json.Unmarshal(data, &m) == nil {
			apiErr.Body = m
			if msg, ok := m["message"].(string); ok {
				apiErr.Message = msg
			}
		}
		if apiErr.Message == "" {
			apiErr.Message = strings.TrimSpace(string(data))
		}
		return apiErr
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var cur sseEvent
	flush := func() error {
		if len(cur.Data) > 0 {
			if err := fn(cur); err != nil {
				return err
			}
			cur = sseEvent{}
		}
		return nil
	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "":
			if err := flush(); err != nil {
				return err
			}
		case strings.HasPrefix(line, "event:"):
			cur.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if cur.Data == nil {
				cur.Data = []byte(payload)
			} else {
				cur.Data = append(cur.Data, '\n')
				cur.Data = append(cur.Data, []byte(payload)...)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read SSE: %w", err)
	}
	return flush()
}

// AISearchRequest is the body for POST /projects/:id/ai/search.
type AISearchRequest struct {
	Query string `json:"query"`
}

// AISearchResponse carries the AI-generated RQL and matched issues.
type AISearchResponse struct {
	RQL         string                   `json:"rql"`
	Explanation string                   `json:"explanation"`
	Issues      []map[string]interface{} `json:"issues"`
}

// AISearch converts natural language into an issue search.
func (c *Client) AISearch(ctx context.Context, projectID uint64, query string) (*AISearchResponse, error) {
	var out AISearchResponse
	_, err := c.PostJSON(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/ai/search", nil,
		AISearchRequest{Query: query}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AIChatReply is the aggregated single-turn chat result.
type AIChatReply struct {
	Text      string
	ThreadID  uint64
	ToolCalls []string // "name(args)" summaries
}

// streamEvent is the backend llm.StreamEvent wire shape.
type streamEvent struct {
	Type       string `json:"type"` // text|tool_call|tool_result|thinking|done|error
	Content    string `json:"content"`
	ToolCall   *struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"tool_call"`
	ThreadID uint64 `json:"thread_id"`
	Error    string `json:"error"`
}

// AIChat sends one message to the project AI chat and aggregates the SSE
// stream into a single reply. Long-running: uses a 5-minute context timeout.
func (c *Client) AIChat(ctx context.Context, projectID uint64, message string, threadID *uint64) (*AIChatReply, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	body := map[string]any{"message": message}
	if threadID != nil {
		body["thread_id"] = *threadID
	}

	reply := &AIChatReply{}
	err := c.postSSE(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/ai/chat", nil, body,
		func(ev sseEvent) error {
			var se streamEvent
			if err := json.Unmarshal(ev.Data, &se); err != nil {
				return nil // ignore non-JSON keepalives
			}
			switch se.Type {
			case "text", "thinking":
				reply.Text += se.Content
			case "tool_call":
				if se.ToolCall != nil {
					reply.ToolCalls = append(reply.ToolCalls,
						se.ToolCall.Name+"("+se.ToolCall.Arguments+")")
				}
			case "error":
				return &APIError{StatusCode: 502, Message: se.Error, Body: map[string]any{"error": se.Error}}
			}
			if se.ThreadID != 0 {
				reply.ThreadID = se.ThreadID
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return reply, nil
}
```

`sdk/client/chat.go`：

```go
package client

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// Message is a chat message.
type Message struct {
	ID         uint64     `json:"id"`
	ChatID     uint64     `json:"chat_id"`
	SenderID   uint64     `json:"sender_id"`
	SenderType string     `json:"sender_type"` // "user" | "agent"
	Content    string     `json:"content"`
	ReplyToID  *uint64    `json:"reply_to_id"`
	EditedAt   *time.Time `json:"edited_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Chat is the GET /issues/:id/chat response.
type Chat struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspace_id"`
	ProjectID   *uint64   `json:"project_id"`
	IssueID     *uint64   `json:"issue_id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"created_at"`
	Messages    []Message `json:"messages"`
}

// GetIssueChat returns (or creates) the chat for an issue.
func (c *Client) GetIssueChat(ctx context.Context, issueID uint64) (*Chat, error) {
	var out Chat
	if _, err := c.GetJSON(ctx, "/issues/"+strconv.FormatUint(issueID, 10)+"/chat", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendMessage posts a user message to a chat (POST /chats/:id/messages).
func (c *Client) SendMessage(ctx context.Context, chatID uint64, content string) (*Message, error) {
	q := url.Values{}
	q.Set("workspace_id", "") // placeholder to keep url import; remove if unused
	_ = q
	var out Message
	_, err := c.PostJSON(ctx, "/chats/"+strconv.FormatUint(chatID, 10)+"/messages", nil,
		map[string]string{"content": content}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
```

`sdk/client/agents.go`：

```go
package client

import (
	"context"
	"encoding/json"
	"strconv"
	"time"
)

// Agent is a workspace AI agent.
type Agent struct {
	ID            uint64     `json:"id"`
	WorkspaceID   uint64     `json:"workspace_id"`
	Name          string     `json:"name"`
	Avatar        string     `json:"avatar"`
	AgentType     string     `json:"agent_type"` // "builtin" | "custom"
	Capabilities  []string   `json:"capabilities"`
	Status        string     `json:"status"` // "active" | "inactive"
	ModelOverride *string    `json:"model_override"`
	SystemPrompt  *string    `json:"system_prompt"`
	TemplateID    *uint64    `json:"template_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ListAgents returns workspace agents (GET /workspaces/:ws/agents).
func (c *Client) ListAgents(ctx context.Context, workspaceID uint64) ([]Agent, error) {
	var out []Agent
	_, err := c.GetJSON(ctx, "/workspaces/"+strconv.FormatUint(workspaceID, 10)+"/agents", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AgentActivity is the dispatch response.
type AgentActivity struct {
	ID           uint64     `json:"id"`
	AgentID      uint64     `json:"agent_id"`
	IssueID      *uint64    `json:"issue_id"`
	Action       string     `json:"action"` // "dispatch" | "auto_triage" | ...
	ResultSummary string    `json:"result_summary"`
	Rating       *int       `json:"rating"`
	ExecutedAt   *time.Time `json:"executed_at"`
	AgentName    string     `json:"agent_name"`
	TaskContext  string     `json:"task_context"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// DispatchAgent triggers an agent (POST /workspaces/:ws/agents/:id/dispatch).
// Long-running: uses a 5-minute context timeout.
func (c *Client) DispatchAgent(ctx context.Context, workspaceID, agentID uint64, task string, issueID, projectID *uint64) (*AgentActivity, error) {
	body := map[string]any{"task": task}
	if issueID != nil {
		body["issue_id"] = *issueID
	}
	if projectID != nil {
		body["project_id"] = *projectID
	}
	var out AgentActivity
	_, err := c.PostJSON(ctx,
		"/workspaces/"+strconv.FormatUint(workspaceID, 10)+"/agents/"+strconv.FormatUint(agentID, 10)+"/dispatch",
		nil, body, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AgentTask is the GET /workspaces/:ws/agent-tasks/:id response.
type AgentTask struct {
	ID            uint64          `json:"id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	Status        string          `json:"status"` // enqueue|claimed|running|completed|failed|cancelled
	Priority      int             `json:"priority"`
	Progress      int             `json:"progress"`
	TaskType      string          `json:"task_type"`
	OutputData    json.RawMessage `json:"output_data"`
	ErrorInfo     string          `json:"error_info"`
	FailureReason string          `json:"failure_reason"`
	WorkspaceID   uint64          `json:"workspace_id"`
	ProjectID     *uint64         `json:"project_id"`
	IssueID       *uint64         `json:"issue_id"`
	EnqueuedAt    *time.Time      `json:"enqueued_at"`
	StartedAt     *time.Time      `json:"started_at"`
	CompletedAt   *time.Time      `json:"completed_at"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// GetAgentTask fetches an agent task by ID.
func (c *Client) GetAgentTask(ctx context.Context, workspaceID, taskID uint64) (*AgentTask, error) {
	var out AgentTask
	path := "/workspaces/" + strconv.FormatUint(workspaceID, 10) + "/agent-tasks/" + strconv.FormatUint(taskID, 10)
	if _, err := c.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TaskLog is one agent task log entry.
type TaskLog struct {
	ID        uint64    `json:"id"`
	TaskID    uint64    `json:"task_id"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Metadata  string    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

// GetAgentTaskLogs fetches task logs.
func (c *Client) GetAgentTaskLogs(ctx context.Context, workspaceID, taskID uint64) ([]TaskLog, error) {
	var out []TaskLog
	path := "/workspaces/" + strconv.FormatUint(workspaceID, 10) + "/agent-tasks/" + strconv.FormatUint(taskID, 10) + "/logs"
	if _, err := c.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd sdk && go test ./client/ -v && go vet ./...`
Expected: PASS（若 `chat.go` 中 `url` import 未被使用导致编译失败，删除 `SendMessage` 里的 placeholder 两行及 `net/url` import）

- [ ] **Step 5: Commit**

```bash
git add sdk/client/ai.go sdk/client/chat.go sdk/client/agents.go sdk/client/ai_test.go sdk/client/agents_test.go
git commit -m "feat(sdk): add AI/chat/agents client with SSE aggregation"
```

---

## Phase C：MCP Server

> mcp-go v1.0.0 已核实 API（module cache 源码逐条核对）：
> - `server.NewMCPServer(name, version string, opts ...ServerOption)`；`WithInputSchemaValidation()` 为 ServerOption
> - `AddTool(tool mcp.Tool, handler ToolHandlerFunc)`，`ToolHandlerFunc func(ctx, mcp.CallToolRequest) (*mcp.CallToolResult, error)`
> - `mcp.NewTool(name, opts ...ToolOption)`；`mcp.WithDescription` / `mcp.WithString` / `mcp.WithInteger` / `mcp.WithBoolean` / `mcp.Required()` / `mcp.Description`
> - `mcp.NewToolResultText(string)` / `mcp.NewToolResultError(string)` / `mcp.NewToolResultErrorf(format, ...)`；`CallToolResult{Content, IsError}`
> - 测试：`servertest.NewTestStreamableHTTPServer(srv, server.WithStateful(true))`（import `github.com/mark3labs/mcp-go/server/servertest`）→ POST initialize（protocolVersion `mcp.LATEST_PROTOCOL_VERSION` = "2026-07-28"）→ 从响应头 `Mcp-Session-Id`（常量 `server.HeaderKeySessionID`）取会话 ID → 后续请求带该头。响应体 `{"id":N,"result":{...}}`。
> - `server.NewStreamableHTTPServer(srv, opts ...StreamableHTTPOption)`：`WithEndpointPath`、`WithStateful`、`WithDisableLocalhostProtection` 均存在
> - stdio：`server.ServeStdio(srv) error`

### Task 10: MCP server 骨架（server 组装 + 双传输入口）

**Files:**
- Create: `sdk/mcp/server.go`
- Create: `sdk/mcp/http.go`（HTTP 传输的 Bearer 认证中间件，spec §5.1）
- Create: `sdk/mcp/server_test.go`（含 Task 11/12 共用的测试辅助函数）
- Create: `sdk/cmd/reqmango-mcp/main.go`

**Interfaces:**
- Consumes: `client.New`（Task 5）
- Produces（Task 11/12/15 依赖）:
  - `const ServerName = "reqmango-mcp"`、`const ServerVersion = "1.0.0"`
  - `func New(cli *client.Client) *server.MCPServer`（Task 11/12 各加一行注册调用）
  - `func BearerAuth(pat string, next http.Handler) http.Handler`（HTTP 模式认证，Task 15 e2e 验证）
  - 测试辅助（同包 `_test.go` 内，Task 11/12 复用）：`mcpPost(t, baseURL, sessionID string, msg map[string]any) (map[string]any, string)`、`newTestMCPServer(t *testing.T, cli *client.Client) (*httptest.Server, string)`、`callTool(t, baseURL, sessionID, name string, args map[string]any) ([]string, bool)`

- [ ] **Step 1: 引入依赖**

Run: `cd sdk && go get github.com/mark3labs/mcp-go@v1.0.0`
Expected: go.mod/go.sum 更新

- [ ] **Step 2: 写失败测试（初始化握手 + 空工具集）**

`sdk/mcp/server_test.go`：

```go
package mcp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mark3labs/mcp-go/server/servertest"

	"github.com/reqmango/tools/client"
)

// mcpPost sends one JSON-RPC message over stateful streamable HTTP and
// returns the decoded response body plus the session id ("" on first call).
func mcpPost(t *testing.T, baseURL, sessionID string, msg map[string]any) (map[string]any, string) {
	t.Helper()
	body, _ := json.Marshal(msg)
	req, _ := http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set(server.HeaderKeySessionID, sessionID)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST status %d: %s", resp.StatusCode, data)
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return out, resp.Header.Get(server.HeaderKeySessionID)
}

// newTestMCPServer starts a stateful streamable HTTP test server, runs the
// initialize handshake and returns the base URL + session id.
func newTestMCPServer(t *testing.T, cli *client.Client) (*httptest.Server, string) {
	t.Helper()
	httpSrv := servertest.NewTestStreamableHTTPServer(New(cli), server.WithStateful(true))
	t.Cleanup(httpSrv.Close)
	_, sessionID := mcpPost(t, httpSrv.URL, "", map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": mcp.LATEST_PROTOCOL_VERSION,
			"clientInfo":      map[string]any{"name": "reqmango-test", "version": "1.0.0"},
			"capabilities":    map[string]any{},
		},
	})
	if sessionID == "" {
		t.Fatal("initialize did not return a session id")
	}
	return httpSrv, sessionID
}

// callTool invokes tools/call and returns (text contents, isError).
func callTool(t *testing.T, baseURL, sessionID, name string, args map[string]any) ([]string, bool) {
	t.Helper()
	resp, _ := mcpPost(t, baseURL, sessionID, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params":  map[string]any{"name": name, "arguments": args},
	})
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("tools/call %s: no result in %v", name, resp)
	}
	isError, _ := result["isError"].(bool)
	var texts []string
	if content, ok := result["content"].([]any); ok {
		for _, c := range content {
			if cm, ok := c.(map[string]any); ok {
				if s, ok := cm["text"].(string); ok {
					texts = append(texts, s)
				}
			}
		}
	}
	return texts, isError
}

func listToolNames(t *testing.T, baseURL, sessionID string) []string {
	t.Helper()
	resp, _ := mcpPost(t, baseURL, sessionID, map[string]any{
		"jsonrpc": "2.0", "id": 3, "method": "tools/list", "params": map[string]any{},
	})
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("tools/list: no result in %v", resp)
	}
	var names []string
	for _, tool := range result["tools"].([]any) {
		names = append(names, tool.(map[string]any)["name"].(string))
	}
	return names
}

func TestNew_Initialize(t *testing.T) {
	httpSrv, sessionID := newTestMCPServer(t, client.New("", "reqmango_pat_test"))
	_ = httpSrv

	resp, _ := mcpPost(t, httpSrv.URL, sessionID, map[string]any{
		"jsonrpc": "2.0", "id": 4, "method": "ping", "params": map[string]any{},
	})
	if _, ok := resp["result"]; !ok {
		t.Fatalf("ping failed: %v", resp)
	}
}

func TestNew_ToolsListInitiallyEmpty(t *testing.T) {
	httpSrv, sessionID := newTestMCPServer(t, client.New("", "reqmango_pat_test"))
	names := listToolNames(t, httpSrv.URL, sessionID)
	if len(names) != 0 {
		t.Fatalf("expected 0 tools before Tasks 11/12, got %v", names)
	}
}

func TestBearerAuth(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	h := BearerAuth("reqmango_pat_secret", inner)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/mcp", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without header, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer reqmango_pat_wrong")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong token, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer reqmango_pat_secret")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 with correct token, got %d", rr.Code)
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `cd sdk && go test ./mcp/ -v`
Expected: FAIL（mcp.go 不存在）

- [ ] **Step 4: 实现 server.go 与 main.go**

`sdk/mcp/server.go`：

```go
// Package mcp assembles the reqmango MCP server on top of the shared client.
package mcp

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/reqmango/tools/client"
)

// ServerName and ServerVersion identify this MCP server to clients.
const (
	ServerName    = "reqmango-mcp"
	ServerVersion = "1.0.0"
)

// New creates the reqmango MCP server backed by cli. Tool registration is
// split across registerCoreTools (19 core) and registerAITools (5 AI).
func New(cli *client.Client) *server.MCPServer {
	s := server.NewMCPServer(ServerName, ServerVersion, server.WithInputSchemaValidation())
	return s
}
```

`sdk/mcp/http.go`：

```go
package mcp

import (
	"crypto/subtle"
	"net/http"
)

// BearerAuth wraps next and rejects requests whose Authorization header is
// not "Bearer <pat>" (spec §5.1: HTTP transport is protected by the PAT).
// Constant-time comparison avoids leaking prefix matches.
func BearerAuth(pat string, next http.Handler) http.Handler {
	want := "Bearer " + pat
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte(want)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"unauthorized: missing or invalid Bearer token"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

`sdk/cmd/reqmango-mcp/main.go`：

```go
// Command reqmango-mcp is the reqmango MCP server.
//
// Stdio transport (default, for Claude Code / Claude Desktop / Cursor):
//
//	REQMANGO_PAT=reqmango_pat_xxx reqmango-mcp
//
// Streamable HTTP transport (for remote / CI):
//
//	REQMANGO_PAT=reqmango_pat_xxx reqmango-mcp --http :8080
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/reqmango/tools/client"
	reqmcp "github.com/reqmango/tools/mcp"
)

func main() {
	httpAddr := flag.String("http", "", "serve streamable HTTP on this address (e.g. :8080) instead of stdio")
	flag.Parse()

	apiURL := os.Getenv("REQMANGO_API_URL")
	if apiURL == "" {
		apiURL = client.DefaultBaseURL
	}
	pat := os.Getenv("REQMANGO_PAT")
	if pat == "" {
		fmt.Fprintln(os.Stderr, "ERROR: REQMANGO_PAT environment variable is required (create one with `reqmango auth login`)")
		os.Exit(1)
	}

	s := reqmcp.New(client.New(apiURL, pat))

	if *httpAddr != "" {
		h := server.NewStreamableHTTPServer(s,
			server.WithEndpointPath("/mcp"),
			server.WithStateful(true),
			server.WithDisableLocalhostProtection(true), // --http is an explicit opt-in for remote serving
		)
		log.Printf("reqmango-mcp listening on %s (endpoint /mcp, Bearer PAT required)", *httpAddr)
		log.Fatal(http.ListenAndServe(*httpAddr, reqmcp.BearerAuth(pat, h)))
	}

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 5: 运行测试确认通过**

Run: `cd sdk && go test ./mcp/ -v && go build ./cmd/reqmango-mcp`
Expected: PASS（注意 mcp 包名与 mcp-go 的 mcp 包名冲突时用 `reqmcp` 别名，测试文件里 import 别名同上）

- [ ] **Step 6: Commit**

```bash
git add sdk/go.mod sdk/go.sum sdk/mcp/server.go sdk/mcp/http.go sdk/mcp/server_test.go sdk/cmd/reqmango-mcp/main.go
git commit -m "feat(mcp): add MCP server scaffold with stdio and streamable HTTP"
```

### Task 11: MCP 核心工具（19 个）

**Files:**
- Create: `sdk/mcp/tools_core.go`
- Create: `sdk/mcp/tools_core_test.go`
- Modify: `sdk/mcp/server.go`（New 中加一行 `registerCoreTools(s, cli)`）

**Interfaces:**
- Consumes: Task 10 的 `New` + 测试辅助；Task 5-8 的 client 方法（`ListWorkspaces`、`ListProjects`、`GetProject`、`CreateIssue`、`ListIssues`、`GetIssue`、`UpdateIssue`、`SearchIssues`、`ResolveIssueCode`、`AddComment`、`ListComments`、`ListCycles`、`GetCycleProgress`、`AddIssueToCycle`、`ListMembers`、`ListStates`、`ListLabels`、`ListIssueTypes`、`ListNotifications`、`ListPages`、`GetPage`）
- Produces: `func registerCoreTools(s *server.MCPServer, cli *client.Client)`、`func decodeArgs(req mcp.CallToolRequest, out any) error`、`func toolResultJSON(v any) *mcp.CallToolResult`、`func toolAPIError(err error) *mcp.CallToolResult`、`func resolveIssueArg(ctx, cli, workspaceID int64, arg string) (uint64, error)`（Task 12 复用前三个辅助）

- [ ] **Step 1: 写失败测试**

`sdk/mcp/tools_core_test.go`：

```go
package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/reqmango/tools/client"
)

// backend returns an httptest server with the routes the core tools exercise.
func backend(t *testing.T, routes map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := routes[r.URL.Path]
		if !ok {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestCoreTools_ToolCountAndNames(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL, "reqmango_pat_test"))

	names := listToolNames(t, httpSrv.URL, sessionID)
	want := []string{
		"list_workspaces", "list_projects", "get_project",
		"create_issue", "list_issues", "get_issue", "update_issue", "search_issues",
		"add_comment", "list_cycles", "get_cycle_progress", "add_issue_to_cycle",
		"list_members", "get_states", "get_labels", "list_issue_types",
		"list_notifications", "list_pages", "get_page",
	}
	// >=（不是 ==）：Task 12 还会注册 5 个 AI 工具，总数变 24。
	if len(names) < len(want) {
		t.Fatalf("expected at least %d core tools, got %d: %v", len(want), len(names), names)
	}
	for _, w := range want {
		found := false
		for _, n := range names {
			if n == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing tool %s", w)
		}
	}
}

func TestListWorkspaces(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/workspaces": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode([]map[string]any{{"id": 2, "name": "Acme", "slug": "acme"}})
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL, "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "list_workspaces", map[string]any{})
	if isError || len(texts) != 1 || !strings.Contains(texts[0], `"slug": "acme"`) {
		t.Fatalf("list_workspaces: texts=%v isError=%v", texts, isError)
	}
}

func TestCreateIssue(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/issues": func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("project_id") != "5" || q.Get("workspace_id") != "2" {
				t.Errorf("unexpected query %s", r.URL.RawQuery)
			}
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["name"] != "Login broken" || body["priority"] != "high" {
				t.Errorf("unexpected body %v", body)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"id": 11, "name": "Login broken", "sequence_id": 42, "priority": "high"})
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL, "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "create_issue", map[string]any{
		"project_id": 5, "workspace_id": 2, "name": "Login broken", "priority": "high",
	})
	if isError || len(texts) != 1 || !strings.Contains(texts[0], `"sequence_id": 42`) {
		t.Fatalf("create_issue: texts=%v isError=%v", texts, isError)
	}
}

func TestGetIssue_WithCode(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/projects": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode([]map[string]any{{"id": 5, "identifier": "DEMO", "workspace_id": 2}})
		},
		"/api/v1/issues": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Total-Count", "1")
			json.NewEncoder(w).Encode([]map[string]any{{"id": 11, "sequence_id": 42, "name": "x"}})
		},
		"/api/v1/issues/11": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"id": 11, "name": "Login broken", "sequence_id": 42})
		},
		"/api/v1/comments/issue/11": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"comments": []map[string]any{{"id": 1, "body": "repro"}}, "total": 1})
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL, "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "get_issue", map[string]any{
		"issue": "DEMO-42", "workspace_id": 2,
	})
	if isError || len(texts) != 1 {
		t.Fatalf("get_issue: texts=%v isError=%v", texts, isError)
	}
	if !strings.Contains(texts[0], `"body": "repro"`) || !strings.Contains(texts[0], `"id": 11`) {
		t.Fatalf("get_issue should merge comments, got %s", texts[0])
	}
}

func TestToolError_401(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/workspaces": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"message": "token expired"})
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL, "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "list_workspaces", map[string]any{})
	if !isError || len(texts) != 1 || !strings.Contains(texts[0], "reqmango auth login") {
		t.Fatalf("expected 401 hint, got texts=%v isError=%v", texts, isError)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd sdk && go test ./mcp/ -run "TestCoreTools|TestListWorkspaces|TestCreateIssue|TestGetIssue_WithCode|TestToolError" -v`
Expected: FAIL（`registerCoreTools` 未调用 / 未定义；`TestNew_ToolsListInitiallyEmpty` 在 Step 4 后会转失败，届时按 Step 4 说明同步更新该断言）

- [ ] **Step 3: 实现 tools_core.go（19 个工具）**

`sdk/mcp/tools_core.go`：

```go
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/reqmango/tools/client"
)

// decodeArgs converts CallToolRequest.Arguments into the target struct.
func decodeArgs(req mcp.CallToolRequest, out any) error {
	raw, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}
	return json.Unmarshal(raw, out)
}

// toolResultJSON renders v as indented JSON text content.
func toolResultJSON(v any) *mcp.CallToolResult {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultErrorf("failed to serialize result: %v", err)
	}
	return mcp.NewToolResultText(string(data))
}

// toolAPIError maps an error to an isError tool result. 401s get a re-login
// hint so the model can tell the user what to fix.
func toolAPIError(err error) *mcp.CallToolResult {
	if apiErr := client.AsAPIError(err); apiErr != nil && apiErr.StatusCode == 401 {
		return mcp.NewToolResultError("authentication failed (401): the configured PAT is invalid or revoked. Create a new one with `reqmango auth login` and restart with the new REQMANGO_PAT.")
	}
	return mcp.NewToolResultError(err.Error())
}

// parseIDList splits a comma-separated ID list into uint64s.
func parseIDList(s string) []uint64 {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []uint64
	for _, part := range strings.Split(s, ",") {
		if n, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64); err == nil && n != 0 {
			out = append(out, n)
		}
	}
	return out
}

// resolveIssueArg accepts "123" or "DEMO-42" and returns the numeric issue ID.
func resolveIssueArg(ctx context.Context, cli *client.Client, workspaceID int64, arg string) (uint64, error) {
	if n, err := strconv.ParseUint(arg, 10, 64); err == nil {
		return n, nil
	}
	if workspaceID == 0 {
		return 0, fmt.Errorf("issue code %q needs workspace_id to resolve the project identifier", arg)
	}
	return cli.ResolveIssueCode(ctx, uint64(workspaceID), arg)
}

type listProjectsArgs struct {
	WorkspaceID int64 `json:"workspace_id"`
}

type getProjectArgs struct {
	ProjectID int64 `json:"project_id"`
}

type createIssueArgs struct {
	WorkspaceID     int64  `json:"workspace_id"`
	ProjectID       int64  `json:"project_id"`
	Name            string `json:"name"`
	DescriptionHTML string `json:"description_html"`
	Priority        string `json:"priority"`
	StateID         int64  `json:"state_id"`
	AssigneeIDs     string `json:"assignee_ids"` // comma-separated user ids
	LabelIDs        string `json:"label_ids"`    // comma-separated label ids
	ParentID        int64  `json:"parent_id"`
	TypeID          int64  `json:"type_id"`
	CycleID         int64  `json:"cycle_id"`
	TargetDate      string `json:"target_date"`
}

type listIssuesArgs struct {
	WorkspaceID int64  `json:"workspace_id"`
	ProjectID   int64  `json:"project_id"`
	RQL         string `json:"rql"`
	StateID     int64  `json:"state_id"`
	Priority    string `json:"priority"`
	AssigneeID  int64  `json:"assignee_id"`
	CycleID     int64  `json:"cycle_id"`
	IssueTypeID int64  `json:"issue_type_id"`
	Search      string `json:"search"`
	SortBy      string `json:"sort_by"`
	SortDir     string `json:"sort_dir"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

type getUpdateIssueArgs struct {
	Issue       string `json:"issue"` // numeric ID or "DEMO-42" (code needs workspace_id)
	WorkspaceID int64  `json:"workspace_id"`
}

type updateIssueArgs struct {
	Issue       string `json:"issue"` // numeric ID or "DEMO-42"
	WorkspaceID int64  `json:"workspace_id"`
	Name        string `json:"name"`
	Priority    string `json:"priority"`
	StateID     int64  `json:"state_id"`
	AssigneeIDs string `json:"assignee_ids"` // replace-all
	LabelIDs    string `json:"label_ids"`    // replace-all
	CycleID     int64  `json:"cycle_id"`
	TargetDate  string `json:"target_date"`
}

type searchIssuesArgs struct {
	WorkspaceID int64  `json:"workspace_id"`
	Query       string `json:"query"`
	ProjectID   int64  `json:"project_id"`
	Limit       int    `json:"limit"`
}

type addCommentArgs struct {
	IssueID int64  `json:"issue_id"`
	Body    string `json:"body"`
}

type listCyclesArgs struct {
	ProjectID int64  `json:"project_id"`
	Status    string `json:"status"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

type cycleIDArgs struct {
	CycleID int64 `json:"cycle_id"`
}

type addIssueToCycleArgs struct {
	CycleID int64 `json:"cycle_id"`
	IssueID int64 `json:"issue_id"`
}

type workspaceIDArgs struct {
	WorkspaceID int64 `json:"workspace_id"`
}

type issueTypesArgs struct {
	WorkspaceID int64 `json:"workspace_id"`
	ProjectID   int64 `json:"project_id"`
}

type notificationsArgs struct {
	UnreadOnly bool `json:"unread_only"`
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
}

type getPageArgs struct {
	ProjectID int64 `json:"project_id"`
	PageID    int64 `json:"page_id"`
}

// registerCoreTools adds the 19 core CRUD/meta tools.
func registerCoreTools(s *server.MCPServer, cli *client.Client) {
	s.AddTool(mcp.NewTool("list_workspaces",
		mcp.WithDescription("List workspaces the current user can access. Call this first to discover workspace IDs.")),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			out, err := cli.ListWorkspaces(ctx)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_projects",
		mcp.WithDescription("List projects in a workspace"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID from list_workspaces"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a listProjectsArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListProjects(ctx, uint64(a.WorkspaceID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_project",
		mcp.WithDescription("Get one project by numeric ID"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getProjectArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.GetProject(ctx, uint64(a.ProjectID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("create_issue",
		mcp.WithDescription("Create an issue in a project"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID")),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Issue title (1-255 chars)")),
		mcp.WithString("description_html", mcp.Description("Issue description (HTML)")),
		mcp.WithString("priority", mcp.Description("Priority: none|low|medium|high|urgent")),
		mcp.WithInteger("state_id", mcp.Description("Initial workflow state ID (from get_states); default state used if omitted")),
		mcp.WithString("assignee_ids", mcp.Description("Comma-separated user IDs to assign")),
		mcp.WithString("label_ids", mcp.Description("Comma-separated label IDs (from get_labels)")),
		mcp.WithInteger("parent_id", mcp.Description("Parent issue ID")),
		mcp.WithInteger("type_id", mcp.Description("Issue type ID (from list_issue_types)")),
		mcp.WithInteger("cycle_id", mcp.Description("Cycle ID to add the issue to")),
		mcp.WithString("target_date", mcp.Description("Target date (ISO 8601)"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a createIssueArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := &client.CreateIssueRequest{
				Name:            a.Name,
				DescriptionHTML: a.DescriptionHTML,
				Priority:        a.Priority,
				AssigneeIDs:     parseIDList(a.AssigneeIDs),
				LabelIDs:        parseIDList(a.LabelIDs),
			}
			if a.StateID != 0 {
				body.StateID = uintPtr64(uint64(a.StateID))
			}
			if a.ParentID != 0 {
				body.ParentID = uintPtr64(uint64(a.ParentID))
			}
			if a.TypeID != 0 {
				body.TypeID = uintPtr64(uint64(a.TypeID))
			}
			if a.CycleID != 0 {
				body.CycleID = uintPtr64(uint64(a.CycleID))
			}
			if a.TargetDate != "" {
				body.TargetDate = &a.TargetDate
			}
			out, err := cli.CreateIssue(ctx, uint64(a.ProjectID), uint64(a.WorkspaceID), body)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_issues",
		mcp.WithDescription("List issues with optional RQL query and filters. Returns items and total count."),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID")),
		mcp.WithInteger("workspace_id", mcp.Description("Workspace ID")),
		mcp.WithString("rql", mcp.Description("RQL query, e.g. `priority = \"high\" and state_group != \"completed\"`")),
		mcp.WithInteger("state_id", mcp.Description("Filter by workflow state ID")),
		mcp.WithString("priority", mcp.Description("Filter by priority")),
		mcp.WithInteger("assignee_id", mcp.Description("Filter by assignee user ID")),
		mcp.WithInteger("cycle_id", mcp.Description("Filter by cycle ID")),
		mcp.WithInteger("issue_type_id", mcp.Description("Filter by issue type ID")),
		mcp.WithString("search", mcp.Description("Free-text search")),
		mcp.WithString("sort_by", mcp.Description("Sort field")),
		mcp.WithString("sort_dir", mcp.Description("Sort direction: asc|desc")),
		mcp.WithInteger("limit", mcp.Description("Page size (default 50)")),
		mcp.WithInteger("offset", mcp.Description("Offset for pagination"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a listIssuesArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			res, err := cli.ListIssues(ctx, client.IssueListOptions{
				ProjectID:   uint64(a.ProjectID),
				WorkspaceID: uint64(a.WorkspaceID),
				RQL:         a.RQL,
				StateID:     uint64(a.StateID),
				Priority:    a.Priority,
				AssigneeID:  uint64(a.AssigneeID),
				CycleID:     uint64(a.CycleID),
				IssueTypeID: uint64(a.IssueTypeID),
				Search:      a.Search,
				SortBy:      a.SortBy,
				SortDir:     a.SortDir,
				Limit:       a.Limit,
				Offset:      a.Offset,
			})
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(map[string]any{"total": res.Total, "items": res.Items}), nil
		})

	s.AddTool(mcp.NewTool("get_issue",
		mcp.WithDescription("Get one issue by numeric ID or code like DEMO-42 (code needs workspace_id), with its comments"),
		mcp.WithString("issue", mcp.Required(), mcp.Description("Numeric issue ID or code like DEMO-42")),
		mcp.WithInteger("workspace_id", mcp.Description("Workspace ID — required when `issue` is a code"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getUpdateIssueArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			id, err := resolveIssueArg(ctx, cli, a.WorkspaceID, a.Issue)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			iss, err := cli.GetIssue(ctx, id)
			if err != nil {
				return toolAPIError(err), nil
			}
			comments, total, err := cli.ListComments(ctx, id, 1, 100)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(map[string]any{"issue": iss, "comments": comments, "comments_total": total}), nil
		})

	s.AddTool(mcp.NewTool("update_issue",
		mcp.WithDescription("Update an issue. State transitions go through the backend workflow validation and may require approval (409 with transition details)."),
		mcp.WithString("issue", mcp.Required(), mcp.Description("Numeric issue ID or code like DEMO-42")),
		mcp.WithInteger("workspace_id", mcp.Description("Workspace ID — required when `issue` is a code")),
		mcp.WithString("name", mcp.Description("New title")),
		mcp.WithString("priority", mcp.Description("New priority: none|low|medium|high|urgent")),
		mcp.WithInteger("state_id", mcp.Description("Target state ID (from get_states)")),
		mcp.WithString("assignee_ids", mcp.Description("Comma-separated user IDs — replaces all assignees")),
		mcp.WithString("label_ids", mcp.Description("Comma-separated label IDs — replaces all labels")),
		mcp.WithInteger("cycle_id", mcp.Description("Move to cycle ID")),
		mcp.WithString("target_date", mcp.Description("New target date (ISO 8601)"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a updateIssueArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			id, err := resolveIssueArg(ctx, cli, a.WorkspaceID, a.Issue)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := &client.UpdateIssueRequest{}
			if a.Name != "" {
				body.Name = &a.Name
			}
			if a.Priority != "" {
				body.Priority = &a.Priority
			}
			if a.StateID != 0 {
				body.StateID = uintPtr64(uint64(a.StateID))
			}
			if a.AssigneeIDs != "" {
				body.AssigneeIDs = parseIDList(a.AssigneeIDs)
			}
			if a.LabelIDs != "" {
				body.LabelIDs = parseIDList(a.LabelIDs)
			}
			if a.CycleID != 0 {
				body.CycleID = uintPtr64(uint64(a.CycleID))
			}
			if a.TargetDate != "" {
				body.TargetDate = &a.TargetDate
			}
			out, err := cli.UpdateIssue(ctx, id, body)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("search_issues",
		mcp.WithDescription("Full-text search across issues in a workspace"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query text")),
		mcp.WithInteger("project_id", mcp.Description("Limit search to one project")),
		mcp.WithInteger("limit", mcp.Description("Max results (default 10)"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a searchIssuesArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var pid *uint64
			if a.ProjectID != 0 {
				pid = uintPtr64(uint64(a.ProjectID))
			}
			out, err := cli.SearchIssues(ctx, uint64(a.WorkspaceID), a.Query, pid, a.Limit)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("add_comment",
		mcp.WithDescription("Add a comment to an issue"),
		mcp.WithInteger("issue_id", mcp.Required(), mcp.Description("Issue ID")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Comment text"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a addCommentArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.AddComment(ctx, uint64(a.IssueID), a.Body, nil)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_cycles",
		mcp.WithDescription("List cycles (sprints) of a project"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID")),
		mcp.WithString("status", mcp.Description("Filter: upcoming|active|completed|cancelled")),
		mcp.WithInteger("limit", mcp.Description("Page size")),
		mcp.WithInteger("offset", mcp.Description("Offset"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a listCyclesArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListCycles(ctx, uint64(a.ProjectID), a.Status, a.Limit, a.Offset)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_cycle_progress",
		mcp.WithDescription("Get a cycle's progress breakdown by state"),
		mcp.WithInteger("cycle_id", mcp.Required(), mcp.Description("Cycle ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a cycleIDArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.GetCycleProgress(ctx, uint64(a.CycleID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("add_issue_to_cycle",
		mcp.WithDescription("Add an existing issue to a cycle"),
		mcp.WithInteger("cycle_id", mcp.Required(), mcp.Description("Cycle ID")),
		mcp.WithInteger("issue_id", mcp.Required(), mcp.Description("Issue ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a addIssueToCycleArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if err := cli.AddIssueToCycle(ctx, uint64(a.CycleID), uint64(a.IssueID)); err != nil {
				return toolAPIError(err), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("added issue %d to cycle %d", a.IssueID, a.CycleID)), nil
		})

	s.AddTool(mcp.NewTool("list_members",
		mcp.WithDescription("List members of a workspace (user IDs for assignment)"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a workspaceIDArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListMembers(ctx, uint64(a.WorkspaceID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_states",
		mcp.WithDescription("List workflow states of a project (state IDs for create/update)"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getProjectArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListStates(ctx, uint64(a.ProjectID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_labels",
		mcp.WithDescription("List labels of a project (label IDs for create/update)"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getProjectArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListLabels(ctx, uint64(a.ProjectID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_issue_types",
		mcp.WithDescription("List issue types available in a workspace (optionally filtered by project)"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID")),
		mcp.WithInteger("project_id", mcp.Description("Project ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a issueTypesArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListIssueTypes(ctx, uint64(a.WorkspaceID), uint64(a.ProjectID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_notifications",
		mcp.WithDescription("List the current user's notifications"),
		mcp.WithBoolean("unread_only", mcp.Description("Only unread (default false)")),
		mcp.WithInteger("limit", mcp.Description("Page size")),
		mcp.WithInteger("offset", mcp.Description("Offset"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a notificationsArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListNotifications(ctx, a.UnreadOnly, a.Limit, a.Offset)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("list_pages",
		mcp.WithDescription("List document pages of a project"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getProjectArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListPages(ctx, uint64(a.ProjectID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_page",
		mcp.WithDescription("Get the content of one document page"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID")),
		mcp.WithInteger("page_id", mcp.Required(), mcp.Description("Page ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a getPageArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.GetPage(ctx, uint64(a.ProjectID), uint64(a.PageID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})
}

// uintPtr64 returns a pointer to v.
func uintPtr64(v uint64) *uint64 { return &v }
```

在 `sdk/mcp/server.go` 的 `New` 中 `return s` 之前加一行：

```go
	registerCoreTools(s, cli)
```

并同步更新 `server_test.go` 中 `TestNew_ToolsListInitiallyEmpty`：断言改为 `len(names) < 19`（测试名可改为 `TestNew_ToolsListCoreRegistered`，改完 `go vet` 确认无遗留引用）。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd sdk && go test ./mcp/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add sdk/mcp/tools_core.go sdk/mcp/tools_core_test.go sdk/mcp/server.go sdk/mcp/server_test.go
git commit -m "feat(mcp): add 19 core tools (workspace/project/issue/cycle/meta)"
```

### Task 12: MCP AI 工具（5 个）

**Files:**
- Create: `sdk/mcp/tools_ai.go`
- Create: `sdk/mcp/tools_ai_test.go`
- Modify: `sdk/mcp/server.go`（New 中加一行 `registerAITools(s, cli)`）

**Interfaces:**
- Consumes: Task 10/11 的 `New`、`decodeArgs`、`toolResultJSON`、`toolAPIError`；Task 9 的 `AISearch`、`AIChat`、`ListAgents`、`DispatchAgent`、`GetAgentTask`、`GetAgentTaskLogs`
- Produces: `func registerAITools(s *server.MCPServer, cli *client.Client)`

- [ ] **Step 1: 写失败测试**

`sdk/mcp/tools_ai_test.go`：

```go
package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/reqmango/tools/client"
)

func TestAITools_ToolCountIs24(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL, "reqmango_pat_test"))

	names := listToolNames(t, httpSrv.URL, sessionID)
	if len(names) != 24 {
		t.Fatalf("expected 24 tools total, got %d: %v", len(names), names)
	}
	for _, w := range []string{"ai_search", "ai_chat", "list_agents", "dispatch_agent", "get_agent_task"} {
		if !contains(names, w) {
			t.Errorf("missing tool %s", w)
		}
	}
}

func contains(hay []string, needle string) bool {
	for _, s := range hay {
		if s == needle {
			return true
		}
	}
	return false
}

func TestAIChat_AggregatesStream(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/projects/5/ai/chat": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			for _, ch := range []map[string]any{
				{"type": "text", "content": "Hello"},
				{"type": "done", "thread_id": 99},
			} {
				b, _ := json.Marshal(ch)
				fmt.Fprintf(w, "data: %s\n\n", b)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL, "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "ai_chat", map[string]any{
		"project_id": 5, "message": "hello",
	})
	if isError || len(texts) != 1 {
		t.Fatalf("ai_chat: texts=%v isError=%v", texts, isError)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(texts[0]), &out); err != nil {
		t.Fatalf("ai_chat output is not JSON: %v", err)
	}
	if out["text"] != "Hello" || out["thread_id"] != float64(99) {
		t.Fatalf("unexpected ai_chat output %v", out)
	}
}

func TestGetAgentTask_MergesLogs(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/workspaces/2/agent-tasks/15": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"id": 15, "title": "T1", "status": "running", "progress": 50})
		},
		"/api/v1/workspaces/2/agent-tasks/15/logs": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "message": "working"}})
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL, "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "get_agent_task", map[string]any{
		"workspace_id": 2, "task_id": 15,
	})
	if isError || len(texts) != 1 {
		t.Fatalf("get_agent_task: texts=%v isError=%v", texts, isError)
	}
	if !strings.Contains(texts[0], `"status": "running"`) || !strings.Contains(texts[0], `"message": "working"`) {
		t.Fatalf("get_agent_task should merge logs, got %s", texts[0])
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd sdk && go test ./mcp/ -run "TestAITools|TestAIChat|TestGetAgentTask" -v`
Expected: FAIL（registerAITools 未调用）

- [ ] **Step 3: 实现 tools_ai.go**

`sdk/mcp/tools_ai.go`：

```go
package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/reqmango/tools/client"
)

type aiSearchArgs struct {
	ProjectID int64  `json:"project_id"`
	Query     string `json:"query"`
}

type aiChatArgs struct {
	ProjectID int64  `json:"project_id"`
	Message   string `json:"message"`
	ThreadID  int64  `json:"thread_id"`
}

type dispatchAgentArgs struct {
	WorkspaceID int64  `json:"workspace_id"`
	AgentID     int64  `json:"agent_id"`
	Task        string `json:"task"`
	IssueID     int64  `json:"issue_id"`
}

type agentTaskArgs struct {
	WorkspaceID int64 `json:"workspace_id"`
	TaskID      int64 `json:"task_id"`
}

// registerAITools adds the 5 AI capability tools.
func registerAITools(s *server.MCPServer, cli *client.Client) {
	s.AddTool(mcp.NewTool("ai_search",
		mcp.WithDescription("Convert a natural-language question into an issue search (returns RQL + matching issues)"),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID to search in")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Natural-language question, e.g. 'high priority bugs from last week'"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a aiSearchArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.AISearch(ctx, uint64(a.ProjectID), a.Query)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("ai_chat",
		mcp.WithDescription("Send one message to the project AI assistant. The SSE stream is aggregated server-side into a single reply. Long-running (up to 5 min)."),
		mcp.WithInteger("project_id", mcp.Required(), mcp.Description("Project ID")),
		mcp.WithString("message", mcp.Required(), mcp.Description("Message to send")),
		mcp.WithInteger("thread_id", mcp.Description("Thread ID from a previous ai_chat reply to continue the conversation"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a aiChatArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var tid *uint64
			if a.ThreadID != 0 {
				tid = uintPtr64(uint64(a.ThreadID))
			}
			reply, err := cli.AIChat(ctx, uint64(a.ProjectID), a.Message, tid)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(map[string]any{
				"text":       reply.Text,
				"thread_id":  reply.ThreadID,
				"tool_calls": reply.ToolCalls,
			}), nil
		})

	s.AddTool(mcp.NewTool("list_agents",
		mcp.WithDescription("List AI agents available in a workspace"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a workspaceIDArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := cli.ListAgents(ctx, uint64(a.WorkspaceID))
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("dispatch_agent",
		mcp.WithDescription("Trigger an AI agent to run a task. Long-running (up to 5 min); returns the activity record."),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID")),
		mcp.WithInteger("agent_id", mcp.Required(), mcp.Description("Agent ID from list_agents")),
		mcp.WithString("task", mcp.Required(), mcp.Description("Task description for the agent")),
		mcp.WithInteger("issue_id", mcp.Description("Related issue ID"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a dispatchAgentArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var issueID *uint64
			if a.IssueID != 0 {
				issueID = uintPtr64(uint64(a.IssueID))
			}
			out, err := cli.DispatchAgent(ctx, uint64(a.WorkspaceID), uint64(a.AgentID), a.Task, issueID, nil)
			if err != nil {
				return toolAPIError(err), nil
			}
			return toolResultJSON(out), nil
		})

	s.AddTool(mcp.NewTool("get_agent_task",
		mcp.WithDescription("Get an agent task's status, progress and recent logs"),
		mcp.WithInteger("workspace_id", mcp.Required(), mcp.Description("Workspace ID")),
		mcp.WithInteger("task_id", mcp.Required(), mcp.Description("Task ID (from agent-tasks listing)"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var a agentTaskArgs
			if err := decodeArgs(req, &a); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			task, err := cli.GetAgentTask(ctx, uint64(a.WorkspaceID), uint64(a.TaskID))
			if err != nil {
				return toolAPIError(err), nil
			}
			logs, _ := cli.GetAgentTaskLogs(ctx, uint64(a.WorkspaceID), uint64(a.TaskID))
			return toolResultJSON(map[string]any{"task": task, "logs": logs}), nil
		})
}
```

在 `sdk/mcp/server.go` 的 `New` 中 `registerCoreTools(s, cli)` 之后加一行：

```go
	registerAITools(s, cli)
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd sdk && go test ./mcp/ -v`
Expected: PASS（`TestCoreTools_ToolCountAndNames` 断言 ≥19 个核心工具，`TestAITools_ToolCountIs24` 断言总计 24 个，两者同时成立）

- [ ] **Step 5: Commit**

```bash
git add sdk/mcp/tools_ai.go sdk/mcp/tools_ai_test.go sdk/mcp/server.go
git commit -m "feat(mcp): add 5 AI tools (search/chat/agents)"
```

---

## Phase D：CLI

### Task 13: CLI 骨架 + config + auth/workspace 命令

**Files:**
- Create: `sdk/cli/config.go`、`sdk/cli/root.go`、`sdk/cli/output.go`、`sdk/cli/cmd_auth.go`、`sdk/cli/cmd_workspace.go`
- Create: `sdk/cmd/reqmango/main.go`
- Create: `sdk/cli/config_test.go`、`sdk/cli/cmd_auth_test.go`

**Interfaces:**
- Consumes: Task 5/6 的 client 方法（`New`、`DefaultBaseURL`、`Login`、`CreatePAT`、`Me`、`ListWorkspaces`、`RevokePAT`、`ListPATs`）
- Produces（Task 14/15 依赖）:
  - `type Config struct { APIURL, PAT string; WorkspaceID, ProjectID uint64 }` + JSON tags `api_url/pat/workspace_id/project_id`
  - `func DefaultConfigPath() (string, error)`、`LoadConfig(path)`、`SaveConfig(path, cfg)`
  - `func NewRootCommand() *cobra.Command`（持久 flag：`--config` `--api-url` `--workspace` `--project` `--output`）
  - `func setup(cmd *cobra.Command) (*client.Client, *Config, error)`、`func resolveWorkspace(cmd, cfg) (uint64, error)`、`func resolveProject(cmd, cfg) (uint64, error)`
  - `func PrintTable(w io.Writer, headers []string, rows [][]string)`、`func PrintJSON(w io.Writer, v any) error`、`func printResult(cmd *cobra.Command, headers []string, rows [][]string, v any) error`

- [ ] **Step 1: 引入依赖**

Run: `cd sdk && go get github.com/spf13/cobra@v1.10.2 && go get golang.org/x/term@latest`
Expected: go.mod/go.sum 更新

- [ ] **Step 2: 写失败测试**

`sdk/cli/config_test.go`：

```go
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reqmango", "config.json")

	cfg := &Config{APIURL: "http://localhost:8000/api/v1", PAT: "reqmango_pat_abc", WorkspaceID: 2, ProjectID: 5}
	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	got, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got.APIURL != cfg.APIURL || got.PAT != cfg.PAT || got.WorkspaceID != 2 || got.ProjectID != 5 {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config should be 0600, got %o", info.Mode().Perm())
	}
}

func TestLoadConfig_Missing(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "nope", "config.json"))
	if err != nil || cfg.APIURL != "" {
		t.Fatalf("expected empty config, got %+v err=%v", cfg, err)
	}
}
```

`sdk/cli/cmd_auth_test.go`：

```go
package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reqmango/tools/client"
)

func TestAuthLogin_WritesConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/login":
			json.NewEncoder(w).Encode(map[string]any{"access_token": "jwt", "token_type": "Bearer", "expires_at": "2026-09-13T00:00:00Z"})
		case "/api/v1/auth/tokens":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"token": "reqmango_pat_secret", "id": 1, "name": "cli"})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	cfgPath := filepath.Join(t.TempDir(), "config.json")
	var out, errBuf bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs([]string{
		"auth", "login", "--config", cfgPath, "--api-url", srv.URL,
		"--email", "a@b.c", "--password", "pw",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v\nstderr: %s", err, errBuf.String())
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.PAT != "reqmango_pat_secret" || cfg.APIURL != srv.URL {
		t.Fatalf("unexpected config %+v", cfg)
	}
}

func TestWorkspaceSwitch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{{"id": 2, "name": "Acme", "slug": "acme"}})
	}))
	defer srv.Close()

	cfgPath := filepath.Join(t.TempDir(), "config.json")
	if err := SaveConfig(cfgPath, &Config{APIURL: srv.URL, PAT: "reqmango_pat_x"}); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"workspace", "switch", "2", "--config", cfgPath})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v\nstderr: %s", err, errBuf.String())
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil || cfg.WorkspaceID != 2 {
		t.Fatalf("workspace not persisted: %+v err=%v", cfg, err)
	}
	if !strings.Contains(out.String(), "Acme") {
		t.Fatalf("expected confirmation with workspace name, got %q", out.String())
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `cd sdk && go test ./cli/ -run "TestConfig|TestAuthLogin|TestWorkspaceSwitch" -v`
Expected: FAIL（cli 包不存在）

- [ ] **Step 4: 实现**

`sdk/cli/config.go`：

```go
// Package cli implements the reqmango command tree.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config is the persisted CLI state.
type Config struct {
	APIURL      string `json:"api_url"`
	PAT         string `json:"pat"`
	WorkspaceID uint64 `json:"workspace_id"`
	ProjectID   uint64 `json:"project_id"`
}

// DefaultConfigPath returns ~/.reqmango/config.json.
func DefaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "reqmango", "config.json"), nil
}

// LoadConfig reads the config file. A missing file yields an empty Config.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// SaveConfig writes the config file (0600, private dir).
func SaveConfig(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
```

`sdk/cli/output.go`：

```go
package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// PrintTable writes aligned columns to w.
func PrintTable(w io.Writer, headers []string, rows [][]string) {
	cols := len(headers)
	widths := make([]int, cols)
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i := 0; i < cols && i < len(row); i++ {
			if l := len(row[i]); l > widths[i] {
				widths[i] = l
			}
		}
	}
	writeRow := func(cells []string) {
		for i := 0; i < cols; i++ {
			if i > 0 {
				fmt.Fprint(w, "  ")
			}
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			fmt.Fprintf(w, "%-*s", widths[i], cell)
		}
		fmt.Fprintln(w)
	}
	writeRow(headers)
	sep := make([]string, cols)
	for i := range sep {
		sep[i] = strings.Repeat("-", widths[i])
	}
	writeRow(sep)
	for _, row := range rows {
		writeRow(row)
	}
}

// PrintJSON writes v as indented JSON.
func PrintJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

// printResult renders rows as a table, or v as JSON when --output=json.
func printResult(cmd *cobra.Command, headers []string, rows [][]string, v any) error {
	if cmd.Flag("output").Value.String() == "json" {
		return PrintJSON(cmd.OutOrStdout(), v)
	}
	PrintTable(cmd.OutOrStdout(), headers, rows)
	return nil
}
```

（output.go 需 `import "strings"`。）

`sdk/cli/root.go`：

```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reqmango/tools/client"
)

// NewRootCommand builds the reqmango command tree.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "reqmango",
		Short:         "reqmango CLI - manage issues, cycles and agents from the terminal",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().String("config", "", "config file (default ~/.reqmango/config.json)")
	root.PersistentFlags().String("api-url", "", "override API base URL")
	root.PersistentFlags().Uint64("workspace", 0, "override workspace ID")
	root.PersistentFlags().Uint64("project", 0, "override project ID")
	root.PersistentFlags().String("output", "table", "output format: table|json")

	root.AddCommand(newAuthCmd())
	root.AddCommand(newWorkspaceCmd())
	return root
}

// configPath resolves the --config flag to a path.
func configPath(cmd *cobra.Command) (string, error) {
	p := cmd.Flag("config").Value.String()
	if p != "" {
		return p, nil
	}
	return DefaultConfigPath()
}

// setup resolves the config file and builds the API client.
func setup(cmd *cobra.Command) (*client.Client, *Config, error) {
	path, err := configPath(cmd)
	if err != nil {
		return nil, nil, err
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		return nil, nil, err
	}
	apiURL := cmd.Flag("api-url").Value.String()
	if apiURL == "" {
		apiURL = cfg.APIURL
	}
	if apiURL == "" {
		apiURL = client.DefaultBaseURL
	}
	if cfg.PAT == "" {
		return nil, nil, fmt.Errorf("not logged in: run `reqmango auth login` first")
	}
	return client.New(apiURL, cfg.PAT), cfg, nil
}

// resolveWorkspace returns --workspace flag > config > error.
func resolveWorkspace(cmd *cobra.Command, cfg *Config) (uint64, error) {
	if v, _ := cmd.Flags().GetUint64("workspace"); v != 0 {
		return v, nil
	}
	if cfg.WorkspaceID != 0 {
		return cfg.WorkspaceID, nil
	}
	return 0, fmt.Errorf("no workspace selected: pass --workspace or run `reqmango workspace switch <id>`")
}

// resolveProject returns --project flag > config > error.
func resolveProject(cmd *cobra.Command, cfg *Config) (uint64, error) {
	if v, _ := cmd.Flags().GetUint64("project"); v != 0 {
		return v, nil
	}
	if cfg.ProjectID != 0 {
		return cfg.ProjectID, nil
	}
	return 0, fmt.Errorf("no project selected: pass --project <id>")
}
```

`sdk/cli/cmd_auth.go`：

```go
package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/reqmango/tools/client"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with reqmango (login / logout / status / revoke)",
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthLogoutCmd(), newAuthStatusCmd(), newAuthRevokeCmd())
	return cmd
}

// promptLine reads one line from stdin after printing prompt to stderr.
func promptLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	var s string
	_, err := fmt.Fscanln(os.Stdin, &s)
	return s, err
}

// readPassword reads a hidden password from the terminal.
func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	data, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func newAuthLoginCmd() *cobra.Command {
	var email, password string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in with email+password and store a PAT in the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			cfg, err := LoadConfig(path)
			if err != nil {
				return err
			}
			apiURL := cmd.Flag("api-url").Value.String()
			if apiURL == "" {
				apiURL = cfg.APIURL
			}
			if apiURL == "" {
				apiURL = client.DefaultBaseURL
			}

			if email == "" {
				email, err = promptLine("Email: ")
				if err != nil {
					return fmt.Errorf("read email: %w", err)
				}
			}
			if password == "" {
				password, err = readPassword("Password: ")
				if err != nil {
					return fmt.Errorf("read password: %w", err)
				}
			}

			ctx := context.Background()
			anon := client.New(apiURL, "")
			tok, err := anon.Login(ctx, email, password)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}
			// Mint a PAT with the short-lived JWT, then discard the JWT.
			jwtCli := client.New(apiURL, tok.AccessToken)
			host, _ := os.Hostname()
			patResp, err := jwtCli.CreatePAT(ctx, client.CreatePATRequest{Name: "cli-" + host})
			if err != nil {
				return fmt.Errorf("create PAT failed: %w", err)
			}

			cfg.APIURL = apiURL
			cfg.PAT = patResp.Token
			if err := SaveConfig(path, cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s. PAT %q saved to %s\n", email, patResp.TokenPrefix, path)
			return nil
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "email (prompted if omitted)")
	cmd.Flags().StringVar(&password, "password", "", "password (hidden prompt if omitted)")
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the stored PAT from the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			cfg, err := LoadConfig(path)
			if err != nil {
				return err
			}
			cfg.PAT = ""
			if err := SaveConfig(path, cfg); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the current config and authenticated user",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			cfg, err := LoadConfig(path)
			if err != nil {
				return err
			}
			apiURL := cfg.APIURL
			if apiURL == "" {
				apiURL = client.DefaultBaseURL
			}
			masked := "<none>"
			if cfg.PAT != "" {
				masked = cfg.PAT[:min(len(cfg.PAT), len("reqmango_pat_ab3d"))] + "…"
			}
			rows := [][]string{
				{"API URL", apiURL},
				{"PAT", masked},
				{"Workspace ID", strconv.FormatUint(cfg.WorkspaceID, 10)},
				{"Project ID", strconv.FormatUint(cfg.ProjectID, 10)},
				{"Config", path},
			}
			PrintTable(cmd.OutOrStdout(), []string{"Setting", "Value"}, rows)

			if cfg.PAT != "" {
				cli := client.New(apiURL, cfg.PAT)
				me, err := cli.Me(context.Background())
				if err != nil {
					if apiErr := client.AsAPIError(err); apiErr != nil && apiErr.StatusCode == 401 {
						fmt.Fprintln(cmd.OutOrStdout(), "\nPAT is invalid or revoked - run `reqmango auth login` to re-authenticate.")
						return nil
					}
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "\nLogged in as %s (%s)\n", me.DisplayName, me.Email)
			}
			return nil
		},
	}
}

func newAuthRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <pat-id>",
		Short: "Revoke a PAT by ID (list them with `reqmango auth revoke --list`)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, _, err := setup(cmd)
			if err != nil {
				return err
			}
			listOnly, _ := cmd.Flags().GetBool("list")
			if listOnly || len(args) == 0 {
				pats, err := cli.ListPATs(context.Background())
				if err != nil {
					return err
				}
				rows := make([][]string, 0, len(pats))
				for _, p := range pats {
					revoked := ""
					if p.RevokedAt != nil {
						revoked = p.RevokedAt.Format("2006-01-02")
					}
					rows = append(rows, []string{strconv.FormatUint(p.ID, 10), p.Name, p.TokenPrefix, revoked})
				}
				PrintTable(cmd.OutOrStdout(), []string{"ID", "Name", "Prefix", "Revoked"}, rows)
				return nil
			}
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid PAT id: %s", args[0])
			}
			if err := cli.RevokePAT(context.Background(), id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Revoked PAT %d.\n", id)
			return nil
		},
	}
}
```

（`min` 为 Go 1.21 内置函数；`strings` import 若未用到可删。）

`sdk/cli/cmd_workspace.go`：

```go
package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "List and switch workspaces",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List accessible workspaces",
			RunE: func(cmd *cobra.Command, args []string) error {
				cli, _, err := setup(cmd)
				if err != nil {
					return err
				}
				ws, err := cli.ListWorkspaces(context.Background())
				if err != nil {
					return err
				}
				rows := make([][]string, 0, len(ws))
				for _, w := range ws {
					rows = append(rows, []string{strconv.FormatUint(w.ID, 10), w.Name, w.Slug})
				}
				return printResult(cmd, []string{"ID", "Name", "Slug"}, rows, ws)
			},
		},
		&cobra.Command{
			Use:   "switch <id>",
			Short: "Remember the current workspace in the config",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid workspace id: %s", args[0])
				}
				path, err := configPath(cmd)
				if err != nil {
					return err
				}
				cfg, err := LoadConfig(path)
				if err != nil {
					return err
				}
				apiURL := cmd.Flag("api-url").Value.String()
				if apiURL == "" {
					apiURL = cfg.APIURL
				}
				name := ""
				if cfg.PAT != "" {
					cli := clientFor(cfg, apiURL)
					ws, err := cli.ListWorkspaces(context.Background())
					if err == nil {
						for _, w := range ws {
							if w.ID == id {
								name = w.Name
								break
							}
						}
					}
				}
				cfg.WorkspaceID = id
				if err := SaveConfig(path, cfg); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Switched to workspace %d %s\n", id, name)
				return nil
			},
		},
	)
	return cmd
}

// clientFor builds a client from config (helper for commands that only
// need config access).
func clientFor(cfg *Config, apiURL string) *client.Client {
	if apiURL == "" {
		apiURL = client.DefaultBaseURL
	}
	return client.New(apiURL, cfg.PAT)
}
```

（cmd_workspace.go 需 `import "github.com/reqmango/tools/client"`。）

`sdk/cmd/reqmango/main.go`：

```go
// Command reqmango is the reqmango daily-ops CLI.
package main

import (
	"os"

	"github.com/reqmango/tools/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 5: 运行测试确认通过**

Run: `cd sdk && go test ./cli/ -v && go build ./cmd/reqmango`
Expected: PASS（若 `output.go` 或 `cmd_auth.go` 有未使用 import 导致编译失败，删除对应 import 行）

- [ ] **Step 6: Commit**

```bash
git add sdk/go.mod sdk/go.sum sdk/cli/config.go sdk/cli/root.go sdk/cli/output.go sdk/cli/cmd_auth.go sdk/cli/cmd_workspace.go sdk/cli/config_test.go sdk/cli/cmd_auth_test.go sdk/cmd/reqmango/main.go
git commit -m "feat(cli): add CLI scaffold with config, auth and workspace commands"
```

### Task 14: CLI issue/project/cycle/meta/agent/ask 命令

**Files:**
- Create: `sdk/cli/cmd_issue.go`、`sdk/cli/cmd_project.go`、`sdk/cli/cmd_cycle.go`、`sdk/cli/cmd_meta.go`、`sdk/cli/cmd_agent.go`、`sdk/cli/cmd_ask.go`
- Create: `sdk/cli/cmd_issue_test.go`、`sdk/cli/cmd_project_test.go`、`sdk/cli/cmd_cycle_test.go`、`sdk/cli/cmd_agent_test.go`
- Modify: `sdk/cli/root.go`（注册 6 个新子命令）

**Interfaces:**
- Consumes: Task 13 的 `setup`/`resolveWorkspace`/`resolveProject`/`printResult`；Task 6-9 的 client 方法
- Produces: `func newIssueCmd()`、`newProjectCmd()`、`newCycleCmd()`、`newMetaCmd()`、`newAgentCmd()`、`newAskCmd()`

- [ ] **Step 1: 写失败测试**

`sdk/cli/cmd_issue_test.go`：

```go
package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// newCLI prepares a root command against a backend and a temp config.
func newCLI(t *testing.T, srvURL string, cfgPatch *Config) (*bytes.Buffer, func(args ...string) error) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	cfg := &Config{APIURL: srvURL, PAT: "reqmango_pat_x", WorkspaceID: 2, ProjectID: 5}
	if cfgPatch != nil {
		*cfg = *cfgPatch
	}
	if err := SaveConfig(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	run := func(args ...string) error {
		out.Reset()
		root.SetArgs(append(args, "--config", cfgPath))
		return root.Execute()
	}
	return &out, run
}

func TestIssueList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issues" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("X-Total-Count", "1")
		json.NewEncoder(w).Encode([]map[string]any{{"id": 11, "name": "Login broken", "sequence_id": 42, "priority": "high", "state_name": "Todo"}})
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL, nil)
	if err := run("issue", "list", "--project", "5"); err != nil {
		t.Fatalf("issue list: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Login broken")) {
		t.Fatalf("expected issue in table output, got %q", out.String())
	}
}

func TestIssueShow_ByCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/projects":
			json.NewEncoder(w).Encode([]map[string]any{{"id": 5, "identifier": "DEMO", "workspace_id": 2}})
		case "/api/v1/issues":
			w.Header().Set("X-Total-Count", "1")
			json.NewEncoder(w).Encode([]map[string]any{{"id": 11, "sequence_id": 42}})
		case "/api/v1/issues/11":
			json.NewEncoder(w).Encode(map[string]any{"id": 11, "name": "Login broken", "sequence_id": 42, "priority": "high", "state_name": "Todo"})
		case "/api/v1/comments/issue/11":
			json.NewEncoder(w).Encode(map[string]any{"comments": []map[string]any{}, "total": 0})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL, nil)
	if err := run("issue", "show", "DEMO-42", "--comments"); err != nil {
		t.Fatalf("issue show: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("DEMO-42")) {
		t.Fatalf("expected code in output, got %q", out.String())
	}
}

func TestIssueCreate_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issues" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "New bug" || body["priority"] != "high" {
			t.Errorf("unexpected body %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": 12, "name": "New bug", "sequence_id": 43, "priority": "high"})
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL, nil)
	if err := run("issue", "create", "--project", "5", "--title", "New bug", "--priority", "high", "--output", "json"); err != nil {
		t.Fatalf("issue create: %v", err)
	}
	var created map[string]any
	if err := json.Unmarshal(out.Bytes(), &created); err != nil {
		t.Fatalf("expected JSON output, got %q: %v", out.String(), err)
	}
	if created["id"] != float64(12) {
		t.Fatalf("unexpected output %v", created)
	}
}
```

`sdk/cli/cmd_project_test.go`：

```go
package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProjectCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects" || r.URL.Query().Get("workspace_id") != "2" {
			t.Errorf("unexpected request %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "NewProj" || body["identifier"] != "NP" {
			t.Errorf("unexpected body %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": 9, "name": "NewProj", "identifier": "NP"})
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL, nil)
	if err := run("project", "create", "--name", "NewProj", "--identifier", "NP", "--workspace", "2"); err != nil {
		t.Fatalf("project create: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("NewProj")) {
		t.Fatalf("unexpected output %q", out.String())
	}
}

func TestProjectShow_ByIdentifier(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/projects" && r.Method == "GET":
			json.NewEncoder(w).Encode([]map[string]any{{"id": 5, "name": "Demo", "identifier": "DEMO", "workspace_id": 2}})
		case r.URL.Path == "/api/v1/projects/5":
			json.NewEncoder(w).Encode(map[string]any{"id": 5, "name": "Demo", "identifier": "DEMO", "workspace_id": 2, "total_issues": 42})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL, nil)
	if err := run("project", "show", "DEMO"); err != nil {
		t.Fatalf("project show: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Demo")) {
		t.Fatalf("unexpected output %q", out.String())
	}
}
```

`sdk/cli/cmd_cycle_test.go`：

```go
package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCycleBurndown_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cycles/3/burndown" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"cycle_id": 3, "cycle_name": "S1", "total_issues": 10, "is_on_track": true, "daily_points": []map[string]any{}})
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL, nil)
	if err := run("cycle", "burndown", "3", "--output", "json"); err != nil {
		t.Fatalf("cycle burndown: %v", err)
	}
	var b map[string]any
	if err := json.Unmarshal(out.Bytes(), &b); err != nil {
		t.Fatalf("expected JSON, got %q: %v", out.String(), err)
	}
	if b["cycle_id"] != float64(3) {
		t.Fatalf("unexpected output %v", b)
	}
}
```

`sdk/cli/cmd_agent_test.go`：

```go
package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAgentDispatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/2/agents/7/dispatch" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["task"] != "triage bugs" {
			t.Errorf("unexpected body %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 1, "agent_id": 7, "action": "dispatch", "agent_name": "Triage"})
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL, nil)
	if err := run("agent", "dispatch", "7", "triage bugs"); err != nil {
		t.Fatalf("agent dispatch: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Triage")) {
		t.Fatalf("unexpected output %q", out.String())
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd sdk && go test ./cli/ -run "TestIssue|TestProject|TestCycle|TestAgentDispatch" -v`
Expected: FAIL（命令未注册）

- [ ] **Step 3: 实现六个命令文件**

`sdk/cli/cmd_issue.go`：

```go
package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reqmango/tools/client"
)

func newIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "issue",
		Aliases: []string{"issues"},
		Short:   "List, show, create, update and search issues",
	}
	cmd.AddCommand(newIssueListCmd(), newIssueShowCmd(), newIssueCreateCmd(), newIssueUpdateCmd(), newIssueSearchCmd())
	return cmd
}

// resolveAssignee maps "" → 0, "me" → current user ID, else parses uint64.
func resolveAssignee(ctx context.Context, cli *client.Client, s string) (uint64, error) {
	if s == "" {
		return 0, nil
	}
	if s == "me" {
		me, err := cli.Me(ctx)
		if err != nil {
			return 0, err
		}
		return me.ID, nil
	}
	return strconv.ParseUint(s, 10, 64)
}

func issueRow(i client.Issue) []string {
	return []string{
		strconv.FormatUint(i.ID, 10),
		strconv.Itoa(i.SequenceID),
		i.Name,
		i.Priority,
		i.StateName,
	}
}

var issueTableHeaders = []string{"ID", "Seq", "Name", "Priority", "State"}

func newIssueListCmd() *cobra.Command {
	var rql, priority, assignee string
	var stateID, cycleID, limit int64
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues (optionally filtered by --rql)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			ctx := context.Background()
			opts := client.IssueListOptions{RQL: rql, Priority: priority, StateID: uint64(stateID), CycleID: uint64(cycleID), Limit: int(limit)}
			if projectID, err := resolveProject(cmd, cfg); err == nil {
				opts.ProjectID = projectID
			}
			if wsID, err := resolveWorkspace(cmd, cfg); err == nil {
				opts.WorkspaceID = wsID
			}
			if assignee != "" {
				id, err := resolveAssignee(ctx, cli, assignee)
				if err != nil {
					return err
				}
				opts.AssigneeID = id
			}
			res, err := cli.ListIssues(ctx, opts)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(res.Items))
			for _, i := range res.Items {
				rows = append(rows, issueRow(i))
			}
			if err := printResult(cmd, issueTableHeaders, rows, res); err != nil {
				return err
			}
			if cmd.Flag("output").Value.String() != "json" {
				fmt.Fprintf(cmd.OutOrStdout(), "\nTotal: %d\n", res.Total)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&rql, "rql", "", "RQL query, e.g. `priority = \"high\"`")
	cmd.Flags().StringVar(&priority, "priority", "", "filter by priority")
	cmd.Flags().Int64Var(&stateID, "state", 0, "filter by state ID")
	cmd.Flags().StringVar(&assignee, "assignee", "", "filter by assignee user ID or `me`")
	cmd.Flags().Int64Var(&cycleID, "cycle", 0, "filter by cycle ID")
	cmd.Flags().Int64Var(&limit, "limit", 0, "page size")
	return cmd
}

func newIssueShowCmd() *cobra.Command {
	var withComments bool
	cmd := &cobra.Command{
		Use:   "show <id|code>",
		Short: "Show one issue by numeric ID or code like DEMO-42",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			ctx := context.Background()
			wsID, _ := resolveWorkspace(cmd, cfg)
			id, err := cli.ResolveIssueCode(ctx, wsID, args[0])
			if err != nil && !isNumeric(args[0]) {
				return err
			}
			if isNumeric(args[0]) {
				id, err = strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid issue id: %s", args[0])
				}
			}
			iss, err := cli.GetIssue(ctx, id)
			if err != nil {
				return err
			}
			rows := [][]string{
				{"ID", strconv.FormatUint(iss.ID, 10)},
				{"Code", args[0]},
				{"Name", iss.Name},
				{"Priority", iss.Priority},
				{"State", iss.StateName},
			}
			PrintTable(cmd.OutOrStdout(), []string{"Field", "Value"}, rows)
			if withComments {
				comments, _, err := cli.ListComments(ctx, id, 1, 100)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "\nComments (%d):\n", len(comments))
				for _, c := range comments {
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", c.CreatedAt.Format("2006-01-02 15:04"), c.Body)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&withComments, "comments", false, "include comments")
	return cmd
}

func isNumeric(s string) bool {
	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}

func newIssueCreateCmd() *cobra.Command {
	var title, desc, priority, assignee, labels string
	var stateID, typeID, parentID, cycleID int64
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			projectID, err := resolveProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			proj, err := cli.GetProject(ctx, projectID)
			if err != nil {
				return err
			}
			body := &client.CreateIssueRequest{Name: title, DescriptionHTML: desc, Priority: priority}
			if stateID != 0 {
				body.StateID = &[]uint64{uint64(stateID)}[0]
			}
			if typeID != 0 {
				body.TypeID = &[]uint64{uint64(typeID)}[0]
			}
			if parentID != 0 {
				body.ParentID = &[]uint64{uint64(parentID)}[0]
			}
			if cycleID != 0 {
				body.CycleID = &[]uint64{uint64(cycleID)}[0]
			}
			if assignee != "" {
				id, err := resolveAssignee(ctx, cli, assignee)
				if err != nil {
					return err
				}
				body.AssigneeIDs = []uint64{id}
			}
			for _, s := range strings.Split(labels, ",") {
				if n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64); err == nil {
					body.LabelIDs = append(body.LabelIDs, n)
				}
			}
			iss, err := cli.CreateIssue(ctx, projectID, proj.WorkspaceID, body)
			if err != nil {
				return err
			}
			if cmd.Flag("output").Value.String() == "json" {
				return PrintJSON(cmd.OutOrStdout(), iss)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created issue %s-%d (id %d)\n", proj.Identifier, iss.SequenceID, iss.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "issue title (required)")
	cmd.MarkFlagRequired("title")
	cmd.Flags().StringVar(&desc, "desc", "", "description (HTML)")
	cmd.Flags().StringVar(&priority, "priority", "", "priority: none|low|medium|high|urgent")
	cmd.Flags().Int64Var(&stateID, "state", 0, "state ID")
	cmd.Flags().Int64Var(&typeID, "type", 0, "issue type ID")
	cmd.Flags().StringVar(&assignee, "assignee", "", "assignee user ID or `me`")
	cmd.Flags().StringVar(&labels, "labels", "", "comma-separated label IDs")
	cmd.Flags().Int64Var(&parentID, "parent", 0, "parent issue ID")
	cmd.Flags().Int64Var(&cycleID, "cycle", 0, "cycle ID")
	return cmd
}

func newIssueUpdateCmd() *cobra.Command {
	var title, desc, priority, assignee, labels string
	var stateID, cycleID int64
	cmd := &cobra.Command{
		Use:   "update <id|code>",
		Short: "Update an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			ctx := context.Background()
			wsID, _ := resolveWorkspace(cmd, cfg)
			id, err := resolveIssueID(ctx, cli, wsID, args[0])
			if err != nil {
				return err
			}
			body := &client.UpdateIssueRequest{}
			if title != "" {
				body.Name = &title
			}
			if desc != "" {
				body.DescriptionHTML = &desc
			}
			if priority != "" {
				body.Priority = &priority
			}
			if stateID != 0 {
				body.StateID = &[]uint64{uint64(stateID)}[0]
			}
			if cycleID != 0 {
				body.CycleID = &[]uint64{uint64(cycleID)}[0]
			}
			if assignee != "" {
				aid, err := resolveAssignee(ctx, cli, assignee)
				if err != nil {
					return err
				}
				body.AssigneeIDs = []uint64{aid}
			}
			if labels != "" {
				for _, s := range strings.Split(labels, ",") {
					if n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64); err == nil {
						body.LabelIDs = append(body.LabelIDs, n)
					}
				}
			}
			iss, err := cli.UpdateIssue(ctx, id, body)
			if err != nil {
				if apiErr := client.AsAPIError(err); apiErr != nil && apiErr.StatusCode == 409 {
					fmt.Fprintf(cmd.OutOrStdout(), "Approval required: %v\n", apiErr.Body)
					return nil
				}
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated issue %d\n", iss.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&desc, "desc", "", "new description (HTML)")
	cmd.Flags().StringVar(&priority, "priority", "", "new priority")
	cmd.Flags().Int64Var(&stateID, "state", 0, "new state ID")
	cmd.Flags().StringVar(&assignee, "assignee", "", "new assignee user ID or `me` (replaces)")
	cmd.Flags().StringVar(&labels, "labels", "", "comma-separated label IDs (replaces)")
	cmd.Flags().Int64Var(&cycleID, "cycle", 0, "move to cycle ID")
	return cmd
}

// resolveIssueID accepts a numeric ID or a code like DEMO-42.
func resolveIssueID(ctx context.Context, cli *client.Client, wsID uint64, arg string) (uint64, error) {
	if isNumeric(arg) {
		return strconv.ParseUint(arg, 10, 64)
	}
	if wsID == 0 {
		return 0, fmt.Errorf("issue code %q needs --workspace to resolve the project identifier", arg)
	}
	return cli.ResolveIssueCode(ctx, wsID, arg)
}

func newIssueSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Full-text search across workspace issues",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			results, err := cli.SearchIssues(context.Background(), wsID, args[0], nil, 20)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(results))
			for _, r := range results {
				rows = append(rows, []string{r.ProjectIdentifier + "-" + strconv.Itoa(r.SequenceID), r.Name, strconv.FormatUint(r.ID, 10)})
			}
			return printResult(cmd, []string{"Code", "Name", "ID"}, rows, results)
		},
	}
}
```

`sdk/cli/cmd_project.go`：

```go
package cli

import (
	"context"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reqmango/tools/client"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"projects"},
		Short:   "List, show and create projects",
	}
	cmd.AddCommand(newProjectListCmd(), newProjectShowCmd(), newProjectCreateCmd())
	return cmd
}

func newProjectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List projects in the current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			projects, err := cli.ListProjects(context.Background(), wsID)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(projects))
			for _, p := range projects {
				rows = append(rows, []string{strconv.FormatUint(p.ID, 10), p.Identifier, p.Name, strconv.FormatInt(p.TotalIssues, 10)})
			}
			return printResult(cmd, []string{"ID", "Identifier", "Name", "Issues"}, rows, projects)
		},
	}
}

func newProjectShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id|identifier>",
		Short: "Show one project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			var proj *client.Project
			if id, err := strconv.ParseUint(args[0], 10, 64); err == nil {
				proj, err = cli.GetProject(context.Background(), id)
				if err != nil {
					return err
				}
			} else {
				projects, err := cli.ListProjects(context.Background(), wsID)
				if err != nil {
					return err
				}
				for i := range projects {
					if strings.EqualFold(projects[i].Identifier, args[0]) {
						proj = &projects[i]
						break
					}
				}
				if proj == nil {
					return fmt.Errorf("project with identifier %q not found", args[0])
				}
			}
			rows := [][]string{
				{"ID", strconv.FormatUint(proj.ID, 10)},
				{"Identifier", proj.Identifier},
				{"Name", proj.Name},
				{"Issues", strconv.FormatInt(proj.TotalIssues, 10)},
			}
			PrintTable(cmd.OutOrStdout(), []string{"Field", "Value"}, rows)
			return nil
		},
	}
}

func newProjectCreateCmd() *cobra.Command {
	var name, identifier, desc string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project in the current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			var created client.Project
			body := map[string]any{"name": name, "identifier": identifier}
			if desc != "" {
				body["description"] = desc
			}
			// POST /projects?workspace_id= returns the bare project object.
			ws := strconv.FormatUint(wsID, 10)
			_ = ws
			q := map[string]string{"workspace_id": strconv.FormatUint(wsID, 10)}
			hdr, err := cli.PostJSON(context.Background(), "/projects", queryFromMap(q), body, &created)
			_ = hdr
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created project %s (id %d)\n", created.Identifier, created.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "project name (required)")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&identifier, "identifier", "", "short identifier, max 10 chars (required)")
	cmd.MarkFlagRequired("identifier")
	cmd.Flags().StringVar(&desc, "desc", "", "description")
	return cmd
}

// queryFromMap converts a map to url.Values (helper for direct endpoints).
func queryFromMap(m map[string]string) url.Values {
	q := url.Values{}
	for k, v := range m {
		q.Set(k, v)
	}
	return q
}
```

（cmd_project.go 需 import `"fmt"` 与 `"net/url"`；`PostJSON` 接受 `url.Values` 参数。）

`sdk/cli/cmd_cycle.go`：

```go
package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newCycleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cycle",
		Aliases: []string{"cycles"},
		Short:   "List cycles and view progress / burndown",
	}
	cmd.AddCommand(newCycleListCmd(), newCycleProgressCmd(), newCycleBurndownCmd())
	return cmd
}

func newCycleListCmd() *cobra.Command {
	var status string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cycles of the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			projectID, err := resolveProject(cmd, cfg)
			if err != nil {
				return err
			}
			res, err := cli.ListCycles(context.Background(), projectID, status, 50, 0)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(res.Items))
			for _, c := range res.Items {
				rows = append(rows, []string{
					strconv.FormatUint(c.ID, 10), c.Name, c.Status,
					fmt.Sprintf("%.0f%%", c.Progress), c.StartDate,
				})
			}
			return printResult(cmd, []string{"ID", "Name", "Status", "Progress", "Start"}, rows, res)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "filter: upcoming|active|completed|cancelled")
	return cmd
}

func newCycleProgressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "progress <cycle-id>",
		Short: "Show a cycle's progress breakdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid cycle id: %s", args[0])
			}
			cli, _, err := setup(cmd)
			if err != nil {
				return err
			}
			p, err := cli.GetCycleProgress(context.Background(), id)
			if err != nil {
				return err
			}
			rows := [][]string{
				{"Cycle", fmt.Sprintf("%s (%d)", p.CycleName, p.CycleID)},
				{"Progress", fmt.Sprintf("%.0f%%", p.Progress)},
				{"Issues", fmt.Sprintf("%d/%d completed", p.CompletedIssues, p.TotalIssues)},
			}
			for _, sb := range p.StateBreakdown {
				rows = append(rows, []string{"State: " + sb.State, strconv.FormatInt(sb.Count, 10)})
			}
			PrintTable(cmd.OutOrStdout(), []string{"Metric", "Value"}, rows)
			return nil
		},
	}
}

func newCycleBurndownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "burndown <cycle-id>",
		Short: "Show a cycle's burndown data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid cycle id: %s", args[0])
			}
			cli, _, err := setup(cmd)
			if err != nil {
				return err
			}
			b, err := cli.GetCycleBurndown(context.Background(), id)
			if err != nil {
				return err
			}
			if cmd.Flag("output").Value.String() == "json" {
				return PrintJSON(cmd.OutOrStdout(), b)
			}
			onTrack := "yes"
			if !b.IsOnTrack {
				onTrack = "no"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Cycle: %s\nIssues: %d total, %d completed\nRemaining: %.0f (ideal %.0f)\nOn track: %s\n",
				b.CycleName, b.TotalIssues, b.ActualCompleted, b.ActualRemaining, b.IdealRemaining, onTrack)
			return nil
		},
	}
}
```

`sdk/cli/cmd_meta.go`：

```go
package cli

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"
)

func newMetaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meta",
		Short: "Project metadata: states, labels, issue types",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "states",
			Short: "List workflow states of the current project",
			RunE: func(cmd *cobra.Command, args []string) error {
				cli, cfg, err := setup(cmd)
				if err != nil {
					return err
				}
				projectID, err := resolveProject(cmd, cfg)
				if err != nil {
					return err
				}
				states, err := cli.ListStates(context.Background(), projectID)
				if err != nil {
					return err
				}
				rows := make([][]string, 0, len(states))
				for _, s := range states {
					rows = append(rows, []string{strconv.FormatUint(s.ID, 10), s.Name, s.Group, s.Color})
				}
				return printResult(cmd, []string{"ID", "Name", "Group", "Color"}, rows, states)
			},
		},
		&cobra.Command{
			Use:   "labels",
			Short: "List labels of the current project",
			RunE: func(cmd *cobra.Command, args []string) error {
				cli, cfg, err := setup(cmd)
				if err != nil {
					return err
				}
				projectID, err := resolveProject(cmd, cfg)
				if err != nil {
					return err
				}
				labels, err := cli.ListLabels(context.Background(), projectID)
				if err != nil {
					return err
				}
				rows := make([][]string, 0, len(labels))
				for _, l := range labels {
					rows = append(rows, []string{strconv.FormatUint(l.ID, 10), l.Name, l.Color})
				}
				return printResult(cmd, []string{"ID", "Name", "Color"}, rows, labels)
			},
		},
		&cobra.Command{
			Use:   "issue-types",
			Short: "List issue types of the current workspace",
			RunE: func(cmd *cobra.Command, args []string) error {
				cli, cfg, err := setup(cmd)
				if err != nil {
					return err
				}
				wsID, err := resolveWorkspace(cmd, cfg)
				if err != nil {
					return err
				}
				var projectID uint64
				if projectID, err = resolveProject(cmd, cfg); err != nil {
					projectID = 0
				}
				types, err := cli.ListIssueTypes(context.Background(), wsID, projectID)
				if err != nil {
					return err
				}
				rows := make([][]string, 0, len(types))
				for _, t := range types {
					rows = append(rows, []string{strconv.FormatUint(t.ID, 10), t.Name, t.Level})
				}
				return printResult(cmd, []string{"ID", "Name", "Level"}, rows, types)
			},
		},
	)
	return cmd
}
```

`sdk/cli/cmd_agent.go`：

```go
package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent",
		Aliases: []string{"agents"},
		Short:   "List and trigger AI agents, inspect agent tasks",
	}
	cmd.AddCommand(newAgentListCmd(), newAgentDispatchCmd(), newAgentTaskCmd())
	return cmd
}

func newAgentListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List agents in the current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			agents, err := cli.ListAgents(context.Background(), wsID)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(agents))
			for _, a := range agents {
				rows = append(rows, []string{strconv.FormatUint(a.ID, 10), a.Name, a.AgentType, a.Status})
			}
			return printResult(cmd, []string{"ID", "Name", "Type", "Status"}, rows, agents)
		},
	}
}

func newAgentDispatchCmd() *cobra.Command {
	var issueArg string
	cmd := &cobra.Command{
		Use:   "dispatch <agent-id> <task...>",
		Short: "Trigger an agent to run a task (long-running)",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid agent id: %s", args[0])
			}
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			task := ""
			for i, part := range args[1:] {
				if i > 0 {
					task += " "
				}
				task += part
			}
			var issueID *uint64
			if issueArg != "" {
				id, err := resolveIssueID(context.Background(), cli, wsID, issueArg)
				if err != nil {
					return err
				}
				issueID = &id
			}
			act, err := cli.DispatchAgent(context.Background(), wsID, agentID, task, issueID, nil)
			if err != nil {
				return err
			}
			if cmd.Flag("output").Value.String() == "json" {
				return PrintJSON(cmd.OutOrStdout(), act)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Dispatched to %s (activity id %d)\n", act.AgentName, act.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&issueArg, "issue", "", "related issue id or code")
	return cmd
}

func newAgentTaskCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "task <task-id>",
		Short: "Show an agent task's status, progress and recent logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task id: %s", args[0])
			}
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			wsID, err := resolveWorkspace(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			task, err := cli.GetAgentTask(ctx, wsID, taskID)
			if err != nil {
				return err
			}
			if cmd.Flag("output").Value.String() == "json" {
				return PrintJSON(cmd.OutOrStdout(), task)
			}
			rows := [][]string{
				{"Title", task.Title},
				{"Status", task.Status},
				{"Progress", fmt.Sprintf("%d%%", task.Progress)},
			}
			if task.ErrorInfo != "" {
				rows = append(rows, []string{"Error", task.ErrorInfo})
			}
			PrintTable(cmd.OutOrStdout(), []string{"Field", "Value"}, rows)
			logs, err := cli.GetAgentTaskLogs(ctx, wsID, taskID)
			if err == nil && len(logs) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\nRecent logs:\n")
				start := 0
				if len(logs) > 10 {
					start = len(logs) - 10
				}
				for _, l := range logs[start:] {
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s %s\n", l.CreatedAt.Format("15:04:05"), l.Level, l.Message)
				}
			}
			return nil
		},
	}
}
```

`sdk/cli/cmd_ask.go`：

```go
package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reqmango/tools/client"
)

func newAskCmd() *cobra.Command {
	var issueArg string
	cmd := &cobra.Command{
		Use:   "ask <question>",
		Short: "Ask the project AI assistant (long-running, stream aggregated)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, cfg, err := setup(cmd)
			if err != nil {
				return err
			}
			question := ""
			for i, part := range args {
				if i > 0 {
					question += " "
				}
				question += part
			}

			projectID, err := resolveProject(cmd, cfg)
			if err != nil {
				return err
			}
			if issueArg != "" {
				wsID, wsErr := resolveWorkspace(cmd, cfg)
				if wsErr != nil {
					return wsErr
				}
				id, err := resolveIssueID(context.Background(), cli, wsID, issueArg)
				if err != nil {
					return err
				}
				iss, err := cli.GetIssue(context.Background(), id)
				if err != nil {
					return err
				}
				projectID = iss.ProjectID
				question = fmt.Sprintf("(About issue %s) %s", issueArg, question)
			}

			reply, err := cli.AIChat(context.Background(), projectID, question, nil)
			if err != nil {
				if apiErr := client.AsAPIError(err); apiErr != nil && apiErr.StatusCode == 401 {
					return fmt.Errorf("PAT is invalid or revoked - run `reqmango auth login`")
				}
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), reply.Text)
			if reply.ThreadID != 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\n(thread %d)\n", reply.ThreadID)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&issueArg, "issue", "", "ask about a specific issue (id or code)")
	return cmd
}
```

在 `sdk/cli/root.go` 的 `NewRootCommand` 中，`root.AddCommand(newWorkspaceCmd())` 之后加：

```go
	root.AddCommand(newProjectCmd())
	root.AddCommand(newIssueCmd())
	root.AddCommand(newCycleCmd())
	root.AddCommand(newMetaCmd())
	root.AddCommand(newAgentCmd())
	root.AddCommand(newAskCmd())
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd sdk && go test ./cli/ -v && go vet ./...`
Expected: PASS（若 `TestAuthLogin`/`TestWorkspaceSwitch` 与新增命令冲突则排查 flag 名；无冲突）

- [ ] **Step 5: Commit**

```bash
git add sdk/cli/cmd_issue.go sdk/cli/cmd_project.go sdk/cli/cmd_cycle.go sdk/cli/cmd_meta.go sdk/cli/cmd_agent.go sdk/cli/cmd_ask.go sdk/cli/cmd_issue_test.go sdk/cli/cmd_project_test.go sdk/cli/cmd_cycle_test.go sdk/cli/cmd_agent_test.go sdk/cli/root.go
git commit -m "feat(cli): add issue/project/cycle/meta/agent/ask commands"
```

---

## Phase E：e2e + 文档 + 收尾

### Task 15: e2e 冒烟 + README + Makefile + 删除旧 mcp-server

**Files:**
- Create: `scripts/e2e_tools.sh`
- Create: `sdk/README.md`
- Modify: `Makefile`（tools/test-tools 目标 + .PHONY）
- Modify: `docs/kb/architecture/project-layout.md`（mcp-server/ → sdk/ 布局）
- Delete: `mcp-server/`（`git rm -r`）

**Interfaces:**
- Consumes: 全部前序任务的产出（两个二进制、PAT、24 tools）

- [ ] **Step 1: 写 e2e 脚本**

`scripts/e2e_tools.sh`：

```bash
#!/usr/bin/env bash
# reqmango tools e2e smoke test.
# Prerequisites: backend running (make dev-backend), database migrated.
# Usage: bash scripts/e2e_tools.sh [api_base_url]
set -euo pipefail

API="${1:-http://localhost:8000/api/v1}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$ROOT/bin"
mkdir -p "$BIN"

echo "==> building tools binaries"
(cd "$ROOT/sdk" && go build -o "$BIN/reqmango" ./cmd/reqmango && go build -o "$BIN/reqmango-mcp" ./cmd/reqmango-mcp)

EMAIL="e2e-$(date +%s)@example.com"
PASSWORD="e2e-pass-123"

echo "==> registering user $EMAIL"
curl -sf -X POST "$API/auth/register" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"username\":\"e2e$(date +%s)\",\"password\":\"$PASSWORD\"}" >/dev/null

echo "==> reqmango auth login"
CONFIG="$(mktemp -d)/config.json"
"$BIN/reqmango" --config "$CONFIG" auth login --api-url "$API" \
  --email "$EMAIL" --password "$PASSWORD" >/dev/null
PAT="$(sed -n 's/.*"pat": "\([^"]*\)".*/\1/p' "$CONFIG")"
[ -n "$PAT" ] || { echo "FAIL: no PAT in config"; exit 1; }

export REQMANGO_API_URL="$API"
export REQMANGO_PAT="$PAT"

echo "==> workspace discovery + switch"
WS_ID="$("$BIN/reqmango" --config "$CONFIG" workspace list --output json | grep -o '"id": [0-9]*' | head -1 | grep -o '[0-9]*')"
[ -n "$WS_ID" ] || { echo "FAIL: no workspace found (create one in the UI first)"; exit 1; }
"$BIN/reqmango" --config "$CONFIG" workspace switch "$WS_ID" >/dev/null

echo "==> project list"
PROJ_ID="$("$BIN/reqmango" --config "$CONFIG" project list --output json | grep -o '"id": [0-9]*' | head -1 | grep -o '[0-9]*')"
[ -n "$PROJ_ID" ] || { echo "FAIL: no project found (create one in the UI first)"; exit 1; }
IDENT="$("$BIN/reqmango" --config "$CONFIG" project list --output json | grep -o '"identifier": "[^"]*"' | head -1 | sed 's/.*"identifier": "\([^"]*\)"/\1/')"

echo "==> issue lifecycle (create/show by code/update)"
CREATE_OUT="$("$BIN/reqmango" --config "$CONFIG" issue create --project "$PROJ_ID" --title "e2e smoke $(date +%s)" --priority medium --output json)"
ISSUE_ID="$(echo "$CREATE_OUT" | grep -o '"id": [0-9]*' | head -1 | grep -o '[0-9]*')"
SEQ="$(echo "$CREATE_OUT" | grep -o '"sequence_id": [0-9]*' | head -1 | grep -o '[0-9]*')"
[ -n "$ISSUE_ID" ] || { echo "FAIL: issue create"; exit 1; }
"$BIN/reqmango" --config "$CONFIG" issue show "$IDENT-$SEQ" >/dev/null
"$BIN/reqmango" --config "$CONFIG" issue update "$ISSUE_ID" --priority high >/dev/null
"$BIN/reqmango" --config "$CONFIG" issue list --project "$PROJ_ID" --limit 5 >/dev/null
echo "   issue $IDENT-$SEQ (id $ISSUE_ID) OK"

echo "==> mcp stdio handshake (initialize + initialized + tools/list)"
STDIO_OUT=$(printf '%s\n%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2026-07-28","clientInfo":{"name":"e2e","version":"1.0.0"},"capabilities":{}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  | "$BIN/reqmango-mcp")
echo "$STDIO_OUT" | grep -q '"tools"' || { echo "FAIL: stdio tools/list"; echo "$STDIO_OUT"; exit 1; }
# 只统计 tools/list 响应行的工具名（initialize 响应的 serverInfo.name 不含 "tools" 关键字，会被过滤）
TOOL_COUNT="$(echo "$STDIO_OUT" | grep '"tools"' | grep -o '"name":"[a-z_]*"' | wc -l)"
[ "$TOOL_COUNT" -ge 24 ] || { echo "FAIL: expected >=24 tools over stdio, got $TOOL_COUNT"; exit 1; }
echo "   stdio OK ($TOOL_COUNT tools)"

echo "==> mcp streamable HTTP smoke"
"$BIN/reqmango-mcp" --http :18080 >/dev/null 2>&1 &
MCP_PID=$!
trap 'kill $MCP_PID 2>/dev/null || true' EXIT
sleep 1
# HTTP 模式需要 Authorization: Bearer <PAT>（spec §5.1）
AUTH_HEADER="Authorization: Bearer $PAT"
# 先验证无凭据被拒
if curl -s -o /dev/null -w '%{http_code}' -X POST "http://localhost:18080/mcp" \
  -H 'Content-Type: application/json' | grep -q 200; then
  echo "FAIL: HTTP endpoint accepted a request without Bearer token"; exit 1
fi
SID=$(curl -sf -X POST "http://localhost:18080/mcp" -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' -H "$AUTH_HEADER" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2026-07-28","clientInfo":{"name":"e2e","version":"1.0.0"},"capabilities":{}}}' \
  -D - -o /dev/null | tr -d '\r' | sed -n 's/[Mm]cp-Session-Id: //p')
[ -n "$SID" ] || { echo "FAIL: no session id from HTTP initialize"; exit 1; }
curl -sf -X POST "http://localhost:18080/mcp" -H 'Content-Type: application/json' \
  -H "Mcp-Session-Id: $SID" -H "$AUTH_HEADER" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_workspaces","arguments":{}}}' \
  | grep -q '"result"' || { echo "FAIL: list_workspaces over HTTP"; exit 1; }
echo "   HTTP OK"

echo "ALL E2E SMOKE TESTS PASSED"
```

- [ ] **Step 2: 写 sdk/README.md**

`sdk/README.md`：

````markdown
# reqmango tools

reqmango 的 MCP server 与 CLI，共享同一个 API 客户端（`client/`）。

```
sdk/
  client/   ← 共享 API 客户端（仅标准库）+ 全部 DTO + 错误映射
  mcp/      ← MCP server（mark3labs/mcp-go）：24 个工具
  cli/      ← reqmango CLI（cobra）
  cmd/reqmango-mcp   ← MCP server 二进制（stdio + streamable HTTP）
  cmd/reqmango       ← CLI 二进制
```

## 构建

```bash
cd sdk
go build -o ../bin/reqmango ./cmd/reqmango
go build -o ../bin/reqmango-mcp ./cmd/reqmango-mcp
```

或 `make tools`（仓库根目录）。

## 认证（PAT）

两种工具都使用 PAT（Personal Access Token）认证：

```bash
# 1. 登录（交换 JWT → 创建 PAT → 存 ~/.reqmango/config.json，JWT 即弃）
reqmango auth login
# 2. 查看 / 吊销
reqmango auth status
reqmango auth revoke --list
reqmango auth revoke <pat-id>
```

环境变量：`REQMANGO_API_URL`（默认 `http://localhost:8000/api/v1`）、`REQMANGO_PAT`。

## MCP server

### Claude Code

```bash
claude mcp add reqmango \
  --env REQMANGO_PAT=reqmango_pat_xxx \
  --env REQMANGO_API_URL=http://localhost:8000/api/v1 \
  -- /path/to/bin/reqmango-mcp
```

### Claude Desktop

`claude_desktop_config.json`：

```json
{
  "mcpServers": {
    "reqmango": {
      "command": "/path/to/bin/reqmango-mcp",
      "env": {
        "REQMANGO_PAT": "reqmango_pat_xxx",
        "REQMANGO_API_URL": "http://localhost:8000/api/v1"
      }
    }
  }
}
```

### Cursor

Cursor Settings → MCP → Add new MCP server：

- Type: `command`
- Command: `/path/to/bin/reqmango-mcp`
- Env: `REQMANGO_PAT=reqmango_pat_xxx`

### HTTP 模式（远程 / CI）

```bash
reqmango-mcp --http :8080
# 端点 POST http://host:8080/mcp（streamable HTTP，先 initialize 拿 Mcp-Session-Id）
# 每个请求必须带 Authorization: Bearer <PAT>，否则 401
```

## 工具清单（24 个）

**核心（19）**：`list_workspaces` `list_projects` `get_project` `create_issue` `list_issues` `get_issue` `update_issue` `search_issues` `add_comment` `list_cycles` `get_cycle_progress` `add_issue_to_cycle` `list_members` `get_states` `get_labels` `list_issue_types` `list_notifications` `list_pages` `get_page`

**AI（5）**：`ai_search` `ai_chat` `list_agents` `dispatch_agent` `get_agent_task`

工具错误以 `isError: true` 的结构化结果返回，不会中断 MCP 会话；401 会附带 `reqmango auth login` 的修复提示。

## CLI

```bash
reqmango auth login|logout|status|revoke
reqmango workspace list|switch <id>
reqmango project list|show <id|identifier>|create
reqmango issue list|show <id|code>|create|update <id|code>|search <query>
reqmango cycle list|progress <id>|burndown <id>
reqmango meta states|labels|issue-types
reqmango agent list|dispatch <agentId> "task..."|task <taskId>
reqmango ask "..." [--issue <id|code>]
```

- 全局 flag：`--workspace` `--project`（记忆在 config，可覆盖）、`--output table|json`
- issue 支持 `DEMO-42` 格式定位（project identifier + sequence_id），也支持数字 ID
- `--assignee me` 表示当前用户

## 测试

```bash
cd sdk && go test ./...          # 单元测试
bash ../scripts/e2e_tools.sh     # e2e 冒烟（需后端在跑）
```
````

- [ ] **Step 3: Makefile + 文档更新 + 删除旧目录**

`Makefile` 的 `.PHONY` 行改为：

```makefile
.PHONY: up down build logs restart clean dev dev-backend db-shell lint lint-fix test test-backend test-frontend ci coverage tools test-tools
```

在 `coverage:` 目标之后添加：

```makefile
# ======== Tools (MCP + CLI) ========

tools:
	cd sdk && go build -o ../bin/reqmango ./cmd/reqmango
	cd sdk && go build -o ../bin/reqmango-mcp ./cmd/reqmango-mcp

test-tools:
	cd sdk && go test ./...
```

`docs/kb/architecture/project-layout.md` 第 30 行附近：把

```
├── mcp-server/              # 独立 MCP Server（Go module）
```

改为

```
├── sdk/                     # MCP server + CLI 共享模块（github.com/reqmango/tools）
│   ├── client/              #   共享 API 客户端 + DTO + 错误映射
│   ├── mcp/                 #   MCP server（mcp-go，24 tools，stdio + HTTP）
│   ├── cli/                 #   reqmango CLI（cobra）
│   └── cmd/                 #   两个二进制入口
```

并同步修正第 84 行 `mcp-server/main.go` 的表格行（改为指向 `sdk/cmd/reqmango-mcp/main.go`）。

删除旧目录：

```bash
git rm -r mcp-server/
```

检查无其他引用后继续：

```bash
grep -rn "mcp-server" docs/ Makefile README.md 2>/dev/null || true
```

（`docs/reqmango-catchup-tasks.md` 是历史任务记录，保留不动。）

- [ ] **Step 4: 全量验证**

Run: `cd sdk && go test ./... && go vet ./... && cd ../backend && go test ./internal/...`
Expected: PASS

手工冒烟（后端在跑时）:

Run: `bash scripts/e2e_tools.sh`
Expected: 输出 `ALL E2E SMOKE TESTS PASSED`

- [ ] **Step 5: Commit**

```bash
git add scripts/e2e_tools.sh sdk/README.md Makefile docs/kb/architecture/project-layout.md
git commit -m "chore: e2e smoke script, sdk README, Makefile tools targets; remove legacy mcp-server"
```

---

## 自审记录

（写入后执行了 writing-plans 自审，逐条如下）

**1. Spec 覆盖：**

| spec 章节 | 对应任务 |
|---|---|
| §2 架构（三包两二进制、共享 client） | Task 5（module + client 核心）、Task 10（mcp 二进制）、Task 13（cli 二进制） |
| §3 模块结构 + 删除旧 mcp-server + Makefile + 环境变量 | Task 15；REQMANGO_API_URL/REQMANGO_PAT 贯穿 Task 10/13/15 |
| §4.1 personal_access_tokens 表 | Task 1（逐字段对应） |
| §4.2 三端点 + AuthMiddleware | Task 3（handler+路由）、Task 4（中间件） |
| §4.2 CLI 登录流（JWT→PAT→config，JWT 即弃） | Task 13（auth login 实现） |
| §5.1 stdio + streamable HTTP | Task 10（双传输入口） |
| §5.2 核心 19 tools | Task 11（19 个一一对应；get_issue 通过合并 ListComments 满足"含评论"） |
| §5.2 AI 5 tools | Task 12（一一对应；ai_chat 采用 project AI chat 聚合方案） |
| §6 CLI 命令树 | Task 13（auth/workspace）+ Task 14（project/issue/cycle/meta/agent/ask） |
| §6 DEMO-42 定位 | Task 7（ResolveIssueCode）+ Task 11/14 复用 |
| §6 配置 ~/.reqmango/config.json | Task 13 |
| §7 错误模型 / 401 提示 / 超时 / 不中断会话 | Task 5（APIError）、Task 10/11（toolAPIError 401 提示）、Task 9（AIChat/Dispatch 5min）、Task 11（isError 结果） |
| §8 测试矩阵 | Task 2-4（sqlmock）、Task 5-9（httptest）、Task 10-12（servertest 全链路）、Task 13-14（cobra Execute）、Task 15（e2e 脚本） |
| §9 交付顺序 | Phase A→E 即对应 5 步顺序 |

**2. 与 spec 的有意偏差（执行时按计划为准）：**
- spec §7 写错误体为 `{error:{code,message}}`，实际后端统一为 `{"message":"..."}`（已由三路 explore 确认）。计划的 APIError/错误解析按实际格式实现，spec §7 的"格式"以实际后端为准。
- `create_issue` 工具新增 workspace_id 必填参数（后端 POST /issues 强制要求 project_id+workspace_id 两个 query 参数）。
- spec 工具表里 get_issue 的"relations"未实现（IssueResponse 无 relations 字段，属 spec 笔误），以评论合成为准。

**3. 占位符扫描：** 无 TBD/TODO；所有代码块为完整实现；唯一"留待确认"点（`common.AppError` 方法名 `Internal/NotFound/Unauthorized`）已由 explore 确认存在。

**4. 类型一致性：** 已核对 —— `client.New`/`GetJSON`/`PostJSON`/`PutJSON`/`DeleteJSON`/`APIError`/`AsAPIError`（Task 5）→ 各资源方法（Task 6-9）→ MCP 工具（Task 10-12）→ CLI 命令（Task 13-14）签名一致；`IssueListOptions.Query()`/`ResolveIssueCode`/`parseIDList`/`resolveIssueArg`/`uintPtr64` 在各调用点匹配；测试辅助 `backend`/`newCLI`/`mcpPost`/`newTestMCPServer`/`callTool`/`listToolNames` 在同包测试文件间共享。

**5. 自审发现并已修复的问题：**
- Task 11 的工具数断言原为 `== 19`，与 Task 12 注册后总数 24 互斥 → 改为 `>= 19`（含 Task 11 Step 3 对 Task 10 测试的同步修改说明）。
- spec §5.1 要求 HTTP 模式 `Authorization: Bearer <PAT>` 认证，初稿遗漏 → 新增 `sdk/mcp/http.go` 的 `BearerAuth` 中间件 + 单测，main.go HTTP 分支接入，e2e 脚本与 README 同步更新（e2e 先断言无凭据 401，再带 PAT 走完整握手）。
- e2e stdio 工具计数会把 initialize 响应中 serverInfo 的 `"name":"reqmango-mcp"` 误计为工具 → 改为先 `grep '"tools"'` 只统计 tools/list 响应行；并补上协议规范的 `notifications/initialized` 消息（mcp-go v1.0.0 的 initialize 即置位 initialized，不补也能工作，补上更贴近规范）。
- 核实 mcp-go v1.0.0：`server.go:1414` 的 `Initialized()` 门槛由 initialize 请求本身置位（`stdio.go:138`），测试与 e2e 的 initialize → tools/call 顺序无需额外握手消息。
