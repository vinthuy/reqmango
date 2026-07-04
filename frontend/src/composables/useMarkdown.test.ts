/**
 * renderMarkdown 单元测试
 */
import { describe, it, expect } from 'vitest'
import { renderMarkdown } from './useMarkdown'

describe('renderMarkdown', () => {
  it('should return empty string for falsy input', () => {
    expect(renderMarkdown('')).toBe('')
  })

  it('should render headings (h2 becomes h3 due to offset)', () => {
    const html = renderMarkdown('## Hello World')
    expect(html).toContain('<h3')
    expect(html).toContain('Hello World')
  })

  it('should render bold text', () => {
    const html = renderMarkdown('This is **bold** text')
    expect(html).toContain('<strong')
    expect(html).toContain('bold')
  })

  it('should render italic text', () => {
    const html = renderMarkdown('This is *italic* text')
    expect(html).toContain('<em>italic</em>')
  })

  it('should render strikethrough', () => {
    const html = renderMarkdown('This is ~~deleted~~ text')
    expect(html).toContain('<del')
    expect(html).toContain('deleted')
  })

  it('should render links', () => {
    const html = renderMarkdown('[Click here](https://example.com)')
    expect(html).toContain('<a href="https://example.com"')
    expect(html).toContain('Click here')
  })

  it('should render inline code', () => {
    const html = renderMarkdown('Use `npm install` command')
    expect(html).toContain('<code')
    expect(html).toContain('npm install')
  })

  it('should render unordered list', () => {
    const html = renderMarkdown('- Item 1\n- Item 2\n- Item 3')
    expect(html).toContain('<ul')
    expect(html).toContain('<li')
    expect(html).toContain('Item 1')
    expect(html).toContain('Item 2')
    expect(html).toContain('Item 3')
  })

  it('should render ordered list', () => {
    const html = renderMarkdown('1. First\n2. Second\n3. Third')
    expect(html).toContain('<ol')
    expect(html).toContain('First')
    expect(html).toContain('Second')
  })

  it('should render code blocks', () => {
    const html = renderMarkdown('```\nconst x = 1;\nconsole.log(x);\n```')
    expect(html).toContain('<pre')
    expect(html).toContain('const x = 1')
    expect(html).toContain('console.log')
  })

  it('should render horizontal rule', () => {
    const html = renderMarkdown('---')
    expect(html).toContain('<hr')
  })

  it('should render blockquotes from HTML-escaped >', () => {
    // In real use, > might be already escaped
    const html = renderMarkdown('> This is a quote')
    expect(html).toContain('<blockquote')
    expect(html).toContain('This is a quote')
  })

  it('should render tables', () => {
    const table = [
      '| Name | Status |',
      '|------|--------|',
      '| Task 1 | Done |',
      '| Task 2 | In Progress |',
    ].join('\n')
    const html = renderMarkdown(table)
    expect(html).toContain('<table')
    expect(html).toContain('Name')
    expect(html).toContain('Task 1')
    expect(html).toContain('Done')
  })

  it('should escape HTML in content', () => {
    const html = renderMarkdown('<script>alert("xss")</script>')
    expect(html).not.toContain('<script>')
    expect(html).toContain('&lt;script&gt;')
  })

  it('should render bold title lines', () => {
    const html = renderMarkdown('**Important Notice**')
    expect(html).toContain('font-semibold')
    expect(html).toContain('Important Notice')
  })

  it('should render regular paragraphs', () => {
    const html = renderMarkdown('This is a regular paragraph.')
    expect(html).toContain('<p')
    expect(html).toContain('regular paragraph')
  })

  it('should handle empty paragraph breaks', () => {
    const html = renderMarkdown('Line 1\n\nLine 2')
    expect(html).toContain('<div class="h-2"></div>')
  })
})
