from sqlalchemy import String, ForeignKey, Integer, BigInteger
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin

class Module(Base, AuditMixin, SoftDeleteMixin):
    __tablename__ = "modules"
    
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    description: Mapped[str | None] = mapped_column(String(1000), nullable=True)
    sequence: Mapped[int] = mapped_column(Integer, default=1)
    
    project_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("projects.id"), nullable=False)
    workspace_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("workspaces.id"), nullable=False)
    parent_id: Mapped[int | None] = mapped_column(BigInteger, ForeignKey("modules.id"), nullable=True)
    
    project: Mapped["Project"] = relationship(back_populates="modules")
    workspace: Mapped["Workspace"] = relationship()
    sub_modules: Mapped[list["Module"]] = relationship(back_populates="parent")
    parent: Mapped["Module | None"] = relationship(remote_side="Module.id", back_populates="sub_modules")
    issue_links: Mapped[list["IssueModule"]] = relationship(back_populates="module")