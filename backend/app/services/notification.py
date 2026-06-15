"""
Notification Service - 通知业务逻辑层
"""
from typing import List, Optional
from uuid import UUID
from datetime import datetime

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy import func, and_

from app.models.notification import Notification
from app.models.user import User
from app.schemas.notification import NotificationCreate, NotificationUpdate
from app.core.exceptions import NotFoundException


async def create_notification(
    db: AsyncSession,
    notification_data: NotificationCreate,
    user_id: Optional[UUID] = None
) -> Notification:
    """创建通知"""
    notification = Notification(
        title=notification_data.title,
        message=notification_data.message,
        type=notification_data.type,
        priority=notification_data.priority,
        action_url=notification_data.action_url,
        recipient_id=notification_data.recipient_id,
        sender_id=notification_data.sender_id or user_id,
        project_id=notification_data.project_id,
        issue_id=notification_data.issue_id,
        created_by_id=user_id
    )

    db.add(notification)
    await db.commit()
    await db.refresh(notification)
    return notification


async def get_notification_by_id(
    db: AsyncSession,
    notification_id: UUID
) -> Notification:
    """获取通知"""
    result = await db.execute(
        select(Notification).where(Notification.id == notification_id)
    )
    notification = result.scalar_one_or_none()
    if not notification or notification.is_deleted:
        raise NotFoundException("Notification not found")
    return notification


async def get_user_notifications(
    db: AsyncSession,
    user_id: UUID,
    unread_only: bool = False,
    limit: int = 50,
    offset: int = 0
) -> tuple[List[Notification], int]:
    """获取用户通知列表"""
    # 构建查询
    query = select(Notification).where(
        Notification.recipient_id == user_id,
        Notification.is_deleted == False
    )

    if unread_only:
        query = query.where(Notification.is_read == False)

    # 获取总数
    count_query = select(func.count(Notification.id)).where(
        Notification.recipient_id == user_id,
        Notification.is_deleted == False
    )
    if unread_only:
        count_query = count_query.where(Notification.is_read == False)

    count_result = await db.execute(count_query)
    total = count_result.scalar()

    # 分页查询
    query = query.order_by(Notification.created_at.desc()).limit(limit).offset(offset)
    result = await db.execute(query)
    notifications = list(result.scalars().all())

    return notifications, total


async def get_notification_summary(
    db: AsyncSession,
    user_id: UUID
) -> dict:
    """获取通知摘要"""
    # 总数
    total_query = select(func.count(Notification.id)).where(
        Notification.recipient_id == user_id,
        Notification.is_deleted == False
    )
    total_result = await db.execute(total_query)
    total = total_result.scalar()

    # 未读数
    unread_query = select(func.count(Notification.id)).where(
        Notification.recipient_id == user_id,
        Notification.is_deleted == False,
        Notification.is_read == False
    )
    unread_result = await db.execute(unread_query)
    unread = unread_result.scalar()

    # 按类型统计未读
    type_query = select(
        Notification.type,
        func.count(Notification.id)
    ).where(
        Notification.recipient_id == user_id,
        Notification.is_deleted == False,
        Notification.is_read == False
    ).group_by(Notification.type)

    type_result = await db.execute(type_query)
    unread_by_type = {row[0]: row[1] for row in type_result}

    return {
        "total": total,
        "unread": unread,
        "unread_by_type": unread_by_type
    }


async def mark_as_read(
    db: AsyncSession,
    notification_id: UUID,
    user_id: UUID
) -> Notification:
    """标记通知已读"""
    notification = await get_notification_by_id(db, notification_id)

    if notification.recipient_id != user_id:
        raise NotFoundException("Notification not found")

    notification.is_read = True
    notification.read_at = datetime.utcnow().isoformat()

    await db.commit()
    await db.refresh(notification)
    return notification


async def mark_all_as_read(
    db: AsyncSession,
    user_id: UUID
) -> int:
    """标记所有通知已读"""
    result = await db.execute(
        select(Notification).where(
            Notification.recipient_id == user_id,
            Notification.is_deleted == False,
            Notification.is_read == False
        )
    )
    notifications = result.scalars().all()

    count = 0
    for notification in notifications:
        notification.is_read = True
        notification.read_at = datetime.utcnow().isoformat()
        count += 1

    await db.commit()
    return count


async def delete_notification(
    db: AsyncSession,
    notification_id: UUID,
    user_id: UUID
) -> None:
    """删除通知"""
    notification = await get_notification_by_id(db, notification_id)

    if notification.recipient_id != user_id:
        raise NotFoundException("Notification not found")

    notification.is_deleted = True
    await db.commit()


async def create_bulk_notification(
    db: AsyncSession,
    notification_data: NotificationCreate,
    recipient_ids: List[UUID],
    user_id: Optional[UUID] = None
) -> List[Notification]:
    """批量创建通知"""
    notifications = []
    for recipient_id in recipient_ids:
        notification_data.recipient_id = recipient_id
        notification = await create_notification(db, notification_data, user_id)
        notifications.append(notification)
    return notifications
