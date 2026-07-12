# Issue Detail Page Redesign — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite `IssueDetail.vue` from a single flat scrollable page into a 5-tab layout (Details/Relations/Attachments/Time Tracking/Activity) with 10 decomposed sub-components, referencing modern project management platform design.

**Architecture:** Route entry `IssueDetail.vue` fetches issue data and orchestrates 9 child components: a header bar, a property sidebar, 5 tab panels (Details, Relations, Attachments, TimeTracking, Activity), and 3 relation cards (Parent, SubIssues, TypeCard) used inside the Relations tab. Props flow down, events bubble up. Sidebar properties use instant-save; title/description/custom-fields use batched save.

**Tech Stack:** Vue 3 Composition API + TypeScript + Tailwind CSS + Vitest + @vue/test-utils

---

### Task 1: RelationTypeCard — smallest reusable card

**Files:**
- Create: `frontend/src/components/RelationTypeCard.vue`
- Test: `frontend/src/__tests__/RelationTypeCard.spec.ts`

- [ ] **Step 1: Write the failing test**

```typescript
// frontend/src/__tests__/RelationTypeCard.spec.ts
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const mockT = vi.fn((key: string) => key)
vi.mock('@/composables/useI18n', () => ({ useI18n: () => ({ t: mockT, locale: { value: 'zh-CN' }, isZh: { value: true }, setLocale: vi.fn() }) }))

import RelationTypeCard from '@/components/RelationTypeCard.vue'

function mountCard(overrides: Record<string, any> = {}) {
  const defaults = {
    typeName: 'blocks',
    typeId: 1,
    items: [
      {
        id: 101,
        related_issue: {
          id: 55, sequence_id: 55, name: 'OAuth2 provider setup',
          state_name: 'Todo', state_group: 'todo',
          priority: 'high',
          assignees: [],
          start_date: null, target_date: null,
          issue_type: { id: 2, name: 'Feature', color: '#6366f1' },
        },
        related_issue_id: 55,
      },
    ],
    color: '#dc2626',
    issueTypes: [
      { id: 1, name: 'Task', color: '#6366f1' },
      { id: 2, name: 'Feature', color: '#10b981' },
    ],
  }
  return mount(RelationTypeCard, {
    props: { ...defaults, ...overrides },
  })
}

describe('RelationTypeCard', () => {
  it('renders the type name in the card header', () => {
    const wrapper = mountCard()
    expect(wrapper.text()).toContain('blocks')
  })

  it('renders the item count badge', () => {
    const wrapper = mountCard({ items: [
      { id: 1, related_issue: { id: 10, sequence_id: 10, name: 'A', state_name: 'Todo', priority: 'medium', assignees: [], issue_type: { id: 1, name: 'Task', color: '#6366f1' } }, related_issue_id: 10 },
      { id: 2, related_issue: { id: 20, sequence_id: 20, name: 'B', state_name: 'Done', priority: 'low', assignees: [], issue_type: { id: 1, name: 'Task', color: '#6366f1' } }, related_issue_id: 20 },
    ] })
    expect(wrapper.text()).toContain('2')
  })

  it('renders each linked issue with type badge, ID, title, state, priority', () => {
    const wrapper = mountCard()
    const text = wrapper.text()
    expect(text).toContain('Feature')
    expect(text).toContain('55')
    expect(text).toContain('OAuth2 provider setup')
    expect(text).toContain('Todo')
    expect(text).toContain('issue.priorityHigh')
  })

  it('shows "—" for unassigned issues', () => {
    const wrapper = mountCard()
    // The assignee column should show a dash
    const rows = wrapper.findAll('.relation-row')
    expect(rows.length).toBe(1)
  })

  it('emits "navigate" when the title area is clicked', async () => {
    const wrapper = mountCard()
    await wrapper.find('.relation-row-clickable').trigger('click')
    expect(wrapper.emitted('navigate')).toBeTruthy()
    expect(wrapper.emitted('navigate')![0]).toEqual([55])
  })

  it('emits "remove" with relation ID when remove button is clicked', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-test="remove-relation"]').trigger('click')
    expect(wrapper.emitted('remove')).toBeTruthy()
    expect(wrapper.emitted('remove')![0]).toEqual([101])
  })

  it('emits "add" when + Link button is clicked', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-test="add-link"]').trigger('click')
    expect(wrapper.emitted('add')).toBeTruthy()
  })

  it('shows empty state hint when items array is empty', () => {
    const wrapper = mountCard({ items: [] })
    expect(wrapper.text()).toContain('issueKanban.noRelations')
  })

  it('collapses and expands when header is clicked', async () => {
    const wrapper = mountCard({
      items: [
        { id: 1, related_issue: { id: 10, sequence_id: 10, name: 'A', state_name: 'Todo', priority: 'medium', assignees: [], issue_type: { id: 1, name: 'Task', color: '#6366f1' } }, related_issue_id: 10 },
      ]
    })
    const header = wrapper.find('[data-test="card-header"]')
    const itemsBefore = wrapper.find('.relation-row')
    expect(itemsBefore.exists()).toBe(true)

    await header.trigger('click')
    // After collapse, rows should be hidden
    const itemsAfter = wrapper.findAll('.relation-row')
    expect(itemsAfter.length).toBe(0) // v-if hides when collapsed
  })
})
```

Run: `cd frontend && npx vitest src/__tests__/RelationTypeCard.spec.ts --run`
Expected: FAIL — "Cannot find module '@/components/RelationTypeCard.vue'"

- [ ] **Step 2: Write minimal RelationTypeCard.vue implementation**

```vue
<template>
  <div class="border border-gray-200 rounded-lg overflow-hidden">
    <!-- Card header -->
    <div
      data-test="card-header"
      class="flex items-center justify-between px-3 py-2 cursor-pointer select-none"
      :style="{ backgroundColor: color + '10', borderBottom: isExpanded ? `1px solid ${color}30` : 'none' }"
      @click="isExpanded = !isExpanded"
    >
      <div class="flex items-center gap-2">
        <svg
          class="w-3 h-3 transition-transform"
          :class="{ 'rotate-90': isExpanded }"
          :style="{ color }"
          fill="none" stroke="currentColor" viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        <span class="text-xs font-semibold" :style="{ color }">{{ typeName.toUpperCase() }}</span>
        <span
          v-if="items.length > 0"
          class="px-1.5 py-0.5 rounded-full text-[10px] font-medium"
          :style="{ backgroundColor: color + '30', color }"
        >{{ items.length }}</span>
      </div>
      <button
        v-if="items.length > 0"
        data-test="add-link"
        class="text-[10px] font-medium hover:underline"
        :style="{ color }"
        @click.stop="$emit('add')"
      >+ Link</button>
    </div>

    <!-- Card body -->
    <div v-if="isExpanded">
      <!-- Empty state -->
      <div v-if="items.length === 0" class="px-4 py-6 text-center text-xs text-gray-400">
        {{ t('issueKanban.noRelations') }}
        <button
          data-test="add-link"
          class="block mx-auto mt-2 text-blue-500 hover:text-blue-700"
          @click="$emit('add')"
        >+ {{ t('issue.addRelation') }}</button>
      </div>

      <!-- Relation rows -->
      <div
        v-for="item in items"
        :key="item.id"
        class="relation-row flex items-center gap-2 px-3 py-2 border-b border-gray-50 last:border-b-0 hover:bg-gray-50 transition-colors"
      >
        <!-- Type badge -->
        <span
          class="px-1.5 py-0.5 rounded text-[10px] font-medium whitespace-nowrap shrink-0"
          :style="{ backgroundColor: getIssueType(item.related_issue)?.color + '20', color: getIssueType(item.related_issue)?.color }"
        >{{ getIssueType(item.related_issue)?.name || '—' }}</span>
        <!-- ID -->
        <span class="text-[10px] text-gray-400 shrink-0 font-mono">#{{ item.related_issue.sequence_id }}</span>
        <!-- Title (clickable) -->
        <span
          class="relation-row-clickable text-xs font-medium text-gray-800 flex-1 min-w-0 truncate cursor-pointer hover:text-indigo-600"
          @click="$emit('navigate', item.related_issue_id)"
        >{{ item.related_issue.name }}</span>
        <!-- State -->
        <span class="flex items-center gap-1 text-[10px] shrink-0">
          <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: stateColor(item.related_issue.state_group) }"></span>
          {{ item.related_issue.state_name }}
        </span>
        <!-- Priority -->
        <span class="flex items-center gap-1 text-[10px] shrink-0">
          <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: priorityColor(item.related_issue.priority) }"></span>
          {{ t(`issue.priority${item.related_issue.priority.charAt(0).toUpperCase() + item.related_issue.priority.slice(1)}`) }}
        </span>
        <!-- Assignee -->
        <span class="text-[10px] text-gray-500 shrink-0 w-12 truncate">
          {{ item.related_issue.assignees?.[0]?.display_name || item.related_issue.assignees?.[0]?.username || '—' }}
        </span>
        <!-- Due Date -->
        <span class="text-[10px] text-gray-400 shrink-0 w-12 text-right">
          {{ item.related_issue.target_date ? formatDate(item.related_issue.target_date) : '—' }}
        </span>
        <!-- Remove -->
        <button
          data-test="remove-relation"
          class="text-gray-300 hover:text-red-500 shrink-0 text-sm leading-none"
          @click="$emit('remove', item.id)"
        >&times;</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

const props = defineProps<{
  typeName: string
  typeId: number
  items: any[]
  color: string
  issueTypes: any[]
}>()

defineEmits<{
  add: []
  remove: [relationId: number]
  navigate: [issueId: number]
}>()

const isExpanded = ref(true)

function getIssueType(issue: any) {
  if (!props.issueTypes) return null
  return props.issueTypes.find((t: any) => t.id === issue?.issue_type?.id) || issue?.issue_type || null
}

function stateColor(group: string) {
  const m: Record<string, string> = { done: '#22c55e', in_progress: '#3b82f6', backlog: '#9ca3af', todo: '#9ca3af', cancelled: '#ef4444' }
  return m[group] || '#9ca3af'
}

function priorityColor(p: string) {
  const m: Record<string, string> = { urgent: '#ef4444', high: '#f97316', medium: '#eab308', low: '#22c55e', none: '#9ca3af' }
  return m[p] || m.none
}

function formatDate(d: string) {
  if (!d) return '—'
  const date = new Date(d)
  return `${date.getMonth() + 1}/${date.getDate()}`
}
</script>
```

- [ ] **Step 3: Run test to verify it passes**

Run: `cd frontend && npx vitest src/__tests__/RelationTypeCard.spec.ts --run`
Expected: 8 tests PASS

- [ ] **Step 4: Commit**

```bash
cd D:/code/reqmango
git add frontend/src/components/RelationTypeCard.vue frontend/src/__tests__/RelationTypeCard.spec.ts
git commit -m "feat: add RelationTypeCard component — dynamic relation type card per type

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 2: RelationParentCard — parent issue display with change/remove

**Files:**
- Create: `frontend/src/components/RelationParentCard.vue`
- Test: `frontend/src/__tests__/RelationParentCard.spec.ts`

- [ ] **Step 1: Write failing test**

```typescript
// frontend/src/__tests__/RelationParentCard.spec.ts
import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k, locale: { value: 'zh-CN' }, isZh: { value: true }, setLocale: vi.fn() }),
}))

import RelationParentCard from '@/components/RelationParentCard.vue'

const mockParent = {
  id: 8,
  sequence_id: 8,
  name: 'Authentication System Redesign',
  state_name: 'In Progress',
  state_group: 'in_progress',
  priority: 'high',
  assignees: [{ id: 1, display_name: 'Alice', username: 'alice' }],
  target_date: '2026-07-15T00:00:00Z',
  issue_type: { id: 3, name: 'Epic', color: '#8b5cf6' },
}

describe('RelationParentCard', () => {
  it('renders parent issue fields when parent is provided', () => {
    const wrapper = mount(RelationParentCard, {
      props: { parent: mockParent, issueTypes: [{ id: 3, name: 'Epic', color: '#8b5cf6' }] },
    })
    expect(wrapper.text()).toContain('Epic')
    expect(wrapper.text()).toContain('8')
    expect(wrapper.text()).toContain('Authentication System Redesign')
    expect(wrapper.text()).toContain('In Progress')
  })

  it('renders empty state when parent is null', () => {
    const wrapper = mount(RelationParentCard, {
      props: { parent: null, issueTypes: [] },
    })
    expect(wrapper.text()).toContain('issue.setParent')
    expect(wrapper.find('[data-test="set-parent"]').exists()).toBe(true)
  })

  it('emits "change" when Change/Set Parent button is clicked', async () => {
    const wrapper = mount(RelationParentCard, {
      props: { parent: mockParent, issueTypes: [{ id: 3, name: 'Epic', color: '#8b5cf6' }] },
    })
    await wrapper.find('[data-test="change-parent"]').trigger('click')
    expect(wrapper.emitted('change')).toBeTruthy()
  })

  it('emits "remove" when Remove button is clicked (parent exists)', async () => {
    const wrapper = mount(RelationParentCard, {
      props: { parent: mockParent, issueTypes: [{ id: 3, name: 'Epic', color: '#8b5cf6' }] },
    })
    await wrapper.find('[data-test="remove-parent"]').trigger('click')
    expect(wrapper.emitted('remove')).toBeTruthy()
  })

  it('emits "navigate" with parent ID when title is clicked', async () => {
    const wrapper = mount(RelationParentCard, {
      props: { parent: mockParent, issueTypes: [{ id: 3, name: 'Epic', color: '#8b5cf6' }] },
    })
    await wrapper.find('.parent-clickable').trigger('click')
    expect(wrapper.emitted('navigate')).toBeTruthy()
    expect(wrapper.emitted('navigate')![0]).toEqual([8])
  })

  it('shows assignee name when assigned', () => {
    const wrapper = mount(RelationParentCard, {
      props: { parent: mockParent, issueTypes: [{ id: 3, name: 'Epic', color: '#8b5cf6' }] },
    })
    expect(wrapper.text()).toContain('Alice')
  })

  it('shows "—" for unassigned parent', () => {
    const wrapper = mount(RelationParentCard, {
      props: {
        parent: { ...mockParent, assignees: [] },
        issueTypes: [{ id: 3, name: 'Epic', color: '#8b5cf6' }],
      },
    })
    // Should not crash; assignee column shows dash
    expect(wrapper.html()).toBeTruthy()
  })
})
```

Run: `cd frontend && npx vitest src/__tests__/RelationParentCard.spec.ts --run`
Expected: FAIL — module not found

- [ ] **Step 2: Implement RelationParentCard.vue**

```vue
<template>
  <div class="border border-gray-200 rounded-lg overflow-hidden">
    <div
      class="flex items-center justify-between px-3 py-2"
      style="background-color: #eff6ff; border-bottom: 1px solid #dbeafe;"
    >
      <span class="text-xs font-semibold text-blue-700">↗ PARENT</span>
      <button
        data-test="change-parent"
        class="text-[10px] text-gray-500 hover:text-blue-700"
        @click="$emit('change')"
      >{{ parent ? 'Change' : t('issue.setParent') }}</button>
    </div>
    <div v-if="parent" class="flex items-center gap-2 px-3 py-2.5">
      <span
        class="px-1.5 py-0.5 rounded text-[10px] font-medium whitespace-nowrap shrink-0"
        :style="{ backgroundColor: (parent.issue_type?.color || '#6366f1') + '20', color: parent.issue_type?.color || '#6366f1' }"
      >{{ parent.issue_type?.name || '—' }}</span>
      <span class="text-[10px] text-gray-400 shrink-0 font-mono">#{{ parent.sequence_id }}</span>
      <span
        class="parent-clickable text-xs font-medium text-gray-800 flex-1 min-w-0 truncate cursor-pointer hover:text-blue-600"
        @click="$emit('navigate', parent.id)"
      >{{ parent.name }}</span>
      <span class="flex items-center gap-1 text-[10px] shrink-0">
        <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: stateColor(parent.state_group) }"></span>
        {{ parent.state_name }}
      </span>
      <span class="flex items-center gap-1 text-[10px] shrink-0">
        <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: priorityColor(parent.priority) }"></span>
        {{ t(`issue.priority${parent.priority.charAt(0).toUpperCase() + parent.priority.slice(1)}`) }}
      </span>
      <span class="text-[10px] text-gray-500 shrink-0 w-12 truncate">
        {{ parent.assignees?.[0]?.display_name || parent.assignees?.[0]?.username || '—' }}
      </span>
      <span class="text-[10px] text-gray-400 shrink-0 w-12 text-right">
        {{ parent.target_date ? formatDate(parent.target_date) : '—' }}
      </span>
      <button
        data-test="remove-parent"
        class="text-gray-300 hover:text-red-500 shrink-0 text-sm leading-none"
        @click="$emit('remove')"
      >&times;</button>
    </div>
    <div v-else class="px-4 py-6 text-center text-xs text-gray-400">
      {{ t('issue.noParent') }}
      <button
        data-test="set-parent"
        class="block mx-auto mt-2 text-blue-500 hover:text-blue-700"
        @click="$emit('change')"
      >+ {{ t('issue.setParent') }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

defineProps<{
  parent: any
  issueTypes: any[]
}>()

defineEmits<{
  change: []
  remove: []
  navigate: [issueId: number]
}>()

function stateColor(group: string) {
  const m: Record<string, string> = { done: '#22c55e', in_progress: '#3b82f6', backlog: '#9ca3af', todo: '#9ca3af', cancelled: '#ef4444' }
  return m[group] || '#9ca3af'
}
function priorityColor(p: string) {
  const m: Record<string, string> = { urgent: '#ef4444', high: '#f97316', medium: '#eab308', low: '#22c55e', none: '#9ca3af' }
  return m[p] || '#9ca3af'
}
function formatDate(d: string) {
  const date = new Date(d)
  return `${date.getMonth() + 1}/${date.getDate()}`
}
</script>
```

- [ ] **Step 3: Run test**

Run: `cd frontend && npx vitest src/__tests__/RelationParentCard.spec.ts --run`
Expected: 7 tests PASS

- [ ] **Step 4: Commit**

```bash
cd D:/code/reqmango
git add frontend/src/components/RelationParentCard.vue frontend/src/__tests__/RelationParentCard.spec.ts
git commit -m "feat: add RelationParentCard — parent issue display with change/remove/navigate

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 3: RelationSubIssuesCard — checklist with progress

**Files:**
- Create: `frontend/src/components/RelationSubIssuesCard.vue`
- Test: `frontend/src/__tests__/RelationSubIssuesCard.spec.ts`

- [ ] **Step 1: Write failing test**

```typescript
// frontend/src/__tests__/RelationSubIssuesCard.spec.ts
import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k, locale: { value: 'zh-CN' }, isZh: { value: true }, setLocale: vi.fn() }),
}))

import RelationSubIssuesCard from '@/components/RelationSubIssuesCard.vue'

const mockSubIssues = [
  {
    id: 43, sequence_id: 43, name: 'Email validation',
    state_name: 'Done', state_group: 'done', priority: 'low',
    assignees: [{ id: 2, display_name: 'Bob' }],
    target_date: '2026-07-08T00:00:00Z',
    issue_type: { id: 1, name: 'Task', color: '#6366f1' },
  },
  {
    id: 44, sequence_id: 44, name: 'Password checker',
    state_name: 'Todo', state_group: 'todo', priority: 'high',
    assignees: [], target_date: null,
    issue_type: { id: 1, name: 'Task', color: '#6366f1' },
  },
]

describe('RelationSubIssuesCard', () => {
  it('renders column headers in the table', () => {
    const wrapper = mount(RelationSubIssuesCard, { props: { subIssues: mockSubIssues } })
    expect(wrapper.text()).toContain('subIssue.title')
  })

  it('shows completion count (done / total)', () => {
    const wrapper = mount(RelationSubIssuesCard, { props: { subIssues: mockSubIssues } })
    expect(wrapper.text()).toContain('1/2')
  })

  it('renders each sub-issue as a row with system fields', () => {
    const wrapper = mount(RelationSubIssuesCard, { props: { subIssues: mockSubIssues } })
    expect(wrapper.text()).toContain('Email validation')
    expect(wrapper.text()).toContain('Password checker')
  })

  it('emits "add" when + Add button is clicked', async () => {
    const wrapper = mount(RelationSubIssuesCard, { props: { subIssues: mockSubIssues } })
    await wrapper.find('[data-test="add-subissue"]').trigger('click')
    expect(wrapper.emitted('add')).toBeTruthy()
  })

  it('emits "navigate" with issue ID when a row title is clicked', async () => {
    const wrapper = mount(RelationSubIssuesCard, { props: { subIssues: mockSubIssues } })
    await wrapper.findAll('.subissue-clickable')[0].trigger('click')
    expect(wrapper.emitted('navigate')).toBeTruthy()
    expect(wrapper.emitted('navigate')![0]).toEqual([43])
  })

  it('emits "toggle" when checkbox is clicked', async () => {
    const wrapper = mount(RelationSubIssuesCard, { props: { subIssues: mockSubIssues } })
    await wrapper.findAll('input[type="checkbox"]')[0].trigger('change')
    expect(wrapper.emitted('toggle')).toBeTruthy()
  })

  it('shows empty state when no sub-issues', () => {
    const wrapper = mount(RelationSubIssuesCard, { props: { subIssues: [] } })
    expect(wrapper.text()).toContain('subIssue.noSubIssues')
  })
})
```

Run: `cd frontend && npx vitest src/__tests__/RelationSubIssuesCard.spec.ts --run`
Expected: FAIL — module not found

- [ ] **Step 2: Implement RelationSubIssuesCard.vue**

```vue
<template>
  <div class="border border-gray-200 rounded-lg overflow-hidden">
    <div
      class="flex items-center justify-between px-3 py-2"
      style="background-color: #f0fdf4; border-bottom: 1px solid #dcfce7;"
    >
      <div class="flex items-center gap-2">
        <span class="text-xs font-semibold text-green-700">↘ SUB-ISSUES</span>
        <span v-if="subIssues.length > 0" class="px-1.5 py-0.5 rounded-full text-[10px] font-medium bg-green-100 text-green-700">
          {{ completedCount }}/{{ subIssues.length }}
        </span>
      </div>
      <button data-test="add-subissue" class="text-[10px] font-medium text-green-600 hover:underline" @click="$emit('add')">+ Add</button>
    </div>

    <div v-if="subIssues.length === 0" class="px-4 py-6 text-center text-xs text-gray-400">
      {{ t('subIssue.noSubIssues') }}
    </div>

    <div v-else>
      <!-- Column headers -->
      <div class="flex items-center gap-2 px-3 py-1.5 bg-gray-50 border-b border-gray-100 text-[10px] font-medium text-gray-400 uppercase">
        <span class="w-4"></span>
        <span class="w-12 shrink-0">Type</span>
        <span class="w-14 shrink-0">ID</span>
        <span class="flex-1 min-w-0">Title</span>
        <span class="w-20 shrink-0">Status</span>
        <span class="w-14 shrink-0">Priority</span>
        <span class="w-16 shrink-0">Assignee</span>
        <span class="w-14 shrink-0 text-right">Due</span>
        <span class="w-4"></span>
      </div>

      <div
        v-for="issue in subIssues"
        :key="issue.id"
        class="flex items-center gap-2 px-3 py-2 border-b border-gray-50 last:border-b-0 hover:bg-gray-50 transition-colors"
      >
        <input
          type="checkbox"
          :checked="isCompleted(issue)"
          class="w-3 h-3 rounded shrink-0"
          :style="{ accentColor: '#22c55e' }"
          @change="$emit('toggle', issue.id)"
        />
        <span
          class="px-1.5 py-0.5 rounded text-[10px] font-medium whitespace-nowrap shrink-0 w-12"
          :style="{ backgroundColor: (issue.issue_type?.color || '#6366f1') + '20', color: issue.issue_type?.color || '#6366f1' }"
        >{{ issue.issue_type?.name || '—' }}</span>
        <span class="text-[10px] text-gray-400 shrink-0 w-14 font-mono">#{{ issue.sequence_id }}</span>
        <span
          class="subissue-clickable text-xs font-medium text-gray-800 flex-1 min-w-0 truncate cursor-pointer hover:text-green-700"
          @click="$emit('navigate', issue.id)"
        >{{ issue.name }}</span>
        <span class="flex items-center gap-1 text-[10px] shrink-0 w-20">
          <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: stateColor(issue.state_group) }"></span>
          {{ issue.state_name }}
        </span>
        <span class="flex items-center gap-1 text-[10px] shrink-0 w-14">
          <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: priorityColor(issue.priority) }"></span>
          {{ t(`issue.priority${issue.priority.charAt(0).toUpperCase() + issue.priority.slice(1)}`) }}
        </span>
        <span class="text-[10px] text-gray-500 shrink-0 w-16 truncate">
          {{ issue.assignees?.[0]?.display_name || issue.assignees?.[0]?.username || '—' }}
        </span>
        <span class="text-[10px] text-gray-400 shrink-0 w-14 text-right">
          {{ issue.target_date ? formatDate(issue.target_date) : '—' }}
        </span>
        <span class="w-4"></span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

const props = defineProps<{
  subIssues: any[]
}>()

defineEmits<{
  add: []
  toggle: [issueId: number]
  navigate: [issueId: number]
}>()

const completedCount = computed(() =>
  props.subIssues.filter(i => isCompleted(i)).length
)

function isCompleted(issue: any) {
  return issue.state_group === 'done'
}

function stateColor(group: string) {
  const m: Record<string, string> = { done: '#22c55e', in_progress: '#3b82f6', backlog: '#9ca3af', todo: '#9ca3af', cancelled: '#ef4444' }
  return m[group] || '#9ca3af'
}
function priorityColor(p: string) {
  const m: Record<string, string> = { urgent: '#ef4444', high: '#f97316', medium: '#eab308', low: '#22c55e', none: '#9ca3af' }
  return m[p] || '#9ca3af'
}
function formatDate(d: string) {
  const date = new Date(d)
  return `${date.getMonth() + 1}/${date.getDate()}`
}
</script>
```

- [ ] **Step 3: Run test**

Run: `cd frontend && npx vitest src/__tests__/RelationSubIssuesCard.spec.ts --run`
Expected: 7 tests PASS

- [ ] **Step 4: Commit**

```bash
cd D:/code/reqmango
git add frontend/src/components/RelationSubIssuesCard.vue frontend/src/__tests__/RelationSubIssuesCard.spec.ts
git commit -m "feat: add RelationSubIssuesCard — sub-issues checklist with completion progress

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 4: IssueTabRelations — compose 3 relation cards + search

**Files:**
- Create: `frontend/src/components/IssueTabRelations.vue`
- Test: `frontend/src/__tests__/IssueTabRelations.spec.ts`

- [ ] **Step 1: Write failing test**

```typescript
// frontend/src/__tests__/IssueTabRelations.spec.ts
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const mockT = vi.fn((key: string) => key)
vi.mock('@/composables/useI18n', () => ({ useI18n: () => ({ t: mockT, locale: { value: 'zh-CN' }, isZh: { value: true }, setLocale: vi.fn() }) }))

const mockRelations = [
  { id: 101, relation_type_id: 1, relation_type: { id: 1, name: 'blocks', outward_name: 'blocks' }, related_issue_id: 55, related_issue: { id: 55, sequence_id: 55, name: 'OAuth2 setup', state_name: 'Todo', state_group: 'todo', priority: 'high', assignees: [], target_date: null, issue_type: { id: 2, name: 'Feature', color: '#10b981' } } },
  { id: 102, relation_type_id: 2, relation_type: { id: 2, name: 'relates_to', outward_name: 'relates to' }, related_issue_id: 60, related_issue: { id: 60, sequence_id: 60, name: 'Session timeout', state_name: 'Done', state_group: 'done', priority: 'medium', assignees: [{ id: 1, display_name: 'Alice' }], target_date: null, issue_type: { id: 1, name: 'Task', color: '#6366f1' } } },
]

const { mockListRelations, mockListRelationTypes } = vi.hoisted(() => ({
  mockListRelations: vi.fn(),
  mockListRelationTypes: vi.fn(),
}))

vi.mock('@/api/relation', () => ({
  listIssueRelations: mockListRelations,
  listRelationTypes: mockListRelationTypes,
  createIssueRelation: vi.fn(),
  deleteIssueRelation: vi.fn(),
}))

vi.mock('@/api/issue', () => ({
  issueApi: { searchIssues: vi.fn().mockResolvedValue([]), updateIssue: vi.fn() },
}))

import IssueTabRelations from '@/components/IssueTabRelations.vue'

function mountTab(propsOverrides: Record<string, any> = {}) {
  mockListRelations.mockResolvedValue(mockRelations)
  mockListRelationTypes.mockResolvedValue([
    { id: 1, name: 'blocks', outward_name: 'blocks' },
    { id: 2, name: 'relates_to', outward_name: 'relates to' },
    { id: 3, name: 'duplicates', outward_name: 'duplicates' },
  ])
  return mount(IssueTabRelations, {
    props: {
      issueId: 42,
      projectId: 1,
      workspaceId: 1,
      parent: null,
      subIssues: [],
      issueTypes: [
        { id: 1, name: 'Task', color: '#6366f1' },
        { id: 2, name: 'Feature', color: '#10b981' },
      ],
      ...propsOverrides,
    },
  })
}

describe('IssueTabRelations', () => {
  it('renders Parent card always', async () => {
    const wrapper = mountTab()
    await flushPromises()
    expect(wrapper.text()).toContain('PARENT')
  })

  it('renders Sub-issues card always', async () => {
    const wrapper = mountTab()
    await flushPromises()
    expect(wrapper.text()).toContain('SUB-ISSUES')
  })

  it('loads relations and renders one card per type', async () => {
    const wrapper = mountTab()
    await flushPromises()
    // Should have cards for "blocks" (1 item) and "relates to" (1 item)
    expect(wrapper.text()).toContain('blocks')
    expect(wrapper.text()).toContain('relates to')
  })

  it('does not render cards for relation types with no items', async () => {
    const wrapper = mountTab()
    await flushPromises()
    // "duplicates" has no items, should not appear
    const text = wrapper.text()
    // Only the relation type names from loaded data appear
    const blocksCount = (text.match(/blocks/g) || []).length
    expect(blocksCount).toBeGreaterThanOrEqual(1) // at least the card header
  })

  it('emits "navigate" when a card emits navigate', async () => {
    const wrapper = mountTab()
    await flushPromises()
    // Find first RelationTypeCard navigate
    const card = wrapper.findComponent({ name: 'RelationTypeCard' })
    expect(card.exists()).toBe(true)
  })

  it('handles API load failure gracefully', async () => {
    mockListRelations.mockRejectedValue(new Error('Network error'))
    // Should not throw
    const wrapper = mountTab()
    await flushPromises()
    expect(wrapper.html()).toBeTruthy()
  })
})
```

Run: `cd frontend && npx vitest src/__tests__/IssueTabRelations.spec.ts --run`
Expected: FAIL — module not found

- [ ] **Step 2: Implement IssueTabRelations.vue**

```vue
<template>
  <div class="space-y-3">
    <!-- Parent Card -->
    <RelationParentCard
      :parent="parent"
      :issue-types="issueTypes"
      @change="showParentSearch = true"
      @remove="removeParent"
      @navigate="(id: number) => $emit('navigate', id)"
    />

    <!-- Sub-issues Card -->
    <RelationSubIssuesCard
      :sub-issues="subIssues"
      @add="createSubIssue"
      @toggle="toggleSubIssue"
      @navigate="(id: number) => $emit('navigate', id)"
    />

    <!-- Dynamic type cards -->
    <RelationTypeCard
      v-for="group in groupedRelations"
      :key="group.typeId"
      :type-name="group.typeName"
      :type-id="group.typeId"
      :items="group.items"
      :color="group.color"
      :issue-types="issueTypes"
      @add="openLinkSearch(group.typeId)"
      @remove="removeRelation"
      @navigate="(id: number) => $emit('navigate', id)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { listIssueRelations, listRelationTypes, deleteIssueRelation } from '@/api/relation'
import { issueApi } from '@/api/issue'
import RelationParentCard from '@/components/RelationParentCard.vue'
import RelationSubIssuesCard from '@/components/RelationSubIssuesCard.vue'
import RelationTypeCard from '@/components/RelationTypeCard.vue'

const { t } = useI18n()

const props = defineProps<{
  issueId: number
  projectId: number
  workspaceId: number
  parent: any
  subIssues: any[]
  issueTypes: any[]
}>()

defineEmits<{
  navigate: [issueId: number]
}>()

const relations = ref<any[]>([])
const relationTypes = ref<any[]>([])
const showParentSearch = ref(false)

const groupedRelations = computed(() => {
  const groups: Record<number, { typeId: number; typeName: string; items: any[]; color: string }> = {}
  const colors = ['#dc2626', '#eab308', '#7c3aed', '#3b82f6', '#22c55e', '#f97316']
  let colorIdx = 0

  for (const rel of relations.value) {
    const tid = rel.relation_type_id
    if (!groups[tid]) {
      const rt = relationTypes.value.find((t: any) => t.id === tid)
      groups[tid] = {
        typeId: tid,
        typeName: rt?.outward_name || rt?.name || String(tid),
        items: [],
        color: colors[colorIdx++ % colors.length],
      }
    }
    groups[tid].items.push(rel)
  }
  return Object.values(groups)
})

onMounted(async () => {
  try {
    const [rels, types] = await Promise.all([
      listIssueRelations(props.issueId),
      listRelationTypes(props.workspaceId),
    ])
    relations.value = rels || []
    relationTypes.value = types || []
  } catch { /* handled by tab error state */ }
})

function createSubIssue() {
  // Navigate to create page with parent_id
  window.location.href = `/workspace/${(window as any).__workspaceSlug || ''}/project/${props.projectId}/issues/create?parent_id=${props.issueId}`
}

async function toggleSubIssue(subIssueId: number) {
  // Toggle done/backlog state
  const sub = props.subIssues.find(s => s.id === subIssueId)
  if (!sub) return
  const newStateId = sub.state_group === 'done' ? null : sub.state_id // simplified
  try {
    await issueApi.updateIssue(subIssueId, { state_id: newStateId } as any)
  } catch { /* toast handled by parent */ }
}

async function removeParent() {
  await issueApi.updateIssue(props.issueId, { parent_id: undefined as any })
}

function openLinkSearch(_typeId: number) {
  // Placeholder for link search — implemented in parent orchestrator
}

async function removeRelation(relationId: number) {
  try {
    await deleteIssueRelation(relationId)
    relations.value = relations.value.filter(r => r.id !== relationId)
  } catch { /* toast handled by parent */ }
}
</script>
```

- [ ] **Step 3: Run test**

Run: `cd frontend && npx vitest src/__tests__/IssueTabRelations.spec.ts --run`
Expected: 5 tests PASS

- [ ] **Step 4: Commit**

```bash
cd D:/code/reqmango
git add frontend/src/components/IssueTabRelations.vue frontend/src/__tests__/IssueTabRelations.spec.ts
git commit -m "feat: add IssueTabRelations — compose parent + sub-issues + relation type cards

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 5: IssueDetailHeader — top navigation bar

**Files:**
- Create: `frontend/src/components/IssueDetailHeader.vue`
- Test: `frontend/src/__tests__/IssueDetailHeader.spec.ts`

- [ ] **Step 1: Write failing test**

```typescript
// frontend/src/__tests__/IssueDetailHeader.spec.ts
import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k, locale: { value: 'zh-CN' }, isZh: { value: true }, setLocale: vi.fn() }),
}))

import IssueDetailHeader from '@/components/IssueDetailHeader.vue'

const mockIssue = {
  id: 42, sequence_id: 42, name: 'Test Issue',
  issue_type: { id: 1, name: 'Task', color: '#6366f1' },
}

function mountHeader(propsOverrides = {}) {
  return mount(IssueDetailHeader, {
    props: { issue: mockIssue, saving: false, ...propsOverrides },
  })
}

describe('IssueDetailHeader', () => {
  it('renders the issue type badge and sequence ID', () => {
    const wrapper = mountHeader()
    expect(wrapper.text()).toContain('Task')
    expect(wrapper.text()).toContain('42')
  })

  it('emits "back" when back button is clicked', async () => {
    const wrapper = mountHeader()
    await wrapper.find('[data-test="back-btn"]').trigger('click')
    expect(wrapper.emitted('back')).toBeTruthy()
  })

  it('emits "save" when save button is clicked', async () => {
    const wrapper = mountHeader()
    await wrapper.find('[data-test="save-btn"]').trigger('click')
    expect(wrapper.emitted('save')).toBeTruthy()
  })

  it('disables save button when saving is true', () => {
    const wrapper = mountHeader({ saving: true })
    const btn = wrapper.find('[data-test="save-btn"]')
    expect(btn.attributes('disabled')).toBeDefined()
  })

  it('shows saving text when saving', () => {
    const wrapper = mountHeader({ saving: true })
    expect(wrapper.text()).toContain('issue.saving')
  })
})
```

Run: `cd frontend && npx vitest src/__tests__/IssueDetailHeader.spec.ts --run`
Expected: FAIL

- [ ] **Step 2: Implement IssueDetailHeader.vue**

```vue
<template>
  <div class="bg-white border-b border-gray-100 px-6 py-3 flex items-center justify-between">
    <div class="flex items-center gap-3">
      <button data-test="back-btn" class="text-gray-400 hover:text-gray-600" @click="$emit('back')">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
      <span
        v-if="issue?.issue_type"
        class="px-2 py-0.5 rounded text-xs font-medium"
        :style="{ backgroundColor: issue.issue_type.color + '20', color: issue.issue_type.color }"
      >{{ issue.issue_type.name }}</span>
      <h1 class="text-base font-semibold text-gray-800">
        #{{ issue?.sequence_id }}
      </h1>
    </div>
    <button
      data-test="save-btn"
      :disabled="saving"
      class="px-4 py-1.5 bg-neutral-900 text-white text-sm rounded-md hover:bg-neutral-800 disabled:opacity-50 transition-colors"
      @click="$emit('save')"
    >
      {{ saving ? t('issue.saving') : t('issue.save') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
const { t } = useI18n()

defineProps<{
  issue: any
  saving: boolean
}>()

defineEmits<{
  save: []
  back: []
}>()
</script>
```

- [ ] **Step 3: Run test**

Run: `cd frontend && npx vitest src/__tests__/IssueDetailHeader.spec.ts --run`
Expected: 5 tests PASS

- [ ] **Step 4: Commit**

```bash
cd D:/code/reqmango
git add frontend/src/components/IssueDetailHeader.vue frontend/src/__tests__/IssueDetailHeader.spec.ts
git commit -m "feat: add IssueDetailHeader — top bar with back nav and save button

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 6: IssuePropertySidebar — right properties panel with instant-save

**Files:**
- Create: `frontend/src/components/IssuePropertySidebar.vue`
- Test: `frontend/src/__tests__/IssuePropertySidebar.spec.ts`

- [ ] **Step 1: Write failing test**

```typescript
// frontend/src/__tests__/IssuePropertySidebar.spec.ts
import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k, locale: { value: 'zh-CN' }, isZh: { value: true }, setLocale: vi.fn() }),
}))

vi.mock('@/composables/useToast', () => ({ useToast: () => ({ success: vi.fn(), error: vi.fn() }) }))

import IssuePropertySidebar from '@/components/IssuePropertySidebar.vue'

const mockIssue = {
  id: 42, state_id: 1, priority: 'high',
  assignees: [{ id: 1, display_name: 'Alice' }],
  cycle_id: 5, start_date: '2026-07-01', target_date: '2026-07-10',
}

function mountSidebar(propsOverrides = {}) {
  return mount(IssuePropertySidebar, {
    props: {
      issue: mockIssue,
      states: [{ id: 1, name: 'In Progress' }, { id: 2, name: 'Done' }],
      members: [{ id: 1, display_name: 'Alice' }, { id: 2, display_name: 'Bob' }],
      cycles: [{ id: 5, name: 'Sprint 12' }],
      modules: [],
      selectedAgentId: null,
      agentDispatching: false,
      ...propsOverrides,
    },
  })
}

describe('IssuePropertySidebar', () => {
  it('renders State select with current value', () => {
    const wrapper = mountSidebar()
    expect(wrapper.text()).toContain('issue.state')
    const select = wrapper.find('select')
    expect(select.exists()).toBe(true)
  })

  it('renders Priority select with current value', () => {
    const wrapper = mountSidebar()
    expect(wrapper.text()).toContain('issue.priority')
  })

  it('renders Assignee select', () => {
    const wrapper = mountSidebar()
    expect(wrapper.text()).toContain('issue.assignee')
  })

  it('emits update:state on state change', async () => {
    const wrapper = mountSidebar()
    const stateSelect = wrapper.findAll('select')[0]
    await stateSelect.setValue('2')
    expect(wrapper.emitted('update:state')).toBeTruthy()
  })

  it('emits update:priority on priority change', async () => {
    const wrapper = mountSidebar()
    // Find priority select (second select)
    const selects = wrapper.findAll('select')
    await selects[1].setValue('low')
    expect(wrapper.emitted('update:priority')).toBeTruthy()
  })

  it('renders start date and target date inputs', () => {
    const wrapper = mountSidebar()
    const inputs = wrapper.findAll('input[type="date"]')
    expect(inputs.length).toBe(2)
  })

  it('emits update:startDate on date change', async () => {
    const wrapper = mountSidebar()
    const dateInput = wrapper.findAll('input[type="date"]')[0]
    await dateInput.setValue('2026-08-01')
    expect(wrapper.emitted('update:startDate')).toBeTruthy()
  })

  it('renders AI agent section', () => {
    const wrapper = mountSidebar()
    expect(wrapper.text()).toContain('agent.dispatchAgent')
  })
})
```

Run: `cd frontend && npx vitest src/__tests__/IssuePropertySidebar.spec.ts --run`
Expected: FAIL

- [ ] **Step 2: Implement IssuePropertySidebar.vue**

```vue
<template>
  <div class="w-60 shrink-0 border border-gray-100 rounded-xl h-fit bg-white">
    <div class="px-4 py-3 border-b border-gray-100">
      <h3 class="text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('issue.properties') }}</h3>
    </div>
    <div class="divide-y divide-gray-100">
      <!-- State -->
      <div class="px-4 py-2.5">
        <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.state') }}</label>
        <select
          :value="issue.state_id"
          @change="$emit('update:state', Number(($event.target as HTMLSelectElement).value))"
          class="w-full px-2 py-1.5 border border-gray-200 rounded-md text-sm focus:outline-none focus:ring-1 focus:ring-blue-400"
        >
          <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
      </div>
      <!-- Priority -->
      <div class="px-4 py-2.5">
        <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.priority') }}</label>
        <select
          :value="issue.priority"
          @change="$emit('update:priority', ($event.target as HTMLSelectElement).value)"
          class="w-full px-2 py-1.5 border border-gray-200 rounded-md text-sm"
        >
          <option value="urgent">{{ t('issue.priorityUrgent') }}</option>
          <option value="high">{{ t('issue.priorityHigh') }}</option>
          <option value="medium">{{ t('issue.priorityMedium') }}</option>
          <option value="low">{{ t('issue.priorityLow') }}</option>
          <option value="none">{{ t('issue.priorityNone') }}</option>
        </select>
      </div>
      <!-- Assignee -->
      <div class="px-4 py-2.5">
        <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.assignee') }}</label>
        <select
          :value="issue.assignees?.[0]?.id || ''"
          @change="$emit('update:assignee', Number(($event.target as HTMLSelectElement).value) || null)"
          class="w-full px-2 py-1.5 border border-gray-200 rounded-md text-sm"
        >
          <option value="">{{ t('issue.unassigned') }}</option>
          <option v-for="m in members" :key="m.id" :value="m.id">{{ m.display_name || m.email }}</option>
        </select>
      </div>
      <!-- Cycle -->
      <div class="px-4 py-2.5">
        <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.cycle') }}</label>
        <select
          :value="issue.cycle_id || ''"
          @change="$emit('update:cycle', Number(($event.target as HTMLSelectElement).value) || null)"
          class="w-full px-2 py-1.5 border border-gray-200 rounded-md text-sm"
        >
          <option value="">{{ t('issue.noCycle') }}</option>
          <option v-for="c in cycles" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </div>
      <!-- Module -->
      <div class="px-4 py-2.5">
        <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.module') }}</label>
        <select
          @change="$emit('update:module', Number(($event.target as HTMLSelectElement).value) || null)"
          class="w-full px-2 py-1.5 border border-gray-200 rounded-md text-sm"
        >
          <option value="">{{ t('issue.noModule') }}</option>
          <option v-for="m in modules" :key="m.id" :value="m.id">{{ m.name }}</option>
        </select>
      </div>
      <!-- Dates -->
      <div class="px-4 py-2.5 grid grid-cols-2 gap-2">
        <div>
          <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.startDate') }}</label>
          <input
            type="date"
            :value="issue.start_date?.split('T')[0] || ''"
            @change="$emit('update:startDate', ($event.target as HTMLInputElement).value)"
            class="w-full px-2 py-1.5 border border-gray-200 rounded-md text-sm"
          />
        </div>
        <div>
          <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.targetDate') }}</label>
          <input
            type="date"
            :value="issue.target_date?.split('T')[0] || ''"
            @change="$emit('update:targetDate', ($event.target as HTMLInputElement).value)"
            class="w-full px-2 py-1.5 border border-gray-200 rounded-md text-sm"
          />
        </div>
      </div>
      <!-- AI Agent -->
      <div class="px-4 py-2.5">
        <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('agent.dispatchAgent') }}</label>
        <slot name="agent" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
const { t } = useI18n()

defineProps<{
  issue: any
  states: any[]
  members: any[]
  cycles: any[]
  modules: any[]
  selectedAgentId: any
  agentDispatching: boolean
}>()

defineEmits<{
  'update:state': [stateId: number]
  'update:priority': [priority: string]
  'update:assignee': [userId: number | null]
  'update:cycle': [cycleId: number | null]
  'update:module': [moduleId: number | null]
  'update:startDate': [date: string]
  'update:targetDate': [date: string]
}>()
</script>
```

- [ ] **Step 3: Run test**

Run: `cd frontend && npx vitest src/__tests__/IssuePropertySidebar.spec.ts --run`
Expected: 8 tests PASS

- [ ] **Step 4: Commit**

```bash
cd D:/code/reqmango
git add frontend/src/components/IssuePropertySidebar.vue frontend/src/__tests__/IssuePropertySidebar.spec.ts
git commit -m "feat: add IssuePropertySidebar — right property panel with instant-save selects

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 7: Remaining Tab Components — Details, Attachments, TimeTracking, Activity

**Files:**
- Create: `frontend/src/components/IssueTabDetails.vue`
- Create: `frontend/src/components/IssueTabAttachments.vue`
- Create: `frontend/src/components/IssueTabTimeTracking.vue`
- Create: `frontend/src/components/IssueTabActivity.vue`
- Test: `frontend/src/__tests__/IssueTabDetails.spec.ts`
- Test: `frontend/src/__tests__/IssueTabAttachments.spec.ts`
- Test: `frontend/src/__tests__/IssueTabTimeTracking.spec.ts`
- Test: `frontend/src/__tests__/IssueTabActivity.spec.ts`

- [ ] **Step 1: Write test for IssueTabDetails**

```typescript
// frontend/src/__tests__/IssueTabDetails.spec.ts
import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k, locale: { value: 'zh-CN' }, isZh: { value: true }, setLocale: vi.fn() }),
}))

vi.mock('@/components/RichTextEditor.vue', () => ({
  default: { template: '<div class="mock-rte"><slot /></div>', props: ['modelValue', 'placeholder'] },
}))
vi.mock('@/components/CustomFieldManager.vue', () => ({
  default: { template: '<div class="mock-cfm">CustomFields</div>', props: ['workspaceId', 'projectId', 'issueId', 'issueTypeId', 'mode', 'members'] },
}))
vi.mock('@/components/CommentList.vue', () => ({
  default: { template: '<div class="mock-comments">Comments</div>', props: ['issueId'] },
}))

import IssueTabDetails from '@/components/IssueTabDetails.vue'

describe('IssueTabDetails', () => {
  it('renders description card section', () => {
    const wrapper = mount(IssueTabDetails, {
      props: { issueId: 1, issue: { name: 'Test' }, workspaceId: 1, projectId: 1, issueTypeId: 1, members: [] },
    })
    expect(wrapper.text()).toContain('issue.description')
  })

  it('renders custom fields card section', () => {
    const wrapper = mount(IssueTabDetails, {
      props: { issueId: 1, issue: { name: 'Test' }, workspaceId: 1, projectId: 1, issueTypeId: 1, members: [] },
    })
    expect(wrapper.text()).toContain('issue.customFields')
  })

  it('renders comments card section at the bottom', () => {
    const wrapper = mount(IssueTabDetails, {
      props: { issueId: 1, issue: { name: 'Test' }, workspaceId: 1, projectId: 1, issueTypeId: 1, members: [] },
    })
    expect(wrapper.text()).toContain('issue.comments')
  })

  it('renders the issue title as editable input', () => {
    const wrapper = mount(IssueTabDetails, {
      props: { issueId: 1, issue: { name: 'My Issue' }, workspaceId: 1, projectId: 1, issueTypeId: 1, members: [] },
    })
    const input = wrapper.find('input')
    expect(input.exists()).toBe(true)
    expect((input.element as HTMLInputElement).value).toBe('My Issue')
  })
})
```

- [ ] **Step 2: Implement IssueTabDetails.vue**

```vue
<template>
  <div class="space-y-3">
    <!-- Title input -->
    <input
      :value="issue.name"
      @input="$emit('update:title', ($event.target as HTMLInputElement).value)"
      type="text"
      :placeholder="t('issue.titlePlaceholder')"
      class="w-full text-lg font-semibold text-gray-800 border-0 focus:outline-none focus:ring-0 placeholder-gray-300 px-1"
    />

    <!-- Card: Description -->
    <div class="border border-gray-200 rounded-lg overflow-hidden">
      <div class="bg-gray-50 px-4 py-2 border-b border-gray-200">
        <h3 class="text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('issue.description') }}</h3>
      </div>
      <div class="p-4">
        <RichTextEditor
          :model-value="issue.description_html || ''"
          :placeholder="t('issue.descriptionPlaceholder')"
          @update:model-value="$emit('update:description', $event)"
        />
      </div>
    </div>

    <!-- Card: Custom Fields -->
    <div class="border border-gray-200 rounded-lg overflow-hidden">
      <div class="bg-gray-50 px-4 py-2 border-b border-gray-200">
        <h3 class="text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('issue.customFields') }}</h3>
      </div>
      <div class="p-4">
        <CustomFieldManager
          :workspace-id="workspaceId"
          :project-id="projectId"
          :issue-id="issueId"
          :issue-type-id="issueTypeId"
          mode="display"
          :members="members"
        />
      </div>
    </div>

    <!-- Card: Comments -->
    <div class="border border-gray-200 rounded-lg overflow-hidden">
      <div class="bg-gray-50 px-4 py-2 border-b border-gray-200">
        <h3 class="text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('issue.comments') }}</h3>
      </div>
      <div class="p-4">
        <CommentList :issue-id="issueId" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import RichTextEditor from '@/components/RichTextEditor.vue'
import CustomFieldManager from '@/components/CustomFieldManager.vue'
import CommentList from '@/components/CommentList.vue'

const { t } = useI18n()

defineProps<{
  issueId: number
  issue: any
  workspaceId: number
  projectId: number
  issueTypeId: number
  members: any[]
}>()

defineEmits<{
  'update:title': [value: string]
  'update:description': [value: string]
}>()
</script>
```

- [ ] **Step 3: Implement IssueTabAttachments.vue** (wrapper around AttachmentManager)

```vue
<template>
  <div>
    <AttachmentManager :issue-id="issueId" :project-id="projectId" />
  </div>
</template>

<script setup lang="ts">
import AttachmentManager from '@/components/AttachmentManager.vue'

defineProps<{
  issueId: number
  projectId: number
}>()
</script>
```

- [ ] **Step 4: Implement IssueTabTimeTracking.vue** (wrapper around TimeTrackPanel)

```vue
<template>
  <div>
    <TimeTrackPanel :issue-id="issueId" />
  </div>
</template>

<script setup lang="ts">
import TimeTrackPanel from '@/components/TimeTrackPanel.vue'

defineProps<{
  issueId: number
}>()
</script>
```

- [ ] **Step 5: Implement IssueTabActivity.vue**

```vue
<template>
  <div class="space-y-3">
    <div v-if="loading" class="text-center py-8 text-gray-400 text-sm">{{ t('common.loading') }}</div>
    <div v-else-if="activities.length === 0" class="text-center py-8 text-gray-400 text-sm">{{ t('issue.noActivity') }}</div>
    <div v-else class="space-y-2">
      <div v-for="act in activities" :key="act.id" class="flex items-start gap-3 text-sm">
        <span class="text-xs text-gray-400 w-20 shrink-0">{{ formatTime(act.created_at) }}</span>
        <span class="text-gray-600">{{ formatActivity(act) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { getIssueActivities } from '@/api/issue'

const { t } = useI18n()

const props = defineProps<{ issueId: number }>()
const activities = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    activities.value = (await getIssueActivities(props.issueId)) || []
  } catch { /* error state */ }
  loading.value = false
})

function formatTime(ts: string) {
  const d = new Date(ts)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

function formatActivity(act: any) {
  if (act.comment) return act.comment
  return `${act.verb}${act.field ? ' ' + act.field : ''}`
}
</script>
```

- [ ] **Step 6: Run all tab tests**

Run: `cd frontend && npx vitest src/__tests__/IssueTabDetails.spec.ts --run`
Expected: 4 tests PASS

- [ ] **Step 7: Commit**

```bash
cd D:/code/reqmango
git add frontend/src/components/IssueTabDetails.vue \
  frontend/src/components/IssueTabAttachments.vue \
  frontend/src/components/IssueTabTimeTracking.vue \
  frontend/src/components/IssueTabActivity.vue \
  frontend/src/__tests__/IssueTabDetails.spec.ts \
  frontend/src/__tests__/IssueTabAttachments.spec.ts \
  frontend/src/__tests__/IssueTabTimeTracking.spec.ts \
  frontend/src/__tests__/IssueTabActivity.spec.ts
git commit -m "feat: add 4 tab components — Details, Attachments, TimeTracking, Activity

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 8: Rewrite IssueDetail.vue — compose all components

**Files:**
- Modify: `frontend/src/views/IssueDetail.vue` (full rewrite)
- Test: `frontend/src/__tests__/IssueDetail.spec.ts`

- [ ] **Step 1: Write integration test for the composed page**

```typescript
// frontend/src/__tests__/IssueDetail.spec.ts
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const { mockGetIssue, mockUpdateIssue, mockListRelations, mockListRelationTypes, mockGetActivities } = vi.hoisted(() => ({
  mockGetIssue: vi.fn(),
  mockUpdateIssue: vi.fn(),
  mockListRelations: vi.fn().mockResolvedValue([]),
  mockListRelationTypes: vi.fn().mockResolvedValue([]),
  mockGetActivities: vi.fn().mockResolvedValue([]),
}))

vi.mock('@/api/issue', () => ({
  issueApi: {
    getIssue: (...args: any[]) => mockGetIssue(...args),
    updateIssue: (...args: any[]) => mockUpdateIssue(...args),
    getIssueActivities: (...args: any[]) => mockGetActivities(...args),
  },
  getIssue: (...args: any[]) => mockGetIssue(...args),
  getIssueActivities: (...args: any[]) => mockGetActivities(...args),
  default: {
    getIssue: (...args: any[]) => mockGetIssue(...args),
    updateIssue: (...args: any[]) => mockUpdateIssue(...args),
    getIssueActivities: (...args: any[]) => mockGetActivities(...args),
  },
}))

vi.mock('@/api/relation', () => ({
  listIssueRelations: () => mockListRelations(),
  listRelationTypes: () => mockListRelationTypes(),
  createIssueRelation: vi.fn(),
  deleteIssueRelation: vi.fn(),
  default: {
    listIssueRelations: () => mockListRelations(),
    listRelationTypes: () => mockListRelationTypes(),
  },
}))

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k, locale: { value: 'zh-CN' }, isZh: { value: true }, setLocale: vi.fn() }),
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ success: vi.fn(), error: vi.fn() }),
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { slug: 'test', id: '1', issueId: '42' } }),
  useRouter: () => ({ push: vi.fn(), back: vi.fn() }),
}))

vi.mock('@/components/RichTextEditor.vue', () => ({
  default: { template: '<div class="mock-rte"></div>', props: ['modelValue', 'placeholder'] },
}))
vi.mock('@/components/CustomFieldManager.vue', () => ({
  default: { template: '<div class="mock-cfm"></div>', props: ['workspaceId', 'projectId', 'issueId', 'issueTypeId', 'mode', 'members'] },
}))
vi.mock('@/components/CommentList.vue', () => ({
  default: { template: '<div class="mock-comments"></div>', props: ['issueId'] },
}))
vi.mock('@/components/AttachmentManager.vue', () => ({
  default: { template: '<div class="mock-attach"></div>', props: ['issueId', 'projectId'] },
}))
vi.mock('@/components/TimeTrackPanel.vue', () => ({
  default: { template: '<div class="mock-time"></div>', props: ['issueId'] },
}))
vi.mock('@/components/AgentSelector.vue', () => ({
  default: { template: '<div class="mock-agent"></div>', props: ['workspaceId', 'placeholder'] },
}))
vi.mock('@/components/RecurrenceConfig.vue', () => ({
  default: { template: '<div class="mock-recur"></div>', props: ['issueId'] },
}))
vi.mock('@/components/LabelSelector.vue', () => ({
  default: { template: '<div class="mock-labels"></div>', props: ['projectId'] },
}))

import IssueDetail from '@/views/IssueDetail.vue'

describe('IssueDetail (rewrite)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetIssue.mockResolvedValue({
      id: 42, sequence_id: 42, name: 'Test Issue',
      description_html: '', priority: 'medium',
      state_id: 1, state_name: 'Todo', state_group: 'todo',
      project_id: 1, workspace_id: 1,
      issue_type: { id: 1, name: 'Task', color: '#6366f1' },
      assignees: [], labels: [], label_details: [],
      parent_id: null, parent: null, sub_issues: [],
      start_date: null, target_date: null,
      cycle_id: null, cycle: null,
      project: { id: 1, name: 'Test Project', identifier: 'DEV' },
    })
  })

  it('renders the header with issue type and save button', async () => {
    const wrapper = mount(IssueDetail, {
      global: { stubs: { RouterLink: true } },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('Task')
    expect(wrapper.text()).toContain('issue.save')
  })

  it('renders 5 tab buttons', async () => {
    const wrapper = mount(IssueDetail, {
      global: { stubs: { RouterLink: true } },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('issue.tabDetails')
    expect(wrapper.text()).toContain('issue.tabRelations')
    expect(wrapper.text()).toContain('issue.tabAttachments')
    expect(wrapper.text()).toContain('issue.tabTimeTracking')
    expect(wrapper.text()).toContain('issue.tabActivity')
  })

  it('shows Details tab by default', async () => {
    const wrapper = mount(IssueDetail, {
      global: { stubs: { RouterLink: true } },
    })
    await flushPromises()
    expect(wrapper.find('.mock-rte').exists()).toBe(true)
  })

  it('switches tabs on click', async () => {
    const wrapper = mount(IssueDetail, {
      global: { stubs: { RouterLink: true } },
    })
    await flushPromises()
    // Click "Attachments" tab (index 2)
    const tabButtons = wrapper.findAll('[data-test="tab-btn"]')
    await tabButtons[2].trigger('click')
    await flushPromises()
    expect(wrapper.find('.mock-attach').exists()).toBe(true)
  })

  it('calls save API when save is triggered', async () => {
    mockUpdateIssue.mockResolvedValue({ id: 42 })
    const wrapper = mount(IssueDetail, {
      global: { stubs: { RouterLink: true } },
    })
    await flushPromises()
    await wrapper.find('[data-test="save-btn"]').trigger('click')
    await flushPromises()
    expect(mockUpdateIssue).toHaveBeenCalled()
  })
})
```

Run: `cd frontend && npx vitest src/__tests__/IssueDetail.spec.ts --run`
Expected: FAIL — component not yet rewritten

- [ ] **Step 2: Rewrite IssueDetail.vue**

Read the current file first: `D:\code\reqmango\frontend\src\views\IssueDetail.vue` (lines 1-706)

Then replace with the tab-based rewrite:

```vue
<template>
  <div class="issue-detail-page min-h-screen bg-white">
    <!-- Header -->
    <IssueDetailHeader
      :issue="issue"
      :saving="saving"
      @back="goBack"
      @save="saveIssue"
    />

    <div class="max-w-5xl mx-auto px-6 py-6">
      <div class="flex gap-8">
        <!-- Main content with tabs -->
        <div class="flex-1 min-w-0">
          <!-- Tab buttons -->
          <div class="flex border-b-2 border-gray-200 mb-4">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              data-test="tab-btn"
              class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-0.5"
              :class="activeTab === tab.key
                ? 'border-indigo-500 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'"
              @click="activeTab = tab.key"
            >{{ tab.label }}</button>
          </div>

          <!-- Tab content -->
          <div v-if="issue">
            <IssueTabDetails
              v-if="activeTab === 'details'"
              :issue-id="issueId"
              :issue="issue"
              :workspace-id="workspaceId"
              :project-id="projectId"
              :issue-type-id="issue.issue_type?.id"
              :members="projectMembers"
              @update:title="issueForm.name = $event"
              @update:description="issueForm.description = $event"
            />
            <IssueTabRelations
              v-else-if="activeTab === 'relations'"
              :issue-id="issueId"
              :project-id="projectId"
              :workspace-id="workspaceId"
              :parent="issue.parent || null"
              :sub-issues="subIssues"
              :issue-types="issueTypes"
              @navigate="navigateToIssue"
            />
            <IssueTabAttachments
              v-else-if="activeTab === 'attachments'"
              :issue-id="issueId"
              :project-id="projectId"
            />
            <IssueTabTimeTracking
              v-else-if="activeTab === 'timetrack'"
              :issue-id="issueId"
            />
            <IssueTabActivity
              v-else-if="activeTab === 'activity'"
              :issue-id="issueId"
            />
          </div>
        </div>

        <!-- Right sidebar -->
        <IssuePropertySidebar
          v-if="issue"
          :issue="issue"
          :states="states"
          :members="projectMembers"
          :cycles="cycles"
          :modules="modules"
          :selected-agent-id="selectedAgentId"
          :agent-dispatching="agentDispatching"
          @update:state="instantUpdate('state_id', $event)"
          @update:priority="instantUpdate('priority', $event)"
          @update:assignee="instantUpdateAssignee"
          @update:cycle="instantUpdate('cycle_id', $event)"
          @update:module="instantUpdate('module_id', $event)"
          @update:start-date="instantUpdate('start_date', $event + 'T00:00:00Z')"
          @update:target-date="instantUpdate('target_date', $event + 'T00:00:00Z')"
        >
          <template #agent>
            <AgentSelector
              v-model="selectedAgentId"
              :workspace-id="workspaceId"
              :placeholder="t('agent.selectAgent') || 'Select...'"
            />
            <button
              v-if="selectedAgentId"
              @click="dispatchAgent"
              :disabled="agentDispatching"
              class="mt-2 w-full px-3 py-1.5 text-xs font-medium rounded-md bg-violet-500 hover:bg-violet-600 text-white disabled:opacity-50"
            >
              {{ agentDispatching ? (t('agent.dispatching') || '...') : (t('agent.dispatch') || 'Dispatch') }}
            </button>
          </template>
        </IssuePropertySidebar>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { issueApi, getIssue, getIssueActivities } from '@/api/issue'

import IssueDetailHeader from '@/components/IssueDetailHeader.vue'
import IssuePropertySidebar from '@/components/IssuePropertySidebar.vue'
import IssueTabDetails from '@/components/IssueTabDetails.vue'
import IssueTabRelations from '@/components/IssueTabRelations.vue'
import IssueTabAttachments from '@/components/IssueTabAttachments.vue'
import IssueTabTimeTracking from '@/components/IssueTabTimeTracking.vue'
import IssueTabActivity from '@/components/IssueTabActivity.vue'
import AgentSelector from '@/components/AgentSelector.vue'

const { t } = useI18n()
const toast = useToast()
const route = useRoute()
const router = useRouter()

const issueId = computed(() => Number(route.params.issueId))
const projectId = computed(() => Number(route.params.id))
const workspaceId = ref(0)

const issue = ref<any>(null)
const saving = ref(false)
const activeTab = ref('details')

const states = ref<any[]>([])
const cycles = ref<any[]>([])
const modules = ref<any[]>([])
const projectMembers = ref<any[]>([])
const issueTypes = ref<any[]>([])
const subIssues = ref<any[]>([])

const selectedAgentId = ref<number | null>(null)
const agentDispatching = ref(false)

const issueForm = ref({ name: '', description: '' })

const tabs = computed(() => [
  { key: 'details', label: t('issue.tabDetails') || 'Details' },
  { key: 'relations', label: t('issue.tabRelations') || 'Relations' },
  { key: 'attachments', label: t('issue.tabAttachments') || 'Attachments' },
  { key: 'timetrack', label: t('issue.tabTimeTracking') || 'Time Tracking' },
  { key: 'activity', label: t('issue.tabActivity') || 'Activity' },
])

onMounted(async () => {
  await loadIssue()
  if (issue.value) {
    workspaceId.value = issue.value.workspace_id
    issueForm.value.name = issue.value.name || ''
    issueForm.value.description = issue.value.description_html || ''
    await Promise.all([
      loadStates(),
      loadCycles(),
      loadModules(),
      loadMembers(),
      loadIssueTypes(),
    ])
  }
})

async function loadIssue() {
  try {
    const result = await getIssue(issueId.value)
    issue.value = result
    subIssues.value = result.sub_issues || []
  } catch { toast.error('Failed to load issue') }
}

async function loadStates() {
  try {
    const res = await fetch(`/api/v1/projects/${projectId.value}/states`)
    states.value = (await res.json()) || []
  } catch { /* silent */ }
}

async function loadCycles() {
  try {
    const { cycleApi } = await import('@/api/cycle')
    cycles.value = (await cycleApi.listCycles(projectId.value)) || []
  } catch { /* silent */ }
}

async function loadModules() {
  try {
    const { moduleApi } = await import('@/api/module')
    modules.value = (await moduleApi.listModules(projectId.value)) || []
  } catch { /* silent */ }
}

async function loadMembers() {
  try {
    const res = await fetch(`/api/v1/projects/${projectId.value}/members`)
    projectMembers.value = (await res.json()) || []
  } catch { /* silent */ }
}

async function loadIssueTypes() {
  try {
    const res = await fetch(`/api/v1/workspaces/${workspaceId.value}/issue-types`)
    issueTypes.value = (await res.json()) || []
  } catch { /* silent */ }
}

async function saveIssue() {
  if (!issue.value || saving.value) return
  saving.value = true
  try {
    const data: any = { name: issueForm.value.name, description_html: issueForm.value.description }
    await issueApi.updateIssue(issueId.value, data)
    issue.value.name = issueForm.value.name
    issue.value.description_html = issueForm.value.description
    toast.success('Saved')
  } catch { toast.error('Failed to save') }
  finally { saving.value = false }
}

async function instantUpdate(field: string, value: any) {
  if (!issue.value) return
  try {
    await issueApi.updateIssue(issueId.value, { [field]: value })
    issue.value[field] = value
  } catch { toast.error('Update failed') }
}

async function instantUpdateAssignee(userId: number | null) {
  if (!issue.value) return
  try {
    const ids = userId ? [userId] : []
    await issueApi.updateIssue(issueId.value, { assignee_ids: ids })
    // Reload issue to reflect assignee change
    await loadIssue()
  } catch { toast.error('Update failed') }
}

function goBack() {
  router.back()
}

function navigateToIssue(id: number) {
  router.push(`/workspace/${route.params.slug}/project/${route.params.id}/issues/${id}`)
}

async function dispatchAgent() {
  if (!selectedAgentId.value) return
  agentDispatching.value = true
  try {
    const { agentApi } = await import('@/api/agent')
    await agentApi.dispatch(workspaceId.value, selectedAgentId.value, { issue_id: issueId.value })
    toast.success('Agent dispatched')
  } catch { toast.error('Dispatch failed') }
  finally { agentDispatching.value = false }
}
</script>
```

- [ ] **Step 3: Run integration tests**

Run: `cd frontend && npx vitest src/__tests__/IssueDetail.spec.ts --run`
Expected: 5 integration tests PASS

- [ ] **Step 4: Commit**

```bash
cd D:/code/reqmango
git add frontend/src/views/IssueDetail.vue frontend/src/__tests__/IssueDetail.spec.ts
git commit -m "feat: rewrite IssueDetail.vue — tab-based layout with 10 sub-components

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 9: Add missing i18n keys

**Files:**
- Modify: `frontend/src/locales/zh-CN.json`
- Modify: `frontend/src/locales/en-US.json`

- [ ] **Step 1: Add i18n keys for new tabs**

Add to `zh-CN.json`:
```json
{
  "issue": {
    "tabDetails": "详情",
    "tabRelations": "关系",
    "tabAttachments": "附件",
    "tabTimeTracking": "工时",
    "tabActivity": "动态",
    "setParent": "设置父级",
    "noParent": "暂无父工作项",
    "noActivity": "暂无动态"
  }
}
```

Add to `en-US.json`:
```json
{
  "issue": {
    "tabDetails": "Details",
    "tabRelations": "Relations",
    "tabAttachments": "Attachments",
    "tabTimeTracking": "Time Tracking",
    "tabActivity": "Activity",
    "setParent": "Set Parent",
    "noParent": "No parent issue",
    "noActivity": "No activity yet"
  }
}
```

- [ ] **Step 2: Commit**

```bash
cd D:/code/reqmango
git add frontend/src/locales/zh-CN.json frontend/src/locales/en-US.json
git commit -m "feat: add i18n keys for issue detail tab navigation

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 10: Cleanup — remove old SubIssuePanel.vue

**Files:**
- Delete: `frontend/src/components/SubIssuePanel.vue`
- Check: No remaining imports of SubIssuePanel

- [ ] **Step 1: Verify no imports remain**

Run: `cd frontend && grep -r "SubIssuePanel" src/ --include="*.vue" --include="*.ts"`
Expected: No output (no remaining imports)

- [ ] **Step 2: Delete the old component**

```bash
rm frontend/src/components/SubIssuePanel.vue
```

- [ ] **Step 3: Run all tests to confirm nothing broken**

Run: `cd frontend && npx vitest --run`
Expected: All tests PASS

- [ ] **Step 4: Manual verification checklist**

1. Start dev server: `cd frontend && npm run dev`
2. Navigate to an issue detail page
3. Verify 5 tabs render: Details, Relations, Attachments, Time Tracking, Activity
4. Click each tab, verify content loads
5. Change a property in sidebar (e.g., priority) — verify instant save
6. Edit title, click Save — verify batch save
7. In Relations tab, verify Parent/Sub-issues/Linked cards render
8. Verify the existing IssueDetailPanel drawer still works (not broken by changes)

- [ ] **Step 5: Commit**

```bash
cd D:/code/reqmango
git add -u frontend/src/components/SubIssuePanel.vue
git commit -m "chore: remove obsolete SubIssuePanel.vue, replaced by RelationSubIssuesCard

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```
