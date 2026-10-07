import { describe, it, expect } from 'vitest'
import { buildDispatchTask } from './agentDispatchTask'

const t = (key: string, params?: Record<string, string>) =>
  `${key}|${params?.seq}|${params?.name}`

const issue = { seq: '81', name: '慢查询' }

describe('buildDispatchTask', () => {
  it('uses the triage prompt for the built-in 请求分诊 agent', () => {
    expect(buildDispatchTask({ name: '请求分诊', capabilities: [] }, issue, t)).toBe(
      'agent.dispatchTaskTriage|81|慢查询',
    )
  })

  it('uses specialised prompts for the other built-in PM agents', () => {
    expect(buildDispatchTask({ name: '交付风险', capabilities: [] }, issue, t)).toContain(
      'dispatchTaskRisk',
    )
    expect(buildDispatchTask({ name: 'Sprint 总结', capabilities: [] }, issue, t)).toContain(
      'dispatchTaskSprint',
    )
    expect(buildDispatchTask({ name: 'Spec 草稿', capabilities: [] }, issue, t)).toContain(
      'dispatchTaskSpec',
    )
  })

  it('falls back to triage when a custom agent can propose changes', () => {
    expect(
      buildDispatchTask(
        { name: '自定义分诊', capabilities: ['get_issue', 'suggest_issue_changes'] },
        issue,
        t,
      ),
    ).toBe('agent.dispatchTaskTriage|81|慢查询')
  })

  it('uses the generic prompt otherwise', () => {
    expect(buildDispatchTask({ name: '随便问问', capabilities: ['get_issue'] }, issue, t)).toBe(
      'agent.dispatchTask|81|慢查询',
    )
    expect(buildDispatchTask(null, issue, t)).toBe('agent.dispatchTask|81|慢查询')
  })
})
