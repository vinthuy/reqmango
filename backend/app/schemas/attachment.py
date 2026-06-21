"""
Attachment Schemas - 附件Schema定义
"""
from typing import Optional, List
from pydantic import BaseModel, Field


class AttachmentBase(BaseModel):
    """附件基础Schema"""
    name: str = Field(..., max_length=255)
    file_name: str = Field(..., max_length=255)
    file_size: int = Field(..., gt=0)
    mime_type: str = Field(..., max_length=100)
    file_url: Optional[str] = Field(None, max_length=500)
    is_protected: bool = Field(default=False)


class AttachmentCreate(AttachmentBase):
    """创建附件"""
    issue_id: Optional[int] = None
    project_id: Optional[int] = None


class AttachmentUpdate(BaseModel):
    """更新附件"""
    name: Optional[str] = Field(None, max_length=255)
    is_protected: Optional[bool] = None


class AttachmentResponse(AttachmentBase):
    """附件响应"""
    id: int
    file_path: str
    issue_id: Optional[int] = None
    project_id: Optional[int] = None
    uploaded_by_id: int
    access_url: Optional[str] = None
    thumbnail_url: Optional[str] = None
    created_at: str
    updated_at: str

    class Config:
        from_attributes = True


class AttachmentListResponse(BaseModel):
    """附件列表响应"""
    items: List[AttachmentResponse]
    total: int


class AttachmentUploadResponse(BaseModel):
    """上传响应"""
    id: int
    name: str
    file_url: str
    file_size: int
    mime_type: str
