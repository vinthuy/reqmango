"""
Comment Model - 评论模型
"""
from sqlalchemy import String, ForeignKey, Boolean, Text, Integer, BigInteger
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin


class Comment(Base, AuditMixin, SoftDeleteMixin):
    __tablename__ = "comments"

    content: Mapped[str] = mapped_column(Text, nullable=False)
    html_content: Mapped[str] = mapped_column(Text, nullable=True)

    # 关联
    issue_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("issues.id"), nullable=False)
    author_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("users.id"), nullable=False)
    parent_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("comments.id"), nullable=True)

    # 状态
    is_resolved: Mapped[bool] = mapped_column(Boolean, default=False)
    resolved_by_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("users.id"), nullable=True)
    resolved_at: Mapped[str] = mapped_column(String, nullable=True)

    # 表情反应计数
    reaction_count: Mapped[int] = mapped_column(Integer, default=0)

    # 关系
    issue: Mapped["Issue"] = relationship(back_populates="comments", foreign_keys=[issue_id])
    author: Mapped["User"] = relationship(back_populates="comments", foreign_keys=[author_id])
    replies: Mapped[list["Comment"]] = relationship(back_populates="parent")
    parent: Mapped["Comment | None"] = relationship(back_populates="replies", remote_side="Comment.id")