# Tech Stack

**Last updated**: 2026-10-25

> This page records **choices** and **major versions** only. Exact patch versions live in
> `backend/go.mod` and `frontend/package.json` — copying them here is the fastest way to
> make documentation stale, so we don't.

---

## Runtimes

| Runtime | Version | Source of truth |
|---------|---------|-----------------|
| Go | 1.25 | `go` directive in `backend/go.mod` |
| Node.js | 24 (LTS) | `engines` in `frontend/package.json`, `.nvmrc` |
| PostgreSQL | 16+ | 16 in both production and CI |
| Browser | Modern browser with ES2022 | Vite build target |

---

## Go backend (primary)

| Component | Choice | Purpose |
|-----------|--------|---------|
| HTTP framework | Gin | Routing, middleware, request handling |
| ORM | GORM | Database mapping, migration, queries |
| Database driver | pgx (postgres) | PostgreSQL driver |
| Database | PostgreSQL | Primary datastore |
| Authentication | golang-jwt/jwt v5 | JWT issuing and verification |
| Passwords | bcrypt | Password hashing |
| Configuration | Viper | Multi-source config (env vars, YAML) |
| HTML sanitizing | bluemonday | HTML sanitization |
| JSON | bytedance/sonic | High-performance JSON encoding |

---

## Frontend

| Component | Choice | Purpose |
|-----------|--------|---------|
| Framework | Vue 3 | Composition API |
| Build | Vite 5 | Dev server + bundling |
| Language | TypeScript | Type safety |
| State | Pinia | Global state |
| Routing | Vue Router 4 | SPA routing |
| CSS | Tailwind CSS | Utility-first styling |
| HTTP | Axios | API calls |
| Rich text | TipTap (vue-3) | Page docs / wiki editing |
| Drag and drop | vue-draggable-plus | Ordering |

---

## Testing and tooling

| Layer | Choice |
|-------|--------|
| Go unit tests | standard `testing` + `go-sqlmock` |
| Frontend unit tests | Vitest |
| E2E | Playwright (`frontend/e2e`) |
| Lint | golangci-lint, ESLint, `vue-tsc` |
| Images | Multi-stage Docker (Go/Node builder + Alpine runtime) |

---

## Why these choices

### Why migrate from Python to Go?

1. **Performance**: Go compiles to a native binary; throughput per host far exceeds async Python
2. **Deployment**: one static binary, no virtualenv
3. **Type safety**: compile-time checks remove a class of runtime errors
4. **Concurrency**: goroutines suit IO-heavy APIs

### Why keep the Vue 3 frontend?

Frontend and backend are decoupled through a REST API, so the frontend does not need to change when the backend language does.

---

## 🌐 Language

- **English** (this document)
- [中文文档](https://vinthuy.github.io/reqmango/kb/architecture/tech-stack/)
