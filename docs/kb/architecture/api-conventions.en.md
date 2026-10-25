# API Conventions

**Last updated**: 2026-10-25

These conventions follow `backend/internal/router/router.go` and the middleware. Per-resource path lists are not copied here.

---

## Basics

| Item | Convention |
|------|------------|
| Prefix | Business routes are under `/api/v1`. `/health`, `/readyz` and `/version` are at the root |
| Content type | `application/json` (attachment uploads excepted) |
| Authentication | `Authorization: Bearer <jwt>`, or `Bearer reqmango_pat_...` (personal access token; only the SHA-256 hash is stored) |

---

## URLs

- Resource names are plural kebab-case: `/custom-fields`, `/issue-types`
- Ownership is nested: `/workspaces/:wsParam/...`, `/projects/:projectId/...`
- The workspace parameter is named `wsParam` and accepts a numeric id or a slug
- The update verb is not uniform. Work items use `PUT /issues/:issueId`; workspaces use `PATCH /workspaces/:wsParam`. Read the route registration instead of assuming one verb

---

## Who can call

### No login

Registered before the auth middleware:

| Method | Path |
|--------|------|
| GET | `/health`, `/readyz`, `/version` |
| POST | `/api/v1/auth/register`, `/api/v1/auth/login` |
| POST | `/api/v1/intake/:projectId` |
| GET | `/api/v1/intake/:projectId/info` |
| POST | `/api/v1/intake-channels/:token/webhook`, `/api/v1/intake-channels/:token/email` |
| POST | `/api/v1/webhook/github/:id`, `/api/v1/webhook/git/:projectId` |
| POST | `/api/v1/autopilot/webhook/:token` |

### Login required

Every other `/api/v1` route uses `ScopedAuthMiddleware` (`middleware/auth.go`): validate the bearer token, then run the tenancy checks in `scope.go`.

Two SSE routes skip that middleware because browser `EventSource` cannot set `Authorization`. They check the token inside the handler: `GET /api/v1/sse` and `GET /api/v1/chats/:chatId/stream`.

`access.go` adds a further gate on some groups:

| Middleware | Admits |
|------------|--------|
| `RequireWorkspaceMember` | Active members of `:wsParam`. Superusers pass |
| `RequireWorkspaceAdmin` | Admins of that workspace |
| `RequireProjectAccess` | Project members, or members of the project's workspace |
| `RequireIssueAccess` | Callers who can access the work item's project |
| `RequirePermission` | An RBAC code such as `issue:edit` |

### Tenancy (`scope.go`)

Every authenticated route is checked before the handler. Superusers skip it.

- Query parameters `workspace_id`, `project_id` and `issue_id` must be reachable by the caller
- `:wsParam` must be a workspace the caller belongs to; the same applies to `:projectId` and `:issueId`
- Other `:param` values are resolved to a table (`segmentTables` / `paramTables`). The row's `project_id` / `workspace_id` / `issue_id` / `page_id` must be reachable and must match the parent ids on the route
- Routes listed in `adminRoutes` (delete, archive, membership changes) also require a workspace or project admin
- A missing child row is left to the handler, which returns 404. A missing project on the path, or a workspace slug that does not match, is a 404 from the middleware itself

A parameter with no table mapping is not checked. A new route that carries an id has to be registered in `scope.go`, or the tenancy check will miss it.

---

## Responses

There is no universal `{data, message}` envelope. Callers cannot parse every endpoint the same way.

| Case | Shape |
|------|-------|
| Work-item list `GET /issues` | A JSON array. The total is the `X-Total-Count` header, not a body field |
| One resource | Usually the object itself |
| Some searches (release search) | `{"data":[...]}` |
| Most errors | `{"message":"..."}` |
| `common.RespondError` | `{"error_code","message","detail"}`. `detail` is omitted when empty |
| A few validations (release search without `q`) | `{"error":"..."}` |

`AppError` status codes come from `internal/common/error_codes.go`: validation is 422, missing auth 401, forbidden 403, missing resource 404, conflict 409.

---

## Pagination and sorting

Pagination uses `limit` and `offset`, not `page` / `page_size`.

```
GET /api/v1/issues?project_id=1&limit=50&offset=0
```

`ParsePagination(limit, offset, defaultLimit, maxLimit)` takes its default and cap from the handler. The work-item list uses a default of 50 and a cap of 100. A non-positive or unparseable `limit` falls back to the default; a value above the cap is truncated.

Work-item sorting uses `sort_by` and `sort_dir` (`asc` / `desc`), or `sort_config` (a JSON array). With neither, the order is `sort_order` then `sequence_id`.

---

## Work-item list filters

`GET /issues` reads `project_id` or `workspace_id`, `state_id`, `priority`, `assignee_id`, `parent_id`, `cycle_id`, `issue_type_id`, `module_id`, `search`, `is_draft`, `cf_field_id`, `cf_value`, `cf_and`, `rql`, `sort_by`, `sort_dir`, `sort_config`, `limit`, `offset`.

---

## Search

The query parameter name is not uniform:

| Path | Parameter |
|------|-----------|
| `GET /workspaces/:wsParam/initiatives/search` | `q` |
| `GET /projects/:projectId/cycles/search` | `q` |
| `GET /projects/:projectId/pages/search` | `q` |
| `GET /projects/:projectId/releases/search` | `q` (400 if missing) |
| `GET /projects/:projectId/settings/labels/search` | `q` |
| `GET /modules/search` | `q`, plus `project_id` / `workspace_id` |
| `GET /issues/search` | `query`, plus `workspace_id` |

---

## RQL

The work-item list accepts `rql`. The standalone endpoint is `POST /api/v1/rql/search`.

Fields registered by `NewIssueQueryContext` (`internal/rql/executor.go`):

`id`, `sequence_id`, `name`, `description`, `state`, `state_id`, `state_group`, `priority`, `type`, `assignee`, `assignee_id`, `reporter`, `label`, `label_id`, `cycle`, `cycle_id`, `module`, `module_id`, `issue_type_id`, `created_by`, `milestone`, `project_id`, `workspace_id`, `created_at`, `updated_at`, `start_date`, `target_date`.

Custom fields use `cf_<field id>`. `custom_field:<id>` is normalized to the `cf_` prefix first.

---

## Frontend Axios

`frontend/src/api/index.ts`:

- `baseURL` is `/api/v1`, timeout is 30 seconds
- The request header carries the `token` from `localStorage`
- On 401 it removes `token` and `user`, then sets `window.location.href = '/login'`
- On 307 it retries once using `Location`

---

## 🌐 Language

- **English** (this document)
- [中文文档](https://vinthuy.github.io/reqmango/kb/architecture/api-conventions/)
