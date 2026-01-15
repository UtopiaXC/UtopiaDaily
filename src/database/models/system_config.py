from sqlalchemy import Column, String, Text, Boolean
from sqlalchemy.dialects.sqlite import JSON
from src.database.models.base_model import BaseModel

class SystemConfig(BaseModel):
    __tablename__ = 'system_config'

    key = Column(String(255), unique=True, nullable=False)
    value = Column(Text, nullable=True)
    default = Column(Text, nullable=True)
    description = Column(Text, nullable=True)
    
    # 新增字段以支持前端动态表单渲染
    group = Column(String(50), nullable=False, default="system") # e.g., "system", "frontend", "push_module"
    type = Column(String(20), nullable=False, default="string") # e.g., "string", "boolean", "int", "select"
    options = Column(JSON, nullable=True) # e.g., ["en", "zh_CN"] for select type
    is_public = Column(Boolean, nullable=False, default=False) # If true, accessible by non-admin (read-only)
    is_editable = Column(Boolean, nullable=False, default=True) # If false, UI shows as disabled/read-only
