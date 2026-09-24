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
`scripts/run-full-e2e.ps1` and documented in `E2E_COVERAGE_REPORT.md §11.4`).

If a CI run or a batch operation triggers HTTP 429 errors, check the backend log
for the `[Webhook]` line counting `429` responses.

---

## 7. Open findings for awareness

| Finding | Severity | Recommendation |
|---------|----------|----------------|
| `.env` / `.claude/settings.local.json` contain live provider keys locally | Medium | Rotate keys if the files were ever shared. Both are gitignored and never committed. |
| `backend/config.yaml` was deleted; the file carried a placeholder JWT secret | Low | Delete ignored copies locally if any remain; the app no longer reads this file. |
| BUG-58: state transitions for agent workflows are stubs | Medium | Implement once the data model is agreed (see docs/bug-list.md). |
| `errorlint` / `nilerr` linting is off due to hundreds of pre-existing patterns | Low | Refactor in a dedicated PR when ready. |
| `govulncheck` currently runs as advisory only | Low | Make blocking once the pinned Go toolchain version matches CI exactly. |
