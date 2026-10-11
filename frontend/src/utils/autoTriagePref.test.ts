import { beforeEach, describe, expect, it } from 'vitest'
import {
  clearAutoTriagePending,
  getAutoTriagePref,
  isAutoTriagePending,
  markAutoTriagePending,
  setAutoTriagePref,
} from './autoTriagePref'

describe('autoTriagePref', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
  })

  it('stores per-project checkbox preference', () => {
    expect(getAutoTriagePref(9)).toBe(false)
    setAutoTriagePref(9, true)
    expect(getAutoTriagePref(9)).toBe(true)
    setAutoTriagePref(9, false)
    expect(getAutoTriagePref(9)).toBe(false)
  })

  it('tracks pending triage for an issue in sessionStorage', () => {
    expect(isAutoTriagePending(42)).toBe(false)
    markAutoTriagePending(42)
    expect(isAutoTriagePending(42)).toBe(true)
    clearAutoTriagePending(42)
    expect(isAutoTriagePending(42)).toBe(false)
  })
})
