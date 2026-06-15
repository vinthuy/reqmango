"""
Attachment Service - 附件业务逻辑层
"""
from typing import List, Optional
from uuid import UUID
import os
from datetime import datetime

from fastapi import UploadFile
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy import func

from app.models.attachment import Attachment
from app.models.project import Project
from app.schemas.attachment import AttachmentCreate, AttachmentUpdate
from app.core.exceptions import NotFoundException, ValidationException

# 附件存储路径
ATTACHMENT_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(__file__))), "uploads", "attachments")


async def upload_attachment(
    db: AsyncSession,
    file: UploadFile,
    upload_data: AttachmentCreate,
    user_id: UUID,
    upload_dir: str = ATTACHMENT_DIR
) -> Attachment:
    """上传附件"""
    # 验证项目（如果指定了项目）
    if upload_data.project_id:
        project = await db.get(Project, upload_data.project_id)
        if not project or project.is_deleted:
            raise NotFoundException("Project not found")

    # 确保上传目录存在
    os.makedirs(upload_dir, exist_ok=True)

    # 生成文件名
    file_id = str(UUID())
    file_ext = os.path.splitext(file.filename)[1] if file.filename else ""
    stored_filename = f"{file_id}{file_ext}"
    file_path = os.path.join(upload_dir, stored_filename)

    # 保存文件
    try:
        content = await file.read()
        with open(file_path, 'wb') as f:
            f.write(content)
    except Exception as e:
        raise ValidationException(f"Failed to save file: {str(e)}")

    # 创建附件记录
    attachment = Attachment(
        name=upload_data.name or file.filename,
        file_name=file.filename,
        file_size=len(content),
        mime_type=file.content_type or "application/octet-stream",
        file_path=file_path,
        file_url=f"/uploads/attachments/{stored_filename}",
        is_protected=upload_data.is_protected,
        issue_id=upload_data.issue_id,
        project_id=upload_data.project_id,
        uploaded_by_id=user_id,
        created_by_id=user_id
    )

    db.add(attachment)
    await db.commit()
    await db.refresh(attachment)
    return attachment


async def get_attachment_by_id(
    db: AsyncSession,
    attachment_id: UUID
) -> Attachment:
    """获取附件"""
    result = await db.execute(
        select(Attachment).where(Attachment.id == attachment_id)
    )
    attachment = result.scalar_one_or_none()
    if not attachment or attachment.is_deleted:
        raise NotFoundException("Attachment not found")
    return attachment


async def get_issue_attachments(
    db: AsyncSession,
    issue_id: UUID
) -> List[Attachment]:
    """获取工作项的所有附件"""
    result = await db.execute(
        select(Attachment).where(
            Attachment.issue_id == issue_id,
            Attachment.is_deleted == False
        ).order_by(Attachment.created_at.desc())
    )
    return list(result.scalars().all())


async def get_project_attachments(
    db: AsyncSession,
    project_id: UUID
) -> List[Attachment]:
    """获取项目的所有附件"""
    result = await db.execute(
        select(Attachment).where(
            Attachment.project_id == project_id,
            Attachment.is_deleted == False
        ).order_by(Attachment.created_at.desc())
    )
    return list(result.scalars().all())


async def update_attachment(
    db: AsyncSession,
    attachment_id: UUID,
    update_data: AttachmentUpdate,
    user_id: UUID
) -> Attachment:
    """更新附件"""
    attachment = await get_attachment_by_id(db, attachment_id)

    if attachment.uploaded_by_id != user_id:
        raise ValidationException("You can only update your own attachments")

    if update_data.name is not None:
        attachment.name = update_data.name
    if update_data.is_protected is not None:
        attachment.is_protected = update_data.is_protected

    await db.commit()
    await db.refresh(attachment)
    return attachment


async def delete_attachment(
    db: AsyncSession,
    attachment_id: UUID,
    user_id: UUID
) -> None:
    """删除附件"""
    attachment = await get_attachment_by_id(db, attachment_id)

    if attachment.uploaded_by_id != user_id:
        raise ValidationException("You can only delete your own attachments")

    # 软删除
    attachment.is_deleted = True

    # 删除物理文件
    if os.path.exists(attachment.file_path):
        try:
            os.remove(attachment.file_path)
        except Exception:
            pass  # 忽略文件删除错误

    await db.commit()


async def get_attachment_download_url(
    db: AsyncSession,
    attachment_id: UUID,
    user_id: UUID
) -> str:
    """获取附件下载URL"""
    attachment = await get_attachment_by_id(db, attachment_id)

    # 如果是受保护的附件，需要验证权限
    if attachment.is_protected:
        # TODO: 实现权限验证
        pass

    return attachment.file_url
