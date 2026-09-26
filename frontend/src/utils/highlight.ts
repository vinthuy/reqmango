export function highlightSearchTerm(text: string, term: string): string {
  if (!term || !text) return text

  // For KEY-123 style searches, highlight the numeric part in sequence columns / names
  const keyMatch = term.trim().match(/^([A-Za-z][A-Za-z0-9]*)-(\d+)$/)
  const highlightSource = keyMatch ? keyMatch[2] : term

  const escaped = highlightSource.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  if (!escaped) return text
  const regex = new RegExp(`(${escaped})`, 'gi')
  return text.replace(regex, '<mark class="bg-yellow-200 text-gray-900 px-0.5 rounded">$1</mark>')
}

export function extractSearchTerm(rql: string): string {
  const likeMatch = rql.match(/name\s+LIKE\s+"%?((?:[^"%\\]|\\.)*)%?"/i)
  if (likeMatch) {
    return likeMatch[1].replace(/\\(.)/g, '$1')
  }

  const seqMatch = rql.match(/\bsequence_id\s*=\s*(\d+)\b/i)
  if (seqMatch) {
    return seqMatch[1]
  }

  return ''
}