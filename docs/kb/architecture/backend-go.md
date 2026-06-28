# Go Backend Architecture（Go 后端架构）

**最后更新**: 2026-06-28

---

## 项目路径

```
backend/
├── cmd/server/main.go              # 入口
├── internal/
│   ├── config/config.go            # 配置加载（Viper + 环境变量）
│   ├── common/
│   │   ├── constants.go            # 角色、优先级、状态组常量
│   │   ├── error_codes.go          # 错误码定义（含 i18n 消息）
│   │   ├── errors.go              # AppError 类型
│   │   └── pagination.go          # 分页辅助
│   ├── model/                      # GORM 模型（36 个文件）
│   │   ├── base.go                # BaseModel 嵌入结构（ID/CreatedAt/UpdatedAt/DeletedAt）
│   │   ├── user.go                # User（含 IsSuperuser、GetID、IsSuper 方法）
│   │   ├── workspace.go           # Workspace, WorkspaceMember
│   │   ├── project.go             # Project, ProjectMember
│   │   ├── issue.go               # Issue, IssueAssignee, IssueLabel, IssueActivity
│   │   ├── issue_type.go          # IssueType + 字段绑定
│   │   ├── issue_type_template.go # Workspace 级问题类型模板
│   │   ├── state.go               # State, StateTransition
│   │   ├── label.go               # Label
│   │   ├── cycle.go               # Cycle
│   │   ├── module.go              # Module, ModuleIssue
│   │   ├── page.go                # Page（层级文档）
│   │   ├── comment.go             # Comment（线程回复）
│   │   ├── relation.go            # RelationType, IssueRelation
│   │   ├── workflow.go            # Workflow + 规则（Allow/Approval）
│   │   ├── attachment.go          # Attachment
│   │   ├── estimate.go            # EstimatePoint
│   │   ├── time_track.go          # TimeTrack
│   │   ├── recurrence_rule.go     # Recurrence（周期性问题）
│   │   ├── notification.go        # Notification（9 种类型）
│   │   ├── webhook.go             # Webhook 配置
│   │   ├── slack.go               # Slack 集成
│   │   ├── github.go              # GitHub 集成
│   │   ├── agent.go               # AI Agent
│   │   ├── ai_config.go           # AI 配置（用户级）
│   │   ├── mcp.go                 # MCP Server 配置
│   │   ├── automation.go          # AutomationRule（自动规则）
│   │   ├── saved_view.go          # SavedView（视图预设）
│   │   ├── project_page_tab.go    # 项目页面 Tab 配置
│   │   ├── project_update.go      # ProjectUpdate（状态时间线）
│   │   ├── project_template.go    # ProjectTemplate
│   │   ├── work_item_template.go  # WorkItemTemplate
│   │   ├── custom_field.go        # CustomField + 7 种类型
│   │   ├── conditional_field.go   # ConditionalField（条件可见性）
│   │   ├── initiative.go          # Initiative（跨项目战略目标）
│   │   ├── release.go             # Release + Roadmap
│   │   ├── role.go                # Role + RoleLevel（Admin=20/Member=15/Guest=5）
│   │   └── permission.go          # Permission + RolePermission 关联表
│   ├── dto/
│   │   ├── request/               # 请求 DTO（25 个文件）
│   │   │   ├── auth.go, workspace.go, project.go
│   │   │   ├── issue.go, cycle.go, module.go, page.go
│   │   │   ├── custom_field.go, conditional_field.go
│   │   │   ├── workflow.go, state.go, label.go
│   │   │   ├── comment.go, relation.go, attachment.go
│   │   │   ├── notification.go, webhook.go
│   │   │   ├── estimate.go, time_track.go, recurrence.go
│   │   │   ├── project_template.go, work_item_template.go
│   │   │   ├── type_template.go, issue_type.go
│   │   │   ├── initiative.go, release.go, saved_view.go
│   │   │   ├── role.go
│   │   │   └── rql.go
│   │   └── response/              # 响应 DTO（22 个文件）
│   │       └── 对应 request 文件的响应结构体
│   ├── service/                    # 业务逻辑（36 个文件）
│   │   ├── auth_service.go        # JWT 签发与验证
│   │   ├── workspace_service.go   # Workspace CRUD + 成员
│   │   ├── project_service.go     # Project CRUD + 成员 + 统计 + 归档
│   │   ├── issue_service.go       # Issue CRUD + 搜索 + 批量 + 导入/导出 + 关联
│   │   ├── cycle_service.go       # Cycle CRUD + 状态流转 + 进度 + 燃尽图
│   │   ├── module_service.go      # Module CRUD + 树形 + Issue 关联 + 统计
│   │   ├── page_service.go        # Page CRUD + 树形 + 归档
│   │   ├── comment_service.go     # Comment CRUD + 回复 + 解决
│   │   ├── relation_service.go    # RelationType CRUD + Issue 关联管理
│   │   ├── workflow_service.go    # Workflow CRUD + 转换验证
│   │   ├── attachment_service.go  # Attachment CRUD
│   │   ├── estimate_service.go    # 3 种估算模式管理
│   │   ├── time_track_service.go  # 工时记录
│   │   ├── recurrence_service.go  # 周期性 Issue 生成
│   │   ├── notification_service.go # 9 种通知 + 标记已读 + 提醒
│   │   ├── webhook_service.go     # Webhook 发送 + HMAC-SHA256 签名
│   │   ├── slack_service.go       # Slack 通知格式化 + 发送
│   │   ├── github_service.go      # GitHub Issues 同步 + Webhook 接收
│   │   ├── ai_service.go          # AI 对话/搜索/创建/分析/图表/分诊
│   │   ├── agent_service.go       # AI Agent CRUD + Dispatch/Triage/Assign
│   │   ├── mcp_service.go         # MCP Server 配置 + 连接管理
│   │   ├── automation_service.go  # 自动化规则 CRUD + 触发执行
│   │   ├── saved_view_service.go  # 视图预设管理
│   │   ├── custom_field_service.go # 自定义字段 7 种类型 + 条件规则
│   │   ├── conditional_field_service.go # 条件可见性评估
│   │   ├── project_settings_service.go # State/Label CRUD + 默认值
│   │   ├── project_page_tab_service.go  # 页面 Tab 配置
│   │   ├── project_update_service.go    # 项目状态更新
│   │   ├── project_template_service.go  # 项目模板
│   │   ├── work_item_template_service.go # 工作项模板
│   │   ├── issue_type_service.go        # 问题类型 CRUD
│   │   ├── type_template_service.go     # Workspace 类型模板
│   │   ├── initiative_service.go        # Initiative CRUD
│   │   ├── release_service.go    # Release + Roadmap
│   │   ├── role_service.go       # RBAC 角色 + 权限查询（8 方法）
│   │   ├── report_service.go     # 报表生成 + 图表数据
│   │   ├── sse_hub.go            # SSE 实时事件中心
│   │   └── llm_client.go         # LLM 客户端（DeepSeek/Anthropic）
│   ├── handler/                    # HTTP Handler（38 个文件）
│   │   └── 每个 service 对应一个 handler + sse_handler.go + role_handler.go
│   ├── rql/                        # RQL 查询语言引擎（9 文件）
│   │   ├── lexer.go               # 词法分析（tokenize）
│   │   ├── parser.go              # 语法分析（AST 构建）
│   │   ├── ast.go                 # 抽象语法树类型定义
│   │   ├── builder.go             # AST → SQL/GORM 查询
│   │   ├── executor.go            # 查询执行器
│   │   ├── handler.go             # HTTP 端点 `/api/v1/rql/search`
│   │   ├── service.go             # 查询服务
│   │   └── *_test.go              # 3 个测试文件
│   ├── middleware/                  # 中间件（6 个）
│   │   ├── auth.go                # JWT 认证：解析 Bearer token → 注入 currentUser
│   │   ├── authorization.go       # RBAC 鉴权：RequirePermission / RequireRoleLevel
│   │   ├── cors.go                # CORS 策略
│   │   ├── language.go            # i18n 语言检测（Header/Query）
│   │   ├── logger.go              # 请求日志
│   │   └── rate_limiter.go        # Token Bucket 限流（默认 500 req/min，0=禁用）
│   ├── i18n/
│   │   ├── i18n.go / i18n_test.go # 国际化逻辑
│   │   ├── messages_en.json       # 英文错误消息
│   │   └── messages_zh.json       # 中文错误消息
│   └── seed/
│       ├── seed.go                # 种子数据（20 用户 + 100 Issues + Workspace/Project）
│       └── seed_rbac.go           # RBAC 种子（55 权限 + 3 默认角色）

---

## 分层架构

```
Router ──→ Middleware Chain ──→ Handler ──→ Service ──→ Model (GORM)
    │           │                    │            │
    │     Auth/CORS/Lang             │            └── DB 操作
    │     Logger/RateLimit           └── DTO 绑定 + HTTP 响应
    └── Gin 路由树（80+ 端点）
```

### Handler 层
- 使用 Gin Context 绑定请求参数（`c.ShouldBindJSON`, `c.Param`, `c.Query`）
- 调用 Service 层方法
- 返回统一格式的 JSON 响应
- **不包含任何业务逻辑**

### Service 层
- 纯 Go 业务逻辑
- 使用 GORM 进行数据库操作
- 返回自定义 `AppError` 类型（含 HTTP 状态码和 i18n 错误码）
- 跨 Model 操作（如 Issue 关联 Assignee、Label、Attachment）
- 注入依赖（Notification、Webhook、Automation、Slack 等服务）

### Model 层
- GORM 标签定义表结构 + 索引
- 所有模型嵌入 `BaseModel`（ID uint64, CreatedAt, UpdatedAt, DeletedAt）
- `DeletedAt gorm.DeletedAt` 实现软删除
- 关联关系通过 GORM Tag + Preload 管理

### DTO 层
- `request/` — 客户端请求结构体，带 `binding` 验证标签
- `response/` — 服务端响应结构体
- 与 Model 分离，避免暴露数据库内部字段

### RQL 层
- 自研 RQL（reqmango Query Language）查询语言
- 完整的词法分析 → 语法分析 → AST → SQL/GORM 生成流水线
- 端点：`POST /api/v1/rql/search`
- 功能对标 Plane PQL，社区版免费可用

---

## 核心组件

### AppError（统一错误处理）

```go
type AppError struct {
    StatusCode int
    ErrorCode  string  // RQL 错误码（含 i18n 支持）
    Message    string
    Err        error
}
```

预定义错误工厂：`NotFound()`, `Conflict()`, `Unauthorized()`, `Forbidden()`, `Validation()`, `Internal()`, `BadRequest()`

### i18n 错误消息
- `messages_en.json` / `messages_zh.json` 双语言
- `ErrorCode` 映射到具体错误消息
- 语言通过 `Accept-Language` Header 或 `lang` Query 参数检测

### Pagination

```go
func ParsePagination(c *gin.Context) (limit int, offset int)
```

默认 limit=20, max=100，从 query params `page` 和 `page_size` 解析。

### 认证流程

1. `POST /api/v1/auth/login` — bcrypt 验证密码，返回 JWT + token_type + expires_at
2. 后续请求在 `Authorization: Bearer <token>` 头中携带 JWT
3. Auth 中间件解析 JWT，查询用户，注入 `c.Set("currentUser", user)`
4. Handler 通过 `c.MustGet("currentUser").(*model.User)` 获取

### RBAC 鉴权

- `authorization.go` 提供 `RequirePermission(db, permCode)` 和 `RequireRoleLevel(db, minLevel)` 两个中间件
- 超级用户（IsSuperuser=true）自动旁路所有权限检查
- 权限格式：`resource:action`（如 `issue:create`, `workspace:manage`）
- 角色级别：Admin=20（全部权限）、Member=15（创建编辑）、Guest=5（只读）
- 支持 workspace-level 和 project-level 自定义角色
- 5 个 RBAC API：`GET/POST /api/v1/workspaces/:wsParam/roles`、`PUT/DELETE /api/v1/workspaces/:wsParam/roles/:id`、`GET /api/v1/permissions`

### 限流

- Token Bucket 算法，按用户 ID 或 IP 分组
- 默认 500 req/min（可通过 `RATE_LIMIT_REQUESTS` 环境变量配置）
- `limit=0` 时完全禁用（开发/CI 环境）
- 响应头：`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

### SSE 实时通知

- `sse_hub.go` — 广播 Hub 实现
- `sse_handler.go` — SSE 端点 `/api/v1/sse`
- 9 种通知类型通过 SSE 实时推送到前端

---

## API 路由总览

所有路由前缀 `/api/v1`，80+ 端点：

| 路由组 | 资源 |
|--------|------|
| `/auth/` | 注册 / 登录 / 当前用户 |
| `/workspaces/` | CRUD + 成员 + AI Agent + MCP + GitHub + Slack 配置 |
| `/projects/` | CRUD + 成员 + 统计 + 报表 + Webhook + 自动化 |
| `/projects/:id/cycles` | Cycle CRUD + 状态流转 + 进度 + 燃尽图 |
| `/projects/:id/modules` | Module CRUD + 树形 + Issue 关联 + 统计 |
| `/projects/:id/pages` | Page CRUD + 归档 |
| `/projects/:id/issue-types` | 问题类型 CRUD + 字段绑定 |
| `/projects/:id/workflows` | Workflow CRUD + 规则 |
| `/projects/:id/releases` | Release + Roadmap |
| `/projects/:id/ai/` | AI 聊天/搜索/图表/分诊 |
| `/projects/:id/automations` | 自动化规则 CRUD + 执行 |
| `/issues/` | CRUD + 搜索 + 批量 + 导入/导出 + 树 + 关联 + 工时 + 周期性 |
| `/comments/` | CRUD + 回复 + 解决 |
| `/custom-fields/` | CRUD + 选项 + 条件规则 |
| `/templates/` | 项目模板 + 类型模板 + 工作项模板 |
| `/rql/search` | RQL 查询执行 |
| `/notifications/` | 列表 + 已读 + 提醒 |
| `/webhook/github/:id` | GitHub Webhook 接收（公开端点） |
| `/intake/:projectId` | 公开问题提交（无需认证） |
| `/sse` | 实时事件推送 |
| `/roles/` | RBAC 角色 CRUD + 权限列表 |
| `/permissions/` | 全局权限枚举查询 |

---

## 配置

通过环境变量 + `config.yaml` 配置：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DATABASE_URL` | PostgreSQL DSN | — |
| `SECRET_KEY` | JWT 签名密钥 | — |
| `PORT` | 监听端口 | 8080 |
| `RATE_LIMIT_REQUESTS` | 限流阈值（req/min） | 500 |
| `RATE_LIMIT_WINDOW_SEC` | 限流窗口 | 60 |
| `AI_PROVIDER` | AI 提供商 | deepseek |
| `AI_MODEL` | AI 模型 | deepseek-chat |
| `AI_BASE_URL` | AI API 地址 | api.deepseek.com |
| `DEBUG` | 调试模式 | false |
| `LANGUAGE` | 默认语言 | zh |
