"""Data types mirroring the Go client's DTO definitions."""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime
from typing import Any


# -- Auth --

@dataclass(slots=True)
class TokenResponse:
    access_token: str = ""
    token_type: str = ""
    expires_at: datetime | None = None


@dataclass(slots=True)
class User:
    id: int = 0
    email: str = ""
    username: str = ""
    display_name: str = ""
    avatar_url: str = ""


@dataclass(slots=True)
class PATResponse:
    id: int = 0
    name: str = ""
    token_prefix: str = ""
    scopes: str = ""
    last_used_at: datetime | None = None
    expires_at: datetime | None = None
    revoked_at: datetime | None = None
    created_at: datetime | None = None


@dataclass(slots=True)
class CreatePATResponse(PATResponse):
    token: str = ""


# -- Workspaces --

@dataclass(slots=True)
class Workspace:
    id: int = 0
    name: str = ""
    slug: str = ""
    created_at: datetime | None = None
    updated_at: datetime | None = None


@dataclass(slots=True)
class UserLite:
    id: int = 0
    display_name: str = ""
    email: str = ""
    avatar_url: str = ""


@dataclass(slots=True)
class Member:
    id: int = 0
    workspace_id: int = 0
    user_id: int = 0
    role: int = 0
    is_active: bool = False
    user: UserLite = field(default_factory=UserLite)
    created_at: datetime | None = None
    updated_at: datetime | None = None


# -- Projects --

@dataclass(slots=True)
class Project:
    id: int = 0
    name: str = ""
    identifier: str = ""
    description: str = ""
    workspace_id: int = 0
    archived_at: datetime | None = None
    total_issues: int = 0
    total_members: int = 0
    is_favorite: bool = False
    created_at: datetime | None = None
    updated_at: datetime | None = None


# -- Issues --

@dataclass(slots=True)
class Issue:
    id: int = 0
    name: str = ""
    description_html: str = ""
    priority: str = ""
    sequence_id: int = 0
    sort_order: float = 0.0
    start_date: datetime | None = None
    target_date: datetime | None = None
    completed_at: datetime | None = None
    is_draft: bool = False
    archived_at: datetime | None = None
    project_id: int = 0
    workspace_id: int = 0
    state_id: int = 0
    state_name: str = ""
    state_group: str = ""
    parent_id: int | None = None
    depth: int = 0
    assignees: list[UserLite] = field(default_factory=list)
    labels: list[int] = field(default_factory=list)
    sub_issues_count: int = 0
    link_count: int = 0
    estimate_point_id: int | None = None
    cycle_id: int | None = None
    module_ids: list[int] = field(default_factory=list)
    release_id: int | None = None
    created_at: datetime | None = None
    updated_at: datetime | None = None


@dataclass(slots=True)
class IssueListResult:
    items: list[Issue] = field(default_factory=list)
    total: int = 0


@dataclass(slots=True)
class IssueSearchResult:
    id: int = 0
    name: str = ""
    sequence_id: int = 0
    project_identifier: str = ""
    project_id: int = 0
    workspace_slug: str = ""


@dataclass(slots=True)
class Comment:
    id: int = 0
    issue_id: int = 0
    author_id: int = 0
    body: str = ""
    is_resolved: bool = False
    parent_id: int | None = None
    created_at: datetime | None = None
    updated_at: datetime | None = None


# -- Cycles --

@dataclass(slots=True)
class Cycle:
    id: int = 0
    name: str = ""
    description: str | None = None
    status: str = ""
    progress: float = 0.0
    total_issues: int = 0
    completed_issues: int = 0
    start_date: str = ""
    end_date: str | None = None
    project_id: int = 0
    workspace_id: int = 0
    created_at: datetime | None = None
    updated_at: datetime | None = None


@dataclass(slots=True)
class CycleListResult:
    items: list[Cycle] = field(default_factory=list)
    total: int = 0
    limit: int = 0
    offset: int = 0


@dataclass(slots=True)
class StateBreakdown:
    state: str = ""
    group: str = ""
    count: int = 0


@dataclass(slots=True)
class CycleProgress:
    cycle_id: int = 0
    cycle_name: str = ""
    total_issues: int = 0
    completed_issues: int = 0
    progress: float = 0.0
    state_breakdown: list[StateBreakdown] = field(default_factory=list)


@dataclass(slots=True)
class BurndownDayPoint:
    day_index: int = 0
    date: str = ""
    ideal_remaining: float = 0.0
    actual_completed: int = 0
    actual_remaining: float = 0.0


@dataclass(slots=True)
class BurndownData:
    cycle_id: int = 0
    cycle_name: str = ""
    start_date: str = ""
    end_date: str = ""
    total_issues: int = 0
    total_days: int = 0
    days_elapsed: int = 0
    ideal_daily_burn: float = 0.0
    ideal_remaining: float = 0.0
    actual_completed: int = 0
    actual_remaining: float = 0.0
    is_on_track: bool = False
    daily_points: list[BurndownDayPoint] = field(default_factory=list)


# -- Meta --

@dataclass(slots=True)
class State:
    id: int = 0
    name: str = ""
    color: str = ""
    group: str = ""
    sequence: int = 0
    is_default: bool = False
    is_active: bool = False
    project_id: int = 0
    workspace_id: int = 0


@dataclass(slots=True)
class Label:
    id: int = 0
    name: str = ""
    color: str = ""
    description: str = ""
    project_id: int = 0
    workspace_id: int = 0


@dataclass(slots=True)
class IssueType:
    id: int = 0
    name: str = ""
    color: str = ""
    icon: str = ""
    description: str = ""
    level: str = ""
    parent_type_id: int | None = None
    is_default: bool = False
    sequence: int = 0
    is_active: bool = False
    project_id: int = 0
    workspace_id: int = 0


@dataclass(slots=True)
class Page:
    id: int = 0
    title: str = ""
    content: str = ""
    published: bool = False
    sequence: int = 0
    parent_id: int | None = None
    depth: int = 0
    project_id: int = 0
    workspace_id: int = 0
    created_at: datetime | None = None
    updated_at: datetime | None = None


@dataclass(slots=True)
class Notification:
    id: int = 0
    title: str = ""
    message: str = ""
    type: str = ""
    priority: str = ""
    is_read: bool = False
    read_at: datetime | None = None
    action_url: str = ""
    issue_id: int | None = None
    created_at: datetime | None = None


# -- AI --

@dataclass(slots=True)
class AISearchResponse:
    rql: str = ""
    explanation: str = ""
    issues: list[dict[str, Any]] = field(default_factory=list)


@dataclass(slots=True)
class AIChatReply:
    text: str = ""
    thread_id: int = 0
    tool_calls: list[str] = field(default_factory=list)


# -- Agents --

@dataclass(slots=True)
class Agent:
    id: int = 0
    workspace_id: int = 0
    name: str = ""
    avatar: str = ""
    agent_type: str = ""
    capabilities: list[str] = field(default_factory=list)
    status: str = ""
    model_override: str | None = None
    system_prompt: str | None = None
    template_id: int | None = None
    created_at: datetime | None = None
    updated_at: datetime | None = None


@dataclass(slots=True)
class AgentActivity:
    id: int = 0
    agent_id: int = 0
    issue_id: int | None = None
    action: str = ""
    result_summary: str = ""
    rating: int | None = None
    executed_at: datetime | None = None
    agent_name: str = ""
    task_context: str = ""
    created_at: datetime | None = None
    updated_at: datetime | None = None


@dataclass(slots=True)
class AgentTask:
    id: int = 0
    title: str = ""
    description: str = ""
    status: str = ""
    priority: int = 0
    progress: int = 0
    task_type: str = ""
    output_data: Any = None
    error_info: str = ""
    failure_reason: str = ""
    workspace_id: int = 0
    project_id: int | None = None
    issue_id: int | None = None
    enqueued_at: datetime | None = None
    started_at: datetime | None = None
    completed_at: datetime | None = None
    created_at: datetime | None = None
    updated_at: datetime | None = None


@dataclass(slots=True)
class TaskLog:
    id: int = 0
    task_id: int = 0
    level: str = ""
    message: str = ""
    created_at: datetime | None = None


# -- Chat --

@dataclass(slots=True)
class Message:
    id: int = 0
    chat_id: int = 0
    sender_id: int = 0
    sender_type: str = ""
    content: str = ""
    reply_to_id: int | None = None
    edited_at: datetime | None = None
    deleted_at: datetime | None = None
    created_at: datetime | None = None


@dataclass(slots=True)
class Chat:
    id: int = 0
    workspace_id: int = 0
    project_id: int | None = None
    issue_id: int | None = None
    type: str = ""
    title: str = ""
    created_at: datetime | None = None
    messages: list[Message] = field(default_factory=list)
