# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[semantic versioning](https://semver.org/spec/v2.0.0.html). While the major
version is `0`, a minor bump may include breaking changes; they are called out
under **Changed** / **Upgrade notes**.

## [Unreleased]

### Changed

- CI moved every GitHub Action to the major that runs on Node 24 natively
  (`checkout@v7`, `setup-go@v7`, `setup-node@v7`, `setup-python@v7`,
  `upload-artifact@v7`, `configure-pages@v6`, `upload-pages-artifact@v5`,
  `deploy-pages@v5`, `docker/*-action@v4`/`v6`/`v7`). GitHub had deprecated the
  Node 20 action runtime and printed a warning on every job; nothing changes for
  contributors. The Go jobs also pin `cache-dependency-path` to `backend/go.sum`,
  so the module cache is restored instead of re-downloaded on every run.
- CI now pins `ubuntu-24.04` instead of `ubuntu-latest`. GitHub is migrating the
  `ubuntu-latest` label to Ubuntu 26 and annotates every job in the meantime; a
  silent base-image swap is also the kind of change that breaks Playwright's
  browser dependencies. Move the label once CI is green on 26.
- The backend runtime image moved from `alpine:3.20` to `alpine:3.24`; 3.20 had
  stopped receiving packages in 2026, and the builder stage was already on 3.24.
- The release workflow has now been exercised end to end against a throwaway
  `v0.0.0-rc.1` tag: both images published for `linux/amd64` and `linux/arm64`
  with provenance and SBOM attestations, `latest` and `0.1` stayed on 0.1.0, and
  the GitHub Release came out as a pre-release. `docs/dev/release.md` records
  the rehearsal and one correction it turned up: a pre-release publishes only
  its exact version tag, never `0.2` or `latest`.
- The frontend now builds and tests on **Node.js 24** (LTS) instead of Node 20,
  which reached end-of-life in April 2026. CI runs 24, the image builder is
  `node:24-alpine`, `frontend/package.json` declares `engines.node >= 24`, and
  `.nvmrc` pins the same major. `@types/node` moved to 24 to match.

## [0.1.0] - 2026-10-24

First tagged release. Reqmango is a self-hosted project manager whose intake
queue triages new requests before they reach the backlog, with AI assisting
triage, issue analysis, label suggestions and cycle summaries.

### Added

- **Work items** — list and kanban views, state transitions, hierarchy up to six
  levels, assignees, labels, relations (blocks / relates / duplicates), comments
  and attachments.
- **Intake** — public intake forms and channels, a triage queue with
  accept/reject, and AI-suggested type, priority and duplicate candidates.
- **Custom fields and templates** — text, number, dropdown, boolean, date,
  member and URL fields; workspace type blueprints; project and issue templates.
- **Workflows and automation** — issue state machines with approvals and role
  gates; trigger → condition → action rules with a dry-run preview.
- **Planning** — cycles/sprints, modules, releases, pages with version history,
  and project metrics/dashboards.
- **Search** — RQL plus multi-field filters, saved views and search templates.
- **AI layer** — provider-agnostic (DeepSeek by default), optional; analyze,
  label suggestions, cycle summary, and a copilot that previews writes before
  applying them.
- **Bulk import** — CSV/JSON with pre-validation, preview (`dry_run`), explicit
  column mapping and per-row warnings.
- **Platform** — JWT auth, role-based access control, API tokens, notification
  centre with SSE, webhooks, audit-friendly tenancy scoping, MCP/CLI client in
  `sdk/`.
- **Operations** — single-command Docker Compose, `GET /health`, `GET /readyz`,
  `GET /metrics` (superuser), slow-request logging, graceful shutdown.
- **Docs** — English/Chinese README and contributing guide, architecture notes,
  API reference, and a production deployment checklist.

### Security

- **Demo and E2E seed data are now opt-in** (`SEED_DEMO`, `SEED_E2E`, both
  off by default). Every release before this one seeded a superuser whose
  password was published in the repository, so a reachable instance could be
  taken over with a documented credential.
- **Bootstrap admin** — a fresh deployment with demo data off creates one
  superuser from `ADMIN_EMAIL`/`ADMIN_PASSWORD`, or prints a random password
  once. An existing account is never promoted automatically.
- Added `SECURITY.md` (private reporting, scope, response targets) and
  `docs/deploy/production.md` (secrets, network exposure, TLS/SSE, backups).
- CI scans the working tree with gitleaks and replays every API route as an
  outsider, a member and a cross-tenant caller (`cmd/route_audit`).

### Fixed

- Workflow transitions are now scoped to their owning domain, so an agent
  workflow id can no longer address an issue state transition (BUG-58).
- The CI Docker job built from the repository root, where no Dockerfile exists;
  it now builds both component images and runs on every change instead of only
  on tags that never existed.

### Upgrade notes

- `SEED_DEMO` defaults to `false`. `docker compose up` with `.env.example`
  keeps the demo experience; any existing deployment should set it explicitly
  and **rotate or delete the old demo accounts** — turning the flag off does not
  remove rows that were already seeded.
- `SECRET_KEY` must be set explicitly, or every restart invalidates all
  sessions.
- Images: `ghcr.io/vinthuy/reqmango-backend` and
  `ghcr.io/vinthuy/reqmango-frontend`, multi-arch (`linux/amd64`,
  `linux/arm64`).

[Unreleased]: https://github.com/vinthuy/reqmango/compare/v0.1.0...master
[0.1.0]: https://github.com/vinthuy/reqmango/releases/tag/v0.1.0
