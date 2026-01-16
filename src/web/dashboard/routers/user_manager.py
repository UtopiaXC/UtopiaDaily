from fastapi import APIRouter, Depends, HTTPException, Query, Request
from sqlalchemy.orm import Session
from typing import List, Optional, Dict, Union, Any
from pydantic import BaseModel, field_validator
import secrets
import string
import hashlib
import re
from src.database.connection import system_db_manager
from src.database.models import User, UserRole, UserSession
from src.utils.constants.permissions import Permissions
from src.utils.security import crypto_manager
from src.utils.i18n import i18n
from src.web.dashboard.routers.auth import get_locale

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

def get_db():
    db = system_db_manager.get_session()
    try:
        yield db
    finally:
        db.close()

def generate_secure_password(length=12):
    alphabet = string.ascii_letters + string.digits + "!@#$%^&*"
    while True:
        password = ''.join(secrets.choice(alphabet) for i in range(length))
        if (any(c.islower() for c in password)
                and any(c.isupper() for c in password)
                and any(c.isdigit() for c in password)
                and any(c in "!@#$%^&*" for c in password)):
            return password

def check_last_admin(db: Session, target_user_id: str):
    """
    Check if the target user is the last active admin.
    Returns True if they are the last admin, False otherwise.
    """
    admin_role = db.query(UserRole).filter(UserRole.name == "admin").first()
    if not admin_role:
        return False

    target_user = db.query(User).filter(User.id == target_user_id).first()
    if not target_user or target_user.role_id != admin_role.id:
        return False

    other_admins_count = db.query(User).filter(
        User.role_id == admin_role.id,
        User.is_active == True,
        User.id != target_user_id
    ).count()

    return other_admins_count == 0

def validate_username(username: str):
    if len(username) < 3:
        return False
    # Disallow '@' to prevent confusion with email login
    if not re.match(r'^[a-zA-Z0-9_\-\.]+$', username):
        return False
    return True

@router.get("/permissions")
async def get_all_permissions():
    """Return all available permission nodes in the system."""
    return Permissions.get_all()

@router.get("/roles", response_model=List[RoleResponse])
async def get_roles(db: Session = Depends(get_db)):
    roles = db.query(UserRole).all()
    result = []
    for role in roles:
        count = db.query(User).filter(User.role_id == role.id).count()
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
        result.append(role_dict)
    return result

@router.post("/roles", response_model=RoleResponse)
async def create_role(role: RoleCreate, db: Session = Depends(get_db)):
    if db.query(UserRole).filter(UserRole.name == role.name).first():
        raise HTTPException(status_code=400, detail="Role name already exists")
    
    new_role = UserRole(
        name=role.name,
        description=role.description,
        permissions=role.permissions
    )
    db.add(new_role)
    db.commit()
    db.refresh(new_role)
    return RoleResponse(id=new_role.id, name=new_role.name, description=new_role.description, permissions=new_role.permissions, user_count=0)

@router.put("/roles/{role_id}", response_model=RoleResponse)
async def update_role(role_id: str, update: RoleUpdate, req: Request, db: Session = Depends(get_db)):
    locale = get_locale(req)
    role = db.query(UserRole).filter(UserRole.id == role_id).first()
    if not role:
        raise HTTPException(status_code=404, detail="Role not found")
    
    # Admin Role Protection
    if role.name == "admin":
        if update.name != "admin":
            detail = i18n.t("user_manager.admin_readonly", locale=locale)
            raise HTTPException(status_code=400, detail=detail)
        
        # Ensure permissions are not changed (or reset to full)
        all_perms = Permissions.get_all()
        update.permissions = {p: True for p in all_perms}
    else:
        if update.name != role.name:
            if db.query(UserRole).filter(UserRole.name == update.name).first():
                raise HTTPException(status_code=400, detail="Role name already exists")

    role.name = update.name
    role.description = update.description
    role.permissions = update.permissions
    
    db.commit()
    db.refresh(role)
    
    count = db.query(User).filter(User.role_id == role.id).count()
    return RoleResponse(id=role.id, name=role.name, description=role.description, permissions=role.permissions, user_count=count)

@router.delete("/roles/{role_id}")
async def delete_role(role_id: str, req: Request, db: Session = Depends(get_db)):
    locale = get_locale(req)
    role = db.query(UserRole).filter(UserRole.id == role_id).first()
    if not role:
        raise HTTPException(status_code=404, detail="Role not found")
    
    user_count = db.query(User).filter(User.role_id == role_id).count()
    if user_count > 0:
        detail = i18n.t("user_manager.error_role_has_users", locale=locale)
        raise HTTPException(status_code=400, detail=detail)
        
    if role.name == "admin":
         detail = i18n.t("user_manager.error_delete_admin_role", locale=locale)
         raise HTTPException(status_code=400, detail=detail)

    db.delete(role)
    db.commit()
    return {"status": "success"}

# --- User Endpoints ---

@router.get("/users", response_model=List[UserResponse])
async def get_users(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    users = db.query(User).offset(skip).limit(limit).all()
    result = []
    for user in users:
        role_name = user.role.name if user.role else "Unknown"
        result.append({
            "id": user.id,
            "username": user.username,
            "nickname": user.nickname,
            "email": user.email,
            "is_active": user.is_active,
            "role_id": user.role_id,
            "role_name": role_name
        })
    return result

@router.post("/users", response_model=UserResponse)
async def create_user(user: UserCreate, req: Request, db: Session = Depends(get_db)):
    locale = get_locale(req)
    
    if not validate_username(user.username):
        detail = i18n.t("login.username_invalid", locale=locale)
        raise HTTPException(status_code=400, detail=detail)

    if db.query(User).filter(User.username == user.username).first():
        detail = i18n.t("login.username_exists", locale=locale)
        raise HTTPException(status_code=400, detail=detail)
    
    if user.email:
        if db.query(User).filter(User.email == user.email).first():
            detail = i18n.t("login.email_exists", locale=locale)
            raise HTTPException(status_code=400, detail=detail)
    

    if len(user.password) != 32:
         pass 

    role = db.query(UserRole).filter(UserRole.id == user.role_id).first()
    if not role:
        raise HTTPException(status_code=400, detail="Invalid Role ID")

    password_hash = crypto_manager.get_password_hash(user.password)
    
    new_user = User(
        username=user.username,
        nickname=user.nickname,
        email=user.email,
        password_hash=password_hash,
        role_id=user.role_id,
        is_active=user.is_active
    )
    db.add(new_user)
    db.commit()
    db.refresh(new_user)
    
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
async def update_user(user_id: str, update: UserUpdate, req: Request, db: Session = Depends(get_db)):
    locale = get_locale(req)
    user = db.query(User).filter(User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    
    # Last Admin Protection
    is_changing_role = update.role_id is not None and update.role_id != user.role_id
    is_deactivating = update.is_active is False # Explicitly False
    
    if is_changing_role or is_deactivating:
        if check_last_admin(db, user_id):
            detail = i18n.t("user_manager.error_last_admin_deactivate", locale=locale)
            raise HTTPException(status_code=400, detail=detail)

    if update.username is not None and update.username != user.username:
        if not validate_username(update.username):
            detail = i18n.t("login.username_invalid", locale=locale)
            raise HTTPException(status_code=400, detail=detail)
        if db.query(User).filter(User.username == update.username).first():
            detail = i18n.t("login.username_exists", locale=locale)
            raise HTTPException(status_code=400, detail=detail)
        user.username = update.username

    if update.nickname is not None:
        user.nickname = update.nickname
        
    if update.email is not None:
        if update.email != user.email:
            if db.query(User).filter(User.email == update.email).first():
                detail = i18n.t("login.email_exists", locale=locale)
                raise HTTPException(status_code=400, detail=detail)
        user.email = update.email
        
    if update.is_active is not None:
        user.is_active = update.is_active
    if update.role_id is not None:
        role = db.query(UserRole).filter(UserRole.id == update.role_id).first()
        if not role:
            raise HTTPException(status_code=400, detail="Invalid Role ID")
        user.role_id = update.role_id
    
    if update.password:
        user.password_hash = crypto_manager.get_password_hash(update.password)
        
    db.commit()
    db.refresh(user)
    
    return {
        "id": user.id,
        "username": user.username,
        "nickname": user.nickname,
        "email": user.email,
        "is_active": user.is_active,
        "role_id": user.role_id,
        "role_name": user.role.name
    }

@router.delete("/users/{user_id}")
async def delete_user(user_id: str, req: Request, db: Session = Depends(get_db)):
    locale = get_locale(req)
    user = db.query(User).filter(User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    
    # Prevent deleting self
    current_user = getattr(req.state, "user", None)
    if current_user and current_user.id == user_id:
        detail = i18n.t("user_manager.error_delete_self", locale=locale)
        raise HTTPException(status_code=400, detail=detail)

    # Last Admin Protection
    if check_last_admin(db, user_id):
        detail = i18n.t("user_manager.error_last_admin_delete", locale=locale)
        raise HTTPException(status_code=400, detail=detail)

    db.delete(user)
    db.commit()
    return {"status": "success"}

@router.post("/users/{user_id}/reset-password", response_model=PasswordResetResponse)
async def reset_password(user_id: str, req: Request, db: Session = Depends(get_db)):
    user = db.query(User).filter(User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    
    current_user = getattr(req.state, "user", None)
    if current_user and current_user.id == user_id:
        raise HTTPException(status_code=400, detail="Cannot force reset your own password. Please use the update profile function.")

    new_password = generate_secure_password()
    
    # Simulate frontend MD5 hashing before storage
    md5_password = hashlib.md5(new_password.encode()).hexdigest()
    user.password_hash = crypto_manager.get_password_hash(md5_password)
    
    # Invalidate all sessions for this user
    db.query(UserSession).filter(UserSession.user_id == user_id).delete()

    db.commit()
    
    return {"new_password": new_password}
