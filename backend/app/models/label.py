from uuid import UUID
from sqlalchemy import String, ForeignKey, Boolean
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin

class Label(Base, AuditMixin, SoftDeleteMixin):
    __tablename__ = "labels"
    
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    color: Mapped[str] = mapped_column(String(50), default="#6B7280")
    description: Mapped[str | None] = mapped_column(String(255), nullable=True)
    
    project_id: Mapped[UUID] = mapped_column(ForeignKey("projects.id"), nullable=False)
    
    project: Mapped["Project"] = relationship(back_populates="labels")
    issue_links: Mapped[list["IssueLabel"]] = relationship(back_populates="label")