from pydantic import BaseModel, EmailStr, Field
from typing import Optional, List
from datetime import datetime
from uuid import UUID
from .base import AuditSchema, SoftDeleteSchema

class UserBase(BaseModel):
    email: EmailStr
    username: str = Field(..., min_length=3, max_length=128)
    display_name: Optional[str] = None
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    user_timezone: str = "UTC"

class UserCreate(UserBase):
    password: str = Field(..., min_length=8)

class UserUpdate(BaseModel):
    display_name: Optional[str] = None
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    avatar: Optional[str] = None
    user_timezone: Optional[str] = None

class UserResponse(AuditSchema, SoftDeleteSchema, UserBase):
    is_active: bool
    is_email_verified: bool
    avatar_url: Optional[str] = None
    cover_image_url: Optional[str] = None
    last_active: Optional[datetime] = None

class UserLite(BaseModel):
    id: UUID
    display_name: str
    email: EmailStr
    avatar_url: Optional[str] = None

class Token(BaseModel):
    access_token: str
    token_type: str = "bearer"
    expires_at: datetime

class LoginRequest(BaseModel):
    email: EmailStr
    password: str