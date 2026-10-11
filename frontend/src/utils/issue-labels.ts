import { useI18n } from '@/composables/useI18n'

export function getPriorityName(priority: string): string {
  const { t } = useI18n()
  const names: Record<string, string> = {
    urgent: t('issueLabels.priority.urgent'),
    high: t('issueLabels.priority.high'),
    medium: t('issueLabels.priority.medium'),
    low: t('issueLabels.priority.low'),
    none: t('issueLabels.priority.none'),
  }
  return names[priority] || priority
}

export function getStateGroupName(group: string): string {
  const { t } = useI18n()
  const names: Record<string, string> = {
    backlog: t('issueLabels.stateGroup.backlog'),
    todo: t('issueLabels.stateGroup.todo'),
    in_progress: t('issueLabels.stateGroup.in_progress'),
    done: t('issueLabels.stateGroup.done'),
    cancelled: t('issueLabels.stateGroup.cancelled'),
  }
  return names[group] || group
}
