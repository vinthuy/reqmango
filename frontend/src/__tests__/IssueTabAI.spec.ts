import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import IssueTabAI from '@/components/IssueTabAI.vue'

const { mockAnalyzeWithAI, mockSuggestLabels, mockAddIssueLabel, mockGetIssue } = vi.hoisted(() => ({
  mockAnalyzeWithAI: vi.fn(),
  mockSuggestLabels: vi.fn(),
  mockAddIssueLabel: vi.fn(),
  mockGetIssue: vi.fn(),
}))

vi.mock('@/api/ai', () => ({
  analyzeWithAI: (...args: any[]) => mockAnalyzeWithAI(...args),
  suggestLabels: (...args: any[]) => mockSuggestLabels(...args),
}))
vi.mock('@/api/issue', () => ({
  default: {
    addIssueLabel: (...args: any[]) => mockAddIssueLabel(...args),
    getIssue: (...args: any[]) => mockGetIssue(...args),
  },
}))
vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))
vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ success: vi.fn(), error: vi.fn() }),
}))

const baseProps = {
  issueId: 42,
  projectId: 1,
  issue: { id: 42, name: 'Login fails', description_html: '<p>500</p>' },
  labels: [{ id: 7, name: 'bug', color: '#f00' }],
}

describe('IssueTabAI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('summarize requests summary mode and shows the result', async () => {
    mockAnalyzeWithAI.mockResolvedValue({ mode: 'summary', summary: 'Healthy enough', insights: ['ok'] })
    const wrapper = mount(IssueTabAI, { props: baseProps })
    await wrapper.find('[data-test="ai-action-summarize"]').trigger('click')
    await flushPromises()
    expect(mockAnalyzeWithAI).toHaveBeenCalledWith(1, 42, 'summary')
    expect(wrapper.find('[data-test="ai-analyze-result"]').text()).toContain('Healthy enough')
    expect(wrapper.find('[data-test="ai-risks"]').exists()).toBe(false)
  })

  it('risk requests risk mode and lists graded risks', async () => {
    mockAnalyzeWithAI.mockResolvedValue({
      mode: 'risk',
      summary: 'Delivery at risk',
      insights: [],
      risks: [
        { level: 'high', title: 'Overdue 7 days', detail: 'Target date passed' },
        { level: 'low', title: 'Thin description', detail: 'Add repro steps' },
      ],
    })
    const wrapper = mount(IssueTabAI, { props: baseProps })
    await wrapper.find('[data-test="ai-action-risk"]').trigger('click')
    await flushPromises()
    expect(mockAnalyzeWithAI).toHaveBeenCalledWith(1, 42, 'risk')
    const risks = wrapper.findAll('[data-test="ai-risk-item"]')
    expect(risks).toHaveLength(2)
    expect(risks[0].text()).toContain('ai.riskLevel.high')
    expect(risks[0].text()).toContain('Overdue 7 days')
    expect(risks[0].text()).toContain('Target date passed')
  })

  it('risk with no findings shows the empty state', async () => {
    mockAnalyzeWithAI.mockResolvedValue({ mode: 'risk', summary: 'Looks fine', insights: [], risks: [] })
    const wrapper = mount(IssueTabAI, { props: baseProps })
    await wrapper.find('[data-test="ai-action-risk"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-test="ai-risks"]').text()).toContain('ai.noRisks')
  })

  it('suggest shows next steps even when label suggestions fail', async () => {
    mockAnalyzeWithAI.mockResolvedValue({ mode: 'next_steps', summary: 'Unblock first', insights: [], next_steps: ['Ask ops for logs', 'Add repro'] })
    mockSuggestLabels.mockRejectedValue(new Error('labels down'))
    const wrapper = mount(IssueTabAI, { props: baseProps })
    await wrapper.find('[data-test="ai-action-suggest"]').trigger('click')
    await flushPromises()
    expect(mockAnalyzeWithAI).toHaveBeenCalledWith(1, 42, 'next_steps')
    const steps = wrapper.findAll('[data-test="ai-next-step"]')
    expect(steps.map((s) => s.text())).toEqual(['Ask ops for logs', 'Add repro'])
    expect(wrapper.find('[data-test="ai-error"]').exists()).toBe(false)
  })

  it('suggest lists labels and applying one emits issue-updated', async () => {
    mockAnalyzeWithAI.mockResolvedValue({ mode: 'next_steps', summary: '', insights: [], next_steps: [] })
    mockSuggestLabels.mockResolvedValue({ suggested_labels: [{ label_id: 7, label_name: 'bug', reason: 'error' }] })
    mockAddIssueLabel.mockResolvedValue({})
    mockGetIssue.mockResolvedValue({ id: 42, name: 'Login fails', labels: [7] })
    const wrapper = mount(IssueTabAI, { props: baseProps })
    await wrapper.find('[data-test="ai-action-suggest"]').trigger('click')
    await flushPromises()
    expect(mockSuggestLabels).toHaveBeenCalledWith(1, 42, expect.objectContaining({ name: 'Login fails' }))
    expect(wrapper.find('[data-test="ai-label-suggestions"]').text()).toContain('bug')

    await wrapper.find('[data-test="ai-apply-label-7"]').trigger('click')
    await flushPromises()
    expect(mockAddIssueLabel).toHaveBeenCalledWith(42, 7)
    expect(wrapper.emitted('issue-updated')![0]).toEqual([{ id: 42, name: 'Login fails', labels: [7] }])
  })

  it('copilot action emits open-copilot without calling AI APIs', async () => {
    const wrapper = mount(IssueTabAI, { props: baseProps })
    await wrapper.find('[data-test="ai-action-copilot"]').trigger('click')
    await wrapper.find('[data-test="ai-open-copilot"]').trigger('click')
    expect(wrapper.emitted('open-copilot')).toHaveLength(2)
    expect(mockAnalyzeWithAI).not.toHaveBeenCalled()
    expect(mockSuggestLabels).not.toHaveBeenCalled()
  })

  it('suggest shows an error only when both requests fail', async () => {
    mockAnalyzeWithAI.mockRejectedValue({ response: { data: { message: 'llm down' } } })
    mockSuggestLabels.mockRejectedValue(new Error('labels down'))
    const wrapper = mount(IssueTabAI, { props: baseProps })
    await wrapper.find('[data-test="ai-action-suggest"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-test="ai-error"]').text()).toBe('llm down')
  })

  it('shows the API error message', async () => {
    mockAnalyzeWithAI.mockRejectedValue({ response: { data: { message: 'quota exceeded' } } })
    const wrapper = mount(IssueTabAI, { props: baseProps })
    await wrapper.find('[data-test="ai-action-risk"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-test="ai-error"]').text()).toBe('quota exceeded')
  })

  it('clears previous results when the issue changes', async () => {
    mockAnalyzeWithAI.mockResolvedValue({ summary: 'Old issue summary' })
    const wrapper = mount(IssueTabAI, { props: baseProps })
    await wrapper.find('[data-test="ai-action-summarize"]').trigger('click')
    await flushPromises()
    await wrapper.setProps({ issueId: 43, issue: { id: 43, name: 'Other' } })
    expect(wrapper.find('[data-test="ai-analyze-result"]').exists()).toBe(false)
  })
})
