"""
Estimate Point Schemas - 估算点Schema定义
"""
from typing import Optional, List
from pydantic import BaseModel, Field


class EstimatePointBase(BaseModel):
    """估算点基础Schema"""
    name: str = Field(..., max_length=50, description="估算点名称")
    value: int = Field(..., ge=0, description="估算点数值")
    is_default: bool = Field(default=False, description="是否为默认选项")
    sequence: int = Field(default=1, ge=0, description="排序顺序")


class EstimatePointCreate(EstimatePointBase):
    """创建估算点"""
    project_id: int


class EstimatePointUpdate(BaseModel):
    """更新估算点"""
    name: Optional[str] = Field(None, max_length=50)
    value: Optional[int] = Field(None, ge=0)
    is_default: Optional[bool] = None
    sequence: Optional[int] = Field(None, ge=0)


class EstimatePointResponse(EstimatePointBase):
    """估算点响应"""
    id: int
    project_id: int
    created_at: str
    updated_at: str

    class Config:
        from_attributes = True


class EstimatePointBulkCreate(BaseModel):
    """批量创建估算点"""
    points: List[EstimatePointCreate]


class EstimatePointReorder(BaseModel):
    """重新排序估算点"""
    point_ids: List[int]
