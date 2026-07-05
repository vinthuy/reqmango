# Kanban Board Drag-and-Drop Enhancement — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace native HTML5 drag-and-drop in `IssueKanban.vue` with `vue-draggable-plus`, adding smooth animations, column-internal ordering with sortOrder persistence, and cross-column drag for all groupBy modes.

**Architecture:** Backend: add `sort_order` to `IssueUpdateRequest` and `BulkUpdateRequest` DTOs, wire into service. Frontend: convert `groupedIssues`/`swimlaneGroupedIssues` from `computed` to `ref` for VueDraggable v-model compatibility; use `@update` (same-column) and `@add` (cross-column) events; compute sortOrder via fractional indexing; handle cross-column field updates per groupBy mode.

**Tech Stack:** Go (backend DTOs + service), Vue 3 + TypeScript + `vue-draggable-plus` ^0.6.1 (frontend)

---

### Task 1: Add `SortOrder` to backend request DTOs

**Files:**
- Modify: `backend/internal/dto/request/issue.go`

- [ ] **Step 1: Add `SortOrder` to `IssueUpdateRequest` struct**

At line 29 (after `StateID`), insert:

```go
SortOrder       *float64 `json:"sort_order"`
```

Full resulting struct:

```go
type IssueUpdateRequest struct {
	Name            *string  `json:"name"`
	DescriptionHTML *string  `json:"description_html"`
	DescriptionJSON *string  `json:"description_json"`
	Priority        *string  `json:"priority"`
	StateID         *uint64  `json:"state_id"`
	SortOrder       *float64 `json:"sort_order"`
	AssigneeIDs     []uint64 `json:"assignee_ids"`
	LabelIDs        []uint64 `json:"label_ids"`
	StartDate       *string  `json:"start_date"`
	TargetDate      *string  `json:"target_date"`
	EstimatePointID *uint64  `json:"estimate_point_id"`
	CycleID         *uint64  `json:"cycle_id"`
	ModuleIDs       []uint64 `json:"module_ids"`
	ParentID        *uint64  `json:"parent_id"`
	TypeID          *uint64  `json:"type_id"`
	CoverImageURL   *string  `json:"cover_image_url"`
}
```

- [ ] **Step 2: Add `SortOrder` to `BulkUpdateRequest` struct**

At line 46 (after `StateID`), insert:

```go
SortOrder   *float64 `json:"sort_order"`
```

Full resulting struct:

```go
type BulkUpdateRequest struct {
	IssueIDs    []uint64 `json:"issue_ids" binding:"required"`
	Priority    *string  `json:"priority"`
	StateID     *uint64  `json:"state_id"`
	SortOrder   *float64 `json:"sort_order"`
	AssigneeIDs []uint64 `json:"assignee_ids"`
	LabelIDs    []uint64 `json:"label_ids"`
	StartDate   *string  `json:"start_date"`
	TargetDate  *string  `json:"target_date"`
}
```

- [ ] **Step 3: Verify compilation**

Run: `cd backend && go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add backend/internal/dto/request/issue.go
git commit -m "feat: add sort_order to IssueUpdateRequest and BulkUpdateRequest DTOs"
```

---

### Task 2: Wire `SortOrder` into backend Update and BulkUpdate service methods

**Files:**
- Modify: `backend/internal/service/issue_service.go`

- [ ] **Step 1: Read the BulkUpdate method to understand its update pattern**

```bash
grep -n "BulkUpdate\|bulkUpdate" backend/internal/service/issue_service.go
```

Then read around line 1018 to see how fields are applied.

- [ ] **Step 2: Add `SortOrder` handling in the `Update` method**

In the `Update` method (~line 533), after the `StateID` block ends (after line 608, the closing `}` of `if req.StateID != nil`), and before `// Save basic fields`, insert:

```go
if req.SortOrder != nil && *req.SortOrder != issue.SortOrder {
    issue.SortOrder = *req.SortOrder
    hasChanges = true
}
```

- [ ] **Step 3: Add `SortOrder` handling in the `BulkUpdate` method**

In the `BulkUpdate` method body, find the pattern used to apply `StateID` and add a matching block for `SortOrder`. The method likely uses either direct field assignment or an `updates` map. If it uses `updates`:

```go
if req.SortOrder != nil {
    updates["sort_order"] = *req.SortOrder
}
```

If it uses individual field assignments on each issue:

```go
if req.SortOrder != nil {
    issue.SortOrder = *req.SortOrder
}
```

- [ ] **Step 4: Verify compilation**

Run: `cd backend && go build ./...`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/issue_service.go
git commit -m "feat: wire sort_order into Update and BulkUpdate service methods"
```

---

### Task 3: Add `sort_order` to frontend types and API

**Files:**
- Modify: `frontend/src/types/issue.ts`
- Modify: `frontend/src/api/issue.ts`

- [ ] **Step 1: Add `sort_order` to `IssueUpdate` interface**

In `frontend/src/types/issue.ts`, at line 65 (after `state_id`), insert:

```typescript
sort_order?: number
```

- [ ] **Step 2: Add `sort_order` to `updateIssue` function signature (if needed)**

The `updateIssue` function already passes `data: IssueUpdate` as the body, so adding `sort_order` to the `IssueUpdate` type is sufficient. No change needed to `issue.ts`.

- [ ] **Step 3: Verify types compile**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 | head -30`
Expected: No new errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/types/issue.ts
git commit -m "feat: add sort_order to frontend IssueUpdate type"
```

---

### Task 4: Rewrite Kanban drag-and-drop with vue-draggable-plus

**Files:**
- Modify: `frontend/src/components/IssueKanban.vue`

This is the core task. The key architectural change: `groupedIssues` and `swimlaneGroupedIssues` must change from `computed` to `ref` because VueDraggable's `v-model` requires a mutable reactive array.

- [ ] **Step 1: Add import for VueDraggable**

After the existing imports (line 188), add:

```typescript
import { VueDraggable } from 'vue-draggable-plus'
```

- [ ] **Step 2: Convert `groupedIssues` from `computed` to `ref` + sync function**

**Remove** the existing `const groupedIssues = computed(...)` block (lines 259-315).

**Replace with:**

```typescript
const groupedIssues = ref<Record<string | number, any[]>>({})

function rebuildGroupedIssues() {
  const map: Record<string | number, any[]> = {}

  if (groupBy.value === 'state') {
    states.value.forEach(s => { map[s.id] = [] })
    issues.value.forEach(i => {
      if (map[i.state_id]) map[i.state_id].push(i)
    })
  } else if (groupBy.value === 'assignee') {
    issues.value.forEach(i => {
      if (i.assignees && i.assignees.length > 0) {
        i.assignees.forEach((a: any) => {
          if (!map[a.id]) map[a.id] = []
        })
      }
    })
    map['__unassigned__'] = []
    issues.value.forEach(i => {
      if (i.assignees && i.assignees.length > 0) {
        i.assignees.forEach((a: any) => {
          map[a.id].push(i)
        })
      } else {
        map['__unassigned__'].push(i)
      }
    })
  } else if (groupBy.value === 'priority') {
    ['urgent', 'high', 'medium', 'low', 'none'].forEach(k => { map[k] = [] })
    issues.value.forEach(i => {
      const key = i.priority || 'none'
      if (map[key]) map[key].push(i)
    })
  } else if (groupBy.value === 'labels') {
    issues.value.forEach(i => {
      if (i.labels && i.labels.length > 0) {
        i.labels.forEach((l: any) => {
          if (!map[l.id]) map[l.id] = []
        })
      }
    })
    map['__nolabel__'] = []
    issues.value.forEach(i => {
      if (i.labels && i.labels.length > 0) {
        i.labels.forEach((l: any) => {
          map[l.id].push(i)
        })
      } else {
        map['__nolabel__'].push(i)
      }
    })
  }

  groupedIssues.value = map
}
```

- [ ] **Step 3: Convert `swimlaneGroupedIssues` from `computed` to `ref` + sync function**

**Remove** the existing `const swimlaneGroupedIssues = computed(...)` block (lines 444-487).

**Replace with:**

```typescript
const swimlaneGrouped = ref<Record<string, Record<string | number, any[]>>>({})

function rebuildSwimlaneGrouped() {
  if (!swimlaneBy.value) {
    swimlaneGrouped.value = {}
    return
  }

  const result: Record<string, Record<string | number, any[]>> = {}
  swimlaneKeys.value.forEach(s => { result[s.key] = {} })
  // Pre-initialize all state columns for state-based grouping
  if (groupBy.value === 'state') {
    states.value.forEach(s => {
      swimlaneKeys.value.forEach(sk => { result[sk.key][s.id] = [] })
    })
  }

  issues.value.forEach(i => {
    const sk = getSwimlaneKeyForIssue(i)
    if (!result[sk]) result[sk] = {}

    if (groupBy.value === 'state') {
      if (!result[sk][i.state_id]) result[sk][i.state_id] = []
      result[sk][i.state_id].push(i)
    } else if (groupBy.value === 'assignee') {
      if (i.assignees && i.assignees.length > 0) {
        i.assignees.forEach((a: any) => {
          if (!result[sk][a.id]) result[sk][a.id] = []
          result[sk][a.id].push(i)
        })
      } else {
        if (!result[sk]['__unassigned__']) result[sk]['__unassigned__'] = []
        result[sk]['__unassigned__'].push(i)
      }
    } else if (groupBy.value === 'priority') {
      const k = i.priority || 'none'
      if (!result[sk][k]) result[sk][k] = []
      result[sk][k].push(i)
    } else if (groupBy.value === 'labels') {
      if (i.labels && i.labels.length > 0) {
        i.labels.forEach((l: any) => {
          if (!result[sk][l.id]) result[sk][l.id] = []
          result[sk][l.id].push(i)
        })
      } else {
        if (!result[sk]['__nolabel__']) result[sk]['__nolabel__'] = []
        result[sk]['__nolabel__'].push(i)
      }
    }
  })

  swimlaneGrouped.value = result
}
```

- [ ] **Step 4: Update `countSwimlaneIssues` to use new ref name**

Update the `countSwimlaneIssues` function (line 489) to use `swimlaneGrouped` instead of `swimlaneGroupedIssues`:

```typescript
function countSwimlaneIssues(swimlaneKey: string): number {
  if (!swimlaneGrouped.value) return 0
  const cols = swimlaneGrouped.value[swimlaneKey]
  if (!cols) return 0
  let count = 0
  for (const k of Object.keys(cols)) {
    count += (cols[k] || []).length
  }
  return count
}
```

- [ ] **Step 5: Call rebuild functions after loading issues**

In the `loadIssues` function (line 543), after `issues.value = result.items`, add:

```typescript
rebuildGroupedIssues()
rebuildSwimlaneGrouped()
```

Also add watchers to rebuild when grouping changes:

```typescript
watch(groupBy, () => { rebuildGroupedIssues(); rebuildSwimlaneGrouped() })
watch(swimlaneBy, () => rebuildSwimlaneGrouped())
```

- [ ] **Step 6: Add drag state and sortOrder computation helper**

After the ref declarations (around line 201, after the watch blocks added above), add:

```typescript
const dragLocked = ref(false)
const pendingRequests = new Map<number, AbortController>()

function computeSortOrder(targetList: any[], newIndex: number): number {
  const prev = newIndex > 0 ? (targetList[newIndex - 1]?.sort_order ?? null) : null
  const next = newIndex < targetList.length - 1 ? (targetList[newIndex + 1]?.sort_order ?? null) : null

  if (!prev && !next) return 65535
  if (!prev && next) return (next as number) / 2
  if (prev && !next) return (prev as number) + 1000
  return ((prev as number) + (next as number)) / 2
}
```

- [ ] **Step 7: Add the drag group name computed**

```typescript
const dragGroupName = computed(() => String(props.projectId))
```

- [ ] **Step 8: Add the `@update` handler (same-column reorder)**

```typescript
async function onDragUpdate(columnKey: string | number, newIndex: number, oldIndex: number, itemEl: any) {
  if (dragLocked.value || newIndex === oldIndex) return

  // Extract issue id — vue-draggable-plus stores item data on the element or wrapper
  const issueId = (itemEl as any)?.__vueDraggableData?.id || (itemEl as any)?.id
  if (!issueId) return

  const prevCtrl = pendingRequests.get(issueId)
  if (prevCtrl) prevCtrl.abort()
  pendingRequests.set(issueId, new AbortController())

  const columnIssues = groupedIssues.value[columnKey]
  if (!columnIssues) return
  const sortOrder = computeSortOrder(columnIssues, newIndex)

  try {
    await issueApi.updateIssue(issueId, { sort_order: sortOrder })
  } catch (err: any) {
    if (err?.name === 'CanceledError' || err?.code === 'ERR_CANCELED') return
    await loadIssues()
    showToast('更新失败，已恢复原位')
  } finally {
    pendingRequests.delete(issueId)
  }
}
```

- [ ] **Step 9: Add the `@add` handler (cross-column move)**

```typescript
async function onDragAdd(columnKey: string | number, newIndex: number, itemEl: any) {
  if (dragLocked.value) return

  const issueId = (itemEl as any)?.__vueDraggableData?.id || (itemEl as any)?.id
  if (!issueId) return

  const prevCtrl = pendingRequests.get(issueId)
  if (prevCtrl) prevCtrl.abort()
  pendingRequests.set(issueId, new AbortController())

  try {
    // Step 1: Update field based on groupBy
    if (groupBy.value === 'state') {
      await issueApi.updateIssue(issueId, { state_id: columnKey as number })
    } else if (groupBy.value === 'assignee') {
      const assigneeIds = columnKey === '__unassigned__' ? [] : [columnKey as number]
      await issueApi.updateIssue(issueId, { assignee_ids: assigneeIds })
    } else if (groupBy.value === 'priority') {
      await issueApi.updateIssue(issueId, { priority: columnKey as string as any })
    } else if (groupBy.value === 'labels') {
      if (columnKey !== '__nolabel__') {
        try { await issueApi.addIssueLabel(issueId, columnKey as number) } catch { /* already exists */ }
      }
    }

    // Step 2: Update sortOrder in target column
    const columnIssues = groupedIssues.value[columnKey]
    if (columnIssues) {
      const sortOrder = computeSortOrder(columnIssues, newIndex)
      await issueApi.updateIssue(issueId, { sort_order: sortOrder })
    }
  } catch (err: any) {
    if (err?.name === 'CanceledError' || err?.code === 'ERR_CANCELED') return
    await loadIssues()
    showToast('更新失败，已恢复原位')
    dragLocked.value = true
    setTimeout(() => { dragLocked.value = false }, 500)
  } finally {
    pendingRequests.delete(issueId)
  }
}
```

- [ ] **Step 10: Rewrite the non-swimlane template**

Replace lines 76-118 (the `v-else-if="!swimlaneBy"` block) with:

```html
<!-- 无泳道模式 -->
<div v-else-if="!swimlaneBy" class="grid gap-4" :style="gridStyle">
  <div
    v-for="column in kanbanColumns"
    :key="column.id"
    class="bg-gray-100 dark:bg-gray-800 rounded-lg p-3 min-h-[200px]"
  >
    <div class="flex items-center justify-between mb-3">
      <div class="flex items-center space-x-2">
        <span class="w-2.5 h-2.5 rounded-full" :style="{ backgroundColor: column.color }"></span>
        <h3 class="text-sm font-medium text-gray-700">{{ column.label }}</h3>
      </div>
      <div class="flex items-center space-x-1">
        <button v-if="groupBy === 'state'" @click="openQuickCreate(column.key as number)" class="w-5 h-5 flex items-center justify-center text-gray-500 hover:text-indigo-600 hover:bg-gray-200 rounded text-sm" :title="t('issueKanban.quickCreate')">+</button>
        <span class="text-xs bg-gray-300 text-gray-600 px-1.5 py-0.5 rounded-full">{{ (groupedIssues[column.key] || []).length }}</span>
      </div>
    </div>
    <div v-if="groupBy === 'state' && quickCreateStateId === column.key" class="mb-3">
      <QuickCreateInput :project-id="projectId" :workspace-id="workspaceId" :issue-types="issueTypes" :default-state-id="(column.key as number)" :show-priority="false" inline show-cancel @created="onQuickCreated" @cancel="closeQuickCreate" />
    </div>
    <VueDraggable
      v-model="groupedIssues[column.key]"
      :group="{ name: dragGroupName, pull: true, put: true }"
      :animation="300"
      ghost-class="kanban-ghost"
      drag-class="kanban-drag"
      :sort="true"
      :disabled="dragLocked"
      item-key="id"
      class="space-y-2 min-h-[4px]"
      @update="(evt: any) => onDragUpdate(column.key, evt.newIndex, evt.oldIndex, evt.item)"
      @add="(evt: any) => onDragAdd(column.key, evt.newIndex, evt.item)"
    >
      <div
        v-for="issue in groupedIssues[column.key] || []"
        :key="issue.id"
        @click="$emit('select', issue)"
        class="bg-white dark:bg-gray-700 rounded-md border border-gray-200 dark:border-gray-600 p-2.5 cursor-pointer hover:shadow-md hover:border-indigo-300 dark:hover:border-indigo-600 transition-shadow relative group"
      >
        <div class="absolute top-1.5 right-1.5 opacity-0 group-hover:opacity-100 transition-opacity" @click.stop>
          <input type="checkbox" :checked="selectedIds.has(issue.id)" @change="toggleSelect(issue.id)" class="rounded border-gray-300 dark:border-gray-500 w-3.5 h-3.5" />
        </div>
        <div class="flex items-center gap-2">
          <div class="flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity cursor-grab text-gray-400" title="Drag to reorder">
            <svg class="w-3 h-3" viewBox="0 0 12 12" fill="currentColor"><circle cx="3" cy="2" r="1.2"/><circle cx="9" cy="2" r="1.2"/><circle cx="3" cy="6" r="1.2"/><circle cx="9" cy="6" r="1.2"/><circle cx="3" cy="10" r="1.2"/><circle cx="9" cy="10" r="1.2"/></svg>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-start justify-between">
              <span class="text-xs text-gray-400 font-mono">{{ projectIdentifier }}-{{ issue.sequence_id }}</span>
              <span :class="priorityDotClass(issue.priority)" class="w-1.5 h-1.5 rounded-full inline-block"></span>
            </div>
            <p class="text-sm text-gray-800 dark:text-gray-100 mt-1 leading-snug line-clamp-2">{{ issue.name }}</p>
            <div class="mt-2 flex items-center justify-between">
              <div class="flex -space-x-1">
                <div v-for="(a, idx) in (issue.assignees || []).slice(0, 2)" :key="a.id" class="w-5 h-5 rounded-full border border-white flex items-center justify-center text-[10px] font-medium text-white" :style="{ backgroundColor: assigneeColor(idx) }" :title="a.display_name || a.username">{{ getInitials(a.display_name || a.username) }}</div>
              </div>
              <span v-if="issue.cycle" class="text-[10px] text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded">{{ issue.cycle.name }}</span>
              <button @click.stop="$emit('select', issue)" class="text-[10px] text-indigo-500 hover:text-indigo-700 font-medium">{{ t('issueKanban.details') }}</button>
            </div>
          </div>
        </div>
      </div>
    </VueDraggable>
    <div v-if="!(groupedIssues[column.key] || []).length" class="text-center text-xs text-gray-400 py-6">{{ t('issueKanban.dragHere') }}</div>
  </div>
</div>
```

- [ ] **Step 11: Rewrite the swimlane template**

Replace lines 121-168 (the `v-else` swimlane block) with:

```html
<!-- 泳道模式 -->
<div v-else class="swimlane-board space-y-6">
  <div
    v-for="swimlane in swimlaneKeys"
    :key="swimlane.key"
    class="swimlane-row"
  >
    <div class="flex items-center space-x-2 mb-2 px-1">
      <span class="w-3 h-3 rounded-full flex-shrink-0" :style="{ backgroundColor: swimlane.color }"></span>
      <span class="text-sm font-semibold text-gray-700 dark:text-gray-200">{{ swimlane.label }}</span>
      <span class="text-xs text-gray-400">{{ countSwimlaneIssues(swimlane.key) }} {{ t('common.items') }}</span>
    </div>
    <div class="grid gap-3" :style="gridStyle">
      <div
        v-for="column in kanbanColumns"
        :key="column.id"
        class="bg-gray-100 dark:bg-gray-800 rounded-lg p-3 min-h-[120px]"
      >
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center space-x-1.5">
            <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: column.color }"></span>
            <h4 class="text-xs font-medium text-gray-500">{{ column.label }}</h4>
          </div>
          <span class="text-[10px] bg-gray-300 text-gray-600 px-1.5 py-0.5 rounded-full">
            {{ (swimlaneGrouped[swimlane.key]?.[column.key] || []).length }}
          </span>
        </div>
        <VueDraggable
          v-if="swimlaneGrouped[swimlane.key]"
          v-model="swimlaneGrouped[swimlane.key][column.key]"
          :group="{ name: dragGroupName, pull: true, put: true }"
          :animation="300"
          ghost-class="kanban-ghost"
          drag-class="kanban-drag"
          :sort="true"
          :disabled="dragLocked"
          item-key="id"
          class="space-y-1.5 min-h-[4px]"
          @update="(evt: any) => onDragUpdate(column.key, evt.newIndex, evt.oldIndex, evt.item)"
          @add="(evt: any) => onDragAdd(column.key, evt.newIndex, evt.item)"
        >
          <div
            v-for="issue in swimlaneGrouped[swimlane.key]?.[column.key] || []"
            :key="issue.id"
            @click="$emit('select', issue)"
            class="bg-white dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600 p-2 cursor-pointer hover:shadow-md hover:border-indigo-300 dark:hover:border-indigo-600 transition-shadow relative group"
          >
            <div class="flex items-center gap-1.5">
              <div class="flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity cursor-grab text-gray-400" title="Drag to reorder">
                <svg class="w-2.5 h-2.5" viewBox="0 0 12 12" fill="currentColor"><circle cx="3" cy="2" r="1.2"/><circle cx="9" cy="2" r="1.2"/><circle cx="3" cy="6" r="1.2"/><circle cx="9" cy="6" r="1.2"/><circle cx="3" cy="10" r="1.2"/><circle cx="9" cy="10" r="1.2"/></svg>
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-start justify-between">
                  <span class="text-[10px] text-gray-400 font-mono">{{ issue.sequence_id }}</span>
                  <span :class="priorityDotClass(issue.priority)" class="w-1.5 h-1.5 rounded-full inline-block"></span>
                </div>
                <p class="text-xs text-gray-800 dark:text-gray-100 mt-0.5 leading-snug line-clamp-2">{{ issue.name }}</p>
              </div>
            </div>
          </div>
        </VueDraggable>
        <div v-if="!(swimlaneGrouped[swimlane.key]?.[column.key] || []).length" class="text-center text-[10px] text-gray-300 py-4">-</div>
      </div>
    </div>
  </div>
</div>
```

- [ ] **Step 12: Remove old DnD handlers and `@dragover`/`@drop` attributes**

Remove the three functions (lines 507-532):
- `onDragStart`
- `onDragOver`
- `onDrop`

The old `draggable="true"`, `@dragstart`, `@dragover`, and `@drop` attributes in the template have already been replaced by steps 10 and 11.

- [ ] **Step 13: Add drag-related CSS**

At the bottom of the SFC (after `</script>` and before the final `</template>` — wait, the file ends with `</script>`, so add a `<style scoped>` block after `</script>`):

```css
<style scoped>
.kanban-ghost {
  opacity: 0.4;
  background-color: #dbeafe;
  border: 2px dashed #60a5fa;
  border-radius: 0.5rem;
}

.kanban-drag {
  opacity: 0.6;
  transform: rotate(2deg);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}
</style>
```

(Using plain CSS instead of `@apply` to avoid Tailwind v4 vs v3 compatibility issues with `@apply` in scoped styles.)

- [ ] **Step 14: Verify frontend compiles**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 | head -40`
Expected: No new errors

- [ ] **Step 15: Commit**

```bash
git add frontend/src/components/IssueKanban.vue
git commit -m "feat: rewrite kanban drag-and-drop with vue-draggable-plus"
```

---

### Task 5: Verify end-to-end and fix issues

**Files:**
- None (verification + fixups)

- [ ] **Step 1: Build backend**

Run: `cd backend && go build ./...`
Expected: Success

- [ ] **Step 2: Typecheck frontend**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: No new errors

- [ ] **Step 3: Run existing E2E tests**

Run: `cd frontend && npx playwright test tests/issue-kanban.test.ts`
Expected: Tests pass (if selectors changed due to template rewrite, update selectors in this step)

- [ ] **Step 4: Manual smoke test matrix**

Start dev server: `cd frontend && npm run dev`

| Scenario | Steps | Expected |
|---|---|---|
| Drag within column | Drag card from position 1 to pos 3 in same column | Animated reorder, sort_order persisted |
| Drag between states | Drag card "Todo" → "In Progress" | state_id updated, card moved |
| Drag between assignees | Group by assignee, drag to another member | assignee_ids updated |
| Drag between priorities | Group by priority, drag "High" → "Medium" | priority updated |
| Drag between labels | Group by labels, drag to "Bug" column | Label added |
| Empty column target | Drag to empty column | sortOrder=65535, card appears |
| Rapid drags | Drag same card 3× in quick succession | Only last result persists |
| Swimlane drag | Enable swimlane, drag across swimlanes+columns | Field + sortOrder correct |
| Grip handle hover | Hover left edge of card | 6-dot grip visible, cursor=grab |

- [ ] **Step 5: Commit fixups**

```bash
git add -A
git commit -m "chore: e2e verification adjustments for kanban drag-and-drop"
```
