/**
 * Lightweight markdown-to-HTML renderer for AI chat output.
 */
export function renderMarkdown(text: string): string {
  if (!text) return ''

  const lines = text.split('\n')
  const out: string[] = []
  let inList: 'ul' | 'ol' | null = null
  let inBlockquote = false

  function flushList() {
    if (inList === 'ul') { out.push('</ul>'); inList = null }
    if (inList === 'ol') { out.push('</ol>'); inList = null }
  }
  function flushBlockquote() {
    if (inBlockquote) { out.push('</blockquote>'); inBlockquote = false }
  }

  for (let i = 0; i < lines.length; i++) {
    let line = lines[i]

    // Escape HTML
    line = line.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

    // Horizontal rule
    if (/^[-*_]{3,}\s*$/.test(line.trim())) {
      flushList()
      flushBlockquote()
      out.push('<hr class="my-2 border-gray-300 dark:border-gray-600">')
      continue
    }

    // Blockquote
    if (line.trim().startsWith('&gt; ') || line.trim().startsWith('> ')) {
      if (!inBlockquote) { out.push('<blockquote class="border-l-3 border-indigo-400 pl-3 my-1.5 italic text-gray-500 dark:text-gray-400">'); inBlockquote = true }
      const content = line.trim().replace(/^&gt;\s*|^>\s*/, '')
      out.push(`<p class="my-0.5">${inlineMarkdown(content)}</p>`)
      continue
    } else {
      flushBlockquote()
    }

    // Headings
    const hMatch = line.match(/^(#{1,6})\s+(.+)/)
    if (hMatch) {
      flushList()
      const level = Math.min(hMatch[1].length, 4) + 1 // h2-h6 -> h3-h6
      const sizeClasses = ['', '', 'text-base font-bold', 'text-sm font-bold', 'text-xs font-bold', 'text-xs font-semibold']
      const spacing = level <= 4 ? 'mt-3 mb-1.5' : 'mt-2 mb-1'
      out.push(`<h${level} class="${spacing} ${sizeClasses[level] || 'text-xs font-medium'} text-gray-900 dark:text-gray-100">${inlineMarkdown(hMatch[2])}</h${level}>`)
      continue
    }

    // Bold title (line that is entirely **text**)
    const boldOnly = line.trim().match(/^\*\*(.+)\*\*$/)
    if (boldOnly && line.trim().length < 100) {
      flushList()
      out.push(`<p class="font-semibold text-gray-900 dark:text-gray-100 mt-2 mb-1">${inlineMarkdown(boldOnly[1])}</p>`)
      continue
    }

    // Unordered list
    const ulMatch = line.match(/^(\s*)[-*]\s+(.+)/)
    if (ulMatch) {
      if (inList !== 'ul') { flushList(); out.push('<ul class="list-disc pl-5 my-1 space-y-0.5">'); inList = 'ul' }
      out.push(`<li class="text-gray-700 dark:text-gray-300">${inlineMarkdown(ulMatch[2])}</li>`)
      continue
    }

    // Ordered list
    const olMatch = line.match(/^(\s*)\d+[.)]\s+(.+)/)
    if (olMatch) {
      if (inList !== 'ol') { flushList(); out.push('<ol class="list-decimal pl-5 my-1 space-y-0.5">'); inList = 'ol' }
      out.push(`<li class="text-gray-700 dark:text-gray-300">${inlineMarkdown(olMatch[2])}</li>`)
      continue
    }

    flushList()

    // Code block (```)
    if (line.trim().startsWith('```')) {
      out.push('<pre class="bg-gray-100 dark:bg-gray-700 rounded p-2 my-1.5 text-xs overflow-x-auto"><code>')
      i++
      while (i < lines.length && !lines[i].trim().startsWith('```')) {
        out.push(lines[i].replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;') + '\n')
        i++
      }
      out.push('</code></pre>')
      continue
    }

    // Inline code
    line = line.replace(/`([^`]+)`/g, '<code class="bg-gray-100 dark:bg-gray-700 px-1 py-0.5 rounded text-xs font-mono text-pink-600 dark:text-pink-400">$1</code>')

    // Table — detect consecutive rows that start AND end with |
    if (line.trim().startsWith('|') && line.trim().endsWith('|')) {
      // Collect all consecutive table rows
      const tableRows: string[][] = []
      let j = i
      while (j < lines.length && lines[j].trim().startsWith('|') && lines[j].trim().endsWith('|')) {
        tableRows.push(lines[j].trim().split('|').slice(1, -1).map(c => c.trim()))
        j++
      }

      if (tableRows.length >= 2) {
        flushList()
        out.push('<div class="overflow-x-auto my-2 rounded-lg border border-gray-200 dark:border-gray-700">')
        out.push('<table class="w-full text-xs">')

        // Header (first row)
        out.push('<thead><tr class="bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">')
        for (const cell of tableRows[0]) {
          out.push(`<th class="text-left px-2.5 py-1.5 text-gray-500 dark:text-gray-400 font-medium whitespace-nowrap">${inlineMarkdown(cell)}</th>`)
        }
        out.push('</tr></thead>')

        // Body (skip row 1 if it's a separator like |---|---|)
        let bodyStart = 1
        if (tableRows[1].every(c => /^[-:]+$/.test(c))) bodyStart = 2

        out.push('<tbody>')
        for (let r = bodyStart; r < tableRows.length; r++) {
          const rowClass = r % 2 === 0 ? 'bg-white dark:bg-gray-900' : 'bg-gray-50/50 dark:bg-gray-850'
          out.push(`<tr class="${rowClass} border-b border-gray-100 dark:border-gray-800 last:border-0">`)
          for (const cell of tableRows[r]) {
            out.push(`<td class="px-2.5 py-1.5 text-gray-700 dark:text-gray-300 whitespace-nowrap">${inlineMarkdown(cell)}</td>`)
          }
          out.push('</tr>')
        }
        out.push('</tbody></table></div>')

        i = j - 1 // skip past table rows
        continue
      }
      // If not enough rows, fall through to regular paragraph
    }

    // Empty line = paragraph break
    if (line.trim() === '') {
      out.push('<div class="h-2"></div>')
      continue
    }

    // Regular paragraph
    out.push(`<p class="my-0.5 text-gray-700 dark:text-gray-300">${inlineMarkdown(line)}</p>`)
  }

  flushList()
  flushBlockquote()

  return out.join('\n')
}

function inlineMarkdown(text: string): string {
  // Bold
  text = text.replace(/\*\*(.+?)\*\*/g, '<strong class="font-semibold text-gray-900 dark:text-gray-100">$1</strong>')
  // Italic
  text = text.replace(/\*(.+?)\*/g, '<em>$1</em>')
  // Strikethrough
  text = text.replace(/~~(.+?)~~/g, '<del class="text-gray-400">$1</del>')
  // Links
  text = text.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" class="text-indigo-600 dark:text-indigo-400 underline" target="_blank">$1</a>')

  return text
}
