from typing import Optional, List, Dict, Any
from datetime import datetime
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload
from sqlalchemy import func, and_

from app.models.project import Project, ProjectMember
from app.models.workspace import Workspace
from app.models.issue import Issue
from app.models.state import State
from app.schemas.project import ProjectCreate, ProjectUpdate
from app.core.exceptions import NotFoundException, ConflictException, ValidationException
from . import project_settings


async def create_project(
    db: AsyncSession,
    project_data: ProjectCreate,
    workspace_id: int,
    user_id: int
) -> Project:
    """创建项目"""
    workspace = await db.get(Workspace, workspace_id)
    if not workspace:
        raise NotFoundException("Workspace not found")
    
    existing = await db.execute(
        select(Project).where(
            Project.workspace_id == workspace_id, 
            Project.identifier == project_data.identifier
        )
    )
    if existing.scalar_one_or_none():
        raise ConflictException("Project identifier already exists in workspace")
    
    project = Project(
        name=project_data.name,
        identifier=project_data.identifier,
        description=project_data.description,
        is_public=project_data.is_public,
        timezone=project_data.timezone,
        workspace_id=workspace_id,
        default_assignee_id=project_data.default_assignee_id,
        created_by_id=user_id,
    )
    
    db.add(project)
    await db.flush()
    
    # 添加创建者为成员
    member = ProjectMember(
        project_id=project.id,
        user_id=user_id,
        role=20,
        created_by_id=user_id,
    )
    db.add(member)
    
    await db.commit()
    await db.refresh(project)
    
    # 创建默认工作项类型
    await project_settings.create_default_issue_types(db, project.id, workspace_id, user_id)
    
    # 创建默认状态
    await project_settings.create_default_states(db, project.id, workspace_id, user_id)
    
    return project


async def get_project_by_id(
    db: AsyncSession,
    project_id: int
) -> Project:
    """获取项目详情"""
    result = await db.execute(
        select(Project)
        .where(Project.id == project_id, Project.is_deleted == False)
        .options(selectinload(Project.members).selectinload(ProjectMember.user))
    )
    project = result.scalar_one_or_none()
    if not project:
        raise NotFoundException("Project not found")
    return project


async def list_workspace_projects(
    db: AsyncSession,
    workspace_id: int,
    include_archived: bool = False,
    limit: int = 50,
    offset: int = 0
) -> List[Project]:
    """列出工作空间的项目"""
    query = select(Project).where(
        Project.workspace_id == workspace_id,
        Project.is_deleted == False
    )
    
    if not include_archived:
        query = query.where(Project.archived_at == None)
    
    query = query.order_by(Project.created_at.desc())
    query = query.limit(limit).offset(offset)
    
    result = await db.execute(query)
    return list(result.scalars().all())


async def update_project(
    db: AsyncSession,
    project_id: int,
    update_data: ProjectUpdate
) -> Project:
    """更新项目"""
    project = await get_project_by_id(db, project_id)
    
    if update_data.name is not None:
        project.name = update_data.name
    if update_data.description is not None:
        project.description = update_data.description
    if update_data.is_public is not None:
        project.is_public = update_data.is_public
    if update_data.archived_at is not None:
        project.archived_at = update_data.archived_at
    if update_data.default_assignee_id is not None:
        project.default_assignee_id = update_data.default_assignee_id
    
    await db.commit()
    await db.refresh(project)
    
    return project


async def delete_project(db: AsyncSession, project_id: int):
    """删除项目（软删除）"""
    project = await get_project_by_id(db, project_id)
    project.is_deleted = True
    await db.commit()


async def archive_project(db: AsyncSession, project_id: int) -> Project:
    """归档项目"""
    project = await get_project_by_id(db, project_id)
    project.archived_at = datetime.utcnow()
    await db.commit()
    await db.refresh(project)
    return project


async def restore_project(db: AsyncSession, project_id: int) -> Project:
    """恢复项目"""
    project = await get_project_by_id(db, project_id)
    project.archived_at = None
    project.is_deleted = False
    await db.commit()
    await db.refresh(project)
    return project


# ==================== Project Member Management ====================

async def list_project_members(
    db: AsyncSession,
    project_id: int,
    only_active: bool = True
) -> List[ProjectMember]:
    """列出项目成员"""
    query = select(ProjectMember).where(ProjectMember.project_id == project_id)
    
    if only_active:
        query = query.where(ProjectMember.is_active == True)
    
    query = query.order_by(ProjectMember.role.desc())
    
    result = await db.execute(query.options(
        selectinload(ProjectMember.user)
    ))
    return list(result.scalars().all())


async def add_project_member(
    db: AsyncSession,
    project_id: int,
    user_id: int,
    added_by: int,
    role: int = 15
) -> ProjectMember:
    """添加项目成员"""
    # 验证项目存在
    project = await get_project_by_id(db, project_id)
    
    # 检查是否已是成员
    existing = await db.execute(
        select(ProjectMember).where(
            ProjectMember.project_id == project_id,
            ProjectMember.user_id == user_id
        )
    )
    if existing.scalar_one_or_none():
        raise ConflictException("User is already a member of this project")
    
    member = ProjectMember(
        project_id=project_id,
        user_id=user_id,
        role=role,
        created_by_id=added_by,
    )
    db.add(member)
    await db.commit()
    await db.refresh(member)
    
    return member


async def update_project_member(
    db: AsyncSession,
    project_id: int,
    user_id: int,
    role: int,
    updated_by: int
) -> ProjectMember:
    """更新项目成员角色"""
    result = await db.execute(
        select(ProjectMember).where(
            ProjectMember.project_id == project_id,
            ProjectMember.user_id == user_id
        )
    )
    member = result.scalar_one_or_none()
    if not member:
        raise NotFoundException("Project member not found")
    
    member.role = role
    await db.commit()
    await db.refresh(member)
    
    return member


async def remove_project_member(
    db: AsyncSession,
    project_id: int,
    user_id: int
) -> Dict[str, Any]:
    """移除项目成员"""
    result = await db.execute(
        select(ProjectMember).where(
            ProjectMember.project_id == project_id,
            ProjectMember.user_id == user_id
        )
    )
    member = result.scalar_one_or_none()
    if not member:
        raise NotFoundException("Project member not found")
    
    await db.delete(member)
    await db.commit()
    
    return {"project_id": project_id, "user_id": user_id, "action": "removed"}


async def deactivate_project_member(
    db: AsyncSession,
    project_id: int,
    user_id: int
) -> ProjectMember:
    """停用项目成员"""
    result = await db.execute(
        select(ProjectMember).where(
            ProjectMember.project_id == project_id,
            ProjectMember.user_id == user_id
        )
    )
    member = result.scalar_one_or_none()
    if not member:
        raise NotFoundException("Project member not found")
    
    member.is_active = False
    await db.commit()
    await db.refresh(member)
    
    return member


async def reactivate_project_member(
    db: AsyncSession,
    project_id: int,
    user_id: int
) -> ProjectMember:
    """重新激活项目成员"""
    result = await db.execute(
        select(ProjectMember).where(
            ProjectMember.project_id == project_id,
            ProjectMember.user_id == user_id
        )
    )
    member = result.scalar_one_or_none()
    if not member:
        raise NotFoundException("Project member not found")
    
    member.is_active = True
    await db.commit()
    await db.refresh(member)
    
    return member


# ==================== Project Statistics ====================

async def get_project_statistics(
    db: AsyncSession,
    project_id: int
) -> Dict[str, Any]:
    """获取项目统计"""
    project = await get_project_by_id(db, project_id)
    
    # 总工作项数
    total_issues_result = await db.execute(
        select(func.count(Issue.id)).where(
            Issue.project_id == project_id,
            Issue.is_deleted == False
        )
    )
    total_issues = total_issues_result.scalar_one_or_none() or 0
    
    # 已完成工作项数
    completed_issues_result = await db.execute(
        select(func.count(Issue.id))
        .join(State, Issue.state_id == State.id)
        .where(
            Issue.project_id == project_id,
            Issue.is_deleted == False,
            State.group == "done"
        )
    )
    completed_issues = completed_issues_result.scalar_one_or_none() or 0
    
    # 按状态分组统计
    state_stats_result = await db.execute(
        select(State.name, State.group, func.count(Issue.id))
        .join(Issue, Issue.state_id == State.id)
        .where(
            Issue.project_id == project_id,
            Issue.is_deleted == False
        )
        .group_by(State.name, State.group)
    )
    state_stats = [
        {"state": row[0], "group": row[1], "count": row[2]}
        for row in state_stats_result.all()
    ]
    
    # 成员数统计
    members_result = await db.execute(
        select(func.count(ProjectMember.id))
        .where(
            ProjectMember.project_id == project_id,
            ProjectMember.is_active == True
        )
    )
    active_members = members_result.scalar_one_or_none() or 0
    
    return {
        "project_id": project_id,
        "project_name": project.name,
        "total_issues": total_issues,
        "completed_issues": completed_issues,
        "progress": round((completed_issues / total_issues * 100) if total_issues > 0 else 0, 2),
        "state_breakdown": state_stats,
        "active_members": active_members,
        "is_archived": project.archived_at is not None
    }


# ==================== Project Issues Summary ====================

async def get_project_issues_summary(
    db: AsyncSession,
    project_id: int
) -> Dict[str, Any]:
    """获取项目工作项摘要"""
    project = await get_project_by_id(db, project_id)
    
    # 待办工作项
    todo_result = await db.execute(
        select(func.count(Issue.id))
        .join(State, Issue.state_id == State.id)
        .where(
            Issue.project_id == project_id,
            Issue.is_deleted == False,
            State.group.in_(["backlog", "todo"])
        )
    )
    todo_count = todo_result.scalar_one_or_none() or 0
    
    # 进行中工作项
    in_progress_result = await db.execute(
        select(func.count(Issue.id))
        .join(State, Issue.state_id == State.id)
        .where(
            Issue.project_id == project_id,
            Issue.is_deleted == False,
            State.group == "in_progress"
        )
    )
    in_progress_count = in_progress_result.scalar_one_or_none() or 0
    
    # 已完成工作项
    done_result = await db.execute(
        select(func.count(Issue.id))
        .join(State, Issue.state_id == State.id)
        .where(
            Issue.project_id == project_id,
            Issue.is_deleted == False,
            State.group == "done"
        )
    )
    done_count = done_result.scalar_one_or_none() or 0
    
    # 已取消工作项
    cancelled_result = await db.execute(
        select(func.count(Issue.id))
        .join(State, Issue.state_id == State.id)
        .where(
            Issue.project_id == project_id,
            Issue.is_deleted == False,
            State.group == "cancelled"
        )
    )
    cancelled_count = cancelled_result.scalar_one_or_none() or 0
    
    return {
        "project_id": project_id,
        "project_name": project.name,
        "issues": {
            "todo": todo_count,
            "in_progress": in_progress_count,
            "done": done_count,
            "cancelled": cancelled_count
        }
    }


# ==================== Helper Functions ====================

async def build_project_response(db: AsyncSession, project: Project) -> Dict[str, Any]:
    """构建Project Response"""
    # 获取工作空间信息
    workspace = await db.get(Workspace, project.workspace_id)
    
    # 获取工作项数量
    issues_result = await db.execute(
        select(func.count(Issue.id)).where(
            Issue.project_id == project.id,
            Issue.is_deleted == False
        )
    )
    total_issues = issues_result.scalar_one_or_none() or 0
    
    # 获取成员数量
    members_result = await db.execute(
        select(func.count(ProjectMember.id)).where(
            ProjectMember.project_id == project.id,
            ProjectMember.is_active == True
        )
    )
    total_members = members_result.scalar_one_or_none() or 0
    
    # 获取默认负责人信息
    default_assignee = None
    if project.default_assignee_id:
        from app.models.user import User
        user = await db.get(User, project.default_assignee_id)
        if user:
            default_assignee = {
                "id": user.id,
                "username": user.username,
                "display_name": user.display_name,
                "avatar": user.avatar
            }
    
    return {
        "id": project.id,
        "name": project.name,
        "identifier": project.identifier,
        "description": project.description,
        "is_public": project.is_public,
        "timezone": project.timezone,
        "archived_at": project.archived_at,
        "workspace_id": project.workspace_id,
        "workspace": {
            "id": workspace.id,
            "name": workspace.name,
            "slug": workspace.slug
        } if workspace else None,
        "total_issues": total_issues,
        "total_members": total_members,
        "default_assignee": default_assignee,
        "logo_url": None,
        "is_favorite": False,
        "created_at": project.created_at,
        "updated_at": project.updated_at,
        "deleted_at": project.deleted_at,
        "is_deleted": project.is_deleted,
        "created_by_id": project.created_by_id,
        "updated_by_id": project.updated_by_id,
    }