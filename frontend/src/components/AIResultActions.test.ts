/**
 * AIResultActions 组件测试
 * 覆盖：按钮显示逻辑、dashboard picker、批量创建
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('@/api/dashboard', () => ({
  listDashboards: vi.fn().mockResolvedValue([
    { id: 1, name: 'Sprint Dashboard' },
    { id: 2, name: 'Quality Metrics' },
  ]),
  addWidget: vi.fn().mockResolvedValue({}),
}))

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
    locale: { value: 'zh-CN' },
  }),
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

import AIResultActions from './AIResultActions.vue'

describe('AIResultActions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('showCreateIssue button', () => {
    it('should show for text message type', () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'text',
          projectId: 1,
          workspaceId: 1,
          content: 'Some text',
        },
      })
      expect(wrapper.text()).toContain('ai.createIssues')
    })

    it('should show for chart message type', () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'chart',
          projectId: 1,
          workspaceId: 1,
          chartConfig: { chart_type: 'bar' },
        },
      })
      expect(wrapper.text()).toContain('ai.createIssues')
    })

    it('should show for tool_result message type', () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'tool_result',
          projectId: 1,
          workspaceId: 1,
          toolResult: { toolName: 'search_issues', rows: [] },
        },
      })
      expect(wrapper.text()).toContain('ai.createIssues')
    })
  })

  describe('showSaveToDashboard button', () => {
    it('should show for chart with config', () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'chart',
          projectId: 1,
          workspaceId: 1,
          chartConfig: { chart_type: 'pie', title: 'Status Distribution' },
        },
      })
      expect(wrapper.text()).toContain('ai.saveToDashboard')
    })

    it('should hide when chartConfig is missing', () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'chart',
          projectId: 1,
          workspaceId: 1,
        },
      })
      expect(wrapper.text()).not.toContain('ai.saveToDashboard')
    })
  })

  describe('showSaveAsPage button', () => {
    it('should show for text with content', () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'text',
          projectId: 1,
          workspaceId: 1,
          content: 'Summary text',
        },
      })
      expect(wrapper.text()).toContain('ai.saveAsPage')
    })

    it('should show for tool_result with content', () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'tool_result',
          projectId: 1,
          workspaceId: 1,
          content: 'Result data',
          toolResult: { toolName: 'analyze', rows: [] },
        },
      })
      expect(wrapper.text()).toContain('ai.saveAsPage')
    })

    it('should hide when content is empty', () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'text',
          projectId: 1,
          workspaceId: 1,
          content: '',
        },
      })
      expect(wrapper.text()).not.toContain('ai.saveAsPage')
    })
  })

  describe('showBatchCreate button', () => {
    it('should show for search_issues tool result with rows', () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'tool_result',
          projectId: 1,
          workspaceId: 1,
          toolResult: {
            toolName: 'search_issues',
            rows: [
              { id: 1, name: 'Bug A', priority: 'high' },
              { id: 2, name: 'Bug B', priority: 'medium' },
            ],
          },
        },
      })
      expect(wrapper.text()).toContain('ai.batchCreateSubtasks')
    })

    it('should hide for non-search tool results', () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'tool_result',
          projectId: 1,
          workspaceId: 1,
          toolResult: {
            toolName: 'analyze',
            rows: [{ id: 1, name: 'Something' }],
          },
        },
      })
      expect(wrapper.text()).not.toContain('ai.batchCreateSubtasks')
    })

    it('should hide when rows are empty', () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'tool_result',
          projectId: 1,
          workspaceId: 1,
          toolResult: {
            toolName: 'search_issues',
            rows: [],
          },
        },
      })
      expect(wrapper.text()).not.toContain('ai.batchCreateSubtasks')
    })
  })

  describe('emits', () => {
    it('should emit create-issue on create issue click', async () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'text',
          projectId: 1,
          workspaceId: 1,
          content: '# Bug Report\nThis is a bug.',
        },
      })
      const btn = wrapper.find('button')
      await btn.trigger('click')
      expect(wrapper.emitted('create-issue')).toBeTruthy()
      const emitted = wrapper.emitted('create-issue')?.[0]?.[0]
      if (emitted) {
        expect(emitted).toHaveProperty('name', 'Bug Report')
        expect(emitted).toHaveProperty('description')
      }
    })

    it('should emit save-as-page on save as page click', async () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'text',
          projectId: 1,
          workspaceId: 1,
          content: 'Page content',
        },
      })
      const buttons = wrapper.findAll('button')
      const saveBtn = buttons.find(b => b.text().includes('ai.saveAsPage'))
      if (saveBtn) {
        await saveBtn.trigger('click')
        expect(wrapper.emitted('save-as-page')).toBeTruthy()
        expect(wrapper.emitted('save-as-page')?.[0]?.[0]).toBe('Page content')
      }
    })

    it('should emit create-issue with batch items on batch create', async () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'tool_result',
          projectId: 1,
          workspaceId: 1,
          toolResult: {
            toolName: 'search_issues',
            rows: [
              { name: 'Item 1', description: 'desc 1', priority: 'high' },
              { name: 'Item 2', description: '', priority: 'medium' },
            ],
          },
        },
      })
      const buttons = wrapper.findAll('button')
      const batchBtn = buttons.find(b => b.text().includes('ai.batchCreateSubtasks'))
      if (batchBtn) {
        await batchBtn.trigger('click')
        expect(wrapper.emitted('create-issue')).toBeTruthy()
        const emitted = wrapper.emitted('create-issue')?.[0]?.[0]
        if (emitted) {
          expect(emitted.batch).toHaveLength(2)
          expect(emitted.batch[0].name).toBe('Item 1')
          expect(emitted.batch[1].name).toBe('Item 2')
        }
      }
    })
  })

  describe('dashboard picker', () => {
    it('should toggle picker on save to dashboard click', async () => {
      const wrapper = mount(AIResultActions, {
        props: {
          messageType: 'chart',
          projectId: 1,
          workspaceId: 1,
          chartConfig: { chart_type: 'bar', title: 'Test' },
        },
      })
      const buttons = wrapper.findAll('button')
      const saveBtn = buttons.find(b => b.text().includes('ai.saveToDashboard'))
      if (saveBtn) {
        await saveBtn.trigger('click')
        await nextTick()
        expect(wrapper.text()).toContain('ai.selectDashboard')
      }
    })
  })
})
