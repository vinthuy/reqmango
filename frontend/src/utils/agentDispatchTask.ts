import type { Agent } from '@/types/agent'

export type DispatchIssueRef = { seq: string; name: string }

type Translate = (key: string, params?: Record<string, string>) => string

/** Built-in PM agents seeded by EnsurePMAgents — match by exact name. */
const BUILTIN_TASK_KEYS: Record<string, string> = {
  请求分诊: 'agent.dispatchTaskTriage',
  交付风险: 'agent.dispatchTaskRisk',
  'Sprint 总结': 'agent.dispatchTaskSprint',
  'Spec 草稿': 'agent.dispatchTaskSpec',
}

/**
 * Pick a dispatch prompt that matches what the agent is for.
 * Prefer the built-in name; otherwise, agents that can propose field changes
 * get the triage prompt so they actually call suggest_issue_changes.
 */
export function buildDispatchTask(
  agent: Pick<Agent, 'name' | 'capabilities'> | null | undefined,
  issue: DispatchIssueRef,
  t: Translate,
): string {
  const params = { seq: issue.seq, name: issue.name }
  const byName = agent?.name ? BUILTIN_TASK_KEYS[agent.name] : undefined
  if (byName) return t(byName, params)

  if (agent?.capabilities?.includes('suggest_issue_changes')) {
    return t('agent.dispatchTaskTriage', params)
  }
  return t('agent.dispatchTask', params)
}
