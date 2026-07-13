# Architecture Overview

reqmango adopts a decoupled frontend-backend architecture with Go + Vue 3 full-stack.

**Last Updated**: 2026-07-13

---

## Tech Stack Overview

| Layer | Technology | Status |
|-------|------------|--------|
| Backend Framework | Go + Gin 1.x | ✅ Primary |
| ORM | GORM 2.x | ✅ |
| Database | PostgreSQL 16+ | ✅ |
| Authentication | JWT (golang-jwt/v5) | ✅ |
| LLM Integration | DeepSeek / Anthropic / OpenAI-compatible | ✅ |
| Frontend Framework | Vue 3 + Composition API | ✅ |
| Frontend Build | Vite | ✅ |
| State Management | Pinia | ✅ |
| CSS | Tailwind CSS | ✅ |
| Type System | TypeScript | ✅ |

---

## System Layers

```
┌──────────────────────────────────────────┐
│              Frontend (Vue 3)             │
│  Views → Components → Stores → API →     │
│         TypeScript Types                 │
└──────────────────┬───────────────────────┘
                   │ HTTP (JSON + JWT) + SSE (AI streaming)
┌──────────────────▼───────────────────────┐
│            Go Backend (Gin)              │
│                                           │
│  Router → Middleware → Handler           │
│                          │                │
│                        Service            │
│                          │                │
│                    Model (GORM)           │
│                          │                │
│                      PostgreSQL           │
└──────────────────────────────────────────┘
```

## Backend Layer Responsibilities

| Layer | Directory | Responsibility |
|-------|-----------|----------------|
| Handler | `internal/handler/` (28) | HTTP request binding, call Service, write response. No business logic |
| Service | `internal/service/` (27) | Pure business logic, cross-Model operations, returns AppError |
| Model | `internal/model/` (28) | GORM model definitions, table structure mapping |
| DTO | `internal/dto/` | Request/response structs, decoupled from Model |
| Middleware | `internal/middleware/` | JWT auth, RBAC, CORS, logging |
| Router | `internal/router/` | Route registration, middleware and Handler mounting |
| Common | `internal/common/` | Error types, constants, pagination utilities |
| RQL | `internal/rql/` | ReqMan Query Language (lexer/parser/executor) |

## Frontend Layer Responsibilities

| Layer | Directory | Responsibility |
|-------|-----------|----------------|
| Views | `src/views/` (16) | Page-level components, corresponding to routes |
| Components | `src/components/` (58) | Reusable components |
| Stores | `src/stores/` (3) | Pinia state management |
| API | `src/api/` (23) | Axios wrapper, backend API calls |
| Types | `src/types/` (21) | TypeScript type definitions (including filters.ts) |
| Composables | `src/composables/` (5) | Composables (useConfirm, useRQL, useAI, usePermission, useFilters) |
| Router | `src/router/` | Vue Router configuration (16 routes, with minRoleLevel guard) |

---

## Implemented Modules

| Module | Go Backend | Frontend | Description |
|--------|------------|----------|-------------|
| Auth | ✅ | ✅ | JWT register/login/me |
| Workspace | ✅ | ✅ | CRUD + member management |
| Project | ✅ | ✅ | CRUD + archive + statistics + Lead + Subscriber |
| Issue | ✅ | ✅ | CRUD + hierarchy/archive/batch/import-export/Issue-Page association |
| State | ✅ | ✅ | 5 fixed groups, project-level CRUD |
| Label | ✅ | ✅ | Project-level CRUD |
| Cycle | ✅ | ✅ | CRUD + start/end/cancel + burndown chart + progress |
| Module | ✅ | ✅ | Tree CRUD + Issue association + progress statistics |
| CustomField | ✅ | ✅ | 7 types + option management + Issue value binding |
| IssueType | ✅ | ✅ | Workspace-level + project-level, hierarchy binding |
| TypeTemplate | ✅ | ✅ | Workspace-level type blueprint |
| ProjectTemplate | ✅ | ✅ | Project template + Apply |
| WorkItemTemplate | ✅ | ✅ | Work item creation template |
| Workflow | ✅ | ✅ | State transitions + approval rules |
| Automation | ✅ | ✅ | Trigger-Condition-Action |
| Comment | ✅ | ✅ | Nested replies + resolve |
| Notification | ✅ | ✅ | CRUD + read/unread + summary statistics |
| Saved Views | ✅ | ✅ | JSONB filters/sort/column config + default view |
| Pages | ✅ | ✅ | Tree hierarchy + archive/restore + Markdown |
| Release | ✅ | ✅ | CRUD + Issue association + progress |
| Estimate | ✅ | ✅ | Points/Categories/Time three modes |
| Attachment | ✅ | ✅ | File upload + metadata |
| Relation | ✅ | ✅ | Custom in/out naming + Issue association |
| RQL | ✅ | ✅ | Custom DSL search |
| AI | ✅ | ✅ | Chat SSE + NL Search + Smart Create + Analyze + Page AI + Triage |
| Time Tracking | ✅ | ✅ | Start/Stop/List/Summary |
| Recurring | ✅ | ✅ | daily/weekly/monthly/cron + UI config |
| Intake & Triage | ✅ | ✅ | Public submission + Accept/Reject + AI analysis |
| Conditional Fields | ✅ | ✅ | Field visibility rules |
| Cover Image | ✅ | ✅ | Issue cover |
| Command Palette (⌘K) | ✅ | ✅ | Keyboard quick navigation |
| AI Settings | ✅ | ✅ | Provider/Model/APIKey UI configuration |
| RBAC | ✅ | ✅ | 55 permissions + 3 default roles + custom role management |
| Quick Create | ✅ | ✅ | Inline quick issue creation |
| Git Integration | ✅ | ✅ | GitHub/GitLab native integration + Webhook |
| Project CustomField Enrollment | ✅ | ✅ | Project-level custom field enable/disable |
| Workspace Workflow | ✅ | ✅ | Workspace-level workflow + project-level override |

---

## In Progress

| Module | Description |
|--------|-------------|
| FilterBar Unified Filter | Unified filter entry + RQL bidirectional sync + semantic operators + State Group + Group By/Order By |
| SavedView Enhancement | New sort_config / columns / group_by fields, complete Views restore chain |

## Future Extensions

| Module | Description |
|--------|-------------|
| Calendar/Gantt View | Calendar/Gantt chart view |

---

## 🌐 Language

- **English** (this document)
- [中文文档](README.md)