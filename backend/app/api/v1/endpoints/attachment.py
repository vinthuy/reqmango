"""
Attachment API - 附件API端点
"""
from typing import List
import os

from fastapi import APIRouter, Depends, HTTPException, Query, UploadFile, File, Form
from fastapi.responses import FileResponse
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.session import get_db
from app.api.deps import get_current_user
from app.models.user import User
from app.schemas.attachment import (
    AttachmentCreate,
    AttachmentUpdate,
    AttachmentResponse,
    AttachmentListResponse,
    AttachmentUploadResponse
)
from app.services import attachment as attachment_service

router = APIRouter(prefix="/attachments", tags=["附件"])

# 附件存储目录
ATTACHMENT_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(__file__)))), "uploads", "attachments")


@router.post("/", response_model=AttachmentResponse, status_code=201)
async def upload_attachment(
    file: UploadFile = File(...),
    name: str = Form(None),
    issue_id: int = Form(None),
    project_id: int = Form(None),
    is_protected: bool = Form(False),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """上传附件"""
    upload_data = AttachmentCreate(
        name=name,
        file_name=file.filename,
        file_size=0,  # 临时值，实际在service中更新
        mime_type=file.content_type or "application/octet-stream",
        issue_id=issue_id,
        project_id=project_id,
        is_protected=is_protected
    )

    uploaded = await attachment_service.upload_attachment(
        db=db,
        file=file,
        upload_data=upload_data,
        user_id=current_user.id,
        upload_dir=ATTACHMENT_DIR
    )
    return uploaded


@router.get("/issue/{issue_id}", response_model=List[AttachmentResponse])
async def list_issue_attachments(
    issue_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """获取工作项的附件列表"""
    attachments = await attachment_service.get_issue_attachments(
        db=db,
        issue_id=issue_id
    )
    return attachments


@router.get("/project/{project_id}", response_model=List[AttachmentResponse])
async def list_project_attachments(
    project_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """获取项目的附件列表"""
    attachments = await attachment_service.get_project_attachments(
        db=db,
        project_id=project_id
    )
    return attachments


@router.get("/{attachment_id}", response_model=AttachmentResponse)
async def get_attachment(
    attachment_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """获取附件详情"""
    att = await attachment_service.get_attachment_by_id(db, attachment_id)
    return att


@router.get("/{attachment_id}/download")
async def download_attachment(
    attachment_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """下载附件"""
    att = await attachment_service.get_attachment_by_id(db, attachment_id)

    if not os.path.exists(att.file_path):
        raise HTTPException(status_code=404, detail="File not found")

    return FileResponse(
        path=att.file_path,
        filename=att.file_name,
        media_type=att.mime_type
    )


@router.patch("/{attachment_id}", response_model=AttachmentResponse)
async def update_attachment(
    attachment_id: int,
    update_data: AttachmentUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """更新附件"""
    att = await attachment_service.update_attachment(
        db=db,
        attachment_id=attachment_id,
        update_data=update_data,
        user_id=current_user.id
    )
    return att


@router.delete("/{attachment_id}", status_code=204)
async def delete_attachment(
    attachment_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """删除附件"""
    await attachment_service.delete_attachment(
        db=db,
        attachment_id=attachment_id,
        user_id=current_user.id
    )
