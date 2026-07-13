# API Conventions

**Last Updated**: 2026-07-13

---

## Basic Conventions

| Item | Convention |
|------|------------|
| Protocol | HTTPS (production) / HTTP (development) |
| Prefix | `/api/v1` |
| Content Type | `application/json` |
| Character Encoding | UTF-8 |
| Authentication | `Authorization: Bearer <jwt_token>` |

---

## URL Design

### Resource Naming

- Use plural form: `/workspaces`, `/projects`, `/issues`
- Use kebab-case: `/custom-fields`
- Nested resources express ownership: `/workspaces/:wid/projects`

### Route Patterns

```
# Nested (Recommended — expresses ownership)
GET  /api/v1/workspaces/:wid/projects

# Top-level + query param (Alternative — when ownership is optional)
GET  /api/v1/projects?workspace_id=N

# Globally unique resources
GET  /api/v1/auth/me
PUT  /api/v1/issues/:id
```

---

## Authentication & Authorization

### Public Endpoints

```
POST /api/v1/auth/register
POST /api/v1/auth/login
```

### Protected Endpoints (🔒)

All other endpoints require JWT Bearer Token. Middleware automatically validates and injects current user.

### Error Responses

| HTTP Status | AppError | Usage |
|-------------|----------|-------|
| 400 | `BadRequest` | Invalid request parameters |
| 401 | `Unauthorized` | Not logged in or token expired |
| 403 | `Forbidden` | Insufficient permissions |
| 404 | `NotFound` | Resource not found |
| 409 | `Conflict` | Resource conflict (e.g., duplicate slug) |
| 422 | `Validation` | Data validation failed |
| 500 | `Internal` | Server internal error |

---

## Response Format

### Success — Single Item

```json
{
  "data": { "id": 1, "name": "..." },
  "message": "ok"
}
```

### Success — List

```json
{
  "data": [ { "id": 1 }, { "id": 2 } ],
  "total": 100,
  "message": "ok"
}
```

### Success — No Body

```json
{
  "message": "deleted"
}
```

### Error

```json
{
  "message": "Description of error"
}
```

---

## Pagination

Use `page` and `page_size` query parameters:

```
GET /api/v1/issues?page=1&page_size=20
```

| Parameter | Type | Default | Max |
|-----------|------|---------|-----|
| `page` | int | 1 | — |
| `page_size` | int | 20 | 100 |

Response includes `total` field indicating total count.

---

## Filtering & Search

| Endpoint | Supported Filter Parameters |
|----------|-----------------------------|
| `GET /projects/:pid/issues` | `state_id`, `priority`, `search`, `label_id`, `assignee_id`, `cycle_id`, `rql`, `sort_by`, `sort_order`, `group_by` |
| `GET /projects/:pid/cycles` | `status` |
| `GET /projects/:pid/modules` | `include_archived` |

The `search` parameter performs fuzzy matching on `name` and `description`.

### Search Endpoints

All entities support dedicated search endpoints using `q` query parameter:

| Entity | Search Endpoint | Search Fields |
|--------|-----------------|---------------|
| Initiative | `GET /workspaces/:wid/initiatives/search?q=` | name |
| Release | `GET /projects/:pid/releases/search?q=` | name, version |
| Cycle | `GET /projects/:pid/cycles/search?q=` | name |
| Module | `GET /modules/search?project_id=&workspace_id=&q=` | name |
| Label | `GET /projects/:pid/settings/labels/search?q=` | name |
| Page | `GET /projects/:pid/pages/search?q=` | title, content |

**Response Format**:
```json
{
  "data": [ { "id": 1, "name": "..." }, ... ]
}
```

### RQL Filtering

Advanced filtering via `rql` query parameter:

```
GET /api/v1/projects/1/issues?rql=priority="urgent" AND state_id IN (1,2)
```

RQL supports 18 fields (including `state_group`, `assignee_id`, `label`, `cycle_id`, `module_id`, etc.), full-text search (`LIKE`), date ranges (`BETWEEN`), null checks (`IS NULL`/`IS NOT NULL`), and custom fields (`cf_` prefix).

### SavedView Filter Restore

Views store complete state (`filters` JSONB + `rql` string + `sort_config` + `columns` + `group_by`), all restored when selecting a view.

---

## Sorting

Default sort is `created_at DESC`. Supports `sort_by` + `sort_order`:

```
GET /api/v1/issues?sort_by=priority&sort_order=asc
```

Available sort fields: `created_at`, `updated_at`, `priority`, `start_date`, `target_date`.

---

## HTTP Method Conventions

| Method | Purpose | Idempotent |
|--------|---------|------------|
| `GET` | Read | ✅ |
| `POST` | Create | ❌ |
| `PUT` | Full update | ✅ |
| `DELETE` | Delete (soft) | ✅ |

---

## Frontend Axios Wrapper

```typescript
// api/index.ts
const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' }
})

// Auto-inject JWT
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Auto-redirect on 401
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

## 🌐 Language

- **English** (this document)
- [中文文档](api-conventions.md)