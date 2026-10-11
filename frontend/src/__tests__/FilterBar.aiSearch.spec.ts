import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

const { mockSearch } = vi.hoisted(() => ({ mockSearch: vi.fn() }))

vi.mock('@/api/ai', () => ({ searchWithAI: (...args: any[]) => mockSearch(...args) }))
vi.mock('@/api/custom-field', () => ({ listCustomFields: vi.fn().mockResolvedValue([]) }))
vi.mock('@/api/issue', () => ({ suggestIssues: vi.fn().mockResolvedValue([]) }))
vi.mock('@/api', () => ({ default: { get: vi.fn().mockResolvedValue({ data: [] }) } }))
vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ success: vi.fn(), error: vi.fn(), warning: vi.fn(), info: vi.fn() }),
}))
vi.mock('@/composables/useI18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }))

import FilterBar from '@/components/FilterBar.vue'

async function searchFor(rql: string) {
  mockSearch.mockResolvedValue({ rql, explanation: '说明', issues: [{ id: 1, name: '登录页白屏', sequence_id: 1 }] })
  const wrapper = mount(FilterBar, {
    props: { projectId: 1, workspaceId: 1, currentView: 'list', projectIdentifier: 'CORE' },
    global: { stubs: { SavedViewSelector: true, SearchTemplateSelector: true } },
  })
  await flushPromises()
  await wrapper.find('[data-test="quick-search-input"]').setValue('紧急的')
  await wrapper.find('[data-test="ai-search-button"]').trigger('click')
  await flushPromises()
  return wrapper
}

describe('FilterBar AI search', () => {
  beforeEach(() => {
    mockSearch.mockReset()
    setActivePinia(createPinia())
  })

  it('shows a generated filter as an editable condition', async () => {
    const wrapper = await searchFor('priority = "urgent"')

    const source = wrapper.find('[data-test="ai-search-source"]')
    expect(source.text()).toContain('filter.aiSearchSourceFilter')
    expect(wrapper.find('[data-test="ai-search-rql"]').text()).toBe('priority = "urgent"')
    expect(wrapper.find('[data-test="ai-search-apply-rql"]').exists()).toBe(true)
  })

  it('flags a matched-issue list instead of presenting it as a filter', async () => {
    const wrapper = await searchFor('sequence_id IN (1, 7)')

    expect(wrapper.find('[data-test="ai-search-source"]').text()).toContain('filter.aiSearchSourceList')
    expect(wrapper.find('[data-test="ai-search-rql"]').exists()).toBe(false)
  })

  it('shows no source line when nothing was matched', async () => {
    const wrapper = await searchFor('')

    expect(wrapper.find('[data-test="ai-search-results"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="ai-search-source"]').exists()).toBe(false)
  })
})
