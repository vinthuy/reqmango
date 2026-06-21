# Go Backend Architecture（Go 后端架构）

**最后更新**: 2026-06-21

---

## 项目路径

```
backend-go/
├── cmd/server/main.go              # 入口
├── internal/
│   ├── config/config.go            # 配置加载（Viper）
│   ├── common/
│   │   ├── constants.go            # 角色、优先级、状态组常量
│   │   ├── errors.go              # AppError 类型
│   │   └── pagination.go          # 分页辅助
│   ├── model/                      # GORM 模型（10 个文件）
│   │   ├── base.go                # BaseModel 嵌入结构
│   │   ├── user.go                # User
│   │   ├── workspace.go           # Workspace, WorkspaceMember
│   │   ├── project.go             # Project, ProjectMember
│   │   ├── issue.go               # Issue, IssueAssignee, IssueLabel, IssueActivity
│   │   ├── state.go               # State, StateTransition
│   │   ├── label.go               # Label
│   │   ├── cycle.go               # Cycle
│   │   └── module.go              # Module, ModuleIssue
│   ├── dto/
│   │   ├── request/               # 请求 DTO（8 个文件）
│   │   └── response/              # 响应 DTO（8 个文件）
│   ├── service/                    # 业务逻辑（7 个文件）
│   │   ├── auth_service.go
│   │   ├── workspace_service.go
│   │   ├── project_service.go
│   │   ├── issue_service.go
│   │   ├── cycle_service.go
│   │   ├── module_service.go
│   │   └── project_settings_service.go
│   ├── handler/                    # HTTP Handler（7 个文件）
│   ├── middleware/                  # 中间件（Auth, CORS, Logger）
│   └── router/router.go           # 路由注册
├── config/config.yaml              # 默认配置
├── go.mod
└── go.sum
```

## 分层架构

```
Router ──→ Middleware ──→ Handler ──→ Service ──→ Model (GORM)
                               │            │
                               │            └── DB 操作
                               └── 参数绑定 + HTTP 响应
```

### Handler 层

- 使用 Gin Context 绑定请求参数（`c.ShouldBindJSON`, `c.Param`, `c.Query`）
- 调用 Service 层方法
- 返回统一格式的 JSON 响应
- **不包含任何业务逻辑**

### Service 层

- 纯 Go 业务逻辑
- 使用 GORM 进行数据库操作
- 返回自定义 `AppError` 类型（含 HTTP 状态码）
- 跨 Model 操作（如 Issue 关联 Assignee、Label）

### Model 层

- GORM 标签定义表结构
- 所有模型嵌入 `BaseModel`（ID uint64, CreatedAt, UpdatedAt, DeletedAt）
- `DeletedAt gorm.DeletedAt` 实现软删除

### DTO 层

- `request/` — 客户端请求结构体，带 `binding` 验证标签
- `response/` — 服务端响应结构体
- 与 Model 分离，避免暴露数据库字段

## 核心组件

### AppError（统一错误处理）

```go
type AppError struct {
    StatusCode int
    Message    string
    Err        error
}
```

预定义错误工厂：`NotFound()`, `Conflict()`, `Unauthorized()`, `Forbidden()`, `Validation()`, `Internal()`, `BadRequest()`

### Pagination

```go
func ParsePagination(c *gin.Context) (limit int, offset int)
```

默认 limit=20, max=100，从 query params `page` 和 `page_size` 解析。

### 认证流程

1. `POST /api/v1/auth/login` — bcrypt 验证密码，返回 JWT access token
2. 后续请求在 `Authorization: Bearer <token>` 头中携带 JWT
3. Auth 中间件解析 JWT，查询用户，注入 `c.Set("currentUser", user)`
4. Handler 通过 `c.MustGet("currentUser").(*model.User)` 获取

## 已实现模块

| 模块 | Service | Handler | 关键功能 |
|------|---------|---------|----------|
| Auth | `auth_service.go` | `auth_handler.go` | Register, Login, GetCurrentUser |
| Workspace | `workspace_service.go` | `workspace_handler.go` | CRUD + 成员管理 |
| Project | `project_service.go` | `project_handler.go` | CRUD + 成员 + 统计 + 归档 |
| Issue | `issue_service.go` | `issue_handler.go` | CRUD + 搜索 + 批量 + 活动 + 关联 |
| Cycle | `cycle_service.go` | `cycle_handler.go` | CRUD + 状态流转 + 进度 + 燃尽图 |
| Module | `module_service.go` | `module_handler.go` | CRUD + 树形 + Issue 关联 + 统计 |
| State/Label | `project_settings_service.go` | `project_settings_handler.go` | CRUD + 默认值 |

## 配置

通过环境变量配置，Viper 自动加载 `.env` 文件：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DATABASE_URL` | PostgreSQL 连接串 | — |
| `SECRET_KEY` | JWT 签名密钥 | — |
| `PORT` | 监听端口 | 8080 |
| `DEBUG` | 调试模式 | false |
