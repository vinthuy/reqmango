import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'

// Simple component test — validates Vue Test Utils works
const HelloWorld = defineComponent({
  props: { msg: { type: String, default: 'Hello' } },
  setup(props) {
    return () => h('div', { class: 'greeting' }, props.msg)
  },
})

describe('HelloWorld', () => {
  it('renders message prop', () => {
    const wrapper = mount(HelloWorld, { props: { msg: 'Hello Vitest!' } })
    expect(wrapper.text()).toBe('Hello Vitest!')
    expect(wrapper.classes()).toContain('greeting')
  })

  it('renders default message', () => {
    const wrapper = mount(HelloWorld)
    expect(wrapper.text()).toBe('Hello')
  })
})

describe('Test Infrastructure', () => {
  it('should have vitest working', () => {
    expect(1 + 1).toBe(2)
  })

  it('should have Vue Test Utils working', () => {
    const wrapper = mount(HelloWorld)
    expect(wrapper.exists()).toBe(true)
  })
})
