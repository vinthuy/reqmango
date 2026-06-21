"""
Estimate Point API - 估算点API端点
"""
from typing import List

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.session import get_db
from app.api.deps import get_current_user
from app.models.user import User
from app.schemas.estimate_point import (
    EstimatePointCreate,
    EstimatePointUpdate,
    EstimatePointResponse,
    EstimatePointBulkCreate,
    EstimatePointReorder
)
from app.services import estimate_point

router = APIRouter(prefix="/estimate-points", tags=["估算点"])


@router.post("/", response_model=EstimatePointResponse, status_code=201)
async def create_estimate_point(
    project_id: int,
    point_data: EstimatePointCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建估算点

    为指定项目创建一个新的估算点选项。
    """
    point_data.project_id = project_id
    point = await estimate_point.create_estimate_point(
        db=db,
        point_data=point_data,
        user_id=current_user.id
    )
    return point


@router.get("/", response_model=List[EstimatePointResponse])
async def list_estimate_points(
    project_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出项目的所有估算点

    返回项目中所有可用的估算点选项，按顺序排列。
    """
    points = await estimate_point.get_project_estimate_points(db, project_id)
    return points


@router.get("/default", response_model=EstimatePointResponse)
async def get_default_estimate_point(
    project_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取项目的默认估算点

    返回项目当前标记为默认的估算点。
    """
    point = await estimate_point.get_default_estimate_point(db, project_id)
    if not point:
        raise HTTPException(status_code=404, detail="No default estimate point found")
    return point


@router.get("/{point_id}", response_model=EstimatePointResponse)
async def get_estimate_point(
    point_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取估算点详情
    """
    point = await estimate_point.get_estimate_point_by_id(db, point_id)
    return point


@router.patch("/{point_id}", response_model=EstimatePointResponse)
async def update_estimate_point(
    point_id: int,
    update_data: EstimatePointUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新估算点

    部分更新估算点的属性。
    """
    point = await estimate_point.update_estimate_point(db, point_id, update_data)
    return point


@router.delete("/{point_id}", status_code=204)
async def delete_estimate_point(
    point_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除估算点

    软删除估算点。
    """
    await estimate_point.delete_estimate_point(db, point_id)


@router.post("/reorder", response_model=List[EstimatePointResponse])
async def reorder_estimate_points(
    project_id: int,
    reorder_data: EstimatePointReorder,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    重新排序估算点

    传入估算点ID的新顺序，系统将更新所有估算点的sequence值。
    """
    points = await estimate_point.reorder_estimate_points(
        db=db,
        project_id=project_id,
        point_ids=reorder_data.point_ids
    )
    return points


@router.post("/bulk", response_model=List[EstimatePointResponse])
async def bulk_create_estimate_points(
    project_id: int,
    bulk_data: EstimatePointBulkCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    批量创建估算点

    为项目一次性创建多个估算点选项。
    """
    # 确保所有点都属于该项目
    for point in bulk_data.points:
        point.project_id = project_id

    points = []
    for point_data in bulk_data.points:
        point = await estimate_point.create_estimate_point(
            db=db,
            point_data=point_data,
            user_id=current_user.id
        )
        points.append(point)

    return points


@router.post("/defaults", response_model=List[EstimatePointResponse])
async def create_default_estimate_points(
    project_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建默认估算点

    为项目创建敏捷Scrum标准的斐波那契数列估算点。
    """
    points = await estimate_point.create_default_estimate_points(
        db=db,
        project_id=project_id,
        user_id=current_user.id
    )
    return points
