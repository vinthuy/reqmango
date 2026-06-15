from fastapi import APIRouter, Depends, Path
from sqlalchemy.ext.asyncio import AsyncSession
from uuid import UUID
from app.schemas.workspace import WorkspaceCreate, WorkspaceUpdate, WorkspaceResponse, WorkspaceLite
from app.services.workspace import create_workspace, get_workspace_by_slug, update_workspace, delete_workspace, list_workspaces
from app.api.deps import get_db, get_current_user
from app.models.user import User

router = APIRouter()

@router.get("/", response_model=list[WorkspaceLite])
async def list(
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    return await list_workspaces(db, current_user.id)

@router.post("/", response_model=WorkspaceResponse, status_code=201)
async def create(
    workspace_data: WorkspaceCreate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    workspace = await create_workspace(db, workspace_data, current_user.id)
    return workspace

@router.get("/{slug}", response_model=WorkspaceResponse)
async def get(
    slug: str = Path(..., description="Workspace slug"),
    db: AsyncSession = Depends(get_db)
):
    return await get_workspace_by_slug(db, slug)

@router.patch("/{workspace_id}", response_model=WorkspaceResponse)
async def update(
    workspace_id: UUID = Path(..., description="Workspace ID"),
    update_data: WorkspaceUpdate = None,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    return await update_workspace(db, workspace_id, update_data)

@router.delete("/{workspace_id}", status_code=204)
async def delete(
    workspace_id: UUID = Path(..., description="Workspace ID"),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    await delete_workspace(db, workspace_id)