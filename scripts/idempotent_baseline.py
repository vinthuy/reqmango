"""Turn a pg_dump --schema-only file into an idempotent baseline migration.

The baseline has to run on two kinds of database:
  * an empty one (new environment, disaster recovery);
  * one already created by the server's AutoMigrate, where every object exists.
Hence IF NOT EXISTS on creates and a duplicate_object guard on constraints.
"""

import re
import sys

src, dst = sys.argv[1], sys.argv[2]
sql = open(src, encoding="utf-8").read()

# psql-only directives from pg_dump 18 and a session-level search_path reset that
# leaks into later migrations over pooled connections.
sql = re.sub(r"(?m)^\\(?:un)?restrict [^\n]*\n", "", sql)
sql = re.sub(r"(?m)^SELECT pg_catalog\.set_config\('search_path', '', false\);\n", "", sql)

sql = re.sub(r"(?m)^CREATE TABLE ", "CREATE TABLE IF NOT EXISTS ", sql)
sql = re.sub(r"(?m)^CREATE SEQUENCE ", "CREATE SEQUENCE IF NOT EXISTS ", sql)
sql = re.sub(r"(?m)^CREATE (UNIQUE )?INDEX ", lambda m: "CREATE %sINDEX IF NOT EXISTS " % (m.group(1) or ""), sql)


def guard_constraint(match):
    stmt = match.group(0).rstrip()
    if not stmt.endswith(";"):
        stmt += ";"
    # 42710 duplicate_object, 42P07 duplicate_table, 42P16 invalid_table_definition
    # (PostgreSQL reports a second PRIMARY KEY on a table as the last one).
    return (
        "DO $baseline$ BEGIN\n  "
        + stmt
        + "\nEXCEPTION\n"
        + "  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;\n"
        + "END $baseline$;"
    )


sql = re.sub(
    r"(?m)^ALTER TABLE (?:ONLY )?[^\n]*(?:\n[ \t]+)?ADD CONSTRAINT [^\n]*;",
    guard_constraint,
    sql,
)

header = """-- Baseline schema (squashed 2026-10-18).
--
-- Regenerate with:
--   pg_dump --schema-only --no-owner --no-privileges --no-comments <db> \\
--     | python3 scripts/idempotent_baseline.py > backend/migrations/000001_baseline.up.sql
--
-- It is written to be re-runnable: an environment created by the server's
-- AutoMigrate already has every object, and adopting migrations there must be
-- a no-op rather than an error.

"""
open(dst, "w", encoding="utf-8", newline="\n").write(header + sql)
print("wrote %s (%d bytes)" % (dst, len(header + sql)))
