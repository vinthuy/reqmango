"""
Comment Service - 评论业务逻辑层
"""
from typing import List, Optional, Tuple
from uuid import UUID
from datetime import datetime

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload

from app.models.comment import Comment
from app.models.issue import Issue
from app.schemas.comment import CommentCreate, CommentUpdate
from app.core.exceptions import NotFoundException, ValidationException


async def create_comment(
    db: AsyncSession,
    comment_data: CommentCreate,
    user_id: UUID
) -> Comment:
    """创建评论"""
    # 验证工作项存在
    issue = await db.get(Issue, comment_data.issue_id)
    if not issue or issue.is_deleted:
        raise NotFoundException("Issue not found")

    # 验证父评论存在（如果是回复）
    if comment_data.parent_id:
        parent = await db.get(Comment, comment_data.parent_id)
        if not parent or parent.is_deleted:
            raise NotFoundException("Parent comment not found")

    comment = Comment(
        content=comment_data.content,
        html_content=comment_data.html_content,
        issue_id=comment_data.issue_id,
        author_id=user_id,
        parent_id=comment_data.parent_id,
        created_by_id=user_id
    )

    db.add(comment)
    await db.commit()
    await db.refresh(comment)
    return comment


async def get_comment_by_id(
    db: AsyncSession,
    comment_id: UUID
) -> Comment:
    """获取评论"""
    result = await db.execute(
        select(Comment).where(Comment.id == comment_id)
    )
    comment = result.scalar_one_or_none()
    if not comment or comment.is_deleted:
        raise NotFoundException("Comment not found")
    return comment


async def get_issue_comments(
    db: AsyncSession,
    issue_id: UUID,
    page: int = 1,
    page_size: int = 20,
    include_replies: bool = True
) -> Tuple[List[Comment], int]:
    """获取工作项的评论列表"""
    # 构建基础查询（只查顶级评论）
    query = select(Comment).where(
        Comment.issue_id == issue_id,
        Comment.is_deleted == False,
        Comment.parent_id == None
    )

    # 获取总数
    count_query = select(func.count(Comment.id)).where(
        Comment.issue_id == issue_id,
        Comment.is_deleted == False,
        Comment.parent_id == None
    )
    count_result = await db.execute(count_query)
    total = count_result.scalar()

    # 分页
    offset = (page - 1) * page_size
    query = query.order_by(Comment.created_at.desc()).limit(page_size).offset(offset)

    result = await db.execute(query)
    comments = list(result.scalars().all())

    # 如果需要，加载回复（但不在对象上设置属性，避免延迟加载问题）
    # 回复将在响应中单独加载
    # 注意：当前实现不包含回复数据，需要在响应层处理

    return comments, total


async def update_comment(
    db: AsyncSession,
    comment_id: UUID,
    update_data: CommentUpdate,
    user_id: UUID
) -> Comment:
    """更新评论"""
    comment = await get_comment_by_id(db, comment_id)

    if comment.author_id != user_id:
        raise ValidationException("You can only edit your own comments")

    if update_data.content is not None:
        comment.content = update_data.content
    if update_data.html_content is not None:
        comment.html_content = update_data.html_content

    await db.commit()
    await db.refresh(comment)
    return comment


async def delete_comment(
    db: AsyncSession,
    comment_id: UUID,
    user_id: UUID
) -> None:
    """删除评论"""
    comment = await get_comment_by_id(db, comment_id)

    if comment.author_id != user_id:
        raise ValidationException("You can only delete your own comments")

    comment.is_deleted = True
    await db.commit()


async def resolve_comment(
    db: AsyncSession,
    comment_id: UUID,
    user_id: UUID
) -> Comment:
    """标记评论为已解决"""
    comment = await get_comment_by_id(db, comment_id)

    comment.is_resolved = True
    comment.resolved_by_id = user_id
    comment.resolved_at = datetime.utcnow().isoformat()

    await db.commit()
    await db.refresh(comment)
    return comment


async def unresolve_comment(
    db: AsyncSession,
    comment_id: UUID,
    user_id: UUID
) -> Comment:
    """取消评论解决状态"""
    comment = await get_comment_by_id(db, comment_id)

    comment.is_resolved = False
    comment.resolved_by_id = None
    comment.resolved_at = None

    await db.commit()
    await db.refresh(comment)
    return comment


# 需要导入func
from sqlalchemy import func
