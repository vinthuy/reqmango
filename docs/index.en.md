# Reqmango

**Triage before backlog.**

Self-hosted project management. An intake form collects requests, the triage queue proposes type, priority and likely duplicates, and you only decide the ones that are genuinely uncertain. Issues, cycles, pages and automation share one data model; AI is a layer inside that workflow rather than a separate console.

Your data stays on your machine.

[![CI](https://github.com/vinthuy/reqmango/actions/workflows/ci.yml/badge.svg)](https://github.com/vinthuy/reqmango/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/vinthuy/reqmango?sort=semver)](https://github.com/vinthuy/reqmango/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/vinthuy/reqmango/blob/master/LICENSE)

!!! note "English docs are partial"
    The knowledge base is written in Chinese and only the architecture pages have English versions so far. Untranslated pages fall back to Chinese; the switcher in the header shows where the gaps are. The [README](https://github.com/vinthuy/reqmango) is a complete English overview.

---

## Run it in 60 seconds

Published multi-arch images (`linux/amd64`, `linux/arm64`) are the fast path:

```bash
git clone https://github.com/vinthuy/reqmango.git
cd reqmango
cp .env.example .env
docker compose pull
docker compose up -d
```

Open **http://localhost** and sign in:

| | |
|---|---|
| Email | `demo@example.com` |
| Password | `demo1234` |

For AI triage, set `AI_API_KEY` in `.env` and restart the backend.

!!! warning "Demo accounts are for local use only"
    `SEED_DEMO=true` creates sample accounts whose passwords are in the public repository. Any instance reachable from the internet must turn it off and follow the [production checklist](deploy/production.md) for `SECRET_KEY`, `DB_PASSWORD` and the administrator account.

The first boot can show a blank page while the frontend container starts. `curl localhost/version` tells you which build is actually running.

![Intake to triage queue](assets/demo-04-triage-queue.png)

---

## What you get

| Area | Capability |
|------|------------|
| **AI inside the work loop** | Intake triage, issue analysis, label suggestions, cycle summaries, Ask / Build copilot with preview before write |
| Work items | List and kanban, state transitions, hierarchy up to six levels |
| Custom fields and templates | Text, number, dropdown, boolean, date, member, URL; workspace type blueprints; project templates |
| Workflow and automation | Transitions, approvals, role gates; trigger → condition → action |
| Relations and search | Blocks / relates / duplicates; RQL and multi-field filters |
| Notifications and security | Unread counts; bluemonday XSS sanitising; JWT + RBAC |
| API | 100+ REST endpoints, see the [API reference](API.md) |

AI sits on top of project management rather than replacing it. The daily path is issues, cycles, intake and pages.

---

## Where to go next

- **Put it on a server**: [production checklist](deploy/production.md) — secrets, demo data off, TLS, backups
- **Change a setting**: [configuration](configuration.md) — every environment variable and its default
- **Read the code**: [architecture overview](kb/architecture/README.en.md) — stack, layers, data model
- **Integrate**: [API reference](API.md) and [API conventions](kb/architecture/api-conventions.en.md)
- **Contribute**: [good first issues](https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md) and the [roadmap](https://github.com/vinthuy/reqmango/blob/master/ROADMAP.md)

## Project

| Entry point | What it is |
|-------------|------------|
| [GitHub](https://github.com/vinthuy/reqmango) | International mirror; CI and releases run here |
| [GitCode](https://gitcode.com/yongfeng9m-/reqmanpy) | Primary hub for Chinese contributors |
| [Gitee](https://gitee.com/vinthuy/reqmango) | Read-only mirror, no CI |
| [Releases](https://github.com/vinthuy/reqmango/releases) | Versions and changelog |
| [Security](https://github.com/vinthuy/reqmango/blob/master/SECURITY.md) | Private reporting; never a public issue |

MIT licensed. Yours to host.
