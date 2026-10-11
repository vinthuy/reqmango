/** Per-project preference for the create-time auto-triage checkbox. */

const prefKey = (projectId: number) => `reqmango:auto_triage_pref:${projectId}`
const pendingKey = (issueId: number) => `reqmango:auto_triage_pending:${issueId}`

export function getAutoTriagePref(projectId: number): boolean {
  if (!projectId) return false
  try {
    return localStorage.getItem(prefKey(projectId)) === '1'
  } catch {
    return false
  }
}

export function setAutoTriagePref(projectId: number, enabled: boolean): void {
  if (!projectId) return
  try {
    if (enabled) localStorage.setItem(prefKey(projectId), '1')
    else localStorage.removeItem(prefKey(projectId))
  } catch { /* ignore */ }
}

/** Mark an issue so the detail view starts watching for triage completion. */
export function markAutoTriagePending(issueId: number): void {
  if (!issueId) return
  try {
    sessionStorage.setItem(pendingKey(issueId), String(Date.now()))
  } catch { /* ignore */ }
}

export function isAutoTriagePending(issueId: number): boolean {
  if (!issueId) return false
  try {
    return !!sessionStorage.getItem(pendingKey(issueId))
  } catch {
    return false
  }
}

export function clearAutoTriagePending(issueId: number): void {
  if (!issueId) return
  try {
    sessionStorage.removeItem(pendingKey(issueId))
  } catch { /* ignore */ }
}
