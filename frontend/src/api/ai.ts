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

  const params = new URLSearchParams({ workspace_id: String(workspaceId) })
  if (request.issue_id) params.set('issue_id', String(request.issue_id))
  if (request.page_id) params.set('page_id', String(request.page_id))

  fetch(`/api/v1/projects/${projectId}/ai/chat?${params}`, {
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
 * Backend requires name (and optional description) in the JSON body.
 */
export async function suggestLabels(
  projectId: number,
  issueId: number,
  payload?: { name: string; description?: string },
): Promise<any> {
  const body = payload ?? { name: `issue-${issueId}`, description: '' }
  const res = await api.post(`/projects/${projectId}/ai/suggest-labels?issue_id=${issueId}`, body)
  return res.data
}

/**
 * AI Sprint Plan — 辅助冲刺计划 / 周期一键总结。
 * Optional cycleId focuses the prompt on that cycle's issues and progress.
 */
export interface AISprintPlanResponse {
  recommended_capacity: number
  suggested_issues: number[]
  reasoning: string
  risks: string[]
}

export async function sprintPlan(projectId: number, cycleId?: number): Promise<AISprintPlanResponse> {
  const qs = cycleId && cycleId > 0 ? `?cycle_id=${cycleId}` : ''
  const res = await api.post(`/projects/${projectId}/ai/sprint-plan${qs}`)
  return res.data
}

/**
 * AI Chart — 自然语言生成结构化图表配置。
 */
export interface AIChartData {
  chart_type: 'bar' | 'pie' | 'doughnut' | 'line' | 'polarArea' | 'radar' | 'bubble' | 'scatter'
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
