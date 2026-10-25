# API Conventions（API 设计约定）

**最后更新**: 2026-10-25

约定以 `backend/internal/router/router.go` 和中间件为准。各资源的路径清单不写在这里。

---

## 基础约定

| 项目 | 约定 |
|------|------|
| 前缀 | 业务接口 `/api/v1`。`/health`、`/readyz`、`/version` 在根上，不带此前缀 |
| 内容类型 | `application/json`（上传附件除外） |
| 认证 | `Authorization: Bearer <jwt>`，或 `Bearer reqmango_pat_...`（个人访问令牌，只存 SHA-256） |

---

## URL

- 资源名用复数、kebab-case：`/custom-fields`、`/issue-types`
- 从属关系用路径嵌套：`/workspaces/:wsParam/...`、`/projects/:projectId/...`
- 工作空间参数名是 `wsParam`，值可以是数字 ID 或 slug
- 更新动词不统一：工作项是 `PUT /issues/:issueId`，工作空间是 `PATCH /workspaces/:wsParam`。以路由注册为准，不要假设全站只用一种

---

## 谁可以调用

### 不需要登录

注册在认证中间件之前：

| 方法 | 路径 |
|------|------|
| GET | `/health`、`/readyz`、`/version` |
| POST | `/api/v1/auth/register`、`/api/v1/auth/login` |
| POST | `/api/v1/intake/:projectId` |
| GET | `/api/v1/intake/:projectId/info` |
| POST | `/api/v1/intake-channels/:token/webhook`、`/api/v1/intake-channels/:token/email` |
| POST | `/api/v1/webhook/github/:id`、`/api/v1/webhook/git/:projectId` |
| POST | `/api/v1/autopilot/webhook/:token` |

### 需要登录

其余 `/api/v1` 路由使用 `ScopedAuthMiddleware`（`middleware/auth.go`）：先校验 Bearer，再跑 `scope.go` 的租户检查。

两个 SSE 接口不走这层中间件，因为浏览器 `EventSource` 不能设置 `Authorization`。它们在 handler 里自己验 token：`GET /api/v1/sse`、`GET /api/v1/chats/:chatId/stream`。

`access.go` 在部分路由组上再加一道门：

| 中间件 | 作用 |
|--------|------|
| `RequireWorkspaceMember` | `:wsParam` 的活跃成员。超级用户通过 |
| `RequireWorkspaceAdmin` | 该工作空间的管理员 |
| `RequireProjectAccess` | 项目成员，或该项目所属工作空间的成员 |
| `RequireIssueAccess` | 能访问该工作项所在项目的人 |
| `RequirePermission` | RBAC 权限码，如 `issue:edit` |

### 租户检查（`scope.go`）

每条已认证路由在 handler 之前做这些事。超级用户跳过。

- 查询参数 `workspace_id`、`project_id`、`issue_id` 必须是调用者能触及的
- 路径上的 `:wsParam` 必须是调用者所在的工作空间；`:projectId`、`:issueId` 同理
- 其它 `:param` 按路径段查表（`segmentTables` / `paramTables`），行上的 `project_id` / `workspace_id` / `issue_id` / `page_id` 必须属于调用者，并且与路径里的父级一致
- `adminRoutes` 里列出的删除、归档、改成员路由还要求工作空间或项目管理员
- 子资源行查不到时放行，由 handler 返回 404。路径上的项目不存在，或工作空间 slug 对不上，中间件自己返回 404

找不到对应表的参数不会被这层拦截。新增带 ID 的路由时，要在 `scope.go` 登记表名，否则租户检查会漏掉它。

---

## 响应

没有统一的 `{data, message}` 信封，调用方不能按一种形状解析所有接口。

| 情况 | 实际形状 |
|------|----------|
| 工作项列表 `GET /issues` | JSON 数组。总数在响应头 `X-Total-Count`，不在 body |
| 单条资源 | 通常直接返回对象 |
| 部分搜索（如发布搜索） | `{"data":[...]}` |
| 多数错误 | `{"message":"..."}` |
| `common.RespondError` | `{"error_code","message","detail"}`。`detail` 可省略 |
| 个别校验（发布搜索缺 `q`） | `{"error":"..."}` |

`AppError` 的 HTTP 状态由错误码决定（`internal/common/error_codes.go`）：校验类是 422，未登录 401，无权限 403，不存在 404，冲突 409。

---

## 分页与排序

分页参数是 `limit` 和 `offset`，不是 `page` / `page_size`。

```
GET /api/v1/issues?project_id=1&limit=50&offset=0
```

`ParsePagination(limit, offset, defaultLimit, maxLimit)` 的默认值和上限由各 handler 传入。工作项列表是默认 50、上限 100。`limit` 非正数或无法解析时用默认值，超过上限时截断。

工作项列表的排序参数是 `sort_by` 和 `sort_dir`（`asc` / `desc`），也可以传 `sort_config`（JSON 数组）。未指定时按 `sort_order`、`sequence_id` 排。

---

## 工作项列表过滤

`GET /issues` 读取这些查询参数：`project_id` 或 `workspace_id`、`state_id`、`priority`、`assignee_id`、`parent_id`、`cycle_id`、`issue_type_id`、`module_id`、`search`、`is_draft`、`cf_field_id`、`cf_value`、`cf_and`、`rql`、`sort_by`、`sort_dir`、`sort_config`、`limit`、`offset`。

---

## 搜索

查询参数名不统一：

| 路径 | 参数 |
|------|------|
| `GET /workspaces/:wsParam/initiatives/search` | `q` |
| `GET /projects/:projectId/cycles/search` | `q` |
| `GET /projects/:projectId/pages/search` | `q` |
| `GET /projects/:projectId/releases/search` | `q`（缺少则 400） |
| `GET /projects/:projectId/settings/labels/search` | `q` |
| `GET /modules/search` | `q`，以及 `project_id` / `workspace_id` |
| `GET /issues/search` | `query`，以及 `workspace_id` |

---

## RQL

工作项列表接受 `rql`。独立执行入口是 `POST /api/v1/rql/search`。

`NewIssueQueryContext`（`internal/rql/executor.go`）登记的字段：

`id`、`sequence_id`、`name`、`description`、`state`、`state_id`、`state_group`、`priority`、`type`、`assignee`、`assignee_id`、`reporter`、`label`、`label_id`、`cycle`、`cycle_id`、`module`、`module_id`、`issue_type_id`、`created_by`、`milestone`、`project_id`、`workspace_id`、`created_at`、`updated_at`、`start_date`、`target_date`。

自定义字段用 `cf_<字段 id>`，`custom_field:<id>` 会先被规范成 `cf_` 前缀。

---

## 前端 Axios

`frontend/src/api/index.ts`：

- `baseURL` 为 `/api/v1`，超时 30 秒
- 请求头带 `localStorage` 里的 `token`
- 401 时清掉 `token` 和 `user`，然后 `window.location.href = '/login'`
- 307 时按 `Location` 重试一次

---

## 🌐 语言

- **中文** (本文档)
- [English](https://vinthuy.github.io/reqmango/en/kb/architecture/api-conventions/)
