from pydantic import BaseModel, Field
from typing import Optional, Dict, Any
from datetime import datetime
from .base import AuditSchema, SoftDeleteSchema
from .user import UserLite
from .project import ProjectLite

class CycleBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description: Optional[str] = None
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    timezone: str = "UTC"

class CycleCreate(CycleBase):
    project_id: int

class CycleUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    archived_at: Optional[datetime] = None

class CycleResponse(AuditSchema, SoftDeleteSchema, CycleBase):
    project: ProjectLite
    owned_by: UserLite
    total_issues: int
    completed_issues: int
    progress_snapshot: Dict[str, Any] = {}
    version: int = 1

class CycleLite(BaseModel):
    id: int
    name: str
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None