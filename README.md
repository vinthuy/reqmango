# Reqmango

[![CI](https://github.com/vinthuy/reqmango/actions/workflows/ci.yml/badge.svg)](https://github.com/vinthuy/reqmango/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/vinthuy/reqmango?sort=semver)](https://github.com/vinthuy/reqmango/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Container images](https://img.shields.io/badge/ghcr.io-reqmango-2496ED?logo=docker&logoColor=white)](https://github.com/vinthuy/reqmango/pkgs/container/reqmango-backend)

**Self-hosted project management where new requests are triaged first.**  
Type, priority, and likely duplicates are suggested before work lands in the backlog — you only decide the uncertain ones.

[中文文档](README-zh.md) · [Contributing](CONTRIBUTING.md) · [国内共建主场 GitCode](https://gitcode.com/yongfeng9m-/reqmanpy) · [GitHub mirror](https://github.com/vinthuy/reqmango)

---

## Try it

Prebuilt multi-arch images (`linux/amd64`, `linux/arm64`) — fastest:

```bash
git clone https://github.com/vinthuy/reqmango.git
# China: git clone https://gitcode.com/yongfeng9m-/reqmanpy.git reqmango
cd reqmango
cp .env.example .env
docker compose pull
docker compose up -d
```

Or build both images locally (several minutes on a cold cache):

```bash
docker compose up --build
```

**First run:** wait for both services to be healthy before opening the app; a
temporary blank page while the frontend image is still starting does not mean
startup failed. `curl localhost/version` answers which build is running.

Open **http://localhost** and sign in:

| | |
|---|---|
| Email | `demo@example.com` |
| Password | `demo1234` |

Optional AI (Intake triage / analyze / labels): set `AI_API_KEY` in `.env`, then restart.

![Intake → triage demo](docs/assets/demo.gif)

### 60-second walkthrough

1. Sign in with the demo account.
2. Open the **Demo** project → **Project settings** → **Intake triage**.
3. Open the Intake form link, submit a short request (e.g. “login sometimes fails on mobile”).
4. Return to the triage queue: accept / reject, and (with `AI_API_KEY`) run AI analyze / label suggestions on issues.

---

## What you get

| Area | Capability |
|------|------------|
| **AI in the work loop** | Intake triage, issue analyze, label suggestions, Cycle / sprint summary, Ask / Build Copilot with preview before write |
| Work items | List + Kanban, state transitions, hierarchy (up to 6 levels) |
| Custom fields & templates | Text / number / dropdown / boolean / date / member / URL; workspace type blueprints; project templates |
| Workflow & automation | Transitions, approvals, role gates; trigger → condition → action |
| Relations & search | Blocks / Relates / Duplicates; RQL + multi-field filters |
| Notifications & security | Unread counts; XSS sanitization (bluemonday); JWT + RBAC |
| API | 100+ REST endpoints |

AI is a layer on top of project management — not a separate “agent IDE” as the default story. Advanced Agent console routes may still exist in the codebase; day-to-day PM paths are Issue, Cycle, Intake, and Pages.

---

## Requirements

| Mode | Need |
|------|------|
| **Docker (recommended)** | Docker + Docker Compose |
| Local dev | Go **1.25+**, PostgreSQL **16+**, Node.js **20+** |

Stack matches CI and `docker-compose.yml` (`postgres:16`, Go 1.25 image).

---

## Local development (without Docker)

```bash
# Database
psql -U postgres -c "CREATE DATABASE reqmango;"

# Backend
cd backend
cat > .env << EOF
DATABASE_URL=postgres://postgres:postgres@localhost:5432/reqmango?sslmode=disable
SECRET_KEY=$(openssl rand -hex 32)
ACCESS_TOKEN_EXPIRE_MINUTES=10080
PORT=8000
DEBUG=true
SEED_DEMO=true
EOF
go run ./cmd/server/
```

The server migrates and seeds on start, so this is all a new environment needs. `SEED_DEMO=true` creates the demo accounts; drop that line and the first boot instead creates one admin and prints a random password to the log (see [production deployment](docs/deploy/production.md)).

### Database migrations

`cmd/migrate` is the explicit alternative for environments that want a reviewed
schema step separate from the application boot:

```bash
cd backend
go run ./cmd/migrate up      # apply the baseline
go run ./cmd/migrate version # 1, dirty=false
```

`migrations/000001_baseline.up.sql` is a single idempotent baseline, so it runs
on an empty database *and* on one already created by the server's AutoMigrate.
An existing database that predates the squash needs a one-time
`go run ./cmd/migrate force 1`; see
[the debt note](docs/dev/debt/2026-10-18-sql-migration-chain.md).

Seed data includes `demo@example.com` / `demo1234`, workspace `reqmango-dev` (projects `CORE` / `MOBILE` / `AI` / `OPENAPI`), sample sprints / modules / issues. Those demo passwords are public, so the demo is for local evaluation only.

```bash
# Frontend (API on :8000)
cd frontend
npm install
npm run dev
```

Open **http://localhost:5173**.

### Production build (local)

```bash
cd backend && go build -o server ./cmd/server/
cd frontend && npm run build
```

---

## Project layout

```
reqmango/
├── backend/     # Go + Gin + GORM API
├── frontend/    # Vue 3 + TypeScript + Vite
├── sdk/         # MCP / CLI (optional; not the default product story)
├── docs/        # API, architecture, product specs
└── docker-compose.yml
```

---

## Docs

- [API Reference](docs/API.md)
- [Architecture](docs/kb/architecture/) — [tech stack](docs/kb/architecture/tech-stack.md), [Go backend](docs/kb/architecture/backend-go.md), [frontend](docs/kb/architecture/frontend.md), [data model](docs/kb/architecture/data-model.md)
- [Production deployment](docs/deploy/production.md) — secrets, demo data off, TLS, backups

---

## Contributing

We welcome Issues and Pull Requests — especially small, focused changes.

- [Contributing guide](CONTRIBUTING.md) · [中文](CONTRIBUTING-zh.md)
- [Good first issues](docs/dev/good-first-issues.md) (curated starter tasks)
- [Security policy](SECURITY.md) — report vulnerabilities privately, never in a public issue
- **CN collaboration hub:** [GitCode](https://gitcode.com/yongfeng9m-/reqmanpy) · **International mirror:** [GitHub](https://github.com/vinthuy/reqmango) — same code, same product story ([dual-remote](docs/dev/dual-remote.md))

## License

[MIT](LICENSE)
