from pydantic import BaseModel, Field
from typing import Optional
from enum import IntEnum
from .base import AuditSchema, SoftDeleteSchema
from .user import UserLite, UserResponse

class WorkspaceRole(IntEnum):
    ADMIN = 20
    MEMBER = 15
    GUEST = 5

class WorkspaceBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    slug: str = Field(..., min_length=1, max_length=50, pattern=r'^[a-z0-9-]+$')
    organization_size: Optional[str] = None
    timezone: str = "UTC"

class WorkspaceCreate(WorkspaceBase):
    pass

class WorkspaceUpdate(BaseModel):
    name: Optional[str] = None
    logo_url: Optional[str] = None
    timezone: Optional[str] = None

class WorkspaceResponse(AuditSchema, SoftDeleteSchema, WorkspaceBase):
    logo_url: Optional[str] = None
    owner_id: int

class WorkspaceLite(BaseModel):
    id: int
    name: str
    slug: str

class WorkspaceMemberBase(BaseModel):
    role: WorkspaceRole

class WorkspaceMemberCreate(WorkspaceMemberBase):
    member_id: int

class WorkspaceMemberResponse(AuditSchema, WorkspaceMemberBase):
    member: UserLite
    is_active: bool
    joining_date: Optional[str] = None