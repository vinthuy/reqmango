"""
Workflow Schemas - 工作流相关数据验证模型
"""
from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any
from uuid import UUID
from enum import Enum

from .base import AuditSchema, SoftDeleteSchema


# ==================== Enums ====================

class TriggerTypeEnum(str, Enum):
    """触发器类型枚举"""
    # 工作项相关
    ISSUE_CREATED = "issue.created"
    ISSUE_UPDATED = "issue.updated"
    ISSUE_DELETED = "issue.deleted"
    ISSUE_ASSIGNED = "issue.assigned"
    STATE_CHANGED = "issue.state_changed"
    PRIORITY_CHANGED = "issue.priority_changed"
    DUE_SOON = "issue.due_soon"
    DUE_DATE_PASSED = "issue.due_date_passed"
    
    # 周期相关
    CYCLE_STARTED = "cycle.started"
    CYCLE_ENDED = "cycle.ended"
    
    # 评论相关
    COMMENT_ADDED = "comment.added"


class ConditionOperatorEnum(str, Enum):
    """条件操作符枚举"""
    EQUALS = "equals"
    NOT_EQUALS = "not_equals"
    CONTAINS = "contains"
    NOT_CONTAINS = "not_contains"
    IN = "in"
    NOT_IN = "not_in"
    GREATER_THAN = "greater_than"
    LESS_THAN = "less_than"
    IS_EMPTY = "is_empty"
    IS_NOT_EMPTY = "is_not_empty"


class ActionTypeEnum(str, Enum):
    """动作类型枚举"""
    # 工作项动作
    ISSUE_UPDATE = "issue.update"
    ISSUE_ASSIGN = "issue.assign"
    ISSUE_ADD_LABEL = "issue.add_label"
    ISSUE_REMOVE_LABEL = "issue.remove_label"
    ISSUE_CHANGE_STATE = "issue.change_state"
    ISSUE_SET_PRIORITY = "issue.set_priority"
    
    # 通知动作
    NOTIFICATION_CREATE = "notification.create"
    EMAIL_SEND = "email.send"
    
    # 自动化动作
    AUTOMATION_TRIGGER = "automation.trigger"


# ==================== State Transition Schema ====================

class StateTransitionBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description: Optional[str] = Field(None, max_length=500)
    is_auto: bool = Field(default=False)


class StateTransitionCreate(StateTransitionBase):
    source_state_id: UUID
    target_state_id: UUID
    issue_type_id: Optional[UUID] = None


class StateTransitionUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=255)
    description: Optional[str] = Field(None, max_length=500)
    is_auto: Optional[bool] = None
    issue_type_id: Optional[UUID] = None


class StateTransitionResponse(AuditSchema, SoftDeleteSchema, StateTransitionBase):
    id: UUID
    source_state_id: UUID
    target_state_id: UUID
    issue_type_id: Optional[UUID] = None
    project_id: UUID
    workspace_id: UUID
    
    class Config:
        from_attributes = True


# ==================== Automation Rule Schema ====================

class TriggerSchema(BaseModel):
    """触发器配置"""
    type: TriggerTypeEnum
    # 额外配置
    state_group: Optional[str] = None  # 用于 state_changed 触发器
    priority: Optional[str] = None  # 用于 priority_changed 触发器
    days_before: Optional[int] = None  # 用于 due_soon 触发器
    field: Optional[str] = None  # 用于 updated 触发器


class ConditionSchema(BaseModel):
    """条件配置"""
    field: str  # 如 "state.group", "priority", "assignee_id"
    operator: ConditionOperatorEnum
    value: Optional[Any] = None  # 比较值


class ActionSchema(BaseModel):
    """动作配置"""
    type: ActionTypeEnum
    # 工作项动作
    field: Optional[str] = None  # 如 "assignee", "priority"
    value: Optional[Any] = None
    label: Optional[str] = None  # 用于添加/移除标签
    state_id: Optional[UUID] = None  # 用于改变状态
    # 通知动作
    message: Optional[str] = None
    subject: Optional[str] = None
    recipients: Optional[List[str]] = None


class AutomationRuleBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description: Optional[str] = Field(None, max_length=500)
    is_enabled: bool = Field(default=True)


class AutomationRuleCreate(AutomationRuleBase):
    project_id: UUID
    trigger: Dict[str, Any]  # 使用 dict 而非 TriggerSchema，方便扩展
    conditions: List[Dict[str, Any]] = Field(default=[])
    actions: List[Dict[str, Any]] = Field(default=[])


class AutomationRuleUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=255)
    description: Optional[str] = Field(None, max_length=500)
    is_enabled: Optional[bool] = None
    trigger: Optional[Dict[str, Any]] = None
    conditions: Optional[List[Dict[str, Any]]] = None
    actions: Optional[List[Dict[str, Any]]] = None


class AutomationRuleResponse(AuditSchema, SoftDeleteSchema, AutomationRuleBase):
    id: UUID
    trigger: Dict[str, Any]
    conditions: List[Dict[str, Any]]
    actions: List[Dict[str, Any]]
    execution_count: int
    last_executed_at: Optional[str] = None
    project_id: UUID
    workspace_id: UUID
    
    class Config:
        from_attributes = True


class AutomationRuleLite(BaseModel):
    """轻量级自动化规则响应"""
    id: UUID
    name: str
    description: Optional[str] = None
    is_enabled: bool
    trigger: Dict[str, Any]
    execution_count: int
    
    class Config:
        from_attributes = True


# ==================== Automation Execution Log Schema ====================

class AutomationExecutionLogResponse(BaseModel):
    id: UUID
    rule_id: UUID
    status: str  # success, failed, skipped
    trigger_event: str
    triggered_issue_id: Optional[UUID] = None
    execution_details: Dict[str, Any]
    error_message: Optional[str] = None
    execution_time_ms: Optional[int] = None
    created_at: str
    
    class Config:
        from_attributes = True


# ==================== Automation Template Schema ====================

class AutomationTemplateResponse(BaseModel):
    id: UUID
    name: str
    description: Optional[str] = None
    category: str
    trigger: Dict[str, Any]
    conditions: List[Dict[str, Any]]
    actions: List[Dict[str, Any]]
    is_system: bool
    usage_count: int
    
    class Config:
        from_attributes = True


# ==================== Helper Functions ====================

def get_trigger_display_name(trigger_type: TriggerTypeEnum) -> str:
    """获取触发器类型的显示名称"""
    names = {
        TriggerTypeEnum.ISSUE_CREATED: "工作项创建时",
        TriggerTypeEnum.ISSUE_UPDATED: "工作项更新时",
        TriggerTypeEnum.ISSUE_DELETED: "工作项删除时",
        TriggerTypeEnum.ISSUE_ASSIGNED: "工作项分配时",
        TriggerTypeEnum.STATE_CHANGED: "状态变更时",
        TriggerTypeEnum.PRIORITY_CHANGED: "优先级变更时",
        TriggerTypeEnum.DUE_SOON: "截止日期临近时",
        TriggerTypeEnum.DUE_DATE_PASSED: "截止日期过期时",
        TriggerTypeEnum.CYCLE_STARTED: "周期开始时",
        TriggerTypeEnum.CYCLE_ENDED: "周期结束时",
        TriggerTypeEnum.COMMENT_ADDED: "添加评论时",
    }
    return names.get(trigger_type, trigger_type)


def get_action_display_name(action_type: ActionTypeEnum) -> str:
    """获取动作类型的显示名称"""
    names = {
        ActionTypeEnum.ISSUE_UPDATE: "更新工作项",
        ActionTypeEnum.ISSUE_ASSIGN: "分配工作项",
        ActionTypeEnum.ISSUE_ADD_LABEL: "添加标签",
        ActionTypeEnum.ISSUE_REMOVE_LABEL: "移除标签",
        ActionTypeEnum.ISSUE_CHANGE_STATE: "改变状态",
        ActionTypeEnum.ISSUE_SET_PRIORITY: "设置优先级",
        ActionTypeEnum.NOTIFICATION_CREATE: "创建通知",
        ActionTypeEnum.EMAIL_SEND: "发送邮件",
        ActionTypeEnum.AUTOMATION_TRIGGER: "触发自动化",
    }
    return names.get(action_type, action_type)


# ==================== Default Workflow Transitions ====================

# 默认状态转换规则：允许大部分状态之间的转换
DEFAULT_STATE_TRANSITIONS = [
    # Backlog 可以转到任何状态
    {"name": "Backlog → Todo", "source_group": "backlog", "target_group": "todo"},
    # Todo 可以转到进行中或取消
    {"name": "Todo → In Progress", "source_group": "todo", "target_group": "in_progress"},
    {"name": "Todo → Cancelled", "source_group": "todo", "target_group": "cancelled"},
    # In Progress 可以转到完成、Review 或回到 Todo
    {"name": "In Progress → In Review", "source_group": "in_progress", "target_group": "in_progress"},
    {"name": "In Progress → Done", "source_group": "in_progress", "target_group": "done"},
    {"name": "In Progress → Cancelled", "source_group": "in_progress", "target_group": "cancelled"},
    # In Review 可以转到完成或回到进行中
    {"name": "In Review → Done", "source_group": "in_progress", "target_group": "done"},
    {"name": "In Review → In Progress", "source_group": "in_progress", "target_group": "in_progress"},
    # Done 和 Cancelled 是终态，可以重新打开
    {"name": "Done → Todo", "source_group": "done", "target_group": "todo"},
    {"name": "Cancelled → Todo", "source_group": "cancelled", "target_group": "todo"},
]
