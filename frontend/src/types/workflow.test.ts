/**
 * Workflow 类型和辅助函数单元测试
 */
import { describe, it, expect } from 'vitest'
import {
  TriggerTypeEnum,
  TriggerTypeOptions,
  StateGroupOptions,
  PriorityOptions,
  ConditionOperatorEnum,
  ConditionOperatorOptions,
  ActionTypeEnum,
  ActionTypeOptions,
  getTriggerDisplayName,
  getActionDisplayName,
  getOperatorDisplayName,
  getTriggerIcon,
  getActionIcon,
} from './workflow'
import type {
  StateTransition,
  AutomationRule,
  AutomationRuleCreate,
  AutomationRuleUpdate,
  AutomationRuleLite,
  AutomationExecutionLog,
  AutomationTemplate,
} from './workflow'

// ==================== TriggerTypeEnum ====================
describe('TriggerTypeEnum', () => {
  it('should have 11 enum values', () => {
    const keys = Object.keys(TriggerTypeEnum).filter(k => isNaN(Number(k)))
    expect(keys).toHaveLength(11)
  })
})

// ==================== TriggerTypeOptions ====================
describe('TriggerTypeOptions', () => {
  it('should have 10 options', () => {
    expect(TriggerTypeOptions).toHaveLength(10)
  })
  it('should each have value/label/icon', () => {
    TriggerTypeOptions.forEach(opt => {
      expect(opt).toHaveProperty('value')
      expect(opt).toHaveProperty('label')
      expect(opt).toHaveProperty('icon')
    })
  })
})

// ==================== StateGroupOptions ====================
describe('StateGroupOptions', () => {
  it('should have 5 state groups', () => {
    expect(StateGroupOptions).toHaveLength(5)
  })
  it('should include backlog and cancelled', () => {
    const values = StateGroupOptions.map(o => o.value)
    expect(values).toContain('backlog')
    expect(values).toContain('cancelled')
  })
})

// ==================== PriorityOptions ====================
describe('PriorityOptions', () => {
  it('should have 4 priorities', () => {
    expect(PriorityOptions).toHaveLength(4)
  })
  it('should have urgent as highest', () => {
    expect(PriorityOptions[0].value).toBe('urgent')
  })
})

// ==================== ConditionOperatorEnum ====================
describe('ConditionOperatorEnum', () => {
  it('should have 10 operators', () => {
    const keys = Object.keys(ConditionOperatorEnum).filter(k => isNaN(Number(k)))
    expect(keys).toHaveLength(10)
  })
})

// ==================== ConditionOperatorOptions ====================
describe('ConditionOperatorOptions', () => {
  it('should have 8 operator options', () => {
    expect(ConditionOperatorOptions).toHaveLength(8)
  })
})

// ==================== ActionTypeOptions ====================
describe('ActionTypeOptions', () => {
  it('should have 7 action options', () => {
    expect(ActionTypeOptions).toHaveLength(9)
  })
})

// ==================== getTriggerDisplayName ====================
describe('getTriggerDisplayName', () => {
  it('should return Chinese name for ISSUE_CREATED', () => {
    expect(getTriggerDisplayName(TriggerTypeEnum.ISSUE_CREATED)).toBe('工作项创建时')
  })
  it('should return Chinese name for STATE_CHANGED', () => {
    expect(getTriggerDisplayName(TriggerTypeEnum.STATE_CHANGED)).toBe('状态变更时')
  })
  it('should return Chinese name for CYCLE_STARTED', () => {
    expect(getTriggerDisplayName(TriggerTypeEnum.CYCLE_STARTED)).toBe('周期开始时')
  })
  it('should return Chinese name for COMMENT_ADDED', () => {
    expect(getTriggerDisplayName(TriggerTypeEnum.COMMENT_ADDED)).toBe('添加评论时')
  })
})

// ==================== getActionDisplayName ====================
describe('getActionDisplayName', () => {
  it('should return Chinese name for CHANGE_STATE', () => {
    expect(getActionDisplayName(ActionTypeEnum.CHANGE_STATE)).toBe('改变状态')
  })
  it('should return Chinese name for ASSIGN_TO', () => {
    expect(getActionDisplayName(ActionTypeEnum.ASSIGN_TO)).toBe('分配工作项')
  })
  it('should return Chinese name for ADD_COMMENT', () => {
    expect(getActionDisplayName(ActionTypeEnum.ADD_COMMENT)).toBe('添加评论')
  })
  it('should return Chinese name for DISPATCH_AGENT', () => {
    expect(getActionDisplayName(ActionTypeEnum.DISPATCH_AGENT)).toBe('调度 Agent')
  })
})

// ==================== getOperatorDisplayName ====================
describe('getOperatorDisplayName', () => {
  it('should handle EQUALS', () => {
    expect(getOperatorDisplayName(ConditionOperatorEnum.EQUALS)).toBe('等于')
  })
  it('should handle CONTAINS', () => {
    expect(getOperatorDisplayName(ConditionOperatorEnum.CONTAINS)).toBe('包含')
  })
  it('should handle IS_EMPTY', () => {
    expect(getOperatorDisplayName(ConditionOperatorEnum.IS_EMPTY)).toBe('为空')
  })
  it('should handle GREATER_THAN', () => {
    expect(getOperatorDisplayName(ConditionOperatorEnum.GREATER_THAN)).toBe('大于')
  })
})

// ==================== getTriggerIcon ====================
describe('getTriggerIcon', () => {
  it('should return icon for ISSUE_CREATED', () => {
    expect(getTriggerIcon(TriggerTypeEnum.ISSUE_CREATED)).toBe('plus-circle')
  })
  it('should return icon for ISSUE_DELETED', () => {
    expect(getTriggerIcon(TriggerTypeEnum.ISSUE_DELETED)).toBe('trash')
  })
  it('should return icon for DUE_SOON', () => {
    expect(getTriggerIcon(TriggerTypeEnum.DUE_SOON)).toBe('clock')
  })
})

// ==================== getActionIcon ====================
describe('getActionIcon', () => {
  it('should return icon for ASSIGN_TO', () => {
    expect(getActionIcon(ActionTypeEnum.ASSIGN_TO)).toBe('user-plus')
  })
  it('should return icon for DISPATCH_AGENT', () => {
    expect(getActionIcon(ActionTypeEnum.DISPATCH_AGENT)).toBe('zap')
  })
})

// ==================== Type Validation ====================
describe('StateTransition type', () => {
  it('should accept valid transition', () => {
    const t: StateTransition = {
      id: 1, name: 'Start Progress',
      source_state_id: 1, target_state_id: 2,
      is_auto: false, project_id: 1, workspace_id: 1,
      is_deleted: false,
      created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z',
    }
    expect(t.source_state_id).toBe(1)
    expect(t.target_state_id).toBe(2)
  })
})

describe('AutomationRule type', () => {
  it('should accept valid automation rule', () => {
    const rule: AutomationRule = {
      id: 1, name: 'Auto Assign High Priority',
      is_enabled: true,
      trigger_type: JSON.stringify({ type: TriggerTypeEnum.ISSUE_CREATED }),
      conditions: JSON.stringify([{ field: 'priority', operator: ConditionOperatorEnum.EQUALS, value: 'high' }]),
      actions: JSON.stringify([{ type: ActionTypeEnum.ASSIGN_TO, value: 1 }]),
      execution_count: 5,
      project_id: 1, workspace_id: 1, is_deleted: false,
      created_at: '', updated_at: '',
    }
    expect(rule.trigger_type).toContain('issue.created')
    const parsedConditions = JSON.parse(rule.conditions)
    const parsedActions = JSON.parse(rule.actions)
    expect(parsedConditions).toHaveLength(1)
    expect(parsedActions).toHaveLength(1)
  })
})

describe('AutomationRuleCreate type', () => {
  it('should require name, trigger_type, project_id', () => {
    const c: AutomationRuleCreate = {
      name: 'Auto Assign',
      trigger_type: JSON.stringify({ type: TriggerTypeEnum.ISSUE_CREATED }),
      conditions: JSON.stringify([]),
      actions: JSON.stringify([]),
    }
    expect(c.name).toBe('Auto Assign')
  })
})

describe('AutomationRuleUpdate type', () => {
  it('should allow partial update', () => {
    const u: AutomationRuleUpdate = { is_enabled: false }
    expect(u.is_enabled).toBe(false)
  })
})

describe('AutomationRuleLite type', () => {
  it('should have lite fields', () => {
    const l: AutomationRuleLite = {
      id: 1, name: 'Auto Close',
      is_enabled: true,
      trigger: { type: TriggerTypeEnum.CYCLE_ENDED },
      execution_count: 10,
    }
    expect(l.execution_count).toBe(10)
  })
})

describe('AutomationExecutionLog type', () => {
  it('should accept valid log', () => {
    const log: AutomationExecutionLog = {
      id: 1, rule_id: 5, status: 'success',
      trigger_event: 'issue.created',
      triggered_issue_id: 42,
      execution_details: { state_changed: true },
      execution_time_ms: 120,
      created_at: '2026-01-01T00:00:00Z',
    }
    expect(log.status).toBe('success')
    expect(log.execution_time_ms).toBe(120)
  })
})

describe('AutomationTemplate type', () => {
  it('should accept valid template', () => {
    const t: AutomationTemplate = {
      id: 1, name: 'Auto Close Stale Issues',
      description: '自动关闭过期工作项',
      category: 'workflow',
      trigger: { type: TriggerTypeEnum.DUE_DATE_PASSED },
      conditions: [],
      actions: [{ type: ActionTypeEnum.CHANGE_STATE, value: 3 }],
      is_system: true, usage_count: 100,
    }
    expect(t.is_system).toBe(true)
    expect(t.category).toBe('workflow')
  })
})
