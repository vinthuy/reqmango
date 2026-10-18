import { afterEach, describe, expect, it } from 'vitest'
import { useI18n } from '@/composables/useI18n'
import { getPriorityName, getStateGroupName } from '@/types/issue'
import { getPriorityName as settingsPriority, getStateGroupName as settingsState } from '@/types/project-settings'

afterEach(() => useI18n().setLocale('zh-CN'))

describe('issue labels', () => {
  it('uses the active locale through both public helper modules', () => {
    const { setLocale } = useI18n()
    const priorities = ['urgent', 'high', 'medium', 'low', 'none']
    const groups = ['backlog', 'todo', 'in_progress', 'done', 'cancelled']
    setLocale('zh-CN')
    expect(priorities.map(getPriorityName)).toEqual(['紧急', '高', '中', '低', '无'])
    expect(groups.map(getStateGroupName)).toEqual(['待办', '计划中', '进行中', '已完成', '已取消'])
    expect(groups.map(settingsState)).toEqual(['待办池', '待办', '进行中', '已完成', '已取消'])
    setLocale('en-US')
    expect(priorities.map(getPriorityName)).toEqual(['Urgent', 'High', 'Medium', 'Low', 'None'])
    expect(groups.map(getStateGroupName)).toEqual(['Backlog', 'Todo', 'In Progress', 'Done', 'Cancelled'])
    expect(priorities.map(settingsPriority)).toEqual(priorities.map(getPriorityName))
    expect(groups.map(settingsState)).toEqual(groups.map(getStateGroupName))
  })

  it('preserves unknown values as a fallback', () => {
    expect(getPriorityName('custom')).toBe('custom')
    expect(getStateGroupName('custom')).toBe('custom')
    expect(settingsState('custom')).toBe('custom')
  })
})
