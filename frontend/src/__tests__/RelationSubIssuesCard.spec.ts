import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))

import RelationSubIssuesCard from '@/components/RelationSubIssuesCard.vue'

const mockSubIssues = [
  {
    id: 43,
    sequence_id: 43,
    name: 'Email validation',
    state_name: 'Done',
    state_group: 'done',
    priority: 'low',
    assignees: [{ id: 2, display_name: 'Bob' }],
    target_date: '2026-07-08T00:00:00Z',
    issue_type: { id: 1, name: 'Task', color: '#6366f1' },
  },
  {
    id: 44,
    sequence_id: 44,
    name: 'Password checker',
    state_name: 'Todo',
    state_group: 'todo',
    priority: 'high',
    assignees: [],
    target_date: null,
    issue_type: { id: 1, name: 'Task', color: '#6366f1' },
  },
]

function mountCard(overrides: Record<string, any> = {}) {
  const defaults = {
    subIssues: mockSubIssues,
  }
  return mount(RelationSubIssuesCard, {
    props: { ...defaults, ...overrides },
  })
}

describe('RelationSubIssuesCard', () => {
  it('renders column headers in the table', () => {
    const wrapper = mountCard()
    const text = wrapper.text()
    expect(text).toContain('subIssue.title')
  })

  it('shows completion count 1/2', () => {
    const wrapper = mountCard()
    expect(wrapper.text()).toContain('1/2')
  })

  it('renders each sub-issue as a row', () => {
    const wrapper = mountCard()
    const text = wrapper.text()
    expect(text).toContain('Email validation')
    expect(text).toContain('Password checker')
  })

  it('emits "add" when + Add button is clicked', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-test="add-subissue"]').trigger('click')
    expect(wrapper.emitted('add')).toBeTruthy()
  })

  it('emits "navigate" with issue ID when title is clicked', async () => {
    const wrapper = mountCard()
    await wrapper.find('.subissue-clickable').trigger('click')
    expect(wrapper.emitted('navigate')).toBeTruthy()
    expect(wrapper.emitted('navigate')![0]).toEqual([43])
  })

  it('emits "toggle" when checkbox is clicked', async () => {
    const wrapper = mountCard()
    const checkbox = wrapper.find('input[type="checkbox"]')
    await checkbox.trigger('change')
    expect(wrapper.emitted('toggle')).toBeTruthy()
  })

  it('shows empty state when no sub-issues', () => {
    const wrapper = mountCard({ subIssues: [] })
    expect(wrapper.text()).toContain('subIssue.noSubIssues')
  })
})
