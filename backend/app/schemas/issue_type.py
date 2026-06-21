"""
Issue Type Schemas - 工作项类型数据验证模型
"""
from pydantic import BaseModel, Field
from typing import Optional, List
from enum import Enum

from .base import AuditSchema, SoftDeleteSchema


class StateGroupEnum(str, Enum):
    """状态分组枚举"""
    BACKLOG = "backlog"
    TODO = "todo"
    IN_PROGRESS = "in_progress"
    DONE = "done"
    CANCELLED = "cancelled"


class IssuePriorityEnum(str, Enum):
    """工作项优先级枚举"""
    URGENT = "urgent"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    NONE = "none"


# ==================== IssueType Schema ====================

class IssueTypeBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    color: str = Field(default="#6B7280", max_length=50)
    icon: str = Field(default="circle", max_length=50)
    is_default: bool = Field(default=False)
    sequence: int = Field(default=1, ge=0)
    is_active: bool = Field(default=True)


class IssueTypeCreate(IssueTypeBase):
    """创建工作项类型"""
    project_id: int
    # 关联的自定义字段ID列表
    custom_fields: Optional[List[int]] = None


class IssueTypeUpdate(BaseModel):
    """更新工作项类型"""
    name: Optional[str] = Field(None, min_length=1, max_length=255)
    color: Optional[str] = Field(None, max_length=50)
    icon: Optional[str] = Field(None, max_length=50)
    is_default: Optional[bool] = None
    sequence: Optional[int] = Field(None, ge=0)
    is_active: Optional[bool] = None
    custom_fields: Optional[List[int]] = None


class IssueTypeResponse(AuditSchema, SoftDeleteSchema, IssueTypeBase):
    """工作项类型响应"""
    id: int
    project_id: int
    workspace_id: int  # 通过 project 获取
    
    class Config:
        from_attributes = True


class IssueTypeLite(BaseModel):
    """轻量级工作项类型响应"""
    id: int
    name: str
    color: str
    icon: str
    is_default: bool
    
    class Config:
        from_attributes = True


# ==================== State Schema ====================

class StateBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    color: str = Field(default="#6B7280", max_length=50)
    group: StateGroupEnum = Field(default=StateGroupEnum.TODO)
    sequence: int = Field(default=1, ge=0)
    is_active: bool = Field(default=True)
    description: Optional[str] = Field(None, max_length=500)


class StateCreate(StateBase):
    """创建状态"""
    project_id: int


class StateUpdate(BaseModel):
    """更新状态"""
    name: Optional[str] = Field(None, min_length=1, max_length=255)
    color: Optional[str] = Field(None, max_length=50)
    group: Optional[StateGroupEnum] = None
    sequence: Optional[int] = Field(None, ge=0)
    is_active: Optional[bool] = None
    description: Optional[str] = Field(None, max_length=500)


class StateResponse(AuditSchema, SoftDeleteSchema, StateBase):
    """状态响应"""
    id: int
    project_id: int
    workspace_id: int
    
    class Config:
        from_attributes = True


class StateLite(BaseModel):
    """轻量级状态响应"""
    id: int
    name: str
    color: str
    group: str
    
    class Config:
        from_attributes = True


# ==================== Label Schema ====================

class LabelBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    color: str = Field(default="#6B7280", max_length=50)
    description: Optional[str] = Field(None, max_length=500)


class LabelCreate(LabelBase):
    """创建标签"""
    project_id: int


class LabelUpdate(BaseModel):
    """更新标签"""
    name: Optional[str] = Field(None, min_length=1, max_length=255)
    color: Optional[str] = Field(None, max_length=50)
    description: Optional[str] = Field(None, max_length=500)


class LabelResponse(AuditSchema, SoftDeleteSchema, LabelBase):
    """标签响应"""
    id: int
    project_id: int
    
    class Config:
        from_attributes = True


class LabelLite(BaseModel):
    """轻量级标签响应"""
    id: int
    name: str
    color: str
    
    class Config:
        from_attributes = True


# ==================== Default Issue Types ====================

DEFAULT_ISSUE_TYPES: List[dict] = [
    {"name": "Issue", "color": "#3B82F6", "icon": "circle", "is_default": True, "sequence": 1},
    {"name": "Task", "color": "#10B981", "icon": "check-circle", "sequence": 2},
    {"name": "Bug", "color": "#EF4444", "icon": "alert-circle", "sequence": 3},
    {"name": "Story", "color": "#F59E0B", "icon": "bookmark", "sequence": 4},
    {"name": "Epic", "color": "#8B5CF6", "icon": "layers", "sequence": 5},
]

# ==================== Default States ====================

DEFAULT_STATES: List[dict] = [
    {"name": "Backlog", "color": "#6B7280", "group": StateGroupEnum.BACKLOG, "sequence": 1},
    {"name": "Todo", "color": "#3B82F6", "group": StateGroupEnum.TODO, "sequence": 2},
    {"name": "In Progress", "color": "#F59E0B", "group": StateGroupEnum.IN_PROGRESS, "sequence": 3},
    {"name": "In Review", "color": "#8B5CF6", "group": StateGroupEnum.IN_PROGRESS, "sequence": 4},
    {"name": "Done", "color": "#10B981", "group": StateGroupEnum.DONE, "sequence": 5},
    {"name": "Cancelled", "color": "#EF4444", "group": StateGroupEnum.CANCELLED, "sequence": 6},
]