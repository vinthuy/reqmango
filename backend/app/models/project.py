from datetime import datetime
from uuid import UUID
from sqlalchemy import String, Boolean, ForeignKey, DateTime, Integer
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin

class Project(Base, AuditMixin, SoftDeleteMixin):
    __tablename__ = "projects"
    
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    identifier: Mapped[str] = mapped_column(String(10), nullable=False)
    description: Mapped[str | None] = mapped_column(String(1000), nullable=True)
    is_public: Mapped[bool] = mapped_column(Boolean, default=False)
    timezone: Mapped[str] = mapped_column(String(255), default="UTC")
    archived_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
    
    workspace_id: Mapped[UUID] = mapped_column(ForeignKey("workspaces.id"), nullable=False)
    default_assignee_id: Mapped[UUID | None] = mapped_column(ForeignKey("users.id"), nullable=True)
    
    workspace: Mapped["Workspace"] = relationship(back_populates="projects")
    members: Mapped[list["ProjectMember"]] = relationship(back_populates="project")
    issues: Mapped[list["Issue"]] = relationship(back_populates="project")
    states: Mapped[list["State"]] = relationship(back_populates="project")
    cycles: Mapped[list["Cycle"]] = relationship(back_populates="project")
    labels: Mapped[list["Label"]] = relationship(back_populates="project")
    modules: Mapped[list["Module"]] = relationship(back_populates="project")
    attachments: Mapped[list["Attachment"]] = relationship(back_populates="project")
    notifications: Mapped[list["Notification"]] = relationship(back_populates="project")
    estimate_points: Mapped[list["EstimatePoint"]] = relationship(back_populates="project")
    issue_types: Mapped[list["IssueType"]] = relationship(back_populates="project")

class ProjectMember(Base, AuditMixin):
    __tablename__ = "project_members"
    
    project_id: Mapped[UUID] = mapped_column(ForeignKey("projects.id"), nullable=False)
    user_id: Mapped[UUID] = mapped_column(ForeignKey("users.id"), nullable=False)
    role: Mapped[int] = mapped_column(Integer, default=15)
    is_active: Mapped[bool] = mapped_column(Boolean, default=True)
    
    project: Mapped["Project"] = relationship(back_populates="members")
    user: Mapped["User"] = relationship(back_populates="projects")