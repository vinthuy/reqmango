"""
Automation Execution API - 自动化执行API端点
"""
from typing import Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, BackgroundTasks
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.session import get_db
from app.core.exceptions import NotFoundException, ValidationException
from app.services.automation import (
    AutomationExecutor,
    TriggerEvent,
    check_due_soon_issues,
    check_overdue_issues,
    check_cycle_status
)
from app.schemas.workflow import TriggerSchema, ConditionSchema, ActionSchema
from pydantic import BaseModel

router = APIRouter(prefix="/automation", tags=["automation"])


# ==================== Request/Response Models ====================

class TriggerEventRequest(BaseModel):
    """触发事件请求"""
    event_type: str
    issue_id: Optional[str] = None
    cycle_id: Optional[str] = None
    project_id: str
    data: dict = {}


class ExecuteRuleRequest(BaseModel):
    """手动执行规则请求"""
    rule_id: str


class ExecutionResult(BaseModel):
    """执行结果"""
    rule_id: str
    rule_name: str
    status: str
    actions_executed: list
    error: Optional[str] = None
    execution_time_ms: int


# ==================== Event Triggering Endpoints ====================

@router.post("/trigger", response_model=list[ExecutionResult])
async def trigger_automation_event(
    request: TriggerEventRequest,
    db: AsyncSession = Depends(get_db)
):
    """
    触发自动化事件

    当工作项、周期等发生变更时，调用此接口触发相应的自动化规则。
    """
    executor = AutomationExecutor(db)

    event_data = {
        "event_type": request.event_type,
        "issue_id": request.issue_id,
        "cycle_id": request.cycle_id,
        "project_id": request.project_id,
        **request.data
    }

    results = await executor.execute_rules_for_event(
        project_id=UUID(request.project_id),
        event_type=request.event_type,
        event_data=event_data
    )

    return results


@router.post("/trigger/issue-created/{project_id}")
async def trigger_issue_created(
    project_id: UUID,
    issue_id: UUID,
    db: AsyncSession = Depends(get_db)
):
    """触发工作项创建事件"""
    executor = AutomationExecutor(db)

    event_data = {
        "event_type": TriggerEvent.ISSUE_CREATED,
        "issue_id": str(issue_id),
        "project_id": str(project_id)
    }

    results = await executor.execute_rules_for_event(
        project_id=project_id,
        event_type=TriggerEvent.ISSUE_CREATED,
        event_data=event_data
    )

    return {"triggered": True, "results": results}


@router.post("/trigger/issue-updated/{project_id}")
async def trigger_issue_updated(
    project_id: UUID,
    issue_id: UUID,
    previous_state_id: Optional[UUID] = None,
    new_state_id: Optional[UUID] = None,
    previous_priority: Optional[str] = None,
    new_priority: Optional[str] = None,
    db: AsyncSession = Depends(get_db)
):
    """触发工作项更新事件"""
    executor = AutomationExecutor(db)
    results = []

    # 基础更新事件
    event_data = {
        "event_type": TriggerEvent.ISSUE_UPDATED,
        "issue_id": str(issue_id),
        "project_id": str(project_id)
    }

    results.extend(await executor.execute_rules_for_event(
        project_id=project_id,
        event_type=TriggerEvent.ISSUE_UPDATED,
        event_data=event_data
    ))

    # 状态变更事件
    if previous_state_id != new_state_id:
        event_data = {
            "event_type": TriggerEvent.STATE_CHANGED,
            "issue_id": str(issue_id),
            "project_id": str(project_id),
            "previous_state_id": str(previous_state_id) if previous_state_id else None,
            "target_state_id": str(new_state_id) if new_state_id else None
        }

        results.extend(await executor.execute_rules_for_event(
            project_id=project_id,
            event_type=TriggerEvent.STATE_CHANGED,
            event_data=event_data
        ))

    # 优先级变更事件
    if previous_priority != new_priority:
        event_data = {
            "event_type": TriggerEvent.PRIORITY_CHANGED,
            "issue_id": str(issue_id),
            "project_id": str(project_id),
            "old_priority": previous_priority,
            "new_priority": new_priority
        }

        results.extend(await executor.execute_rules_for_event(
            project_id=project_id,
            event_type=TriggerEvent.PRIORITY_CHANGED,
            event_data=event_data
        ))

    return {"triggered": True, "results": results}


@router.post("/trigger/issue-assigned/{project_id}")
async def trigger_issue_assigned(
    project_id: UUID,
    issue_id: UUID,
    assignee_id: UUID,
    db: AsyncSession = Depends(get_db)
):
    """触发工作项分配事件"""
    executor = AutomationExecutor(db)

    event_data = {
        "event_type": TriggerEvent.ISSUE_ASSIGNED,
        "issue_id": str(issue_id),
        "project_id": str(project_id),
        "assignee_id": str(assignee_id)
    }

    results = await executor.execute_rules_for_event(
        project_id=project_id,
        event_type=TriggerEvent.ISSUE_ASSIGNED,
        event_data=event_data
    )

    return {"triggered": True, "results": results}


@router.post("/trigger/cycle-started/{project_id}")
async def trigger_cycle_started(
    project_id: UUID,
    cycle_id: UUID,
    db: AsyncSession = Depends(get_db)
):
    """触发周期开始事件"""
    executor = AutomationExecutor(db)

    event_data = {
        "event_type": TriggerEvent.CYCLE_STARTED,
        "cycle_id": str(cycle_id),
        "project_id": str(project_id)
    }

    results = await executor.execute_rules_for_event(
        project_id=project_id,
        event_type=TriggerEvent.CYCLE_STARTED,
        event_data=event_data
    )

    return {"triggered": True, "results": results}


@router.post("/trigger/cycle-ended/{project_id}")
async def trigger_cycle_ended(
    project_id: UUID,
    cycle_id: UUID,
    db: AsyncSession = Depends(get_db)
):
    """触发周期结束事件"""
    executor = AutomationExecutor(db)

    event_data = {
        "event_type": TriggerEvent.CYCLE_ENDED,
        "cycle_id": str(cycle_id),
        "project_id": str(project_id)
    }

    results = await executor.execute_rules_for_event(
        project_id=project_id,
        event_type=TriggerEvent.CYCLE_ENDED,
        event_data=event_data
    )

    return {"triggered": True, "results": results}


# ==================== Scheduled Task Endpoints ====================

@router.post("/check/due-soon")
async def check_due_soon(
    days_before: int = 1,
    db: AsyncSession = Depends(get_db)
):
    """
    检查即将到期的工作项

    由定时任务调用，触发 DUE_SOON 事件。
    """
    results = await check_due_soon_issues(db, days_before)
    return {"checked": True, "triggered": len(results), "results": results}


@router.post("/check/overdue")
async def check_overdue(
    db: AsyncSession = Depends(get_db)
):
    """
    检查已过期的工作项

    由定时任务调用，触发 DUE_DATE_PASSED 事件。
    """
    results = await check_overdue_issues(db)
    return {"checked": True, "triggered": len(results), "results": results}


@router.post("/check/cycles")
async def check_cycles(
    db: AsyncSession = Depends(get_db)
):
    """
    检查周期状态

    由定时任务调用，触发 CYCLE_STARTED 和 CYCLE_ENDED 事件。
    """
    results = await check_cycle_status(db)
    return {"checked": True, "triggered": len(results), "results": results}


# ==================== Manual Execution Endpoints ====================

@router.post("/execute/{rule_id}")
async def execute_rule(
    rule_id: UUID,
    issue_id: Optional[UUID] = None,
    db: AsyncSession = Depends(get_db)
):
    """
    手动执行自动化规则

    用于测试或在特定场景下手动触发规则。
    """
    from app.services.workflow import get_automation_rule_by_id

    try:
        rule = await get_automation_rule_by_id(db, rule_id)
    except NotFoundException:
        raise HTTPException(status_code=404, detail="Rule not found")

    executor = AutomationExecutor(db)

    event_data = {
        "event_type": "manual",
        "issue_id": str(issue_id) if issue_id else None,
        "project_id": str(rule.project_id),
        "manual_trigger": True
    }

    results = await executor.execute_rules_for_event(
        project_id=rule.project_id,
        event_type=rule.trigger.get("type", "manual"),
        event_data=event_data
    )

    return {"executed": True, "results": results}


@router.post("/dry-run")
async def dry_run_automation(
    project_id: UUID,
    trigger: TriggerSchema,
    conditions: list[ConditionSchema],
    actions: list[ActionSchema],
    issue_id: Optional[UUID] = None,
    db: AsyncSession = Depends(get_db)
):
    """
    预演自动化规则

    测试规则在不实际执行的情况下会匹配哪些工作项。
    """
    from app.services.workflow import evaluate_conditions

    # 获取工作项数据
    issue_data = {}
    if issue_id:
        from app.models.issue import Issue
        result = await db.execute(
            select(Issue).where(Issue.id == issue_id)
        )
        issue = result.scalar_one_or_none()
        if issue:
            issue_data = {
                "id": str(issue.id),
                "name": issue.name,
                "description": issue.description,
                "state_id": str(issue.state_id) if issue.state_id else None,
                "priority": issue.priority,
                "assignee_id": str(issue.assignee_id) if issue.assignee_id else None,
                "cycle_id": str(issue.cycle_id) if issue.cycle_id else None,
                "start_date": issue.start_date.isoformat() if issue.start_date else None,
                "target_date": issue.target_date.isoformat() if issue.target_date else None
            }

    # 评估条件
    conditions_met = evaluate_conditions(conditions, issue_data)

    return {
        "would_execute": conditions_met,
        "issue_id": str(issue_id) if issue_id else None,
        "issue_data": issue_data,
        "conditions_evaluated": len(conditions),
        "actions_to_execute": len(actions) if conditions_met else 0
    }
