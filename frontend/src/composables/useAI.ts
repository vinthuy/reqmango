/**
 * useAI — AI Chat composable
 */
import { ref } from 'vue'
import { chatWithAI } from '@/api/ai'
import type { StreamEvent, AIChatMessage } from '@/types/ai'

export function useAI() {
  const messages = ref<AIChatMessage[]>([])
  const isStreaming = ref(false)
  const error = ref('')
  let controller: AbortController | null = null

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
    const aiMsg: AIChatMessage = { role: 'assistant', content: '', toolCalls: [] }
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
            break
          case 'tool_result':
            messages.value[aiIndex].content += `\n\n> 🔧 查询结果: ${evt.tool_result?.content?.substring(0, 200) || ''}...`
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
