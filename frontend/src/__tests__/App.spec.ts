import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { reactive, nextTick } from 'vue'
import App from '@/App.vue'

const authStore = vi.hoisted(() => ({ token: '', fetchUser: vi.fn() }))
let route = reactive({ meta: { requiresAuth: false } })
vi.mock('vue-router', () => ({ useRoute: () => route }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => authStore }))

describe('App authentication shell', () => {
  let wrapper: VueWrapper | undefined
  beforeEach(() => {
    route = reactive({ meta: { requiresAuth: false } })
    authStore.token = ''
    authStore.fetchUser.mockReset()
  })
  afterEach(() => wrapper?.unmount())

  function render() {
    wrapper = mount(App, {
      global: {
        stubs: ['TopBar', 'ConfirmDialog', 'ShortcutsPanel', 'ToastContainer', 'router-view'],
      },
    })
    return wrapper
  }

  it('keeps public routes free of authenticated navigation and user requests', () => {
    expect(render().find('top-bar-stub').exists()).toBe(false)
    expect(authStore.fetchUser).not.toHaveBeenCalled()
  })

  it('does not validate stale login tokens on public routes', () => {
    authStore.token = 'stale-token'
    expect(render().find('top-bar-stub').exists()).toBe(false)
    expect(authStore.fetchUser).not.toHaveBeenCalled()
  })

  it('shows navigation and loads the user on authenticated routes', () => {
    route.meta.requiresAuth = true
    authStore.token = 'valid-token'
    expect(render().find('top-bar-stub').exists()).toBe(true)
    expect(authStore.fetchUser).toHaveBeenCalledOnce()
  })

  it('unmounts authenticated navigation when moving to a public route', async () => {
    route.meta.requiresAuth = true
    const app = render()
    expect(app.find('top-bar-stub').exists()).toBe(true)
    route.meta.requiresAuth = false
    await nextTick()
    expect(app.find('top-bar-stub').exists()).toBe(false)
  })
})
