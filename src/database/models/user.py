from sqlalchemy import Column, String, Boolean, ForeignKey, Text, BigInteger
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.sqlite import JSON
from src.database.models.base_model import BaseModel

class UserRole(BaseModel):
    """
    角色表 (原 Type 表)
    用于定义用户组和权限范围
    """
    __tablename__ = 'user_role'

    name = Column(String(50), unique=True, nullable=False) # e.g., "admin", "editor"
    description = Column(String(255), nullable=True)
    
    # 核心权限字段，使用 JSON 存储以支持灵活配置
    # 示例: {"is_admin": false, "allowed_modules": ["telegram"], "push_channels": ["email"]}
    permissions = Column(JSON, nullable=False, default={}) 

    users = relationship("User", back_populates="role")

class User(BaseModel):
    """
    用户表
    """
    __tablename__ = 'user'

    username = Column(String(50), unique=True, nullable=False, index=True)
    password_hash = Column(String(255), nullable=False) # 存储哈希后的密码，绝不存明文
    email = Column(String(255), nullable=True)
    is_active = Column(Boolean, default=True, nullable=False)
    
    role_id = Column(String(36), ForeignKey('user_role.id'), nullable=False)
    role = relationship("UserRole", back_populates="users")
    
    sessions = relationship("UserSession", back_populates="user", cascade="all, delete-orphan")
    logs = relationship("UserLog", back_populates="user")
    
    # 新增关联
    push_config = relationship("UserPushConfig", back_populates="user", uselist=False, cascade="all, delete-orphan")

class UserSession(BaseModel):
    """
    用户会话表
    用于保持登录状态 (RefreshToken)
    """
    __tablename__ = 'user_session'

    user_id = Column(String(36), ForeignKey('user.id'), nullable=False)
    token_hash = Column(String(255), unique=True, nullable=False, index=True) # 存储 Token 的哈希
    expires_at = Column(BigInteger, nullable=False) # 过期时间戳 (ms)
    
    ip_address = Column(String(45), nullable=True)
    user_agent = Column(Text, nullable=True)
    is_valid = Column(Boolean, default=True) # 可用于强制下线

    user = relationship("User", back_populates="sessions")

class UserLog(BaseModel):
    """
    用户操作审计日志
    """
    __tablename__ = 'user_log'

    user_id = Column(String(36), ForeignKey('user.id'), nullable=True) # 允许为空(如登录失败)
    action = Column(String(50), nullable=False) # e.g., "LOGIN", "UPDATE_CONFIG"
    status = Column(String(20), nullable=False) # "SUCCESS", "FAILURE"
    ip_address = Column(String(45), nullable=True)
    details = Column(JSON, nullable=True) # 存储详细信息，如修改了哪个配置

    user = relationship("User", back_populates="logs")
