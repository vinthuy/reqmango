"""
Workflow Models - 工作流相关数据模型
包含：状态转换规则、自动化规则、触发器、条件、动作
"""
import uuid
from uuid import UUID
from sqlalchemy import String, Integer, ForeignKey, Boolean, Text, JSON
from sqlalchemy.dialects.postgresql import UUID as PGUUID
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin


# ==================== State Transition (状态转换规则) ====================

class StateTransition(Base, AuditMixin, SoftDeleteMixin):
    """
    状态转换规则
    定义哪些状态可以转换到哪些状态
    """
    __tablename__ = "state_transitions"

    name: Mapped[str] = mapped_column(String(255), nullable=False)
    description: Mapped[str | None] = mapped_column(String(500), nullable=True)
    
    # 源状态和目标状态
    source_state_id: Mapped[UUID] = mapped_column(ForeignKey("states.id"), nullable=False)
    target_state_id: Mapped[UUID] = mapped_column(ForeignKey("states.id"), nullable=False)
    
    # 关联的工作项类型（可选，为空表示适用于所有类型）
    issue_type_id: Mapped[UUID | None] = mapped_column(ForeignKey("issue_types.id"), nullable=True)
    
    # 是否自动应用（当工作项满足条件时）
    is_auto: Mapped[bool] = mapped_column(Boolean, default=False)
    
    project_id: Mapped[UUID] = mapped_column(ForeignKey("projects.id"), nullable=False)
    workspace_id: Mapped[UUID] = mapped_column(ForeignKey("workspaces.id"), nullable=False)
    
    source_state: Mapped["State"] = relationship(
        back_populates="target_transitions",
        foreign_keys=[source_state_id]
    )
    target_state: Mapped["State"] = relationship(
        back_populates="source_transitions",
        foreign_keys=[target_state_id]
    )
    issue_type: Mapped["IssueType | None"] = relationship()
    project: Mapped["Project"] = relationship()


# ==================== Automation (自动化规则) ====================

class AutomationRule(Base, AuditMixin, SoftDeleteMixin):
    """
    自动化规则
    包含触发器、条件和动作
    """
    __tablename__ = "automation_rules"

    name: Mapped[str] = mapped_column(String(255), nullable=False)
    description: Mapped[str | None] = mapped_column(String(500), nullable=True)
    
    # 是否启用
    is_enabled: Mapped[bool] = mapped_column(Boolean, default=True)
    
    # 触发器配置 (JSON)
    trigger: Mapped[dict] = mapped_column(JSON, nullable=False)
    
    # 条件配置 (JSON数组)
    conditions: Mapped[list] = mapped_column(JSON, default=[])
    
    # 动作配置 (JSON数组)
    actions: Mapped[list] = mapped_column(JSON, default=[])
    
    # 执行次数统计
    execution_count: Mapped[int] = mapped_column(Integer, default=0)
    last_executed_at: Mapped[str | None] = mapped_column(String(50), nullable=True)
    
    project_id: Mapped[UUID] = mapped_column(ForeignKey("projects.id"), nullable=False)
    workspace_id: Mapped[UUID] = mapped_column(ForeignKey("workspaces.id"), nullable=False)
    
    project: Mapped["Project"] = relationship()
    execution_logs: Mapped[list["AutomationExecutionLog"]] = relationship(
        back_populates="rule", cascade="all, delete-orphan"
    )


class AutomationExecutionLog(Base, AuditMixin):
    """
    自动化执行日志
    记录每次自动化规则的执行情况
    """
    __tablename__ = "automation_execution_logs"

    rule_id: Mapped[UUID] = mapped_column(ForeignKey("automation_rules.id"), nullable=False)
    
    # 执行状态
    status: Mapped[str] = mapped_column(String(20), nullable=False)  # success, failed, skipped
    
    # 触发的事件类型
    trigger_event: Mapped[str] = mapped_column(String(50), nullable=False)
    
    # 触发的工作项ID
    triggered_issue_id: Mapped[UUID | None] = mapped_column(ForeignKey("issues.id"), nullable=True)
    
    # 执行详情 (JSON)
    execution_details: Mapped[dict] = mapped_column(JSON, default={})
    
    # 错误信息
    error_message: Mapped[str | None] = mapped_column(Text, nullable=True)
    
    # 执行时间
    execution_time_ms: Mapped[int | None] = mapped_column(Integer, nullable=True)
    
    rule: Mapped["AutomationRule"] = relationship(back_populates="execution_logs")
    triggered_issue: Mapped["Issue | None"] = relationship()


# ==================== Automation Templates (自动化模板) ====================

class AutomationTemplate(Base):
    """
    自动化模板
    预定义的自动化规则模板，方便快速创建
    """
    __tablename__ = "automation_templates"

    id: Mapped[UUID] = mapped_column(PGUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    description: Mapped[str | None] = mapped_column(String(500), nullable=True)
    
    # 模板分类
    category: Mapped[str] = mapped_column(String(50), nullable=False)  # workflow, notification, etc.
    
    # 模板配置 (JSON)
    trigger: Mapped[dict] = mapped_column(JSON, nullable=False)
    conditions: Mapped[list] = mapped_column(JSON, default=[])
    actions: Mapped[list] = mapped_column(JSON, default=[])
    
    # 是否为系统内置模板
    is_system: Mapped[bool] = mapped_column(Boolean, default=False)
    
    # 使用次数统计
    usage_count: Mapped[int] = mapped_column(Integer, default=0)


# ==================== Default Automation Templates ====================

DEFAULT_AUTOMATION_TEMPLATES = [
    {
        "name": "自动分配负责人",
        "description": "当创建工作项时，自动分配给项目负责人",
        "category": "workflow",
        "trigger": {"type": "issue.created"},
        "conditions": [],
        "actions": [
            {"type": "issue.assign", "field": "assignee", "value": "{{project.default_assignee}}"}
        ]
    },
    {
        "name": "状态同步",
        "description": "当工作项标记为完成时，通知相关人员",
        "category": "notification",
        "trigger": {"type": "issue.state_changed", "state_group": "done"},
        "conditions": [],
        "actions": [
            {"type": "notification.create", "message": "工作项 {{issue.name}} 已完成"}
        ]
    },
    {
        "name": "截止日期提醒",
        "description": "当工作项截止日期临近时发送提醒",
        "category": "notification",
        "trigger": {"type": "issue.due_soon", "days_before": 1},
        "conditions": [
            {"field": "state.group", "operator": "not_in", "value": ["done", "cancelled"]}
        ],
        "actions": [
            {"type": "notification.create", "message": "工作项 {{issue.name}} 即将到期"}
        ]
    },
    {
        "name": "自动添加标签",
        "description": "当工作项优先级为紧急时，自动添加紧急标签",
        "category": "workflow",
        "trigger": {"type": "issue.priority_changed", "priority": "urgent"},
        "conditions": [],
        "actions": [
            {"type": "issue.add_label", "label": "紧急"}
        ]
    }
]
