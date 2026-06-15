"""
Issue Services - 工作项业务逻辑层
"""
from typing import Optional, List, Dict, Any
from uuid import UUID
from datetime import datetime

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload
from sqlalchemy import and_, or_, func, delete

from app.models.issue import Issue, IssueActivity, IssueAssignee, IssueLabel, IssueCycle
from app.models.project import Project
from app.models.state import State
from app.models.user import User
from app.models.label import Label
from app.models.cycle import Cycle
from app.models.issue import IssueModule
from app.models.module import Module
from app.models.issue_type import IssueType
from app.models.custom_field import IssueCustomFieldValue
from app.schemas.issue import (
    IssueCreate,
    IssueUpdate,
    IssuePriority,
    IssueResponse
)
from app.core.exceptions import NotFoundException, ValidationException


# ==================== Issue CRUD ====================

async def create_issue(
    db: AsyncSession,
    issue_data: IssueCreate,
    project_id: UUID,
    workspace_id: UUID,
    user_id: UUID
) -> Issue:
    """创建工作项"""
    # 验证项目存在
    project = await db.get(Project, project_id)
    if not project or project.is_deleted:
        raise NotFoundException("Project not found")
    
    # 验证状态存在（如果提供了 state_id）
    if issue_data.state_id:
        state = await db.get(State, issue_data.state_id)
        if not state or state.is_deleted:
            raise NotFoundException("State not found")
    else:
        # 获取默认状态（Backlog）
        result = await db.execute(
            select(State).where(
                State.project_id == project_id,
                State.group == "backlog",
                State.is_deleted == False
            ).order_by(State.sequence)
        )
        state = result.scalars().first()
        if not state:
            raise ValidationException("No default state found for project")
        issue_data.state_id = state.id
    
    # 获取下一个 sequence_id
    result = await db.execute(
        select(func.max(Issue.sequence_id)).where(
            Issue.project_id == project_id
        )
    )
    max_seq = result.scalar_one_or_none()
    next_seq = (max_seq or 0) + 1
    
    # 创建工作项
    issue = Issue(
        name=issue_data.name,
        description_html=issue_data.description_html,
        description_json=issue_data.description_json,
        priority=issue_data.priority.value,
        sequence_id=next_seq,
        start_date=issue_data.start_date,
        target_date=issue_data.target_date,
        state_id=issue_data.state_id,
        parent_id=issue_data.parent_id,
        project_id=project_id,
        workspace_id=workspace_id,
        external_id=issue_data.external_id,
        external_source=issue_data.external_source,
        created_by_id=user_id
    )
    
    db.add(issue)
    await db.flush()
    
    # 添加活动记录
    activity = IssueActivity(
        issue_id=issue.id,
        verb="created",
        actor_id=user_id,
        created_by_id=user_id
    )
    db.add(activity)
    
    # 处理 assignees（如果有）
    if issue_data.assignee_ids:
        await _add_assignees(db, issue.id, issue_data.assignee_ids)
    
    # 处理 labels（如果有）
    if issue_data.label_ids:
        await _add_labels(db, issue.id, issue_data.label_ids)
    
    await db.commit()
    await db.refresh(issue)
    
    # 预加载关联数据以避免延迟加载问题
    result = await db.execute(
        select(Issue)
        .where(Issue.id == issue.id)
        .options(
            selectinload(Issue.project),
            selectinload(Issue.state),
            selectinload(Issue.assignee_links).selectinload(IssueAssignee.user),
            selectinload(Issue.label_links).selectinload(IssueLabel.label),
            selectinload(Issue.cycle_link).selectinload(IssueCycle.cycle),
            selectinload(Issue.module_links).selectinload(IssueModule.module),
            selectinload(Issue.sub_issues),
            selectinload(Issue.attachments)
        )
    )
    issue = result.scalar_one_or_none()
    
    return issue


async def get_issue_by_id(db: AsyncSession, issue_id: UUID) -> Issue:
    """获取工作项详情"""
    result = await db.execute(
        select(Issue)
        .where(Issue.id == issue_id, Issue.is_deleted == False)
        .options(
            selectinload(Issue.project),
            selectinload(Issue.state),
            selectinload(Issue.parent),
            selectinload(Issue.sub_issues),
            selectinload(Issue.custom_field_values),
            selectinload(Issue.assignee_links).selectinload(IssueAssignee.user),
            selectinload(Issue.label_links).selectinload(IssueLabel.label),
            selectinload(Issue.cycle_link).selectinload(IssueCycle.cycle),
            selectinload(Issue.module_links).selectinload(IssueModule.module),
            selectinload(Issue.attachments)
        )
    )
    issue = result.scalar_one_or_none()
    if not issue:
        raise NotFoundException("Issue not found")
    return issue


async def list_project_issues(
    db: AsyncSession,
    project_id: UUID,
    state_id: Optional[UUID] = None,
    priority: Optional[IssuePriority] = None,
    assignee_id: Optional[UUID] = None,
    parent_id: Optional[UUID] = None,
    cycle_id: Optional[UUID] = None,
    module_id: Optional[UUID] = None,
    search: Optional[str] = None,
    is_draft: Optional[bool] = None,
    limit: int = 50,
    offset: int = 0
) -> List[Issue]:
    """列出项目的工作项"""
    query = select(Issue).where(
        Issue.project_id == project_id,
        Issue.is_deleted == False
    )
    
    # 状态筛选
    if state_id:
        query = query.where(Issue.state_id == state_id)
    
    # 优先级筛选
    if priority:
        query = query.where(Issue.priority == priority.value)
    
    # 父工作项筛选
    if parent_id:
        query = query.where(Issue.parent_id == parent_id)
    
    # Draft 状态筛选
    if is_draft is not None:
        query = query.where(Issue.is_draft == is_draft)
    
    # 负责人筛选
    if assignee_id:
        query = query.join(IssueAssignee).where(IssueAssignee.user_id == assignee_id)
    
    # 周期筛选
    if cycle_id:
        query = query.join(IssueCycle).where(IssueCycle.cycle_id == cycle_id)
    
    # 模块筛选
    if module_id:
        query = query.join(IssueModule).where(IssueModule.module_id == module_id)
    
    # 搜索
    if search:
        query = query.where(
            or_(
                Issue.name.ilike(f"%{search}%"),
                Issue.description_stripped.ilike(f"%{search}%")
            )
        )
    
    # 排序
    query = query.order_by(Issue.sort_order, Issue.sequence_id.desc())
    
    # 分页
    query = query.limit(limit).offset(offset)
    
    result = await db.execute(query.options(
        selectinload(Issue.project),
        selectinload(Issue.state),
        selectinload(Issue.assignee_links).selectinload(IssueAssignee.user),
        selectinload(Issue.label_links).selectinload(IssueLabel.label),
        selectinload(Issue.cycle_link).selectinload(IssueCycle.cycle),
        selectinload(Issue.module_links).selectinload(IssueModule.module),
        selectinload(Issue.sub_issues),
        selectinload(Issue.attachments)
    ))
    
    return list(result.scalars().all())


async def update_issue(
    db: AsyncSession,
    issue_id: UUID,
    update_data: IssueUpdate,
    user_id: UUID
) -> Issue:
    """更新工作项"""
    issue = await get_issue_by_id(db, issue_id)
    
    # 记录变更
    changes = []
    
    # 更新名称
    if update_data.name is not None and update_data.name != issue.name:
        changes.append(("name", issue.name, update_data.name))
        issue.name = update_data.name
    
    # 更新描述
    if update_data.description_html is not None:
        changes.append(("description", issue.description_html, update_data.description_html))
        issue.description_html = update_data.description_html
    
    # 更新优先级
    if update_data.priority is not None and update_data.priority.value != issue.priority:
        changes.append(("priority", issue.priority, update_data.priority.value))
        issue.priority = update_data.priority.value
    
    # 更新状态
    if update_data.state_id is not None and update_data.state_id != issue.state_id:
        # 验证新状态
        new_state = await db.get(State, update_data.state_id)
        if not new_state or new_state.is_deleted:
            raise NotFoundException("State not found")
        
        old_state = await db.get(State, issue.state_id)
        changes.append(("state", old_state.name if old_state else None, new_state.name))
        issue.state_id = update_data.state_id
        
        # 如果状态变为完成，记录完成时间
        if new_state.group == "done":
            issue.completed_at = datetime.utcnow()
    
    # 更新日期
    if update_data.start_date is not None:
        changes.append(("start_date", str(issue.start_date), str(update_data.start_date)))
        issue.start_date = update_data.start_date
    
    if update_data.target_date is not None:
        changes.append(("target_date", str(issue.target_date), str(update_data.target_date)))
        issue.target_date = update_data.target_date
    
    # 更新 assignees
    if update_data.assignee_ids is not None:
        await _update_assignees(db, issue_id, update_data.assignee_ids)
        changes.append(("assignees", None, str(update_data.assignee_ids)))
    
    # 更新 labels
    if update_data.label_ids is not None:
        await _update_labels(db, issue_id, update_data.label_ids)
        changes.append(("labels", None, str(update_data.label_ids)))
    
    # 更新 cycle
    if update_data.cycle_id is not None:
        await _update_cycle(db, issue_id, update_data.cycle_id)
        changes.append(("cycle", None, str(update_data.cycle_id)))
    
    # 更新 modules
    if update_data.module_ids is not None:
        await _update_modules(db, issue_id, update_data.module_ids)
        changes.append(("modules", None, str(update_data.module_ids)))
    
    # 记录活动
    for field, old_value, new_value in changes:
        activity = IssueActivity(
            issue_id=issue.id,
            verb="updated",
            field=field,
            old_value=old_value,
            new_value=new_value,
            actor_id=user_id,
            created_by_id=user_id
        )
        db.add(activity)
    
    await db.commit()
    await db.refresh(issue)
    
    return issue


async def delete_issue(db: AsyncSession, issue_id: UUID):
    """删除工作项（软删除）"""
    issue = await get_issue_by_id(db, issue_id)
    issue.is_deleted = True
    await db.commit()


async def archive_issue(db: AsyncSession, issue_id: UUID):
    """归档工作项"""
    issue = await get_issue_by_id(db, issue_id)
    issue.archived_at = datetime.utcnow().date()
    await db.commit()


async def restore_issue(db: AsyncSession, issue_id: UUID):
    """恢复工作项"""
    issue = await get_issue_by_id(db, issue_id)
    issue.archived_at = None
    issue.is_deleted = False
    await db.commit()


# ==================== Assignee Management ====================

async def _add_assignees(db: AsyncSession, issue_id: UUID, assignee_ids: List[UUID]):
    """添加负责人"""
    for user_id in assignee_ids:
        user = await db.get(User, user_id)
        if not user:
            raise NotFoundException(f"User {user_id} not found")
        
        association = IssueAssignee(
            issue_id=issue_id,
            user_id=user_id
        )
        db.add(association)
    
    await db.flush()


async def _update_assignees(db: AsyncSession, issue_id: UUID, assignee_ids: List[UUID]):
    """更新负责人"""
    # 删除旧的关联
    await db.execute(
        delete(IssueAssignee).where(IssueAssignee.issue_id == issue_id)
    )
    
    # 添加新的关联
    await _add_assignees(db, issue_id, assignee_ids)


async def add_assignee(
    db: AsyncSession,
    issue_id: UUID,
    user_id: UUID,
    actor_id: UUID
) -> Dict[str, Any]:
    """添加单个负责人"""
    issue = await get_issue_by_id(db, issue_id)
    user = await db.get(User, user_id)
    if not user:
        raise NotFoundException("User not found")
    
    # 检查是否已存在
    result = await db.execute(
        select(IssueAssignee).where(
            IssueAssignee.issue_id == issue_id,
            IssueAssignee.user_id == user_id
        )
    )
    existing = result.scalar_one_or_none()
    if existing:
        raise ValidationException("User is already assigned to this issue")
    
    association = IssueAssignee(
        issue_id=issue_id,
        user_id=user_id
    )
    db.add(association)
    
    # 记录活动
    activity = IssueActivity(
        issue_id=issue_id,
        verb="updated",
        field="assignees",
        new_value=f"added {user.display_name or user.username}",
        actor_id=actor_id,
        created_by_id=actor_id
    )
    db.add(activity)
    
    await db.commit()
    
    return {"issue_id": str(issue_id), "user_id": str(user_id), "action": "added"}


async def remove_assignee(
    db: AsyncSession,
    issue_id: UUID,
    user_id: UUID,
    actor_id: UUID
) -> Dict[str, Any]:
    """移除单个负责人"""
    issue = await get_issue_by_id(db, issue_id)
    
    result = await db.execute(
        select(IssueAssignee).where(
            IssueAssignee.issue_id == issue_id,
            IssueAssignee.user_id == user_id
        )
    )
    association = result.scalar_one_or_none()
    
    if not association:
        raise ValidationException("User is not assigned to this issue")
    
    await db.delete(association)
    
    # 记录活动
    user = await db.get(User, user_id)
    activity = IssueActivity(
        issue_id=issue_id,
        verb="updated",
        field="assignees",
        new_value=f"removed {user.display_name or user.username if user else user_id}",
        actor_id=actor_id,
        created_by_id=actor_id
    )
    db.add(activity)
    
    await db.commit()
    
    return {"issue_id": str(issue_id), "user_id": str(user_id), "action": "removed"}


# ==================== Label Management ====================

async def _add_labels(db: AsyncSession, issue_id: UUID, label_ids: List[UUID]):
    """添加标签"""
    for label_id in label_ids:
        label = await db.get(Label, label_id)
        if not label or label.is_deleted:
            raise NotFoundException(f"Label {label_id} not found")
        
        association = IssueLabel(
            issue_id=issue_id,
            label_id=label_id
        )
        db.add(association)
    
    await db.flush()


async def _update_labels(db: AsyncSession, issue_id: UUID, label_ids: List[UUID]):
    """更新标签"""
    # 删除旧的关联
    await db.execute(
        delete(IssueLabel).where(IssueLabel.issue_id == issue_id)
    )
    
    # 添加新的关联
    await _add_labels(db, issue_id, label_ids)


async def add_label(
    db: AsyncSession,
    issue_id: UUID,
    label_id: UUID,
    actor_id: UUID
) -> Dict[str, Any]:
    """添加单个标签"""
    issue = await get_issue_by_id(db, issue_id)
    label = await db.get(Label, label_id)
    if not label or label.is_deleted:
        raise NotFoundException("Label not found")
    
    # 检查是否已存在
    result = await db.execute(
        select(IssueLabel).where(
            IssueLabel.issue_id == issue_id,
            IssueLabel.label_id == label_id
        )
    )
    existing = result.scalar_one_or_none()
    if existing:
        raise ValidationException("Label is already attached to this issue")
    
    association = IssueLabel(
        issue_id=issue_id,
        label_id=label_id
    )
    db.add(association)
    
    # 记录活动
    activity = IssueActivity(
        issue_id=issue_id,
        verb="updated",
        field="labels",
        new_value=f"added {label.name}",
        actor_id=actor_id,
        created_by_id=actor_id
    )
    db.add(activity)
    
    await db.commit()
    
    return {"issue_id": str(issue_id), "label_id": str(label_id), "action": "added"}


async def remove_label(
    db: AsyncSession,
    issue_id: UUID,
    label_id: UUID,
    actor_id: UUID
) -> Dict[str, Any]:
    """移除单个标签"""
    issue = await get_issue_by_id(db, issue_id)
    
    result = await db.execute(
        select(IssueLabel).where(
            IssueLabel.issue_id == issue_id,
            IssueLabel.label_id == label_id
        )
    )
    association = result.scalar_one_or_none()
    
    if not association:
        raise ValidationException("Label is not attached to this issue")
    
    await db.delete(association)
    
    # 记录活动
    label = await db.get(Label, label_id)
    activity = IssueActivity(
        issue_id=issue_id,
        verb="updated",
        field="labels",
        new_value=f"removed {label.name if label else label_id}",
        actor_id=actor_id,
        created_by_id=actor_id
    )
    db.add(activity)
    
    await db.commit()
    
    return {"issue_id": str(issue_id), "label_id": str(label_id), "action": "removed"}


# ==================== Cycle Management ====================

async def _add_cycle(db: AsyncSession, issue_id: UUID, cycle_id: UUID):
    """添加周期关联"""
    cycle = await db.get(Cycle, cycle_id)
    if not cycle or cycle.is_deleted:
        raise NotFoundException(f"Cycle {cycle_id} not found")
    
    association = IssueCycle(
        issue_id=issue_id,
        cycle_id=cycle_id
    )
    db.add(association)
    await db.flush()


async def _update_cycle(db: AsyncSession, issue_id: UUID, cycle_id: Optional[UUID]):
    """更新周期关联"""
    # 删除旧的关联
    await db.execute(
        delete(IssueCycle).where(IssueCycle.issue_id == issue_id)
    )
    
    # 如果提供了新的周期ID，添加关联
    if cycle_id:
        await _add_cycle(db, issue_id, cycle_id)


async def set_issue_cycle(
    db: AsyncSession,
    issue_id: UUID,
    cycle_id: UUID,
    actor_id: UUID
) -> Dict[str, Any]:
    """设置工作项周期"""
    issue = await get_issue_by_id(db, issue_id)
    cycle = await db.get(Cycle, cycle_id)
    if not cycle or cycle.is_deleted:
        raise NotFoundException("Cycle not found")
    
    # 验证周期和工作项属于同一项目
    if cycle.project_id != issue.project_id:
        raise ValidationException("Cycle does not belong to this issue's project")
    
    # 删除旧的关联
    await db.execute(
        delete(IssueCycle).where(IssueCycle.issue_id == issue_id)
    )
    
    # 添加新的关联
    association = IssueCycle(
        issue_id=issue_id,
        cycle_id=cycle_id
    )
    db.add(association)
    
    # 记录活动
    activity = IssueActivity(
        issue_id=issue_id,
        verb="updated",
        field="cycle",
        new_value=cycle.name,
        actor_id=actor_id,
        created_by_id=actor_id
    )
    db.add(activity)
    
    await db.commit()
    
    return {"issue_id": str(issue_id), "cycle_id": str(cycle_id), "action": "set"}


async def remove_issue_cycle(
    db: AsyncSession,
    issue_id: UUID,
    actor_id: UUID
) -> Dict[str, Any]:
    """移除工作项周期"""
    issue = await get_issue_by_id(db, issue_id)
    
    result = await db.execute(
        select(IssueCycle).where(IssueCycle.issue_id == issue_id)
    )
    association = result.scalar_one_or_none()
    
    if not association:
        raise ValidationException("Issue is not in any cycle")
    
    cycle = await db.get(Cycle, association.cycle_id)
    cycle_name = cycle.name if cycle else None
    
    await db.delete(association)
    
    # 记录活动
    activity = IssueActivity(
        issue_id=issue_id,
        verb="updated",
        field="cycle",
        old_value=cycle_name,
        new_value="removed",
        actor_id=actor_id,
        created_by_id=actor_id
    )
    db.add(activity)
    
    await db.commit()
    
    return {"issue_id": str(issue_id), "cycle_id": None, "action": "removed"}


# ==================== Module Management ====================

async def _update_modules(db: AsyncSession, issue_id: UUID, module_ids: List[UUID]):
    """更新模块关联"""
    # 验证模块存在
    for module_id in module_ids:
        module = await db.get(Module, module_id)
        if not module or module.is_deleted:
            raise NotFoundException(f"Module {module_id} not found")
    
    # 删除旧的关联
    await db.execute(
        delete(IssueModule).where(IssueModule.issue_id == issue_id)
    )
    
    # 创建新的关联
    for module_id in module_ids:
        association = IssueModule(
            module_id=module_id,
            issue_id=issue_id
        )
        db.add(association)
    
    await db.flush()


# ==================== Issue Activity ====================

async def get_issue_activities(
    db: AsyncSession,
    issue_id: UUID,
    limit: int = 50,
    offset: int = 0
) -> List[IssueActivity]:
    """获取工作项活动历史"""
    result = await db.execute(
        select(IssueActivity)
        .where(IssueActivity.issue_id == issue_id)
        .order_by(IssueActivity.created_at.desc())
        .limit(limit)
        .offset(offset)
    )
    return list(result.scalars().all())


# ==================== Issue Statistics ====================

async def get_issue_statistics(db: AsyncSession, project_id: UUID) -> Dict[str, Any]:
    """获取项目工作项统计"""
    # 总数
    total_result = await db.execute(
        select(func.count(Issue.id)).where(
            Issue.project_id == project_id,
            Issue.is_deleted == False
        )
    )
    total = total_result.scalar_one()
    
    # 按状态分组统计
    state_stats_result = await db.execute(
        select(State.name, func.count(Issue.id))
        .join(Issue, Issue.state_id == State.id)
        .where(Issue.project_id == project_id, Issue.is_deleted == False)
        .group_by(State.name)
    )
    state_stats = {row[0]: row[1] for row in state_stats_result.all()}
    
    # 按优先级分组统计
    priority_stats_result = await db.execute(
        select(Issue.priority, func.count(Issue.id))
        .where(Issue.project_id == project_id, Issue.is_deleted == False)
        .group_by(Issue.priority)
    )
    priority_stats = {row[0]: row[1] for row in priority_stats_result.all()}
    
    return {
        "total": total,
        "by_state": state_stats,
        "by_priority": priority_stats
    }


# ==================== Helper Functions ====================

def build_issue_response(issue: Issue) -> Dict[str, Any]:
    """构建 Issue Response"""
    # 获取负责人列表
    assignees = []
    if issue.assignee_links:
        for link in issue.assignee_links:
            if link.user:
                assignees.append({
                    "id": link.user.id,
                    "display_name": link.user.display_name or "",
                    "email": link.user.email,
                    "avatar_url": link.user.avatar
                })
    
    # 获取标签ID列表
    labels = []
    if issue.label_links:
        for link in issue.label_links:
            if link.label:
                labels.append(link.label.id)
    
    # 获取项目信息
    project = None
    if issue.project:
        project = {
            "id": issue.project.id,
            "name": issue.project.name,
            "identifier": issue.project.identifier
        }
    
    # 获取周期ID
    cycle_id = issue.cycle_link.cycle.id if issue.cycle_link and issue.cycle_link.cycle else None
    
    # 获取模块ID列表
    module_ids = []
    if issue.module_links:
        for link in issue.module_links:
            if link.module:
                module_ids.append(link.module.id)
    
    return {
        "id": issue.id,
        "name": issue.name,
        "description_html": issue.description_html,
        "description_json": issue.description_json,
        "priority": issue.priority,
        "sequence_id": issue.sequence_id,
        "sort_order": issue.sort_order,
        "start_date": issue.start_date,
        "target_date": issue.target_date,
        "completed_at": issue.completed_at,
        "is_draft": issue.is_draft,
        "archived_at": issue.archived_at,
        "project_id": issue.project_id,
        "workspace_id": issue.workspace_id,
        "parent_id": issue.parent_id,
        "state_id": issue.state_id,
        "state_name": issue.state.name if issue.state else None,
        "state_group": issue.state.group if issue.state else None,
        "project": project,
        "assignees": assignees,
        "labels": labels,
        "sub_issues_count": len(issue.sub_issues) if issue.sub_issues else 0,
        "link_count": 0,
        "attachment_count": len(issue.attachments) if issue.attachments else 0,
        "estimate_point_id": None,
        "cycle_id": cycle_id,
        "module_ids": module_ids,
        "created_at": issue.created_at,
        "updated_at": issue.updated_at,
        "created_by": issue.created_by_id,
        "updated_by": issue.updated_by_id,
        "deleted_at": issue.deleted_at,
        "is_deleted": issue.is_deleted,
    }