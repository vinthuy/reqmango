from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any
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
    parent_id: Optional[int] = None
    state_id: Optional[int] = None
    assignee_ids: Optional[List[int]] = []
    label_ids: Optional[List[int]] = []
    estimate_point_id: Optional[int] = None
    type_id: Optional[int] = None
    external_id: Optional[str] = None
    external_source: Optional[str] = None

class IssueUpdate(BaseModel):
    name: Optional[str] = None
    description_html: Optional[str] = None
    priority: Optional[IssuePriority] = None
    state_id: Optional[int] = None
    assignee_ids: Optional[List[int]] = None
    label_ids: Optional[List[int]] = None
    start_date: Optional[date] = None
    target_date: Optional[date] = None
    estimate_point_id: Optional[int] = None
    cycle_id: Optional[int] = None
    module_ids: Optional[List[int]] = None

class IssueResponse(AuditSchema, SoftDeleteSchema, IssueBase):
    project: ProjectLite
    sequence_id: int
    state_id: int
    state_name: str
    state_group: str
    assignees: List[UserLite]
    labels: List[int]
    sub_issues_count: int
    link_count: int
    attachment_count: int
    completed_at: Optional[datetime] = None
    is_draft: bool = False
    parent_id: Optional[int] = None
    estimate_point_id: Optional[int] = None
    cycle_id: Optional[int] = None
    module_ids: List[int] = []

class IssueLite(BaseModel):
    id: int
    name: str
    sequence_id: int
    priority: IssuePriority
    state_id: int
    project_id: int
    project_identifier: str

class IssueSearchResult(BaseModel):
    id: int
    name: str
    sequence_id: int
    project_identifier: str
    project_id: int
    workspace_slug: str