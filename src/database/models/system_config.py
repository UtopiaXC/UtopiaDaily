from sqlalchemy import Column, String, Text
from src.database.models.base_model import BaseModel

class SystemConfig(BaseModel):
    __tablename__ = 'system_config'

    key = Column(String(255), unique=True, nullable=False)
    value = Column(Text, nullable=True)
    default = Column(Text, nullable=True)
    description = Column(Text, nullable=True)