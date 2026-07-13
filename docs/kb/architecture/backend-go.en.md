# Go Backend Architecture

**Last Updated**: 2026-07-13

---

## Project Structure

```
backend/
├── cmd/server/main.go              # Entry point
├── internal/
│   ├── config/config.go            # Config loading (Viper + env vars)
│   ├── common/
│   │   ├── constants.go            # Role, priority, state group constants
│   │   ├── error_codes.go          # Error code definitions (with i18n messages)
│   │   ├── errors.go              # AppError type
│   │   └── pagination.go          # Pagination helpers
│   ├── model/                      # GORM models (44 files)
│   │   ├── base.go                # BaseModel embedded struct (ID/CreatedAt/UpdatedAt/DeletedAt)
│   │   ├── user.go                # User (with IsSuperuser, GetID, IsSuper methods)
│   │   ├── workspace.go           # Workspace, WorkspaceMember
│   │   ├── project.go             # Project, ProjectMember
│   │   ├── issue.go               # Issue, IssueAssignee, IssueLabel, IssueActivity
│   │   ├── issue_type.go          # IssueType + field binding
│   │   ├── issue_type_template.go # Workspace-level issue type templates
│   │   ├── state.go               # State, StateTransition
│   │   ├── label.go               # Label
│   │   ├── cycle.go               # Cycle
│   │   ├── module.go              # Module, ModuleIssue
│   │   ├── page.go                # Page (hierarchical documents)
│   │   ├── page_template.go       # PageTemplate
│   │   ├── page_version.go        # PageVersion
│   │   ├── comment.go             # Comment (threaded replies)
│   │   ├── relation.go            # RelationType, IssueRelation
│   │   ├── workflow.go            # Workflow + AutomationRule + AutomationExecution
│   │   ├── attachment.go          # Attachment
│   │   ├── estimate.go            # EstimatePoint, EstimateCategory, EstimateTime, ProjectEstimateSettings
│   │   ├── time_track.go          # TimeTrack
│   │   ├── recurrence_rule.go     # RecurrenceRule (periodic issues)
│   │   ├── notification.go        # Notification (9 types)
│   │   ├── webhook.go             # Webhook configuration
│   │   ├── slack.go               # Slack integration
│   │   ├── github.go              # GitHub integration
│   │   ├── agent.go               # AI Agent + AgentActivity
│   │   ├── ai_config.go           # AI configuration (user-level)
│   │   ├── mcp.go                 # MCP Server configuration
│   │   ├── saved_view.go          # SavedView (view presets, with sort_config/columns/group_by/RQL)
│   │   ├── saved_dashboard.go     # SavedDashboard
│   │   ├── saved_report.go        # SavedReport
│   │   ├── dashboard_widget.go    # DashboardWidget
│   │   ├── search_template.go     # SearchTemplate
│   │   ├── plugin.go              # Plugin + PluginEventLog
│   │   ├── project_page_tab.go    # Project page tab configuration
│   │   ├── project_update.go      # ProjectUpdate (status timeline)
│   │   ├── project_template.go    # ProjectTemplate
│   │   ├── work_item_template.go  # WorkItemTemplate
│   │   ├── custom_field.go        # CustomField + 7 types
│   │   ├── conditional_field.go   # ConditionalField (conditional visibility)
│   │   ├── initiative.go          # Initiative + InitiativeProject (cross-project strategic goals)
│   │   ├── release.go             # Release + Roadmap
│   │   ├── role.go                # Role + RoleLevel (Admin=20/Member=15/Guest=5)
│   │   └── permission.go          # Permission + RolePermission join table
│   ├── dto/
│   │   ├── request/               # Request DTOs (31 files)
│   │   │   ├── auth.go, workspace.go, project.go
│   │   │   ├── issue.go, cycle.go, module.go, page.go
│   │   │   ├── custom_field.go, conditional_field.go
│   │   │   ├── workflow.go, state.go, label.go
│   │   │   ├── relation.go, notification.go
│   │   │   ├── time_track.go, recurrence.go
│   │   │   ├── project_template.go, work_item_template.go
│   │   │   ├── type_template.go, issue_type.go
│   │   │   ├── initiative.go, release.go, saved_view.go
│   │   │   ├── role.go, rql.go
│   │   │   ├── dashboard.go, page_lock.go, page_template.go
│   │   │   ├── page_version.go, plugin.go, saved_report.go
│   │   │   └── project.go
│   │   └── response/              # Response DTOs (27 files)
│   │       ├── auth.go, workspace.go, project.go
│   │       ├── issue.go, cycle.go, module.go, page.go
│   │       ├── page_template.go, page_version.go
│   │       ├── custom_field.go, conditional_field.go
│   │       ├── workflow.go, state.go, label.go
│   │       ├── relation.go, notification.go
│   │       ├── time_track.go, recurrence.go
│   │       ├── project_template.go, work_item_template.go
│   │       ├── type_template.go, issue_type.go
│   │       ├── release.go, saved_view.go
│   │       ├── role.go
│   │       ├── dashboard.go, plugin.go, search_template.go, saved_report.go
│   │       └── corresponding request files
│   ├── service/                    # Business logic (45 source files + 6 test files)
│   │   ├── auth_service.go            # JWT signing and verification
│   │   ├── workspace_service.go       # Workspace CRUD + members
│   │   ├── project_service.go         # Project CRUD + members + statistics + archive
│   │   ├── issue_service.go           # Issue CRUD + search + batch + import/export + relations
│   │   ├── cycle_service.go           # Cycle CRUD + state transitions + progress + burndown chart + search
│   │   ├── module_service.go          # Module CRUD + tree + Issue relations + statistics + search
│   │   ├── page_service.go            # Page CRUD + tree + archive + search
│   │   ├── page_template_service.go   # PageTemplate CRUD
│   │   ├── page_version_service.go    # PageVersion management
│   │   ├── comment_service.go         # Comment CRUD + replies + resolve
│   │   ├── relation_service.go        # RelationType CRUD + Issue relation management
│   │   ├── workflow_service.go        # Workflow CRUD + transition validation
│   │   ├── attachment_service.go      # Attachment CRUD
│   │   ├── estimate_service.go        # 3 estimation mode management
│   │   ├── time_track_service.go      # Time tracking
│   │   ├── recurrence_service.go      # Periodic Issue generation
│   │   ├── notification_service.go    # 9 notification types + mark as read + reminders
│   │   ├── webhook_service.go         # Webhook sending + HMAC-SHA256 signing
│   │   ├── slack_service.go           # Slack notification formatting + sending
│   │   ├── github_service.go          # GitHub Issues sync + Webhook receiving
│   │   ├── ai_service.go              # AI chat/search/create/analyze/charts/triage
│   │   ├── agent_service.go           # AI Agent CRUD + Dispatch/Triage/Assign
│   │   ├── mcp_service.go             # MCP Server config + connection management
│   │   ├── automation_service.go      # Automation rule CRUD + trigger execution
│   │   ├── saved_view_service.go      # View preset management
│   │   ├── custom_field_service.go     # Custom field 7 types + conditional rules
│   │   ├── conditional_field_service.go # Conditional visibility evaluation
│   │   ├── project_settings_service.go # State/Label CRUD + defaults + search
│   │   ├── project_page_tab_service.go  # Page tab configuration
│   │   ├── project_update_service.go    # Project status updates
│   │   ├── project_template_service.go  # Project templates
│   │   ├── work_item_template_service.go # Work item templates
│   │   ├── issue_type_service.go        # Issue type CRUD + field binding + sorting + copy
│   │   ├── type_template_service.go     # Workspace type templates
│   │   ├── initiative_service.go        # Initiative CRUD + search
│   │   ├── release_service.go           # Release + Roadmap + search
│   │   ├── role_service.go              # RBAC roles + permission queries (8 methods)
│   │   ├── report_service.go            # Report generation + chart data
│   │   ├── plugin_service.go            # Plugin management
│   │   ├── search_template_service.go   # SearchTemplate CRUD
│   │   ├── saved_report_service.go      # SavedReport management
│   │   ├── dashboard_service.go         # Dashboard CRUD + Widget
│   │   ├── field_permission_service.go  # Field permission management
│   │   ├── sse_hub.go                   # SSE real-time event hub
│   │   ├── llm_client.go                # LLM client (DeepSeek/Anthropic)
│   │   └── *_test.go                    # 6 test files
│   ├── handler/                    # HTTP Handlers (46 files)
│   │   └── One handler per service + sse_handler.go + role_handler.go + project_issue_type_handler.go + field_permission_handler.go + intake_handler.go
│   ├── rql/                        # RQL query language engine (10 files: 7 source + 3 tests)
│   │   ├── lexer.go               # Lexer (tokenize)
│   │   ├── parser.go              # Parser (AST building)
│   │   ├── ast.go                 # AST type definitions
│   │   ├── builder.go             # AST → SQL/GORM query
│   │   ├── executor.go            # Query executor
│   │   ├── handler.go             # HTTP endpoint `/api/v1/rql/search`
│   │   ├── service.go             # Query service
│   │   └── *_test.go              # 3 test files
│   ├── middleware/                  # Middleware (6)
│   │   ├── auth.go                # JWT authentication: parse Bearer token → inject currentUser
│   │   ├── authorization.go       # RBAC authorization: RequirePermission / RequireRoleLevel
│   │   ├── cors.go                # CORS policy
│   │   ├── language.go            # i18n language detection (Header/Query)
│   │   ├── logger.go              # Request logging
│   │   └── rate_limiter.go        # Token Bucket rate limiting (default 500 req/min, 0=disabled)
│   ├── i18n/
│   │   ├── i18n.go / i18n_test.go # Internationalization logic
│   │   ├── messages_en.json       # English error messages
│   │   └── messages_zh.json       # Chinese error messages
│   └── seed/
│       ├── seed.go                # Seed data (20 users + 1000 Issues + Workspace/Project)
│       └── seed_rbac.go           # RBAC seed (55 permissions + 3 default roles)
```

---

## Layered Architecture

```
Router ──→ Middleware Chain ──→ Handler ──→ Service ──→ Model (GORM)
    │           │                    │            │
    │     Auth/CORS/Lang             │            └── DB operations
    │     Logger/RateLimit           └── DTO binding + HTTP response
    └── Gin route tree (80+ endpoints)
```

### Handler Layer
- Use Gin Context to bind request parameters (`c.ShouldBindJSON`, `c.Param`, `c.Query`)
- Call Service layer methods
- Return unified JSON response format
- **No business logic**

### Service Layer
- Pure Go business logic
- Use GORM for database operations
- Return custom `AppError` type (with HTTP status code and i18n error code)
- Cross-Model operations (e.g., Issue with Assignee, Label, Attachment)
- Inject dependencies (Notification, Webhook, Automation, Slack services)

### Model Layer
- GORM tags define table structure + indexes
- All models embed `BaseModel` (ID uint64, CreatedAt, UpdatedAt, DeletedAt)
- `DeletedAt gorm.DeletedAt` for soft delete
- Associations managed via GORM Tag + Preload

### DTO Layer
- `request/` — Client request structs with `binding` validation tags
- `response/` — Server response structs
- Separated from Model to avoid exposing internal DB fields

### RQL Layer
- Custom RQL (reqmango Query Language)
- Complete pipeline: Lexer → Parser → AST → SQL/GORM generation
- Endpoint: `POST /api/v1/rql/search`
- Feature parity with mainstream query languages, free in community edition

---

## Core Components

### AppError (Unified Error Handling)

```go
type AppError struct {
    StatusCode int
    ErrorCode  string  // RQL error code (with i18n support)
    Message    string
    Err        error
}
```

Predefined error factories: `NotFound()`, `Conflict()`, `Unauthorized()`, `Forbidden()`, `Validation()`, `Internal()`, `BadRequest()`

### i18n Error Messages
- `messages_en.json` / `messages_zh.json` bilingual
- `ErrorCode` maps to specific error message
- Language detected via `Accept-Language` Header or `lang` Query parameter

### Pagination

```go
func ParsePagination(c *gin.Context) (limit int, offset int)
```

Default limit=20, max=100, parsed from query params `page` and `page_size`.

### Authentication Flow

1. `POST /api/v1/auth/login` — bcrypt password verification, returns JWT + token_type + expires_at
2. Subsequent requests carry JWT in `Authorization: Bearer <token>` header
3. Auth middleware parses JWT, queries user, injects via `c.Set("currentUser", user)`
4. Handler retrieves via `c.MustGet("currentUser").(*model.User)`

### RBAC Authorization

- `authorization.go` provides `RequirePermission(db, permCode)` and `RequireRoleLevel(db, minLevel)` middleware
- Superusers (IsSuperuser=true) bypass all permission checks automatically
- Permission format: `resource:action` (e.g., `issue:create`, `workspace:manage`)
- Role levels: Admin=20 (full access), Member=15 (create/edit), Guest=5 (read-only)
- Supports workspace-level and project-level custom roles
- 5 RBAC APIs: `GET/POST /api/v1/workspaces/:wsParam/roles`, `PUT/DELETE /api/v1/workspaces/:wsParam/roles/:id`, `GET /api/v1/permissions`

### Rate Limiting

- Token Bucket algorithm, grouped by user ID or IP
- Default 500 req/min (configurable via `RATE_LIMIT_REQUESTS` env var)
- `limit=0` disables completely (dev/CI environments)
- Response headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

### SSE Real-time Notifications

- `sse_hub.go` — Broadcast Hub implementation
- `sse_handler.go` — SSE endpoint `/api/v1/sse`
- 9 notification types pushed to frontend via SSE

---

## API Route Overview

All routes prefixed with `/api/v1`, 80+ endpoints:

| Route Group | Resources |
|-------------|-----------|
| `/auth/` | Register / Login / Current User |
| `/workspaces/` | CRUD + Members + AI Agent + MCP + GitHub + Slack config |
| `/projects/` | CRUD + Members + Statistics + Reports + Webhook + Automation |
| `/projects/:id/cycles` | Cycle CRUD + State transitions + Progress + Burndown |
| `/projects/:id/modules` | Module CRUD + Tree + Issue relations + Statistics |
| `/projects/:id/pages` | Page CRUD + Archive |
| `/projects/:id/issue-types` | Issue type CRUD + Field binding |
| `/projects/:id/workflows` | Workflow CRUD + Rules |
| `/projects/:id/releases` | Release + Roadmap |
| `/projects/:id/ai/` | AI Chat/Search/Charts/Triage |
| `/projects/:id/automations` | Automation rule CRUD + Execution |
| `/issues/` | CRUD + Search + Batch + Import/Export + Tree + Relations + Time tracking + Recurrence |
| `/comments/` | CRUD + Replies + Resolve |
| `/custom-fields/` | CRUD + Options + Conditional rules |
| `/templates/` | Project templates + Type templates + Work item templates |
| `/rql/search` | RQL query execution |
| `/notifications/` | List + Read + Reminders |
| `/webhook/github/:id` | GitHub Webhook receiving (public endpoint) |
| `/intake/:projectId` | Public issue submission (no auth required) |
| `/sse` | Real-time event streaming |
| `/roles/` | RBAC role CRUD + Permission list |
| `/permissions/` | Global permission enumeration |

---

## Configuration

Configured via environment variables + `config.yaml`:

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL DSN | — |
| `SECRET_KEY` | JWT signing secret | — |
| `PORT` | Listen port | 8080 |
| `RATE_LIMIT_REQUESTS` | Rate limit threshold (req/min) | 500 |
| `RATE_LIMIT_WINDOW_SEC` | Rate limit window | 60 |
| `AI_PROVIDER` | AI provider | deepseek |
| `AI_MODEL` | AI model | deepseek-chat |
| `AI_BASE_URL` | AI API endpoint | api.deepseek.com |
| `DEBUG` | Debug mode | false |
| `LANGUAGE` | Default language | zh |

---

## 🌐 Language

- **English** (this document)
- [中文文档](backend-go.md)