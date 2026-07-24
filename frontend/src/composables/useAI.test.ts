/**
 * useAI Composable 单元测试
 * 覆盖：sendMessage（SSE 流处理），cancel，clear，状态管理
 *
 * 使用 vi.hoisted() 解决 mock 变量提升问题
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

describe('useAI - public API', () => {
  let ai: ReturnType<typeof useAI>

  beforeEach(() => {
    vi.clearAllMocks()
    ai = useAI()
    ai.clear()
  })

  describe('initial state', () => {
    it('should initialize with empty messages', () => {
      expect(ai.messages.value).toEqual([])
    })
    it('should initialize with isStreaming false', () => {
      expect(ai.isStreaming.value).toBe(false)
    })
    it('should initialize with empty error', () => {
      expect(ai.error.value).toBe('')
    })
  })

  describe('clear()', () => {
    it('should reset messages and error', () => {
      ai.messages.value.push({ role: 'user', content: 'hi' })
      ai.error.value = 'some error'
      ai.clear()
      expect(ai.messages.value).toEqual([])
      expect(ai.error.value).toBe('')
    })
  })

  describe('cancel()', () => {
    it('should stop streaming when cancel is called', () => {
      ai.cancel()
      expect(ai.isStreaming.value).toBe(false)
    })
  })

  describe('sendMessage - text event', () => {
    it('should add user message and stream text responses', () => {
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, onEvent: (e: StreamEvent) => void, onDone: () => void) => {
          onEvent({ type: 'text', content: 'Hello ' })
          onEvent({ type: 'text', content: 'World!' })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('Hi', 1, 1, 'ask')

      expect(ai.messages.value.length).toBe(2)
      expect(ai.messages.value[0]).toMatchObject({ role: 'user', content: 'Hi' })
      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg).toBeDefined()
      expect(assistantMsg!.content).toBe('Hello World!')
      expect(ai.isStreaming.value).toBe(false)
    })
  })

  describe('sendMessage - tool events', () => {
    it('should handle tool_call event', () => {
      const toolCall: ToolCall = { id: 't1', name: 'list_issues', input: {} }
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, onEvent: (e: StreamEvent) => void, onDone: () => void) => {
          onEvent({ type: 'tool_call', tool_call: toolCall })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('list issues', 1, 1, 'build')
      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg!.toolCalls).toHaveLength(1)
      expect(assistantMsg!.toolCalls![0].name).toBe('list_issues')
    })

    it('should handle tool_result with array data', () => {
      const items = [{ id: 1, name: 'Issue A', priority: 'high' }]
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, onEvent: (e: StreamEvent) => void, onDone: () => void) => {
          onEvent({ type: 'tool_call', tool_call: { id: 't2', name: 'search', input: {} } })
          onEvent({ type: 'tool_result', tool_result: { content: JSON.stringify(items) } })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('search issues', 1, 1, 'ask')
      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg!.toolResults).toHaveLength(1)
      expect(assistantMsg!.toolResults![0].toolName).toBe('search')
      expect(assistantMsg!.toolResults![0].columns).toContain('name')
    })

    it('should handle tool_result with object data', () => {
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, onEvent: (e: StreamEvent) => void, onDone: () => void) => {
          onEvent({ type: 'tool_call', tool_call: { id: 't3', name: 'get_config', input: {} } })
          onEvent({ type: 'tool_result', tool_result: { content: JSON.stringify({ api_url: 'https://api.example.com', timeout: 30 }) } })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('get config', 1, 1, 'ask')
      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg!.toolResults).toHaveLength(1)
      expect(assistantMsg!.toolResults![0].columns).toEqual(['key', 'value'])
    })

    it('should skip tool_result with empty content', () => {
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, onEvent: (e: StreamEvent) => void, onDone: () => void) => {
          onEvent({ type: 'tool_call', tool_call: { id: 't5', name: 'empty_tool', input: {} } })
          onEvent({ type: 'tool_result', tool_result: { content: '' } })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('call empty tool', 1, 1, 'build')
      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      // parseToolResult returns null for empty content → toolResults stays as []
      expect(assistantMsg!.toolResults).toEqual([])
    })
  })

  describe('sendMessage - error handling', () => {
    it('should handle stream error event', () => {
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, onEvent: (e: StreamEvent) => void) => {
          onEvent({ type: 'error', error: 'AI service unavailable' })
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('test', 1, 1, 'ask')
      expect(ai.error.value).toBe('AI service unavailable')
      expect(ai.isStreaming.value).toBe(false)
    })

    it('should handle error via onError callback', () => {
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, _onEvent: (e: StreamEvent) => void, _onDone: () => void, onError: (err: string) => void) => {
          onError('Internal server error')
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('test', 1, 1, 'ask')
      expect(ai.error.value).toBe('Internal server error')
      expect(ai.isStreaming.value).toBe(false)
    })
  })

  describe('sendMessage - parseToolResult edge cases', () => {
    it('should handle tool_result with invalid JSON', () => {
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, onEvent: (e: StreamEvent) => void, onDone: () => void) => {
          onEvent({ type: 'tool_call', tool_call: { id: 't6', name: 'bad_tool', input: {} } })
          onEvent({ type: 'tool_result', tool_result: { content: 'not-json' } })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('call bad tool', 1, 1, 'build')
      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      // parseToolResult tries JSON.parse and catches → returns null → stays empty
      expect(assistantMsg!.toolResults).toEqual([])
    })

    it('should handle tool_result with null JSON', () => {
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, onEvent: (e: StreamEvent) => void, onDone: () => void) => {
          onEvent({ type: 'tool_call', tool_call: { id: 't7', name: 'null_tool', input: {} } })
          onEvent({ type: 'tool_result', tool_result: { content: 'null' } })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('call null tool', 1, 1, 'ask')
      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      // JSON.parse('null') = null → parseToolResult returns null → stays empty
      expect(assistantMsg!.toolResults).toEqual([])
    })

    it('should handle tool_result with large array', () => {
      const items = Array.from({ length: 30 }, (_, i) => ({ id: i, name: `Item ${i}` }))
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, onEvent: (e: StreamEvent) => void, onDone: () => void) => {
          onEvent({ type: 'tool_call', tool_call: { id: 't8', name: 'list_all', input: {} } })
          onEvent({ type: 'tool_result', tool_result: { content: JSON.stringify(items) } })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('list all', 1, 1, 'ask')
      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg!.toolResults).toHaveLength(1)
      expect(assistantMsg!.toolResults![0]?.rows?.length ?? 0).toBeLessThanOrEqual(20)
    })

    it('should handle tool_result with empty array', () => {
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, onEvent: (e: StreamEvent) => void, onDone: () => void) => {
          onEvent({ type: 'tool_call', tool_call: { id: 't9', name: 'empty_result', input: {} } })
          onEvent({ type: 'tool_result', tool_result: { content: '[]' } })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('call empty result tool', 1, 1, 'build')
      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      // Array.isArray([]) → data.length === 0 → returns null → stays empty
      expect(assistantMsg!.toolResults).toEqual([])
    })

    it('should handle tool_result without preceding tool_call', () => {
      mockChatWithAI.mockImplementation(
        (_pid: number, _wid: number, _req: any, onEvent: (e: StreamEvent) => void, onDone: () => void) => {
          onEvent({ type: 'tool_result', tool_result: { content: JSON.stringify([{ id: 1 }]) } })
          onDone()
          return { abort: vi.fn() }
        }
      )

      ai.sendMessage('result without call', 1, 1, 'ask')
      const assistantMsg = ai.messages.value.find(m => m.role === 'assistant')
      // Tool result is still parsed even without preceding tool_call (toolName will be '')
      expect(assistantMsg!.toolResults).toBeDefined()
      expect(assistantMsg!.toolResults!.length).toBe(1)
      expect(assistantMsg!.toolResults![0].toolName).toBe('')
    })
  })
})
