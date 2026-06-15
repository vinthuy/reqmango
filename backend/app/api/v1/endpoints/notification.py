"""
Notification API - 通知API端点
"""
from typing import List
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.session import get_db
from app.api.deps import get_current_user
from app.models.user import User
from app.schemas.notification import (
    NotificationCreate,
    NotificationUpdate,
    NotificationResponse,
    NotificationMarkRead,
    NotificationSummary
)
from app.services import notification

router = APIRouter(tags=["通知"])


@router.post("/", response_model=NotificationResponse, status_code=201)
async def create_notification(
    notification_data: NotificationCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """创建通知"""
    notif = await notification.create_notification(
        db=db,
        notification_data=notification_data,
        user_id=current_user.id
    )
    return notif


@router.get("/", response_model=List[NotificationResponse])
async def list_notifications(
    unread_only: bool = Query(False),
    limit: int = Query(50, le=100),
    offset: int = Query(0, ge=0),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """获取当前用户的通知列表"""
    notifications, _ = await notification.get_user_notifications(
        db=db,
        user_id=current_user.id,
        unread_only=unread_only,
        limit=limit,
        offset=offset
    )
    return notifications


@router.get("/summary", response_model=NotificationSummary)
async def get_notification_summary(
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """获取通知摘要"""
    summary = await notification.get_notification_summary(
        db=db,
        user_id=current_user.id
    )
    return summary


@router.get("/{notification_id}", response_model=NotificationResponse)
async def get_notification(
    notification_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """获取通知详情"""
    notif = await notification.get_notification_by_id(db, notification_id)

    if notif.recipient_id != current_user.id:
        raise HTTPException(status_code=404, detail="Notification not found")

    return notif


@router.patch("/{notification_id}/read", response_model=NotificationResponse)
async def mark_as_read(
    notification_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """标记通知已读"""
    notif = await notification.mark_as_read(
        db=db,
        notification_id=notification_id,
        user_id=current_user.id
    )
    return notif


@router.post("/read-all")
async def mark_all_as_read(
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """标记所有通知已读"""
    count = await notification.mark_all_as_read(
        db=db,
        user_id=current_user.id
    )
    return {"marked_count": count}


@router.delete("/{notification_id}", status_code=204)
async def delete_notification(
    notification_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """删除通知"""
    await notification.delete_notification(
        db=db,
        notification_id=notification_id,
        user_id=current_user.id
    )


@router.post("/bulk", response_model=List[NotificationResponse])
async def create_bulk_notification(
    notification_data: NotificationCreate,
    recipient_ids: List[UUID],
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """批量创建通知"""
    notifications = await notification.create_bulk_notification(
        db=db,
        notification_data=notification_data,
        recipient_ids=recipient_ids,
        user_id=current_user.id
    )
    return notifications
