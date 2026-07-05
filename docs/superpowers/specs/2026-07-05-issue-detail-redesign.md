# Issue Detail Page Redesign

**Date:** 2026-07-05
**Status:** Design Approved

## Overview

Rewrite the issue detail page (`IssueDetail.vue`) from a single flat scrollable page into a tab-based layout with clean component decomposition. Reference Plane AI design for visual structure. All relationships (parent, sub-issues, linked issues) are consolidated into a single "Relations" tab with card-based organization.

## Architecture

### Component Tree

```
IssueDetail.vue                    ← Route entry, orchestrates sub-components
├── IssueDetailHeader.vue          ← Top bar: back nav, type badge, ID, save button
├── IssuePropertySidebar.vue       ← Right sidebar: state/priority/assignee/labels/dates/AI agent
├── IssueTabDetails.vue            ← Tab 1: Description + Custom Fields grid + Comments card
├── IssueTabRelations.vue          ← Tab 2: Parent + Sub-issues + Linked Issues
│   ├── RelationParentCard.vue     ←   Parent relationship card
│   ├── RelationSubIssuesCard.vue  ←   Sub-issues checklist card
│   └── RelationTypeCard.vue       ←   Dynamic relation type card (reusable)
├── IssueTabAttachments.vue        ← Tab 3: File attachments
├── IssueTabTimeTracking.vue       ← Tab 4: Time tracking logs
└── IssueTabActivity.vue           ← Tab 5: System activity log
```

### Tab Structure (5 Tabs)

| Tab | Content | Component | Source |
|---|---|---|---|
| Details | Description + Custom Fields grid + Comments card | IssueTabDetails | Integrate RichTextEditor, CustomFieldManager, CommentList |
| Relations | Parent card + Sub-issues card + dynamic Link type cards | IssueTabRelations | Merge old SubIssuePanel + relation panel |
| Attachments | File list + upload | IssueTabAttachments | Reuse AttachmentManager |
| Time Tracking | Work log entries | IssueTabTimeTracking | Reuse TimeTrackPanel |
| Activity | System event timeline | IssueTabActivity | Existing activity logic |

### Right Sidebar (IssuePropertySidebar)

Properties (vertical order):
- State (select)
- Priority (select)
- Assignee (select)
- Labels (LabelSelector)
- Cycle (select)
- Module (select)
- Start Date (date input)
- Target Date (date input)
- AI Agent (AgentSelector + Dispatch button)
- Recurrence (RecurrenceConfig - collapsed by default)

## Relations Tab Design

### 3 Card Types

1. **Parent** — Always rendered. Shows parent issue row with system fields. Empty state: "+ Set Parent" button when no parent exists. Actions: Change, Remove.

2. **Sub-issues** — Always rendered. Checklist table with column headers. Completion counter (e.g., "2/5"). Toggle expand per row for details. Actions: + Add, checkbox toggle.

3. **Linked Issues** — One card per relation type (blocks, relates to, duplicates, is blocked by, etc.). Dynamically generated from `relationTypes` data. Only types with linked issues are shown. Card header: relation outward_name + count badge. Actions: + Link, per-row Remove.

### Relation Row Fields

Each row (Parent / Sub-issue / Linked item) shows these system fields:

| Field | Width | Notes |
|---|---|---|
| Type badge | fix | Task/Bug/Feature/Epic — colored badge |
| ID | fix | Project identifier-sequeno, e.g., DEV-43 |
| Title | flex | Truncated with ellipsis |
| State | fix | State name with color dot |
| Priority | fix | Priority color dot + text |
| Assignee | fix | Display name or "—" |
| Due Date | fix | MM/DD format or "—" |
| Action | fix | Remove button (×) |

Sub-issues have a leading checkbox column and visible column headers.

### Responsive Handling

When viewport is narrow (< 768px), hide Assignee and Due Date columns first. Keep Type/ID/Title/State/Priority always visible.

## Details Tab Design

### 3 Card Sections (Vertical)

1. **Description** — Full-width RichTextEditor. Editable with debounced auto-save.

2. **Custom Fields** — 2-column grid layout. Each field shows label + value/input. Types: text, number, date, boolean (switch), dropdown (single/multi-select), member (user picker).

3. **Comments** — Bottom card with threaded comment list + input. Shows author avatar, timestamp, content. New comment input at bottom.

## Data Flow

### State Management

- `IssueDetail.vue` fetches issue data via `getIssue(issueId)` on mount
- Holds reactive `issue: IssueResponse` passed down as props to children
- Child components emit events for changes (`update:state`, `update:priority`, etc.)
- `IssueDetail.vue` handles API calls and optimistic updates
- Save button in header triggers `saveIssue()` which batches pending changes

### Relations Tab Data Loading

- `IssueTabRelations` loads:
  - `issue.parent` from parent issue data (already in response)
  - `issue.sub_issues` from sub-issues data (already in response)
  - Relations via `listIssueRelations(issueId)` API
  - Relation types via `listRelationTypes(workspaceId)` API
- Groups relations by `relation_type_id` for dynamic cards

## Error Handling

| Scenario | Handling |
|---|---|
| Save fails | Toast error, revert form to last saved state |
| Relation add fails | Toast error, remove from local list |
| Relation remove fails | Toast error, re-add to local list |
| Tab content load fails | Inline error state per tab with retry button |
| Custom field validation | Inline field error message, field border red |

## Component Interfaces

### IssueDetailHeader
```
Props: issue: IssueResponse, saving: boolean
Events: @save, @back
```

### IssuePropertySidebar
```
Props: issue: IssueResponse, states: State[], members: User[], 
       cycles: Cycle[], modules: Module[]
Events: @update:state, @update:priority, @update:assignee, 
        @update:cycle, @update:module, @update:startDate, 
        @update:targetDate, @update:labels
```

### IssueTabDetails
```
Props: issueId: number, issue: IssueResponse, 
       workspaceId: number, projectId: number
Events: — (uses CustomFieldManager and CommentList internally)
```

### IssueTabRelations
```
Props: issueId: number, projectId: number, workspaceId: number
Events: @navigate(issueId: number)
```

### RelationParentCard
```
Props: parent: IssueResponse | null, issueTypes: IssueType[]
Events: @change, @remove, @navigate(issueId)
```

### RelationSubIssuesCard
```
Props: subIssues: IssueResponse[]
Events: @add, @toggle(issueId), @navigate(issueId)
```

### RelationTypeCard
```
Props: typeName: string, typeId: number, items: RelationItem[], 
       color: string, issueTypes: IssueType[]
Events: @add, @remove(relationId), @navigate(issueId)
```

## Backend Changes

No backend changes required. All data is already available through existing APIs.

## Testing

- **Unit (Vitest + Vue Test Utils):** Each component tested in isolation
  - Props rendering (correct fields, wrong data, missing optional fields)
  - Events emission (correct payload on user interaction)
  - Edge cases (null parent, empty sub-issues, 0 relations, very long titles)
  - Responsive column hiding (simulate narrow viewport)
- **Integration:** IssueDetail.vue composition test — all child components render with mock data
- **E2E:** Existing tests in `tests/` updated for new selectors. Full flow: navigate to issue → switch tabs → add/remove relations → save changes

## Files

### Rewrite
- `frontend/src/views/IssueDetail.vue` — Full rewrite as orchestrator component

### Create
- `frontend/src/components/IssueDetailHeader.vue`
- `frontend/src/components/IssuePropertySidebar.vue`
- `frontend/src/components/IssueTabDetails.vue`
- `frontend/src/components/IssueTabRelations.vue`
- `frontend/src/components/IssueTabAttachments.vue`
- `frontend/src/components/IssueTabTimeTracking.vue`
- `frontend/src/components/IssueTabActivity.vue`
- `frontend/src/components/RelationParentCard.vue`
- `frontend/src/components/RelationSubIssuesCard.vue`
- `frontend/src/components/RelationTypeCard.vue`

### Reuse (no changes required)
- `frontend/src/components/RichTextEditor.vue`
- `frontend/src/components/CustomFieldManager.vue`
- `frontend/src/components/CommentList.vue`
- `frontend/src/components/AttachmentManager.vue`
- `frontend/src/components/TimeTrackPanel.vue`
- `frontend/src/components/LabelSelector.vue`
- `frontend/src/components/AgentSelector.vue`
- `frontend/src/components/RecurrenceConfig.vue`

### Delete (after migration)
- `frontend/src/components/SubIssuePanel.vue`

### Test Files
- `frontend/src/components/__tests__/IssueDetailHeader.test.ts`
- `frontend/src/components/__tests__/IssuePropertySidebar.test.ts`
- `frontend/src/components/__tests__/IssueTabDetails.test.ts`
- `frontend/src/components/__tests__/IssueTabRelations.test.ts`
- `frontend/src/components/__tests__/IssueTabAttachments.test.ts`
- `frontend/src/components/__tests__/IssueTabTimeTracking.test.ts`
- `frontend/src/components/__tests__/IssueTabActivity.test.ts`
- `frontend/src/components/__tests__/RelationParentCard.test.ts`
- `frontend/src/components/__tests__/RelationSubIssuesCard.test.ts`
- `frontend/src/components/__tests__/RelationTypeCard.test.ts`

## Migration

- Delete `SubIssuePanel.vue` after extracting logic into `RelationSubIssuesCard.vue`
- `IssueDetailPanel.vue` (slide-in drawer) is NOT in scope — keep as-is
- Router paths unchanged
- Existing API functions (`issueApi`, `relationApi`) reused as-is
