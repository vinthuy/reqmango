export function highlightSearchTerm(text: string, term: string): string {
  if (!term || !text) return text
  
  const escaped = term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const regex = new RegExp(`(${escaped})`, 'gi')
  return text.replace(regex, '<mark class="bg-yellow-200 text-gray-900 px-0.5 rounded">$1</mark>')
}

export function extractSearchTerm(rql: string): string {
  const match = rql.match(/name\s+LIKE\s+["']([^"']+)["']/)
  if (match) {
    return match[1].replace(/^%/, '').replace(/%$/, '')
  }
  
  const quickMatch = rql.match(/["']([^"']+)["']/)
  if (quickMatch) {
    return quickMatch[1]
  }
  
  return ''
}