from datetime import datetime, date
from uuid import UUID
from sqlalchemy import String, ForeignKey, Boolean, Integer, Date, DateTime
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin

class Cycle(Base, AuditMixin, SoftDeleteMixin):
    __tablename__ = "cycles"
    
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    description: Mapped[str | None] = mapped_column(String(1000), nullable=True)
    
    start_date: Mapped[date] = mapped_column(Date, nullable=False)
    end_date: Mapped[date | None] = mapped_column(Date, nullable=True)
    completed_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
    
    project_id: Mapped[UUID] = mapped_column(ForeignKey("projects.id"), nullable=False)
    workspace_id: Mapped[UUID] = mapped_column(ForeignKey("workspaces.id"), nullable=False)
    
    project: Mapped["Project"] = relationship(back_populates="cycles")
    workspace: Mapped["Workspace"] = relationship()
    issue_links: Mapped[list["IssueCycle"]] = relationship(back_populates="cycle")