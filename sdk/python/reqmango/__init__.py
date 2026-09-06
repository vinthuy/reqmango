"""reqmango Python SDK — client for the reqmango project management API."""

from __future__ import annotations

from typing import Any

from ._http import APIError, DEFAULT_BASE_URL, HTTPClient
from ._types import (
    Agent,
    AgentActivity,
    AgentTask,
    AIChatReply,
    AISearchResponse,
    BurndownData,
    Chat,
    Comment,
    CreatePATResponse,
    Cycle,
    CycleListResult,
    CycleProgress,
    Issue,
    IssueListResult,
    IssueSearchResult,
    IssueType,
    Label,
    Member,
    Message,
    Notification,
    Page,
    PATResponse,
    Project,
    State,
    TaskLog,
    TokenResponse,
    User,
    Workspace,
)

__all__ = [
    "ReqMangoClient",
    "APIError",
    # types
    "Agent", "AgentActivity", "AgentTask", "AIChatReply", "AISearchResponse",
    "BurndownData", "Chat", "Comment", "CreatePATResponse", "Cycle",
    "CycleListResult", "CycleProgress", "Issue", "IssueListResult",
    "IssueSearchResult", "IssueType", "Label", "Member", "Message",
    "Notification", "Page", "PATResponse", "Project", "State", "TaskLog",
    "TokenResponse", "User", "Workspace",
]


def _to(cls: type, data: Any) -> Any:
    """Convert a dict to a dataclass, recursively handling nested lists/dicts."""
    if data is None:
        return None
    if isinstance(data, list):
        return [_to(cls, item) for item in data]
    if isinstance(data, cls):
        return data
    if not isinstance(data, dict):
        return data
    import dataclasses
    field_types = {f.name: f.type for f in dataclasses.fields(cls)}
    kwargs: dict[str, Any] = {}
    for fname, ftype in field_types.items():
        json_key = fname  # dataclass field names match JSON keys (already snake_case)
        val = data.get(json_key)
        if val is None:
            continue
        # Handle nested dataclass types in list annotations
        if isinstance(ftype, str) and "list[" in ftype:
            inner_name = ftype.split("[")[1].rstrip("]")
            import sys
            inner_cls = getattr(sys.modules[__name__], inner_name, None)
            if inner_cls and isinstance(val, list):
                val = [_to(inner_cls, v) for v in val]
        elif isinstance(ftype, str):
            import sys
            nested_cls = getattr(sys.modules[__name__], ftype, None)
            if nested_cls and isinstance(val, dict):
                val = _to(nested_cls, val)
        kwargs[fname] = val
    return cls(**kwargs)


class ReqMangoClient:
    """High-level client for the reqmango REST API.

    Usage::

        with ReqMangoClient(token="reqmango_pat_xxx") as client:
            issues = client.list_issues(project_id=1)
    """

    def __init__(self, base_url: str = "", token: str = ""):
        self._http = HTTPClient(base_url=base_url, token=token)

    @property
    def base_url(self) -> str:
        return self._http.base_url

    def close(self) -> None:
        self._http.close()

    def __enter__(self) -> ReqMangoClient:
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()

    # ------------------------------------------------------------------ Auth

    def login(self, email: str, password: str) -> TokenResponse:
        data = self._http.post_json("/auth/login", {"email": email, "password": password})
        return _to(TokenResponse, data)

    def me(self) -> User:
        data = self._http.get_json("/auth/me")
        return _to(User, data)

    def create_pat(self, name: str, expires_at: str | None = None) -> CreatePATResponse:
        body: dict[str, Any] = {"name": name}
        if expires_at:
            body["expires_at"] = expires_at
        data = self._http.post_json("/auth/tokens", body)
        return _to(CreatePATResponse, data)

    def list_pats(self) -> list[PATResponse]:
        data = self._http.get_json("/auth/tokens")
        return _to(PATResponse, data)

    def revoke_pat(self, pat_id: int) -> None:
        self._http.delete_json(f"/auth/tokens/{pat_id}")

    # ----------------------------------------------------------- Workspaces

    def list_workspaces(self) -> list[Workspace]:
        data = self._http.get_json("/workspaces")
        return _to(Workspace, data)

    def create_workspace(self, name: str, slug: str) -> Workspace:
        data = self._http.post_json("/workspaces", {"name": name, "slug": slug})
        return _to(Workspace, data)

    def list_members(self, workspace_id: int) -> list[Member]:
        data = self._http.get_json(f"/workspaces/{workspace_id}/members")
        return _to(Member, data)

    # ------------------------------------------------------------- Projects

    def list_projects(self, workspace_id: int) -> list[Project]:
        data = self._http.get_json("/projects", {"workspace_id": workspace_id})
        return _to(Project, data)

    def get_project(self, project_id: int) -> Project:
        data = self._http.get_json(f"/projects/{project_id}")
        return _to(Project, data)

    def create_project(self, workspace_id: int, name: str, identifier: str, description: str = "") -> Project:
        body: dict[str, Any] = {"name": name, "identifier": identifier}
        if description:
            body["description"] = description
        data = self._http.post_json("/projects", body, {"workspace_id": workspace_id})
        return _to(Project, data)

    # --------------------------------------------------------------- Issues

    def list_issues(
        self,
        *,
        project_id: int = 0,
        workspace_id: int = 0,
        rql: str = "",
        state_id: int = 0,
        priority: str = "",
        assignee_id: int = 0,
        cycle_id: int = 0,
        issue_type_id: int = 0,
        search: str = "",
        sort_by: str = "",
        sort_dir: str = "",
        limit: int = 0,
        offset: int = 0,
    ) -> IssueListResult:
        query: dict[str, Any] = {}
        if project_id: query["project_id"] = project_id
        if workspace_id: query["workspace_id"] = workspace_id
        if rql: query["rql"] = rql
        if state_id: query["state_id"] = state_id
        if priority: query["priority"] = priority
        if assignee_id: query["assignee_id"] = assignee_id
        if cycle_id: query["cycle_id"] = cycle_id
        if issue_type_id: query["issue_type_id"] = issue_type_id
        if search: query["search"] = search
        if sort_by: query["sort_by"] = sort_by
        if sort_dir: query["sort_dir"] = sort_dir
        if limit: query["limit"] = limit
        if offset: query["offset"] = offset

        # ListIssues returns bare array + X-Total-Count header.
        # We need the header, so use the raw HTTP client.
        url = f"{self._http.base_url}/issues"
        import httpx
        resp = self._http._client.request("GET", url, params={k: v for k, v in query.items() if v})
        if resp.status_code < 200 or resp.status_code >= 300:
            body_data: dict[str, Any] = {}
            try:
                body_data = resp.json()
            except Exception:
                pass
            raise APIError(resp.status_code, body_data.get("message", resp.text.strip()), body_data)

        items = [_to(Issue, item) for item in resp.json()] if resp.content else []
        total = int(resp.headers.get("X-Total-Count", "0"))
        return IssueListResult(items=items, total=total)

    def create_issue(self, project_id: int, workspace_id: int, **kwargs: Any) -> Issue:
        data = self._http.post_json("/issues", kwargs, {"project_id": project_id, "workspace_id": workspace_id})
        return _to(Issue, data)

    def get_issue(self, issue_id: int) -> Issue:
        data = self._http.get_json(f"/issues/{issue_id}")
        return _to(Issue, data)

    def update_issue(self, issue_id: int, **kwargs: Any) -> Issue:
        data = self._http.put_json(f"/issues/{issue_id}", kwargs)
        return _to(Issue, data)

    def search_issues(
        self, workspace_id: int, query: str, *, project_id: int | None = None, limit: int = 10
    ) -> list[IssueSearchResult]:
        params: dict[str, Any] = {"workspace_id": workspace_id, "query": query, "limit": limit}
        if project_id:
            params["project_id"] = project_id
        data = self._http.get_json("/issues/search", params)
        return _to(IssueSearchResult, data)

    def resolve_issue_code(self, workspace_id: int, code: str) -> int:
        """Resolve "DEMO-42" to a numeric issue ID."""
        parts = code.split("-", 1)
        if len(parts) != 2:
            raise ValueError(f"invalid issue code {code!r} (expected IDENTIFIER-NUMBER)")
        identifier, seq_str = parts
        seq = int(seq_str)

        projects = self.list_projects(workspace_id)
        project_id = 0
        for p in projects:
            if p.identifier.upper() == identifier.upper():
                project_id = p.id
                break
        if not project_id:
            raise ValueError(f"project with identifier {identifier!r} not found")

        result = self.list_issues(project_id=project_id, search=seq_str, limit=100)
        for item in result.items:
            if item.sequence_id == seq:
                return item.id
        raise ValueError(f"issue {code} not found")

    def add_comment(self, issue_id: int, body: str, parent_id: int | None = None) -> Comment:
        req: dict[str, Any] = {"issue_id": issue_id, "body": body}
        if parent_id:
            req["parent_id"] = parent_id
        data = self._http.post_json("/comments", req)
        return _to(Comment, data)

    def list_comments(self, issue_id: int, page: int = 1, page_size: int = 20) -> tuple[list[Comment], int]:
        query: dict[str, Any] = {}
        if page > 1: query["page"] = page
        if page_size != 20: query["page_size"] = page_size
        data = self._http.get_json(f"/comments/issue/{issue_id}", query)
        comments = [_to(Comment, c) for c in data.get("comments", [])]
        total = data.get("total", 0)
        return comments, total

    # ---------------------------------------------------------------- Cycles

    def list_cycles(self, project_id: int, *, status: str = "", limit: int = 50, offset: int = 0) -> CycleListResult:
        query: dict[str, Any] = {}
        if status: query["status"] = status
        if limit: query["limit"] = limit
        if offset: query["offset"] = offset
        data = self._http.get_json(f"/projects/{project_id}/cycles", query)
        return _to(CycleListResult, data)

    def get_cycle(self, cycle_id: int) -> Cycle:
        data = self._http.get_json(f"/cycles/{cycle_id}")
        return _to(Cycle, data)

    def get_cycle_progress(self, cycle_id: int) -> CycleProgress:
        data = self._http.get_json(f"/cycles/{cycle_id}/progress")
        return _to(CycleProgress, data)

    def get_cycle_burndown(self, cycle_id: int) -> BurndownData:
        data = self._http.get_json(f"/cycles/{cycle_id}/burndown")
        return _to(BurndownData, data)

    def add_issue_to_cycle(self, cycle_id: int, issue_id: int) -> None:
        self._http.post_json(f"/cycles/{cycle_id}/issues", query={"issue_id": issue_id})

    # ----------------------------------------------------------------- Meta

    def list_states(self, project_id: int) -> list[State]:
        data = self._http.get_json(f"/projects/{project_id}/settings/states")
        return _to(State, data)

    def list_labels(self, project_id: int) -> list[Label]:
        data = self._http.get_json(f"/projects/{project_id}/settings/labels")
        return _to(Label, data)

    def list_issue_types(self, workspace_id: int, project_id: int = 0) -> list[IssueType]:
        query: dict[str, Any] = {"workspace_id": workspace_id}
        if project_id: query["project_id"] = project_id
        data = self._http.get_json("/issue-types", query)
        return _to(IssueType, data)

    def list_pages(self, project_id: int) -> list[Page]:
        data = self._http.get_json(f"/projects/{project_id}/pages")
        return _to(Page, data)

    def get_page(self, project_id: int, page_id: int) -> Page:
        data = self._http.get_json(f"/projects/{project_id}/pages/{page_id}")
        return _to(Page, data)

    def list_notifications(self, *, unread_only: bool = False, limit: int = 0, offset: int = 0) -> list[Notification]:
        query: dict[str, Any] = {}
        if unread_only: query["unread_only"] = "true"
        if limit: query["limit"] = limit
        if offset: query["offset"] = offset
        data = self._http.get_json("/notifications", query)
        return _to(Notification, data)

    # ------------------------------------------------------------------- AI

    def ai_search(self, project_id: int, query: str) -> AISearchResponse:
        data = self._http.post_json(f"/projects/{project_id}/ai/search", {"query": query})
        return _to(AISearchResponse, data)

    def ai_chat(self, project_id: int, message: str, thread_id: int | None = None) -> AIChatReply:
        body: dict[str, Any] = {"message": message}
        if thread_id:
            body["thread_id"] = thread_id
        # SSE stream — use raw httpx with stream
        import httpx as _httpx
        url = f"{self._http.base_url}/projects/{project_id}/ai/chat"
        with _httpx.Client(timeout=300.0) as hc:
            req = hc.build_request("POST", url, json=body, headers={"Authorization": f"Bearer {self._http._token}"})
            resp = hc.send(req, stream=True)
            if resp.status_code < 200 or resp.status_code >= 300:
                raise APIError(resp.status_code, resp.text.strip())

            reply = AIChatReply()
            for line in resp.iter_lines():
                line = line.strip()
                if not line or not line.startswith("data: "):
                    continue
                import json
                try:
                    ev = json.loads(line[6:])
                except json.JSONDecodeError:
                    continue
                match ev.get("type"):
                    case "text" | "thinking":
                        reply.text += ev.get("content", "")
                    case "tool_call":
                        tc = ev.get("tool_call", {})
                        reply.tool_calls.append(f"{tc.get('name', '')}({tc.get('arguments', '')})")
                    case "error":
                        raise APIError(502, ev.get("error", "stream error"))
                if ev.get("thread_id"):
                    reply.thread_id = ev["thread_id"]
            return reply

    # --------------------------------------------------------------- Agents

    def list_agents(self, workspace_id: int) -> list[Agent]:
        data = self._http.get_json(f"/workspaces/{workspace_id}/agents")
        return _to(Agent, data)

    def dispatch_agent(
        self, workspace_id: int, agent_id: int, task: str, *, issue_id: int | None = None
    ) -> AgentActivity:
        body: dict[str, Any] = {"task": task}
        if issue_id:
            body["issue_id"] = issue_id
        data = self._http.post_json(
            f"/workspaces/{workspace_id}/agents/{agent_id}/dispatch", body, timeout=300.0
        )
        return _to(AgentActivity, data)

    def get_agent_task(self, workspace_id: int, task_id: int) -> AgentTask:
        data = self._http.get_json(f"/workspaces/{workspace_id}/agent-tasks/{task_id}")
        return _to(AgentTask, data)

    def get_agent_task_logs(self, workspace_id: int, task_id: int) -> list[TaskLog]:
        data = self._http.get_json(f"/workspaces/{workspace_id}/agent-tasks/{task_id}/logs")
        return _to(TaskLog, data)

    # ----------------------------------------------------------------- Chat

    def get_issue_chat(self, issue_id: int) -> Chat:
        data = self._http.get_json(f"/issues/{issue_id}/chat")
        return _to(Chat, data)

    def send_message(self, chat_id: int, content: str) -> Message:
        data = self._http.post_json(f"/chats/{chat_id}/messages", {"content": content})
        return _to(Message, data)
