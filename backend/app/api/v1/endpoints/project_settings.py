"""
Project Settings API Endpoints - 项目设置管理接口（IssueType、State、Label）
"""
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from uuid import UUID
from typing import Optional, List

from app.db.session import get_db
from app.api.deps import get_current_user
from app.models.user import User
from app.schemas.issue_type import (
    IssueTypeCreate,
    IssueTypeUpdate,
    IssueTypeResponse,
    IssueTypeLite,
    StateCreate,
    StateUpdate,
    StateResponse,
    StateLite,
    LabelCreate,
    LabelUpdate,
    LabelResponse,
    LabelLite
)
from app.services import project_settings as settings_service
from app.core.exceptions import NotFoundException

router = APIRouter()


# ==================== IssueType Endpoints ====================

@router.post("/issue-types", response_model=IssueTypeResponse, status_code=201)
async def create_issue_type(
    project_id: UUID,
    workspace_id: UUID,
    type_data: IssueTypeCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建工作项类型
    
    为项目创建一个新的工作项类型（如 Issue、Task、Bug 等）。
    """
    # 确保 type_data.project_id 与传入的 project_id 一致
    type_data.project_id = project_id
    
    issue_type = await settings_service.create_issue_type(
        db=db,
        type_data=type_data,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    return issue_type


@router.get("/issue-types", response_model=List[IssueTypeResponse])
async def list_issue_types(
    project_id: UUID,
    include_inactive: bool = False,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出项目的工作项类型
    """
    issue_types = await settings_service.get_project_issue_types(
        db=db,
        project_id=project_id,
        include_inactive=include_inactive
    )
    return issue_types


@router.get("/issue-types/{type_id}", response_model=IssueTypeResponse)
async def get_issue_type(
    type_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取工作项类型详情
    """
    issue_type = await settings_service.get_issue_type_by_id(db, type_id)
    return issue_type


@router.put("/issue-types/{type_id}", response_model=IssueTypeResponse)
async def update_issue_type(
    type_id: UUID,
    update_data: IssueTypeUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新工作项类型
    """
    issue_type = await settings_service.update_issue_type(
        db=db,
        type_id=type_id,
        update_data=update_data
    )
    return issue_type


@router.delete("/issue-types/{type_id}", status_code=204)
async def delete_issue_type(
    type_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除工作项类型（软删除）
    """
    await settings_service.delete_issue_type(db, type_id)
    return None


@router.post("/issue-types/default", response_model=List[IssueTypeResponse], status_code=201)
async def create_default_issue_types(
    project_id: UUID,
    workspace_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建默认的工作项类型
    
    自动创建 Issue、Task、Bug、Story、Epic 五种默认类型。
    """
    issue_types = await settings_service.create_default_issue_types(
        db=db,
        project_id=project_id,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    return issue_types


# ==================== State Endpoints ====================

@router.post("/states", response_model=StateResponse, status_code=201)
async def create_state(
    project_id: UUID,
    workspace_id: UUID,
    state_data: StateCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建状态
    
    为项目创建一个新的工作流状态。
    """
    state_data.project_id = project_id
    
    state = await settings_service.create_state(
        db=db,
        state_data=state_data,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    return state


@router.get("/states", response_model=List[StateResponse])
async def list_states(
    project_id: UUID,
    include_inactive: bool = False,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出项目的状态
    """
    states = await settings_service.get_project_states(
        db=db,
        project_id=project_id,
        include_inactive=include_inactive
    )
    return states


@router.get("/states/{state_id}", response_model=StateResponse)
async def get_state(
    state_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取状态详情
    """
    state = await settings_service.get_state_by_id(db, state_id)
    return state


@router.put("/states/{state_id}", response_model=StateResponse)
async def update_state(
    state_id: UUID,
    update_data: StateUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新状态
    """
    state = await settings_service.update_state(
        db=db,
        state_id=state_id,
        update_data=update_data
    )
    return state


@router.delete("/states/{state_id}", status_code=204)
async def delete_state(
    state_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除状态（软删除）
    """
    await settings_service.delete_state(db, state_id)
    return None


@router.post("/states/default", response_model=List[StateResponse], status_code=201)
async def create_default_states(
    project_id: UUID,
    workspace_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建默认状态
    
    自动创建 Backlog、Todo、In Progress、In Review、Done、Cancelled 六种默认状态。
    """
    states = await settings_service.create_default_states(
        db=db,
        project_id=project_id,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    return states


# ==================== Label Endpoints ====================

@router.post("/labels", response_model=LabelResponse, status_code=201)
async def create_label(
    project_id: UUID,
    workspace_id: UUID,
    label_data: LabelCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建标签
    
    为项目创建一个新的标签，用于分类工作项。
    """
    label_data.project_id = project_id
    
    label = await settings_service.create_label(
        db=db,
        label_data=label_data,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    return label


@router.get("/labels", response_model=List[LabelResponse])
async def list_labels(
    project_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出项目的标签
    """
    labels = await settings_service.get_project_labels(db, project_id)
    return labels


@router.get("/labels/{label_id}", response_model=LabelResponse)
async def get_label(
    label_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取标签详情
    """
    label = await settings_service.get_label_by_id(db, label_id)
    return label


@router.put("/labels/{label_id}", response_model=LabelResponse)
async def update_label(
    label_id: UUID,
    update_data: LabelUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新标签
    """
    label = await settings_service.update_label(
        db=db,
        label_id=label_id,
        update_data=update_data
    )
    return label


@router.delete("/labels/{label_id}", status_code=204)
async def delete_label(
    label_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除标签（软删除）
    """
    await settings_service.delete_label(db, label_id)
    return None