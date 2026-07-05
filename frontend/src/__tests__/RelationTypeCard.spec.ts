import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))

import RelationTypeCard from '@/components/RelationTypeCard.vue'

const mockItem = {
  id: 101,
  related_issue: {
    id: 55,
    sequence_id: 55,
    name: 'OAuth2 provider setup',
    state_name: 'Todo',
    state_group: 'todo',
    priority: 'high',
    assignees: [{ id: 2, display_name: 'Bob' }],
    start_date: null,
    target_date: '2026-07-15T00:00:00Z',
    issue_type: { id: 2, name: 'Feature', color: '#10b981' },
  },
  related_issue_id: 55,
}

function mountCard(overrides: Record<string, any> = {}) {
  const defaults = {
    typeName: 'blocks',
    items: [mockItem],
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
    expect(wrapper.text()).toContain('BLOCKS')
  })

  it('renders column headers in the table', () => {
    const wrapper = mountCard()
    const text = wrapper.text()
    expect(text).toContain('issue.type')
    expect(text).toContain('issue.title')
    expect(text).toContain('issue.status')
    expect(text).toContain('issue.priority')
    expect(text).toContain('issue.assignee')
    expect(text).toContain('issue.targetDate')
  })

  it('renders the item count badge', () => {
    const wrapper = mountCard({
      items: [
        { ...mockItem, id: 1 },
        { ...mockItem, id: 2, related_issue: { ...mockItem.related_issue, id: 20, sequence_id: 20 } },
      ],
    })
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
    const wrapper = mountCard({
      items: [{
        ...mockItem,
        related_issue: { ...mockItem.related_issue, assignees: [] },
      }],
    })
    expect(wrapper.text()).toContain('—')
  })

  it('shows dropdown menu when + Add button is clicked', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-test="add-relation"]').trigger('click')
    expect(wrapper.find('[data-test="add-existing-relation"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="quick-create-relation"]').exists()).toBe(true)
  })

  it('emits "add-existing" when select existing is clicked', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-test="add-relation"]').trigger('click')
    await wrapper.find('[data-test="add-existing-relation"]').trigger('click')
    expect(wrapper.emitted('add-existing')).toBeTruthy()
  })

  it('shows inline input when quick create is clicked', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-test="add-relation"]').trigger('click')
    await wrapper.find('[data-test="quick-create-relation"]').trigger('click')
    expect(wrapper.find('input[type="text"]').exists()).toBe(true)
  })

  it('emits "quick-create" with name when Enter is pressed in quick input', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-test="add-relation"]').trigger('click')
    await wrapper.find('[data-test="quick-create-relation"]').trigger('click')
    const input = wrapper.find('input[type="text"]')
    await input.setValue('New related issue')
    await input.trigger('keydown.enter')
    expect(wrapper.emitted('quick-create')).toBeTruthy()
    expect(wrapper.emitted('quick-create')![0]).toEqual(['New related issue'])
  })

  it('emits "navigate" when the title is clicked', async () => {
    const wrapper = mountCard()
    await wrapper.find('.relation-type-clickable').trigger('click')
    expect(wrapper.emitted('navigate')).toBeTruthy()
    expect(wrapper.emitted('navigate')![0]).toEqual([55])
  })

  it('emits "remove" with relation ID when remove button is clicked', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-test="remove-relation"]').trigger('click')
    expect(wrapper.emitted('remove')).toBeTruthy()
    expect(wrapper.emitted('remove')![0]).toEqual([101])
  })

  it('shows empty state when no items', () => {
    const wrapper = mountCard({ items: [] })
    expect(wrapper.text()).toContain('issueKanban.noRelations')
  })

  it('renders "—" for missing target_date', () => {
    const wrapper = mountCard({
      items: [{
        ...mockItem,
        related_issue: { ...mockItem.related_issue, target_date: null },
      }],
    })
    expect(wrapper.text()).toContain('—')
  })

  it('shows assignee display_name when present', () => {
    const wrapper = mountCard()
    expect(wrapper.text()).toContain('Bob')
  })
})
