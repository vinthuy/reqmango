"""
AI Thread Model - AI 对话线程持久化模型
"""
from datetime import datetime
from uuid import UUID
from sqlalchemy import String, ForeignKey, DateTime, Text, JSON, Boolean
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin


class AIThread(Base, AuditMixin):
    """AI 对话线程"""
    __tablename__ = "ai_threads"
    
    title: Mapped[str] = mapped_column(String(255), nullable=False, default="New Conversation")
    summary: Mapped[str | None] = mapped_column(Text, nullable=True)
    
    workspace_id: Mapped[UUID] = mapped_column(ForeignKey("workspaces.id"), nullable=False)
    project_id: Mapped[UUID | None] = mapped_column(ForeignKey("projects.id"), nullable=True)
    user_id: Mapped[UUID] = mapped_column(ForeignKey("users.id"), nullable=False)
    
    is_archived: Mapped[bool] = mapped_column(Boolean, default=False)
    
    workspace: Mapped["Workspace"] = relationship()
    project: Mapped["Project | None"] = relationship()
    user: Mapped["User"] = relationship()
    messages: Mapped[list["AIMessage"]] = relationship(back_populates="thread", cascade="all, delete-orphan")


class AIMessage(Base, AuditMixin):
    """AI 对话消息"""
    __tablename__ = "ai_messages"
    
    thread_id: Mapped[UUID] = mapped_column(ForeignKey("ai_threads.id"), nullable=False)
    
    role: Mapped[str] = mapped_column(String(20), nullable=False)  # user, assistant, system
    content: Mapped[str] = mapped_column(Text, nullable=False)
    
    mode: Mapped[str] = mapped_column(String(20), default="ask")  # ask, build
    intent: Mapped[str | None] = mapped_column(String(20), nullable=True)  # search, create, update, analyze, help
    
    # 上下文信息
    context_snapshot: Mapped[JSON | None] = mapped_column(JSON, nullable=True)
    
    # AI 响应信息
    plan: Mapped[JSON | None] = mapped_column(JSON, nullable=True)
    results: Mapped[JSON | None] = mapped_column(JSON, nullable=True)
    suggestions: Mapped[JSON | None] = mapped_column(JSON, nullable=True)
    
    # 附件引用
    attachments: Mapped[JSON | None] = mapped_column(JSON, nullable=True)
    
    # 执行状态
    is_executed: Mapped[bool] = mapped_column(Boolean, default=False)
    executed_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
    
    thread: Mapped["AIThread"] = relationship(back_populates="messages")


class AIActionLog(Base, AuditMixin):
    """AI 操作日志 - 记录 AI 执行的操作"""
    __tablename__ = "ai_action_logs"
    
    thread_id: Mapped[UUID | None] = mapped_column(ForeignKey("ai_threads.id"), nullable=True)
    message_id: Mapped[UUID | None] = mapped_column(ForeignKey("ai_messages.id"), nullable=True)
    
    action_type: Mapped[str] = mapped_column(String(50), nullable=False)
    target_type: Mapped[str] = mapped_column(String(50), nullable=False)  # issue, project, etc.
    target_id: Mapped[UUID | None] = mapped_column(nullable=True)
    
    changes: Mapped[JSON | None] = mapped_column(JSON, nullable=True)
    description: Mapped[str | None] = mapped_column(Text, nullable=True)
    
    status: Mapped[str] = mapped_column(String(20), default="pending")  # pending, executed, failed
    error_message: Mapped[str | None] = mapped_column(Text, nullable=True)
    
    thread: Mapped["AIThread | None"] = relationship()
    message: Mapped["AIMessage | None"] = relationship()