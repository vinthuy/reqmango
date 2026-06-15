from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from pydantic import BaseModel
from app.schemas.user import UserCreate, LoginRequest, Token, UserResponse
from app.services.auth import register_user, login
from app.api.deps import get_db, get_current_user
from app.models.user import User
from app.core.security import verify_password, get_password_hash

router = APIRouter()

class UpdatePasswordRequest(BaseModel):
    current_password: str
    new_password: str

@router.post("/register", response_model=UserResponse, status_code=201)
async def register(
    user_data: UserCreate,
    db: AsyncSession = Depends(get_db)
):
    user = await register_user(db, user_data)
    return user

@router.post("/login", response_model=Token)
async def login_user(
    login_data: LoginRequest,
    db: AsyncSession = Depends(get_db)
):
    return await login(db, login_data)

@router.get("/me", response_model=UserResponse)
async def get_current_user(
    user: User = Depends(get_current_user)
):
    return user

@router.post("/update-password")
async def update_password(
    password_data: UpdatePasswordRequest,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    # 验证当前密码
    if not verify_password(password_data.current_password, current_user.password_hash):
        raise HTTPException(status_code=400, detail="Current password is incorrect")
    
    # 更新密码
    current_user.password_hash = get_password_hash(password_data.new_password)
    await db.commit()
    
    return {"message": "Password updated successfully"}