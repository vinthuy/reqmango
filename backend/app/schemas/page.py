from pydantic import BaseModel, Field
from typing import Optional
from .base import AuditSchema, SoftDeleteSchema
from .project import ProjectLite

class PageBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description_html: str = "<p></p>"
    is_locked: bool = False

class PageCreate(PageBase):
    project_id: int

class PageUpdate(BaseModel):
    name: Optional[str] = None
    description_html: Optional[str] = None
    is_locked: Optional[bool] = None

class PageResponse(AuditSchema, SoftDeleteSchema, PageBase):
    project: ProjectLite
    version: int = 1

class PageLite(BaseModel):
    id: int
    name: str