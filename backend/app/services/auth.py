from datetime import timedelta
from typing import Optional
from uuid import UUID
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from app.models.user import User
from app.schemas.user import UserCreate, LoginRequest, Token
from app.core.security import verify_password, get_password_hash, create_access_token
from app.core.config import settings
from app.core.exceptions import UnauthorizedException, ConflictException

async def register_user(db: AsyncSession, user_data: UserCreate) -> User:
    existing = await db.execute(select(User).where(User.email == user_data.email))
    if existing.scalar_one_or_none():
        raise ConflictException("Email already registered")
    
    existing_username = await db.execute(select(User).where(User.username == user_data.username))
    if existing_username.scalar_one_or_none():
        raise ConflictException("Username already taken")
    
    user = User(
        email=user_data.email,
        username=user_data.username,
        display_name=user_data.display_name or user_data.username,
        password_hash=get_password_hash(user_data.password),
        first_name=user_data.first_name,
        last_name=user_data.last_name,
        user_timezone=user_data.user_timezone,
    )
    
    db.add(user)
    await db.commit()
    await db.refresh(user)
    
    return user

async def login(db: AsyncSession, login_data: LoginRequest) -> Token:
    user = await db.execute(select(User).where(User.email == login_data.email))
    user = user.scalar_one_or_none()
    
    if not user or not verify_password(login_data.password, user.password_hash):
        raise UnauthorizedException("Invalid email or password")
    
    if not user.is_active:
        raise UnauthorizedException("Account is disabled")
    
    access_token, expires_at = create_access_token(
        data={"sub": str(user.id)},
        expires_delta=timedelta(minutes=settings.ACCESS_TOKEN_EXPIRE_MINUTES)
    )
    
    return Token(
        access_token=access_token,
        token_type="bearer",
        expires_at=expires_at
    )

async def get_user_by_id(db: AsyncSession, user_id: UUID) -> Optional[User]:
    return await db.get(User, user_id)