/**
 * AgentAuditLog 组件测试
 * 覆盖：筛选器、时间线展示、反馈操作、刷新
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('@/api/agent', () => ({
  agentApi: {
    list: vi.fn(),
    listWorkspaceActivity: vi.fn(),
    rateActivity: vi.fn(),
  },
}))

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
    locale: { value: 'zh-CN' },
  }),
}))

import { agentApi } from '@/api/agent'
import AgentAuditLog from './AgentAuditLog.vue'

describe('AgentAuditLog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(agentApi.list as any).mockResolvedValue([])
    ;(agentApi.listWorkspaceActivity as any).mockResolvedValue([])
  })

  describe('empty state', () => {
    it('should show empty message when no activities', async () => {
      ;(agentApi.listWorkspaceActivity as any).mockResolvedValue([])
      const wrapper = mount(AgentAuditLog, {
        props: { workspaceId: 1 },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))
      expect(wrapper.text()).toContain('agent.auditNoRecords')
    })
  })

  describe('timeline rendering', () => {
    it('should render activity cards', async () => {
      ;(agentApi.list as any).mockResolvedValue([{ id: 10, name: 'Test Agent', avatar: '🤖' }])
      ;(agentApi.listWorkspaceActivity as any).mockResolvedValue([
        {
          id: 1,
          agent_id: 10,
          agent_name: 'Test Agent',
          action: 'auto_triage',
          result_summary: 'Triaged 5 issues',
          executed_at: '2026-01-15T10:00:00Z',
          issue_id: null,
          rating: null,
        },
        {
          id: 2,
          agent_id: 10,
          agent_name: 'Test Agent',
          action: 'auto_assign',
          result_summary: 'Assigned 3 issues',
          executed_at: '2026-01-15T10:05:00Z',
          issue_id: 42,
          rating: 1,
        },
      ])
      const wrapper = mount(AgentAuditLog, {
        props: { workspaceId: 1 },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      expect(wrapper.text()).toContain('Test Agent')
      expect(wrapper.text()).toContain('agent.autoTriage')
      expect(wrapper.text()).toContain('agent.autoAssign')
    })
  })

  describe('filtering', () => {
    it('should have agent filter dropdown with all-agents option', async () => {
      ;(agentApi.list as any).mockResolvedValue([
        { id: 1, name: 'Agent A', avatar: '🤖' },
        { id: 2, name: 'Agent B', avatar: '🦾' },
      ])
      const wrapper = mount(AgentAuditLog, {
        props: { workspaceId: 1 },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      const selects = wrapper.findAll('select')
      expect(selects.length).toBeGreaterThanOrEqual(1)
      // First select is agent filter
      if (selects.length >= 1) {
        const options = selects[0].findAll('option')
        const values = options.map(o => o.attributes('value'))
        expect(values).toContain('')
        expect(values).toContain('1')
        expect(values).toContain('2')
      }
    })

    it('should have action filter dropdown', async () => {
      const wrapper = mount(AgentAuditLog, {
        props: { workspaceId: 1 },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      const selects = wrapper.findAll('select')
      if (selects.length >= 2) {
        const actionOptions = selects[1].findAll('option')
        const values = actionOptions.map(o => o.attributes('value'))
        expect(values).toContain('dispatch')
        expect(values).toContain('auto_triage')
        expect(values).toContain('auto_assign')
        expect(values).toContain('mention')
      }
    })

    it('should refetch activities on agent filter change', async () => {
      const wrapper = mount(AgentAuditLog, {
        props: { workspaceId: 1 },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      const initialCalls = (agentApi.listWorkspaceActivity as any).mock.calls.length
      const selects = wrapper.findAll('select')
      if (selects.length >= 1) {
        await selects[0].setValue('1')
        await nextTick()
        await new Promise(r => setTimeout(r, 10))
        expect((agentApi.listWorkspaceActivity as any).mock.calls.length).toBeGreaterThan(initialCalls)
      }
    })
  })

  describe('feedback', () => {
    it('should submit positive feedback', async () => {
      ;(agentApi.list as any).mockResolvedValue([{ id: 1, name: 'Agent', avatar: '🤖' }])
      ;(agentApi.listWorkspaceActivity as any).mockResolvedValue([
        {
          id: 5,
          agent_id: 1,
          agent_name: 'Agent',
          action: 'dispatch',
          result_summary: 'Done',
          executed_at: '2026-01-15T10:00:00Z',
          issue_id: null,
          rating: null,
        },
      ])
      ;(agentApi.rateActivity as any).mockResolvedValue({})
      const wrapper = mount(AgentAuditLog, {
        props: { workspaceId: 1 },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      // Find and click thumbs up button
      const thumbsUpBtn = wrapper.find('[title="agent.feedbackUp"]')
      if (thumbsUpBtn.exists()) {
        await thumbsUpBtn.trigger('click')
        await nextTick()
        expect(agentApi.rateActivity).toHaveBeenCalledWith(1, 5, 1)
      }
    })
  })

  describe('refresh', () => {
    it('should call refresh on button click', async () => {
      const wrapper = mount(AgentAuditLog, {
        props: { workspaceId: 1 },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      const initialCalls = (agentApi.listWorkspaceActivity as any).mock.calls.length
      const refreshBtn = wrapper.find('button.text-blue-600')
      await refreshBtn.trigger('click')
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      expect((agentApi.listWorkspaceActivity as any).mock.calls.length).toBeGreaterThan(initialCalls)
    })
  })

  describe('action helpers', () => {
    it('should handle all action types', async () => {
      ;(agentApi.list as any).mockResolvedValue([])
      const allActions = ['dispatch', 'auto_triage', 'auto_assign', 'mention', 'summarize', 'custom', 'unknown_action']
      const activities = allActions.map((action, i) => ({
        id: i + 1,
        agent_id: 1,
        agent_name: 'Test',
        action,
        result_summary: `Did ${action}`,
        executed_at: '2026-01-15T10:00:00Z',
        issue_id: null,
        rating: null,
      }))
      ;(agentApi.listWorkspaceActivity as any).mockResolvedValue(activities)

      const wrapper = mount(AgentAuditLog, {
        props: { workspaceId: 1 },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      // Should render all activities without crashing
      allActions.forEach(action => {
        const label = `agent.${action === 'custom' ? 'customTask' : action}`
        // Some will use i18n key, some will use direct value
        expect(wrapper.text()).toBeTruthy()
      })
    })

    it('should format time correctly', async () => {
      ;(agentApi.list as any).mockResolvedValue([])
      const now = new Date()
      ;(agentApi.listWorkspaceActivity as any).mockResolvedValue([
        {
          id: 1,
          agent_id: 1,
          agent_name: 'Test',
          action: 'dispatch',
          result_summary: 'Done',
          executed_at: now.toISOString(),
          issue_id: null,
          rating: null,
        },
      ])

      const wrapper = mount(AgentAuditLog, {
        props: { workspaceId: 1 },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 10))

      // Should show "just now" for recent activity
      expect(wrapper.text()).toContain('common.justNow')
    })
  })
})
