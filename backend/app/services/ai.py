"""
AI Service - Plane AI 核心服务
实现类似 Plane AI 的功能：
1. 工作区上下文感知 - AI 可以读取整个工作区的项目、问题、周期等数据
2. AI Agent - 自动分类、分配、跟踪阻塞等
3. 自然语言转 PQL - 将自然语言转换为查询语言
"""
from typing import Optional, List, Dict, Any, AsyncGenerator
from datetime import datetime
import json
from enum import Enum

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload

from app.core.config import settings
from app.models.issue import Issue
from app.models.project import Project
from app.models.workspace import Workspace
from app.models.cycle import Cycle
from app.models.module import Module
from app.models.state import State
from app.models.user import User
from app.schemas.ai import AIRequest, AIResponse, AIIntent, AIPlan, AIAction, AIMode


class AIContextAggregator:
    """AI 上下文聚合器 - 收集工作区数据作为 AI 上下文"""
    
    def __init__(self, db: AsyncSession):
        self.db = db
    
    async def get_workspace_context(
        self, 
        workspace_id: int, 
        project_id: Optional[int] = None,
        include_issues: bool = True,
        include_cycles: bool = True,
        include_modules: bool = True,
        issue_limit: int = 50
    ) -> Dict[str, Any]:
        """获取工作区完整上下文"""
        context = {
            "workspace": await self._get_workspace(workspace_id),
            "projects": [],
            "issues": [],
            "cycles": [],
            "modules": [],
            "states": [],
            "users": [],
            "summary": {}
        }
        
        # 获取项目
        if project_id:
            context["projects"] = [await self._get_project(project_id)]
        else:
            context["projects"] = await self._get_workspace_projects(workspace_id)
        
        # 获取问题
        if include_issues:
            context["issues"] = await self._get_issues(
                workspace_id, 
                project_id, 
                limit=issue_limit
            )
        
        # 获取周期
        if include_cycles:
            context["cycles"] = await self._get_cycles(workspace_id, project_id)
        
        # 获取模块
        if include_modules:
            context["modules"] = await self._get_modules(workspace_id, project_id)
        
        # 获取状态
        context["states"] = await self._get_states(workspace_id)
        
        # 获取用户
        context["users"] = await self._get_workspace_users(workspace_id)
        
        # 生成摘要
        context["summary"] = self._generate_summary(context)
        
        return context
    
    async def _get_workspace(self, workspace_id: int) -> Dict[str, Any]:
        workspace = await self.db.get(Workspace, workspace_id)
        if not workspace:
            return {}
        return {
            "id": workspace.id,
            "name": workspace.name,
            "slug": workspace.slug,
            "timezone": workspace.timezone
        }
    
    async def _get_workspace_projects(self, workspace_id: int) -> List[Dict[str, Any]]:
        result = await self.db.execute(
            select(Project).where(
                Project.workspace_id == workspace_id,
                Project.is_deleted == False
            )
        )
        projects = result.scalars().all()
        return [
            {
                "id": p.id,
                "name": p.name,
                "identifier": p.identifier,
                "description": p.description,
                "is_public": p.is_public
            }
            for p in projects
        ]
    
    async def _get_project(self, project_id: int) -> Dict[str, Any]:
        project = await self.db.get(Project, project_id)
        if not project:
            return {}
        return {
            "id": project.id,
            "name": project.name,
            "identifier": project.identifier,
            "description": project.description,
            "is_public": project.is_public
        }
    
    async def _get_issues(
        self, 
        workspace_id: int, 
        project_id: Optional[int] = None,
        limit: int = 50
    ) -> List[Dict[str, Any]]:
        query = select(Issue).where(
            Issue.workspace_id == workspace_id,
            Issue.is_deleted == False
        ).options(selectinload(Issue.state))
        
        if project_id:
            query = query.where(Issue.project_id == project_id)
        
        query = query.order_by(Issue.created_at.desc()).limit(limit)
        result = await self.db.execute(query)
        issues = result.scalars().all()
        
        return [
            {
                "id": i.id,
                "name": i.name,
                "priority": i.priority,
                "state": i.state.name if i.state else None,
                "state_group": i.state.group if i.state else None,
                "start_date": str(i.start_date) if i.start_date else None,
                "target_date": str(i.target_date) if i.target_date else None,
                "is_draft": i.is_draft,
                "project_id": i.project_id,
                "parent_id": i.parent_id if i.parent_id else None,
                "created_at": str(i.created_at),
                "completed_at": str(i.completed_at) if i.completed_at else None
            }
            for i in issues
        ]
    
    async def _get_cycles(
        self, 
        workspace_id: int, 
        project_id: Optional[int] = None
    ) -> List[Dict[str, Any]]:
        query = select(Cycle).where(
            Cycle.workspace_id == workspace_id,
            Cycle.is_deleted == False
        )
        if project_id:
            query = query.where(Cycle.project_id == project_id)
        
        result = await self.db.execute(query)
        cycles = result.scalars().all()
        
        return [
            {
                "id": c.id,
                "name": c.name,
                "description": c.description,
                "start_date": str(c.start_date) if c.start_date else None,
                "end_date": str(c.end_date) if c.end_date else None,
                "project_id": c.project_id
            }
            for c in cycles
        ]
    
    async def _get_modules(
        self, 
        workspace_id: int, 
        project_id: Optional[int] = None
    ) -> List[Dict[str, Any]]:
        query = select(Module).where(
            Module.workspace_id == workspace_id,
            Module.is_deleted == False
        )
        if project_id:
            query = query.where(Module.project_id == project_id)
        
        result = await self.db.execute(query)
        modules = result.scalars().all()
        
        return [
            {
                "id": m.id,
                "name": m.name,
                "description": m.description,
                "project_id": m.project_id
            }
            for m in modules
        ]
    
    async def _get_states(self, workspace_id: int) -> List[Dict[str, Any]]:
        result = await self.db.execute(
            select(State).where(State.workspace_id == workspace_id)
        )
        states = result.scalars().all()
        
        return [
            {
                "id": s.id,
                "name": s.name,
                "group": s.group,
                "color": s.color
            }
            for s in states
        ]
    
    async def _get_workspace_users(self, workspace_id: int) -> List[Dict[str, Any]]:
        from app.models.workspace import WorkspaceMember
        result = await self.db.execute(
            select(WorkspaceMember).where(
                WorkspaceMember.workspace_id == workspace_id,
                WorkspaceMember.is_active == True
            ).options(selectinload(WorkspaceMember.user))
        )
        members = result.scalars().all()
        
        return [
            {
                "id": m.user.id,
                "name": m.user.display_name or m.user.email,
                "email": m.user.email,
                "role": m.role
            }
            for m in members if m.user
        ]
    
    def _generate_summary(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """生成工作区摘要"""
        issues = context.get("issues", [])
        states = context.get("states", [])
        
        # 按状态分组统计
        status_counts = {}
        for issue in issues:
            state = issue.get("state", "Unknown")
            status_counts[state] = status_counts.get(state, 0) + 1
        
        # 按优先级分组统计
        priority_counts = {}
        for issue in issues:
            priority = issue.get("priority", "none")
            priority_counts[priority] = priority_counts.get(priority, 0) + 1
        
        # 计算完成率
        completed_states = ["completed", "cancelled"]
        completed_count = sum(
            count for state, count in status_counts.items()
            if state and state.lower() in completed_states
        )
        total_count = len(issues)
        
        return {
            "total_issues": total_count,
            "total_projects": len(context.get("projects", [])),
            "total_cycles": len(context.get("cycles", [])),
            "total_modules": len(context.get("modules", [])),
            "total_users": len(context.get("users", [])),
            "status_distribution": status_counts,
            "priority_distribution": priority_counts,
            "completion_rate": completed_count / total_count if total_count > 0 else 0
        }


class AIAgent:
    """AI Agent - 执行具体任务的代理"""
    
    def __init__(self, db: AsyncSession, context: Dict[str, Any]):
        self.db = db
        self.context = context
    
    async def triage_issue(self, issue_id: int) -> AIAction:
        """自动分类问题"""
        issue = await self.db.get(Issue, issue_id)
        if not issue:
            return AIAction(
                action_type="triage",
                target_type="issue",
                target_id=issue_id,
                changes={},
                description="Issue not found"
            )
        
        # 基于上下文分析问题
        analysis = self._analyze_issue_content(issue)
        
        return AIAction(
            action_type="triage",
            target_type="issue",
            target_id=issue_id,
            changes=analysis.get("suggestions", {}),
            description=analysis.get("reasoning", "Auto-triaged based on content analysis")
        )
    
    async def suggest_assignee(self, issue_id: int) -> AIAction:
        """建议分配人员"""
        issue = await self.db.get(Issue, issue_id)
        if not issue:
            return AIAction(
                action_type="assign",
                target_type="issue",
                target_id=issue_id,
                changes={},
                description="Issue not found"
            )
        
        # 基于工作负载和技能匹配建议分配人员
        users = self.context.get("users", [])
        if not users:
            return AIAction(
                action_type="assign",
                target_type="issue",
                target_id=issue_id,
                changes={},
                description="No users available for assignment"
            )
        
        # 简单的工作负载均衡
        issues = self.context.get("issues", [])
        workload = {}
        for i in issues:
            # 统计每个用户的未完成任务数
            pass  # TODO: 实现工作负载计算
        
        # 选择工作负载最低的用户
        suggested_user = users[0] if users else None
        
        return AIAction(
            action_type="assign",
            target_type="issue",
            target_id=issue_id,
            changes={"assignee_id": suggested_user["id"]} if suggested_user else {},
            description=f"Suggested assignee: {suggested_user['name']}" if suggested_user else "No suggestion available"
        )
    
    async def track_blockers(self, project_id: Optional[int] = None) -> List[Dict[str, Any]]:
        """跟踪阻塞项"""
        issues = self.context.get("issues", [])
        blockers = []
        
        for issue in issues:
            # 检测可能的阻塞条件
            if issue.get("state_group") == "backlog":
                # 检查是否有长期未处理的backlog问题
                created_at = datetime.fromisoformat(issue.get("created_at", ""))
                days_old = (datetime.now() - created_at).days
                if days_old > 30:
                    blockers.append({
                        "type": "stale_backlog",
                        "issue_id": issue["id"],
                        "issue_name": issue["name"],
                        "severity": "medium",
                        "description": f"Issue has been in backlog for {days_old} days"
                    })
            
            if issue.get("target_date"):
                # 检查是否有即将到期但未完成的问题
                target_date = datetime.fromisoformat(issue["target_date"])
                days_until_due = (target_date - datetime.now()).days
                if days_until_due < 3 and issue.get("state_group") not in ["completed", "cancelled"]:
                    blockers.append({
                        "type": "approaching_deadline",
                        "issue_id": issue["id"],
                        "issue_name": issue["name"],
                        "severity": "high",
                        "description": f"Issue due in {days_until_due} days"
                    })
        
        return blockers
    
    def _analyze_issue_content(self, issue: Issue) -> Dict[str, Any]:
        """分析问题内容"""
        suggestions = {}
        reasoning = []
        
        # 基于名称和描述分析优先级
        name_lower = issue.name.lower()
        description = issue.description_stripped or ""
        
        # 检测紧急关键词
        urgent_keywords = ["urgent", "critical", "blocker", "asap", "紧急", "严重"]
        if any(kw in name_lower or kw in description.lower() for kw in urgent_keywords):
            suggestions["priority"] = "urgent"
            reasoning.append("Detected urgent keywords in issue")
        
        # 检测 Bug 关键词
        bug_keywords = ["bug", "fix", "error", "crash", "问题", "错误"]
        if any(kw in name_lower or kw in description.lower() for kw in bug_keywords):
            suggestions["type"] = "bug"
            reasoning.append("Detected bug-related keywords")
        
        return {
            "suggestions": suggestions,
            "reasoning": ". ".join(reasoning) if reasoning else "No specific suggestions"
        }


class PQLGenerator:
    """PQL (Plane Query Language) 生成器 - 将自然语言转换为查询"""
    
    # PQL 操作符映射
    OPERATORS = {
        "equals": "=",
        "not_equals": "!=",
        "in": "IN",
        "not_in": "NOT IN",
        "contains": "~",
        "greater_than": ">",
        "less_than": "<",
        "between": "BETWEEN",
        "is_null": "IS NULL",
        "is_not_null": "IS NOT NULL"
    }
    
    # 字段映射
    FIELDS = {
        "priority": "priority",
        "state": "state",
        "status": "state",
        "assignee": "assignee",
        "label": "label",
        "project": "project",
        "cycle": "cycle",
        "module": "module",
        "type": "type",
        "due": "dueDate",
        "due_date": "dueDate",
        "start": "startDate",
        "start_date": "startDate",
        "created": "createdAt",
        "updated": "updatedAt",
        "title": "title",
        "name": "title"
    }
    
    # 内置函数
    FUNCTIONS = [
        "isOverdue",
        "hasNoAssignee", 
        "hasNoLabel",
        "isTopLevel",
        "isSubWorkItem",
        "hasChildren",
        "hasStartAndDueDates"
    ]
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
    
    def generate_from_natural_language(self, query: str) -> str:
        """将自然语言转换为 PQL 查询"""
        query_lower = query.lower()
        
        # 检测意图
        if "overdue" in query_lower or "过期" in query_lower or "逾期" in query_lower:
            return "isOverdue()"
        
        if "unassigned" in query_lower or "未分配" in query_lower or "无负责人" in query_lower:
            return "hasNoAssignee()"
        
        if "no label" in query_lower or "无标签" in query_lower:
            return "hasNoLabel()"
        
        # 解析优先级
        priority_pql = self._parse_priority(query_lower)
        if priority_pql:
            return priority_pql
        
        # 解析状态
        state_pql = self._parse_state(query_lower)
        if state_pql:
            return state_pql
        
        # 解析项目
        project_pql = self._parse_project(query_lower)
        if project_pql:
            return project_pql
        
        # 默认返回标题搜索
        return f'title ~ "{query}"'
    
    def _parse_priority(self, query: str) -> Optional[str]:
        """解析优先级查询"""
        priority_map = {
            "urgent": "urgent",
            "critical": "urgent",
            "紧急": "urgent",
            "严重": "urgent",
            "high": "high",
            "高": "high",
            "medium": "medium",
            "中": "medium",
            "low": "low",
            "低": "low"
        }
        
        for keyword, priority in priority_map.items():
            if keyword in query:
                return f'priority = {priority.upper()}'
        
        return None
    
    def _parse_state(self, query: str) -> Optional[str]:
        """解析状态查询"""
        state_map = {
            "todo": "Todo",
            "backlog": "Backlog",
            "in progress": "In Progress",
            "进行中": "In Progress",
            "done": "Done",
            "completed": "Done",
            "完成": "Done",
            "cancelled": "Cancelled",
            "取消": "Cancelled"
        }
        
        for keyword, state in state_map.items():
            if keyword in query:
                return f'state = "{state}"'
        
        return None
    
    def _parse_project(self, query: str) -> Optional[str]:
        """解析项目查询"""
        projects = self.context.get("projects", [])
        
        for project in projects:
            project_name = project.get("name", "").lower()
            project_identifier = project.get("identifier", "").lower()
            
            if project_name in query or project_identifier in query:
                return f'project = "{project["name"]}"'
        
        return None


class AIService:
    """AI 服务主类"""
    
    def __init__(self, db: AsyncSession):
        self.db = db
        self.context_aggregator = AIContextAggregator(db)
        self.pql_generator = None  # 延迟初始化
    
    async def process_request(self, request: AIRequest, user_id: int) -> AIResponse:
        """处理 AI 请求"""
        # 获取上下文
        context = await self.context_aggregator.get_workspace_context(
            workspace_id=request.workspace_id,
            project_id=request.project_id
        )
        
        # 初始化 PQL 生成器
        self.pql_generator = PQLGenerator(context)
        
        # 初始化 Agent
        agent = AIAgent(self.db, context)
        
        # 分析意图
        intent = self._analyze_intent(request.message.content)
        
        # 根据模式处理
        if request.message.mode == AIMode.ASK:
            return await self._handle_ask_mode(request, context, intent)
        elif request.message.mode == AIMode.BUILD:
            return await self._handle_build_mode(request, context, intent, agent)
        
        return AIResponse(
            content="Unknown mode",
            intent=intent,
            thread_id=request.thread_id or 0
        )
    
    def _analyze_intent(self, content: str) -> AIIntent:
        """分析用户意图"""
        content_lower = content.lower()
        
        # 创建意图
        create_keywords = ["create", "add", "new", "创建", "添加", "新建"]
        if any(kw in content_lower for kw in create_keywords):
            return AIIntent.CREATE
        
        # 更新意图
        update_keywords = ["update", "change", "modify", "assign", "更新", "修改", "分配"]
        if any(kw in content_lower for kw in update_keywords):
            return AIIntent.UPDATE
        
        # 分析意图
        analyze_keywords = ["analyze", "report", "summary", "status", "分析", "报告", "摘要", "状态"]
        if any(kw in content_lower for kw in analyze_keywords):
            return AIIntent.ANALYZE
        
        # 搜索意图
        search_keywords = ["find", "search", "show", "list", "查找", "搜索", "显示", "列出"]
        if any(kw in content_lower for kw in search_keywords):
            return AIIntent.SEARCH
        
        return AIIntent.HELP
    
    async def _handle_ask_mode(
        self, 
        request: AIRequest, 
        context: Dict[str, Any], 
        intent: AIIntent
    ) -> AIResponse:
        """处理 ASK 模式 - 回答问题"""
        content = request.message.content
        summary = context.get("summary", {})
        
        # 生成 PQL 查询
        pql_query = self.pql_generator.generate_from_natural_language(content)
        
        # 构建响应
        response_content = self._generate_response(content, context, intent)
        
        return AIResponse(
            content=response_content,
            intent=intent,
            results=[{"pql": pql_query}] if intent == AIIntent.SEARCH else None,
            suggestions=self._generate_suggestions(context),
            thread_id=request.thread_id or 0
        )
    
    async def _handle_build_mode(
        self,
        request: AIRequest,
        context: Dict[str, Any],
        intent: AIIntent,
        agent: AIAgent
    ) -> AIResponse:
        """处理 BUILD 模式 - 执行操作"""
        content = request.message.content
        
        # 创建操作计划
        plan = await self._create_action_plan(content, context, agent)
        
        return AIResponse(
            content="I've analyzed your request and prepared an action plan.",
            intent=intent,
            plan=plan,
            suggestions=self._generate_suggestions(context),
            thread_id=request.thread_id or 0
        )
    
    async def _create_action_plan(
        self,
        content: str,
        context: Dict[str, Any],
        agent: AIAgent
    ) -> AIPlan:
        """创建操作计划"""
        actions = []
        content_lower = content.lower()
        
        # 检测是否需要跟踪阻塞
        if "blocker" in content_lower or "阻塞" in content_lower or "障碍" in content_lower:
            blockers = await agent.track_blockers()
            for blocker in blockers:
                actions.append(AIAction(
                    action_type="identify_blocker",
                    target_type="issue",
                    target_id=int(blocker["issue_id"]),
                    changes={"blocker_type": blocker["type"], "severity": blocker["severity"]},
                    description=blocker["description"]
                ))
        
        # 检测是否需要分类
        if "triage" in content_lower or "分类" in content_lower:
            issues = context.get("issues", [])
            for issue in issues[:5]:  # 限制处理数量
                action = await agent.triage_issue(int(issue["id"]))
                actions.append(action)
        
        return AIPlan(
            actions=actions,
            requires_confirmation=True,
            estimated_impact=f"Will affect {len(actions)} items"
        )
    
    def _generate_response(
        self, 
        content: str, 
        context: Dict[str, Any],
        intent: AIIntent
    ) -> str:
        """生成响应内容"""
        summary = context.get("summary", {})
        
        if intent == AIIntent.ANALYZE:
            return self._format_analysis(summary, context)
        elif intent == AIIntent.SEARCH:
            return self._format_search_results(context)
        elif intent == AIIntent.HELP:
            return self._format_help()
        else:
            return f"Based on your workspace, you have {summary.get('total_issues', 0)} issues across {summary.get('total_projects', 0)} projects."
    
    def _format_analysis(self, summary: Dict[str, Any], context: Dict[str, Any]) -> str:
        """格式化分析结果"""
        lines = [
            "## Workspace Analysis",
            "",
            f"**Total Issues:** {summary.get('total_issues', 0)}",
            f"**Total Projects:** {summary.get('total_projects', 0)}",
            f"**Active Cycles:** {summary.get('total_cycles', 0)}",
            f"**Active Modules:** {summary.get('total_modules', 0)}",
            f"**Team Members:** {summary.get('total_users', 0)}",
            "",
            "**Status Distribution:**"
        ]
        
        for status, count in summary.get("status_distribution", {}).items():
            lines.append(f"  - {status}: {count}")
        
        lines.append("")
        lines.append("**Priority Distribution:**")
        
        for priority, count in summary.get("priority_distribution", {}).items():
            lines.append(f"  - {priority}: {count}")
        
        lines.append("")
        lines.append(f"**Completion Rate:** {summary.get('completion_rate', 0):.1%}")
        
        return "\n".join(lines)
    
    def _format_search_results(self, context: Dict[str, Any]) -> str:
        """格式化搜索结果"""
        issues = context.get("issues", [])
        
        if not issues:
            return "No issues found matching your criteria."
        
        lines = [f"Found {len(issues)} issues:", ""]
        
        for issue in issues[:10]:  # 限制显示数量
            lines.append(f"- **{issue.get('name', 'Unknown')}**")
            lines.append(f"  Priority: {issue.get('priority', 'none')}, State: {issue.get('state', 'Unknown')}")
        
        if len(issues) > 10:
            lines.append(f"\n... and {len(issues) - 10} more")
        
        return "\n".join(lines)
    
    def _format_help(self) -> str:
        """格式化帮助信息"""
        return """
## AI Assistant Help

I can help you with:

**Ask Mode:**
- Search for issues: "Show me high priority bugs"
- Analyze workspace: "What's the status of my project?"
- Generate queries: "Find overdue issues"

**Build Mode:**
- Triage issues: "Triage all new issues"
- Track blockers: "Identify blockers in the project"
- Assign work: "Suggest assignees for unassigned issues"

**PQL Examples:**
- `priority = HIGH`
- `state IN (Todo, "In Progress")`
- `isOverdue()`
- `hasNoAssignee()`
"""
    
    def _generate_suggestions(self, context: Dict[str, Any]) -> List[str]:
        """生成建议"""
        suggestions = []
        summary = context.get("summary", {})
        
        # 基于工作区状态生成建议
        if summary.get("total_issues", 0) > 0:
            completion_rate = summary.get("completion_rate", 0)
            if completion_rate < 0.3:
                suggestions.append("Consider reviewing stalled issues - completion rate is below 30%")
            
            priority_dist = summary.get("priority_distribution", {})
            if priority_dist.get("urgent", 0) > 0:
                suggestions.append(f"You have {priority_dist['urgent']} urgent issues that need attention")
        
        if summary.get("total_cycles", 0) == 0:
            suggestions.append("Consider creating a cycle to organize your work")
        
        return suggestions[:5]  # 最多返回5个建议