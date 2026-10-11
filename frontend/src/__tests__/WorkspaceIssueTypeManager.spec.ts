import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import WorkspaceIssueTypeManager from '@/components/WorkspaceIssueTypeManager.vue'
import { resetProjectPermissionsCache } from '@/composables/usePermission'

const mockGet = vi.fn()
vi.mock('@/api', () => ({
  default: { get: (...args: any[]) => mockGet(...args) },
}))
vi.mock('@/api/issue-type', () => ({
  getIssueTypes: vi.fn().mockResolvedValue([
    { id: 1, name: 'Task', color: '#6366f1', icon: 'task', is_default: false, is_active: true, sequence: 1 },
  ]),
  getIssueTypeFields: vi.fn().mockResolvedValue([]),
}))
vi.mock('@/api/custom-field', () => ({
  listCustomFields: vi.fn().mockResolvedValue([]),
}))
vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))
vi.mock('@/composables/useConfirm', () => ({
  useConfirm: () => ({ confirm: vi.fn() }),
}))

describe('WorkspaceIssueTypeManager', () => {
  beforeEach(() => {
    resetProjectPermissionsCache()
    mockGet.mockReset()
  })

  it('hides create and row actions from members who can only view settings', async () => {
    mockGet.mockResolvedValue({ data: { permissions: ['settings:view'], role_level: 15 } })
    const wrapper = mount(WorkspaceIssueTypeManager, { props: { workspaceId: 3 } })
    await flushPromises()
    expect(mockGet).toHaveBeenCalledWith('/workspaces/3/my-permissions')
    expect(wrapper.text()).toContain('Task')
    expect(wrapper.find('[data-test="create-type-btn"]').exists()).toBe(false)
    expect(wrapper.find('.type-row-actions').exists()).toBe(false)
  })

  it('shows create and row actions to settings managers', async () => {
    mockGet.mockResolvedValue({ data: { permissions: ['settings:view', 'settings:manage'], role_level: 20 } })
    const wrapper = mount(WorkspaceIssueTypeManager, { props: { workspaceId: 3 } })
    await flushPromises()
    expect(wrapper.find('[data-test="create-type-btn"]').exists()).toBe(true)
    expect(wrapper.find('.type-row-actions').exists()).toBe(true)
  })
})
