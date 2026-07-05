import { describe, it, expect, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

const { mockListRelations, mockListRelationTypes, mockDeleteRelation, mockUpdateIssue } = vi.hoisted(() => ({
  mockListRelations: vi.fn().mockResolvedValue([]),
  mockListRelationTypes: vi.fn().mockResolvedValue([]),
  mockDeleteRelation: vi.fn().mockResolvedValue(undefined),
  mockUpdateIssue: vi.fn().mockResolvedValue({}),
}))

vi.mock('@/api/relation', () => ({
  listIssueRelations: (...args: any[]) => mockListRelations(...args),
  listRelationTypes: (...args: any[]) => mockListRelationTypes(...args),
  deleteIssueRelation: (...args: any[]) => mockDeleteRelation(...args),
}))

vi.mock('@/api/issue', () => ({
  issueApi: { updateIssue: (...args: any[]) => mockUpdateIssue(...args) },
}))

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))

import IssueTabRelations from '@/components/IssueTabRelations.vue'
import RelationTypeCard from '@/components/RelationTypeCard.vue'

const mockRelations = [
  {
    id: 101,
    relation_type_id: 1,
    relation_type: { id: 1, name: 'blocks', outward_name: 'blocks' },
    related_issue_id: 55,
    related_issue: {
      id: 55,
      sequence_id: 55,
      name: 'OAuth2 setup',
      state_name: 'Todo',
      state_group: 'todo',
      priority: 'high',
      assignees: [],
      target_date: null,
      issue_type: { id: 2, name: 'Feature', color: '#10b981' },
    },
  },
  {
    id: 102,
    relation_type_id: 2,
    relation_type: { id: 2, name: 'relates_to', outward_name: 'relates to' },
    related_issue_id: 60,
    related_issue: {
      id: 60,
      sequence_id: 60,
      name: 'Session timeout',
      state_name: 'Done',
      state_group: 'done',
      priority: 'medium',
      assignees: [{ id: 1, display_name: 'Alice' }],
      target_date: null,
      issue_type: { id: 1, name: 'Task', color: '#6366f1' },
    },
  },
]

const mockRelationTypes = [
  { id: 1, name: 'blocks', outward_name: 'blocks' },
  { id: 2, name: 'relates_to', outward_name: 'relates to' },
  { id: 3, name: 'duplicates', outward_name: 'duplicates' },
]

const defaultProps = {
  issueId: 123,
  projectId: 1,
  workspaceId: 1,
  parent: null,
  subIssues: [],
  issueTypes: [],
}

function mountComponent(overrides: Record<string, any> = {}) {
  return mount(IssueTabRelations, {
    props: { ...defaultProps, ...overrides },
  })
}

describe('IssueTabRelations', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockListRelations.mockResolvedValue(mockRelations)
    mockListRelationTypes.mockResolvedValue(mockRelationTypes)
  })

  it('renders Parent card', () => {
    const wrapper = mountComponent()
    expect(wrapper.text()).toContain('PARENT')
  })

  it('renders Sub-issues card', () => {
    const wrapper = mountComponent()
    expect(wrapper.text()).toContain('subIssue.title')
  })

  it('loads relations and renders one card per relation type', async () => {
    const wrapper = mountComponent()
    await flushPromises()

    expect(mockListRelations).toHaveBeenCalledWith(123)
    expect(mockListRelationTypes).toHaveBeenCalledWith(1)

    const cards = wrapper.findAllComponents(RelationTypeCard)
    expect(cards.length).toBe(2)
  })

  it('does not render cards for types with no items', async () => {
    const wrapper = mountComponent()
    await flushPromises()

    // "duplicates" type has 0 relations so no card should render for it
    const cards = wrapper.findAllComponents(RelationTypeCard)
    expect(cards.length).toBe(2)
    // Only "blocks" and "relates to" should have cards
    expect(wrapper.text()).not.toContain('duplicates')
  })

  it('handles API load failure gracefully', async () => {
    mockListRelations.mockRejectedValue(new Error('Network error'))
    const wrapper = mountComponent()
    await flushPromises()

    // Component should not throw and still render parent/sub-issues cards
    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('PARENT')
    expect(wrapper.text()).toContain('subIssue.title')
    // No relation type cards should render
    const cards = wrapper.findAllComponents(RelationTypeCard)
    expect(cards.length).toBe(0)
  })

  it('emits navigate when a child card emits navigate', async () => {
    const wrapper = mountComponent()
    await flushPromises()

    const relationRows = wrapper.findAll('.relation-row-clickable')
    expect(relationRows.length).toBeGreaterThan(0)
    await relationRows[0].trigger('click')
    expect(wrapper.emitted('navigate')).toBeTruthy()
    expect(wrapper.emitted('navigate')![0]).toEqual([55])
  })
})
