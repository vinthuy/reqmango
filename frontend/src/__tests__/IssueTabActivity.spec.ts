import { describe, it, expect, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import IssueTabActivity from '@/components/IssueTabActivity.vue'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))

const { mockGetActivities } = vi.hoisted(() => ({
  mockGetActivities: vi.fn().mockResolvedValue([]),
}))

vi.mock('@/api/issue', () => ({
  getIssueActivities: (...args: any[]) => mockGetActivities(...args),
}))

const mockActivities = [
  {
    id: 1,
    created_at: '2026-07-01T10:00:00Z',
    verb: 'updated',
    field: 'priority',
    comment: '',
    actor_id: 1,
    actor_display_name: 'User',
    old_value: 'low',
    new_value: 'high',
  },
  {
    id: 2,
    created_at: '2026-07-02T14:30:00Z',
    verb: 'created',
    field: '',
    comment: '',
    actor_id: 2,
    actor_display_name: 'Admin',
  },
]

describe('IssueTabActivity', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading state initially', () => {
    mockGetActivities.mockResolvedValue([])
    const wrapper = mount(IssueTabActivity, {
      props: { issueId: 1 },
    })
    expect(wrapper.text()).toContain('common.loading')
  })

  it('shows empty state when API returns empty array', async () => {
    mockGetActivities.mockResolvedValue([])
    const wrapper = mount(IssueTabActivity, {
      props: { issueId: 1 },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('issue.noActivity')
  })

  it('renders activities when API returns data', async () => {
    mockGetActivities.mockResolvedValue(mockActivities)
    const wrapper = mount(IssueTabActivity, {
      props: { issueId: 1 },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('activity.changedPriority')
    expect(wrapper.text()).toContain('activity.created')
    expect(wrapper.text()).toContain('User')
    expect(wrapper.text()).toContain('Admin')
  })

  it('handles API error gracefully', async () => {
    mockGetActivities.mockRejectedValue(new Error('Network error'))
    const wrapper = mount(IssueTabActivity, {
      props: { issueId: 1 },
    })
    await flushPromises()
    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).not.toContain('common.loading')
  })
})
