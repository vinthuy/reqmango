# Security & Operations Guide

## 1. JWT signing key (`SECRET_KEY`)

`SECRET_KEY` signs every access token issued by the backend. A missing or
publicly known value lets an attacker mint valid tokens for any user.

| Action | Command |
|--------|---------|
| Generate a fresh key | `openssl rand -hex 32` |
| Rotate without downtime | Set `SECRET_KEY` to a new value, restart; existing sessions will be rejected immediately, which is the intended safe state. |
| Verify the active key | `curl -s localhost:8000/api/v1/auth/login ... \| jq .access_token \| cut -d. -f2 \| base64 -d` — the `sub` claim is the user id; a valid signature can be confirmed with `jwt.io`. |

If `SECRET_KEY` is unset or still the placeholder `change-me-in-production`, the
process starts with a random key and prints a `[SECURITY WARNING]` banner. This
means existing tokens become invalid and will be invalidated again on each restart.

---

## 2. Database log level (`DB_LOG_LEVEL`)

GORM's Info level prints every SQL statement together with its parameters, which
can both flood logs (tens of MB in a 20 minute E2E run) and persist user data to
disk.

| Value | Behaviour |
|-------|-----------|
| `silent` | No SQL logging |
| `error` | Only connection / fatal errors |
| `warn` | Errors + slow query warnings (recommended production default) |
| `info` | Every statement with parameters (default only when `DEBUG=true`) |

For long-running E2E or CI runs, set `DB_LOG_LEVEL=warn` (or `DEBUG=false`).

---

## 3. GitHub remote credential management

The repository's remote URL may contain a Personal Access Token (PAT) used for
push access. This is a local concern (not committed), but the token is exposed
in `.git/config` and in any `git remote -v` output.

### Switch to the credential manager / SSH

```bash
# Option A – credential-helper (HTTPS, no token in URL)
git remote set-url origin https://github.com/vinthuy/reqmango.git
git config --global credential.helper store  # or manager / osxkeychain

# Option B – SSH
git remote set-url origin git@github.com:vinthuy/reqmango.git
```

### Rotate the PAT

1. Revoke the current token at <https://github.com/settings/tokens>.
2. Generate a new one with `repo` scope.
3. Push once; the credential helper will prompt for the new token and persist it.

---

## 4. `.git/config` PAT embedded in the `github` remote

The repository still carries a `github` remote whose URL contains an embedded
PAT. When this changes or the token is rotated, the old URL stops working and
pushing fails. Switch the `github` remote to the same credential-helper path as
`origin` (see §3) and delete the PAT from the URL.

---

## 5. Secret scanning (gitleaks)

A `gitleaks` step runs in the CI security job. Configuration lives in
`.gitleaks.toml`, which extends the default rules and allowlists audited false
positives (test-only fixture credentials, documentation examples).

### Adding a new allowlist entry

1. Verify the match is genuinely non-secret (e.g. a prefixed placeholder in a
   test file).
2. Prefer a path-based allowlist over a regex for the smallest possible surface
   area.
3. Commit the `.gitleaks.toml` change in the same PR.

---

## 6. Rate limiting

The default limit is **500 requests per 60-second window per client IP**.

For E2E suites the limit is raised via `RATE_LIMIT_REQUESTS=200000` (set in
`scripts/run-full-e2e.ps1` and documented in `docs/E2E_COVERAGE_REPORT.md` §11.4).

If a CI run or a batch operation triggers HTTP 429 errors, check the backend log
for the `[Webhook]` line counting `429` responses.

---

## 7. Open findings for awareness

| Finding | Severity | Recommendation |
|---------|----------|----------------|
| `.env` / `.claude/settings.local.json` contain live provider keys locally | Medium | Rotate keys if the files were ever shared. Both are gitignored and never committed. |
| `backend/config.yaml` was deleted; the file carried a placeholder JWT secret | Low | Delete ignored copies locally if any remain; the app no longer reads this file. |
| BUG-58: state transitions were keyed only by `workflow_id`, which is ambiguous | **Fixed** | `state_transitions.workflow_kind` (`state`/`agent`) now disambiguates the two id spaces; see section 11. |
| `errorlint` / `nilerr` linting is off due to hundreds of pre-existing patterns | Low | Refactor in a dedicated PR when ready. |
| `govulncheck` currently runs as advisory only | Low | Make blocking once the pinned Go toolchain version matches CI exactly. |

---

## 8. Tenancy guard and the route audit (`backend/cmd/route_audit`)

`ScopedAuthMiddleware` is the single tenancy guard: it resolves the workspace,
project and child-resource IDs a request carries in its path and query string
and checks them against the caller's memberships. A route is therefore only as
safe as the IDs it exposes, and nothing but a replay of every route proves it.
`backend/cmd/route_audit` does that replay against a disposable copy of the
database.

The tool mutates the database it runs against, so point it at a copy whose
name contains `audit` — it refuses to start otherwise — and keep outbound HTTP
dead (it already forces a dead proxy for LLM and webhook calls):

```bash
createdb reqmango_audit
pg_dump reqmango | psql reqmango_audit
cd backend
mkdir -p /tmp/audit
ROUTE_AUDIT_DATABASE_URL='postgres://postgres:postgres@localhost:5432/reqmango_audit?sslmode=disable' \
  go run ./cmd/route_audit -seed -cross -out /tmp/audit/cross.md
```

| Mode | Caller and IDs used | A hit means |
|------|---------------------|-------------|
| (default) | a user with no memberships, victim IDs in path *and* query | `LEAK` / `WRITE` — tenant data or writes reachable from outside |
| `-no-query` | same, but only path IDs | the same, for routes that take their IDs from the path alone |
| `-member` | a role-15 member of the victim workspace | `ESCALATE` — an admin-only route accepted a member |
| `-cross` | admin of a second workspace, with victim child IDs behind its own parents | `XTENANT` — a route checks membership but not ownership |
| `-seed` | — | inserts one victim-owned row for every path parameter without a fixture, so the route is probed instead of reported `unresolved` |

A clean tree reports empty `XTENANT`, `ESCALATE`, `LEAK` and `WRITE` sections;
the remaining `denied` / `invalid` / `ok-empty` counts are the guard refusing
requests. `status500` and `status503` entries are unconfigured integrations
(GitHub, Slack, MCP credentials) and the AI key check, not access results —
verify those by hand before reading anything into them.

`-strict` turns the report into a gate: it exits non-zero on any `XTENANT`,
`ESCALATE`, `LEAK`, `WRITE`, `status500` or `unresolved` route. The CI job
`Tenancy audit` runs all four modes this way, each against a fresh copy of the
seeded database. Routes that intentionally carry no tenant data — the
permissions catalog, the static import template, and the caller's own
`notifications/read-all` — are allowlisted in `benignPublic` and still probed.

Pass `-seed` to **every** mode, not just `-cross`. The member probe performs
the writes a member is allowed, which deletes the very fixture rows later
routes need, so on the same copy `unresolved` climbs from run to run and
coverage quietly shrinks (4 → 35 in one session here). Seeding first keeps it
at one or two unresolvable parameters per mode.

Reports are generated output (`route_audit.md` / `.json`); they are gitignored,
so keep them out of commits.

---

## 9. Health and readiness probes

| Endpoint | Meaning | Use |
|----------|---------|-----|
| `GET /health` | the process is running; touches no dependency | container `HEALTHCHECK`, restart decisions |
| `GET /readyz` | the database answers a ping within 2 s; `503` otherwise | load balancers, CI/E2E startup, rollout gates |
| `GET /version` | build identity: version, commit, build date, Go version | "what is deployed?", support reports |

Public on purpose: none of the three reveals tenant data, and an operator
should not need a token to answer them. `/version` is injected at build time
(`-ldflags -X .../internal/version.Version=...`); a local `go run` reports
`"dev"`.

Both are public and registered before the language middleware and the rate
limiter, so probe traffic is cheap and cannot be throttled. Liveness stays
green when the database is down on purpose: a liveness failure restarts the
container, which would not fix a broken database and would drop the process
that could still serve cached and read-only paths.

`backend/Dockerfile` wires `/health` into `HEALTHCHECK`. Orchestrators should
route on `/readyz`, not on the liveness endpoint.

---

## 10. Request metrics and slow-request logging

`middleware.MetricsCollector` aggregates request counts, 4xx/5xx counts and a
latency histogram per **route template** (`c.FullPath()`), so memory is bounded
by the number of registered routes rather than by the URLs seen. Unmatched
paths collapse into `<unmatched>`. `/health` and `/readyz` are excluded so
orchestrator polling cannot drown real traffic.

`GET /api/v1/metrics` returns the snapshot. It is process-wide and spans every
tenant, so it requires a **superuser**; a member gets `403`. The payload has
`uptime_seconds`, `slow_threshold_ms`, and per route `requests`,
`client_errors`, `server_errors`, `avg_ms`, `max_ms`, `p95_ms` (estimated from
the histogram bucket, i.e. accurate to the bucket edge and never below the true
value).

Requests slower than `SLOW_REQUEST_MS` (default 1000, `0` disables) are logged
as `[SLOW] <method> <route> <status> <duration>`. The route template is used
instead of the raw path, so no path IDs and no query string reach the log.

This is intentionally a minimal in-process view, not a Prometheus exporter: it
answers "which route is slow and how often it fails" without adding a
metrics dependency. Add an exporter if dashboards or alerting are needed.

---

## 11. Workflow ownership: one table per domain

Two workflow tables exist, each owned by exactly one domain and one service:

| Table | Domain | Owner | Routes |
|-------|--------|-------|--------|
| `workflows` (+ `state_transitions`) | issue status machine | `StateWorkflowService` | `/workspaces/:wsParam/workflows/*`, read by the rules engine (`internal/workflow`) |
| `agent_workflows` (+ `workflow_nodes`, `workflow_edges`, `workflow_runs`) | multi-agent DAG orchestration | `WorkflowService` | `/projects/:projectId/workflows/*` |

Their id ranges overlap, but no table references the other, so a query filtered
on `workflow_id` alone can no longer mix domains: `state_transitions.workflow_id`
carries a foreign key (`fk_workflows_transitions`) to `workflows(id)` only, and
`workflow_nodes`/`workflow_edges` reference `agent_workflows(id)`.

Each route group binds to the table it addresses:

- `/projects/:id/workflows/:workflowId/*` addresses `agent_workflows`;
  `RequireProjectWorkflow` validates `:workflowId`, `:nodeId` and `:edgeId`
  there. This group has no transition routes — agent orchestration has nodes and
  edges, not states.
- `/workspaces/:wsParam/workflows/:workflowId/*` addresses `workflows`;
  `RequireWorkspaceWorkflow` validates the id against that table and the
  transition against `transition.workflow_id = :workflowId`.
- Workspace detail and list responses read transitions through the FK, so a
  transition from another domain cannot appear under a state workflow.
- The rules engine reads `workflows` and their transitions for the workspace,
  scoped by project or workspace-level default.

Historically a `workflow_kind` discriminator (`state` / `agent`) was added to
`state_transitions` to keep the two apart at query time, and both route groups
wrote into that one table. It was a stopgap: the column had to be threaded
through every query, the project group owned transitions that were really state
machine rows, and a wrong kind silently attached a row to the other domain. It
was removed in `000002_separate_state_transitions`: agent and orphan transitions
were archived to `state_transitions_agent_archive`, the column was dropped, and
the foreign key to `workflows` was restored. `cmd/server` performs the same
retirement before `AutoMigrate`, so a deployment that only ever runs
`AutoMigrate` converges to the same schema. See
`docs/superpowers/specs/2026-10-18-workflow-model-separation-design.md`.

New transitions inherit their parent from the route, never from the id.

---

## 12. Gaps the fresh-database audit found

Probing a database built only by `AutoMigrate` (the Docker path) rather than by
`cmd/migrate` exposed three defects that a long-lived local database hides:

| Defect | Effect | Fix |
|--------|--------|-----|
| Six models were missing from `AutoMigrate`: `ProjectAgentMember`, `AgentDecision`, `AgentCostBudget`, `AgentSLA`, `TaskLog`, `PersonalAccessToken` | every handler reading them returned 500 on a fresh install | added to the `AutoMigrate` list in `cmd/server` |
| `GET /workspaces/:ws/autopilot-tasks/:taskId` and `/executions[/:executionId]` loaded the row by id only | an admin of another workspace could read or act on a victim task by guessing its id (`XTENANT`) | `middleware.RequireWorkspaceAutopilotTask` binds both ids to `:wsParam`, returning 404 |
| `distribution` reports built `issues.id IN ()` when an RQL filter matched nothing | PostgreSQL syntax error surfaced as a failed report | empty match returns an empty report |
| `initiative` progress counted `project_id` after joining `states` | ambiguous column, progress silently stayed 0 | filter qualified as `issues.project_id` |

`-strict` in CI now runs against a database seeded the same way, so these
regressions fail the build instead of accumulating.

---

## 13. Seed data and the bootstrap admin (`SEED_DEMO`)

The server seeds on every boot. Two of those seed phases create accounts with
passwords published in this repository, so they are opt-in:

| Setting | Creates | Default |
|---------|---------|---------|
| `SEED_DEMO=true` | Sample workspaces/projects/issues, plus `admin@reqmango.com` and `demo@example.com` with password `demo1234` | **off** |
| `SEED_E2E=true` | `qa_tester@reqmango.com` with password `Test@12345`, workspace `qa-test`, project `QAT` | **off** |

CI sets both where the suites need them. Local development opts in through
`.env.example` (`SEED_DEMO=true`); the Compose demo path is the only place that
should be on.

With both off, a fresh database would have no way in, so the server bootstraps
one superuser when — and only when — no superuser exists:

- `ADMIN_EMAIL` (default `admin@reqmango.com`) and `ADMIN_PASSWORD` if set;
- otherwise a random 24-character password printed once to the log;
- an existing account with that email is **never** promoted automatically
  (otherwise the first person to register would get admin).

Deployments that predate this flag keep whatever rows they already have:
turning `SEED_DEMO` off does not delete the demo accounts. Change their
passwords or remove them by hand. See
[deploy/production.md](deploy/production.md) for the full deployment checklist.
