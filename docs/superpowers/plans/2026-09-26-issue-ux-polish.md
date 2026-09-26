# Issue List & Detail UX Polish Implementation Plan

> **For agentic workers:** Execute task-by-task. Steps use checkbox syntax.

**Goal:** Ship P0 consistency (instant save, URL sync, panel parity) + P1 compact property sidebar rows per the approved spec.

**Architecture:** Frontend-only. Shared `IssuePropertySidebar` restyle; `Project.vue` owns `?view=` / `?issue=` query; full-page and panel share instant-save title/description and header actions.

**Tech Stack:** Vue 3 + Vue Router + existing toast/confirm composables.

**Spec:** `docs/superpowers/specs/2026-09-26-issue-ux-polish-design.md`

## Global Constraints

- No Linear-style Popovers (option A only).
- Panel does not gain AI/Chat/Git tabs.
- Prefer slug routes: `/workspace/:slug/project/:id`.
- No new backend APIs.

---

### Task 1: Compact property sidebar

**Files:** `frontend/src/components/IssuePropertySidebar.vue`

- [x] Row layout (label left, control right); section dividers
- [x] Disable Release when `isLocked`
- [x] Remove auto-assign on `localAgentId` watch (Assign button only; keep unassign on clear)

### Task 2: Detail header + full-page save model

**Files:** `IssueDetailHeader.vue`, `IssueDetail.vue`

- [x] Editable title (blur/Enter → emit `update:title`); remove Save button
- [x] Instant description save; load relation summary on mount; `useConfirm` for approvals

### Task 3: Panel parity

**Files:** `IssueDetailPanel.vue`

- [x] Copy link / watch / delete in header
- [x] Filter states like full page; approval via `useConfirm`

### Task 4: Project URL sync + slug links

**Files:** `Project.vue`, `IssueList.vue`, `IssueCreate.vue` (as needed)

- [x] Sync `view` and `issue` query; Back/Esc closes panel
- [x] Fix create links to slug routes

### Task 5: Verify

- [x] Unit tests: IssueDetailHeader + IssuePropertySidebar (12 passed)
- [x] Touched-file type errors cleared (`IssueList` route import)
