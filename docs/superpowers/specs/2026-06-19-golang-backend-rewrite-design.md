# Go 后端重构设计文档

**日期：** 2026-06-19
**状态：** 已确认

---

## 目标

将 ReqManPy 后端从 Python/FastAPI 重构为 Golang，第一期聚焦核心 CRUD 和 Issue 管理，逐步替换。

## 技术栈

| 层 | 选型 | 理由 |
|---|---|---|
| HTTP 框架 | **Gin** | 最流行，类 FastAPI 风格，生态丰富 |
| ORM | **GORM** | Go 最流行 ORM，自动迁移，关联加载 |
| 数据库 | **PostgreSQL** | 生产级 |
| 认证 | **JWT** | Access Token + Refresh Token |
| 配置 | **Viper** | 读取环境变量 |
| 迁移 | **GORM AutoMigrate** | 开发阶段自动建表 |

## 项目结构（经典分层）

```
backend-go/
├── cmd/server/main.go          # 入口：初始化 DB、注册路由、启动
├── internal/
│   ├── config/                 # Viper 配置加载
│   │   └── config.go
│   ├── model/                  # GORM 模型
│   │   ├── base.go             # 公共字段（ID, Timestamps, SoftDelete）
│   │   ├── user.go
│   │   ├── workspace.go
│   │   ├── project.go
│   │   ├── state.go
│   │   ├── label.go
│   │   └── issue.go
│   ├── dto/                    # 请求/响应结构体
│   ├── repository/             # 数据访问层
│   ├── service/                # 业务逻辑层
│   ├── handler/                # HTTP 层（Gin handlers）
│   ├── middleware/              # JWT Auth、CORS、Recovery
│   └── router/                 # 路由注册
├── config/config.yaml          # 默认配置
└── go.mod
```

分层职责：
- **handler** — 绑定请求参数 → 调 service → 写 HTTP 响应
- **service** — 业务逻辑，跨 repository 调用
- **repository** — 封装 GORM 查询，单一数据源操作
- **model** — 表结构定义，GORM 标签
- **dto** — 请求/响应 DTO，binding 验证标签

## 数据模型

### 通用字段（嵌入所有模型）

```go
type BaseModel struct {
    ID        uint64    `gorm:"primaryKey;autoIncrement"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

所有表统一：bigint 主键、created_at/updated_at 自动维护、软删除。

### 1. User
- email (unique), username (unique), display_name
- first_name, last_name, avatar
- password_hash
- is_active, is_superuser, is_email_verified
- 关系：Workspaces (via WorkspaceMember), Projects (via ProjectMember), AssignedIssues (via IssueAssignee)

### 2. Workspace + WorkspaceMember
- name, slug (unique), owner_id (FK→users)
- WorkspaceMember: workspace_id + user_id + role
- 关系：has many Projects, has many Members

### 3. Project + ProjectMember
- name, identifier, description, workspace_id (FK)
- ProjectMember: project_id + user_id + role
- 关系：has many Issues/States/Labels

### 4. State
- name, color, group (backlog/unstarted/started/completed/cancelled)
- sequence, is_default
- project_id (FK), workspace_id (FK)

### 5. Label
- name, color, description
- project_id (FK)
- 关系：many-to-many with Issue (via issue_labels)

### 6. Issue（核心）
- name, description_html, description_json
- priority (urgent/high/medium/low/none)
- sequence_id, sort_order
- start_date, target_date
- project_id (FK), workspace_id (FK), state_id (FK)
- parent_id (FK→self, 子任务支持)
- 多对多：IssueAssignee (issue_id + user_id), IssueLabel (issue_id + label_id)

## API 路由

前缀：`/api/v1`

### Auth（公开）
- `POST /api/v1/auth/register` — 注册
- `POST /api/v1/auth/login` — 登录，返回 JWT
- `GET  /api/v1/auth/me` 🔒 — 当前用户信息

### Workspace 🔒
- `GET    /api/v1/workspaces` — 用户的工作空间列表
- `POST   /api/v1/workspaces` — 创建
- `GET    /api/v1/workspaces/:id` — 详情
- `PUT    /api/v1/workspaces/:id` — 更新
- `DELETE /api/v1/workspaces/:id` — 软删除

### Project 🔒
- `GET    /api/v1/workspaces/:wid/projects` — 项目列表
- `POST   /api/v1/workspaces/:wid/projects` — 创建
- `GET    /api/v1/projects/:id` — 详情
- `PUT    /api/v1/projects/:id` — 更新
- `DELETE /api/v1/projects/:id` — 软删除

### State 🔒
- `GET    /api/v1/projects/:pid/states` — 状态列表
- `POST   /api/v1/projects/:pid/states` — 创建
- `PUT    /api/v1/states/:id` — 更新
- `DELETE /api/v1/states/:id` — 删除

### Label 🔒
- `GET    /api/v1/projects/:pid/labels` — 标签列表
- `POST   /api/v1/projects/:pid/labels` — 创建
- `PUT    /api/v1/labels/:id` — 更新
- `DELETE /api/v1/labels/:id` — 删除

### Issue 🔒
- `GET    /api/v1/projects/:pid/issues` — Issue 列表（支持 ?state_id=&priority=&search= 过滤）
- `POST   /api/v1/projects/:pid/issues` — 创建（可带 assignees、labels）
- `GET    /api/v1/issues/:id` — 详情（含关联数据）
- `PUT    /api/v1/issues/:id` — 更新（含状态变更、分配人/label 调整）
- `DELETE /api/v1/issues/:id` — 软删除

### 响应格式
```json
// 单条
{ "data": {...}, "message": "ok" }
// 列表
{ "data": [...], "total": 100, "message": "ok" }
// 错误
{ "message": "错误描述" }
```

## 实现顺序（7 步）

| 步骤 | 内容 | 依赖 |
|---|---|---|
| 1 | 项目脚手架：go mod、目录、Gin 启动、配置、GORM 连接 | 无 |
| 2 | User 模型 + Auth：注册/登录、JWT 中间件 | 1 |
| 3 | Workspace CRUD（含 WorkspaceMember） | 2 |
| 4 | Project CRUD（含 ProjectMember） | 3 |
| 5 | State CRUD | 4 |
| 6 | Label CRUD | 4 |
| 7 | Issue CRUD + 关联表 + 过滤查询 + 状态变更 | 4,5,6 |

每步完成后 curl 验证。路由保持与 FastAPI 兼容，前端无需改动。

## 不包含（第一期）

- Cycle、Module、Comment、Attachment、Notification
- Workflow、Automation、CustomField
- AI 相关功能
- 估算点、项目设置
