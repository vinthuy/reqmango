from sqlalchemy import String, ForeignKey, Boolean, Integer, BigInteger
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin

class IssueType(Base, AuditMixin, SoftDeleteMixin):
    __tablename__ = "issue_types"
    
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    color: Mapped[str] = mapped_column(String(50), default="#6B7280")
    icon: Mapped[str] = mapped_column(String(50), default="circle")
    is_default: Mapped[bool] = mapped_column(Boolean, default=False)
    sequence: Mapped[int] = mapped_column(Integer, default=1)
    is_active: Mapped[bool] = mapped_column(Boolean, default=True)
    
    project_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("projects.id"), nullable=False)
    workspace_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("workspaces.id"), nullable=False)
    
    project: Mapped["Project"] = relationship(back_populates="issue_types")
    workspace: Mapped["Workspace"] = relationship()