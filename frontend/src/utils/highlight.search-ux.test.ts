import { describe, it, expect } from 'vitest'
import { highlightSearchTerm, extractSearchTerm } from './highlight'

describe('search UX highlight helpers', () => {
  it('highlights title substrings', () => {
    const html = highlightSearchTerm('fix login bug', 'login')
    expect(html).toContain('<mark')
    expect(html).toContain('login')
  })

  it('for KEY-n terms highlights the numeric part', () => {
    const html = highlightSearchTerm('ABC-12', 'abc-12')
    expect(html).toContain('<mark')
    expect(html).toContain('12')
  })

  it('extracts LIKE term without percent signs', () => {
    expect(extractSearchTerm('(name LIKE "%bug%" OR description LIKE "%bug%")')).toBe('bug')
  })

  it('extracts sequence_id from key RQL', () => {
    expect(extractSearchTerm('sequence_id = 12')).toBe('12')
  })
})
