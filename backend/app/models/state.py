from sqlalchemy import String, Integer, ForeignKey, Boolean, BigInteger
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin

class State(Base, AuditMixin, SoftDeleteMixin):
    __tablename__ = "states"
    
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    color: Mapped[str] = mapped_column(String(50), default="#6B7280")
    group: Mapped[str] = mapped_column(String(50), default="backlog")
    sequence: Mapped[int] = mapped_column(Integer, default=1)
    is_default: Mapped[bool] = mapped_column(Boolean, default=False)
    is_active: Mapped[bool] = mapped_column(Boolean, default=True)
    
    project_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("projects.id"), nullable=False)
    workspace_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("workspaces.id"), nullable=False)
    
    project: Mapped["Project"] = relationship(back_populates="states")
    workspace: Mapped["Workspace"] = relationship()
    issues: Mapped[list["Issue"]] = relationship(back_populates="state")
    
    # 工作流转换关系
    source_transitions: Mapped[list["StateTransition"]] = relationship(
        back_populates="source_state",
        foreign_keys="StateTransition.source_state_id"
    )
    target_transitions: Mapped[list["StateTransition"]] = relationship(
        back_populates="target_state",
        foreign_keys="StateTransition.target_state_id"
    )