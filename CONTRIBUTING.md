# Contributing to Reqmango

Thanks for helping. Reqmango is a **self-hosted project management** app: new requests are triaged before they hit the backlog; AI sits inside Issue / Intake / Cycle — not as a separate agent IDE.

[中文版](CONTRIBUTING-zh.md)

## Dual remotes (same product)

| | URL | Role |
|---|---|---|
| **GitCode (primary for CN)** | https://gitcode.com/yongfeng9m-/reqmanpy | Day-to-day Issues / PRs in Chinese |
| **GitHub (international mirror)** | https://github.com/vinthuy/reqmango | Discovery abroad; keep in sync with primary |

One codebase, **one product sentence**. Do not treat the mirrors as different products. Maintainer sync: [docs/dev/dual-remote.md](docs/dev/dual-remote.md).

## Before you start

1. Read the product one-liner in [README.md](README.md).
2. Prefer a **small, focused PR** over a large rewrite.
3. Check open issues on the host you use, plus [good first issues](docs/dev/good-first-issues.md).
4. Optional DX check after i18n work: `cd frontend && node scripts/scan-missing-i18n-keys.mjs`

## Quick setup

### Docker (recommended)

```bash
# International mirror
git clone https://github.com/vinthuy/reqmango.git
cd reqmango
cp .env.example .env
docker compose up --build
```

Contributors in China may prefer:

```bash
git clone https://gitcode.com/yongfeng9m-/reqmanpy.git reqmango
cd reqmango
cp .env.example .env
docker compose up --build
```

Open http://localhost — `demo@example.com` / `demo1234`.

Optional AI: set `AI_API_KEY` in `.env`, restart.

### Local (without Docker)

- Go **1.25+**, PostgreSQL **16+**, Node.js **20+**
- Backend: `cd backend && go run ./cmd/server/` (API on `:8000`)
- Frontend: `cd frontend && npm install && npm run dev` (http://localhost:5173)

See README for seed accounts and env details.

## Where to change what

| Area | Path |
|------|------|
| UI / i18n | `frontend/src/` — locales in `frontend/src/locales/{zh-CN,en-US}.json` |
| API / domain | `backend/internal/` |
| Compose / deploy | `docker-compose.yml`, `backend/Dockerfile`, `frontend/Dockerfile` |
| Docs | `docs/`, root `README.md` / `README-zh.md` |

Keep **zh-CN and en-US in sync** when you add UI strings. Prefer `t('namespace.key')` over hardcoded copy. Do not leave raw keys like `common.refresh` visible in the UI.

## Development workflow

1. Fork the host you contribute on (GitCode for CN, GitHub for international) and branch from `master`.
2. Make the change; keep scope to the issue.
3. Verify locally:
   - Frontend: `cd frontend && npx vitest run` (and `npx vue-tsc --noEmit` if you touched types)
   - Backend: `cd backend && go test ./internal/...`
   - Manual: reproduce the bug or walk the happy path once
4. Open a PR using the template. Link the issue. Maintainers sync both remotes.

CI (`.github/workflows/ci.yml`) runs lint, Go tests, frontend tests, and e2e — PRs should stay green.

## What we welcome

- Bug fixes with a clear repro
- i18n / docs / DX (Docker, `.env.example`, error messages)
- Small UX improvements on Issue, Intake, Cycle, Pages
- Tests for existing behavior

## What to avoid (unless agreed in an issue)

- Large “Agent console / marketplace” expansions as the default product story
- Silent breaking API changes without migration notes
- Drive-by dependency upgrades with no functional need
- Committing secrets, `.env` with real keys, or large binaries

## Issue labels (maintainer)

| Label | Meaning |
|-------|---------|
| `good first issue` | Scoped, files named, acceptance clear |
| `help wanted` | Maintainer wants outside help |
| `bug` / `enhancement` | Type |
| `area:frontend` / `area:backend` / `area:docs` | Rough ownership |

## Code of conduct

Be respectful. Assume good intent. No harassment or personal attacks. Maintainers may close out-of-scope or hostile threads.

## Maintainers

- Dual-remote sync: [docs/dev/dual-remote.md](docs/dev/dual-remote.md)
- GitHub About / Topics: [docs/dev/github-public-profile.md](docs/dev/github-public-profile.md)
- Publishing curated tasks: [good-first-issues.md](docs/dev/good-first-issues.md)

## License

By contributing, you agree your contributions are licensed under the [MIT License](LICENSE).
