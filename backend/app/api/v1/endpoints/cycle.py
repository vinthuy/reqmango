"""
Cycle API Endpoints - 周期管理接口
"""
from fastapi import APIRouter, Depends, Query
from sqlalchemy.ext.asyncio import AsyncSession
from typing import Optional, List

from app.db.session import get_db
from app.api.deps import get_current_user
from app.models.user import User
from app.models.project import Project
from app.schemas.cycle import (
    CycleCreate,
    CycleUpdate,
    CycleResponse
)
from app.services import cycle as cycle_service
from app.core.exceptions import NotFoundException, ValidationException

router = APIRouter()


# ==================== Cycle CRUD ====================

@router.post("/", response_model=CycleResponse, status_code=201)
async def create_cycle(
    workspace_id: int,
    cycle_data: CycleCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建周期
    
    在指定项目中创建一个新的迭代周期。
    """
    cycle = await cycle_service.create_cycle(
        db=db,
        cycle_data=cycle_data,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    
    return cycle_service.build_cycle_response(cycle)


@router.get("/", response_model=List[CycleResponse])
async def list_cycles(
    project_id: int,
    workspace_id: int,
    status: Optional[str] = None,
    include_completed: bool = False,
    limit: int = Query(50, ge=1, le=100),
    offset: int = Query(0, ge=0),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出项目的周期
    
    支持按状态筛选和分页。
    """
    cycles = await cycle_service.list_project_cycles(
        db=db,
        project_id=project_id,
        status=status,
        include_completed=include_completed,
        limit=limit,
        offset=offset
    )
    
    # 获取每个周期的统计信息
    results = []
    for cycle in cycles:
        progress = await cycle_service.get_cycle_progress(db, cycle.id)
        results.append(cycle_service.build_cycle_response(
            cycle,
            progress.get('total_issues', 0),
            progress.get('completed_issues', 0)
        ))
    
    return results


@router.get("/{cycle_id}", response_model=CycleResponse)
async def get_cycle(
    cycle_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取周期详情
    
    返回周期的完整信息，包括进度统计。
    """
    cycle = await cycle_service.get_cycle_by_id(db, cycle_id)
    progress = await cycle_service.get_cycle_progress(db, cycle_id)
    
    return cycle_service.build_cycle_response(
        cycle,
        progress.get('total_issues', 0),
        progress.get('completed_issues', 0)
    )


@router.put("/{cycle_id}", response_model=CycleResponse)
async def update_cycle(
    cycle_id: int,
    update_data: CycleUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新周期
    
    可以更新周期的名称、描述、开始/结束日期。
    """
    cycle = await cycle_service.update_cycle(
        db=db,
        cycle_id=cycle_id,
        update_data=update_data,
        user_id=current_user.id
    )
    
    progress = await cycle_service.get_cycle_progress(db, cycle_id)
    return cycle_service.build_cycle_response(
        cycle,
        progress.get('total_issues', 0),
        progress.get('completed_issues', 0)
    )


@router.delete("/{cycle_id}", status_code=204)
async def delete_cycle(
    cycle_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除周期（软删除）
    
    周期不会真正删除，而是标记为已删除状态。
    """
    await cycle_service.delete_cycle(db, cycle_id)
    return None


# ==================== Cycle Status Management ====================

@router.post("/{cycle_id}/start", response_model=CycleResponse)
async def start_cycle(
    cycle_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    开始周期
    
    将周期状态从 upcoming 设置为 active。
    """
    cycle = await cycle_service.start_cycle(
        db=db,
        cycle_id=cycle_id,
        user_id=current_user.id
    )
    
    progress = await cycle_service.get_cycle_progress(db, cycle_id)
    return cycle_service.build_cycle_response(
        cycle,
        progress.get('total_issues', 0),
        progress.get('completed_issues', 0)
    )


@router.post("/{cycle_id}/end", response_model=CycleResponse)
async def end_cycle(
    cycle_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    结束周期
    
    将周期状态从 active 设置为 completed。
    """
    cycle = await cycle_service.end_cycle(
        db=db,
        cycle_id=cycle_id,
        user_id=current_user.id
    )
    
    progress = await cycle_service.get_cycle_progress(db, cycle_id)
    return cycle_service.build_cycle_response(
        cycle,
        progress.get('total_issues', 0),
        progress.get('completed_issues', 0)
    )


@router.post("/{cycle_id}/cancel", response_model=CycleResponse)
async def cancel_cycle(
    cycle_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    取消周期
    
    将周期状态设置为 cancelled。
    """
    cycle = await cycle_service.cancel_cycle(
        db=db,
        cycle_id=cycle_id,
        user_id=current_user.id
    )
    
    progress = await cycle_service.get_cycle_progress(db, cycle_id)
    return cycle_service.build_cycle_response(
        cycle,
        progress.get('total_issues', 0),
        progress.get('completed_issues', 0)
    )


# ==================== Cycle Issue Management ====================

@router.post("/{cycle_id}/issues", response_model=dict)
async def add_issue_to_cycle(
    cycle_id: int,
    issue_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    将工作项添加到周期
    """
    result = await cycle_service.add_issue_to_cycle(
        db=db,
        cycle_id=cycle_id,
        issue_id=issue_id,
        user_id=current_user.id
    )
    return result


@router.delete("/{cycle_id}/issues/{issue_id}", response_model=dict)
async def remove_issue_from_cycle(
    cycle_id: int,
    issue_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    从周期移除工作项
    """
    result = await cycle_service.remove_issue_from_cycle(
        db=db,
        cycle_id=cycle_id,
        issue_id=issue_id,
        user_id=current_user.id
    )
    return result


@router.get("/{cycle_id}/issues", response_model=List[dict])
async def get_cycle_issues(
    cycle_id: int,
    state_id: Optional[int] = None,
    priority: Optional[str] = None,
    limit: int = Query(50, ge=1, le=100),
    offset: int = Query(0, ge=0),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取周期内的工作项
    
    支持按状态和优先级筛选。
    """
    from app.services.issue import build_issue_response
    
    issues = await cycle_service.get_cycle_issues(
        db=db,
        cycle_id=cycle_id,
        state_id=state_id,
        priority=priority,
        limit=limit,
        offset=offset
    )
    
    return [build_issue_response(issue) for issue in issues]


# ==================== Cycle Progress & Statistics ====================

@router.get("/{cycle_id}/progress", response_model=dict)
async def get_cycle_progress(
    cycle_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取周期进度
    
    返回周期的完成进度和状态分布统计。
    """
    progress = await cycle_service.get_cycle_progress(db, cycle_id)
    return progress


@router.get("/{cycle_id}/statistics", response_model=dict)
async def get_cycle_statistics(
    cycle_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取周期详细统计
    
    返回周期的完整统计数据，包括进度、优先级分布等。
    """
    stats = await cycle_service.get_cycle_statistics(db, cycle_id)
    return stats


@router.get("/{cycle_id}/burndown", response_model=dict)
async def get_burndown_data(
    cycle_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取燃尽图数据
    
    返回燃尽图所需的每日数据点。
    """
    burndown = await cycle_service.get_burndown_data(db, cycle_id)
    return burndown