"""
Custom Field Schemas - 自定义字段数据验证模型
"""
from pydantic import BaseModel, Field, field_validator
from typing import Optional, List, Dict, Any, Union
from datetime import date, datetime
from enum import Enum

from .base import AuditSchema, SoftDeleteSchema


class CustomFieldTypeEnum(str, Enum):
    """自定义字段类型"""
    TEXT = "text"
    NUMBER = "number"
    DROPDOWN = "dropdown"
    BOOLEAN = "boolean"
    DATE = "date"
    MEMBER = "member"
    URL = "url"


class TextTypeEnum(str, Enum):
    """文本类型"""
    SINGLE = "single"       # 单行文本
    PARAGRAPH = "paragraph"  # 多行文本
    READONLY = "readonly"   # 只读文本


# ==================== CustomFieldOption Schema ====================

class CustomFieldOptionBase(BaseModel):
    value: str = Field(..., min_length=1, max_length=255)
    color: Optional[str] = None
    sequence: int = 0
    is_default: bool = False
    is_active: bool = True


class CustomFieldOptionCreate(CustomFieldOptionBase):
    pass


class CustomFieldOptionUpdate(BaseModel):
    value: Optional[str] = Field(None, min_length=1, max_length=255)
    color: Optional[str] = None
    sequence: Optional[int] = None
    is_default: Optional[bool] = None
    is_active: Optional[bool] = None


class CustomFieldOptionResponse(CustomFieldOptionBase):
    id: int

    class Config:
        from_attributes = True


# ==================== CustomField Schema ====================

class CustomFieldBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description: Optional[str] = Field(None, max_length=500)
    field_type: CustomFieldTypeEnum
    is_required: bool = False
    is_readonly: bool = False
    is_active: bool = True
    
    # 文本类型属性
    text_type: Optional[TextTypeEnum] = None
    placeholder: Optional[str] = Field(None, max_length=255)
    
    # 数字类型属性
    number_default: Optional[float] = None
    number_min: Optional[float] = None
    number_max: Optional[float] = None
    
    # 下拉类型属性
    is_multi_select: bool = False
    
    # 日期类型属性
    date_format: Optional[str] = None
    
    # 序列号
    sequence: int = 0


class CustomFieldCreate(CustomFieldBase):
    project_id: Optional[int] = None
    issue_type_id: Optional[int] = None
    options: Optional[List[CustomFieldOptionCreate]] = []


class CustomFieldUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=255)
    description: Optional[str] = Field(None, max_length=500)
    is_required: Optional[bool] = None
    is_readonly: Optional[bool] = None
    is_active: Optional[bool] = None
    
    # 文本类型属性
    text_type: Optional[TextTypeEnum] = None
    placeholder: Optional[str] = Field(None, max_length=255)
    
    # 数字类型属性
    number_default: Optional[float] = None
    number_min: Optional[float] = None
    number_max: Optional[float] = None
    
    # 下拉类型属性
    is_multi_select: Optional[bool] = None
    
    # 日期类型属性
    date_format: Optional[str] = None
    
    # 序列号
    sequence: Optional[int] = None


class CustomFieldResponse(AuditSchema, SoftDeleteSchema, CustomFieldBase):
    id: int
    workspace_id: int
    project_id: Optional[int] = None
    issue_type_id: Optional[int] = None
    options: List[CustomFieldOptionResponse] = []

    class Config:
        from_attributes = True


class CustomFieldLite(BaseModel):
    """轻量级自定义字段响应"""
    id: int
    name: str
    field_type: str
    is_required: bool
    is_readonly: bool
    options: List[CustomFieldOptionResponse] = []

    class Config:
        from_attributes = True


# ==================== IssueCustomFieldValue Schema ====================

class IssueCustomFieldValueBase(BaseModel):
    field_id: int
    # 值可以是多种类型，根据字段类型使用不同的值字段
    text_value: Optional[str] = None
    number_value: Optional[float] = None
    boolean_value: Optional[bool] = None
    date_value: Optional[date] = None
    url_value: Optional[str] = Field(None, max_length=2000)
    # 下拉和成员类型使用 JSON 数组存储选项ID
    json_value: Optional[List[str]] = None


class IssueCustomFieldValueCreate(IssueCustomFieldValueBase):
    issue_id: int


class IssueCustomFieldValueUpdate(BaseModel):
    """只更新值，不更新字段ID"""
    text_value: Optional[str] = None
    number_value: Optional[float] = None
    boolean_value: Optional[bool] = None
    date_value: Optional[date] = None
    url_value: Optional[str] = Field(None, max_length=2000)
    json_value: Optional[List[str]] = None


class IssueCustomFieldValueResponse(BaseModel):
    id: int
    issue_id: int
    field_id: int
    field_name: str
    field_type: str
    # 返回统一格式的值
    value: Optional[Any] = None  # 统一返回处理后的值
    # 原始值也返回，方便前端处理
    text_value: Optional[str] = None
    number_value: Optional[float] = None
    boolean_value: Optional[bool] = None
    date_value: Optional[date] = None
    url_value: Optional[str] = None
    json_value: Optional[List[str]] = None

    class Config:
        from_attributes = True


class BulkCustomFieldValueUpdate(BaseModel):
    """批量更新工作项的自定义字段值"""
    issue_id: int
    values: List[IssueCustomFieldValueUpdate]


# ==================== Custom Field with Values Schema ====================

class CustomFieldWithValues(BaseModel):
    """自定义字段定义及其值"""
    field: CustomFieldLite
    value: Optional[Any] = None


class IssueCustomFieldsResponse(BaseModel):
    """工作项的所有自定义字段及其值"""
    issue_id: int
    fields: List[CustomFieldWithValues]


# ==================== Custom Field Validation ====================

class CustomFieldValidationRequest(BaseModel):
    """验证自定义字段值"""
    field_id: int
    value: Any


class CustomFieldValidationResponse(BaseModel):
    """验证结果"""
    is_valid: bool
    errors: List[str] = []
    field_id: int
