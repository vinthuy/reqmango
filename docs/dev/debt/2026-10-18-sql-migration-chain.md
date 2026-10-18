# SQL migration chain could not bootstrap a database

**Status:** resolved 2026-10-18
**Impact when open:** `go run ./cmd/migrate up` failed on an empty PostgreSQL
database, so the SQL migration path could not create an environment. Only the
server's `AutoMigrate` worked from scratch.

## What was wrong

| Step | Failure |
|------|---------|
| `000001_initial_schema.up.sql` | `\restrict <key>` / `\unrestrict` are psql meta-commands from pg_dump 18; `golang-migrate` executes raw SQL. |
| `000001` | `SELECT pg_catalog.set_config('search_path', '', false)` set the search path for the *session*; the driver reuses pooled connections, so later unqualified statements stopped resolving. |
| `000007_add_workspace_id.up.sql` | `ALTER TABLE … ADD INDEX` is MySQL syntax. |
| `000007` | `ALTER TABLE states ADD COLUMN workspace_id …` → column already exists, because `000001` was a **dump of the final schema**, not of the schema at that point in history. Adding `IF NOT EXISTS` everywhere would have produced a chain that only accidentally matched. |

## What was done

The chain was squashed into one idempotent baseline, generated from the schema
`AutoMigrate` produces (the actual source of truth):

```bash
pg_dump --schema-only --no-owner --no-privileges --no-comments <db> \
  | python3 scripts/idempotent_baseline.py > backend/migrations/000001_baseline.up.sql
```

`scripts/idempotent_baseline.py` rewrites the dump so it can run twice and on a
database that already has every object:

- `CREATE TABLE` / `CREATE SEQUENCE` / `CREATE [UNIQUE] INDEX` → `IF NOT EXISTS`
- every `ADD CONSTRAINT` is wrapped in a `DO` block that swallows
  `duplicate_object`, `duplicate_table` and `invalid_table_definition`
  (PostgreSQL reports a second `PRIMARY KEY` as the last one)
- the psql directives and the session-level `search_path` reset are removed

The old files are in `migrations/archive/` for reference; `golang-migrate` does
not recurse, so they never run. `000001_baseline.down.sql` drops every object in
`public` except `schema_migrations`.

## Verified

| Scenario | Result |
|----------|--------|
| empty database → `migrate up` → start server | 132 tables, version 1, `/readyz` 200 in 6 s |
| `migrate up` a second time | `migrate up: OK`, exit 0 |
| database built by `AutoMigrate` (no `schema_migrations`) → `migrate up` | adopts baseline, version 1, exit 0 |
| database already at version 22 | needs a one-time `migrate force 1`; then `up` is a no-op |

## Upgrading an existing deployment

The schema already matches the baseline, so only the bookkeeping changes:

```bash
go run ./cmd/migrate force 1
go run ./cmd/migrate up     # no-op, confirms version 1
```

`force` is explicit rather than automatic on purpose: rewriting migration state
behind the operator's back is how a database ends up at a version nobody chose.

## Related finding

The same session found that the absence of a migrate-then-start check in CI hid
a real startup crash: `AgentCostBudget` and `AgentSLA` declared `uniqueIndex`
while the SQL migration creates a single-column `UNIQUE` constraint, so
`AutoMigrate` tried to drop `uni_<table>_<column>` and aborted (`SQLSTATE
42704`). Fixed in `agent_budget.go` by using the `unique` tag. The CI job
`Migration path` now applies the baseline and boots the server on it, so this
class of divergence fails the build.
