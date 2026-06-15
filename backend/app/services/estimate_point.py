"""
Estimate Point Service - 估算点业务逻辑层
"""
from typing import List, Optional
from uuid import UUID

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload

from app.models.estimate_point import EstimatePoint
from app.models.project import Project
from app.schemas.estimate_point import (
    EstimatePointCreate,
    EstimatePointUpdate
)
from app.core.exceptions import NotFoundException, ValidationException


async def create_estimate_point(
    db: AsyncSession,
    point_data: EstimatePointCreate,
    user_id: UUID
) -> EstimatePoint:
    """创建估算点"""
    # 验证项目存在
    project = await db.get(Project, point_data.project_id)
    if not project or project.is_deleted:
        raise NotFoundException("Project not found")

    # 如果设置为默认，取消其他默认
    if point_data.is_default:
        await _clear_default_points(db, point_data.project_id)

    # 如果是第一个点，设置为默认
    existing = await get_project_estimate_points(db, point_data.project_id)
    if len(existing) == 0:
        point_data.is_default = True

    point = EstimatePoint(
        name=point_data.name,
        value=point_data.value,
        is_default=point_data.is_default,
        sequence=point_data.sequence,
        project_id=point_data.project_id,
        created_by_id=user_id
    )

    db.add(point)
    await db.commit()
    await db.refresh(point)
    return point


async def get_estimate_point_by_id(
    db: AsyncSession,
    point_id: UUID
) -> EstimatePoint:
    """获取估算点"""
    result = await db.execute(
        select(EstimatePoint).where(EstimatePoint.id == point_id)
    )
    point = result.scalar_one_or_none()
    if not point or point.is_deleted:
        raise NotFoundException("Estimate point not found")
    return point


async def get_project_estimate_points(
    db: AsyncSession,
    project_id: UUID
) -> List[EstimatePoint]:
    """获取项目的所有估算点"""
    result = await db.execute(
        select(EstimatePoint)
        .where(
            EstimatePoint.project_id == project_id,
            EstimatePoint.is_deleted == False
        )
        .order_by(EstimatePoint.sequence)
    )
    return list(result.scalars().all())


async def get_default_estimate_point(
    db: AsyncSession,
    project_id: UUID
) -> Optional[EstimatePoint]:
    """获取项目的默认估算点"""
    result = await db.execute(
        select(EstimatePoint)
        .where(
            EstimatePoint.project_id == project_id,
            EstimatePoint.is_default == True,
            EstimatePoint.is_deleted == False
        )
    )
    return result.scalar_one_or_none()


async def update_estimate_point(
    db: AsyncSession,
    point_id: UUID,
    update_data: EstimatePointUpdate
) -> EstimatePoint:
    """更新估算点"""
    point = await get_estimate_point_by_id(db, point_id)

    if update_data.name is not None:
        point.name = update_data.name
    if update_data.value is not None:
        point.value = update_data.value
    if update_data.sequence is not None:
        point.sequence = update_data.sequence

    # 处理默认切换
    if update_data.is_default is not None:
        if update_data.is_default:
            await _clear_default_points(db, point.project_id)
            point.is_default = True
        else:
            point.is_default = False

    await db.commit()
    await db.refresh(point)
    return point


async def delete_estimate_point(
    db: AsyncSession,
    point_id: UUID
) -> None:
    """删除估算点（软删除）"""
    point = await get_estimate_point_by_id(db, point_id)
    point.is_deleted = True
    await db.commit()


async def reorder_estimate_points(
    db: AsyncSession,
    project_id: UUID,
    point_ids: List[UUID]
) -> List[EstimatePoint]:
    """重新排序估算点"""
    # 验证所有点都属于该项目
    points = await get_project_estimate_points(db, project_id)
    point_map = {p.id: p for p in points}

    for i, pid in enumerate(point_ids):
        if pid in point_map:
            point_map[pid].sequence = i

    await db.commit()

    # 返回更新后的列表
    return await get_project_estimate_points(db, project_id)


async def _clear_default_points(
    db: AsyncSession,
    project_id: UUID
) -> None:
    """清除项目的所有默认标记"""
    result = await db.execute(
        select(EstimatePoint)
        .where(
            EstimatePoint.project_id == project_id,
            EstimatePoint.is_default == True,
            EstimatePoint.is_deleted == False
        )
    )
    for point in result.scalars():
        point.is_default = False


async def create_default_estimate_points(
    db: AsyncSession,
    project_id: UUID,
    user_id: UUID
) -> List[EstimatePoint]:
    """为项目创建默认估算点（敏捷Scrum标准）"""
    # 标准斐波那契数列
    default_points = [
        {"name": "0 - 不需要估算", "value": 0, "is_default": False, "sequence": 0},
        {"name": "1 - 很小", "value": 1, "is_default": False, "sequence": 1},
        {"name": "2 - 小", "value": 2, "is_default": False, "sequence": 2},
        {"name": "3 - 中等", "value": 3, "is_default": True, "sequence": 3},
        {"name": "5 - 较大", "value": 5, "is_default": False, "sequence": 4},
        {"name": "8 - 大", "value": 8, "is_default": False, "sequence": 5},
        {"name": "13 - 很大", "value": 13, "is_default": False, "sequence": 6},
        {"name": "21 - 巨大", "value": 21, "is_default": False, "sequence": 7},
    ]

    points = []
    for data in default_points:
        point = EstimatePoint(
            name=data["name"],
            value=data["value"],
            is_default=data["is_default"],
            sequence=data["sequence"],
            project_id=project_id,
            created_by_id=user_id
        )
        db.add(point)
        points.append(point)

    await db.commit()
    return points
