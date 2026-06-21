"""
Attachment Model - 附件模型
"""
from sqlalchemy import String, ForeignKey, Boolean, Integer, BigInteger
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin


class Attachment(Base, AuditMixin, SoftDeleteMixin):
    __tablename__ = "attachments"

    name: Mapped[str] = mapped_column(String(255), nullable=False)
    file_name: Mapped[str] = mapped_column(String(255), nullable=False)
    file_size: Mapped[int] = mapped_column(BigInteger, nullable=False)
    mime_type: Mapped[str] = mapped_column(String(100), nullable=False)
    file_path: Mapped[str] = mapped_column(String(500), nullable=False)
    file_url: Mapped[str] = mapped_column(String(500), nullable=True)

    # 关联
    issue_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("issues.id"), nullable=True)
    project_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("projects.id"), nullable=True)
    uploaded_by_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("users.id"), nullable=False)

    # 元数据
    is_protected: Mapped[bool] = mapped_column(Boolean, default=False)  # 是否受保护（不能公开下载）
    access_url: Mapped[str] = mapped_column(String(500), nullable=True)
    thumbnail_url: Mapped[str] = mapped_column(String(500), nullable=True)

    # 关系
    issue: Mapped["Issue"] = relationship(back_populates="attachments", foreign_keys=[issue_id])
    project: Mapped["Project"] = relationship(back_populates="attachments")
    uploaded_by: Mapped["User"] = relationship(back_populates="attachments")