/**
 * useAI — AI Chat composable
 */
import { ref } from 'vue'
import { chatWithAI } from '@/api/ai'
import type { StreamEvent, AIChatMessage } from '@/types/ai'

export interface ParsedToolResult {
  toolName: string
  rows: Record<string, any>[]
  columns: string[]
}

export function useAI() {
  const messages = ref<AIChatMessage[]>([])
  const isStreaming = ref(false)
  const error = ref('')
  let controller: AbortController | null = null
  let currentToolName = ''

  function sendMessage(
    text: string,
    projectId: number,
    workspaceId: number,
    mode: 'ask' | 'build' | 'chart' = 'ask',
  ) {
    messages.value.push({ role: 'user', content: text })
    isStreaming.value = true
    error.value = ''

    // Placeholder for AI response
    const aiMsg: AIChatMessage = { role: 'assistant', content: '', toolCalls: [], toolResults: [] }
    const aiIndex = messages.value.length
    messages.value.push(aiMsg)

    controller = chatWithAI(
      projectId,
      workspaceId,
      { message: text, mode },
      (evt: StreamEvent) => {
        switch (evt.type) {
          case 'text':
            messages.value[aiIndex].content += evt.content || ''
            break
          case 'tool_call':
            messages.value[aiIndex].toolCalls = messages.value[aiIndex].toolCalls || []
            messages.value[aiIndex].toolCalls!.push(evt.tool_call!)
            currentToolName = evt.tool_call?.name || ''
            break
          case 'tool_result':
            const parsed = parseToolResult(currentToolName, evt.tool_result?.content || '')
            if (parsed) {
              messages.value[aiIndex].toolResults = messages.value[aiIndex].toolResults || []
              messages.value[aiIndex].toolResults!.push(parsed)
            }
            currentToolName = ''
            break
          case 'error':
            error.value = evt.error || 'AI error'
            isStreaming.value = false
            break
          case 'done':
            isStreaming.value = false
            break
        }
      },
      () => {
        isStreaming.value = false
      },
      (err: string) => {
        error.value = err
        isStreaming.value = false
      },
    )
  }

  function parseToolResult(toolName: string, raw: string): ParsedToolResult | null {
    if (!raw) return null
    try {
      const data = JSON.parse(raw)
      let rows: Record<string, any>[] = []
      let columns: string[] = []

      if (Array.isArray(data)) {
        if (data.length === 0) return null
        rows = data
        // Derive display columns from first item
        const first = data[0]
        const primaryCols = ['name', 'title', 'display_name', 'username']
        const metaCols = ['priority', 'state', 'status']
        const hasPrimary = primaryCols.some(k => first[k])
        // Show primary + meta, exclude raw color/group/id when name exists
        columns = [...primaryCols.filter(k => first[k]), ...metaCols.filter(k => first[k])]
        // Only include color/id/group if no primary column
        if (!hasPrimary) {
          columns.push(...['color', 'group', 'id'].filter(k => first[k]))
        }
        if (columns.length === 0) columns = Object.keys(first).slice(0, 5)
      } else if (typeof data === 'object' && data !== null) {
        // Single object → key-value rows
        rows = Object.entries(data)
          .filter(([_, v]) => typeof v !== 'object' || v === null)
          .map(([k, v]) => ({ key: k, value: v }))
        if (rows.length > 0) columns = ['key', 'value']
        else {
          // Nested: flatten first level
          rows = [data]
          columns = Object.keys(data).slice(0, 6)
        }
      } else {
        return null
      }

      return { toolName, rows: rows.slice(0, 20), columns }
    } catch {
      return null
    }
  }

  function cancel() {
    controller?.abort()
    isStreaming.value = false
  }

  function clear() {
    messages.value = []
    error.value = ''
  }

  return { messages, isStreaming, error, sendMessage, cancel, clear }
}
