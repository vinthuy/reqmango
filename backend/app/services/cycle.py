"""
Cycle Services - 周期管理业务逻辑层
"""
from typing import Optional, List, Dict, Any
from datetime import datetime, date

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload
from sqlalchemy import and_, func

from app.models.cycle import Cycle
from app.models.issue import Issue
from app.models.project import Project
from app.models.state import State
from app.schemas.cycle import (
    CycleCreate,
    CycleUpdate,
    CycleResponse
)
from app.core.exceptions import NotFoundException, ValidationException


# ==================== Cycle Status Enum ====================

class CycleStatus:
    UPCOMING = "upcoming"
    ACTIVE = "active"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


# ==================== Cycle CRUD ====================

async def create_cycle(
    db: AsyncSession,
    cycle_data: CycleCreate,
    workspace_id: int,
    user_id: int
) -> Cycle:
    """创建周期"""
    # 验证项目存在
    project = await db.get(Project, cycle_data.project_id)
    if not project or project.is_deleted:
        raise NotFoundException("Project not found")
    
    # 验证日期范围
    if cycle_data.start_date and cycle_data.end_date:
        if cycle_data.start_date > cycle_data.end_date:
            raise ValidationException("Start date must be before end date")
    
    # 创建周期
    cycle = Cycle(
        name=cycle_data.name,
        description=cycle_data.description,
        start_date=cycle_data.start_date.date() if cycle_data.start_date else None,
        end_date=cycle_data.end_date.date() if cycle_data.end_date else None,
        project_id=cycle_data.project_id,
        workspace_id=workspace_id,
        created_by_id=user_id
    )
    
    db.add(cycle)
    await db.commit()
    await db.refresh(cycle)
    
    return cycle


async def get_cycle_by_id(db: AsyncSession, cycle_id: int) -> Cycle:
    """获取周期详情"""
    result = await db.execute(
        select(Cycle)
        .where(Cycle.id == cycle_id, Cycle.is_deleted == False)
        .options(selectinload(Cycle.project))
    )
    cycle = result.scalar_one_or_none()
    if not cycle:
        raise NotFoundException("Cycle not found")
    return cycle


async def list_project_cycles(
    db: AsyncSession,
    project_id: int,
    status: Optional[str] = None,
    include_completed: bool = False,
    limit: int = 50,
    offset: int = 0
) -> List[Cycle]:
    """列出项目的周期"""
    query = select(Cycle).where(
        Cycle.project_id == project_id,
        Cycle.is_deleted == False
    )
    
    # 状态筛选
    if status:
        query = query.where(Cycle.status == status)
    
    # 排序：活跃周期优先，然后按开始日期排序
    query = query.order_by(
        Cycle.start_date.desc()
    )
    
    # 分页
    query = query.limit(limit).offset(offset)
    
    result = await db.execute(query.options(selectinload(Cycle.project)))
    return list(result.scalars().all())


async def update_cycle(
    db: AsyncSession,
    cycle_id: int,
    update_data: CycleUpdate,
    user_id: int
) -> Cycle:
    """更新周期"""
    cycle = await get_cycle_by_id(db, cycle_id)
    
    # 更新字段
    if update_data.name is not None:
        cycle.name = update_data.name
    if update_data.description is not None:
        cycle.description = update_data.description
    if update_data.start_date is not None:
        cycle.start_date = update_data.start_date.date() if update_data.start_date else None
    if update_data.end_date is not None:
        cycle.end_date = update_data.end_date.date() if update_data.end_date else None
    
    # 验证日期范围
    if cycle.start_date and cycle.end_date:
        if cycle.start_date > cycle.end_date:
            raise ValidationException("Start date must be before end date")
    
    await db.commit()
    await db.refresh(cycle)
    
    return cycle


async def delete_cycle(db: AsyncSession, cycle_id: int):
    """删除周期（软删除）"""
    cycle = await get_cycle_by_id(db, cycle_id)
    cycle.is_deleted = True
    await db.commit()


# ==================== Cycle Status Management ====================

async def start_cycle(
    db: AsyncSession,
    cycle_id: int,
    user_id: int
) -> Cycle:
    """开始周期"""
    cycle = await get_cycle_by_id(db, cycle_id)
    
    if cycle.status == CycleStatus.ACTIVE:
        raise ValidationException("Cycle is already active")
    
    if cycle.status == CycleStatus.COMPLETED:
        raise ValidationException("Cannot start a completed cycle")
    
    cycle.status = CycleStatus.ACTIVE
    await db.commit()
    await db.refresh(cycle)
    
    return cycle


async def end_cycle(
    db: AsyncSession,
    cycle_id: int,
    user_id: int
) -> Cycle:
    """结束周期"""
    cycle = await get_cycle_by_id(db, cycle_id)
    
    if cycle.status == CycleStatus.COMPLETED:
        raise ValidationException("Cycle is already completed")
    
    if cycle.status == CycleStatus.CANCELLED:
        raise ValidationException("Cannot end a cancelled cycle")
    
    cycle.status = CycleStatus.COMPLETED
    cycle.completed_at = datetime.utcnow()
    await db.commit()
    await db.refresh(cycle)
    
    return cycle


async def cancel_cycle(
    db: AsyncSession,
    cycle_id: int,
    user_id: int
) -> Cycle:
    """取消周期"""
    cycle = await get_cycle_by_id(db, cycle_id)
    
    if cycle.status == CycleStatus.COMPLETED:
        raise ValidationException("Cannot cancel a completed cycle")
    
    cycle.status = CycleStatus.CANCELLED
    await db.commit()
    await db.refresh(cycle)
    
    return cycle


# ==================== Cycle Issue Management ====================

async def add_issue_to_cycle(
    db: AsyncSession,
    cycle_id: int,
    issue_id: int,
    user_id: int
) -> Dict[str, Any]:
    """将工作项添加到周期"""
    # 验证周期存在
    cycle = await get_cycle_by_id(db, cycle_id)
    
    # 验证工作项存在
    result = await db.execute(
        select(Issue).where(Issue.id == issue_id, Issue.is_deleted == False)
    )
    issue = result.scalar_one_or_none()
    if not issue:
        raise NotFoundException("Issue not found")
    
    # 验证工作项属于同一项目
    if issue.project_id != cycle.project_id:
        raise ValidationException("Issue does not belong to this cycle's project")
    
    # 检查是否已在周期中
    if issue.cycle_id == cycle_id:
        raise ValidationException("Issue is already in this cycle")
    
    # 更新工作项的周期
    issue.cycle_id = cycle_id
    await db.commit()
    
    return {
        "cycle_id": cycle_id,
        "issue_id": issue_id,
        "action": "added"
    }


async def remove_issue_from_cycle(
    db: AsyncSession,
    cycle_id: int,
    issue_id: int,
    user_id: int
) -> Dict[str, Any]:
    """从周期移除工作项"""
    # 验证周期存在
    cycle = await get_cycle_by_id(db, cycle_id)
    
    # 验证工作项存在
    result = await db.execute(
        select(Issue).where(Issue.id == issue_id, Issue.is_deleted == False)
    )
    issue = result.scalar_one_or_none()
    if not issue:
        raise NotFoundException("Issue not found")
    
    # 检查是否在周期中
    if issue.cycle_id != cycle_id:
        raise ValidationException("Issue is not in this cycle")
    
    # 移除工作项的周期关联
    issue.cycle_id = None
    await db.commit()
    
    return {
        "cycle_id": cycle_id,
        "issue_id": issue_id,
        "action": "removed"
    }


async def get_cycle_issues(
    db: AsyncSession,
    cycle_id: int,
    state_id: Optional[int] = None,
    priority: Optional[str] = None,
    limit: int = 50,
    offset: int = 0
) -> List[Issue]:
    """获取周期内的工作项"""
    # 验证周期存在
    cycle = await get_cycle_by_id(db, cycle_id)
    
    # 查询工作项
    query = select(Issue).where(
        Issue.cycle_id == cycle_id,
        Issue.is_deleted == False
    )
    
    # 状态筛选
    if state_id:
        query = query.where(Issue.state_id == state_id)
    
    # 优先级筛选
    if priority:
        query = query.where(Issue.priority == priority)
    
    # 排序
    query = query.order_by(Issue.sort_order, Issue.sequence_id.desc())
    
    # 分页
    query = query.limit(limit).offset(offset)
    
    result = await db.execute(query.options(
        selectinload(Issue.project),
        selectinload(Issue.state)
    ))
    
    return list(result.scalars().all())


# ==================== Cycle Progress & Statistics ====================

async def get_cycle_progress(
    db: AsyncSession,
    cycle_id: int
) -> Dict[str, Any]:
    """获取周期进度"""
    cycle = await get_cycle_by_id(db, cycle_id)
    
    # 统计总数
    total_result = await db.execute(
        select(func.count(Issue.id)).where(
            Issue.cycle_id == cycle_id,
            Issue.is_deleted == False
        )
    )
    total_issues = total_result.scalar_one_or_none() or 0
    
    # 统计已完成数（通过state group判断）
    completed_result = await db.execute(
        select(func.count(Issue.id))
        .join(State, Issue.state_id == State.id)
        .where(
            Issue.cycle_id == cycle_id,
            Issue.is_deleted == False,
            State.group == "done"
        )
    )
    completed_issues = completed_result.scalar_one_or_none() or 0
    
    # 计算进度
    progress = (completed_issues / total_issues * 100) if total_issues > 0 else 0
    
    # 按状态分组统计
    state_stats_result = await db.execute(
        select(State.name, State.group, func.count(Issue.id))
        .join(State, Issue.state_id == State.id)
        .where(
            Issue.cycle_id == cycle_id,
            Issue.is_deleted == False
        )
        .group_by(State.name, State.group)
    )
    state_stats = [
        {"state": row[0], "group": row[1], "count": row[2]}
        for row in state_stats_result.all()
    ]
    
    return {
        "cycle_id": cycle_id,
        "cycle_name": cycle.name,
        "total_issues": total_issues,
        "completed_issues": completed_issues,
        "progress": round(progress, 2),
        "state_breakdown": state_stats
    }


async def get_cycle_statistics(
    db: AsyncSession,
    cycle_id: int
) -> Dict[str, Any]:
    """获取周期详细统计"""
    progress = await get_cycle_progress(db, cycle_id)
    cycle = await get_cycle_by_id(db, cycle_id)
    
    # 优先级统计
    priority_result = await db.execute(
        select(Issue.priority, func.count(Issue.id))
        .where(
            Issue.cycle_id == cycle_id,
            Issue.is_deleted == False
        )
        .group_by(Issue.priority)
    )
    priority_stats = {row[0]: row[1] for row in priority_result.all()}
    
    # 工作项统计
    issue_stats_result = await db.execute(
        select(
            func.count(Issue.id).label('total'),
            func.count(Issue.start_date).label('with_start'),
            func.count(Issue.target_date).label('with_target'),
        )
        .where(
            Issue.cycle_id == cycle_id,
            Issue.is_deleted == False
        )
    )
    issue_stats = issue_stats_result.first()
    
    return {
        **progress,
        "priority_breakdown": priority_stats,
        "issue_stats": {
            "total": issue_stats.total if issue_stats else 0,
            "with_start_date": issue_stats.with_start if issue_stats else 0,
            "with_target_date": issue_stats.with_target if issue_stats else 0
        },
        "date_range": {
            "start_date": str(cycle.start_date) if cycle.start_date else None,
            "end_date": str(cycle.end_date) if cycle.end_date else None
        }
    }


# ==================== Burndown Chart Data ====================

async def get_burndown_data(
    db: AsyncSession,
    cycle_id: int
) -> Dict[str, Any]:
    """获取燃尽图数据"""
    cycle = await get_cycle_by_id(db, cycle_id)
    
    if not cycle.start_date or not cycle.end_date:
        raise ValidationException("Cycle does not have start and end dates")
    
    # 获取周期内的所有工作项
    total_result = await db.execute(
        select(func.count(Issue.id)).where(
            Issue.cycle_id == cycle_id,
            Issue.is_deleted == False
        )
    )
    total_issues = total_result.scalar_one_or_none() or 0
    
    # 估算每天应该完成的工作项
    total_days = (cycle.end_date - cycle.start_date).days
    if total_days <= 0:
        total_days = 1
    
    ideal_daily_burn = total_issues / total_days
    
    # 按日期统计每天完成的工作项
    # 这里简化处理，实际可能需要IssueActivity表来追踪真实完成日期
    today = date.today()
    days_elapsed = (today - cycle.start_date).days if today >= cycle.start_date else 0
    ideal_remaining = max(0, total_issues - (ideal_daily_burn * days_elapsed))
    
    # 实际完成数
    completed_result = await db.execute(
        select(func.count(Issue.id))
        .join(State, Issue.state_id == State.id)
        .where(
            Issue.cycle_id == cycle_id,
            Issue.is_deleted == False,
            State.group == "done"
        )
    )
    actual_completed = completed_result.scalar_one_or_none() or 0
    
    return {
        "cycle_id": cycle_id,
        "cycle_name": cycle.name,
        "start_date": str(cycle.start_date),
        "end_date": str(cycle.end_date),
        "total_issues": total_issues,
        "total_days": total_days,
        "days_elapsed": days_elapsed,
        "ideal_daily_burn": round(ideal_daily_burn, 2),
        "ideal_remaining": round(ideal_remaining, 2),
        "actual_completed": actual_completed,
        "actual_remaining": total_issues - actual_completed,
        "is_on_track": (total_issues - actual_completed) <= ideal_remaining
    }


# ==================== Helper Functions ====================

def build_cycle_response(cycle: Cycle, total_issues: int = 0, completed_issues: int = 0) -> Dict[str, Any]:
    """构建Cycle Response"""
    progress = (completed_issues / total_issues * 100) if total_issues > 0 else 0
    
    return {
        "id": cycle.id,
        "name": cycle.name,
        "description": cycle.description,
        "start_date": cycle.start_date,
        "end_date": cycle.end_date,
        "status": getattr(cycle, 'status', CycleStatus.UPCOMING),
        "progress": round(progress, 2),
        "total_issues": total_issues,
        "completed_issues": completed_issues,
        "project_id": cycle.project_id,
        "workspace_id": cycle.workspace_id,
        "created_at": cycle.created_at,
        "updated_at": cycle.updated_at,
    }