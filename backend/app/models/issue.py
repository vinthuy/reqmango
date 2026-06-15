from datetime import datetime, date
from uuid import UUID
from sqlalchemy import String, Integer, ForeignKey, Date, DateTime, Boolean, Text, JSON, Float, Table, Column
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import ARRAY
from .base import Base, AuditMixin, SoftDeleteMixin


# ==================== Issue Association Tables ====================

class IssueAssignee(Base):
    """工作项负责人关联表"""
    __tablename__ = "issue_assignees"
    
    issue_id: Mapped[UUID] = mapped_column(ForeignKey("issues.id", ondelete="CASCADE"), primary_key=True)
    user_id: Mapped[UUID] = mapped_column(ForeignKey("users.id", ondelete="CASCADE"), primary_key=True)
    
    # Relationships
    issue: Mapped["Issue"] = relationship(back_populates="assignee_links")
    user: Mapped["User"] = relationship(back_populates="assigned_issues")


class IssueLabel(Base):
    """工作项标签关联表"""
    __tablename__ = "issue_labels"
    
    issue_id: Mapped[UUID] = mapped_column(ForeignKey("issues.id", ondelete="CASCADE"), primary_key=True)
    label_id: Mapped[UUID] = mapped_column(ForeignKey("labels.id", ondelete="CASCADE"), primary_key=True)
    
    # Relationships
    issue: Mapped["Issue"] = relationship(back_populates="label_links")
    label: Mapped["Label"] = relationship(back_populates="issue_links")


class IssueCycle(Base):
    """工作项周期关联表"""
    __tablename__ = "issue_cycles"
    
    issue_id: Mapped[UUID] = mapped_column(ForeignKey("issues.id", ondelete="CASCADE"), primary_key=True)
    cycle_id: Mapped[UUID] = mapped_column(ForeignKey("cycles.id", ondelete="CASCADE"), primary_key=True)
    
    # Relationships
    issue: Mapped["Issue"] = relationship(back_populates="cycle_link")
    cycle: Mapped["Cycle"] = relationship(back_populates="issue_links")


class IssueModule(Base):
    """工作项模块关联表"""
    __tablename__ = "issue_modules"
    
    issue_id: Mapped[UUID] = mapped_column(ForeignKey("issues.id", ondelete="CASCADE"), primary_key=True)
    module_id: Mapped[UUID] = mapped_column(ForeignKey("modules.id", ondelete="CASCADE"), primary_key=True)
    
    # Relationships
    issue: Mapped["Issue"] = relationship(back_populates="module_links")
    module: Mapped["Module"] = relationship(back_populates="issue_links")


# ==================== Issue Model ====================

class Issue(Base, AuditMixin, SoftDeleteMixin):
    __tablename__ = "issues"
    
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    description_html: Mapped[str] = mapped_column(Text, default="<p></p>")
    description_json: Mapped[JSON] = mapped_column(JSON, default=dict)
    description_stripped: Mapped[str | None] = mapped_column(Text, nullable=True)
    
    priority: Mapped[str] = mapped_column(String(30), default="none")
    sequence_id: Mapped[int] = mapped_column(Integer, default=1)
    sort_order: Mapped[float] = mapped_column(Float, default=65535)
    
    start_date: Mapped[date | None] = mapped_column(Date, nullable=True)
    target_date: Mapped[date | None] = mapped_column(Date, nullable=True)
    completed_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
    
    is_draft: Mapped[bool] = mapped_column(Boolean, default=False)
    archived_at: Mapped[date | None] = mapped_column(Date, nullable=True)
    
    project_id: Mapped[UUID] = mapped_column(ForeignKey("projects.id"), nullable=False)
    workspace_id: Mapped[UUID] = mapped_column(ForeignKey("workspaces.id"), nullable=False)
    parent_id: Mapped[UUID | None] = mapped_column(ForeignKey("issues.id"), nullable=True)
    state_id: Mapped[UUID] = mapped_column(ForeignKey("states.id"), nullable=False)
    
    external_id: Mapped[str | None] = mapped_column(String(255), nullable=True)
    external_source: Mapped[str | None] = mapped_column(String(255), nullable=True)
    
    # Relationships
    project: Mapped["Project"] = relationship(back_populates="issues")
    workspace: Mapped["Workspace"] = relationship()
    sub_issues: Mapped[list["Issue"]] = relationship(back_populates="parent")
    parent: Mapped["Issue | None"] = relationship(remote_side="Issue.id", back_populates="sub_issues")
    state: Mapped["State"] = relationship(back_populates="issues")
    custom_field_values: Mapped[list["IssueCustomFieldValue"]] = relationship(
        back_populates="issue", cascade="all, delete-orphan"
    )
    
    # Association relationships (use links for many-to-many)
    assignee_links: Mapped[list["IssueAssignee"]] = relationship(
        back_populates="issue", cascade="all, delete-orphan"
    )
    label_links: Mapped[list["IssueLabel"]] = relationship(
        back_populates="issue", cascade="all, delete-orphan"
    )
    cycle_link: Mapped["IssueCycle | None"] = relationship(
        back_populates="issue", cascade="all, delete-orphan", uselist=False
    )
    module_links: Mapped[list["IssueModule"]] = relationship(
        back_populates="issue", cascade="all, delete-orphan"
    )
    activities: Mapped[list["IssueActivity"]] = relationship(back_populates="issue")
    attachments: Mapped[list["Attachment"]] = relationship(back_populates="issue")
    comments: Mapped[list["Comment"]] = relationship(back_populates="issue")
    notifications: Mapped[list["Notification"]] = relationship(back_populates="issue")
    
    # Convenience accessors (computed properties)
    @property
    def assignees(self) -> list["User"]:
        """获取负责人列表"""
        return [link.user for link in self.assignee_links] if self.assignee_links else []
    
    @property
    def labels(self) -> list["Label"]:
        """获取标签列表"""
        return [link.label for link in self.label_links] if self.label_links else []
    
    @property
    def cycle(self) -> "Cycle | None":
        """获取周期"""
        return self.cycle_link.cycle if self.cycle_link else None


# ==================== Issue Activity ====================

class IssueActivity(Base, AuditMixin):
    __tablename__ = "issue_activities"
    
    issue_id: Mapped[UUID | None] = mapped_column(ForeignKey("issues.id"), nullable=True)
    verb: Mapped[str] = mapped_column(String(255), default="created")
    field: Mapped[str | None] = mapped_column(String(255), nullable=True)
    old_value: Mapped[str | None] = mapped_column(Text, nullable=True)
    new_value: Mapped[str | None] = mapped_column(Text, nullable=True)
    comment: Mapped[str | None] = mapped_column(Text, nullable=True)
    actor_id: Mapped[UUID | None] = mapped_column(ForeignKey("users.id"), nullable=True)
    
    issue: Mapped["Issue"] = relationship(back_populates="activities")