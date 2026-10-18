# Archived migrations

These are the original `000001`–`000022` migrations, kept for reference only.
`golang-migrate` reads only the top level of `migrations/`, so nothing here runs.

They were retired on 2026-10-18 because they could not create a database from
scratch: `000001` was a `pg_dump` of the **final** schema, while the later files
re-applied evolution steps (`ADD COLUMN workspace_id`, …) to tables that already
had those columns. They also contained psql-only directives (`\restrict`) and
MySQL syntax (`ALTER TABLE … ADD INDEX`).

The replacement is a single idempotent baseline generated from the schema the
server's `AutoMigrate` produces:

```bash
pg_dump --schema-only --no-owner --no-privileges --no-comments <db> \
  | python3 scripts/idempotent_baseline.py > backend/migrations/000001_baseline.up.sql
```

It can run on an empty database and on one already built by `AutoMigrate`, so
adopting migrations in an existing environment is a no-op. See
[the debt note](../debt/2026-10-18-sql-migration-chain.md) for details.
