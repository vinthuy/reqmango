"""
Workflow Services - 工作流和自动化业务逻辑层
"""
from typing import Optional, List
from datetime import datetime

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload

from app.models.workflow import (
    StateTransition,
    AutomationRule,
    AutomationExecutionLog,
    AutomationTemplate
)
from app.models.state import State
from app.models.project import Project
from app.schemas.workflow import (
    StateTransitionCreate,
    StateTransitionUpdate,
    AutomationRuleCreate,
    AutomationRuleUpdate,
    DEFAULT_STATE_TRANSITIONS
)
from app.core.exceptions import NotFoundException, ValidationException


# ==================== State Transition Service ====================

async def create_state_transition(
    db: AsyncSession,
    transition_data: StateTransitionCreate,
    workspace_id: int,
    user_id: int
) -> StateTransition:
    """创建状态转换规则"""
    # 验证项目存在
    project = await db.get(Project, transition_data.project_id)
    if not project or project.is_deleted:
        raise NotFoundException("Project not found")
    
    # 验证源状态和目标状态存在
    source_state = await db.get(State, transition_data.source_state_id)
    if not source_state or source_state.is_deleted:
        raise NotFoundException("Source state not found")
    
    target_state = await db.get(State, transition_data.target_state_id)
    if not target_state or target_state.is_deleted:
        raise NotFoundException("Target state not found")
    
    # 验证不是同一状态
    if transition_data.source_state_id == transition_data.target_state_id:
        raise ValidationException("Source and target state cannot be the same")
    
    # 创建转换规则
    transition = StateTransition(
        name=transition_data.name,
        description=transition_data.description,
        source_state_id=transition_data.source_state_id,
        target_state_id=transition_data.target_state_id,
        issue_type_id=transition_data.issue_type_id,
        is_auto=transition_data.is_auto,
        project_id=transition_data.project_id,
        workspace_id=workspace_id,
        created_by_id=user_id
    )
    
    db.add(transition)
    await db.commit()
    await db.refresh(transition)
    return transition


async def get_state_transition_by_id(
    db: AsyncSession,
    transition_id: int
) -> StateTransition:
    """获取状态转换规则"""
    result = await db.execute(
        select(StateTransition).where(StateTransition.id == transition_id)
    )
    transition = result.scalar_one_or_none()
    if not transition or transition.is_deleted:
        raise NotFoundException("State transition not found")
    return transition


async def get_project_state_transitions(
    db: AsyncSession,
    project_id: int,
    source_state_id: Optional[int] = None
) -> List[StateTransition]:
    """获取项目的状态转换规则"""
    query = select(StateTransition).where(
        StateTransition.project_id == project_id,
        StateTransition.is_deleted == False
    )
    
    if source_state_id:
        query = query.where(StateTransition.source_state_id == source_state_id)
    
    query = query.order_by(StateTransition.name)
    
    result = await db.execute(query)
    return list(result.scalars().all())


async def can_transition(
    db: AsyncSession,
    project_id: int,
    source_state_id: int,
    target_state_id: int,
    issue_type_id: Optional[int] = None
) -> bool:
    """检查是否允许从源状态转换到目标状态"""
    query = select(StateTransition).where(
        StateTransition.project_id == project_id,
        StateTransition.source_state_id == source_state_id,
        StateTransition.target_state_id == target_state_id,
        StateTransition.is_deleted == False,
        StateTransition.is_active == True
    )
    
    # 检查是否有通用的转换规则或特定类型的规则
    result = await db.execute(query)
    transitions = result.scalars().all()
    
    for t in transitions:
        if t.issue_type_id is None or t.issue_type_id == issue_type_id:
            return True
    
    return False


async def update_state_transition(
    db: AsyncSession,
    transition_id: int,
    update_data: StateTransitionUpdate
) -> StateTransition:
    """更新状态转换规则"""
    transition = await get_state_transition_by_id(db, transition_id)
    
    if update_data.name is not None:
        transition.name = update_data.name
    if update_data.description is not None:
        transition.description = update_data.description
    if update_data.is_auto is not None:
        transition.is_auto = update_data.is_auto
    if update_data.issue_type_id is not None:
        transition.issue_type_id = update_data.issue_type_id
    
    await db.commit()
    await db.refresh(transition)
    return transition


async def delete_state_transition(db: AsyncSession, transition_id: int):
    """删除状态转换规则（软删除）"""
    transition = await get_state_transition_by_id(db, transition_id)
    transition.is_deleted = True
    await db.commit()


async def create_default_transitions(
    db: AsyncSession,
    project_id: int,
    workspace_id: int,
    user_id: int,
    states: List[State]
) -> List[StateTransition]:
    """创建默认状态转换规则"""
    # 构建状态分组映射
    state_by_group = {}
    for state in states:
        if state.group not in state_by_group:
            state_by_group[state.group] = []
        state_by_group[state.group].append(state)
    
    transitions = []
    
    for default_trans in DEFAULT_STATE_TRANSITIONS:
        source_group = default_trans["source_group"]
        target_group = default_trans["target_group"]
        
        if source_group in state_by_group and target_group in state_by_group:
            # 使用第一个匹配的状态
            source_state = state_by_group[source_group][0]
            target_state = state_by_group[target_group][0]
            
            transition_data = StateTransitionCreate(
                name=default_trans["name"],
                source_state_id=source_state.id,
                target_state_id=target_state.id,
                project_id=project_id,
                is_auto=False
            )
            
            transition = await create_state_transition(
                db, transition_data, workspace_id, user_id
            )
            transitions.append(transition)
    
    return transitions


# ==================== Automation Rule Service ====================

async def create_automation_rule(
    db: AsyncSession,
    rule_data: AutomationRuleCreate,
    workspace_id: int,
    user_id: int
) -> AutomationRule:
    """创建自动化规则"""
    # 验证项目存在
    project = await db.get(Project, rule_data.project_id)
    if not project or project.is_deleted:
        raise NotFoundException("Project not found")
    
    # 创建规则
    rule = AutomationRule(
        name=rule_data.name,
        description=rule_data.description,
        is_enabled=rule_data.is_enabled,
        trigger=rule_data.trigger,
        conditions=rule_data.conditions,
        actions=rule_data.actions,
        project_id=rule_data.project_id,
        workspace_id=workspace_id,
        created_by_id=user_id
    )
    
    db.add(rule)
    await db.commit()
    await db.refresh(rule)
    return rule


async def get_automation_rule_by_id(
    db: AsyncSession,
    rule_id: int
) -> AutomationRule:
    """获取自动化规则"""
    result = await db.execute(
        select(AutomationRule).where(AutomationRule.id == rule_id)
    )
    rule = result.scalar_one_or_none()
    if not rule or rule.is_deleted:
        raise NotFoundException("Automation rule not found")
    return rule


async def get_project_automation_rules(
    db: AsyncSession,
    project_id: int,
    enabled_only: bool = False
) -> List[AutomationRule]:
    """获取项目的自动化规则"""
    query = select(AutomationRule).where(
        AutomationRule.project_id == project_id,
        AutomationRule.is_deleted == False
    )
    
    if enabled_only:
        query = query.where(AutomationRule.is_enabled == True)
    
    query = query.order_by(AutomationRule.created_at.desc())
    
    result = await db.execute(query)
    return list(result.scalars().all())


async def update_automation_rule(
    db: AsyncSession,
    rule_id: int,
    update_data: AutomationRuleUpdate
) -> AutomationRule:
    """更新自动化规则"""
    rule = await get_automation_rule_by_id(db, rule_id)
    
    if update_data.name is not None:
        rule.name = update_data.name
    if update_data.description is not None:
        rule.description = update_data.description
    if update_data.is_enabled is not None:
        rule.is_enabled = update_data.is_enabled
    if update_data.trigger is not None:
        rule.trigger = update_data.trigger
    if update_data.conditions is not None:
        rule.conditions = update_data.conditions
    if update_data.actions is not None:
        rule.actions = update_data.actions
    
    await db.commit()
    await db.refresh(rule)
    return rule


async def delete_automation_rule(db: AsyncSession, rule_id: int):
    """删除自动化规则（软删除）"""
    rule = await get_automation_rule_by_id(db, rule_id)
    rule.is_deleted = True
    await db.commit()


async def toggle_automation_rule(
    db: AsyncSession,
    rule_id: int,
    enabled: bool
) -> AutomationRule:
    """启用/禁用自动化规则"""
    rule = await get_automation_rule_by_id(db, rule_id)
    rule.is_enabled = enabled
    await db.commit()
    await db.refresh(rule)
    return rule


# ==================== Automation Execution Log Service ====================

async def create_execution_log(
    db: AsyncSession,
    rule_id: int,
    status: str,
    trigger_event: str,
    triggered_issue_id: Optional[int] = None,
    execution_details: dict = None,
    error_message: Optional[str] = None,
    execution_time_ms: Optional[int] = None
) -> AutomationExecutionLog:
    """创建执行日志"""
    log = AutomationExecutionLog(
        rule_id=rule_id,
        status=status,
        trigger_event=trigger_event,
        triggered_issue_id=triggered_issue_id,
        execution_details=execution_details or {},
        error_message=error_message,
        execution_time_ms=execution_time_ms,
        created_by_id=triggered_issue_id or 0
    )
    
    db.add(log)
    
    # 更新规则的执行统计
    rule = await get_automation_rule_by_id(db, rule_id)
    rule.execution_count += 1
    rule.last_executed_at = datetime.utcnow().isoformat()
    
    await db.commit()
    await db.refresh(log)
    return log


async def get_rule_execution_logs(
    db: AsyncSession,
    rule_id: int,
    limit: int = 50
) -> List[AutomationExecutionLog]:
    """获取规则的执行日志"""
    query = select(AutomationExecutionLog).where(
        AutomationExecutionLog.rule_id == rule_id
    ).order_by(AutomationExecutionLog.created_at.desc()).limit(limit)
    
    result = await db.execute(query)
    return list(result.scalars().all())


# ==================== Automation Template Service ====================

async def get_all_templates(
    db: AsyncSession,
    category: Optional[str] = None
) -> List[AutomationTemplate]:
    """获取所有自动化模板"""
    query = select(AutomationTemplate)
    
    if category:
        query = query.where(AutomationTemplate.category == category)
    
    query = query.order_by(AutomationTemplate.usage_count.desc())
    
    result = await db.execute(query)
    return list(result.scalars().all())


async def create_rule_from_template(
    db: AsyncSession,
    template_id: int,
    project_id: int,
    workspace_id: int,
    user_id: int,
    name_override: Optional[str] = None
) -> AutomationRule:
    """从模板创建自动化规则"""
    # 获取模板
    result = await db.execute(
        select(AutomationTemplate).where(AutomationTemplate.id == template_id)
    )
    template = result.scalar_one_or_none()
    if not template:
        raise NotFoundException("Template not found")
    
    # 更新模板使用次数
    template.usage_count += 1
    
    # 创建规则
    rule_data = AutomationRuleCreate(
        name=name_override or template.name,
        description=template.description,
        project_id=project_id,
        trigger=template.trigger,
        conditions=template.conditions,
        actions=template.actions
    )
    
    rule = await create_automation_rule(db, rule_data, workspace_id, user_id)
    return rule


# ==================== Condition Evaluation ====================

def evaluate_condition(condition: dict, issue_data: dict) -> bool:
    """评估条件是否满足"""
    field = condition.get("field")
    operator = condition.get("operator")
    expected_value = condition.get("value")
    
    # 获取实际值
    actual_value = issue_data.get(field)
    
    # 处理空值情况
    if operator == "is_empty":
        return actual_value is None or actual_value == ""
    if operator == "is_not_empty":
        return actual_value is not None and actual_value != ""
    
    # 处理比较
    if operator == "equals":
        return actual_value == expected_value
    if operator == "not_equals":
        return actual_value != expected_value
    if operator == "contains":
        return expected_value in str(actual_value)
    if operator == "not_contains":
        return expected_value not in str(actual_value)
    if operator == "in":
        return actual_value in expected_value
    if operator == "not_in":
        return actual_value not in expected_value
    if operator == "greater_than":
        return actual_value > expected_value
    if operator == "less_than":
        return actual_value < expected_value
    
    return False


def evaluate_conditions(conditions: list, issue_data: dict) -> bool:
    """评估所有条件（AND 关系）"""
    if not conditions:
        return True
    
    for condition in conditions:
        if not evaluate_condition(condition, issue_data):
            return False
    return True
