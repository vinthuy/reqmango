import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@/composables/useI18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }))

import RelationParentCard from '@/components/RelationParentCard.vue'

const parentIssue = {
  id: 55,
  sequence_id: 55,
  name: 'OAuth2 provider setup',
  state_name: 'Todo',
  state_group: 'todo',
  priority: 'high',
  assignees: [{ id: 1, display_name: 'Alice', username: 'alice' }],
  target_date: '2026-08-15',
  issue_type: { id: 2, name: 'Feature', color: '#6366f1' },
}

const issueTypes = [
  { id: 1, name: 'Task', color: '#6366f1' },
  { id: 2, name: 'Feature', color: '#10b981' },
]

function mountCard(overrides: Record<string, any> = {}) {
  const defaults = {
    parent: parentIssue,
    issueTypes,
  }
  return mount(RelationParentCard, {
    props: { ...defaults, ...overrides },
  })
}

describe('RelationParentCard', () => {
  it('renders parent issue fields when parent is provided', () => {
    const wrapper = mountCard()
    const text = wrapper.text()
    expect(text).toContain('PARENT')
    expect(text).toContain('Feature')
    expect(text).toContain('55')
    expect(text).toContain('OAuth2 provider setup')
    expect(text).toContain('Todo')
  })

  it('renders empty state when parent is null', () => {
    const wrapper = mountCard({ parent: null })
    expect(wrapper.text()).toContain('issue.setParent')
    expect(wrapper.find('[data-test="set-parent"]').exists()).toBe(true)
  })

  it('emits "change" when Change / Set Parent button is clicked', async () => {
    // Empty state — Set Parent button
    const emptyWrapper = mountCard({ parent: null })
    await emptyWrapper.find('[data-test="set-parent"]').trigger('click')
    expect(emptyWrapper.emitted('change')).toBeTruthy()

    // With parent — Change button
    const wrapper = mountCard()
    await wrapper.find('[data-test="change-parent"]').trigger('click')
    expect(wrapper.emitted('change')).toBeTruthy()
  })

  it('emits "remove" when Remove button is clicked (parent exists)', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-test="remove-parent"]').trigger('click')
    expect(wrapper.emitted('remove')).toBeTruthy()
  })

  it('emits "navigate" with parent ID when title is clicked', async () => {
    const wrapper = mountCard()
    const titleEl = wrapper.find('.relation-parent-title')
    expect(titleEl.exists()).toBe(true)
    await titleEl.trigger('click')
    expect(wrapper.emitted('navigate')).toBeTruthy()
    expect(wrapper.emitted('navigate')![0]).toEqual([55])
  })

  it('shows assignee name when assigned', () => {
    const wrapper = mountCard()
    expect(wrapper.text()).toContain('Alice')
  })

  it('does not crash for unassigned parent (assignees: [])', () => {
    const wrapper = mountCard({
      parent: { ...parentIssue, assignees: [] },
    })
    expect(wrapper.exists()).toBe(true)
    // Should show "—" for unassigned
    expect(wrapper.text()).toContain('—')
  })

  it('renders "—" for invalid/unparseable target_date', () => {
    const wrapper = mountCard({
      parent: { ...parentIssue, target_date: 'not-a-date' },
    })
    expect(wrapper.text()).toContain('—')
  })

  it('does not crash when parent has no issue_type', () => {
    const wrapper = mountCard({
      parent: { ...parentIssue, issue_type: undefined },
    })
    expect(wrapper.exists()).toBe(true)
    // Should show "—" for unknown type
    expect(wrapper.text()).toContain('—')
  })
})
