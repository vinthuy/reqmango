/**
 * Agent capabilities are stored either as editor categories ("search") or as
 * tool names ("get_issue"); both are shown with the editor's category labels.
 */
const CAPABILITY_LABEL_KEYS: Record<string, string> = {
  search: 'agent.search',
  create: 'agent.createIssues',
  update: 'agent.updateIssues',
  analyze: 'agent.analyze',
  comment: 'agent.comment',
  list: 'agent.listResources',
  summarize: 'agent.summarize',

  search_issues: 'agent.search',
  get_issue: 'agent.search',
  get_issue_activities: 'agent.search',
  create_issue: 'agent.createIssues',
  update_issue: 'agent.updateIssues',
  get_project_stats: 'agent.analyze',
  get_issues_summary: 'agent.analyze',
  get_cycle_progress: 'agent.analyze',
  add_comment: 'agent.comment',
  suggest_issue_changes: 'agent.suggestChanges',

  triage: 'agent.capTriage',
  label: 'agent.capLabel',
  assign: 'agent.capAssign',
  plan: 'agent.capPlan',
}

function labelKey(cap: string): string | undefined {
  if (CAPABILITY_LABEL_KEYS[cap]) return CAPABILITY_LABEL_KEYS[cap]
  if (cap.startsWith('list_')) return 'agent.listResources'
  return undefined
}

export function capabilityLabels(caps: string[] | undefined | null, t: (key: string) => string): string[] {
  const labels: string[] = []
  for (const cap of caps ?? []) {
    const key = labelKey(cap)
    const label = key ? t(key) : cap
    if (!labels.includes(label)) labels.push(label)
  }
  return labels
}
