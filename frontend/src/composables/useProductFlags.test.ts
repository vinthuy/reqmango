import { describe, it, expect, beforeEach } from 'vitest'
import {
  isInitiativesEnabled,
  setInitiativesEnabled,
  SHOW_INITIATIVES_KEY,
} from './useProductFlags'

describe('useProductFlags initiatives', () => {
  beforeEach(() => {
    localStorage.removeItem(SHOW_INITIATIVES_KEY)
  })

  it('defaults to hidden', () => {
    expect(isInitiativesEnabled()).toBe(false)
  })

  it('enables and disables via setter', () => {
    setInitiativesEnabled(true)
    expect(isInitiativesEnabled()).toBe(true)
    expect(localStorage.getItem(SHOW_INITIATIVES_KEY)).toBe('1')
    setInitiativesEnabled(false)
    expect(isInitiativesEnabled()).toBe(false)
  })
})
