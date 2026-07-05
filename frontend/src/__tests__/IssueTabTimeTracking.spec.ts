import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import IssueTabTimeTracking from '@/components/IssueTabTimeTracking.vue'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))

vi.mock('@/components/TimeTrackPanel.vue', () => ({
  default: {
    template: '<div class="mock-time">TimeTracking</div>',
    props: ['issueId'],
  },
}))

const defaultProps = {
  issueId: 1,
}

function mountComponent(overrides: Record<string, any> = {}) {
  return mount(IssueTabTimeTracking, {
    props: { ...defaultProps, ...overrides },
  })
}

describe('IssueTabTimeTracking', () => {
  it('renders TimeTrackPanel', () => {
    const wrapper = mountComponent()
    expect(wrapper.text()).toContain('TimeTracking')
  })

  it('passes issueId prop', () => {
    const wrapper = mountComponent({ issueId: 42 })
    expect(wrapper.exists()).toBe(true)
    expect(wrapper.find('.mock-time').exists()).toBe(true)
  })
})
