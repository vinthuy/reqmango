import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

// Hoisted mocks for API calls
const {
  mockGetIssue, mockUpdateIssue, mockListStates, mockListCycles, mockListModules,
  mockListProjectMembers, mockGetIssueTypes, mockDispatch, mockAnalyzeWithAI,
  mockSuggestLabels, mockAddIssueLabel,
} = vi.hoisted(() => ({
  mockGetIssue: vi.fn().mockResolvedValue({
    id: 42,
    sequence_id: 42,
    name: 'Test Issue',
    description_html: '<p>Hello</p>',
    priority: 'medium',
    state_id: 1,
    state_name: 'Todo',
    state_group: 'todo',
    project_id: 1,
    workspace_id: 1,
    issue_type: { id: 1, name: 'Task', color: '#6366f1' },
    assignees: [],
    labels: [],
    label_details: [],
    parent_id: null,
    parent: null,
    sub_issues: [],
    start_date: null,
    target_date: null,
    cycle_id: null,
    cycle: null,
    project: { id: 1, name: 'Test Project', identifier: 'DEV' },
  }),
  mockUpdateIssue: vi.fn().mockResolvedValue({ id: 42 }),
  mockListStates: vi.fn().mockResolvedValue([]),
  mockListCycles: vi.fn().mockResolvedValue({ items: [] }),
  mockListModules: vi.fn().mockResolvedValue([]),
  mockListProjectMembers: vi.fn().mockResolvedValue([]),
  mockGetIssueTypes: vi.fn().mockResolvedValue([]),
  mockDispatch: vi.fn().mockResolvedValue({}),
  mockAnalyzeWithAI: vi.fn().mockResolvedValue({
    summary: 'Healthy enough',
    insights: ['Insight A'],
    bottlenecks: [{ issue_id: 7, issue_name: 'Stuck', days_in_state: 5, state_name: 'Doing' }],
  }),
  mockSuggestLabels: vi.fn().mockResolvedValue({
    suggested_labels: [{ label_id: 3, label_name: 'bug', confidence: 0.9, reason: 'crash' }],
  }),
  mockAddIssueLabel: vi.fn().mockResolvedValue({ issue_id: 42, label_id: 3, action: 'added' }),
}))

// Mock all API modules
vi.mock('@/api/issue', () => ({
  getIssue: (...args: any[]) => mockGetIssue(...args),
  updateIssue: (...args: any[]) => mockUpdateIssue(...args),
  addIssueLabel: (...args: any[]) => mockAddIssueLabel(...args),
  listWatchers: vi.fn().mockResolvedValue({ watchers: [] }),
  addWatcher: vi.fn().mockResolvedValue({}),
  removeWatcher: vi.fn().mockResolvedValue({}),
  issueApi: {
    getIssue: (...args: any[]) => mockGetIssue(...args),
    updateIssue: (...args: any[]) => mockUpdateIssue(...args),
  },
}))

vi.mock('@/api/ai', () => ({
  analyzeWithAI: (...args: any[]) => mockAnalyzeWithAI(...args),
  suggestLabels: (...args: any[]) => mockSuggestLabels(...args),
}))

vi.mock('@/api/project-settings', () => ({
  listStates: (...args: any[]) => mockListStates(...args),
}))

vi.mock('@/api/cycle', () => ({
  listCycles: (...args: any[]) => mockListCycles(...args),
}))

vi.mock('@/api/module', () => ({
  listModules: (...args: any[]) => mockListModules(...args),
}))

vi.mock('@/api/project', () => ({
  default: {
    listProjectMembers: (...args: any[]) => mockListProjectMembers(...args),
  },
}))

vi.mock('@/api/issue-type', () => ({
  getIssueTypes: (...args: any[]) => mockGetIssueTypes(...args),
}))

vi.mock('@/api/agent', () => ({
  agentApi: {
    dispatch: (...args: any[]) => mockDispatch(...args),
  },
}))

vi.mock('@/components/AICopilot.vue', () => ({
  default: {
    template: '<div data-test="mock-aicopilot" v-if="visible">AICopilot</div>',
    props: ['visible', 'projectId', 'workspaceId', 'projectName', 'issueId', 'issueLabel', 'view'],
    emits: ['close'],
  },
}))

// Mock composables
vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ success: vi.fn(), error: vi.fn() }),
}))

vi.mock('@/composables/useConfirm', () => ({
  useConfirm: () => ({ confirm: vi.fn().mockResolvedValue(true) }),
}))

// Mock vue-router
const mockPush = vi.fn()
const mockBack = vi.fn()
vi.mock('vue-router', () => ({
  useRoute: () => ({
    params: { issueId: '42', id: '1', slug: 'test' },
  }),
  useRouter: () => ({
    push: mockPush,
    back: mockBack,
  }),
}))

// Mock child components — all stubs are simple divs
vi.mock('@/components/IssueDetailHeader.vue', () => ({
  default: {
    template:
      '<div data-test="mock-header"><span>{{ issue?.issue_type?.name }}</span><button data-test="save-btn" @click="$emit(\'save\')">{{ saving ? \'issue.saving\' : \'issue.save\' }}</button></div>',
    props: ['issue', 'saving'],
    emits: ['save', 'back'],
  },
}))

vi.mock('@/components/IssuePropertySidebar.vue', () => ({
  default: {
    template: '<div data-test="mock-sidebar">Property Sidebar</div>',
    props: [
      'issue', 'states', 'members', 'cycles', 'modules',
      'selectedAgentId', 'agentDispatching',
    ],
  },
}))

vi.mock('@/components/IssueTabDetails.vue', () => ({
  default: {
    template:
      '<div data-test="mock-details">Details Tab</div>',
    props: ['issueId', 'issue', 'workspaceId', 'projectId', 'issueTypeId', 'members'],
    emits: ['update:title', 'update:description'],
  },
}))

vi.mock('@/components/IssueTabRelations.vue', () => ({
  default: {
    template: '<div data-test="mock-relations">Relations Tab</div>',
    props: ['issueId', 'projectId', 'workspaceId', 'parent', 'subIssues', 'issueTypes'],
    emits: ['navigate'],
  },
}))

vi.mock('@/components/IssueTabAttachments.vue', () => ({
  default: {
    template: '<div data-test="mock-attachments">Attachments Tab</div>',
    props: ['issueId', 'projectId'],
  },
}))

vi.mock('@/components/IssueTabTimeTracking.vue', () => ({
  default: {
    template: '<div data-test="mock-timetrack">Time Tracking Tab</div>',
    props: ['issueId'],
  },
}))

vi.mock('@/components/IssueTabActivity.vue', () => ({
  default: {
    template: '<div data-test="mock-activity">Activity Tab</div>',
    props: ['issueId'],
  },
}))

vi.mock('@/components/IssueGitPanel.vue', () => ({
  default: {
    template: '<div data-test="mock-git">Git Panel</div>',
    props: ['workspaceId', 'issueId'],
  },
}))

vi.mock('@/components/AgentSelector.vue', () => ({
  default: {
    template: '<select data-test="mock-agent-selector"><option>Agent</option></select>',
    props: ['modelValue', 'workspaceId'],
  },
}))

import IssueDetail from '@/views/IssueDetail.vue'

describe('IssueDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  function mountComponent() {
    return mount(IssueDetail, {
      attachTo: document.body,
    })
  }

  it('renders the header with issue type and save button', async () => {
    const wrapper = mountComponent()
    await nextTick() // wait for onMounted
    await nextTick()

    expect(mockGetIssue).toHaveBeenCalledWith(42)
    expect(wrapper.find('[data-test="mock-header"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Task')
    expect(wrapper.text()).toContain('issue.save')
  })

  it('renders 8 tab buttons', async () => {
    const wrapper = mountComponent()
    await nextTick()
    await nextTick()

    const tabBtns = wrapper.findAll('[data-test="tab-btn"]')
    expect(tabBtns.length).toBe(8)
    expect(tabBtns[0].text()).toBe('issue.tabDetails')
    expect(tabBtns[1].text()).toBe('issue.tabRelations')
    expect(tabBtns[2].text()).toBe('issue.tabAttachments')
    expect(tabBtns[3].text()).toBe('gitIntegration.title')
    expect(tabBtns[4].text()).toBe('issue.tabTimetrack')
    expect(tabBtns[5].text()).toBe('issue.tabActivity')
    expect(tabBtns[6].text()).toBe('🤖 AI')
    expect(tabBtns[7].text()).toBe('issue.tabChat')
  })

  it('shows Details tab by default', async () => {
    const wrapper = mountComponent()
    await nextTick()
    await nextTick()

    expect(wrapper.find('[data-test="mock-details"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="mock-relations"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="mock-attachments"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="mock-timetrack"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="mock-activity"]').exists()).toBe(false)
  })

  it('switches tabs on click', async () => {
    const wrapper = mountComponent()
    await nextTick()
    await nextTick()

    const tabBtns = wrapper.findAll('[data-test="tab-btn"]')
    
    await tabBtns[1].trigger('click')
    await nextTick()

    expect(wrapper.find('[data-test="mock-details"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="mock-relations"]').exists()).toBe(true)

    await tabBtns[5].trigger('click')
    await nextTick()

    expect(wrapper.find('[data-test="mock-relations"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="mock-activity"]').exists()).toBe(true)
  })

  it('calls save API when save is triggered', async () => {
    const wrapper = mountComponent()
    await nextTick()
    await nextTick()

    // Click save button inside the header
    const saveBtn = wrapper.find('[data-test="save-btn"]')
    await saveBtn.trigger('click')
    await nextTick()

    // saveIssue should call updateIssue
    expect(mockUpdateIssue).toHaveBeenCalledWith(42, expect.objectContaining({
      name: 'Test Issue',
    }))
  })

  async function openAiTab(wrapper: ReturnType<typeof mount>) {
    await nextTick()
    await nextTick()
    const tabBtns = wrapper.findAll('[data-test="tab-btn"]')
    await tabBtns[6].trigger('click')
    await nextTick()
  }

  it('AI summarize calls analyzeWithAI and shows result panel', async () => {
    const wrapper = mountComponent()
    await openAiTab(wrapper)

    expect(wrapper.find('[data-test="ai-tab"]').exists()).toBe(true)
    await wrapper.find('[data-test="ai-action-summarize"]').trigger('click')
    await nextTick()
    await nextTick()

    expect(mockAnalyzeWithAI).toHaveBeenCalledWith(1, 42)
    expect(wrapper.find('[data-test="ai-analyze-result"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Healthy enough')
    expect(wrapper.find('[data-test="mock-aicopilot"]').exists()).toBe(false)
  })

  it('AI suggest calls suggestLabels and lists suggestions', async () => {
    const wrapper = mountComponent()
    await openAiTab(wrapper)

    await wrapper.find('[data-test="ai-action-suggest"]').trigger('click')
    await nextTick()
    await nextTick()

    expect(mockSuggestLabels).toHaveBeenCalledWith(1, 42, expect.objectContaining({
      name: 'Test Issue',
    }))
    expect(wrapper.find('[data-test="ai-label-suggestions"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('bug')
  })

  it('AI copilot action opens AICopilot without calling analyze APIs', async () => {
    const wrapper = mountComponent()
    await openAiTab(wrapper)

    await wrapper.find('[data-test="ai-action-copilot"]').trigger('click')
    await nextTick()

    expect(mockAnalyzeWithAI).not.toHaveBeenCalled()
    expect(mockSuggestLabels).not.toHaveBeenCalled()
    expect(wrapper.find('[data-test="mock-aicopilot"]').exists()).toBe(true)
  })
})
