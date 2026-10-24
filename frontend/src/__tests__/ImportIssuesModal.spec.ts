import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ImportIssuesModal from '@/components/ImportIssuesModal.vue'

const previewImportCSV = vi.fn()
const importIssuesCSV = vi.fn()
const previewImportJSON = vi.fn()

vi.mock('@/api/issue', async () => {
  const actual = await vi.importActual<typeof import('@/api/issue')>('@/api/issue')
  return {
    ...actual,
    default: {
      previewImportCSV: (...a: unknown[]) => previewImportCSV(...a),
      importIssuesCSV: (...a: unknown[]) => importIssuesCSV(...a),
      previewImportJSON: (...a: unknown[]) => previewImportJSON(...a),
      downloadImportTemplate: vi.fn(),
      importIssuesJSON: vi.fn(),
    },
  }
})
vi.mock('@/composables/useI18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }))

function mountModal() {
  return mount(ImportIssuesModal, {
    props: { visible: true, projectId: 4, workspaceId: 1 },
  })
}

function csvFile() {
  return new File(['title,state\nA,Todo\n'], 'issues.csv', { type: 'text/csv' })
}

describe('ImportIssuesModal', () => {
  beforeEach(() => {
    previewImportCSV.mockReset()
    importIssuesCSV.mockReset()
    previewImportJSON.mockReset()
  })

  it('previews the file as soon as it is selected', async () => {
    previewImportCSV.mockResolvedValue({
      success_count: 1,
      fail_count: 0,
      errors: [],
      imported_ids: [],
      dry_run: true,
      total_rows: 1,
      headers: ['title', 'state'],
      mapping: { name: 'title', state: 'state' },
      warnings: [],
    })
    const wrapper = mountModal()
    const input = wrapper.find('input[type="file"]')
    Object.defineProperty(input.element, 'files', { value: [csvFile()], configurable: true })
    await input.trigger('change')
    await flushPromises()

    expect(previewImportCSV).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('import.previewSummary')
  })

  it('renders one mapping select per canonical field and re-previews on change', async () => {
    previewImportCSV.mockResolvedValue({
      success_count: 1,
      fail_count: 0,
      errors: [],
      imported_ids: [],
      dry_run: true,
      total_rows: 1,
      headers: ['title', 'state'],
      mapping: { name: 'title', state: 'state' },
      warnings: [],
    })
    const wrapper = mountModal()
    const input = wrapper.find('input[type="file"]')
    Object.defineProperty(input.element, 'files', { value: [csvFile()], configurable: true })
    await input.trigger('change')
    await flushPromises()

    const selects = wrapper.findAll('[data-testid^="import-map-"]')
    expect(selects).toHaveLength(13)

    const stateSelect = wrapper.find('[data-testid="import-map-state"]')
    expect((stateSelect.element as HTMLSelectElement).value).toBe('state')
    await stateSelect.setValue('')
    await flushPromises()

    expect(previewImportCSV).toHaveBeenCalledTimes(2)
    const lastMapping = previewImportCSV.mock.calls[1][3]
    expect(lastMapping).not.toHaveProperty('state')
    expect(lastMapping.name).toBe('title')
  })

  it('imports with the mapping shown in the preview', async () => {
    previewImportCSV.mockResolvedValue({
      success_count: 1,
      fail_count: 0,
      errors: [],
      imported_ids: [],
      dry_run: true,
      total_rows: 1,
      headers: ['title'],
      mapping: { name: 'title' },
      warnings: [],
    })
    importIssuesCSV.mockResolvedValue({ success_count: 1, fail_count: 0, errors: [], imported_ids: [9] })

    const wrapper = mountModal()
    const input = wrapper.find('input[type="file"]')
    Object.defineProperty(input.element, 'files', { value: [csvFile()], configurable: true })
    await input.trigger('change')
    await flushPromises()

    const importButton = wrapper.findAll('button').find(b => b.text().includes('import.startImport'))
    expect(importButton).toBeTruthy()
    await importButton!.trigger('click')
    await flushPromises()

    expect(importIssuesCSV).toHaveBeenCalledTimes(1)
    expect(importIssuesCSV.mock.calls[0][3]).toEqual({ mapping: { name: 'title' } })
  })

  it('surfaces preview failures instead of offering an import', async () => {
    previewImportCSV.mockRejectedValue({ response: { data: { message: 'CSV 缺少标题列' } } })
    const wrapper = mountModal()
    const input = wrapper.find('input[type="file"]')
    Object.defineProperty(input.element, 'files', { value: [csvFile()], configurable: true })
    await input.trigger('change')
    await flushPromises()

    expect(wrapper.text()).toContain('CSV 缺少标题列')
    const importButton = wrapper.findAll('button').find(b => b.text().includes('import.startImport'))
    expect(importButton!.attributes('disabled')).toBeDefined()
  })

  // Auto-detection fails on unfamiliar headers; the mapping UI must still be
  // usable, otherwise such a file can never be imported.
  it('keeps the mapping UI when auto-detection fails and recovers after mapping', async () => {
    previewImportCSV.mockRejectedValueOnce({ response: { data: { message: 'CSV 缺少标题列' } } })
    previewImportCSV.mockResolvedValueOnce({
      success_count: 1,
      fail_count: 0,
      errors: [],
      imported_ids: [],
      dry_run: true,
      total_rows: 1,
      headers: ['Work Item', 'Stage'],
      mapping: { name: 'Work Item', state: 'Stage' },
      warnings: [],
    })

    const wrapper = mountModal()
    const input = wrapper.find('input[type="file"]')
    const file = new File(['Work Item,Stage\nA,Todo\n'], 'custom.csv', { type: 'text/csv' })
    Object.defineProperty(input.element, 'files', { value: [file], configurable: true })
    await input.trigger('change')
    await flushPromises()

    const nameSelect = wrapper.find('[data-testid="import-map-name"]')
    expect(nameSelect.exists()).toBe(true)
    const options = nameSelect.findAll('option').map(o => o.text())
    expect(options).toContain('Work Item')

    await nameSelect.setValue('Work Item')
    await flushPromises()

    expect(previewImportCSV).toHaveBeenCalledTimes(2)
    expect(previewImportCSV.mock.calls[1][3]).toEqual({ name: 'Work Item' })
    expect(wrapper.text()).toContain('import.previewSummary')
  })
})
