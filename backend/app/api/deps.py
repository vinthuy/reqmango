from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from sqlalchemy.ext.asyncio import AsyncSession
from typing import Optional
from uuid import UUID
from app.db.session import get_db
from app.core.security import decode_access_token
from app.models.user import User
from app.core.exceptions import UnauthorizedException, ForbiddenException

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/api/v1/auth/login")

async def get_current_user(
    token: str = Depends(oauth2_scheme),
    db: AsyncSession = Depends(get_db)
) -> User:
    payload = decode_access_token(token)
    if not payload:
        raise UnauthorizedException()
    
    user_id = payload.get("sub")
    if not user_id:
        raise UnauthorizedException()
    
    user = await db.get(User, UUID(user_id))
    if not user or user.is_deleted:
        raise UnauthorizedException()
    
    return user

async def require_workspace_access(
    workspace_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
) -> User:
    from app.models.workspace import WorkspaceMember
    
    member = await db.get(WorkspaceMember, {"workspace_id": workspace_id, "user_id": current_user.id})
    if not member or not member.is_active:
        raise ForbiddenException("Not a member of this workspace")
    
    return current_user