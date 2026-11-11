from sqlalchemy import Column, String, Integer
from src.database.models.base_model import BaseModel

class MigrationVersion(BaseModel):
    __tablename__ = 'migration_version'

    version_name = Column(String(255), unique=True, nullable=False)
    version_code = Column(Integer, nullable=False)
    description = Column(String(255), nullable=True)
