from fastapi import APIRouter, Depends, HTTPException, Request
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func, delete
from sqlalchemy.orm import selectinload
from starlette.concurrency import run_in_threadpool
from typing import List, Optional, Dict
from pydantic import BaseModel, field_validator
import secrets
import string
import hashlib
import re
from src.database.models import User, UserRole, UserSession
from src.utils.constants.permissions import Permissions
from src.utils.security import crypto_manager
from src.utils.i18n import i18n
from src.web.dashboard.routers.auth import get_locale
from src.utils.event import EventManager
from src.web.dependencies import get_async_db

router = APIRouter(prefix="/api/dashboard/user-manager", tags=["User Manager"])
class RoleBase(BaseModel):
    name: str
    description: Optional[str] = None
    permissions: Dict[str, bool] = {} 

    @field_validator('permissions', mode='before')
    @classmethod
    def parse_permissions(cls, v):
        if isinstance(v, list):
            return {perm: True for perm in v}
        return v or {}

class RoleCreate(RoleBase):
    pass

class RoleUpdate(RoleBase):
    pass

class RoleResponse(RoleBase):
    id: str
    user_count: int = 0

    class Config:
        from_attributes = True

class UserBase(BaseModel):
    username: str
    nickname: Optional[str] = None
    email: Optional[str] = None
    is_active: bool = True
    role_id: str

class UserCreate(UserBase):
    password: str

class UserUpdate(BaseModel):
    username: Optional[str] = None
    nickname: Optional[str] = None
    email: Optional[str] = None
    is_active: Optional[bool] = None
    role_id: Optional[str] = None
    password: Optional[str] = None

class UserResponse(UserBase):
    id: str
    role_name: str

    class Config:
        from_attributes = True

class PasswordResetResponse(BaseModel):
    new_password: str

def generate_secure_password(length=12):
    alphabet = string.ascii_letters + string.digits + "!@#$%^&*"
    while True:
        password = ''.join(secrets.choice(alphabet) for i in range(length))
        if (any(c.islower() for c in password)
                and any(c.isupper() for c in password)
                and any(c.isdigit() for c in password)
                and any(c in "!@#$%^&*" for c in password)):
            return password

async def check_last_admin(db: AsyncSession, target_user_id: str):
    result = await db.execute(select(UserRole).where(UserRole.name == "admin"))
    admin_role = result.scalars().first()
    if not admin_role:
        return False

    user_result = await db.execute(select(User).where(User.id == target_user_id))
    target_user = user_result.scalars().first()
    
    if not target_user or target_user.role_id != admin_role.id:
        return False

    count_result = await db.execute(
        select(func.count(User.id)).where(
            User.role_id == admin_role.id,
            User.is_active == True,
            User.id != target_user_id
        )
    )
    other_admins_count = count_result.scalar()

    return other_admins_count == 0

def validate_username(username: str):
    if len(username) < 3:
        return False
    # Disallow '@' to prevent confusion with email login
    if not re.match(r'^[a-zA-Z0-9_\-\.]+$', username):
        return False
    return True

@router.get("/permissions")
def get_all_permissions():
    return Permissions.get_all()

@router.get("/roles", response_model=List[RoleResponse])
async def get_roles(db: AsyncSession = Depends(get_async_db)):
    result = await db.execute(select(UserRole))
    roles = result.scalars().all()
    
    role_responses = []
    for role in roles:
        count_result = await db.execute(select(func.count(User.id)).where(User.role_id == role.id))
        count = count_result.scalar()
        
        perms = role.permissions
        if isinstance(perms, list):
            perms = {p: True for p in perms}
        role_dict = {
            "id": role.id,
            "name": role.name,
            "description": role.description,
            "permissions": perms,
            "user_count": count
        }
        role_responses.append(role_dict)
    return role_responses

@router.post("/roles", response_model=RoleResponse)
async def create_role(role: RoleCreate, req: Request, db: AsyncSession = Depends(get_async_db)):
    result = await db.execute(select(UserRole).where(UserRole.name == role.name))
    if result.scalars().first():
        raise HTTPException(status_code=400, detail="Role name already exists")
    
    new_role = UserRole(
        name=role.name,
        description=role.description,
        permissions=role.permissions
    )
    db.add(new_role)
    await db.commit()
    await db.refresh(new_role)
    
    current_user = getattr(req.state, "user", None)
    await EventManager.record_async(
        level=EventManager.LEVEL_NORMAL,
        category=EventManager.CATEGORY_USER,
        event_type="role_created",
        summary=f"Role created: {new_role.name}",
        details={"role_id": new_role.id, "created_by": current_user.username if current_user else "unknown"},
        source_id=current_user.id if current_user else None
    )
    
    return RoleResponse(id=new_role.id, name=new_role.name, description=new_role.description, permissions=new_role.permissions, user_count=0)

@router.put("/roles/{role_id}", response_model=RoleResponse)
async def update_role(role_id: str, update: RoleUpdate, req: Request, db: AsyncSession = Depends(get_async_db)):
    locale = get_locale(req)
    result = await db.execute(select(UserRole).where(UserRole.id == role_id))
    role = result.scalars().first()
    
    if not role:
        raise HTTPException(status_code=404, detail="Role not found")
    if role.name == "admin":
        if update.name != "admin":
            detail = i18n.t("user_manager.admin_readonly", locale=locale)
            raise HTTPException(status_code=400, detail=detail)

        all_perms = Permissions.get_all()
        update.permissions = {p: True for p in all_perms}
    else:
        if update.name != role.name:
            name_check = await db.execute(select(UserRole).where(UserRole.name == update.name))
            if name_check.scalars().first():
                raise HTTPException(status_code=400, detail="Role name already exists")

    role.name = update.name
    role.description = update.description
    role.permissions = update.permissions
    
    await db.commit()
    await db.refresh(role)
    
    current_user = getattr(req.state, "user", None)
    await EventManager.record_async(
        level=EventManager.LEVEL_NORMAL,
        category=EventManager.CATEGORY_USER,
        event_type="role_updated",
        summary=f"Role updated: {role.name}",
        details={"role_id": role.id, "updated_by": current_user.username if current_user else "unknown"},
        source_id=current_user.id if current_user else None
    )
    
    count_result = await db.execute(select(func.count(User.id)).where(User.role_id == role.id))
    count = count_result.scalar()
    
    return RoleResponse(id=role.id, name=role.name, description=role.description, permissions=role.permissions, user_count=count)

@router.delete("/roles/{role_id}")
async def delete_role(role_id: str, req: Request, db: AsyncSession = Depends(get_async_db)):
    locale = get_locale(req)
    result = await db.execute(select(UserRole).where(UserRole.id == role_id))
    role = result.scalars().first()
    
    if not role:
        raise HTTPException(status_code=404, detail="Role not found")
    
    count_result = await db.execute(select(func.count(User.id)).where(User.role_id == role_id))
    user_count = count_result.scalar()
    
    if user_count > 0:
        detail = i18n.t("user_manager.error_role_has_users", locale=locale)
        raise HTTPException(status_code=400, detail=detail)
        
    if role.name == "admin":
         detail = i18n.t("user_manager.error_delete_admin_role", locale=locale)
         raise HTTPException(status_code=400, detail=detail)

    role_name = role.name
    await db.delete(role)
    await db.commit()
    
    current_user = getattr(req.state, "user", None)
    await EventManager.record_async(
        level=EventManager.LEVEL_WARNING,
        category=EventManager.CATEGORY_USER,
        event_type="role_deleted",
        summary=f"Role deleted: {role_name}",
        details={"role_id": role_id, "deleted_by": current_user.username if current_user else "unknown"},
        source_id=current_user.id if current_user else None
    )
    
    return {"status": "success"}

@router.get("/users", response_model=List[UserResponse])
async def get_users(skip: int = 0, limit: int = 100, db: AsyncSession = Depends(get_async_db)):
    # Use selectinload to eagerly load the role relationship asynchronously
    stmt = select(User).options(selectinload(User.role)).offset(skip).limit(limit)
    result = await db.execute(stmt)
    users = result.scalars().all()
    
    user_responses = []
    for user in users:
        role_name = user.role.name if user.role else "Unknown"
        user_responses.append({
            "id": user.id,
            "username": user.username,
            "nickname": user.nickname,
            "email": user.email,
            "is_active": user.is_active,
            "role_id": user.role_id,
            "role_name": role_name
        })
    return user_responses

@router.post("/users", response_model=UserResponse)
async def create_user(user: UserCreate, req: Request, db: AsyncSession = Depends(get_async_db)):
    locale = get_locale(req)
    
    if not validate_username(user.username):
        detail = i18n.t("login.username_invalid", locale=locale)
        raise HTTPException(status_code=400, detail=detail)

    username_check = await db.execute(select(User).where(User.username == user.username))
    if username_check.scalars().first():
        detail = i18n.t("login.username_exists", locale=locale)
        raise HTTPException(status_code=400, detail=detail)
    
    if user.email:
        email_check = await db.execute(select(User).where(User.email == user.email))
        if email_check.scalars().first():
            detail = i18n.t("login.email_exists", locale=locale)
            raise HTTPException(status_code=400, detail=detail)
    
    role_check = await db.execute(select(UserRole).where(UserRole.id == user.role_id))
    role = role_check.scalars().first()
    if not role:
        raise HTTPException(status_code=400, detail="Invalid Role ID")

    # Password hashing is CPU intensive, run in threadpool
    password_hash = await run_in_threadpool(crypto_manager.get_password_hash, user.password)
    
    new_user = User(
        username=user.username,
        nickname=user.nickname,
        email=user.email,
        password_hash=password_hash,
        role_id=user.role_id,
        is_active=user.is_active
    )
    db.add(new_user)
    await db.commit()
    await db.refresh(new_user)
    
    current_user = getattr(req.state, "user", None)
    await EventManager.record_async(
        level=EventManager.LEVEL_NORMAL,
        category=EventManager.CATEGORY_USER,
        event_type="user_created",
        summary=f"User created: {new_user.username}",
        details={"user_id": new_user.id, "role": role.name, "created_by": current_user.username if current_user else "unknown"},
        source_id=current_user.id if current_user else None
    )
    
    return {
        "id": new_user.id,
        "username": new_user.username,
        "nickname": new_user.nickname,
        "email": new_user.email,
        "is_active": new_user.is_active,
        "role_id": new_user.role_id,
        "role_name": role.name
    }

@router.put("/users/{user_id}", response_model=UserResponse)
async def update_user(user_id: str, update: UserUpdate, req: Request, db: AsyncSession = Depends(get_async_db)):
    locale = get_locale(req)
    # Eager load role to return correct response later
    stmt = select(User).options(selectinload(User.role)).where(User.id == user_id)
    result = await db.execute(stmt)
    user = result.scalars().first()
    
    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    is_changing_role = update.role_id is not None and update.role_id != user.role_id
    is_deactivating = update.is_active is False # Explicitly False
    
    if is_changing_role or is_deactivating:
        if await check_last_admin(db, user_id):
            detail = i18n.t("user_manager.error_last_admin_deactivate", locale=locale)
            raise HTTPException(status_code=400, detail=detail)

    if update.username is not None and update.username != user.username:
        if not validate_username(update.username):
            detail = i18n.t("login.username_invalid", locale=locale)
            raise HTTPException(status_code=400, detail=detail)
        
        username_check = await db.execute(select(User).where(User.username == update.username))
        if username_check.scalars().first():
            detail = i18n.t("login.username_exists", locale=locale)
            raise HTTPException(status_code=400, detail=detail)
        user.username = update.username

    if update.nickname is not None:
        user.nickname = update.nickname
        
    if update.email is not None:
        if update.email != user.email:
            email_check = await db.execute(select(User).where(User.email == update.email))
            if email_check.scalars().first():
                detail = i18n.t("login.email_exists", locale=locale)
                raise HTTPException(status_code=400, detail=detail)
        user.email = update.email
        
    if update.is_active is not None:
        user.is_active = update.is_active
    if update.role_id is not None:
        role_check = await db.execute(select(UserRole).where(UserRole.id == update.role_id))
        role = role_check.scalars().first()
        if not role:
            raise HTTPException(status_code=400, detail="Invalid Role ID")
        user.role_id = update.role_id
    
    if update.password:
        user.password_hash = await run_in_threadpool(crypto_manager.get_password_hash, update.password)
        
    await db.commit()
    stmt = select(User).options(selectinload(User.role)).where(User.id == user_id)
    result = await db.execute(stmt)
    user = result.scalars().first()

    current_user = getattr(req.state, "user", None)
    await EventManager.record_async(
        level=EventManager.LEVEL_NORMAL,
        category=EventManager.CATEGORY_USER,
        event_type="user_updated",
        summary=f"User updated: {user.username}",
        details={"user_id": user.id, "updated_by": current_user.username if current_user else "unknown"},
        source_id=current_user.id if current_user else None
    )
    
    return {
        "id": user.id,
        "username": user.username,
        "nickname": user.nickname,
        "email": user.email,
        "is_active": user.is_active,
        "role_id": user.role_id,
        "role_name": user.role.name if user.role else "Unknown"
    }

@router.delete("/users/{user_id}")
async def delete_user(user_id: str, req: Request, db: AsyncSession = Depends(get_async_db)):
    locale = get_locale(req)
    result = await db.execute(select(User).where(User.id == user_id))
    user = result.scalars().first()
    
    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    current_user = getattr(req.state, "user", None)
    if current_user and current_user.id == user_id:
        detail = i18n.t("user_manager.error_delete_self", locale=locale)
        raise HTTPException(status_code=400, detail=detail)

    if await check_last_admin(db, user_id):
        detail = i18n.t("user_manager.error_last_admin_delete", locale=locale)
        raise HTTPException(status_code=400, detail=detail)

    username = user.username
    await db.delete(user)
    await db.commit()
    
    await EventManager.record_async(
        level=EventManager.LEVEL_WARNING,
        category=EventManager.CATEGORY_USER,
        event_type="user_deleted",
        summary=f"User deleted: {username}",
        details={"user_id": user_id, "deleted_by": current_user.username if current_user else "unknown"},
        source_id=current_user.id if current_user else None
    )
    
    return {"status": "success"}

@router.post("/users/{user_id}/reset-password", response_model=PasswordResetResponse)
async def reset_password(user_id: str, req: Request, db: AsyncSession = Depends(get_async_db)):
    result = await db.execute(select(User).where(User.id == user_id))
    user = result.scalars().first()
    
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    
    current_user = getattr(req.state, "user", None)
    if current_user and current_user.id == user_id:
        raise HTTPException(status_code=400, detail="Cannot force reset your own password. Please use the update profile function.")

    new_password = generate_secure_password()
    md5_password = hashlib.md5(new_password.encode()).hexdigest()

    user.password_hash = await run_in_threadpool(crypto_manager.get_password_hash, md5_password)
    
    await db.execute(delete(UserSession).where(UserSession.user_id == user_id))

    await db.commit()
    
    await EventManager.record_async(
        level=EventManager.LEVEL_WARNING,
        category=EventManager.CATEGORY_USER,
        event_type="password_reset",
        summary=f"Password reset for user: {user.username}",
        details={"user_id": user_id, "reset_by": current_user.username if current_user else "unknown"},
        source_id=current_user.id if current_user else None
    )
    
    return {"new_password": new_password}
