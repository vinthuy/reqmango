"""
Module API Endpoints - 模块管理接口
"""
from fastapi import APIRouter, Depends, Query
from sqlalchemy.ext.asyncio import AsyncSession
from uuid import UUID
from typing import Optional, List

from app.db.session import get_db
from app.api.deps import get_current_user
from app.models.user import User
from app.models.project import Project
from app.schemas.module import (
    ModuleCreate,
    ModuleUpdate,
    ModuleResponse
)
from app.services import module as module_service
from app.core.exceptions import NotFoundException, ValidationException

router = APIRouter()


# ==================== Module CRUD ====================

@router.post("/", response_model=ModuleResponse, status_code=201)
async def create_module(
    workspace_id: UUID,
    module_data: ModuleCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建模块
    
    在指定项目中创建一个新的功能模块。
    """
    module = await module_service.create_module(
        db=db,
        module_data=module_data,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    
    return module_service.build_module_response(module)


@router.get("/", response_model=List[ModuleResponse])
async def list_modules(
    project_id: UUID,
    workspace_id: UUID,
    parent_id: Optional[UUID] = None,
    include_archived: bool = False,
    limit: int = Query(50, ge=1, le=100),
    offset: int = Query(0, ge=0),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出项目的模块
    
    支持按父模块筛选和分页。
    """
    modules = await module_service.list_project_modules(
        db=db,
        project_id=project_id,
        parent_id=parent_id,
        include_archived=include_archived,
        limit=limit,
        offset=offset
    )
    
    # 获取每个模块的统计信息
    results = []
    for module in modules:
        progress = await module_service.get_module_progress(db, module.id)
        results.append(module_service.build_module_response(
            module,
            progress.get('total_issues', 0),
            progress.get('completed_issues', 0)
        ))
    
    return results


@router.get("/{module_id}", response_model=ModuleResponse)
async def get_module(
    module_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取模块详情
    
    返回模块的完整信息，包括进度统计。
    """
    module = await module_service.get_module_by_id(db, module_id)
    progress = await module_service.get_module_progress(db, module_id)
    
    return module_service.build_module_response(
        module,
        progress.get('total_issues', 0),
        progress.get('completed_issues', 0)
    )


@router.put("/{module_id}", response_model=ModuleResponse)
async def update_module(
    module_id: UUID,
    update_data: ModuleUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新模块
    
    可以更新模块的名称、描述、目标日期。
    """
    module = await module_service.update_module(
        db=db,
        module_id=module_id,
        update_data=update_data,
        user_id=current_user.id
    )
    
    progress = await module_service.get_module_progress(db, module_id)
    return module_service.build_module_response(
        module,
        progress.get('total_issues', 0),
        progress.get('completed_issues', 0)
    )


@router.delete("/{module_id}", status_code=204)
async def delete_module(
    module_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除模块（软删除）
    
    模块不会真正删除，而是标记为已删除状态。
    """
    await module_service.delete_module(db, module_id)
    return None


# ==================== Module Issue Management ====================

@router.post("/{module_id}/issues", response_model=dict)
async def add_issue_to_module(
    module_id: UUID,
    issue_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    将工作项添加到模块
    """
    result = await module_service.add_issue_to_module(
        db=db,
        module_id=module_id,
        issue_id=issue_id,
        user_id=current_user.id
    )
    return result


@router.delete("/{module_id}/issues/{issue_id}", response_model=dict)
async def remove_issue_from_module(
    module_id: UUID,
    issue_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    从模块移除工作项
    """
    result = await module_service.remove_issue_from_module(
        db=db,
        module_id=module_id,
        issue_id=issue_id,
        user_id=current_user.id
    )
    return result


@router.get("/{module_id}/issues", response_model=List[dict])
async def get_module_issues(
    module_id: UUID,
    state_id: Optional[UUID] = None,
    priority: Optional[str] = None,
    limit: int = Query(50, ge=1, le=100),
    offset: int = Query(0, ge=0),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取模块内的工作项
    
    支持按状态和优先级筛选。
    """
    from app.services.issue import build_issue_response
    
    issues = await module_service.get_module_issues(
        db=db,
        module_id=module_id,
        state_id=state_id,
        priority=priority,
        limit=limit,
        offset=offset
    )
    
    return [build_issue_response(issue) for issue in issues]


# ==================== Module Progress & Statistics ====================

@router.get("/{module_id}/progress", response_model=dict)
async def get_module_progress(
    module_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取模块进度
    
    返回模块的完成进度和状态分布统计。
    """
    progress = await module_service.get_module_progress(db, module_id)
    return progress


@router.get("/{module_id}/statistics", response_model=dict)
async def get_module_statistics(
    module_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取模块详细统计
    
    返回模块的完整统计数据，包括进度、优先级分布等。
    """
    stats = await module_service.get_module_statistics(db, module_id)
    return stats


# ==================== Module Tree ====================

@router.get("/tree", response_model=List[dict])
async def get_module_tree(
    project_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取模块树形结构
    
    返回项目的模块层级树，包括子模块。
    """
    tree = await module_service.get_module_tree(db, project_id)
    return tree