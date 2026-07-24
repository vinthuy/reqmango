/**
 * AgentSelector 组件测试
 * 覆盖：成员/Agent 渲染、选择值绑定、emit
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('@/api/agent', () => ({
  agentApi: {
    list: vi.fn().mockResolvedValue([]),
  },
}))

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
    locale: { value: 'zh-CN' },
  }),
}))

import { agentApi } from '@/api/agent'
import AgentSelector from './AgentSelector.vue'

describe('AgentSelector', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(agentApi.list as any).mockResolvedValue([])
  })

  describe('rendering', () => {
    it('should render a select element', () => {
      const wrapper = mount(AgentSelector, {
        props: {
          modelValue: '',
          workspaceId: 1,
          members: [],
        },
      })
      expect(wrapper.find('select').exists()).toBe(true)
    })
  })

  describe('members', () => {
    it('should display members from prop', async () => {
      const members = [
        { id: 1, display_name: 'Alice' },
        { id: 2, display_name: 'Bob' },
      ]
      const wrapper = mount(AgentSelector, {
        props: {
          modelValue: '',
          workspaceId: 1,
          members,
        },
      })
      await nextTick()
      const options = wrapper.findAll('option')
      const texts = options.map(o => o.text())
      expect(texts.some(t => t.includes('Alice'))).toBe(true)
      expect(texts.some(t => t.includes('Bob'))).toBe(true)
    })
  })

  describe('agents', () => {
    it('should display AI agents fetched from API', async () => {
      ;(agentApi.list as any).mockResolvedValue([
        { id: 101, name: 'Triage Bot', avatar: '🤖', agent_type: 'builtin', status: 'active' },
        { id: 102, name: 'Code Reviewer', avatar: '🔍', agent_type: 'builtin', status: 'active' },
      ])
      const wrapper = mount(AgentSelector, {
        props: {
          modelValue: '',
          workspaceId: 1,
          members: [],
        },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      const options = wrapper.findAll('option')
      const texts = options.map(o => o.text())
      expect(texts.some(t => t.includes('Triage Bot'))).toBe(true)
      expect(texts.some(t => t.includes('Code Reviewer'))).toBe(true)
    })
  })

  describe('model value', () => {
    it('should select member by ID as string', () => {
      const wrapper = mount(AgentSelector, {
        props: {
          modelValue: '1',
          workspaceId: 1,
          members: [{ id: 1, display_name: 'Alice' }],
        },
      })
      const select = wrapper.find('select')
      expect(select.element.value).toBe('1')
    })

    it('should select agent by agent: prefix', async () => {
      ;(agentApi.list as any).mockResolvedValue([
        { id: 101, name: 'Bot', avatar: '🤖', agent_type: 'builtin', status: 'active' },
      ])
      const wrapper = mount(AgentSelector, {
        props: {
          modelValue: 'agent:101',
          workspaceId: 1,
          members: [],
        },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))
      const select = wrapper.find('select')
      expect(select.element.value).toBe('agent:101')
    })

    it('should have empty value when modelValue is empty string', () => {
      const wrapper = mount(AgentSelector, {
        props: {
          modelValue: '',
          workspaceId: 1,
          members: [],
        },
      })
      const select = wrapper.find('select')
      expect(select.element.value).toBe('')
    })
  })

  describe('emits', () => {
    it('should emit update:modelValue with member id string when member selected', async () => {
      const wrapper = mount(AgentSelector, {
        props: {
          modelValue: '',
          workspaceId: 1,
          members: [{ id: 1, display_name: 'Alice' }],
        },
      })
      const select = wrapper.find('select')
      await select.setValue('1')
      expect(wrapper.emitted('update:modelValue')).toBeTruthy()
      const emitted = wrapper.emitted('update:modelValue')?.[0]?.[0]
      expect(emitted).toBe('1')
    })

    it('should emit update:modelValue with agent: prefix when agent selected', async () => {
      ;(agentApi.list as any).mockResolvedValue([
        { id: 101, name: 'Bot', avatar: '🤖', agent_type: 'builtin', status: 'active' },
      ])
      const wrapper = mount(AgentSelector, {
        props: {
          modelValue: '',
          workspaceId: 1,
          members: [],
        },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      const select = wrapper.find('select')
      await select.setValue('agent:101')
      expect(wrapper.emitted('update:modelValue')).toBeTruthy()
      const emitted = wrapper.emitted('update:modelValue')?.[0]?.[0]
      expect(emitted).toBe('agent:101')
    })
  })

  describe('edge cases', () => {
    it('should handle empty members and no agents', () => {
      const wrapper = mount(AgentSelector, {
        props: {
          modelValue: '',
          workspaceId: 1,
          members: [],
        },
      })
      expect(wrapper.find('select').exists()).toBe(true)
    })

    it('should handle both members and agents', async () => {
      ;(agentApi.list as any).mockResolvedValue([
        { id: 101, name: 'Bot', avatar: '🤖', agent_type: 'builtin', status: 'active' },
      ])
      const wrapper = mount(AgentSelector, {
        props: {
          modelValue: '',
          workspaceId: 1,
          members: [{ id: 1, display_name: 'Alice' }],
        },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      const options = wrapper.findAll('option')
      // Should have: empty/unassigned + 1 member + 1 agent
      expect(options.length).toBeGreaterThanOrEqual(3)
    })
  })
})
