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

  it('summarize calls analyzeWithAI and shows the result', async () => {
    mockAnalyzeWithAI.mockResolvedValue({ summary: 'Healthy enough', insights: ['ok'], bottlenecks: [] })
    const wrapper = mount(IssueTabAI, { props: baseProps })
    await wrapper.find('[data-test="ai-action-summarize"]').trigger('click')
    await flushPromises()
    expect(mockAnalyzeWithAI).toHaveBeenCalledWith(1, 42)
    expect(wrapper.find('[data-test="ai-analyze-result"]').text()).toContain('Healthy enough')
  })

  it('suggest lists labels and applying one emits issue-updated', async () => {
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
