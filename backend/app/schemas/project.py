from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime
from .base import AuditSchema, SoftDeleteSchema
from .user import UserLite
from .workspace import WorkspaceLite

class ProjectBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    identifier: str = Field(..., min_length=1, max_length=10, pattern=r'^[A-Z]+$')
    description: Optional[str] = None
    is_public: bool = False
    timezone: str = "UTC"

class ProjectCreate(ProjectBase):
    default_assignee_id: Optional[int] = None

class ProjectUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    is_public: Optional[bool] = None
    archived_at: Optional[datetime] = None

class ProjectResponse(AuditSchema, SoftDeleteSchema, ProjectBase):
    workspace: WorkspaceLite
    total_issues: int
    total_members: int
    default_assignee: Optional[UserLite] = None
    logo_url: Optional[str] = None
    is_favorite: bool = False

class ProjectLite(BaseModel):
    id: int
    name: str
    identifier: str

class ProjectMemberBase(BaseModel):
    role: int = 15

class ProjectMemberCreate(ProjectMemberBase):
    user_id: int

class ProjectMemberResponse(AuditSchema, ProjectMemberBase):
    user: UserLite
    is_active: bool