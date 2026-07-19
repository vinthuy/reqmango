/**
 * AISettingsPanel 组件测试
 * 覆盖：表单渲染、provider/model 列表、保存/测试连接逻辑
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

const mockGet = vi.fn()
const mockPut = vi.fn()
const mockPost = vi.fn()

vi.mock('@/api', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    put: (...args: any[]) => mockPut(...args),
    post: (...args: any[]) => mockPost(...args),
  },
}))

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
    locale: { value: 'zh-CN' },
  }),
}))

import AISettingsPanel from './AISettingsPanel.vue'

async function flushPromises() {
  await nextTick()
  await new Promise(r => setTimeout(r, 50))
  await nextTick()
}

describe('AISettingsPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGet.mockResolvedValue({
      data: {
        configured: false,
        provider: 'deepseek',
        model: 'deepseek-chat',
        max_tokens: 4096,
        is_active: true,
      },
    })
  })

  describe('initial render', () => {
    it('should render settings title', async () => {
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()
      expect(wrapper.text()).toContain('ai.settingsTitle')
    })

    it('should show not-configured badge when not configured', async () => {
      mockGet.mockResolvedValue({
        data: { configured: false },
      })
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()
      expect(wrapper.text()).toContain('ai.notConfigured')
    })

    it('should show configured badge when configured', async () => {
      mockGet.mockResolvedValue({
        data: { configured: true, provider: 'openai' },
      })
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()
      expect(wrapper.text()).toContain('ai.configured')
    })
  })

  describe('provider selection', () => {
    it('should have deepseek, anthropic, and openai options', async () => {
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()
      const select = wrapper.findAll('select')[0]
      if (select) {
        const options = select.findAll('option')
        const optionValues = options.map(o => o.attributes('value'))
        expect(optionValues).toContain('deepseek')
        expect(optionValues).toContain('anthropic')
        expect(optionValues).toContain('openai')
      }
    })
  })

  describe('model selection', () => {
    it('should have model options with optgroups', async () => {
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()
      const modelSelect = wrapper.findAll('select')[1]
      if (modelSelect) {
        const optgroups = modelSelect.findAll('optgroup')
        expect(optgroups.length).toBeGreaterThanOrEqual(3)
      }
    })
  })

  describe('save', () => {
    it('should call PUT with correct payload', async () => {
      mockPut.mockResolvedValue({ data: {} })
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()

      const saveBtn = wrapper.find('button.bg-indigo-600')
      await saveBtn.trigger('click')
      await flushPromises()

      expect(mockPut).toHaveBeenCalled()
      const [url, payload] = mockPut.mock.calls[0]
      expect(url).toContain('/workspaces/1/ai-config')
      expect(payload).toHaveProperty('provider')
      expect(payload).toHaveProperty('model')
      expect(payload).toHaveProperty('max_tokens')
      expect(payload).toHaveProperty('is_active')
    })

    it('should show save success message', async () => {
      mockPut.mockResolvedValue({ data: {} })
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()

      const saveBtn = wrapper.find('button.bg-indigo-600')
      await saveBtn.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('ai.saveSuccess')
    })

    it('should show error message on failure', async () => {
      mockPut.mockRejectedValue({
        response: { data: { message: 'Invalid API key' } },
      })
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()

      const saveBtn = wrapper.find('button.bg-indigo-600')
      await saveBtn.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('ai.saveFailed')
    })
  })

  describe('test connection', () => {
    it('should show success result on OK response', async () => {
      mockPost.mockResolvedValue({})
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()

      const buttons = wrapper.findAll('button')
      const testBtn = buttons.find(b => b.text() === 'ai.testConnection')
      if (testBtn) {
        await testBtn.trigger('click')
        await flushPromises()
        expect(wrapper.text()).toContain('ai.connectionSuccess')
      }
    })

    it('should show failure result on error', async () => {
      mockPost.mockRejectedValue({
        response: { data: { message: '401 Unauthorized' } },
      })
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()

      const buttons = wrapper.findAll('button')
      const testBtn = buttons.find(b => b.text() === 'ai.testConnection')
      if (testBtn) {
        await testBtn.trigger('click')
        await flushPromises()
        expect(wrapper.text()).toContain('ai.connectionFailed')
      }
    })
  })

  describe('loading state', () => {
    it('should show loading text initially', () => {
      mockGet.mockImplementation(() => new Promise(() => {})) // never resolves
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      expect(wrapper.text()).toContain('common.loading')
    })
  })

  describe('404 handling', () => {
    it('should still render form on 404', async () => {
      mockGet.mockRejectedValue({ response: { status: 404 } })
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()
      // Should still render form without error
      expect(wrapper.find('select').exists()).toBe(true)
    })
  })

  describe('form fields', () => {
    it('should render max_tokens as number input', async () => {
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()
      const numberInput = wrapper.find('input[type="number"]')
      expect(numberInput.exists()).toBe(true)
    })

    it('should render is_active checkbox', async () => {
      const wrapper = mount(AISettingsPanel, {
        props: { workspaceId: 1 },
      })
      await flushPromises()
      const checkbox = wrapper.find('input[type="checkbox"]')
      expect(checkbox.exists()).toBe(true)
    })
  })
})
