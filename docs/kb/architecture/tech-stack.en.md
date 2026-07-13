# Tech Stack

**Last Updated**: 2026-07-13

---

## Current Tech Stack

### Go Backend (Primary)

| Component | Selection | Version | Purpose |
|-----------|-----------|---------|---------|
| HTTP Framework | Gin | v1.x | Routing, middleware, request handling |
| ORM | GORM | v2.x | Database mapping, migration, queries |
| Database Driver | pgx (postgres) | — | PostgreSQL driver |
| Database | PostgreSQL | 16+ | Primary database |
| Authentication | golang-jwt/jwt | v5 | JWT token signing and verification |
| Password | bcrypt | — | Password hashing |
| Configuration | Viper | — | Multi-source config (env vars, YAML) |
| LLM Integration | DeepSeek / Anthropic / OpenAI-compatible | — | AI capabilities |

### Frontend

| Component | Selection | Purpose |
|-----------|-----------|---------|
| Framework | Vue 3 | Composition API |
| Build | Vite | Dev server + bundling |
| Language | TypeScript | Type safety |
| State Management | Pinia | Global state |
| Router | Vue Router 4 | SPA routing |
| CSS | Tailwind CSS | Atomic styling |
| HTTP | Axios | API calls |
| Rich Text | TipTap | Rich text editing |
| SSE | fetch + ReadableStream | AI streaming |

### Python Backend (Legacy, being phased out)

| Component | Selection | Purpose |
|-----------|-----------|---------|
| Framework | FastAPI | Async web framework |
| ORM | SQLAlchemy 2.0 | Async database operations |
| Validation | Pydantic V2 | Data validation and serialization |
| Database | SQLite / PostgreSQL | Development/production database |
| Authentication | python-jose | JWT |

### Integrations

| Component | Purpose |
|-----------|---------|
| GitHub API | Native GitHub integration |
| GitLab API | Native GitLab integration |
| Slack API | Slack notifications and issue creation |
| Webhooks | External system integration |
| MCP Server | AI tool integration protocol |

---

## Rationale

### Why Migrate from Python to Go?

1. **Performance**: Go compiles to native binary, single-machine throughput far exceeds Python async
2. **Deployment**: Single static binary, no virtual environment required
3. **Concurrency**: Goroutines provide lightweight concurrency
4. **Standard Library**: Rich standard library, fewer dependencies
5. **Type Safety**: Static typing catches errors early

---

## 🌐 Language

- **English** (this document)
- [中文文档](tech-stack.md)