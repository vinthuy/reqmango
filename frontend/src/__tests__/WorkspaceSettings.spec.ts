import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import WorkspaceSettings from '@/views/WorkspaceSettings.vue'
import { resetProjectPermissionsCache } from '@/composables/usePermission'

const mockGet = vi.fn()
vi.mock('@/api', () => ({
  default: {
    get: (...args: unknown[]) => mockGet(...args),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))
vi.mock('@/api/workspace', () => ({
  workspaceApi: {
    getBySlug: vi.fn().mockResolvedValue({ id: 3, name: 'acme', slug: 'acme' }),
    listMembers: vi.fn().mockResolvedValue([]),
  },
}))
vi.mock('@/api/project', () => ({
  listProjects: vi.fn().mockResolvedValue([]),
}))
vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))
vi.mock('@/composables/useConfirm', () => ({
  useConfirm: () => ({ confirm: vi.fn() }),
}))
vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ success: vi.fn(), error: vi.fn(), info: vi.fn(), warning: vi.fn() }),
}))
vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { slug: 'acme' }, query: { section: 'states' } }),
  useRouter: () => ({ replace: vi.fn(), push: vi.fn() }),
}))

function primeApi(permissions: string[], roleLevel: number) {
  mockGet.mockImplementation((url: string) => {
    if (url.includes('my-permissions')) {
      return Promise.resolve({ data: { permissions, role_level: roleLevel } })
    }
    if (url.includes('/settings/states')) {
      return Promise.resolve({ data: [{ id: 1, name: 'Todo', color: '#6B7280', group: 'backlog' }] })
    }
    return Promise.resolve({ data: [] })
  })
}

function mountSection() {
  return mount(WorkspaceSettings, { shallow: true })
}

describe('WorkspaceSettings workspace states', () => {
  beforeEach(() => {
    resetProjectPermissionsCache()
    mockGet.mockReset()
  })

  it('hides state mutations from members who may only view settings', async () => {
    primeApi(['settings:view'], 15)
    const wrapper = mountSection()
    await flushPromises()

    expect(mockGet).toHaveBeenCalledWith('/workspaces/3/my-permissions')
    expect(wrapper.find('[data-test="workspace-states-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="create-workspace-state-btn"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="workspace-state-row-actions"]').exists()).toBe(false)
  })

  it('shows state mutations to settings managers', async () => {
    primeApi(['settings:view', 'settings:manage'], 20)
    const wrapper = mountSection()
    await flushPromises()

    expect(wrapper.find('[data-test="workspace-states-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="create-workspace-state-btn"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="workspace-state-row-actions"]').exists()).toBe(true)
  })
})
