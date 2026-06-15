from datetime import datetime, date
from uuid import UUID
from sqlalchemy import String, Boolean, Text, DateTime
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin

class User(Base, AuditMixin, SoftDeleteMixin):
    __tablename__ = "users"
    
    email: Mapped[str] = mapped_column(String(255), unique=True, nullable=False)
    username: Mapped[str] = mapped_column(String(128), unique=True, nullable=False)
    display_name: Mapped[str] = mapped_column(String(255), default="")
    first_name: Mapped[str | None] = mapped_column(String(255), nullable=True)
    last_name: Mapped[str | None] = mapped_column(String(255), nullable=True)
    avatar: Mapped[str | None] = mapped_column(Text, nullable=True)
    password_hash: Mapped[str] = mapped_column(String(255), nullable=False)
    
    is_active: Mapped[bool] = mapped_column(Boolean, default=True)
    is_superuser: Mapped[bool] = mapped_column(Boolean, default=False)
    is_email_verified: Mapped[bool] = mapped_column(Boolean, default=False)
    
    user_timezone: Mapped[str] = mapped_column(String(255), default="UTC")
    last_active: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
    
    workspaces: Mapped[list["WorkspaceMember"]] = relationship(back_populates="user")
    projects: Mapped[list["ProjectMember"]] = relationship(back_populates="user")
    assigned_issues: Mapped[list["IssueAssignee"]] = relationship(back_populates="user")
    notifications: Mapped[list["Notification"]] = relationship(back_populates="recipient", foreign_keys="Notification.recipient_id")
    sent_notifications: Mapped[list["Notification"]] = relationship(back_populates="sender", foreign_keys="Notification.sender_id")
    attachments: Mapped[list["Attachment"]] = relationship(back_populates="uploaded_by")
    comments: Mapped[list["Comment"]] = relationship(back_populates="author", foreign_keys="Comment.author_id")