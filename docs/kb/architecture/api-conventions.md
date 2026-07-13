# API Conventions（API 设计约定）

**最后更新**: 2026-07-04

---

## 基础约定

| 项目 | 约定 |
|------|------|
| 协议 | HTTPS（生产）/ HTTP（开发） |
| 前缀 | `/api/v1` |
| 内容类型 | `application/json` |
| 字符编码 | UTF-8 |
| 认证 | `Authorization: Bearer <jwt_token>` |

---

## URL 设计

### 资源命名

- 使用复数形式：`/workspaces`, `/projects`, `/issues`
- 使用 kebab-case：`/custom-fields`
- 嵌套资源表示从属关系：`/workspaces/:wid/projects`

### 路由模式

```
# 嵌套（推荐 — 表达从属关系）
GET  /api/v1/workspaces/:wid/projects

# 顶级 + query param（备选 — 当从属关系可选时）
GET  /api/v1/projects?workspace_id=N

# 全局唯一资源
GET  /api/v1/auth/me
PUT  /api/v1/issues/:id
```

---

## 认证与授权

### Public Endpoints

```
POST /api/v1/auth/register
POST /api/v1/auth/login
```

### Protected Endpoints（🔒）

所有其他接口需要 JWT Bearer Token。中间件自动验证并注入当前用户。

### 错误响应

| HTTP 状态码 | AppError | 使用场景 |
|-------------|----------|----------|
| 400 | `BadRequest` | 请求参数不合法 |
| 401 | `Unauthorized` | 未登录或 token 过期 |
| 403 | `Forbidden` | 权限不足 |
| 404 | `NotFound` | 资源不存在 |
| 409 | `Conflict` | 资源冲突（如重复 slug） |
| 422 | `Validation` | 数据验证失败 |
| 500 | `Internal` | 服务器内部错误 |

---

## 响应格式

### 成功 — 单条

```json
{
  "data": { "id": 1, "name": "..." },
  "message": "ok"
}
```

### 成功 — 列表

```json
{
  "data": [ { "id": 1 }, { "id": 2 } ],
  "total": 100,
  "message": "ok"
}
```

### 成功 — 无返回体

```json
{
  "message": "deleted"
}
```

### 错误

```json
{
  "message": "描述错误原因"
}
```

---

## 分页

使用 `page` 和 `page_size` 查询参数：

```
GET /api/v1/issues?page=1&page_size=20
```

| 参数 | 类型 | 默认值 | 最大值 |
|------|------|--------|--------|
| `page` | int | 1 | — |
| `page_size` | int | 20 | 100 |

响应中包含 `total` 字段表示总数。

---

## 过滤与搜索

| 接口 | 支持的过滤参数 |
|------|---------------|
| `GET /projects/:pid/issues` | `state_id`, `priority`, `search`, `label_id`, `assignee_id`, `cycle_id`, `rql`, `sort_by`, `sort_order`, `group_by` |
| `GET /projects/:pid/cycles` | `status` |
| `GET /projects/:pid/modules` | `include_archived` |

搜索参数 `search` 通常对 `name` 和 `description` 进行模糊匹配。

### 搜索端点

所有实体支持独立的搜索端点，使用 `q` 查询参数进行模糊搜索：

| 实体 | 搜索端点 | 搜索字段 |
|------|---------|---------|
| Initiative | `GET /workspaces/:wid/initiatives/search?q=` | name |
| Release | `GET /projects/:pid/releases/search?q=` | name, version |
| Cycle | `GET /projects/:pid/cycles/search?q=` | name |
| Module | `GET /modules/search?project_id=&workspace_id=&q=` | name |
| Label | `GET /projects/:pid/settings/labels/search?q=` | name |
| Page | `GET /projects/:pid/pages/search?q=` | title, content |

**响应格式**：
```json
{
  "data": [ { "id": 1, "name": "..." }, ... ]
}
```

### RQL 筛选

支持通过 `rql` 查询参数传递 RQL 表达式进行高级筛选：

```
GET /api/v1/projects/1/issues?rql=priority="urgent" AND state_id IN (1,2)
```

RQL 支持 18 个字段（含 `state_group`、`assignee_id`、`label`、`cycle_id`、`module_id` 等），支持全文搜索（`LIKE`）、日期范围（`BETWEEN`）、空值检查（`IS NULL`/`IS NOT NULL`）、自定义字段（`cf_` 前缀）。

### SavedView 筛选恢复

视图保存时存储完整状态（`filters` JSONB + `rql` 字符串 + `sort_config` + `columns` + `group_by`），选择视图时全部恢复。

---

## 排序

默认按 `created_at DESC`。支持 `sort_by` + `sort_order`：

```
GET /api/v1/issues?sort_by=priority&sort_order=asc
```

可选排序字段：`created_at`, `updated_at`, `priority`, `start_date`, `target_date`。

---

## 请求方法约定

| 方法 | 用途 | 幂等 |
|------|------|------|
| `GET` | 读取 | ✅ |
| `POST` | 创建 | ❌ |
| `PUT` | 全量更新 | ✅ |
| `DELETE` | 删除（软删除） | ✅ |

---

## 前端 Axios 封装

```typescript
// api/index.ts
const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' }
})

// 自动注入 JWT
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// 401 自动跳转登录
api.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      router.push('/login')
    }
    return Promise.reject(err)
  }
)
```

---

## 🌐 语言

- **中文** (本文档)
- [English](api-conventions.en.md)
