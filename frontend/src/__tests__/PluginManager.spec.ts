/**
 * PluginManager.vue - Full component functional tests
 *
 * Functional test cases covered:
 *  1. Initial loading state with spinner
 *  2. Catalog rendering - plugin cards with metadata
 *  3. Installed plugins list with status indicators
 *  4. Install flow - click Install → API call → appears in installed
 *  5. Uninstall flow - click Remove → confirm → API call → removed
 *  6. Enable/Disable toggle - both card and list toggle
 *  7. Config modal - open, edit config JSON, save, cancel, error handling
 *  8. Logs modal - open, view logs, close, empty state
 *  9. Test execution - click Test → see result, clear result
 * 10. Event subscriptions displayed on catalog cards
 * 11. Empty state when no plugins installed
 * 12. Error handling - API failures show alerts
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

// ============================================================
// vi.hoisted mock data – these are preserved through mock hoisting
// ============================================================
const { mockApi, mockCatalog, mockInstalledPlugins, mockLogs } = vi.hoisted(() => {
  const catalog = [
    {
      slug: 'outgoing-webhook',
      name: 'Outgoing Webhook',
      description: 'Send outgoing webhooks on events',
      author: 'ReqMango',
      version: '1.0.0',
      type: 'webhook' as const,
      icon_url: '',
      entry_point: 'https://hooks.example.com',
      config_schema: { url: { type: 'string', required: true } },
      subscribed_events: ['issue.created', 'issue.updated'],
    },
    {
      slug: 'slack-notifier',
      name: 'Slack Notifier',
      description: 'Send Slack notifications on events',
      author: 'ReqMango',
      version: '1.0.0',
      type: 'notification' as const,
      icon_url: '',
      entry_point: 'https://slack.com/api',
      config_schema: { webhook_url: { type: 'string' } },
      subscribed_events: ['issue.created'],
    },
  ]

  const installed = [
    {
      id: 1,
      name: 'Outgoing Webhook',
      slug: 'outgoing-webhook',
      description: 'Send outgoing webhooks on events',
      author: 'ReqMango',
      version: '1.0.0',
      icon_url: '',
      type: 'webhook',
      entry_point: 'https://hooks.example.com',
      config_schema: { url: { type: 'string', required: true } },
      config: { url: 'https://myhook.com' },
      subscribed_events: ['issue.created', 'issue.updated'],
      enabled: true,
      workspace_id: 1,
      installed_by_id: 1,
      created_at: '2026-07-04T12:00:00Z',
      updated_at: '2026-07-04T12:00:00Z',
    },
  ]

  const logs = [
    {
      id: 1, plugin_id: 1, event_type: 'issue.created',
      status: 'success' as const, request_body: '{}',
      response_body: '{"ok":true}', status_code: 200,
      duration_ms: 150, created_at: '2026-07-04T12:00:00Z',
    },
    {
      id: 2, plugin_id: 1, event_type: 'issue.updated',
      status: 'error' as const, request_body: '{}',
      response_body: '{"error":"timeout"}', status_code: 504,
      duration_ms: 3000, created_at: '2026-07-04T11:00:00Z',
    },
    {
      id: 3, plugin_id: 1, event_type: 'comment.created',
      status: 'skipped' as const, request_body: '',
      response_body: '', status_code: 0,
      duration_ms: 0, created_at: '2026-07-04T10:00:00Z',
    },
  ]

  const api = {
    catalog: vi.fn(),
    list: vi.fn(),
    install: vi.fn(),
    get: vi.fn(),
    update: vi.fn(),
    uninstall: vi.fn(),
    enable: vi.fn(),
    disable: vi.fn(),
    logs: vi.fn(),
    test: vi.fn(),
  }

  return { mockApi: api, mockCatalog: catalog, mockInstalledPlugins: installed, mockLogs: logs }
})

// ============================================================
// Mock the pluginApi module
// ============================================================
vi.mock('@/api/plugin', () => ({
  pluginApi: mockApi,
}))

// Now safe to import component
import PluginManager from '@/views/PluginManager.vue'

// ============================================================
// Setup
// ============================================================
function mountComponent() {
  return mount(PluginManager, {
    props: { workspaceId: 1 },
  })
}

function setupDefaultData() {
  mockApi.catalog.mockResolvedValue(mockCatalog)
  mockApi.list.mockResolvedValue([])
}

function setupWithInstalled() {
  mockApi.catalog.mockResolvedValue(mockCatalog)
  mockApi.list.mockResolvedValue(mockInstalledPlugins)
}

beforeEach(() => {
  vi.clearAllMocks()
  // Re-stub globals after clearAllMocks
  vi.stubGlobal('confirm', vi.fn((_msg?: string) => true))
  vi.stubGlobal('alert', vi.fn())
  setupDefaultData()
})

// ============================================================
// 1. INITIAL LOADING STATE
// ============================================================
describe('Initial loading state', () => {
  it('should show loading spinner on mount', () => {
    const wrapper = mountComponent()
    expect(wrapper.find('.pm-loading').exists()).toBe(true)
    expect(wrapper.find('.spinner').exists()).toBe(true)
    expect(wrapper.text()).toContain('Loading plugins...')
  })

  it('should hide loading spinner after data resolves', async () => {
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.find('.pm-loading').exists()).toBe(false)
  })

  it('should call catalog and list APIs on mount', async () => {
    mountComponent()
    await flushPromises()
    expect(mockApi.catalog).toHaveBeenCalledWith(1)
    expect(mockApi.list).toHaveBeenCalledWith(1)
  })

  it('should handle API load failure gracefully', async () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(vi.fn())
    mockApi.catalog.mockRejectedValue(new Error('Network Error'))
    mockApi.list.mockRejectedValue(new Error('Network Error'))
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.find('.pm-loading').exists()).toBe(false)
    spy.mockRestore()
  })
})

// ============================================================
// 2. CATALOG RENDERING
// ============================================================
describe('Catalog rendering', () => {
  it('should render all catalog items as cards', async () => {
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.findAll('.pm-card')).toHaveLength(mockCatalog.length)
  })

  it('should display plugin name, description, version, author', async () => {
    const wrapper = mountComponent()
    await flushPromises()
    const card = wrapper.find('.pm-card')
    expect(card.text()).toContain(mockCatalog[0].name)
    expect(card.text()).toContain(mockCatalog[0].description)
    expect(card.text()).toContain(`v${mockCatalog[0].version}`)
    expect(card.text()).toContain(`by ${mockCatalog[0].author}`)
  })

  it('should display plugin type icon and label', async () => {
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.find('.pm-card-icon').text()).toBe('🔗')
    expect(wrapper.find('.pm-card-type').text()).toBe('Webhook')
  })

  it('should display event subscription tags', async () => {
    const wrapper = mountComponent()
    await flushPromises()
    const tags = wrapper.findAll('.pm-event-tag')
    expect(tags.length).toBeGreaterThan(0)
    expect(tags[0].text()).toBe('Issue Created')
  })

  it('should show "Installed" badge for installed plugins', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.find('.pm-badge').text()).toBe('Installed')
  })

  it('should NOT show "Installed" badge for uninstalled plugins', async () => {
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.find('.pm-badge').exists()).toBe(false)
  })
})

// ============================================================
// 3. INSTALLED LIST RENDERING
// ============================================================
describe('Installed plugins list', () => {
  it('should render installed items when plugins exist', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.findAll('.pm-installed-item')).toHaveLength(mockInstalledPlugins.length)
  })

  it('should show name and version in installed list', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const item = wrapper.find('.pm-installed-item')
    expect(item.text()).toContain(mockInstalledPlugins[0].name)
    expect(item.text()).toContain(`v${mockInstalledPlugins[0].version}`)
  })

  it('should show active dot when enabled', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const dot = wrapper.find('.pm-status-dot')
    expect(dot.classes()).toContain('active')
  })

  it('should show disabled opacity and no active dot when disabled', async () => {
    mockApi.list.mockResolvedValue([{ ...mockInstalledPlugins[0], enabled: false }])
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.find('.pm-installed-item').classes()).toContain('disabled')
    expect(wrapper.find('.pm-status-dot').classes()).not.toContain('active')
  })

  it('should have Config/Logs/Test/Remove action buttons', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const actions = wrapper.find('.pm-item-actions')
    expect(actions.text()).toContain('Config')
    expect(actions.text()).toContain('Logs')
    expect(actions.text()).toContain('Test')
    expect(actions.text()).toContain('Remove')
  })

  it('should have toggle switch with correct checked state', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const toggle = wrapper.find('.pm-toggle input[type="checkbox"]')
    expect((toggle.element as HTMLInputElement).checked).toBe(true)
  })
})

// ============================================================
// 4. INSTALL FLOW
// ============================================================
describe('Install flow', () => {
  it('should show Install button for uninstalled plugins', async () => {
    mockApi.catalog.mockResolvedValue([mockCatalog[1]])
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.find('.pm-card .btn-primary')
    expect(btn.exists()).toBe(true)
    expect(btn.text()).toBe('Install')
  })

  it('should call install API on button click', async () => {
    mockApi.catalog.mockResolvedValue([mockCatalog[1]])
    mockApi.install.mockResolvedValue({
      id: 2, slug: 'slack-notifier', name: 'Slack Notifier', enabled: true,
      type: 'notification', version: '1.0.0', config: {}, config_schema: {},
      subscribed_events: [], workspace_id: 1, installed_by_id: 1,
      created_at: '', updated_at: '', description: '', author: '', icon_url: '', entry_point: '',
    })
    const wrapper = mountComponent()
    await flushPromises()
    await wrapper.find('.pm-card .btn-primary').trigger('click')
    await flushPromises()
    expect(mockApi.install).toHaveBeenCalledWith(1, { slug: 'slack-notifier' })
  })

  it('should show error alert on install failure', async () => {
    mockApi.catalog.mockResolvedValue([mockCatalog[1]])
    mockApi.install.mockRejectedValue({
      response: { data: { message: 'Plugin already installed' } },
    })
    const wrapper = mountComponent()
    await flushPromises()
    await wrapper.find('.pm-card .btn-primary').trigger('click')
    await flushPromises()
    expect(window.alert).toHaveBeenCalledWith('Plugin already installed')
  })

  it('should show generic error on install failure without response', async () => {
    mockApi.catalog.mockResolvedValue([mockCatalog[1]])
    mockApi.install.mockRejectedValue(new Error('network'))
    const wrapper = mountComponent()
    await flushPromises()
    await wrapper.find('.pm-card .btn-primary').trigger('click')
    await flushPromises()
    expect(window.alert).toHaveBeenCalledWith('Failed to install plugin')
  })
})

// ============================================================
// 5. UNINSTALL FLOW
// ============================================================
describe('Uninstall flow', () => {
  it('should prompt confirmation before uninstall', async () => {
    setupWithInstalled()
    vi.mocked(window.confirm).mockReturnValue(true)
    mockApi.uninstall.mockResolvedValue(undefined)
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('button').find(b => b.text() === 'Remove')!
    await btn.trigger('click')
    expect(window.confirm).toHaveBeenCalledWith('Are you sure you want to uninstall this plugin?')
  })

  it('should NOT uninstall if user cancels', async () => {
    setupWithInstalled()
    vi.mocked(window.confirm).mockReturnValue(false)
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('button').find(b => b.text() === 'Remove')!
    await btn.trigger('click')
    expect(mockApi.uninstall).not.toHaveBeenCalled()
  })

  it('should call uninstall and show empty state on confirm', async () => {
    setupWithInstalled()
    vi.mocked(window.confirm).mockReturnValue(true)
    mockApi.uninstall.mockResolvedValue(undefined)
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.findAll('.pm-installed-item')).toHaveLength(1)
    const btn = wrapper.findAll('button').find(b => b.text() === 'Remove')!
    await btn.trigger('click')
    await flushPromises()
    expect(mockApi.uninstall).toHaveBeenCalledWith(1, 1)
    expect(wrapper.find('.pm-empty').exists()).toBe(true)
  })

  it('should show alert on uninstall failure', async () => {
    setupWithInstalled()
    vi.mocked(window.confirm).mockReturnValue(true)
    mockApi.uninstall.mockRejectedValue({
      response: { data: { message: 'Cannot uninstall: locked' } },
    })
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('button').find(b => b.text() === 'Remove')!
    await btn.trigger('click')
    await flushPromises()
    expect(window.alert).toHaveBeenCalledWith('Cannot uninstall: locked')
  })
})

// ============================================================
// 6. ENABLE / DISABLE
// ============================================================
describe('Enable/Disable toggle', () => {
  it('should call disable API when toggling off', async () => {
    setupWithInstalled()
    mockApi.disable.mockResolvedValue({ ...mockInstalledPlugins[0], enabled: false })
    const wrapper = mountComponent()
    await flushPromises()
    await wrapper.find('.pm-toggle input[type="checkbox"]').setValue(false)
    await flushPromises()
    expect(mockApi.disable).toHaveBeenCalledWith(1, 1)
  })

  it('should call enable API when toggling on disabled plugin', async () => {
    mockApi.list.mockResolvedValue([{ ...mockInstalledPlugins[0], enabled: false }])
    mockApi.enable.mockResolvedValue({ ...mockInstalledPlugins[0], enabled: true })
    const wrapper = mountComponent()
    await flushPromises()
    const input = wrapper.find('.pm-toggle input[type="checkbox"]')
    expect((input.element as HTMLInputElement).checked).toBe(false)
    await input.setValue(true)
    await flushPromises()
    expect(mockApi.enable).toHaveBeenCalledWith(1, 1)
  })

  it('should show alert on toggle failure', async () => {
    mockApi.list.mockResolvedValue([{ ...mockInstalledPlugins[0], enabled: false }])
    mockApi.enable.mockRejectedValue({
      response: { data: { message: 'Cannot enable: locked' } },
    })
    const wrapper = mountComponent()
    await flushPromises()
    // Click toggle on disabled plugin - should trigger enable which fails
    const checkbox = wrapper.find('.pm-toggle input[type="checkbox"]')
    await checkbox.setValue(true)
    await flushPromises()
    expect(window.alert).toHaveBeenCalledWith('Cannot enable: locked')
  })
})

// ============================================================
// 7. CONFIG MODAL
// ============================================================
describe('Config modal', () => {
  it('should open config modal on Config button click', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Config')!
    await btn.trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-modal-config').exists()).toBe(true)
    expect(wrapper.find('.pm-modal-config h3').text()).toContain('Outgoing Webhook')
  })

  it('should display config schema if available', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Config')!
    await btn.trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-config-schema').exists()).toBe(true)
  })

  it('should display current config JSON in textarea', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Config')!
    await btn.trigger('click')
    await flushPromises()
    const val = (wrapper.find('.pm-config-textarea').element as HTMLTextAreaElement).value
    expect(JSON.parse(val)).toEqual({ url: 'https://myhook.com' })
  })

  it('should save config and close modal', async () => {
    setupWithInstalled()
    mockApi.update.mockResolvedValue({
      ...mockInstalledPlugins[0], config: { url: 'https://updated.com' },
    })
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Config')!
    await btn.trigger('click')
    await flushPromises()
    await wrapper.find('.pm-config-textarea').setValue('{"url":"https://updated.com"}')
    await wrapper.find('.pm-modal-footer .btn-primary').trigger('click')
    await flushPromises()
    expect(mockApi.update).toHaveBeenCalledWith(1, 1, { config: { url: 'https://updated.com' } })
    expect(wrapper.find('.pm-modal-config').exists()).toBe(false)
  })

  it('should show error for invalid JSON', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Config')!
    await btn.trigger('click')
    await flushPromises()
    await wrapper.find('.pm-config-textarea').setValue('{invalid')
    await wrapper.find('.pm-modal-footer .btn-primary').trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-config-error').text()).toBe('Invalid JSON format')
    expect(mockApi.update).not.toHaveBeenCalled()
  })

  it('should show API error on save failure', async () => {
    setupWithInstalled()
    mockApi.update.mockRejectedValue({
      response: { data: { message: 'Config validation failed' } },
    })
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Config')!
    await btn.trigger('click')
    await flushPromises()
    await wrapper.find('.pm-config-textarea').setValue('{}')
    await wrapper.find('.pm-modal-footer .btn-primary').trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-config-error').text()).toBe('Config validation failed')
  })

  it('should close modal on Cancel click', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Config')!
    await btn.trigger('click')
    await flushPromises()
    await wrapper.find('.pm-modal-footer .btn-outline').trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-modal-config').exists()).toBe(false)
  })

  it('should close modal on X button click', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Config')!
    await btn.trigger('click')
    await flushPromises()
    await wrapper.find('.pm-modal-close').trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-modal-config').exists()).toBe(false)
  })

  it('should open config from catalog card Config button', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.pm-card-actions .btn-outline').find(b => b.text() === 'Config')!
    await btn.trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-modal-config').exists()).toBe(true)
  })
})

// ============================================================
// 8. LOGS MODAL
// ============================================================
describe('Logs modal', () => {
  it('should open logs modal on Logs button click', async () => {
    setupWithInstalled()
    mockApi.logs.mockResolvedValue(mockLogs)
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Logs')!
    await btn.trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-modal-header h3').text()).toContain('Outgoing Webhook')
  })

  it('should fetch logs with limit 50', async () => {
    setupWithInstalled()
    mockApi.logs.mockResolvedValue([])
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Logs')!
    await btn.trigger('click')
    await flushPromises()
    expect(mockApi.logs).toHaveBeenCalledWith(1, 1, 50)
  })

  it('should display log entries with status/event/duration/HTTP code', async () => {
    setupWithInstalled()
    mockApi.logs.mockResolvedValue(mockLogs)
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Logs')!
    await btn.trigger('click')
    await flushPromises()
    const entries = wrapper.findAll('.pm-log-entry')
    expect(entries).toHaveLength(mockLogs.length)
    expect(entries[0].classes()).toContain('success')
    expect(entries[0].text()).toContain('150ms')
    expect(entries[0].text()).toContain('HTTP 200')
    expect(entries[1].classes()).toContain('error')
    expect(entries[1].text()).toContain('HTTP 504')
    expect(entries[2].classes()).toContain('skipped')
  })

  it('should show response body details when present', async () => {
    setupWithInstalled()
    mockApi.logs.mockResolvedValue([mockLogs[0]])
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Logs')!
    await btn.trigger('click')
    await flushPromises()
    const detail = wrapper.find('.pm-log-detail')
    expect(detail.exists()).toBe(true)
    expect(detail.find('pre').text()).toContain('{"ok":true}')
  })

  it('should show empty state when no logs', async () => {
    setupWithInstalled()
    mockApi.logs.mockResolvedValue([])
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Logs')!
    await btn.trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-logs-empty').text()).toBe('No execution logs yet.')
  })

  it('should handle logs API failure gracefully', async () => {
    setupWithInstalled()
    const spy = vi.spyOn(console, 'error').mockImplementation(vi.fn())
    mockApi.logs.mockRejectedValue(new Error('Failed'))
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Logs')!
    await btn.trigger('click')
    await flushPromises()
    expect(spy).toHaveBeenCalledWith('Failed to load logs:', expect.anything())
    expect(wrapper.find('.pm-modal').exists()).toBe(true)
    spy.mockRestore()
  })

  it('should close logs modal on X click', async () => {
    setupWithInstalled()
    mockApi.logs.mockResolvedValue([])
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Logs')!
    await btn.trigger('click')
    await flushPromises()
    await wrapper.find('.pm-modal-close').trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-modal').exists()).toBe(false)
  })
})

// ============================================================
// 9. TEST EXECUTION
// ============================================================
describe('Test execution', () => {
  it('should call test API and display result', async () => {
    setupWithInstalled()
    mockApi.test.mockResolvedValue({ message: 'Test completed', payload: '{"status":"ok"}' })
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Test')!
    await btn.trigger('click')
    await flushPromises()
    expect(mockApi.test).toHaveBeenCalledWith(1, 1)
    expect(wrapper.find('.pm-test-result').text()).toContain('Test completed')
  })

  it('should show error on test failure with response', async () => {
    setupWithInstalled()
    mockApi.test.mockRejectedValue({
      response: { data: { message: 'Connection refused' } },
    })
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Test')!
    await btn.trigger('click')
    await flushPromises()
    const text = wrapper.find('.pm-test-result').text()
    expect(text).toContain('Test failed')
    expect(text).toContain('Connection refused')
  })

  it('should show generic error on test failure without response', async () => {
    setupWithInstalled()
    mockApi.test.mockRejectedValue(new Error('Network Error'))
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Test')!
    await btn.trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-test-result').text()).toContain('Network Error')
  })

  it('should clear test result on Clear click', async () => {
    setupWithInstalled()
    mockApi.test.mockResolvedValue({ message: 'Test completed', payload: '{}' })
    const wrapper = mountComponent()
    await flushPromises()
    const btn = wrapper.findAll('.btn-outline').find(b => b.text().trim() === 'Test')!
    await btn.trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-test-result').exists()).toBe(true)
    await wrapper.find('.pm-test-result button').trigger('click')
    await flushPromises()
    expect(wrapper.find('.pm-test-result').exists()).toBe(false)
  })
})

// ============================================================
// 10. EMPTY STATE
// ============================================================
describe('Empty state', () => {
  it('should show empty message when no plugins installed', async () => {
    const wrapper = mountComponent()
    await flushPromises()
    const empty = wrapper.find('.pm-empty')
    expect(empty.exists()).toBe(true)
    expect(empty.text()).toContain('No plugins installed yet')
  })

  it('should NOT show installed section when empty', async () => {
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.find('.pm-installed-list').exists()).toBe(false)
  })

  it('should hide empty state when plugins are installed', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.find('.pm-empty').exists()).toBe(false)
    expect(wrapper.find('.pm-installed-list').exists()).toBe(true)
  })
})

// ============================================================
// 11. EVENT LABEL DISPLAY
// ============================================================
describe('Event label display', () => {
  it('should display known event labels', async () => {
    const wrapper = mountComponent()
    await flushPromises()
    const labels = wrapper.findAll('.pm-event-tag').map(t => t.text())
    expect(labels).toContain('Issue Created')
    expect(labels).toContain('Issue Updated')
  })

  it('should display raw event type for unknown events', async () => {
    mockApi.catalog.mockResolvedValue([
      { ...mockCatalog[0], subscribed_events: ['custom.event', 'issue.created'] },
    ])
    const wrapper = mountComponent()
    await flushPromises()
    const labels = wrapper.findAll('.pm-event-tag').map(t => t.text())
    expect(labels).toContain('custom.event')
    expect(labels).toContain('Issue Created')
  })
})

// ============================================================
// 12. HEADER AND SECTIONS
// ============================================================
describe('Header and sections', () => {
  it('should render page title and subtitle', async () => {
    const wrapper = mountComponent()
    await flushPromises()
    expect(wrapper.find('.pm-header h2').text()).toBe('Plugin Manager')
    expect(wrapper.find('.pm-subtitle').text()).toContain('Install and manage plugins')
  })

  it('should show installed count in section heading', async () => {
    setupWithInstalled()
    const wrapper = mountComponent()
    await flushPromises()
    const heading = wrapper.findAll('h3').find(h => h.text().includes('Installed Plugins'))!
    expect(heading.text()).toContain('(1)')
  })
})
