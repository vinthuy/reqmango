import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import CommentList from '@/components/CommentList.vue'

const { mockList, mockCreate, mockAgentList, mockGet } = vi.hoisted(() => ({
  mockList: vi.fn(),
  mockCreate: vi.fn(),
  mockAgentList: vi.fn(),
  mockGet: vi.fn(),
}))

vi.mock('@/api/comment', () => ({
  default: {
    listIssueComments: (...args: any[]) => mockList(...args),
    createComment: (...args: any[]) => mockCreate(...args),
  },
}))
vi.mock('@/api', () => ({ default: { get: (...args: any[]) => mockGet(...args) } }))
vi.mock('@/api/agent', () => ({ agentApi: { list: (...args: any[]) => mockAgentList(...args) } }))
vi.mock('@/composables/useConfirm', () => ({ useConfirm: () => ({ confirm: vi.fn() }) }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ user: { id: 1, display_name: 'Leo' } }) }))
vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string, p?: Record<string, string>) => (p ? `${k}:${JSON.stringify(p)}` : k) }),
}))

// The component mutates comments it receives (it attaches replies), so every test
// needs its own fixtures rather than shared module-level objects.
function makeComment(overrides: Record<string, unknown> = {}) {
  return {
    id: 10,
    issue_id: 42,
    author_id: 1,
    author: { id: 1, username: 'leo', display_name: 'Leo', email: '' },
    body: '@Assistant Agent 请分诊',
    is_resolved: false,
    reaction_count: 0,
    created_at: '2026-09-27T10:00:00Z',
    updated_at: '2026-09-27T10:00:00Z',
    ...overrides,
  }
}

function makeAgentReply(overrides: Record<string, unknown> = {}) {
  return {
    id: 11,
    issue_id: 42,
    author_id: null,
    agent_id: 3,
    agent: { id: 3, name: 'Assistant Agent', avatar: '🦾' },
    parent_id: 10,
    body: '建议优先级调整为高',
    is_resolved: false,
    reaction_count: 0,
    created_at: '2026-09-27T10:00:05Z',
    updated_at: '2026-09-27T10:00:05Z',
    ...overrides,
  }
}

function mountList() {
  return mount(CommentList, {
    props: { issueId: 42, projectId: 1, workspaceId: 1 },
    global: { stubs: { teleport: true } },
  })
}

describe('CommentList agent replies', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGet.mockResolvedValue({ data: [] })
    mockAgentList.mockResolvedValue([{ id: 3, name: 'Assistant Agent', status: 'active' }])
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders agent-authored replies with the agent name, avatar and badge', async () => {
    mockList.mockResolvedValue({ comments: [makeComment(), makeAgentReply()], total: 2 })
    const wrapper = mountList()
    await flushPromises()

    const authors = wrapper.findAll('[data-test="comment-author"]').map(a => a.text())
    expect(authors).toEqual(['Leo', 'Assistant Agent'])
    expect(wrapper.findAll('[data-test="comment-agent-badge"]')).toHaveLength(1)
    expect(wrapper.text()).toContain('🦾')
    const mention = wrapper.findAll('span.bg-indigo-50').map(s => s.text())
    expect(mention).toContain('@Assistant Agent')
  })

  it('renders agent markdown replies (headings, lists, tables)', async () => {
    const reply = makeAgentReply({
      body: '## 分诊建议\n\n- 改为 Bug\n- 指派李四\n\n| 项 | 建议 |\n|---|---|\n| 类型 | Bug |',
    })
    mockList.mockResolvedValue({ comments: [makeComment({ body: 'hi' }), reply], total: 2 })
    const wrapper = mountList()
    await flushPromises()

    const body = wrapper.findAll('[data-test="comment-body"]')[1]
    expect(body.find('h3').exists()).toBe(true)
    expect(body.find('ul li').text()).toBe('改为 Bug')
    expect(body.find('table').exists()).toBe(true)
    expect(body.text()).toContain('类型')
  })

  it('shows a pending indicator after mentioning an agent and polls until it replies', async () => {
    vi.useFakeTimers()
    mockList.mockResolvedValueOnce({ comments: [], total: 0 })
    mockCreate.mockResolvedValue(makeComment())
    const wrapper = mountList()
    await flushPromises()

    await wrapper.find('textarea').setValue('@Assistant Agent 请分诊')
    await wrapper.find('button.bg-indigo-600').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-test="agent-pending"]').text()).toContain('Assistant Agent')

    mockList.mockResolvedValueOnce({ comments: [makeComment()], total: 1 })
    await vi.advanceTimersByTimeAsync(3000)
    await flushPromises()
    expect(wrapper.find('[data-test="agent-pending"]').exists()).toBe(true)

    mockList.mockResolvedValueOnce({ comments: [makeComment(), makeAgentReply()], total: 2 })
    await vi.advanceTimersByTimeAsync(3000)
    await flushPromises()
    expect(wrapper.find('[data-test="agent-pending"]').exists()).toBe(false)
    expect(wrapper.findAll('[data-test="comment-author"]').map(a => a.text())).toContain('Assistant Agent')
  })

  it('silently reloads when refreshKey changes', async () => {
    mockList.mockResolvedValueOnce({ comments: [], total: 0 })
    const wrapper = mountList()
    await flushPromises()

    mockList.mockResolvedValueOnce({ comments: [makeAgentReply({ parent_id: undefined })], total: 1 })
    await wrapper.setProps({ refreshKey: 1 })
    await flushPromises()
    expect(mockList).toHaveBeenCalledTimes(2)
    expect(wrapper.findAll('[data-test="comment-author"]').map(a => a.text())).toEqual(['Assistant Agent'])
  })

  it('watches for a dispatched agent reply and stops once it lands', async () => {
    vi.useFakeTimers()
    mockList.mockResolvedValueOnce({ comments: [], total: 0 })
    const wrapper = mountList()
    await flushPromises()

    mockList.mockResolvedValueOnce({ comments: [], total: 0 })
    await wrapper.setProps({ refreshKey: 1 })
    await flushPromises()
    expect(wrapper.find('[data-test="agent-pending"]').text()).toBe('comment.agentDispatched')

    mockList.mockResolvedValueOnce({ comments: [makeComment(), makeAgentReply()], total: 2 })
    await vi.advanceTimersByTimeAsync(3000)
    await flushPromises()
    expect(wrapper.find('[data-test="agent-pending"]').exists()).toBe(false)
    expect(wrapper.findAll('[data-test="comment-author"]').map(a => a.text())).toContain('Assistant Agent')
  })

  it('does not poll when no agent is mentioned', async () => {
    vi.useFakeTimers()
    mockList.mockResolvedValue({ comments: [], total: 0 })
    mockCreate.mockResolvedValue(makeComment({ body: 'plain' }))
    const wrapper = mountList()
    await flushPromises()

    await wrapper.find('textarea').setValue('plain comment')
    await wrapper.find('button.bg-indigo-600').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-test="agent-pending"]').exists()).toBe(false)
    await vi.advanceTimersByTimeAsync(6000)
    expect(mockList).toHaveBeenCalledTimes(1)
  })
})
