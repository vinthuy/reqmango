"""
Project Settings Services - 项目设置相关服务（IssueType、State、Label）
"""
from typing import Optional, List

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload

from app.models.issue_type import IssueType
from app.models.state import State
from app.models.label import Label
from app.models.project import Project
from app.schemas.issue_type import (
    IssueTypeCreate,
    IssueTypeUpdate,
    StateCreate,
    StateUpdate,
    LabelCreate,
    LabelUpdate,
    DEFAULT_ISSUE_TYPES,
    DEFAULT_STATES
)
from app.core.exceptions import NotFoundException, ValidationException


# ==================== IssueType Service ====================

async def create_issue_type(
    db: AsyncSession,
    type_data: IssueTypeCreate,
    workspace_id: int,
    user_id: int
) -> IssueType:
    """创建工作项类型"""
    # 验证项目存在
    project = await db.get(Project, type_data.project_id)
    if not project or project.is_deleted:
        raise NotFoundException("Project not found")
    
    # 创建类型
    issue_type = IssueType(
        name=type_data.name,
        color=type_data.color,
        icon=type_data.icon,
        is_default=type_data.is_default,
        sequence=type_data.sequence,
        is_active=type_data.is_active,
        project_id=type_data.project_id,
        workspace_id=workspace_id,
        created_by_id=user_id
    )
    
    db.add(issue_type)
    await db.commit()
    await db.refresh(issue_type)
    return issue_type


async def create_default_issue_types(
    db: AsyncSession,
    project_id: int,
    workspace_id: int,
    user_id: int
) -> List[IssueType]:
    """创建默认的工作项类型"""
    types = []
    for default_type in DEFAULT_ISSUE_TYPES:
        type_data = IssueTypeCreate(
            name=default_type['name'],
            color=default_type['color'],
            icon=default_type['icon'],
            is_default=default_type.get('is_default', False),
            sequence=default_type['sequence'],
            is_active=True,
            project_id=project_id
        )
        issue_type = await create_issue_type(db, type_data, workspace_id, user_id)
        types.append(issue_type)
    return types


async def get_issue_type_by_id(db: AsyncSession, type_id: int) -> IssueType:
    """获取工作项类型"""
    result = await db.execute(
        select(IssueType).where(IssueType.id == type_id)
    )
    issue_type = result.scalar_one_or_none()
    if not issue_type or issue_type.is_deleted:
        raise NotFoundException("Issue type not found")
    return issue_type


async def get_project_issue_types(
    db: AsyncSession,
    project_id: int,
    include_inactive: bool = False
) -> List[IssueType]:
    """获取项目的工作项类型列表"""
    query = select(IssueType).where(
        IssueType.project_id == project_id,
        IssueType.is_deleted == False
    )
    
    if not include_inactive:
        query = query.where(IssueType.is_active == True)
    
    query = query.order_by(IssueType.sequence)
    
    result = await db.execute(query)
    return list(result.scalars().all())


async def update_issue_type(
    db: AsyncSession,
    type_id: int,
    update_data: IssueTypeUpdate
) -> IssueType:
    """更新工作项类型"""
    issue_type = await get_issue_type_by_id(db, type_id)
    
    if update_data.name is not None:
        issue_type.name = update_data.name
    if update_data.color is not None:
        issue_type.color = update_data.color
    if update_data.icon is not None:
        issue_type.icon = update_data.icon
    if update_data.is_default is not None:
        # 如果设置为默认，需要取消其他默认
        if update_data.is_default:
            await _unset_other_defaults(db, issue_type.project_id)
        issue_type.is_default = update_data.is_default
    if update_data.sequence is not None:
        issue_type.sequence = update_data.sequence
    if update_data.is_active is not None:
        issue_type.is_active = update_data.is_active
    
    await db.commit()
    await db.refresh(issue_type)
    return issue_type


async def _unset_other_defaults(db: AsyncSession, project_id: int):
    """取消项目中其他类型的默认状态"""
    result = await db.execute(
        select(IssueType).where(
            IssueType.project_id == project_id,
            IssueType.is_default == True
        )
    )
    for issue_type in result.scalars().all():
        issue_type.is_default = False
    await db.flush()


async def delete_issue_type(db: AsyncSession, type_id: int):
    """删除工作项类型（软删除）"""
    issue_type = await get_issue_type_by_id(db, type_id)
    issue_type.is_deleted = True
    await db.commit()


# ==================== State Service ====================

async def create_state(
    db: AsyncSession,
    state_data: StateCreate,
    workspace_id: int,
    user_id: int
) -> State:
    """创建状态"""
    # 验证项目存在
    project = await db.get(Project, state_data.project_id)
    if not project or project.is_deleted:
        raise NotFoundException("Project not found")
    
    state = State(
        name=state_data.name,
        color=state_data.color,
        group=state_data.group.value,
        sequence=state_data.sequence,
        is_active=state_data.is_active,
        project_id=state_data.project_id,
        workspace_id=workspace_id,
        created_by_id=user_id
    )
    
    db.add(state)
    await db.commit()
    await db.refresh(state)
    return state


async def create_default_states(
    db: AsyncSession,
    project_id: int,
    workspace_id: int,
    user_id: int
) -> List[State]:
    """创建默认状态"""
    states = []
    for default_state in DEFAULT_STATES:
        state_data = StateCreate(
            name=default_state['name'],
            color=default_state['color'],
            group=default_state['group'],
            sequence=default_state['sequence'],
            is_active=True,
            description=None,
            project_id=project_id
        )
        state = await create_state(db, state_data, workspace_id, user_id)
        states.append(state)
    return states


async def get_state_by_id(db: AsyncSession, state_id: int) -> State:
    """获取状态"""
    result = await db.execute(
        select(State).where(State.id == state_id)
    )
    state = result.scalar_one_or_none()
    if not state or state.is_deleted:
        raise NotFoundException("State not found")
    return state


async def get_project_states(
    db: AsyncSession,
    project_id: int,
    include_inactive: bool = False
) -> List[State]:
    """获取项目的状态列表"""
    query = select(State).where(
        State.project_id == project_id,
        State.is_deleted == False
    )
    
    if not include_inactive:
        query = query.where(State.is_active == True)
    
    query = query.order_by(State.sequence)
    
    result = await db.execute(query)
    return list(result.scalars().all())


async def update_state(
    db: AsyncSession,
    state_id: int,
    update_data: StateUpdate
) -> State:
    """更新状态"""
    state = await get_state_by_id(db, state_id)
    
    if update_data.name is not None:
        state.name = update_data.name
    if update_data.color is not None:
        state.color = update_data.color
    if update_data.group is not None:
        state.group = update_data.group.value
    if update_data.sequence is not None:
        state.sequence = update_data.sequence
    if update_data.is_active is not None:
        state.is_active = update_data.is_active
    if update_data.description is not None:
        state.description = update_data.description
    
    await db.commit()
    await db.refresh(state)
    return state


async def delete_state(db: AsyncSession, state_id: int):
    """删除状态（软删除）"""
    state = await get_state_by_id(db, state_id)
    state.is_deleted = True
    await db.commit()


# ==================== Label Service ====================

async def create_label(
    db: AsyncSession,
    label_data: LabelCreate,
    workspace_id: int,
    user_id: int
) -> Label:
    """创建标签"""
    # 验证项目存在
    project = await db.get(Project, label_data.project_id)
    if not project or project.is_deleted:
        raise NotFoundException("Project not found")
    
    label = Label(
        name=label_data.name,
        color=label_data.color,
        description=label_data.description,
        project_id=label_data.project_id,
        workspace_id=workspace_id,
        created_by_id=user_id
    )
    
    db.add(label)
    await db.commit()
    await db.refresh(label)
    return label


async def get_label_by_id(db: AsyncSession, label_id: int) -> Label:
    """获取标签"""
    result = await db.execute(
        select(Label).where(Label.id == label_id)
    )
    label = result.scalar_one_or_none()
    if not label or label.is_deleted:
        raise NotFoundException("Label not found")
    return label


async def get_project_labels(
    db: AsyncSession,
    project_id: int
) -> List[Label]:
    """获取项目的标签列表"""
    query = select(Label).where(
        Label.project_id == project_id,
        Label.is_deleted == False
    ).order_by(Label.name)
    
    result = await db.execute(query)
    return list(result.scalars().all())


async def update_label(
    db: AsyncSession,
    label_id: int,
    update_data: LabelUpdate
) -> Label:
    """更新标签"""
    label = await get_label_by_id(db, label_id)
    
    if update_data.name is not None:
        label.name = update_data.name
    if update_data.color is not None:
        label.color = update_data.color
    if update_data.description is not None:
        label.description = update_data.description
    
    await db.commit()
    await db.refresh(label)
    return label


async def delete_label(db: AsyncSession, label_id: int):
    """删除标签（软删除）"""
    label = await get_label_by_id(db, label_id)
    label.is_deleted = True
    await db.commit()