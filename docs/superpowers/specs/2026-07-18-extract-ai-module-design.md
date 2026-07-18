# Extract AI Module into agent-service

Date: 2026-07-18
Status: draft

## Goal

Move all AI/Agent backend code from `backend/internal/` into the existing `agent-service/` microservice, making it the single owner of AI functionality. Frontend code stays unchanged.

## Current State

**Main backend (`backend/internal/`)** AI code:
- `handler/agent_handler.go` — Agent CRUD, dispatch, triage, assign, mention
- `handler/ai_handler.go` — AI chat, search, create, analyze, suggest, sprint-plan, chart, page AI, triage analyze, assist comment
- `service/agent_service.go` (17KB) — Agent business logic
- `service/ai_service.go` (57KB) — AI business logic
- `service/llm_client.go` (20KB) — LLM client (DeepSeek/OpenAI compatible)
- `model/agent.go` — Agent, AgentActivity models
- `model/ai_config.go` — AIConfig, AIThread, AIMessage models

**agent-service/*** already has:
- Gin router on :8001, Postgres DB, JWT auth
- Loop, Pipeline, Session, Agent Registry handlers/services/models
- `internal/client/agent_client.go` — calls main backend agent APIs (will need refactor)
- `internal/common/errors.go` — error code system

**Cross-dependencies:**
- `automation_service` → `agentSvc` (for triggering agent actions)
- `comment_service` → `agentSvc` (for @agent-name mentions)
- `ai_service` → `issueSvc`, `projectSvc` (for data access)
- `agent_service` → `aiSvc` (for LLM tools)

## Target Architecture

```
          Frontend (unchanged)
               │
               ▼
    ┌─────────────────────┐
    │    Main Backend      │
    │   (Gin :8000)        │
    │                      │
    │  /api/v1/workspaces/ │──┐
    │    :ws/agents/*      │  │  HTTP Proxy ──┐
    │    :ws/ai-config/*   │  │               │
    │    :ws/loops/*       │──┤  (existing)    │
    │    :ws/pipelines/*   │  │               │
    │    :ws/agent-sess/*  │──┘               ▼
    │                      │        ┌──────────────────────┐
    │  /api/v1/projects/   │        │   agent-service      │
    │    :pid/ai/*   ──────┼─HTTP──▶│   (Gin :8001)        │
    │    :pid/agent/* ─────┼─proxy  │                      │
    │                      │        │  /api/v1/workspaces/ │
    │  issue/project/user  │◀─HTTP──│    :ws/agents/*      │
    │  (internal data API) │  call  │    :ws/ai-config/*   │
    │                      │        │    :ws/loops/*        │
    └──────────────────────┘        │    :ws/pipelines/*    │
                                    │    :ws/agent-sess/*   │
                                    │                      │
                                    │  + /ai/* endpoints    │
                                    │  + LLM Client        │
                                    │  + agents/activities │
                                    │  + ai_configs/threads│
                                    └──────────────────────┘
```

**Main backend keeps:** Proxy handlers for all AI routes (no frontend URL changes), internal data API for agent-service to query.

**agent-service gains:** All agent/AI handlers, services, models, LLM client, plus DB tables for agents, agent_activities, ai_configs, ai_threads, ai_messages.

## HTTP API Contracts

### agent-service → Main Backend (data queries, internal only)

```
GET  /api/internal/issues/:id          → { id, title, ... }
GET  /api/internal/issues?ids=1,2,3    → [{...}, ...]
GET  /api/internal/projects/:id        → { id, name, identifier, workspace_id }
GET  /api/internal/users/:id           → { id, name, email }
POST /api/internal/issues/search       → { query: RQL } → [{...}]
```

Protected by shared JWT. Not exposed to frontend.

### Main Backend → agent-service (function calls)

```
POST /api/internal/agents/:id/dispatch    → { task, issue_id, project_id, triggered_by }
POST /api/internal/agents/:id/auto-triage → { issue_id }
POST /api/internal/agents/:id/auto-assign → { issue_id }
POST /api/internal/ai/chat                → { messages, tools, context }
POST /api/internal/ai/analyze             → { issue_id, project_id }
POST /api/internal/agents/:id/mention     → { comment_id, issue_id, content }
```

These are the internal endpoints agent-service exposes. Main backend's public routes proxy to these.

## Migration Phases

### Phase 1: Infrastructure

- Add `BackendClient` to agent-service (calls main backend `/api/internal/*` for issue/project/user data)
- Add `/api/internal/*` data API to main backend (JWT-protected)
- Add DB migrations in agent-service for `agents`, `agent_activities`, `ai_configs`, `ai_threads`, `ai_messages` tables
- Refactor `automation_service` and `comment_service` in main backend: replace direct `agentSvc` calls with HTTP client calls to agent-service

### Phase 2: Agent CRUD + Dispatch

- Move `model/agent.go` → agent-service
- Move `service/agent_service.go` → agent-service
- Move `handler/agent_handler.go` → agent-service
- Replace main backend agent routes (`/workspaces/:ws/agents/*`) with reverse proxy
- Remove agent code from main backend

### Phase 3: AI Features

- Move `model/ai_config.go` → agent-service
- Move `service/ai_service.go` → agent-service
- Move `service/llm_client.go` → agent-service
- Move `handler/ai_handler.go` → agent-service
- Replace main backend AI routes (`/ai/*`, `/ai-config`, `/ai-analyze`) with reverse proxy
- Remove AI code from main backend

### Phase 4: Cleanup

- Remove residual agent/AI type references and imports from main backend
- Refactor `agent-service/internal/client/agent_client.go` — no longer calls main backend agent API, uses local service instead
- Full end-to-end regression test

## Error Handling

- agent-service uses existing error codes from `internal/common/errors.go`
- Cross-service HTTP failures: main backend proxy returns `502 Bad Gateway` with agent-service's error message
- BackendClient data API failures: agent-service returns clear error, does not swallow or crash

## Testing

- After each phase: run both main backend and agent-service handler tests
- Phase 2 verification: agent CRUD, dispatch, mention, activity end-to-end
- Phase 3 verification: AI chat, search, analyze, suggest-labels, page AI end-to-end
- Phase 4: full regression suite

## Open Questions

- None currently — to be resolved during implementation if discovered
