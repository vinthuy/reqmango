# Kanban Board Drag-and-Drop Enhancement

**Date:** 2026-07-05
**Status:** Approved

## Overview

Replace native HTML5 drag-and-drop in `IssueKanban.vue` with `vue-draggable-plus` (SortableJS wrapper), adding: smooth animations, column-internal ordering with sortOrder persistence, and cross-column drag for all groupBy modes.

## Architecture

### Component Changes

Primary: `IssueKanban.vue`
- Replace `draggable="true"` + `@dragstart/@dragOver/@drop` handlers with `<VueDraggable>` wrapping each column's card list
- Each column's `<VueDraggable>` shares a common `group` config so cards can move across columns
- `onUpdate` callback handles same-column reorder → compute new `sortOrder` → API
- `onChange` callback handles cross-column move → update the appropriate field per groupBy mode + sortOrder → API

Supporting files:
- `frontend/src/api/issue.ts`: `updateIssue` accepts `sort_order` parameter
- `frontend/src/types/issue.ts`: `IssueUpdate` type adds `sort_order?: number`
- Backend: Confirm `PUT /issues/:issueId` handler parses `sort_order`; add field if missing

### Library

`vue-draggable-plus` (^0.6.1) — already in `package.json`, not yet imported.

## Data Flow

### onUpdate (same-column reorder)

```
User drags card within same column
  → VueDraggable v-model mutates the local array
  → onUpdate(newIndex, oldIndex, from, to)
  → Compute sortOrder = midpoint(prev, next)
  → AbortController cancels any pending update for this issue
  → PUT /issues/:issueId { sort_order: computedValue }
  → On failure: revert to saved snapshot + toast "更新失败，已恢复原位"
```

### onChange (cross-column move)

```
User drags card to a different column
  → VueDraggable v-model mutates both local arrays
  → onChange(evt)
  → Determine update field from current groupBy:
      state    → { state_id: targetKey }
      assignee → { assignee_ids: targetKey === 'unassigned' ? [] : [targetKey] }
      priority → { priority: targetKey }
      labels   → POST /issues/:issueId/labels { label_id: targetKey } (skip if already present)
                  "no-label" → skip label API, only sortOrder
  → Compute sortOrder in target column
  → PUT /issues/:issueId { ...updateField, sort_order: computedValue }
  → On failure: revert + toast
```

### GroupBy Mapping

| groupBy | Target column key | API field updated |
|---|---|---|
| state | state_id | `state_id` |
| assignee | user_id / `unassigned` | `assignee_ids` |
| priority | priority value string | `priority` |
| labels | label_id / `no-label` | Add label via `POST /issues/:id/labels` (skip if duplicate). "No Label" column = no-op for labels, only sortOrder updated. Two API calls: label POST first, then sortOrder PUT. |

## SortOrder Calculation

Fractional indexing, backed by `sort_order` field (float64, default 65535):

| Position | Formula |
|---|---|
| Between two cards | `(prev.sortOrder + next.sortOrder) / 2` |
| Top of non-empty column | `first.sortOrder / 2` |
| Bottom of non-empty column | `last.sortOrder + 1000` |
| Only card in column | `65535` (default) |

Rebalance trigger: when gap < 0.001, renumber entire column with spacing of 1000 (1000, 2000, 3000...), via `POST /issues/bulk/update`.

## Visual Feedback

- **Drag handle:** 6-dot grip icon on left edge of card, visible on hover, `cursor: grab`
- **Ghost:** SortableJS default semi-transparent clone
- **Placeholder:** Empty card-height slot with dashed border in source position
- **Animation:** `animation: 300` for smooth settle transitions
- **Column hover:** Target column background shifts to `bg-blue-50` via SortableJS ghost class or `onMove` callback
- **Touch:** SortableJS native touch support via `forceFallback: false`

## Swimlane Support

Each cell in the swimlane grid gets its own `<VueDraggable>`. All cells share a common `group` name so cards can move across swimlanes and columns freely. `onChange` detects both swimlane and column changes from the group context.

## Error Handling

| Scenario | Handling |
|---|---|
| API failure | Revert to pre-drag snapshot, toast "更新失败，已恢复原位", lock drag 500ms |
| Rapid successive drags | `AbortController` cancels pending request for same issue |
| Drag to same position | `moved.oldIndex === moved.newIndex` && same column → skip API call |
| Empty column target | sortOrder = 65535, normal update |
| Rebalance ongoing | Column shows loading overlay (semi-transparent + spinner), drag disabled until complete |
| Two users drag same card | Last-write-wins (no optimistic locking for v1) |

## Backend Changes

- `PUT /issues/:issueId` handler: confirm `sort_order` float field is parsed from request body and applied to the model
- If `sort_order` not yet in the update struct, add it

## Testing

- Unit/manual: drag within column (top, middle, bottom), drag between columns, drag across swimlanes
- Verify sortOrder persistence via API response
- Verify abort on rapid drag
- Verify revert on simulated API failure
- Existing E2E tests in `tests/issue-kanban.test.ts` should continue to pass (update selectors if DOM structure changes)
