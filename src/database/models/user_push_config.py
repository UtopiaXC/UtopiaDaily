from sqlalchemy import Column, String, ForeignKey
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.sqlite import JSON
from src.database.models.base_model import BaseModel

class UserPushConfig(BaseModel):
    """
    用户个人推送配置表
    允许普通用户配置自己的推送规则
    """
    __tablename__ = 'user_push_config'

    user_id = Column(String(36), ForeignKey('user.id'), nullable=False, unique=True)
    config = Column(JSON, nullable=False, default={})

    user = relationship("User", back_populates="push_config")
