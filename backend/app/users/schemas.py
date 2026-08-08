from datetime import datetime

from pydantic import BaseModel, ConfigDict, EmailStr

from app.users.models import UserRole


class UserCreate(BaseModel):
    email: EmailStr
    password: str
    full_name: str


class UserRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: str
    email: EmailStr
    full_name: str
    role: UserRole
    is_active: bool
    is_business_verified: bool
    business_category: str | None
    quality_flagged_at: datetime | None
    created_at: datetime


class UserUpdate(BaseModel):
    full_name: str | None = None
    business_category: str | None = None
