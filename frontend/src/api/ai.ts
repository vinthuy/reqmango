/**
 * AI API — AI 智能助手 API 模块
 */
import api from './index'
import type { AISearchRequest, AISearchResponse, AICreateRequest, AICreateResponse, AIChatRequest, StreamEvent } from '@/types/ai'

/**
 * AI Chat — SSE 流式对话。
 * Returns an AbortController for cancellation.
 */
export function chatWithAI(
  projectId: number,
  workspaceId: number,
  request: AIChatRequest,
  onEvent: (evt: StreamEvent) => void,
  onDone: () => void,
  onError: (err: string) => void,
): AbortController {
  const controller = new AbortController()

  fetch(`/api/v1/projects/${projectId}/ai/chat?workspace_id=${workspaceId}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
    },
    body: JSON.stringify(request),
    signal: controller.signal,
  }).then(async (response) => {
    if (!response.ok) {
      const text = await response.text()
      onError(`HTTP ${response.status}: ${text}`)
      return
    }
    const reader = response.body?.getReader()
    if (!reader) { onError('No response body'); return }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        const trimmed = line.trim()
        if (!trimmed.startsWith('data: ')) continue
        const data = trimmed.slice(6)
        try {
          const evt: StreamEvent = JSON.parse(data)
          onEvent(evt)
          if (evt.type === 'done') onDone()
          if (evt.type === 'error') onError(evt.error || 'Unknown error')
        } catch { /* skip malformed */ }
      }
    }
  }).catch((err) => {
    if (err.name !== 'AbortError') onError(err.message)
  })

  return controller
}

/**
 * AI Search — NL 搜索工作项。
 */
export async function searchWithAI(
  projectId: number,
  request: AISearchRequest,
): Promise<AISearchResponse> {
  const response = await api.post(`/projects/${projectId}/ai/search`, request)
  return response.data
}

/**
 * AI Create Preview — NL 生成工作项预览。
 */
export async function createPreviewWithAI(
  projectId: number,
  request: AICreateRequest,
): Promise<AICreateResponse> {
  const response = await api.post(`/projects/${projectId}/ai/create`, request)
  return response.data
}

export default { chatWithAI, searchWithAI, createPreviewWithAI }
