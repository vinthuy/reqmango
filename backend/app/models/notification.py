"""
Notification Model - 通知模型
"""
from uuid import UUID
from sqlalchemy import String, ForeignKey, Boolean, Text, Integer
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin


class Notification(Base, AuditMixin, SoftDeleteMixin):
    __tablename__ = "notifications"

    title: Mapped[str] = mapped_column(String(255), nullable=False)
    message: Mapped[str] = mapped_column(Text, nullable=False)
    type: Mapped[str] = mapped_column(String(50), default="info")  # info, warning, error, success
    is_read: Mapped[bool] = mapped_column(Boolean, default=False)
    read_at: Mapped[str] = mapped_column(String, nullable=True)

    # 关联
    recipient_id: Mapped[UUID] = mapped_column(ForeignKey("users.id"), nullable=False)
    sender_id: Mapped[UUID] = mapped_column(ForeignKey("users.id"), nullable=True)
    project_id: Mapped[UUID] = mapped_column(ForeignKey("projects.id"), nullable=True)
    issue_id: Mapped[UUID] = mapped_column(ForeignKey("issues.id"), nullable=True)

    # 优先级
    priority: Mapped[str] = mapped_column(String(20), default="medium")  # low, medium, high, urgent

    # 跳转链接
    action_url: Mapped[str] = mapped_column(String(500), nullable=True)

    # 聚合字段
    aggregated_count: Mapped[int] = mapped_column(Integer, default=1)

    # 关系
    recipient: Mapped["User"] = relationship(back_populates="notifications", foreign_keys=[recipient_id])
    sender: Mapped["User"] = relationship(back_populates="sent_notifications", foreign_keys=[sender_id])
    project: Mapped["Project"] = relationship(back_populates="notifications")
    issue: Mapped["Issue"] = relationship(back_populates="notifications")
