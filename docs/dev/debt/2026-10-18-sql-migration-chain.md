# SQL migration chain cannot bootstrap a database

**Status:** open (found 2026-10-18 while adding a CI gate for the migrate-then-start path)
**Impact:** `go run ./cmd/migrate up` fails on an empty PostgreSQL database, so the
SQL migration path is not a supported way to create a database. The server's
`AutoMigrate` is the only path that works from scratch.
**Severity:** medium — every existing deployment already has a database, so this
only blocks new environments and disaster recovery, where it blocks them
completely.

## What happens

```
createdb reqmango_mig
DATABASE_URL=.../reqmango_mig go run ./cmd/migrate up
```

| Step | Failure |
|------|---------|
| `000001_initial_schema.up.sql` | `\restrict <key>` / `\unrestrict` are psql meta-commands from pg_dump 18; `golang-migrate` executes raw SQL. **Fixed** by deleting both lines. |
| `000001` | `SELECT pg_catalog.set_config('search_path', '', false)` sets the search path for the *session*, and the driver reuses pooled connections, so later unqualified statements (`ALTER TABLE states …`) stop resolving. **Fixed** by deleting the line; every statement in the dump is already `public.`-qualified. |
| `000007_add_workspace_id.up.sql` | `ALTER TABLE … ADD INDEX` is MySQL syntax. **Fixed** by rewriting the three statements as `CREATE INDEX IF NOT EXISTS`. |
| `000007` | `ALTER TABLE states ADD COLUMN workspace_id …` → column already exists, because `000001` is a **dump of the final schema**, not of the schema at that point in history. |

The last one is structural. `000001` is a 118 KB `pg_dump` of a database that
already contained everything later migrations add, so the chain re-applies
evolution steps to columns that exist. Adding `IF NOT EXISTS` to every `ALTER`
would make it run, but the result would depend on which snapshot `000001` came
from — not a real migration history.

## Options

1. **Squash (recommended).** Regenerate `000001_initial_schema.up.sql` from the
   current schema, move every later migration into `migrations/archive/`, and
   make `cmd/migrate` the documented bootstrap. Requires re-baselining
   `schema_migrations` on existing databases, so it needs a maintenance window.
2. **Declare AutoMigrate the only path.** Delete `cmd/migrate` and the
   `migrations/` tree, and document that a new environment is created by
   starting the server. Cheapest, but loses the explicit, reviewable history.
3. **Add `IF NOT EXISTS` to every `ALTER`.** Unblocks a fresh database without a
   maintenance window, but leaves a chain that only accidentally matches the
   schema. Recommended only as a stopgap.

## Related

The same session found that `command/migrate`'s absence from CI hid a real
startup crash: `AgentCostBudget` and `AgentSLA` declared `uniqueIndex` while the
SQL migration creates a single-column `UNIQUE` constraint, so `AutoMigrate`
tried to drop a constraint named `uni_<table>_<column>` and aborted
(`SQLSTATE 42704`). Fixed in `agent_budget.go` by using the `unique` tag.

A CI job that runs `cmd/migrate up` and then boots the server would have caught
it. It is not added yet because the chain fails for the structural reason above;
add it when option 1 or 3 lands.
