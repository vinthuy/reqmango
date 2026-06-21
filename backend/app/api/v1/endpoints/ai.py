"""
AI API Endpoints - Plane AI 对话和操作接口
"""
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from typing import Optional
from datetime import datetime

from app.db.session import get_db
from app.api.deps import get_current_user, require_workspace_access
from app.models.user import User
from app.models.ai_thread import AIThread, AIMessage, AIActionLog
from app.schemas.ai import (
    AIRequest, AIResponse, AIMessage as AIMessageSchema, AIMode, AIIntent,
    AIPlan, AIAction, AIThread as AIThreadSchema,
    TaskExtractionRequest, TaskExtractionResponse,
    TaskCreateFromNLPRequest, TaskCreateFromNLPResponse
)
from app.services.ai import AIService, AIContextAggregator, PQLGenerator
from app.services.nlp_parser import NLPParser, TaskExtraction

router = APIRouter()


# ==================== AI Thread Management ====================

@router.post("/threads", response_model=dict)
async def create_thread(
    workspace_id: int,
    project_id: Optional[int] = None,
    title: str = "New Conversation",
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建新的 AI 对话线程
    """
    await require_workspace_access(workspace_id, current_user, db)
    
    thread = AIThread(
        title=title,
        workspace_id=workspace_id,
        project_id=project_id,
        user_id=current_user.id,
        created_by_id=current_user.id
    )
    
    db.add(thread)
    await db.commit()
    await db.refresh(thread)
    
    return {
        "id": thread.id,
        "title": thread.title,
        "workspace_id": thread.workspace_id,
        "project_id": thread.project_id if thread.project_id else None,
        "created_at": str(thread.created_at)
    }


@router.get("/threads/{thread_id}", response_model=dict)
async def get_thread(
    thread_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取 AI 对话线程详情
    """
    thread = await db.get(AIThread, thread_id)
    if not thread:
        raise HTTPException(status_code=404, detail="Thread not found")
    
    # 验证用户权限
    await require_workspace_access(thread.workspace_id, current_user, db)
    
    # 获取消息
    result = await db.execute(
        select(AIMessage).where(AIMessage.thread_id == thread_id).order_by(AIMessage.created_at)
    )
    messages = result.scalars().all()
    
    return {
        "id": thread.id,
        "title": thread.title,
        "summary": thread.summary,
        "workspace_id": thread.workspace_id,
        "project_id": thread.project_id if thread.project_id else None,
        "is_archived": thread.is_archived,
        "created_at": str(thread.created_at),
        "updated_at": str(thread.updated_at),
        "messages": [
            {
                "id": m.id,
                "role": m.role,
                "content": m.content,
                "mode": m.mode,
                "intent": m.intent,
                "created_at": str(m.created_at),
                "is_executed": m.is_executed
            }
            for m in messages
        ]
    }


@router.get("/threads/workspace/{workspace_id}", response_model=list)
async def list_threads(
    workspace_id: int,
    project_id: Optional[int] = None,
    is_archived: bool = False,
    limit: int = 20,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出工作区的 AI 对话线程
    """
    await require_workspace_access(workspace_id, current_user, db)
    
    query = select(AIThread).where(
        AIThread.workspace_id == workspace_id,
        AIThread.user_id == current_user.id,
        AIThread.is_archived == is_archived
    )
    
    if project_id:
        query = query.where(AIThread.project_id == project_id)
    
    query = query.order_by(AIThread.updated_at.desc()).limit(limit)
    result = await db.execute(query)
    threads = result.scalars().all()
    
    return [
        {
            "id": t.id,
            "title": t.title,
            "summary": t.summary,
            "project_id": t.project_id if t.project_id else None,
            "created_at": str(t.created_at),
            "updated_at": str(t.updated_at)
        }
        for t in threads
    ]


@router.put("/threads/{thread_id}", response_model=dict)
async def update_thread(
    thread_id: int,
    title: Optional[str] = None,
    is_archived: Optional[bool] = None,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新 AI 对话线程
    """
    thread = await db.get(AIThread, thread_id)
    if not thread:
        raise HTTPException(status_code=404, detail="Thread not found")
    
    await require_workspace_access(thread.workspace_id, current_user, db)
    
    if title is not None:
        thread.title = title
    if is_archived is not None:
        thread.is_archived = is_archived
    
    await db.commit()
    await db.refresh(thread)
    
    return {
        "id": thread.id,
        "title": thread.title,
        "is_archived": thread.is_archived,
        "updated_at": str(thread.updated_at)
    }


@router.delete("/threads/{thread_id}")
async def delete_thread(
    thread_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除 AI 对话线程
    """
    thread = await db.get(AIThread, thread_id)
    if not thread:
        raise HTTPException(status_code=404, detail="Thread not found")
    
    await require_workspace_access(thread.workspace_id, current_user, db)
    
    await db.delete(thread)
    await db.commit()
    
    return {"message": "Thread deleted successfully"}


# ==================== AI Chat ====================

@router.post("/chat", response_model=AIResponse)
async def ai_chat(
    request: AIRequest,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    AI 对话接口 - 处理用户的 AI 请求
    
    支持两种模式:
    - ASK: 回答问题、搜索、分析
    - BUILD: 执行操作、创建计划
    
    类似 Plane AI 的功能:
    - 工作区上下文感知
    - 自然语言转 PQL
    - AI Agent 操作
    """
    ai_service = AIService(db)
    
    # 验证工作区访问权限
    if request.workspace_id:
        await require_workspace_access(request.workspace_id, current_user, db)
    
    # 如果有 thread_id，保存消息到线程
    if request.thread_id:
        thread = await db.get(AIThread, request.thread_id)
        if thread:
            # 保存用户消息
            user_message = AIMessage(
                thread_id=request.thread_id,
                role="user",
                content=request.message.content,
                mode=request.message.mode.value,
                created_by_id=current_user.id
            )
            db.add(user_message)
    
    response = await ai_service.process_request(request, current_user.id)
    
    # 保存 AI 响应
    if request.thread_id:
        ai_message = AIMessage(
            thread_id=request.thread_id,
            role="assistant",
            content=response.content,
            mode=request.message.mode.value,
            intent=response.intent.value if response.intent else None,
            plan=response.plan.model_dump() if response.plan else None,
            results=response.results,
            suggestions=response.suggestions,
            created_by_id=current_user.id
        )
        db.add(ai_message)
        
        # 更新线程
        thread = await db.get(AIThread, request.thread_id)
        if thread:
            thread.updated_at = datetime.utcnow()
    
    await db.commit()
    
    return response


@router.post("/pql", response_model=dict)
async def generate_pql(
    query: str,
    workspace_id: int,
    project_id: Optional[int] = None,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    自然语言转 PQL 接口
    
    将用户的自然语言描述转换为 PQL (Plane Query Language) 查询
    
    示例:
    - "Show me high priority bugs" -> "priority = HIGH AND type = BUG"
    - "Find overdue issues" -> "isOverdue()"
    - "Unassigned tasks" -> "hasNoAssignee()"
    """
    await require_workspace_access(workspace_id, current_user, db)
    
    # 获取上下文
    context_aggregator = AIContextAggregator(db)
    context = await context_aggregator.get_workspace_context(
        workspace_id=workspace_id,
        project_id=project_id,
        include_issues=False  # PQL生成不需要完整问题数据
    )
    
    # 生成 PQL
    pql_generator = PQLGenerator(context)
    pql_query = pql_generator.generate_from_natural_language(query)
    
    return {
        "natural_language": query,
        "pql": pql_query,
        "workspace_id": str(workspace_id)
    }


@router.get("/context/{workspace_id}", response_model=dict)
async def get_workspace_context(
    workspace_id: int,
    project_id: Optional[int] = None,
    include_issues: bool = True,
    include_cycles: bool = True,
    include_modules: bool = True,
    issue_limit: int = 50,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取工作区上下文
    
    返回工作区的完整数据上下文，供 AI 使用
    
    包括:
    - 项目列表
    - 问题列表
    - 周期列表
    - 模块列表
    - 状态列表
    - 用户列表
    - 汇总统计
    """
    await require_workspace_access(workspace_id, current_user, db)
    
    context_aggregator = AIContextAggregator(db)
    context = await context_aggregator.get_workspace_context(
        workspace_id=workspace_id,
        project_id=project_id,
        include_issues=include_issues,
        include_cycles=include_cycles,
        include_modules=include_modules,
        issue_limit=issue_limit
    )
    
    return context


# ==================== AI Agent ====================

@router.post("/agent/triage/{issue_id}", response_model=AIAction)
async def triage_issue(
    issue_id: int,
    workspace_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    AI Agent: 自动分类问题
    
    分析问题内容并建议:
    - 优先级
    - 类型
    - 标签
    """
    await require_workspace_access(workspace_id, current_user, db)
    
    from app.services.ai import AIAgent
    
    # 获取上下文
    context_aggregator = AIContextAggregator(db)
    context = await context_aggregator.get_workspace_context(workspace_id)
    
    # 执行分类
    agent = AIAgent(db, context)
    action = await agent.triage_issue(issue_id)
    
    return action


@router.post("/agent/assign/{issue_id}", response_model=AIAction)
async def suggest_assignee(
    issue_id: int,
    workspace_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    AI Agent: 建议分配人员
    
    基于工作负载和技能匹配建议合适的分配人员
    """
    await require_workspace_access(workspace_id, current_user, db)
    
    from app.services.ai import AIAgent
    
    # 获取上下文
    context_aggregator = AIContextAggregator(db)
    context = await context_aggregator.get_workspace_context(workspace_id)
    
    # 执行分配建议
    agent = AIAgent(db, context)
    action = await agent.suggest_assignee(issue_id)
    
    return action


@router.get("/agent/blockers/{workspace_id}", response_model=list)
async def track_blockers(
    workspace_id: int,
    project_id: Optional[int] = None,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    AI Agent: 跟踪阻塞项
    
    检测工作区中的阻塞项:
    - 长期未处理的 backlog 问题
    - 即将到期但未完成的问题
    - 其他潜在阻塞
    """
    await require_workspace_access(workspace_id, current_user, db)
    
    from app.services.ai import AIAgent
    
    # 获取上下文
    context_aggregator = AIContextAggregator(db)
    context = await context_aggregator.get_workspace_context(
        workspace_id=workspace_id,
        project_id=project_id
    )
    
    # 执行阻塞跟踪
    agent = AIAgent(db, context)
    blockers = await agent.track_blockers(project_id)
    
    return blockers


# ==================== Plan Execution ====================

@router.post("/execute-plan", response_model=dict)
async def execute_plan(
    plan: AIPlan,
    workspace_id: int,
    thread_id: Optional[int] = None,
    message_id: Optional[int] = None,
    confirm: bool = False,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    执行 AI 操作计划
    
    当用户确认后，执行 AI 生成的操作计划
    
    注意: 需要用户确认才能执行实际操作
    """
    await require_workspace_access(workspace_id, current_user, db)
    
    if not confirm:
        return {
            "status": "pending_confirmation",
            "message": "Plan requires user confirmation before execution",
            "plan": plan,
            "actions_count": len(plan.actions)
        }
    
    # 执行计划中的操作
    executed_actions = []
    for action in plan.actions:
        result = await _execute_action(db, action)
        executed_actions.append({
            "action_type": action.action_type,
            "target_type": action.target_type,
            "target_id": str(action.target_id) if action.target_id else None,
            "result": result
        })
        
        # 记录操作日志
        log = AIActionLog(
            thread_id=thread_id,
            message_id=message_id,
            action_type=action.action_type,
            target_type=action.target_type,
            target_id=action.target_id,
            changes=action.changes,
            description=action.description,
            status="executed" if result.get("success") else "failed",
            error_message=result.get("error") if not result.get("success") else None,
            created_by_id=current_user.id
        )
        db.add(log)
    
    await db.commit()
    
    return {
        "status": "executed",
        "executed_actions": executed_actions,
        "total_actions": len(executed_actions)
    }


async def _execute_action(db: AsyncSession, action: AIAction) -> dict:
    """执行单个 AI 操作"""
    from app.models.issue import Issue
    
    if action.target_type == "issue" and action.target_id:
        issue = await db.get(Issue, action.target_id)
        if not issue:
            return {"success": False, "error": "Issue not found"}
        
        # 应用变更
        for field, value in action.changes.items():
            if hasattr(issue, field):
                setattr(issue, field, value)
        
        return {"success": True, "changes_applied": list(action.changes.keys())}
    
    return {"success": False, "error": "Unknown target type"}


# ==================== Suggestions ====================

@router.get("/suggestions/{workspace_id}", response_model=list)
async def get_ai_suggestions(
    workspace_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取 AI 建议
    
    基于工作区状态生成智能建议
    """
    await require_workspace_access(workspace_id, current_user, db)
    
    ai_service = AIService(db)
    context = await ai_service.context_aggregator.get_workspace_context(workspace_id)
    
    suggestions = ai_service._generate_suggestions(context)
    
    return suggestions


# ==================== NLP Task Creation ====================

@router.post("/tasks/parse", response_model=TaskExtractionResponse)
async def parse_task_from_nlp(
    request: TaskExtractionRequest,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    从自然语言解析任务信息
    
    将用户的自然语言输入解析为结构化的任务数据，包括：
    - 标题
    - 描述
    - 优先级
    - 截止日期
    - 负责人
    - 标签
    
    示例输入：
    "创建一个高优先级任务，标题为修复登录页面bug，描述为用户反馈使用Chrome浏览器时无法登录，截止日期为本周五，负责人分配给张三"
    """
    await require_workspace_access(request.workspace_id, current_user, db)
    
    # 验证项目属于工作区
    from app.models.project import Project
    project = await db.get(Project, request.project_id)
    if not project or project.workspace_id != request.workspace_id:
        raise HTTPException(status_code=404, detail="Project not found in workspace")
    
    # 解析自然语言
    parser = NLPParser()
    extraction = parser.parse_task_creation(request.text)
    
    # 验证提取结果
    warnings = parser.validate_extraction(extraction)
    
    # 生成建议
    suggestions = []
    if not extraction.title:
        suggestions.append("请提供任务标题以便更清晰地描述任务")
    if not extraction.priority:
        suggestions.append("未检测到优先级，将使用默认优先级")
    if extraction.confidence < 0.7:
        suggestions.append("解析置信度较低，建议手动检查提取的信息")
    
    return TaskExtractionResponse(
        extracted_data={
            "title": extraction.title,
            "description": extraction.description,
            "priority": extraction.priority.value if extraction.priority else None,
            "due_date": str(extraction.due_date) if extraction.due_date else None,
            "start_date": str(extraction.start_date) if extraction.start_date else None,
            "assignee_email": extraction.assignee_email,
            "assignee_name": extraction.assignee_name,
            "labels": extraction.labels,
        },
        confidence=extraction.confidence,
        warnings=warnings,
        parsing_notes=extraction.parsing_notes,
        suggestions=suggestions
    )


@router.post("/tasks/create-from-nlp", response_model=TaskCreateFromNLPResponse)
async def create_task_from_nlp(
    request: TaskCreateFromNLPRequest,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    从自然语言创建任务
    
    解析自然语言并自动创建任务。如果 review_before_create=True，
    则先返回提取结果供用户确认。
    
    流程：
    1. 解析自然语言
    2. 提取任务信息
    3. 如果需要审核，返回提取结果
    4. 如果自动创建，创建任务并返回结果
    """
    await require_workspace_access(request.workspace_id, current_user, db)
    
    # 验证项目属于工作区
    from app.models.project import Project
    project = await db.get(Project, request.project_id)
    if not project or project.workspace_id != request.workspace_id:
        raise HTTPException(status_code=404, detail="Project not found in workspace")
    
    # 解析自然语言
    parser = NLPParser()
    extraction = parser.parse_task_creation(request.text)
    
    # 验证提取结果
    warnings = parser.validate_extraction(extraction)
    
    # 如果不需要自动创建或需要审核，返回提取结果
    if not request.auto_create or request.review_before_create:
        return TaskCreateFromNLPResponse(
            created=False,
            extraction_details={
                "title": extraction.title,
                "description": extraction.description,
                "priority": extraction.priority.value if extraction.priority else None,
                "due_date": str(extraction.due_date) if extraction.due_date else None,
                "start_date": str(extraction.start_date) if extraction.start_date else None,
                "assignee_email": extraction.assignee_email,
                "assignee_name": extraction.assignee_name,
                "labels": extraction.labels,
            },
            confidence=extraction.confidence,
            warnings=warnings
        )
    
    # 自动创建任务
    try:
        from app.models.issue import Issue
        from app.models.state import State
        
        # 获取默认状态（Todo）
        state_result = await db.execute(
            select(State).where(
                State.workspace_id == request.workspace_id,
                State.group == "todo",
                State.is_default == True
            ).limit(1)
        )
        default_state = state_result.scalar_one_or_none()
        
        if not default_state:
            # 如果没有默认状态，获取第一个 Todo 状态
            state_result = await db.execute(
                select(State).where(
                    State.workspace_id == request.workspace_id,
                    State.group == "todo"
                ).limit(1)
            )
            default_state = state_result.scalar_one_or_none()
        
        # 查找负责人（如果有邮箱或姓名）
        assignee_id = None
        if extraction.assignee_email:
            user_result = await db.execute(
                select(User).where(User.email == extraction.assignee_email)
            )
            assignee_user = user_result.scalar_one_or_none()
            if assignee_user:
                assignee_id = assignee_user.id
        
        # 创建 Issue
        issue = Issue(
            name=extraction.title or f"Task from NLP: {request.text[:50]}",
            description=extraction.description,
            priority=extraction.priority.value if extraction.priority else "medium",
            state_id=default_state.id if default_state else None,
            project_id=request.project_id,
            workspace_id=request.workspace_id,
            assignee_id=assignee_id,
            start_date=extraction.start_date,
            target_date=extraction.due_date,
            is_draft=False,
            created_by_id=current_user.id
        )
        
        db.add(issue)
        await db.flush()  # 获取 issue ID
        
        # 添加标签（如果有）
        if extraction.labels:
            from app.models.label import Label
            for label_name in extraction.labels:
                # 查找或创建标签
                label_result = await db.execute(
                    select(Label).where(
                        Label.workspace_id == request.workspace_id,
                        Label.name == label_name
                    ).limit(1)
                )
                label = label_result.scalar_one_or_none()
                
                if not label:
                    label = Label(
                        name=label_name,
                        color="#3B82F6",  # 默认蓝色
                        workspace_id=request.workspace_id,
                        created_by_id=current_user.id
                    )
                    db.add(label)
                    await db.flush()
                
                # 关联标签到问题
                issue.labels.append(label)
        
        await db.commit()
        await db.refresh(issue)
        
        return TaskCreateFromNLPResponse(
            task_id=issue.id,
            created=True,
            extraction_details={
                "title": extraction.title,
                "description": extraction.description,
                "priority": extraction.priority.value if extraction.priority else None,
                "due_date": str(extraction.due_date) if extraction.due_date else None,
                "start_date": str(extraction.start_date) if extraction.start_date else None,
                "assignee_email": extraction.assignee_email,
                "assignee_name": extraction.assignee_name,
                "labels": extraction.labels,
            },
            confidence=extraction.confidence,
            warnings=warnings
        )
    
    except Exception as e:
        await db.rollback()
        raise HTTPException(
            status_code=500,
            detail=f"Failed to create task: {str(e)}"
        )
