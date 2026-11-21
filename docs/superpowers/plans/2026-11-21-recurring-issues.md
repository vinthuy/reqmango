# Recurring Issues Implementation Plan

> **For agentic workers:** Use subagent-driven-development or executing-plans. Checkbox steps.

**Goal:** Mount recurrence UI on issue detail and harden permissions; optionally add a minimal runner.

**Architecture:** Wire existing `RecurrenceConfig.vue` + `/issues/:id/recurrence` API. Hide cron in P0. Add `issue:edit` on mutate routes. P1: poller clones issues.

**Spec:** [docs/specs/2026-11-21-recurring-issues.md](../../specs/2026-11-21-recurring-issues.md)

---

### Task 1: Mount RecurrenceConfig + trim cron (AC-01/02)

- [ ] Import in `IssueDetail.vue` + `IssueDetailPanel.vue`
- [ ] Hide/disable cron option in `RecurrenceConfig.vue`; i18n toasts
- [ ] Vitest smoke if practical
- [ ] Commit: `feat(ui): mount recurrence config on issue detail`

### Task 2: RBAC (AC-03)

- [ ] `RequirePermission(..., "issue:edit", "project")` on POST/PUT/DELETE
- [ ] Commit: `fix(recurrence): require issue:edit on mutate`

### Task 3 (P1): Executor

- [ ] Background ticker or cron entrypoint
- [ ] Clone template issue; advance `next_run`
- [ ] Tests
- [ ] Commit: `feat(recurrence): spawn issues on next_run`
