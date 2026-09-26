import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import IssueDetailHeader from '@/components/IssueDetailHeader.vue'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))
vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ success: vi.fn(), error: vi.fn() }),
}))

const mockIssue = {
  id: 42,
  sequence_id: 42,
  name: 'Test Issue',
  issue_type: { id: 1, name: 'Task', color: '#6366f1' },
}

describe('IssueDetailHeader', () => {
  it('renders the issue type badge and sequence ID', () => {
    const wrapper = mount(IssueDetailHeader, { props: { issue: mockIssue, projectIdentifier: 'DEV' } })
    expect(wrapper.text()).toContain('Task')
    expect(wrapper.text()).toContain('DEV-42')
  })

  it('emits "back" when back button is clicked', () => {
    const wrapper = mount(IssueDetailHeader, { props: { issue: mockIssue } })
    wrapper.find('[data-test="back-btn"]').trigger('click')
    expect(wrapper.emitted('back')).toBeTruthy()
    expect(wrapper.emitted('back')!.length).toBe(1)
  })

  it('shows editable title input', () => {
    const wrapper = mount(IssueDetailHeader, { props: { issue: mockIssue } })
    const input = wrapper.find('[data-test="issue-title-input"]')
    expect(input.exists()).toBe(true)
    expect((input.element as HTMLInputElement).value).toBe('Test Issue')
  })

  it('emits update:title on blur when title changes', async () => {
    const wrapper = mount(IssueDetailHeader, { props: { issue: mockIssue } })
    const input = wrapper.find('[data-test="issue-title-input"]')
    await input.setValue('Renamed Issue')
    await input.trigger('blur')
    expect(wrapper.emitted('update:title')).toBeTruthy()
    expect(wrapper.emitted('update:title')![0]).toEqual(['Renamed Issue'])
  })

  it('does not show a primary save button', () => {
    const wrapper = mount(IssueDetailHeader, { props: { issue: mockIssue } })
    expect(wrapper.find('[data-test="save-btn"]').exists()).toBe(false)
  })
})
