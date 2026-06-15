from typing import Optional
from uuid import UUID
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from app.models.workspace import Workspace, WorkspaceMember
from app.models.user import User
from app.schemas.workspace import WorkspaceCreate, WorkspaceUpdate
from app.core.exceptions import NotFoundException, ConflictException

async def create_workspace(db: AsyncSession, workspace_data: WorkspaceCreate, owner_id: UUID) -> Workspace:
    existing = await db.execute(select(Workspace).where(Workspace.slug == workspace_data.slug))
    if existing.scalar_one_or_none():
        raise ConflictException("Workspace slug already exists")
    
    workspace = Workspace(
        name=workspace_data.name,
        slug=workspace_data.slug,
        organization_size=workspace_data.organization_size,
        timezone=workspace_data.timezone,
        owner_id=owner_id,
        created_by_id=owner_id,
    )
    
    db.add(workspace)
    await db.commit()
    await db.refresh(workspace)
    
    member = WorkspaceMember(
        workspace_id=workspace.id,
        user_id=owner_id,
        role=20,
        created_by_id=owner_id,
    )
    db.add(member)
    await db.commit()
    
    return workspace

async def get_workspace_by_slug(db: AsyncSession, slug: str) -> Workspace:
    result = await db.execute(select(Workspace).where(Workspace.slug == slug, Workspace.is_deleted == False))
    workspace = result.scalar_one_or_none()
    
    if not workspace:
        raise NotFoundException("Workspace not found")
    
    return workspace

async def update_workspace(db: AsyncSession, workspace_id: UUID, update_data: WorkspaceUpdate) -> Workspace:
    workspace = await db.get(Workspace, workspace_id)
    if not workspace:
        raise NotFoundException("Workspace not found")
    
    if update_data.name is not None:
        workspace.name = update_data.name
    if update_data.logo_url is not None:
        workspace.logo_url = update_data.logo_url
    if update_data.timezone is not None:
        workspace.timezone = update_data.timezone
    
    await db.commit()
    await db.refresh(workspace)
    
    return workspace

async def delete_workspace(db: AsyncSession, workspace_id: UUID):
    workspace = await db.get(Workspace, workspace_id)
    if not workspace:
        raise NotFoundException("Workspace not found")
    
    workspace.is_deleted = True
    await db.commit()

async def list_workspaces(db: AsyncSession, user_id: UUID) -> list[Workspace]:
    result = await db.execute(
        select(Workspace)
        .join(WorkspaceMember, WorkspaceMember.workspace_id == Workspace.id)
        .where(WorkspaceMember.user_id == user_id, Workspace.is_deleted == False)
    )
    return result.scalars().all()