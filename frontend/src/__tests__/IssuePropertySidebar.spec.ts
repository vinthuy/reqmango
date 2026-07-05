import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import IssuePropertySidebar from '@/components/IssuePropertySidebar.vue'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))
vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ success: vi.fn(), error: vi.fn() }),
}))

const mockIssue = {
  id: 42, state_id: 1, priority: 'high',
  assignees: [{ id: 1, display_name: 'Alice' }],
  cycle_id: 5, start_date: '2026-07-01', target_date: '2026-07-10',
}

const mockStates = [
  { id: 1, name: 'To Do' },
  { id: 2, name: 'In Progress' },
]
const mockMembers = [
  { id: 1, display_name: 'Alice' },
  { id: 2, display_name: 'Bob' },
]
const mockCycles = [
  { id: 5, name: 'Sprint 5' },
  { id: 6, name: 'Sprint 6' },
]
const mockModules = [
  { id: 1, name: 'Module A' },
  { id: 2, name: 'Module B' },
]

const mountOptions = {
  props: {
    issue: mockIssue,
    states: mockStates,
    members: mockMembers,
    cycles: mockCycles,
    modules: mockModules,
    customFields: [],
  },
}

describe('IssuePropertySidebar', () => {
  it('renders State select with label "issue.state"', () => {
    const wrapper = mount(IssuePropertySidebar, mountOptions)
    expect(wrapper.text()).toContain('issue.state')
    const selects = wrapper.findAll('select')
    expect(selects[0].exists()).toBe(true)
  })

  it('renders Priority select with label "issue.priority"', () => {
    const wrapper = mount(IssuePropertySidebar, mountOptions)
    expect(wrapper.text()).toContain('issue.priority')
    const selects = wrapper.findAll('select')
    expect(selects[1].exists()).toBe(true)
  })

  it('renders Assignee select with label "issue.assignee"', () => {
    const wrapper = mount(IssuePropertySidebar, mountOptions)
    expect(wrapper.text()).toContain('issue.assignee')
    const selects = wrapper.findAll('select')
    expect(selects[2].exists()).toBe(true)
  })

  it('emits update:state on state change', () => {
    const wrapper = mount(IssuePropertySidebar, mountOptions)
    const selects = wrapper.findAll('select')
    selects[0].setValue(2)
    expect(wrapper.emitted('update:state')).toBeTruthy()
    expect(wrapper.emitted('update:state')![0]).toEqual([2])
  })

  it('emits update:priority on priority change', () => {
    const wrapper = mount(IssuePropertySidebar, mountOptions)
    const selects = wrapper.findAll('select')
    selects[1].setValue('urgent')
    expect(wrapper.emitted('update:priority')).toBeTruthy()
    expect(wrapper.emitted('update:priority')![0]).toEqual(['urgent'])
  })

  it('renders start date and target date inputs', () => {
    const wrapper = mount(IssuePropertySidebar, mountOptions)
    const dateInputs = wrapper.findAll('input[type="date"]')
    expect(dateInputs.length).toBe(2)
    expect(dateInputs[0].element.value).toBe('2026-07-01')
    expect(dateInputs[1].element.value).toBe('2026-07-10')
  })

  it('emits update:startDate on date change', () => {
    const wrapper = mount(IssuePropertySidebar, mountOptions)
    const dateInputs = wrapper.findAll('input[type="date"]')
    dateInputs[0].setValue('2026-08-01')
    expect(wrapper.emitted('update:startDate')).toBeTruthy()
    expect(wrapper.emitted('update:startDate')![0]).toEqual(['2026-08-01'])
  })

})
