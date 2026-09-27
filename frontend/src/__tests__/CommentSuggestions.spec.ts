import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import CommentSuggestions from '@/components/CommentSuggestions.vue'
import type { Comment, IssueSuggestion } from '@/types/comment'

const { mockApply, mockToastSuccess, mockToastError, mockToastWarning } = vi.hoisted(() => ({
  mockApply: vi.fn(),
  mockToastSuccess: vi.fn(),
  mockToastError: vi.fn(),
  mockToastWarning: vi.fn(),
}))

vi.mock('@/api/comment', () => ({
  default: { applySuggestions: (...args: any[]) => mockApply(...args) },
}))
vi.mock('@/composables/useToast', () => ({
  useToast: () => ({
    success: mockToastSuccess,
    error: mockToastError,
    warning: mockToastWarning,
  }),
}))
vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string, p?: Record<string, unknown>) => (p ? `${k}:${JSON.stringify(p)}` : k) }),
}))

function makeComment(suggestions: IssueSuggestion[], overrides: Partial<Comment> = {}): Comment {
  return {
    id: 11,
    issue_id: 42,
    author_id: null,
    agent_id: 3,
    body: '分诊建议',
    is_resolved: false,
    reaction_count: 0,
    created_at: '2026-09-27T10:00:00Z',
    updated_at: '2026-09-27T10:00:00Z',
    suggestions,
    ...overrides,
  } as Comment
}

const TITLE_AND_TYPE: IssueSuggestion[] = [
  { field: 'title', value: '登录页返回 500', label: '登录页返回 500', current: 'dispatch-1 登录页 500' },
  { field: 'type', value: 3, label: 'Bug (type_id=3)', current: 'Feature', reason: '服务端 5xx 应为 Bug' },
]

function mountCard(comment: Comment) {
  return mount(CommentSuggestions, { props: { issueId: 42, comment } })
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe('CommentSuggestions', () => {
  it('renders one actionable row per proposal', () => {
    const wrapper = mountCard(makeComment(TITLE_AND_TYPE))
    expect(wrapper.findAll('[data-test="suggestion-row"]')).toHaveLength(2)
    expect(wrapper.text()).toContain('Bug (type_id=3)')
    expect(wrapper.text()).toContain('服务端 5xx 应为 Bug')
  })

  it('renders nothing when the comment carries no proposals', () => {
    const wrapper = mountCard(makeComment([]))
    expect(wrapper.find('[data-test="comment-suggestions"]').exists()).toBe(false)
  })

  it('renders nothing once the proposals were already adopted', () => {
    // Otherwise a resolved thread would keep offering to re-apply old changes.
    const wrapper = mountCard(makeComment(TITLE_AND_TYPE, { suggestions_applied_at: '2026-09-27T11:00:00Z' }))
    expect(wrapper.find('[data-test="comment-suggestions"]').exists()).toBe(false)
  })

  it('applies a single field and hides only that row', async () => {
    mockApply.mockResolvedValue({ applied: ['title'] })
    const comment = makeComment(TITLE_AND_TYPE)
    const wrapper = mountCard(comment)

    await wrapper.findAll('[data-test="apply-suggestion"]')[0].trigger('click')
    await flushPromises()

    expect(mockApply).toHaveBeenCalledWith(42, 11, ['title'])
    const rows = wrapper.findAll('[data-test="suggestion-row"]')
    expect(rows).toHaveLength(1)
    expect(rows[0].text()).toContain('Bug (type_id=3)')
    expect(comment.suggestions).toHaveLength(1)
    expect(wrapper.emitted('applied')?.[0]).toEqual([['title']])
  })

  it('applies every field and collapses the block', async () => {
    mockApply.mockResolvedValue({ applied: ['title', 'type'] })
    const wrapper = mountCard(makeComment(TITLE_AND_TYPE))

    await wrapper.find('[data-test="apply-all-suggestions"]').trigger('click')
    await flushPromises()

    expect(mockApply).toHaveBeenCalledWith(42, 11, ['title', 'type'])
    expect(wrapper.find('[data-test="comment-suggestions"]').exists()).toBe(false)
  })

  it('marks a keep-as-is proposal as reasoning only, without an apply action', () => {
    // "保持 high" is signal for the reader; offering to apply it would imply a change.
    const wrapper = mountCard(makeComment([
      { field: 'title', value: '登录页返回 500', label: '登录页返回 500' },
      { field: 'priority', value: 'high', label: 'high', current: 'high', noop: true, reason: '影响面有限' },
    ]))

    const rows = wrapper.findAll('[data-test="suggestion-row"]')
    expect(rows).toHaveLength(2)
    expect(wrapper.findAll('[data-test="suggestion-noop"]')).toHaveLength(1)
    expect(wrapper.findAll('[data-test="apply-suggestion"]')).toHaveLength(1)
    expect(rows[1].text()).toContain('影响面有限')
  })

  it('excludes keep-as-is proposals from apply-all', async () => {
    mockApply.mockResolvedValue({ applied: ['title'] })
    const wrapper = mountCard(makeComment([
      { field: 'title', value: '登录页返回 500', label: '登录页返回 500' },
      { field: 'priority', value: 'high', label: 'high', current: 'high', noop: true },
    ]))

    await wrapper.find('[data-test="apply-all-suggestions"]').trigger('click')
    await flushPromises()

    expect(mockApply).toHaveBeenCalledWith(42, 11, ['title'])
  })

  it('offers no apply-all button when every proposal is a keep-as-is', () => {
    const wrapper = mountCard(makeComment([
      { field: 'priority', value: 'high', label: 'high', current: 'high', noop: true },
    ]))

    expect(wrapper.find('[data-test="apply-all-suggestions"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('comment.noChangesNeeded')
  })

  it('reports a refused proposal without discarding the ones that landed', async () => {
    // A workflow that forbids the transition must not make "apply all" look
    // broken: the other fields still apply and the refusal is explained.
    mockApply.mockResolvedValue({
      applied: ['title'],
      skipped: [{ field: 'state', reason: '状态转换「待处理 → 待办」不被当前工作流允许' }],
    })
    const comment = makeComment([
      { field: 'title', value: '登录页返回 500', label: '登录页返回 500' },
      { field: 'state', value: 2, label: '待办 (Todo)', current: '待处理 (Backlog)' },
    ])
    const wrapper = mountCard(comment)

    await wrapper.find('[data-test="apply-all-suggestions"]').trigger('click')
    await flushPromises()

    expect(mockToastSuccess).toHaveBeenCalled()
    expect(mockToastWarning).toHaveBeenCalledWith(
      expect.stringContaining('状态转换「待处理 → 待办」不被当前工作流允许')
    )
    // The refused row is gone too — retrying it verbatim would fail the same way.
    expect(comment.suggestions).toHaveLength(0)
    expect(wrapper.find('[data-test="comment-suggestions"]').exists()).toBe(false)
  })

  it('does not claim success when every proposal was refused', async () => {
    // The server answers 400 when nothing could be written; the UI must not
    // show a success toast alongside the error.
    mockApply.mockRejectedValue({
      response: { status: 400, data: { message: '状态转换不被允许' } },
    })
    const wrapper = mountCard(makeComment(TITLE_AND_TYPE))

    await wrapper.find('[data-test="apply-all-suggestions"]').trigger('click')
    await flushPromises()

    expect(mockToastSuccess).not.toHaveBeenCalled()
    expect(mockToastError).toHaveBeenCalledWith('状态转换不被允许')
  })

  it('treats an already-applied conflict as settled instead of an error', async () => {
    // A concurrent second click would 409; the UI must converge, not nag.
    mockApply.mockRejectedValue({ response: { status: 409 } })
    const wrapper = mountCard(makeComment(TITLE_AND_TYPE))

    await wrapper.find('[data-test="apply-all-suggestions"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-test="comment-suggestions"]').exists()).toBe(false)
    expect(mockToastError).not.toHaveBeenCalled()
  })

  it('surfaces a failure and keeps the rows actionable', async () => {
    mockApply.mockRejectedValue({ response: { status: 403, data: { message: 'no permission' } } })
    const wrapper = mountCard(makeComment(TITLE_AND_TYPE))

    await wrapper.find('[data-test="apply-all-suggestions"]').trigger('click')
    await flushPromises()

    expect(mockToastError).toHaveBeenCalledWith('no permission')
    expect(wrapper.findAll('[data-test="suggestion-row"]')).toHaveLength(2)
    expect(wrapper.find('[data-test="apply-all-suggestions"]').attributes('disabled')).toBeUndefined()
  })

  it('does not fire a second apply while one is in flight', async () => {
    let resolve: (v: unknown) => void = () => {}
    mockApply.mockReturnValue(new Promise((r) => { resolve = r }))
    const wrapper = mountCard(makeComment(TITLE_AND_TYPE))

    await wrapper.findAll('[data-test="apply-suggestion"]')[0].trigger('click')
    await wrapper.find('[data-test="apply-all-suggestions"]').trigger('click')
    expect(mockApply).toHaveBeenCalledTimes(1)

    resolve({ applied: ['title'] })
    await flushPromises()
  })
})
