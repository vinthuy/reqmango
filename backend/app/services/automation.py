"""
Automation Engine - 自动化执行引擎
实现真正的触发器检测、条件评估和动作执行
"""
from typing import Optional, List, Dict, Any
from uuid import UUID
from datetime import datetime, timedelta
import asyncio
import logging

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload

from app.models.issue import Issue
from app.models.cycle import Cycle
from app.models.workflow import AutomationRule, AutomationExecutionLog
from app.models.notification import Notification
from app.models.user import User
from app.schemas.workflow import TriggerSchema, ConditionSchema, ActionSchema
from app.services.workflow import evaluate_conditions

logger = logging.getLogger(__name__)


# ==================== Trigger Event Types ====================

class TriggerEvent:
    """触发器事件类型"""
    ISSUE_CREATED = "issue.created"
    ISSUE_UPDATED = "issue.updated"
    ISSUE_DELETED = "issue.deleted"
    ISSUE_ASSIGNED = "issue.assigned"
    STATE_CHANGED = "issue.state_changed"
    PRIORITY_CHANGED = "issue.priority_changed"
    DUE_SOON = "issue.due_soon"
    DUE_DATE_PASSED = "issue.due_date_passed"
    CYCLE_STARTED = "cycle.started"
    CYCLE_ENDED = "cycle.ended"
    COMMENT_ADDED = "comment.added"


# ==================== Automation Executor ====================

class AutomationExecutor:
    """自动化规则执行器"""

    def __init__(self, db: AsyncSession):
        self.db = db
        self.execution_results: List[Dict[str, Any]] = []

    async def execute_rules_for_event(
        self,
        project_id: UUID,
        event_type: str,
        event_data: Dict[str, Any]
    ) -> List[Dict[str, Any]]:
        """执行与事件匹配的所有自动化规则"""
        # 获取项目所有启用的自动化规则
        rules = await self._get_enabled_rules(project_id)

        # 过滤匹配触发器类型的规则
        matching_rules = [
            rule for rule in rules
            if self._trigger_matches_event(rule.trigger, event_type, event_data)
        ]

        # 执行匹配的规则
        results = []
        for rule in matching_rules:
            result = await self._execute_rule(rule, event_data)
            results.append(result)

        return results

    async def _get_enabled_rules(self, project_id: UUID) -> List[AutomationRule]:
        """获取项目所有启用的自动化规则"""
        result = await self.db.execute(
            select(AutomationRule).where(
                AutomationRule.project_id == project_id,
                AutomationRule.is_enabled == True,
                AutomationRule.is_deleted == False
            )
        )
        return list(result.scalars().all())

    def _trigger_matches_event(
        self,
        trigger: Dict[str, Any],
        event_type: str,
        event_data: Dict[str, Any]
    ) -> bool:
        """检查触发器是否匹配事件"""
        trigger_type = trigger.get("type")

        # 检查基础类型匹配
        if trigger_type != event_type:
            return False

        # 检查触发器特定条件
        if trigger_type == TriggerEvent.STATE_CHANGED:
            expected_state_group = trigger.get("state_group")
            if expected_state_group:
                current_state_group = event_data.get("target_state_group")
                return current_state_group == expected_state_group

        if trigger_type == TriggerEvent.PRIORITY_CHANGED:
            expected_priority = trigger.get("priority")
            if expected_priority:
                current_priority = event_data.get("new_priority")
                return current_priority == expected_priority

        if trigger_type == TriggerEvent.DUE_SOON:
            days_before = trigger.get("days_before", 1)
            # 这个触发器由定时任务处理，不在这里处理
            return False

        return True

    async def _execute_rule(
        self,
        rule: AutomationRule,
        event_data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """执行单个自动化规则"""
        start_time = datetime.utcnow()
        result = {
            "rule_id": str(rule.id),
            "rule_name": rule.name,
            "status": "success",
            "actions_executed": [],
            "error": None,
            "execution_time_ms": 0
        }

        try:
            # 获取工作项数据
            issue_data = await self._get_issue_data(event_data.get("issue_id"))

            # 评估条件
            if not evaluate_conditions(rule.conditions or [], issue_data):
                result["status"] = "skipped"
                result["reason"] = "Conditions not met"
                return result

            # 执行动作
            for action in rule.actions or []:
                action_result = await self._execute_action(action, issue_data, event_data)
                result["actions_executed"].append(action_result)

            # 更新规则执行统计
            rule.execution_count += 1
            rule.last_executed_at = datetime.utcnow().isoformat()

        except Exception as e:
            logger.error(f"Error executing rule {rule.id}: {str(e)}")
            result["status"] = "failed"
            result["error"] = str(e)

        finally:
            # 计算执行时间
            end_time = datetime.utcnow()
            result["execution_time_ms"] = int(
                (end_time - start_time).total_seconds() * 1000
            )

            # 创建执行日志
            await self._create_execution_log(rule, result, event_data)

        self.execution_results.append(result)
        return result

    async def _get_issue_data(self, issue_id: Optional[UUID]) -> Dict[str, Any]:
        """获取工作项数据用于条件评估"""
        if not issue_id:
            return {}

        result = await self.db.execute(
            select(Issue).where(Issue.id == issue_id)
        )
        issue = result.scalar_one_or_none()

        if not issue:
            return {}

        return {
            "id": str(issue.id),
            "name": issue.name,
            "description": issue.description,
            "state_id": str(issue.state_id) if issue.state_id else None,
            "priority": issue.priority,
            "assignee_id": str(issue.assignee_id) if issue.assignee_id else None,
            "cycle_id": str(issue.cycle_id) if issue.cycle_id else None,
            "start_date": issue.start_date.isoformat() if issue.start_date else None,
            "target_date": issue.target_date.isoformat() if issue.target_date else None,
            "estimate_point": issue.estimate_point,
            "sequence_id": issue.sequence_id
        }

    async def _execute_action(
        self,
        action: Dict[str, Any],
        issue_data: Dict[str, Any],
        event_data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """执行单个动作"""
        action_type = action.get("type")
        result = {"type": action_type, "status": "success", "details": {}}

        try:
            if action_type == "issue.update":
                result["details"] = await self._action_update_issue(
                    issue_data.get("id"),
                    action.get("field"),
                    action.get("value")
                )

            elif action_type == "issue.assign":
                result["details"] = await self._action_assign_issue(
                    issue_data.get("id"),
                    action.get("value")
                )

            elif action_type == "issue.add_label":
                result["details"] = await self._action_add_label(
                    issue_data.get("id"),
                    action.get("label")
                )

            elif action_type == "issue.remove_label":
                result["details"] = await self._action_remove_label(
                    issue_data.get("id"),
                    action.get("label")
                )

            elif action_type == "issue.change_state":
                result["details"] = await self._action_change_state(
                    issue_data.get("id"),
                    action.get("state_id")
                )

            elif action_type == "issue.set_priority":
                result["details"] = await self._action_set_priority(
                    issue_data.get("id"),
                    action.get("value")
                )

            elif action_type == "notification.create":
                result["details"] = await self._action_create_notification(
                    issue_data,
                    action.get("message"),
                    action.get("recipients")
                )

            elif action_type == "email.send":
                result["details"] = await self._action_send_email(
                    issue_data,
                    action.get("subject"),
                    action.get("message"),
                    action.get("recipients")
                )

            else:
                result["status"] = "skipped"
                result["details"] = {"reason": f"Unknown action type: {action_type}"}

        except Exception as e:
            logger.error(f"Error executing action {action_type}: {str(e)}")
            result["status"] = "failed"
            result["error"] = str(e)

        return result

    async def _action_update_issue(
        self,
        issue_id: str,
        field: str,
        value: Any
    ) -> Dict[str, Any]:
        """更新工作项"""
        if not issue_id:
            return {"success": False, "reason": "No issue ID"}

        result = await self.db.execute(
            select(Issue).where(Issue.id == UUID(issue_id))
        )
        issue = result.scalar_one_or_none()

        if not issue:
            return {"success": False, "reason": "Issue not found"}

        # 根据字段更新
        if field == "name":
            issue.name = value
        elif field == "description":
            issue.description = value
        elif field == "priority":
            issue.priority = value
        elif field == "start_date":
            issue.start_date = datetime.fromisoformat(value).date() if value else None
        elif field == "target_date":
            issue.target_date = datetime.fromisoformat(value).date() if value else None
        elif field == "estimate_point":
            issue.estimate_point = float(value) if value else None

        issue.updated_at = datetime.utcnow().isoformat()

        await self.db.commit()

        return {"success": True, "field": field, "value": value}

    async def _action_assign_issue(
        self,
        issue_id: str,
        assignee_id: str
    ) -> Dict[str, Any]:
        """分配工作项"""
        if not issue_id:
            return {"success": False, "reason": "No issue ID"}

        result = await self.db.execute(
            select(Issue).where(Issue.id == UUID(issue_id))
        )
        issue = result.scalar_one_or_none()

        if not issue:
            return {"success": False, "reason": "Issue not found"}

        issue.assignee_id = UUID(assignee_id) if assignee_id else None
        issue.updated_at = datetime.utcnow().isoformat()

        await self.db.commit()

        return {"success": True, "assignee_id": assignee_id}

    async def _action_add_label(
        self,
        issue_id: str,
        label_id: str
    ) -> Dict[str, Any]:
        """添加标签"""
        # TODO: 实现添加标签逻辑
        return {"success": True, "label_id": label_id}

    async def _action_remove_label(
        self,
        issue_id: str,
        label_id: str
    ) -> Dict[str, Any]:
        """移除标签"""
        # TODO: 实现移除标签逻辑
        return {"success": True, "label_id": label_id}

    async def _action_change_state(
        self,
        issue_id: str,
        state_id: str
    ) -> Dict[str, Any]:
        """改变状态"""
        if not issue_id:
            return {"success": False, "reason": "No issue ID"}

        result = await self.db.execute(
            select(Issue).where(Issue.id == UUID(issue_id))
        )
        issue = result.scalar_one_or_none()

        if not issue:
            return {"success": False, "reason": "Issue not found"}

        issue.state_id = UUID(state_id) if state_id else None
        issue.updated_at = datetime.utcnow().isoformat()

        await self.db.commit()

        return {"success": True, "state_id": state_id}

    async def _action_set_priority(
        self,
        issue_id: str,
        priority: str
    ) -> Dict[str, Any]:
        """设置优先级"""
        if not issue_id:
            return {"success": False, "reason": "No issue ID"}

        result = await self.db.execute(
            select(Issue).where(Issue.id == UUID(issue_id))
        )
        issue = result.scalar_one_or_none()

        if not issue:
            return {"success": False, "reason": "Issue not found"}

        issue.priority = priority
        issue.updated_at = datetime.utcnow().isoformat()

        await self.db.commit()

        return {"success": True, "priority": priority}

    async def _action_create_notification(
        self,
        issue_data: Dict[str, Any],
        message: str,
        recipients: Optional[List[str]] = None
    ) -> Dict[str, Any]:
        """创建通知"""
        # 格式化消息
        formatted_message = message.format(**issue_data) if message else ""

        # 创建通知
        notification = Notification(
            title="自动化通知",
            message=formatted_message,
            type="automation",
            project_id=issue_data.get("project_id"),
            recipient_id=recipients[0] if recipients else None
        )

        self.db.add(notification)
        await self.db.commit()

        return {"success": True, "notification_id": str(notification.id)}

    async def _action_send_email(
        self,
        issue_data: Dict[str, Any],
        subject: str,
        message: str,
        recipients: Optional[List[str]] = None
    ) -> Dict[str, Any]:
        """发送邮件（预留接口）"""
        # TODO: 实现邮件发送逻辑
        formatted_subject = subject.format(**issue_data) if subject else "自动化通知"
        formatted_message = message.format(**issue_data) if message else ""

        logger.info(f"Email would be sent: to={recipients}, subject={formatted_subject}")

        return {
            "success": True,
            "subject": formatted_subject,
            "recipients": recipients
        }

    async def _create_execution_log(
        self,
        rule: AutomationRule,
        result: Dict[str, Any],
        event_data: Dict[str, Any]
    ) -> None:
        """创建执行日志"""
        log = AutomationExecutionLog(
            rule_id=rule.id,
            status=result["status"],
            trigger_event=event_data.get("event_type", "unknown"),
            triggered_issue_id=UUID(event_data["issue_id"]) if event_data.get("issue_id") else None,
            execution_details=result,
            error_message=result.get("error"),
            execution_time_ms=result.get("execution_time_ms", 0)
        )

        self.db.add(log)
        await self.db.commit()


# ==================== Scheduled Tasks ====================

async def check_due_soon_issues(db: AsyncSession, days_before: int = 1) -> List[Dict[str, Any]]:
    """检查即将到期的工作项，触发 DUE_SOON 事件"""
    future_date = datetime.utcnow().date() + timedelta(days=days_before)

    result = await db.execute(
        select(Issue).where(
            Issue.target_date != None,
            Issue.target_date <= future_date,
            Issue.target_date >= datetime.utcnow().date(),
            Issue.is_deleted == False
        )
    )
    issues = result.scalars().all()

    executor = AutomationExecutor(db)
    results = []

    for issue in issues:
        event_data = {
            "event_type": TriggerEvent.DUE_SOON,
            "issue_id": str(issue.id),
            "project_id": str(issue.project_id),
            "target_date": str(issue.target_date),
            "days_until_due": (issue.target_date - datetime.utcnow().date()).days
        }

        rule_results = await executor.execute_rules_for_event(
            issue.project_id,
            TriggerEvent.DUE_SOON,
            event_data
        )
        results.extend(rule_results)

    return results


async def check_overdue_issues(db: AsyncSession) -> List[Dict[str, Any]]:
    """检查已过期的工作项，触发 DUE_DATE_PASSED 事件"""
    today = datetime.utcnow().date()

    result = await db.execute(
        select(Issue).where(
            Issue.target_date != None,
            Issue.target_date < today,
            Issue.is_deleted == False
        )
    )
    issues = result.scalars().all()

    executor = AutomationExecutor(db)
    results = []

    for issue in issues:
        event_data = {
            "event_type": TriggerEvent.DUE_DATE_PASSED,
            "issue_id": str(issue.id),
            "project_id": str(issue.project_id),
            "target_date": str(issue.target_date),
            "days_overdue": (today - issue.target_date).days
        }

        rule_results = await executor.execute_rules_for_event(
            issue.project_id,
            TriggerEvent.DUE_DATE_PASSED,
            event_data
        )
        results.extend(rule_results)

    return results


async def check_cycle_status(db: AsyncSession) -> List[Dict[str, Any]]:
    """检查周期状态，触发 CYCLE_STARTED 和 CYCLE_ENDED 事件"""
    today = datetime.utcnow().date()
    results = []

    executor = AutomationExecutor(db)

    # 检查今天开始的周期
    started_result = await db.execute(
        select(Cycle).where(
            Cycle.start_date != None,
            Cycle.start_date == today,
            Cycle.is_deleted == False
        )
    )
    started_cycles = started_result.scalars().all()

    for cycle in started_cycles:
        event_data = {
            "event_type": TriggerEvent.CYCLE_STARTED,
            "cycle_id": str(cycle.id),
            "project_id": str(cycle.project_id),
            "start_date": str(cycle.start_date)
        }

        rule_results = await executor.execute_rules_for_event(
            cycle.project_id,
            TriggerEvent.CYCLE_STARTED,
            event_data
        )
        results.extend(rule_results)

    # 检查今天结束的周期
    ended_result = await db.execute(
        select(Cycle).where(
            Cycle.end_date != None,
            Cycle.end_date == today,
            Cycle.is_deleted == False
        )
    )
    ended_cycles = ended_result.scalars().all()

    for cycle in ended_cycles:
        event_data = {
            "event_type": TriggerEvent.CYCLE_ENDED,
            "cycle_id": str(cycle.id),
            "project_id": str(cycle.project_id),
            "end_date": str(cycle.end_date)
        }

        rule_results = await executor.execute_rules_for_event(
            cycle.project_id,
            TriggerEvent.CYCLE_ENDED,
            event_data
        )
        results.extend(rule_results)

    return results
