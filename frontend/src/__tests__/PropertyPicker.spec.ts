import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import PropertyPicker from '@/components/PropertyPicker.vue'

vi.mock('@/composables/useI18n', () => ({
  useI18n: () => ({ t: (k: string) => k }),
}))

const options = [
  { value: 1, label: 'To Do', color: '#999' },
  { value: 2, label: 'In Progress', color: '#36f' },
  { value: 3, label: 'Done', color: '#0a0' },
]

function mountPicker(props: Record<string, unknown> = {}) {
  return mount(PropertyPicker, {
    props: { modelValue: 1, options, ...props },
    global: { stubs: { teleport: true } },
  })
}

describe('PropertyPicker', () => {
  it('shows the selected option label on the trigger', () => {
    const wrapper = mountPicker()
    expect(wrapper.find('button').text()).toContain('To Do')
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
  })

  it('shows placeholder when nothing is selected', () => {
    const wrapper = mountPicker({ modelValue: null, placeholder: 'Pick one' })
    expect(wrapper.find('button').text()).toContain('Pick one')
  })

  it('opens on click and emits the clicked option', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')
    const items = wrapper.findAll('[role="option"]')
    expect(items).toHaveLength(3)
    await items[2].trigger('mousedown')
    expect(wrapper.emitted('update:modelValue')![0]).toEqual([3])
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
  })

  it('filters options by search query', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')
    await wrapper.find('input').setValue('prog')
    const items = wrapper.findAll('[role="option"]')
    expect(items).toHaveLength(1)
    expect(items[0].text()).toContain('In Progress')
  })

  it('supports keyboard navigation and enter to select', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')
    const input = wrapper.find('input')
    await input.trigger('keydown', { key: 'ArrowDown' })
    await input.trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('update:modelValue')![0]).toEqual([2])
  })

  it('closes on escape without emitting', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')
    await wrapper.find('input').trigger('keydown', { key: 'Escape' })
    await nextTick()
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
    expect(wrapper.emitted('update:modelValue')).toBeFalsy()
  })

  it('offers a clear option when clearable', async () => {
    const wrapper = mountPicker({ clearable: true })
    await wrapper.find('button').trigger('click')
    const items = wrapper.findAll('[role="option"]')
    expect(items).toHaveLength(4)
    await items[0].trigger('mousedown')
    expect(wrapper.emitted('update:modelValue')![0]).toEqual([null])
  })

  it('does not emit when picking the current value', async () => {
    const wrapper = mountPicker()
    await wrapper.find('button').trigger('click')
    await wrapper.findAll('[role="option"]')[0].trigger('mousedown')
    expect(wrapper.emitted('update:modelValue')).toBeFalsy()
  })

  it('does not open when disabled', async () => {
    const wrapper = mountPicker({ disabled: true })
    await wrapper.find('button').trigger('click')
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
  })
})
