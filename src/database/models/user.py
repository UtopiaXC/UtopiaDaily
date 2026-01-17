from sqlalchemy import Column, String, Boolean, ForeignKey, Text, BigInteger
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.sqlite import JSON
from src.database.models.base_model import BaseModel

class UserRole(BaseModel):
    __tablename__ = 'user_role'

    name = Column(String(50), unique=True, nullable=False) # e.g., "admin", "editor"
    description = Column(String(255), nullable=True)
    permissions = Column(JSON, nullable=False, default={}) 

    users = relationship("User", back_populates="role")

class User(BaseModel):
    __tablename__ = 'user'

    username = Column(String(50), unique=True, nullable=False, index=True)
    nickname = Column(String(100), nullable=True)
    password_hash = Column(String(255), nullable=False)
    email = Column(String(255), nullable=True)
    is_active = Column(Boolean, default=True, nullable=False)
    role_id = Column(String(36), ForeignKey('user_role.id'), nullable=False)
    role = relationship("UserRole", back_populates="users")
    sessions = relationship("UserSession", back_populates="user", cascade="all, delete-orphan")
    # logs = relationship("UserLog", back_populates="user") # Deprecated
    push_config = relationship("UserPushConfig", back_populates="user", uselist=False, cascade="all, delete-orphan")

class UserSession(BaseModel):
    __tablename__ = 'user_session'

    user_id = Column(String(36), ForeignKey('user.id'), nullable=False)
    token_hash = Column(String(255), unique=True, nullable=False, index=True)
    expires_at = Column(BigInteger, nullable=False)
    ip_address = Column(String(45), nullable=True)
    user_agent = Column(Text, nullable=True)
    is_valid = Column(Boolean, default=True)
    user = relationship("User", back_populates="sessions")

# UserLog class removed. Use SystemEvent instead.
