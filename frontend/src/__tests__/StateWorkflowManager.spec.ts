import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import StateWorkflowManager from '@/components/StateWorkflowManager.vue'

const get = vi.fn()
vi.mock('@/api', () => ({ default: { get: (...a: unknown[]) => get(...a) } }))
vi.mock('@/composables/useI18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }))
vi.mock('@/composables/useConfirm', () => ({ useConfirm: () => ({ confirm: vi.fn() }) }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ user: { id: 1, is_superuser: true } }) }))

describe('StateWorkflowManager', () => {
  beforeEach(() => {
    get.mockReset()
  })

  it('only lists workspace workflows', async () => {
    get.mockImplementation((url: string) => {
      if (url.endsWith('/workflows')) return Promise.resolve({ data: [] })
      if (url.endsWith('/settings/states')) return Promise.resolve({ data: { data: [] } })
      if (url === '/issue-types') return Promise.resolve({ data: { data: [] } })
      return Promise.resolve({ data: [] })
    })
    mount(StateWorkflowManager, { props: { workspaceId: 1 } })
    await flushPromises()
    expect(get).toHaveBeenCalledWith('/workspaces/1/workflows')
    expect(get.mock.calls.some(([u]) => String(u).includes('/projects/'))).toBe(false)
  })

  it('renders the transitions returned by the workspace route', async () => {
    get.mockImplementation((url: string) => {
      if (url === '/workspaces/1/workflows') {
        return Promise.resolve({
          data: [{
            id: 5,
            name: '默认流程',
            description: '',
            transitions: [{ id: 9, from_name: 'Todo', to_name: 'Done', rule_type: 'allow' }],
          }],
        })
      }
      if (url.endsWith('/settings/states')) {
        return Promise.resolve({ data: { data: [{ id: 1, name: 'Todo' }, { id: 2, name: 'Done' }] } })
      }
      return Promise.resolve({ data: { data: [] } })
    })
    const wrapper = mount(StateWorkflowManager, { props: { workspaceId: 1 } })
    await flushPromises()
    expect(wrapper.text()).toContain('默认流程')
    expect(wrapper.text()).toContain('Todo')
    expect(wrapper.text()).toContain('Done')
  })

  // `canManage` is declared as an optional boolean. Vue casts an absent
  // boolean prop to `false`, so a `!== undefined` guard silently hides every
  // mutating control unless the prop carries an explicit `default: undefined`.
  it('shows the create control when canManage is omitted and the user is a superuser', async () => {
    get.mockImplementation(() => Promise.resolve({ data: [] }))
    const wrapper = mount(StateWorkflowManager, { props: { workspaceId: 1 } })
    await flushPromises()
    expect(wrapper.text()).toContain('workflow.create')
  })

  it('hides the create control when canManage is explicitly false', async () => {
    get.mockImplementation(() => Promise.resolve({ data: [] }))
    const wrapper = mount(StateWorkflowManager, { props: { workspaceId: 1, canManage: false } })
    await flushPromises()
    expect(wrapper.text()).not.toContain('workflow.create')
  })

  // The empty-state hint must wait for the request to settle; showing it while
  // the list is still in flight makes a populated workspace look empty.
  it('shows the empty hint only after the workflow request settles', async () => {
    let resolveWorkflows: (value: unknown) => void = () => {}
    get.mockImplementation((url: string) => {
      if (url === '/workspaces/1/workflows') {
        return new Promise((resolve) => { resolveWorkflows = resolve })
      }
      return Promise.resolve({ data: { data: [] } })
    })
    const wrapper = mount(StateWorkflowManager, { props: { workspaceId: 1 } })
    await flushPromises()
    expect(wrapper.text()).not.toContain('workflow.empty')

    resolveWorkflows({ data: [] })
    await flushPromises()
    expect(wrapper.text()).toContain('workflow.empty')
  })
})
