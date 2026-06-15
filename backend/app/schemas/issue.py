from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any
from uuid import UUID
from datetime import datetime, date
from enum import Enum
from .base import AuditSchema, SoftDeleteSchema
from .user import UserLite
from .project import ProjectLite

class IssuePriority(str, Enum):
    URGENT = "urgent"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    NONE = "none"

class IssueType(str, Enum):
    ISSUE = "issue"
    TASK = "task"
    BUG = "bug"
    STORY = "story"
    EPIC = "epic"

class IssueBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description_html: str = "<p></p>"
    description_json: Dict[str, Any] = {}
    priority: IssuePriority = IssuePriority.NONE
    start_date: Optional[date] = None
    target_date: Optional[date] = None

class IssueCreate(IssueBase):
    parent_id: Optional[UUID] = None
    state_id: Optional[UUID] = None
    assignee_ids: Optional[List[UUID]] = []
    label_ids: Optional[List[UUID]] = []
    estimate_point_id: Optional[UUID] = None
    type_id: Optional[UUID] = None
    external_id: Optional[str] = None
    external_source: Optional[str] = None

class IssueUpdate(BaseModel):
    name: Optional[str] = None
    description_html: Optional[str] = None
    priority: Optional[IssuePriority] = None
    state_id: Optional[UUID] = None
    assignee_ids: Optional[List[UUID]] = None
    label_ids: Optional[List[UUID]] = None
    start_date: Optional[date] = None
    target_date: Optional[date] = None
    estimate_point_id: Optional[UUID] = None
    cycle_id: Optional[UUID] = None
    module_ids: Optional[List[UUID]] = None

class IssueResponse(AuditSchema, SoftDeleteSchema, IssueBase):
    project: ProjectLite
    sequence_id: int
    state_id: UUID
    state_name: str
    state_group: str
    assignees: List[UserLite]
    labels: List[UUID]
    sub_issues_count: int
    link_count: int
    attachment_count: int
    completed_at: Optional[datetime] = None
    is_draft: bool = False
    parent_id: Optional[UUID] = None
    estimate_point_id: Optional[UUID] = None
    cycle_id: Optional[UUID] = None
    module_ids: List[UUID] = []

class IssueLite(BaseModel):
    id: UUID
    name: str
    sequence_id: int
    priority: IssuePriority
    state_id: UUID
    project_id: UUID
    project_identifier: str

class IssueSearchResult(BaseModel):
    id: UUID
    name: str
    sequence_id: int
    project_identifier: str
    project_id: UUID
    workspace_slug: str