"""
Custom Field Models - 自定义字段数据模型
支持的工作项自定义字段类型：
- Text (文本)
- Number (数字)
- Dropdown (下拉)
- Boolean (布尔)
- Date (日期)
- Member (成员选择)
- URL (链接)
"""
from datetime import datetime, date
from sqlalchemy import String, ForeignKey, Boolean, Integer, Text, JSON, Float, Date, BigInteger
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base, AuditMixin, SoftDeleteMixin


class CustomFieldType(str):
    """自定义字段类型枚举"""
    TEXT = "text"
    NUMBER = "number"
    DROPDOWN = "dropdown"
    BOOLEAN = "boolean"
    DATE = "date"
    MEMBER = "member"
    URL = "url"


class CustomField(Base, AuditMixin, SoftDeleteMixin):
    """
    自定义字段定义
    定义一个可附加到工作项类型的自定义字段
    """
    __tablename__ = "custom_fields"

    name: Mapped[str] = mapped_column(String(255), nullable=False)
    description: Mapped[str | None] = mapped_column(String(500), nullable=True)
    
    # 字段类型
    field_type: Mapped[str] = mapped_column(String(50), nullable=False)  # text, number, dropdown, boolean, date, member, url
    
    # 字段属性
    is_required: Mapped[bool] = mapped_column(Boolean, default=False)
    is_readonly: Mapped[bool] = mapped_column(Boolean, default=False)
    is_active: Mapped[bool] = mapped_column(Boolean, default=True)
    
    # 文本类型属性
    text_type: Mapped[str | None] = mapped_column(String(20), nullable=True)  # single, paragraph, readonly
    placeholder: Mapped[str | None] = mapped_column(String(255), nullable=True)
    
    # 数字类型属性
    number_default: Mapped[float | None] = mapped_column(Float, nullable=True)
    number_min: Mapped[float | None] = mapped_column(Float, nullable=True)
    number_max: Mapped[float | None] = mapped_column(Float, nullable=True)
    
    # 下拉类型属性
    is_multi_select: Mapped[bool] = mapped_column(Boolean, default=False)
    
    # 日期类型属性
    date_format: Mapped[str | None] = mapped_column(String(20), nullable=True)
    
    # 序列号用于排序
    sequence: Mapped[int] = mapped_column(Integer, default=0)
    
    # 关联关系
    workspace_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("workspaces.id"), nullable=False)
    project_id: Mapped[int | None] = mapped_column(BigInteger, ForeignKey("projects.id"), nullable=True)  # None 表示项目级
    issue_type_id: Mapped[int | None] = mapped_column(BigInteger, ForeignKey("issue_types.id"), nullable=True)  # 关联到工作项类型
    
    workspace: Mapped["Workspace"] = relationship()
    project: Mapped["Project | None"] = relationship()
    issue_type: Mapped["IssueType | None"] = relationship()
    options: Mapped[list["CustomFieldOption"]] = relationship(back_populates="field", cascade="all, delete-orphan")
    values: Mapped[list["IssueCustomFieldValue"]] = relationship(back_populates="field", cascade="all, delete-orphan")


class CustomFieldOption(Base, AuditMixin):
    """
    自定义字段选项（用于下拉类型）
    """
    __tablename__ = "custom_field_options"

    field_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("custom_fields.id"), nullable=False)
    value: Mapped[str] = mapped_column(String(255), nullable=False)
    color: Mapped[str | None] = mapped_column(String(20), nullable=True)
    sequence: Mapped[int] = mapped_column(Integer, default=0)
    
    is_default: Mapped[bool] = mapped_column(Boolean, default=False)
    is_active: Mapped[bool] = mapped_column(Boolean, default=True)
    
    field: Mapped["CustomField"] = relationship(back_populates="options")


class IssueCustomFieldValue(Base, AuditMixin):
    """
    工作项的自定义字段值
    存储每个工作项的自定义字段具体值
    """
    __tablename__ = "issue_custom_field_values"

    issue_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("issues.id"), nullable=False)
    field_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("custom_fields.id"), nullable=False)
    
    # 根据字段类型存储不同的值
    text_value: Mapped[str | None] = mapped_column(Text, nullable=True)
    number_value: Mapped[float | None] = mapped_column(Float, nullable=True)
    boolean_value: Mapped[bool | None] = mapped_column(Boolean, nullable=True)
    date_value: Mapped[date | None] = mapped_column(Date, nullable=True)
    url_value: Mapped[str | None] = mapped_column(String(2000), nullable=True)
    # 下拉和成员类型存储为 JSON 数组
    json_value: Mapped[JSON | None] = mapped_column(JSON, nullable=True)
    
    issue: Mapped["Issue"] = relationship(back_populates="custom_field_values")
    field: Mapped["CustomField"] = relationship(back_populates="values")


class CustomFieldValueHistory(Base, AuditMixin):
    """
    自定义字段值变更历史
    """
    __tablename__ = "custom_field_value_history"

    issue_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("issues.id"), nullable=False)
    field_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("custom_fields.id"), nullable=False)
    
    old_value: Mapped[str | None] = mapped_column(Text, nullable=True)
    new_value: Mapped[str | None] = mapped_column(Text, nullable=True)
    
    actor_id: Mapped[int | None] = mapped_column(BigInteger, ForeignKey("users.id"), nullable=True)
    
    issue: Mapped["Issue"] = relationship()
    field: Mapped["CustomField"] = relationship()
    actor: Mapped["User | None"] = relationship()