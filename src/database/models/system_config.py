from sqlalchemy import Column, String, Text, Boolean, Integer
from sqlalchemy.dialects.sqlite import JSON
from src.database.models.base_model import BaseModel

class SystemConfig(BaseModel):
    __tablename__ = 'system_config'

    key = Column(String(255), unique=True, nullable=False)
    value = Column(Text, nullable=True)
    default = Column(Text, nullable=True)
    description = Column(Text, nullable=True)

    group = Column(String(50), nullable=False, default="system")
    type = Column(String(20), nullable=False, default="string")
    options = Column(JSON, nullable=True)
    is_public = Column(Boolean, nullable=False, default=False)
    is_editable = Column(Boolean, nullable=False, default=True)
    order = Column(Integer, nullable=False, default=0)
