/**
 * AI API 单元测试
 * 覆盖：所有 AI API 函数的参数验证和请求格式
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()
const mockPost = vi.fn()

vi.mock('./index', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
  },
}))

import {
  searchWithAI, createPreviewWithAI, analyzeWithAI,
  suggestLabels, sprintPlan, generateChart,
} from './ai'
import type { AIChartData } from './ai'

beforeEach(() => {
  vi.clearAllMocks()
})

describe('searchWithAI', () => {
  it('should POST to correct endpoint with query', async () => {
    mockPost.mockResolvedValue({
      data: { rql: 'test', explanation: 'found', issues: [] },
    })
    const result = await searchWithAI(1, { query: 'high priority bugs' })
    expect(mockPost).toHaveBeenCalledWith(
      '/projects/1/ai/search', { query: 'high priority bugs' }
    )
    expect(result.rql).toBe('test')
  })
})

describe('createPreviewWithAI', () => {
  it('should POST to create endpoint with description', async () => {
    mockPost.mockResolvedValue({
      data: { preview: { name: 'Bug' }, explanation: 'Created' },
    })
    const result = await createPreviewWithAI(1, 1, {
      description: 'A critical login bug',
      workspace_id: 1,
    })
    expect(mockPost).toHaveBeenCalledWith(
      '/projects/1/ai/create',
      { description: 'A critical login bug', workspace_id: 1 },
      { params: { workspace_id: 1 } }
    )
    expect(result.preview.name).toBe('Bug')
  })
})

describe('analyzeWithAI', () => {
  it('should POST to analyze endpoint with issue_id', async () => {
    mockPost.mockResolvedValue({
      data: { insights: ['Bottleneck at review'], summary: 'Good' },
    })
    const result = await analyzeWithAI(1, 42)
    expect(mockPost).toHaveBeenCalledWith(
      '/projects/1/ai/analyze?issue_id=42'
    )
    expect(result.insights).toHaveLength(1)
  })
})

describe('suggestLabels', () => {
  it('should POST to suggest-labels endpoint with issue_id', async () => {
    mockPost.mockResolvedValue({
      data: { labels: ['bug', 'critical'], confidence: 0.95 },
    })
    const result = await suggestLabels(1, 42)
    expect(mockPost).toHaveBeenCalledWith(
      '/projects/1/ai/suggest-labels?issue_id=42'
    )
    expect(result.labels).toContain('bug')
  })
})

describe('sprintPlan', () => {
  it('should POST to sprint-plan endpoint', async () => {
    mockPost.mockResolvedValue({
      data: { plan: [{ issue_id: 1, suggestion: 'Move to next sprint' }] },
    })
    const result = await sprintPlan(1)
    expect(mockPost).toHaveBeenCalledWith(
      '/projects/1/ai/sprint-plan'
    )
    expect(result.plan).toHaveLength(1)
  })
})

describe('generateChart', () => {
  it('should POST to chart endpoint with query', async () => {
    const chartData: AIChartData = {
      chart_type: 'bar',
      labels: ['Open', 'Closed'],
      datasets: [{ label: 'Issues', data: [10, 20] }],
      title: 'Status Report',
    }
    mockPost.mockResolvedValue({
      data: { data: chartData },
    })
    const result = await generateChart(1, 1, '按状态分布')
    expect(mockPost).toHaveBeenCalledWith(
      '/projects/1/ai/chart',
      { query: '按状态分布' },
      { params: { workspace_id: 1 } }
    )
    expect(result.chart_type).toBe('bar')
    expect(result.labels).toEqual(['Open', 'Closed'])
  })
})

describe('endpoint patterns', () => {
  it('should always use project_id in the URL path', async () => {
    mockPost.mockResolvedValue({ data: {} })
    await sprintPlan(99)
    const url = mockPost.mock.calls[0][0] as string
    expect(url).toContain('/projects/99/')
  })

  it('searchWithAI should not require workspace_id', async () => {
    mockPost.mockResolvedValue({ data: { rql: '', explanation: '', issues: [] } })
    await searchWithAI(1, { query: 'test' })
    const url = mockPost.mock.calls[0][0] as string
    expect(url).toBe('/projects/1/ai/search')
  })

  it('createPreviewWithAI should include workspace_id in params', async () => {
    mockPost.mockResolvedValue({ data: { preview: {}, explanation: '' } })
    await createPreviewWithAI(7, 3, { description: 'test' })
    const url = mockPost.mock.calls[0][0] as string
    const options = mockPost.mock.calls[0][2]
    expect(url).toBe('/projects/7/ai/create')
    expect(options.params.workspace_id).toBe(3)
  })
})
