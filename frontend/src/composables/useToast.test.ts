/**
 * useToast Composable 单元测试
 * Module-level toasts ref is shared — we clear it between tests
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useToast } from './useToast'

describe('useToast', () => {
  let toastModule: ReturnType<typeof useToast>

  beforeEach(() => {
    vi.useFakeTimers()
    // Get fresh access and clear module-level state
    toastModule = useToast()
    toastModule.toasts.value = []
  })

  it('should add toast with show()', () => {
    toastModule.show('Hello World', 'info')
    expect(toastModule.toasts.value).toHaveLength(1)
    expect(toastModule.toasts.value[0].message).toBe('Hello World')
    expect(toastModule.toasts.value[0].type).toBe('info')
  })

  it('should assign auto-incrementing ids', () => {
    toastModule.show('First')
    toastModule.show('Second')
    expect(toastModule.toasts.value[0].id).toBeLessThan(toastModule.toasts.value[1].id)
  })

  it('success() should create success toast', () => {
    toastModule.success('Operation completed')
    expect(toastModule.toasts.value[0].type).toBe('success')
    expect(toastModule.toasts.value[0].message).toBe('Operation completed')
  })

  it('error() should create error toast with 5s duration', () => {
    toastModule.error('Something went wrong')
    expect(toastModule.toasts.value[0].type).toBe('error')
    expect(toastModule.toasts.value[0].duration).toBe(5000)
  })

  it('warning() should create warning toast with 4s duration', () => {
    toastModule.warning('Be careful')
    expect(toastModule.toasts.value[0].type).toBe('warning')
    expect(toastModule.toasts.value[0].duration).toBe(4000)
  })

  it('info() should use default type and 3s duration', () => {
    toastModule.info('Info message')
    expect(toastModule.toasts.value[0].type).toBe('info')
    expect(toastModule.toasts.value[0].duration).toBe(3000)
  })

  it('should remove toast after duration', () => {
    toastModule.show('Temporary', 'info', 1000)
    expect(toastModule.toasts.value).toHaveLength(1)

    vi.advanceTimersByTime(1000)
    expect(toastModule.toasts.value).toHaveLength(0)
  })

  it('should allow custom duration', () => {
    toastModule.success('Custom', 5000)
    expect(toastModule.toasts.value[0].duration).toBe(5000)
  })

  it('should handle multiple toasts with different durations', () => {
    toastModule.show('Toast 1', 'info', 1000)
    toastModule.show('Toast 2', 'error', 3000)

    expect(toastModule.toasts.value).toHaveLength(2)

    vi.advanceTimersByTime(1000)
    expect(toastModule.toasts.value).toHaveLength(1)
    expect(toastModule.toasts.value[0].message).toBe('Toast 2')

    vi.advanceTimersByTime(2000)
    expect(toastModule.toasts.value).toHaveLength(0)
  })

  it('should return all helper methods', () => {
    expect(typeof toastModule.success).toBe('function')
    expect(typeof toastModule.error).toBe('function')
    expect(typeof toastModule.warning).toBe('function')
    expect(typeof toastModule.info).toBe('function')
    expect(typeof toastModule.show).toBe('function')
  })
})
