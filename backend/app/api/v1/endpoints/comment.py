"""
Comment API - 评论API端点
"""
from typing import List
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.session import get_db
from app.api.deps import get_current_user
from app.models.user import User
from app.schemas.comment import (
    CommentCreate,
    CommentUpdate,
    CommentResponse,
    CommentResolve,
    CommentListResponse
)
from app.services import comment

router = APIRouter(tags=["评论"])


@router.post("/", response_model=CommentResponse, status_code=201)
async def create_comment(
    comment_data: CommentCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """创建评论"""
    new_comment = await comment.create_comment(
        db=db,
        comment_data=comment_data,
        user_id=current_user.id
    )
    return new_comment


@router.get("/issue/{issue_id}", response_model=CommentListResponse)
async def list_issue_comments(
    issue_id: UUID,
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """获取工作项的评论列表"""
    comments, total = await comment.get_issue_comments(
        db=db,
        issue_id=issue_id,
        page=page,
        page_size=page_size
    )
    return {
        "items": comments,
        "total": total,
        "page": page,
        "page_size": page_size
    }


@router.get("/{comment_id}", response_model=CommentResponse)
async def get_comment(
    comment_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """获取评论详情"""
    c = await comment.get_comment_by_id(db, comment_id)
    return c


@router.patch("/{comment_id}", response_model=CommentResponse)
async def update_comment(
    comment_id: UUID,
    update_data: CommentUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """更新评论"""
    c = await comment.update_comment(
        db=db,
        comment_id=comment_id,
        update_data=update_data,
        user_id=current_user.id
    )
    return c


@router.delete("/{comment_id}", status_code=204)
async def delete_comment(
    comment_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """删除评论"""
    await comment.delete_comment(
        db=db,
        comment_id=comment_id,
        user_id=current_user.id
    )


@router.post("/{comment_id}/resolve", response_model=CommentResponse)
async def resolve_comment(
    comment_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """标记评论为已解决"""
    c = await comment.resolve_comment(
        db=db,
        comment_id=comment_id,
        user_id=current_user.id
    )
    return c


@router.post("/{comment_id}/unresolve", response_model=CommentResponse)
async def unresolve_comment(
    comment_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """取消评论解决状态"""
    c = await comment.unresolve_comment(
        db=db,
        comment_id=comment_id,
        user_id=current_user.id
    )
    return c
