# Roadmap

This is a direction, not a schedule. Work happens in the open as issues and
pull requests; anything here can move if a bug, a security report, or a
well-argued use case gets in front of it.

The source of truth for what is finished is [CHANGELOG.md](CHANGELOG.md) and the
[releases](https://github.com/vinthuy/reqmango/releases). The internal pipeline
status (Chinese) lives in [docs/dev/pipeline-status.md](docs/dev/pipeline-status.md).

## Now — make the project easy to adopt and safe to self-host

- [x] One-time security baseline: demo seed data opt-in, bootstrap admin,
      `SECURITY.md`, production checklist (v0.1.0)
- [x] Tagged releases and multi-arch `ghcr.io` images (v0.1.0)
- [x] Documentation site (MkDocs, published from `docs/`) with a browsable
      quickstart, deployment, configuration and API reference
- [x] Language handling that follows the browser on first visit instead of
      hardcoding Chinese, with a visible switcher
- [x] Community files and automation: code of conduct, support guide,
      Dependabot, protected `master`, public roadmap (this file)
- [x] Screenshots for the README, covering intake, triage, issues, cycles and
      analytics (`docs/assets/screenshot-*.png`)

## Next — close the remaining work-item gaps

**Shipped in [v0.2.0](https://github.com/vinthuy/reqmango/releases/tag/v0.2.0)** ([PR #32](https://github.com/vinthuy/reqmango/pull/32)).
SRS + acceptance: [docs/specs/](docs/specs/README.md), [docs/dev/acceptance/](docs/dev/acceptance/).
Process: [docs/dev/sdd.md](docs/dev/sdd.md).

| Candidate | Status |
|-----------|--------|
| Quick create from the work-item list | Done |
| Mandatory field validation + conditional fields | Done (import CF column mapping still a follow-up) |
| Type hierarchy rules | Done |
| Sub-items panel | Done |
| Page linking | Done |
| Notification auto-trigger | Done (due-soon remains P1) |
| Attachment backend | Done |

## Later — deeper workflows

- [x] [Merge duplicate work items](docs/specs/2026-11-07-merge-duplicates.md) ([PR #33](https://github.com/vinthuy/reqmango/pull/33))
- [x] Time tracking and estimate categories (P0 — `docs/specs/2026-11-07-time-tracking-estimates.md`; categories/time-mode on issues still Later)
- [ ] [Recurring work items](docs/specs/2026-11-21-recurring-issues.md) (P0: mount UI + RBAC; executor P1)
- [ ] Release picker polish ([#36](https://github.com/vinthuy/reqmango/issues/36))
- [ ] [Public demo playground](docs/deploy/demo.md) ([#38](https://github.com/vinthuy/reqmango/issues/38))

## Not planned

- The archived **standalone agent platform / full-SDLC agent** direction. It is
  kept in [`docs/superseded/`](docs/superseded/) for history and is not on the
  roadmap. AI stays a layer inside Issue / Intake / Cycle.
- MCP/CLI as a growth story. The SDK exists and is maintained, but it is not the
  default product path.
- A paid/cloud edition. Reqmango is self-hosted and MIT-licensed.

## How priorities are chosen

1. Security and tenancy defects.
2. Data loss or corruption.
3. Install/upgrade failures.
4. Issues with the most 👍 and the clearest reproduction.
5. The candidates above.

If you want something sooner, the most effective move is a focused pull request
or a well-scoped issue — see [CONTRIBUTING.md](CONTRIBUTING.md).
