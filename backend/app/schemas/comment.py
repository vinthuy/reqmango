"""
Comment Schemas - 评论Schema定义
"""
from typing import Optional, List
from datetime import datetime
from pydantic import BaseModel, Field


class CommentBase(BaseModel):
    """评论基础Schema"""
    content: str
    html_content: Optional[str] = None


class CommentCreate(CommentBase):
    """创建评论"""
    issue_id: int
    parent_id: Optional[int] = None


class CommentUpdate(BaseModel):
    """更新评论"""
    content: Optional[str] = None
    html_content: Optional[str] = None


class CommentResponse(CommentBase):
    """评论响应"""
    id: int
    issue_id: int
    author_id: int
    parent_id: Optional[int] = None
    is_resolved: bool
    resolved_by_id: Optional[int] = None
    resolved_at: Optional[datetime] = None
    reaction_count: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CommentResolve(BaseModel):
    """解决评论"""
    is_resolved: bool = True


class CommentListResponse(BaseModel):
    """评论列表响应"""
    items: List[CommentResponse]
    total: int
    page: int
    page_size: int
