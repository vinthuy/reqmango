# Reqmango

A modern project management platform supporting work item management, custom fields, type templates, workflows, and automation.

---

## 🌐 Language

- **English** (this document)
- [中文文档](README-zh.md)

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.21+ + Gin + GORM |
| Database | PostgreSQL 16+ |
| Frontend | Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS |
| Authentication | JWT (golang-jwt/v5) |

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 16+
- Node.js 18+

### 1. Clone the Project

```bash
git clone https://github.com/vinthuy/reqmango.git
cd reqmango
```

### 2. Configure Database

```bash
# Create database
psql -U postgres -c "CREATE DATABASE reqmango;"
```

### 3. Configure Backend

```bash
cd backend

# Create environment configuration file
cat > .env << EOF
DATABASE_URL=postgres://postgres:postgres@localhost:5432/reqmango?sslmode=disable
SECRET_KEY=change-me-in-production-use-a-long-random-string
ACCESS_TOKEN_EXPIRE_MINUTES=10080
PORT=8000
DEBUG=true
EOF

# Start backend (auto-migrate + seed data)
go run ./cmd/server/
```

Upon startup, the backend automatically creates database tables and inserts demo data:
- Admin account: `demo@example.com` / `demo1234`
- Test accounts: `demo1@reqman.local` ~ `demo19@reqman.local` (same password)
- Demo Workspace (slug: demo) + Demo Project (identifier: DEMO)
- 6 default states, 4 Sprints, 5 modules, 100 work items
- 3 default issue types (Bug/Feature/Epic)
- 3 custom fields (Priority/Target Date/Version)

### 4. Configure Frontend

```bash
cd frontend
npm install
npm run dev
```

Open browser at `http://localhost:5173`, login with `demo@example.com` / `demo1234`.

### 5. Production Build

```bash
# Backend
cd backend && go build -o server ./cmd/server/

# Frontend
cd frontend && npm run build
```

---

## Project Structure

```
reqmango/
├── backend/              # Go Backend
│   ├── cmd/server/          # Entry point
│   ├── internal/
│   │   ├── model/           # GORM Data Models (34 files)
│   │   ├── dto/             # Request/Response DTOs (45 files)
│   │   ├── service/         # Business Logic (35 files)
│   │   ├── handler/         # HTTP Handlers (37 files)
│   │   ├── rql/             # RQL Query Language Engine
│   │   ├── middleware/      # Middleware (Auth/CORS/Lang/Log/RateLimit)
│   │   ├── i18n/            # Internationalization (en/zh)
│   │   ├── seed/            # Seed Data
│   │   ├── common/          # Utility Functions
│   │   └── config/          # Configuration Loading
│   └── config/              # YAML Configuration
├── sdk/                     # MCP Server + CLI (shared Go module)
├── frontend/                # Vue 3 Frontend
│   └── src/
│       ├── api/             # API Calls (35 modules)
│       ├── types/           # TypeScript Types
│       ├── stores/          # Pinia State Management
│       ├── views/           # Pages
│       ├── components/      # Components
│       └── router/          # Routing
└── docs/                    # Documentation
    ├── kb/                  # Knowledge Base (Architecture Docs)
    ├── dev/                 # Development Pipeline
    └── superseded/          # Historical Archive
```

## Core Features

| Feature | Description |
|---------|-------------|
| Work Item Management | CRUD + State Transitions + List/Kanban Views |
| Custom Fields | 7 types (text/number/dropdown/boolean/date/member/url) |
| Type Templates | Workspace-level type blueprints + Hierarchy + Field Binding |
| Project Templates | Package type templates, apply to projects in one click |
| Workflows | State transition rules + Approval + Role restrictions |
| Automation | Trigger → Condition → Action rule engine |
| Relations | Custom relation types (Blocks/Relates/Duplicates) |
| Hierarchy System | Up to 6 levels of work item hierarchy + Type validation |
| Advanced Search | Multi-field AND combination filtering |
| API | 90+ RESTful endpoints |

## API Documentation

See architecture documents in the [docs/kb/architecture/](docs/kb/architecture/) directory.

## Architecture Documents

- [Tech Stack](docs/kb/architecture/tech-stack.md)
- [Go Backend Architecture](docs/kb/architecture/backend-go.md)
- [Frontend Architecture](docs/kb/architecture/frontend.md)
- [Data Model](docs/kb/architecture/data-model.md)
- [API Conventions](docs/kb/architecture/api-conventions.md)
- [Type Hierarchy & Template Design](docs/kb/architecture/type-hierarchy-template-design.md)
- [Relation System Design](docs/kb/architecture/relation-system-design.md)

## Contributing

Contributions are welcome. Please submit Issues and Pull Requests.

## License

MIT