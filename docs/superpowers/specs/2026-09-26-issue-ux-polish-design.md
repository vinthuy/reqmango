# Issue List & Detail UX Polish

**Date:** 2026-09-26  
**Status:** Draft — awaiting user review before implementation plan  
**Scope:** P0 consistency + P1 compact property sidebar (option A)  
**Out of scope this round:** Linear-style property popovers, panel AI/Chat/Git tabs, list j/k keyboard nav (can follow later)

Related prior work: `2026-07-05-issue-detail-redesign.md`, `2026-06-28-modern-issue-pages-design.md`.

## Goals

Make the project issues hub and issue detail feel consistent and closer to Linear/Plane for day-to-day editing:

1. One save model everywhere (instant).
2. Shareable, refresh-safe detail peek via URL.
3. Full-page and slide-over parity for core actions.
4. Property sidebar as compact label|value rows (not a stack of form controls).

## Non-goals

- Rebuilding Filters / Kanban / Command Palette.
- New Popover pickers for state/assignee/priority (option B — next iteration).
- Adding AI / Chat / Git tabs to the slide-over panel.
- Global `j/k` list navigation (tracked as follow-up).

## Design

### 1. Save model

| Surface | Behavior |
|---------|----------|
| Title | Editable; save on blur (and Enter). Toast on failure. |
| Description | Save on editor blur / debounce idle (match panel today). |
| Property sidebar | Keep emit-on-change → parent PATCH; toast on failure. |
| Full-page header | Remove primary **Save** button. Watch / copy link / delete remain. |

`IssueDetail.vue` must wire title into the same path as `IssueDetailPanel` (today full-page title is effectively read-only because `IssueTabDetails` title is gated and the header is static).

### 2. URL and navigation

Canonical routes use **slug**:

- Project hub: `/workspace/:slug/project/:id`
- Full detail: `/workspace/:slug/project/:id/issues/:issueId`

Project hub query params:

| Param | Meaning |
|-------|---------|
| `view` | `list` \| `kanban` \| `tree` \| `calendar` \| `gantt` — written on view change, read on mount |
| `issue` | Numeric issue id — opens `IssueDetailPanel`; cleared when panel closes |

Rules:

- Selecting an issue from list/kanban/tree sets `?issue=` (replace or push — prefer **push** so browser Back closes the panel).
- Closing the panel (X / Esc) removes `issue` from the query.
- On mount, if `issue` is present, open the panel for that id.
- Create / “New issue” links and post-create redirects use slug routes (eliminate mixed `/workspaces/:workspaceId/...` jumps from Project, IssueList, IssueCreate where those surfaces are touched).

### 3. Full-page ↔ panel parity

| Capability | Full page | Panel (this round) |
|------------|-----------|--------------------|
| Title edit + instant save | Yes | Yes (already) |
| Copy link | Yes | **Add** |
| Watch / unwatch | Yes | **Add** |
| Delete | Yes | **Add** (confirm via `useConfirm`) |
| Property sidebar | Shared | Shared |
| AI / Chat / Git tabs | Keep | **Not** added |
| State list filtering | Filter inactive / E2E | **Match** full page |
| Approval dialogs | `useConfirm` / toast | Same (replace `window.confirm` / `alert`) |
| Relation summary in sidebar | Load on detail open | Load on panel open (do not wait for Relations tab mount) |
| Agent assign | Explicit button only | Explicit button only (remove auto-assign on selector change) |

Release field respects `isLocked` like other properties.

### 4. Property sidebar — compact rows (option A)

Visual structure (single column, ~280px):

```
[ Relations summary — if any ]

── Status & people ──
State          [ select ]
Priority       [ select ]
Assignee       [ select ]

── Planning ──
Cycle          [ select ]
Module         [ select ]
Release        [ select ]

── Dates & labels ──
Start          [ date ]
Target         [ date ]
Labels         [ LabelSelector ]

── Agent ──
… existing agent block …

── Custom fields ──
Field name *   [ control ]
```

Implementation notes:

- Row layout: left label (`text-xs text-gray-500`), right control (`flex-1 min-w-0`).
- Reduce “card-in-card”: light section dividers instead of heavy nested boxes where possible.
- Keep existing control types (native select / date / text); no new Popovers.
- Required custom fields show `*`.
- Shared by `IssueDetail`, `IssueDetailPanel`, and preferably `IssueCreate` right column if low-cost; if create reuse slips, create page can stay as-is this round (not a blocker).

### 5. Files expected to change

- `frontend/src/components/IssuePropertySidebar.vue` — row layout, lock Release, agent watch fix
- `frontend/src/views/IssueDetail.vue` — instant title/description, remove Save, relation summary load, approval confirm
- `frontend/src/components/IssueDetailPanel.vue` — header actions, state filter, URL sync props, approval confirm
- `frontend/src/components/IssueDetailHeader.vue` — editable title; drop/repurpose Save
- `frontend/src/views/Project.vue` — `view` + `issue` query sync; Esc closes panel
- `frontend/src/views/IssueCreate.vue` / `IssueList.vue` — slug redirects / create links as touched
- i18n keys only if new strings (copy link on panel, etc.)

### 6. Acceptance criteria

1. Full-page issue title can be edited and persists without a primary Save button.
2. Description changes persist without Save.
3. Opening an issue from the project hub sets `?issue=<id>`; refresh reopens the panel; Back/close clears it.
4. Changing list/kanban/… view updates `?view=` and survives refresh.
5. Panel header supports copy link, watch, and delete.
6. Property sidebar renders as compact label|value rows; Release disabled when approval-locked.
7. Choosing an agent in the selector does **not** assign until the Assign button is clicked.
8. Sidebar relation summary appears without visiting the Relations tab first.
9. No new `/workspaces/:numericId/...` links introduced from the touched create/list entry points; existing touched ones migrate to slug.

### 7. Follow-ups (explicitly deferred)

- Option B: Linear-style property Popovers.
- Panel AI / Chat / Git tabs.
- List `j/k` + Enter open / Esc close.
- Deduplicate remaining shared logic between `IssueDetail` and `IssueDetailPanel` into a composable (optional cleanup if time remains).

## Testing

- Manual: project hub list → open panel → copy URL in new tab → panel opens; Back closes.
- Manual: full-page title + description + sidebar field each persist after reload.
- Manual: pending-approval issue — Release and other fields locked; state select still follows existing approval UX.
- Manual: agent select then Assign — only Assign calls API.
- Smoke: create issue from project hub still lands on correct slug project/detail.

No new backend APIs required for this round.
