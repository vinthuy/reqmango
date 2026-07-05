export function stateColor(group: string) {
  const m: Record<string, string> = { done: '#22c55e', in_progress: '#3b82f6', backlog: '#9ca3af', todo: '#9ca3af', cancelled: '#ef4444' }
  return m[group] || '#9ca3af'
}

export function priorityColor(p: string) {
  const m: Record<string, string> = { urgent: '#ef4444', high: '#f97316', medium: '#eab308', low: '#22c55e', none: '#9ca3af' }
  return m[p] || m['none']
}

export function formatDate(d: string) {
  if (!d) return '—'
  const date = new Date(d)
  if (isNaN(date.getTime())) return '—'
  return `${date.getMonth() + 1}/${date.getDate()}`
}
