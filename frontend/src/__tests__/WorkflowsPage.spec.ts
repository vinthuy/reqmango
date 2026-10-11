import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import Antd from 'ant-design-vue'
import WorkflowsPage from '@/views/agent/WorkflowsPage.vue'
import { resetProjectPermissionsCache } from '@/composables/usePermission'

const mockGet = vi.fn()
vi.mock('@/api', () => ({
  default: { get: (...args: unknown[]) => mockGet(...args) },
}))
vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))
vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { id: '7', slug: 'ws' } }),
  useRouter: () => ({ push: vi.fn() }),
}))
vi.mock('@/api/workflow', () => ({
  workflowApi: {
    list: vi.fn().mockResolvedValue([]),
    listRuns: vi.fn().mockResolvedValue({ data: { data: [] } }),
  },
}))
function mountPage() {
  return mount(WorkflowsPage, {
    global: { plugins: [Antd] },
  })
}

describe('WorkflowsPage', () => {
  beforeEach(() => {
    resetProjectPermissionsCache()
    mockGet.mockReset()
    mockGet.mockResolvedValue({ data: { permissions: ['workflow:execute'], role_level: 15 } })
  })

  it('shows create workflow only to users who may manage workflows in the project', async () => {
    const member = mountPage()
    await flushPromises()
    expect(mockGet).toHaveBeenCalledWith('/projects/7/my-permissions')
    expect(member.find('[data-test="create-workflow-btn"]').exists()).toBe(false)

    resetProjectPermissionsCache()
    mockGet.mockResolvedValue({ data: { permissions: ['workflow:manage'], role_level: 20 } })
    const admin = mountPage()
    await flushPromises()
    expect(admin.find('[data-test="create-workflow-btn"]').exists()).toBe(true)
  })
})
