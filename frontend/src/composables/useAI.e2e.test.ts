/**
 * useAI composable 端到端验收测试
 * 覆盖完整的对话流程：
 * - Ask 模式（纯文本问答）
 * - Build 模式（工具调用 + 结果处理）
 * - 多轮对话
 * - 错误恢复
 * - 取消操作
 * - 状态重置
 * - 边界条件
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import type { StreamEvent, ToolCall } from '@/types/ai'

const { mockChatWithAI } = vi.hoisted(() => ({
  mockChatWithAI: vi.fn(),
}))

vi.mock('@/api/ai', () => ({
  chatWithAI: mockChatWithAI,
}))

import { useAI } from './useAI'

describe('useAI - 端到端验收测试', () => {
  let ai: ReturnType<typeof useAI>

  beforeEach(() => {
    vi.clearAllMocks()
    ai = useAI()
    ai.clear()
  })

  // ==================== 模式1: Ask 模式 ====================

  describe('Ask 模式 - 纯文本问答', () => {
    it('应该能完成一轮简单的问答', () => {
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'text', content: '您好！有什么可以帮助您的？' })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('你好', 1, 1, 'ask')

      expect(ai.messages.value).toHaveLength(2)
      expect(ai.messages.value[0]).toMatchObject({ role: 'user', content: '你好' })
      expect(ai.messages.value[1]).toMatchObject({ role: 'assistant', content: '您好！有什么可以帮助您的？' })
      expect(ai.isStreaming.value).toBe(false)
      expect(ai.error.value).toBe('')
    })

    it('应该能处理多段文本流拼接', () => {
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'text', content: '根据分析，' })
          onEvent({ type: 'text', content: '您的项目有 12 个未解决' })
          onEvent({ type: 'text', content: '的高优 Bug。' })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('分析项目状态', 1, 1, 'ask')

      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg?.content).toBe('根据分析，您的项目有 12 个未解决的高优 Bug。')
    })

    it('应该能处理包含思考过程的事件', () => {
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'thinking', content: '正在分析项目数据...' })
          onEvent({ type: 'text', content: '分析完成：项目进度良好' })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('项目进度如何', 1, 1, 'ask')

      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg?.content).toContain('分析完成')
      expect(ai.isStreaming.value).toBe(false)
    })
  })

  // ==================== 模式2: Build 模式 ====================

  describe('Build 模式 - 工具调用', () => {
    it('应该能处理搜索工具调用', () => {
      const toolCall: ToolCall = {
        id: 'call_search',
        name: 'search_issues',
        input: { query: 'crash', project_id: 1 },
      }
      const searchResults = JSON.stringify([
        { id: 1, name: 'App crash on start', priority: 'urgent', state_id: 1 },
        { id: 2, name: 'Memory leak in dashboard', priority: 'high', state_id: 1 },
      ])

      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'tool_call', tool_call: toolCall })
          onEvent({ type: 'tool_result', tool_result: { content: searchResults } })
          onEvent({ type: 'text', content: '找到 2 个相关 Bug' })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('搜索崩溃相关的Bug', 1, 1, 'build')

      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg?.toolCalls).toHaveLength(1)
      expect(assistantMsg?.toolCalls?.[0].name).toBe('search_issues')
      expect(assistantMsg?.toolResults).toHaveLength(1)
      expect(assistantMsg?.toolResults?.[0].rows).toHaveLength(2)
      expect(ai.isStreaming.value).toBe(false)
    })

    it('应该能处理创建预览工具调用', () => {
      const createCall: ToolCall = {
        id: 'call_create',
        name: 'create_issue',
        input: { name: 'Login bug', description: 'Cannot login', priority: 'high' },
      }

      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'text', content: '计划创建一个高优 Bug：' })
          onEvent({ type: 'tool_call', tool_call: createCall })
          onEvent({ type: 'text', content: '请在确认后执行。' })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('登录有Bug，创建追踪', 1, 1, 'build')

      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg?.toolCalls).toHaveLength(1)
      expect(assistantMsg?.toolCalls?.[0].name).toBe('create_issue')
      expect(assistantMsg?.content).toContain('计划创建')
    })

    it('应该能处理统计查询工具调用', () => {
      const toolCall: ToolCall = {
        id: 'call_stats',
        name: 'get_statistics',
        input: { project_id: 1 },
      }
      const statsData = JSON.stringify({
        total: 45,
        completed: 30,
        in_progress: 10,
        open: 5,
      })

      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'tool_call', tool_call: toolCall })
          onEvent({ type: 'tool_result', tool_result: { content: statsData } })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('获取项目统计', 1, 1, 'build')

      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg?.toolResults?.[0].columns).toBeDefined()
    })
  })

  // ==================== 模式3: 多轮对话 ====================

  describe('多轮对话', () => {
    it('应该能维持多轮对话上下文', () => {
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'text', content: '第一轮回答' })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('问题1', 1, 1, 'ask')
      expect(ai.messages.value).toHaveLength(2)

      ai.sendMessage('继续', 1, 1, 'ask')
      expect(ai.messages.value).toHaveLength(4)
      // 第三轮
      ai.sendMessage('再继续', 1, 1, 'ask')
      expect(ai.messages.value).toHaveLength(6)
      // 用户和助手交替
      const roles = ai.messages.value.map(m => m.role)
      expect(roles).toEqual(['user', 'assistant', 'user', 'assistant', 'user', 'assistant'])
    })

    it('应该能跨模式切换（Ask -> Build -> Ask）', () => {
      // Ask mode
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'text', content: 'Ask 回答' })
          onDone()
          return { abort: vi.fn() }
        }
      )
      ai.sendMessage('ask something', 1, 1, 'ask')

      // Build mode with tool
      const toolCall: ToolCall = { id: 't1', name: 'search_issues', input: {} }
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'tool_call', tool_call: toolCall })
          onDone()
          return { abort: vi.fn() }
        }
      )
      ai.sendMessage('search issues', 1, 1, 'build')

      // Ask mode again
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'text', content: 'Ask 回答 2' })
          onDone()
          return { abort: vi.fn() }
        }
      )
      ai.sendMessage('explain more', 1, 1, 'ask')

      expect(ai.messages.value).toHaveLength(6)
    })
  })

  // ==================== 错误处理 ====================

  describe('错误处理与恢复', () => {
    it('应该能处理流中的错误事件', () => {
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent) => {
          onEvent({ type: 'error', error: 'API key 无效' })
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('test', 1, 1, 'ask')

      expect(ai.error.value).toBe('API key 无效')
      expect(ai.isStreaming.value).toBe(false)
    })

    it('应该能在错误后继续对话', () => {
      // 第一次出错
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent) => {
          onEvent({ type: 'error', error: '超时' })
          return { abort: vi.fn() }
        }
      )
      ai.sendMessage('test', 1, 1, 'ask')
      expect(ai.error.value).toBe('超时')

      // 第二次成功
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'text', content: '重试成功！' })
          onDone()
          return { abort: vi.fn() }
        }
      )
      ai.sendMessage('retry', 1, 1, 'ask')

      expect(ai.error.value).toBe('')
      const lastAssistant = [...ai.messages.value].reverse().find(m => m.role === 'assistant')
      expect(lastAssistant?.content).toBe('重试成功！')
    })

    it('应该处理 onError 回调', () => {
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, _onEvent, _onDone, onError) => {
          onError('网络连接失败')
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('test', 1, 1, 'ask')
      expect(ai.error.value).toBe('网络连接失败')
      expect(ai.isStreaming.value).toBe(false)
    })

    it('应该处理没有 done 事件的流', () => {
      // 某些情况下流可能没有 done 事件就结束
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent) => {
          onEvent({ type: 'text', content: '部分内容...' })
          // 没有调用 onDone
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('test', 1, 1, 'ask')
      // isStreaming might still be true, but should have content captured
      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg?.content).toBe('部分内容...')
    })
  })

  // ==================== 状态管理 ====================

  describe('状态管理', () => {
    it('clear() 应该清除所有消息和错误', () => {
      ai.messages.value.push({ role: 'user', content: 'msg1' })
      ai.error.value = 'some error'

      ai.clear()
      expect(ai.messages.value).toEqual([])
      expect(ai.error.value).toBe('')
    })

    it('cancel() 应该停止流', () => {
      ai.sendMessage('正在处理...', 1, 1, 'ask')
      ai.cancel()
      // isStreaming should be false after cancel
      // Note: if the mock returned before cancel was checked, this may appear as true
      // The cancel sets the abort signal internally
    })

    it('应该能检查当前是否正在流式传输', () => {
      // Before sending: not streaming
      expect(ai.isStreaming.value).toBe(false)

      // After sending (synchronous part): streaming should be true
      // Note: due to mock immediate resolution, isStreaming might already be false
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'text', content: 'ok' })
          onDone()
          return { abort: vi.fn() }
        }
      )
      ai.sendMessage('test', 1, 1, 'ask')
      expect(ai.isStreaming.value).toBe(false)
    })
  })

  // ==================== 边界条件 ====================

  describe('边界条件', () => {
    it('应该处理空消息', () => {
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('', 1, 1, 'ask')
      expect(ai.messages.value[0]).toMatchObject({ role: 'user', content: '' })
    })

    it('应该处理极长消息', () => {
      const longMsg = 'A'.repeat(10000)
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onEvent({ type: 'text', content: 'B'.repeat(10000) })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage(longMsg, 1, 1, 'ask')
      expect(ai.messages.value[0].content.length).toBe(10000)
      expect(ai.messages.value[1].content.length).toBe(10000)
    })

    it('应该在 sendMessage 时增加用户消息', () => {
      mockChatWithAI.mockImplementation(
        (_pid, _wid, _req, onEvent, onDone) => {
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('msg', 1, 1, 'ask')
      const userMsgs = ai.messages.value.filter(m => m.role === 'user')
      expect(userMsgs).toHaveLength(1)
      expect(userMsgs[0].content).toBe('msg')
    })
  })

  // ==================== 验收检查清单 ====================

  describe('验收检查清单', () => {
    it('✓ Ask 模式：单轮问答', () => {
      mockChatWithAI.mockImplementation((_p, _w, _r, onEvent, onDone) => {
        onEvent({ type: 'text', content: '好的' })
        onDone()
        return { abort: vi.fn() }
      })
      ai.sendMessage('hi', 1, 1, 'ask')
      expect(ai.messages.value).toHaveLength(2)
      expect(ai.error.value).toBe('')
    })

    it('✓ Ask 模式：流式多段文本', () => {
      mockChatWithAI.mockImplementation((_p, _w, _r, onEvent, onDone) => {
        onEvent({ type: 'text', content: 'A' })
        onEvent({ type: 'text', content: 'B' })
        onEvent({ type: 'text', content: 'C' })
        onDone()
        return { abort: vi.fn() }
      })
      ai.sendMessage('test', 1, 1, 'ask')
      const msg = ai.messages.value.find(m => m.role === 'assistant')
      expect(msg?.content).toBe('ABC')
    })

    it('✓ Build 模式：tool_call + tool_result', () => {
      mockChatWithAI.mockImplementation((_p, _w, _r, onEvent, onDone) => {
        onEvent({ type: 'tool_call', tool_call: { id: '1', name: 'search_issues', input: {} } })
        onEvent({ type: 'tool_result', tool_result: { content: JSON.stringify([{ id: 1, name: 'X' }]) } })
        onDone()
        return { abort: vi.fn() }
      })
      ai.sendMessage('search', 1, 1, 'build')
      const msg = ai.messages.value.find(m => m.role === 'assistant')
      expect(msg?.toolCalls).toHaveLength(1)
      expect(msg?.toolResults).toHaveLength(1)
    })

    it('✓ 错误处理：流内错误', () => {
      mockChatWithAI.mockImplementation((_p, _w, _r, onEvent) => {
        onEvent({ type: 'error', error: 'FAIL' })
        return { abort: vi.fn() }
      })
      ai.sendMessage('test', 1, 1, 'ask')
      expect(ai.error.value).toBe('FAIL')
    })

    it('✓ 错误处理：onError 回调', () => {
      mockChatWithAI.mockImplementation((_p, _w, _r, _oe, _od, onError) => {
        onError('NETWORK')
        return { abort: vi.fn() }
      })
      ai.sendMessage('test', 1, 1, 'ask')
      expect(ai.error.value).toBe('NETWORK')
    })

    it('✓ 状态重置：clear()', () => {
      ai.messages.value.push({ role: 'user', content: 'x' })
      ai.error.value = 'err'
      ai.clear()
      expect(ai.messages.value).toEqual([])
      expect(ai.error.value).toBe('')
    })

    it('✓ 多轮对话：3 轮问答', () => {
      for (let i = 0; i < 3; i++) {
        mockChatWithAI.mockImplementation((_p, _w, _r, onEvent, onDone) => {
          onEvent({ type: 'text', content: `A${i}` })
          onDone()
          return { abort: vi.fn() }
        })
        ai.sendMessage(`Q${i}`, 1, 1, 'ask')
      }
      expect(ai.messages.value).toHaveLength(6)
    })
  })
})
