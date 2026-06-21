from datetime import datetime
from sqlalchemy import DateTime, Boolean, BigInteger, Integer, func
from sqlalchemy.ext.asyncio import AsyncAttrs
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column

class Base(AsyncAttrs, DeclarativeBase):
    pass

class TimestampMixin:
    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=func.now(), nullable=False
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, default=func.now(), onupdate=func.now(), nullable=False
    )

class SoftDeleteMixin:
    deleted_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
    is_deleted: Mapped[bool] = mapped_column(Boolean, default=False)

class BigIntMixin:
    """使用 BigInteger 作为主键，与 Go 后端保持一致"""
    # 使用Integer.with_variant实现跨数据库兼容
    # SQLite: INTEGER (支持自增)
    # PostgreSQL: BIGINT
    id: Mapped[int] = mapped_column(
        Integer, primary_key=True, autoincrement=True
    )

class AuditMixin(BigIntMixin, TimestampMixin):
    created_by_id: Mapped[int | None] = mapped_column(BigInteger, nullable=True)
    updated_by_id: Mapped[int | None] = mapped_column(BigInteger, nullable=True)