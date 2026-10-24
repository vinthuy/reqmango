# Security Policy

Reqmango is self-hosted software: the operator runs the database, the API and
the reverse proxy. This policy covers the code in this repository and the
default configuration it ships. It does not cover the infrastructure you run.

## Supported versions

| Version | Supported |
|---------|-----------|
| Latest `master` | Yes |
| Latest tagged release | Yes |
| Older releases | Best effort, security fixes land on `master` first |

There is no long-term-support branch yet. If you are running an older commit,
the fix you need will usually be a `git pull` away.

## Reporting a vulnerability

**Please do not open a public issue for a security problem.**

Use GitHub's private reporting flow: go to the **Security** tab of
[vinthuy/reqmango](https://github.com/vinthuy/reqmango) and choose
**Report a vulnerability**. That opens a private advisory visible only to the
maintainers.

If private reporting is unavailable to you, email the maintainer listed in the
repository profile and mention "security" in the subject.

Please include, as far as you can:

- affected version or commit hash,
- the deployment shape (Docker Compose, bare binary, reverse proxy in front?),
- reproduction steps or a proof of concept,
- the impact you believe it has,
- whether you plan to disclose it publicly, and on what timeline.

### What to expect

- **Acknowledgement** within 3 business days.
- **Triage and a first assessment** within 7 business days.
- A fix for confirmed high-severity issues as soon as practical, with a
  security advisory and a release that names the fixed version.
- Credit in the advisory unless you ask to stay anonymous.

We follow coordinated disclosure: we ask for **90 days** before public
disclosure, and we will coordinate with you if a fix takes longer. If a
vulnerability is already being exploited, tell us and we will shorten that
window.

## Scope

In scope:

- authentication and session handling (JWT issuance, token refresh, password
  reset, API tokens),
- authorization, including cross-workspace and cross-project access
  (`workspace_id` / `project_id` scoping),
- injection: SQL, command, template, and stored/reflected XSS,
- SSRF through webhooks, AI providers, or integrations,
- file upload handling and path traversal,
- secret leakage in the repository,
- privilege escalation through roles, workflow transitions, or automation rules.

Out of scope, unless you can show concrete impact:

- findings that require a compromised host, database, or admin account,
- missing security headers with no demonstrated exploit,
- rate-limit tuning opinions,
- reports produced only by an automated scanner with no reproduction,
- the demo accounts described below, when `SEED_DEMO=true` is set deliberately.

## Deploying safely

The defaults are chosen so that an untouched deployment is not trivially
compromised. Two settings deserve explicit attention:

1. **`SEED_DEMO`.** When `true`, the server creates sample workspaces and the
   demo logins documented in the README
   (`admin@reqmango.com` / `demo@example.com`, password `demo1234`). Those
   passwords are public: **never enable this on an instance reachable from the
   internet.** With `SEED_DEMO=false` (the default) no demo account is created;
   instead the first boot creates one superuser and prints a random password
   once to the log, or uses `ADMIN_EMAIL` / `ADMIN_PASSWORD` if you set them.

2. **`SECRET_KEY`.** It signs access tokens. If it is unset, the server
   generates a random per-process key and warns: every restart invalidates all
   sessions. Generate one with `openssl rand -hex 32` and keep it out of git.

See [production deployment](docs/deploy/production.md) for the full checklist
(TLS, database exposure, backups, secret rotation).

## Automated checks in CI

Every pull request runs:

- `gitleaks` over the working tree (`.gitleaks.toml` records accepted
  exceptions and why),
- Go dependency and vulnerability checks,
- the tenancy audit (`backend/cmd/route_audit`), which probes every route as a
  user without memberships, as a member, and as a cross-tenant caller,
- the full test and E2E suites.

These reduce the chance of shipping a regression; they are not a substitute for
reviewing a change that touches authentication or tenancy.
