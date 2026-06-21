"""
Notification Schemas - 通知Schema定义
"""
from typing import Optional, List
from pydantic import BaseModel, Field


class NotificationBase(BaseModel):
    """通知基础Schema"""
    title: str = Field(..., max_length=255)
    message: str
    type: str = Field(default="info")  # info, warning, error, success
    priority: str = Field(default="medium")  # low, medium, high, urgent
    action_url: Optional[str] = Field(None, max_length=500)


class NotificationCreate(NotificationBase):
    """创建通知"""
    recipient_id: int
    sender_id: Optional[int] = None
    project_id: Optional[int] = None
    issue_id: Optional[int] = None


class NotificationUpdate(BaseModel):
    """更新通知"""
    is_read: Optional[bool] = None
    title: Optional[str] = Field(None, max_length=255)
    message: Optional[str] = None


class NotificationResponse(NotificationBase):
    """通知响应"""
    id: int
    recipient_id: int
    sender_id: Optional[int] = None
    project_id: Optional[int] = None
    issue_id: Optional[int] = None
    is_read: bool
    read_at: Optional[str] = None
    created_at: str
    updated_at: str

    class Config:
        from_attributes = True


class NotificationMarkRead(BaseModel):
    """标记已读"""
    notification_ids: List[int]


class NotificationSummary(BaseModel):
    """通知摘要"""
    total: int
    unread: int
    unread_by_type: dict
