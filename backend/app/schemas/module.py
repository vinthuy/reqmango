from pydantic import BaseModel, Field
from typing import Optional
from uuid import UUID
from datetime import datetime
from .base import AuditSchema, SoftDeleteSchema
from .project import ProjectLite

class ModuleBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description: Optional[str] = None
    target_date: Optional[datetime] = None

class ModuleCreate(ModuleBase):
    project_id: UUID

class ModuleUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    target_date: Optional[datetime] = None
    archived_at: Optional[datetime] = None

class ModuleResponse(AuditSchema, SoftDeleteSchema, ModuleBase):
    project: ProjectLite
    total_issues: int
    completed_issues: int

class ModuleLite(BaseModel):
    id: UUID
    name: str