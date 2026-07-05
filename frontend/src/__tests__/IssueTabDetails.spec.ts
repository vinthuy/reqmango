import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import IssueTabDetails from '@/components/IssueTabDetails.vue'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))

vi.mock('@/components/RichTextEditor.vue', () => ({
  default: {
    template: '<div class="mock-rte"></div>',
    props: ['modelValue', 'placeholder'],
  },
}))

vi.mock('@/components/CustomFieldManager.vue', () => ({
  default: {
    template: '<div class="mock-cfm">CustomFields</div>',
    props: ['workspaceId', 'projectId', 'issueId', 'issueTypeId', 'mode', 'members'],
  },
}))

vi.mock('@/components/CommentList.vue', () => ({
  default: {
    template: '<div class="mock-comments">Comments</div>',
    props: ['issueId'],
  },
}))

const defaultProps = {
  issueId: 1,
  issue: { id: 1, name: 'Test Issue', description: 'Test description' },
  workspaceId: 1,
  projectId: 1,
  issueTypeId: 1,
  members: [],
}

function mountComponent(overrides: Record<string, any> = {}) {
  return mount(IssueTabDetails, {
    props: { ...defaultProps, ...overrides },
  })
}

describe('IssueTabDetails', () => {
  it('renders "issue.description" text', () => {
    const wrapper = mountComponent()
    expect(wrapper.text()).toContain('issue.description')
  })

  it('renders "issue.customFields" text', () => {
    const wrapper = mountComponent()
    expect(wrapper.text()).toContain('issue.customFields')
  })

  it('renders "issue.comments" text', () => {
    const wrapper = mountComponent()
    expect(wrapper.text()).toContain('issue.comments')
  })

  it('renders the issue title as an editable input with correct value', () => {
    const wrapper = mountComponent()
    const input = wrapper.find('input')
    expect(input.exists()).toBe(true)
    expect(input.element.value).toBe('Test Issue')
  })
})
