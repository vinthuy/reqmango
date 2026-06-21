from fastapi import APIRouter, Depends, Path, Query
from sqlalchemy.ext.asyncio import AsyncSession
from typing import List

from app.db.session import get_db
from app.api.deps import get_current_user
from app.models.user import User
from app.schemas.project import ProjectCreate, ProjectUpdate, ProjectResponse
from app.services import project as project_service
from app.core.exceptions import NotFoundException, ConflictException

router = APIRouter()


# ==================== Project CRUD ====================

@router.post("/", response_model=ProjectResponse, status_code=201)
async def create_project(
    workspace_id: int,
    project_data: ProjectCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建项目
    
    在指定工作空间中创建一个新的项目。
    """
    project = await project_service.create_project(
        db=db,
        project_data=project_data,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    return await project_service.build_project_response(db, project)


@router.get("/", response_model=List[ProjectResponse])
async def list_projects(
    workspace_id: int,
    include_archived: bool = False,
    limit: int = Query(50, ge=1, le=100),
    offset: int = Query(0, ge=0),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出工作空间的项目
    
    返回工作空间下的所有项目，支持分页。
    """
    projects = await project_service.list_workspace_projects(
        db=db,
        workspace_id=workspace_id,
        include_archived=include_archived,
        limit=limit,
        offset=offset
    )
    return [await project_service.build_project_response(db, p) for p in projects]


@router.get("/{project_id}", response_model=ProjectResponse)
async def get_project(
    project_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取项目详情
    
    返回项目的完整信息。
    """
    project = await project_service.get_project_by_id(db, project_id)
    return await project_service.build_project_response(db, project)


@router.patch("/{project_id}", response_model=ProjectResponse)
async def update_project(
    project_id: int,
    update_data: ProjectUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新项目
    
    可以更新项目的名称、描述、是否公开等属性。
    """
    project = await project_service.update_project(
        db=db,
        project_id=project_id,
        update_data=update_data
    )
    return await project_service.build_project_response(db, project)


@router.delete("/{project_id}", status_code=204)
async def delete_project(
    project_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除项目（软删除）
    
    项目不会真正删除，而是标记为已删除状态。
    """
    await project_service.delete_project(db, project_id)
    return None


# ==================== Project Archive ====================

@router.post("/{project_id}/archive", response_model=ProjectResponse)
async def archive_project(
    project_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    归档项目
    
    归档后的项目将从常规列表中隐藏。
    """
    project = await project_service.archive_project(db, project_id)
    return await project_service.build_project_response(db, project)


@router.post("/{project_id}/restore", response_model=ProjectResponse)
async def restore_project(
    project_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    恢复项目
    
    恢复已归档的项目。
    """
    project = await project_service.restore_project(db, project_id)
    return await project_service.build_project_response(db, project)


# ==================== Project Members ====================

@router.get("/{project_id}/members", response_model=List[dict])
async def list_project_members(
    project_id: int,
    only_active: bool = True,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出项目成员
    
    返回项目的所有成员及其角色。
    """
    members = await project_service.list_project_members(
        db=db,
        project_id=project_id,
        only_active=only_active
    )
    return [
        {
            "id": member.id,
            "user_id": member.user_id,
            "role": member.role,
            "is_active": member.is_active,
            "user": {
                "id": member.user.id,
                "display_name": member.user.display_name,
                "username": member.user.username,
                "avatar": member.user.avatar
            } if member.user else None
        }
        for member in members
    ]


@router.post("/{project_id}/members", response_model=dict)
async def add_project_member(
    project_id: int,
    user_id: int,
    role: int = Query(15, ge=1, le=20),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    添加项目成员
    
    将指定用户添加为项目成员。
    """
    member = await project_service.add_project_member(
        db=db,
        project_id=project_id,
        user_id=user_id,
        role=role,
        added_by=current_user.id
    )
    return {
        "id": member.id,
        "project_id": member.project_id,
        "user_id": member.user_id,
        "role": member.role
    }


@router.patch("/{project_id}/members/{user_id}", response_model=dict)
async def update_project_member(
    project_id: int,
    user_id: int,
    role: int = Query(..., ge=1, le=20),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新项目成员角色
    """
    member = await project_service.update_project_member(
        db=db,
        project_id=project_id,
        user_id=user_id,
        role=role,
        updated_by=current_user.id
    )
    return {
        "id": member.id,
        "project_id": member.project_id,
        "user_id": member.user_id,
        "role": member.role
    }


@router.delete("/{project_id}/members/{user_id}", response_model=dict)
async def remove_project_member(
    project_id: int,
    user_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    移除项目成员
    
    从项目中移除指定成员。
    """
    result = await project_service.remove_project_member(
        db=db,
        project_id=project_id,
        user_id=user_id
    )
    return result


# ==================== Project Statistics ====================

@router.get("/{project_id}/statistics", response_model=dict)
async def get_project_statistics(
    project_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取项目统计
    
    返回项目的工作项统计、进度、成员数等。
    """
    stats = await project_service.get_project_statistics(db, project_id)
    return stats


@router.get("/{project_id}/issues-summary", response_model=dict)
async def get_project_issues_summary(
    project_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取项目工作项摘要
    
    按状态分组返回工作项数量。
    """
    summary = await project_service.get_project_issues_summary(db, project_id)
    return summary