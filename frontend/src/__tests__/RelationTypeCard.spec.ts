import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'

const mockT = vi.fn((key: string) => key)
vi.mock('@/composables/useI18n', () => ({ useI18n: () => ({ t: mockT }) }))

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
    expect(wrapper.text()).toContain('BLOCKS')
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
    const row = wrapper.find('.relation-row')
    expect(row.text()).toContain('—')
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
    await wrapper.find('[data-test="add-link-header"]').trigger('click')
    expect(wrapper.emitted('add')).toBeTruthy()
  })

  it('shows empty state hint when items array is empty', () => {
    const wrapper = mountCard({ items: [] })
    expect(wrapper.text()).toContain('issueKanban.noRelations')
  })

  it('renders "—" for invalid/unparseable target_date', () => {
    const wrapper = mountCard({
      items: [
        {
          id: 101,
          related_issue: {
            id: 55, sequence_id: 55, name: 'OAuth2 provider setup',
            state_name: 'Todo', state_group: 'todo',
            priority: 'high',
            assignees: [],
            start_date: null, target_date: 'invalid-date',
            issue_type: { id: 2, name: 'Feature', color: '#6366f1' },
          },
          related_issue_id: 55,
        },
      ],
    })
    expect(wrapper.text()).toContain('—')
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
    const itemsAfter = wrapper.findAll('.relation-row')
    expect(itemsAfter.length).toBe(0)
  })
})
