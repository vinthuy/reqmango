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
  workspaceId: number,
  request: AICreateRequest,
): Promise<AICreateResponse> {
  const response = await api.post(`/projects/${projectId}/ai/create`, request, {
    params: { workspace_id: workspaceId },
  })
  return response.data
}

export default { chatWithAI, searchWithAI, createPreviewWithAI, generateChart }

// ==================== AI 扩展端点 ====================

/**
 * AI Analyze — 分析工作项并生成洞察。
 */
export async function analyzeWithAI(projectId: number, issueId: number): Promise<any> {
  const res = await api.post(`/projects/${projectId}/ai/analyze?issue_id=${issueId}`)
  return res.data
}

/**
 * AI Suggest Labels — 根据内容推荐标签。
 */
export async function suggestLabels(projectId: number, issueId: number): Promise<any> {
  const res = await api.post(`/projects/${projectId}/ai/suggest-labels?issue_id=${issueId}`)
  return res.data
}

/**
 * AI Sprint Plan — 辅助冲刺计划。
 */
export async function sprintPlan(projectId: number): Promise<any> {
  const res = await api.post(`/projects/${projectId}/ai/sprint-plan`)
  return res.data
}

/**
 * AI Chart — 自然语言生成结构化图表配置。
 */
export interface AIChartData {
  chart_type: 'bar' | 'pie' | 'doughnut' | 'line' | 'polarArea' | 'radar'
  title: string
  labels: string[]
  datasets: {
    label: string
    data: number[]
    backgroundColor?: string[]
    borderColor?: string[]
    fill?: boolean
    tension?: number
  }[]
  options?: {
    indexAxis?: string
    stacked?: boolean
    showLegend?: boolean
  }
}

export async function generateChart(
  projectId: number,
  workspaceId: number,
  query: string,
): Promise<AIChartData> {
  const response = await api.post(`/projects/${projectId}/ai/chart`, { query }, {
    params: { workspace_id: workspaceId },
  })
  return response.data.data
}
