import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import IssueDetailHeader from '@/components/IssueDetailHeader.vue'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))

const mockIssue = {
  id: 42,
  sequence_id: 42,
  name: 'Test Issue',
  issue_type: { id: 1, name: 'Task', color: '#6366f1' },
}

describe('IssueDetailHeader', () => {
  it('renders the issue type badge and sequence ID', () => {
    const wrapper = mount(IssueDetailHeader, { props: { issue: mockIssue, saving: false, projectIdentifier: 'DEV' } })
    expect(wrapper.text()).toContain('Task')
    expect(wrapper.text()).toContain('DEV-42')
  })

  it('emits "back" when back button is clicked', () => {
    const wrapper = mount(IssueDetailHeader, { props: { issue: mockIssue, saving: false } })
    wrapper.find('[data-test="back-btn"]').trigger('click')
    expect(wrapper.emitted('back')).toBeTruthy()
    expect(wrapper.emitted('back')!.length).toBe(1)
  })

  it('emits "save" when save button is clicked', () => {
    const wrapper = mount(IssueDetailHeader, { props: { issue: mockIssue, saving: false } })
    wrapper.find('[data-test="save-btn"]').trigger('click')
    expect(wrapper.emitted('save')).toBeTruthy()
    expect(wrapper.emitted('save')!.length).toBe(1)
  })

  it('disables save button when saving is true', () => {
    const wrapper = mount(IssueDetailHeader, { props: { issue: mockIssue, saving: true } })
    const btn = wrapper.find('[data-test="save-btn"]')
    expect((btn.element as HTMLButtonElement).disabled).toBe(true)
  })

  it('shows saving text when saving is true', () => {
    const wrapper = mount(IssueDetailHeader, { props: { issue: mockIssue, saving: true } })
    expect(wrapper.text()).toContain('issue.saving')
    expect(wrapper.text()).not.toContain('issue.save')
  })
})
