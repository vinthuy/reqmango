import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ProjectSettings from '@/views/ProjectSettings.vue'
import { resetProjectPermissionsCache } from '@/composables/usePermission'

const routeState = vi.hoisted(() => ({ section: 'states' }))
const mockGet = vi.fn()
vi.mock('@/api', () => ({
  default: {
    get: (...args: unknown[]) => mockGet(...args),
    post: vi.fn().mockResolvedValue({ data: {} }),
    put: vi.fn().mockResolvedValue({ data: {} }),
    delete: vi.fn().mockResolvedValue({ data: {} }),
  },
}))
vi.mock('vue-router', () => ({
  RouterLink: { template: '<a><slot /></a>' },
  useRoute: () => ({ params: { slug: 'acme', id: '7' }, query: { section: routeState.section } }),
  useRouter: () => ({ replace: vi.fn(), push: vi.fn() }),
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
  return mount(ProjectSettings, { shallow: true })
}

describe('ProjectSettings states and labels', () => {
  beforeEach(() => {
    resetProjectPermissionsCache()
    mockGet.mockReset()
    routeState.section = 'states'
  })

  it('hides state and label mutations from members who may only view settings', async () => {
    primeApi(['settings:view'], 15)

    const states = mountSection()
    await flushPromises()
    expect(mockGet).toHaveBeenCalledWith('/projects/7/my-permissions')
    expect(states.find('[data-test="project-states-section"]').exists()).toBe(true)
    expect(states.find('[data-test="create-state-btn"]').exists()).toBe(false)
    expect(states.find('[data-test="project-state-row-actions"]').exists()).toBe(false)

    routeState.section = 'labels'
    const labels = mountSection()
    await flushPromises()
    expect(labels.find('[data-test="project-labels-section"]').exists()).toBe(true)
    expect(labels.find('[data-test="create-label-btn"]').exists()).toBe(false)
  })

  it('shows state mutations to settings managers', async () => {
    primeApi(['settings:view', 'settings:manage'], 20)
    const wrapper = mountSection()
    await flushPromises()

    expect(wrapper.find('[data-test="project-states-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="create-state-btn"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="project-state-row-actions"]').exists()).toBe(true)
  })
})
