from sqlalchemy import Column, String, Boolean, Text
from sqlalchemy.dialects.sqlite import JSON
from src.database.models.base_model import BaseModel

class ScraperModule(BaseModel):
    __tablename__ = 'scraper_modules'

    module_id = Column(String(100), unique=True, nullable=False, index=True)
    name = Column(String(100), nullable=False)
    description = Column(Text, nullable=True)
    version = Column(String(50), nullable=False)
    author = Column(String(100), nullable=True)
    meta = Column(JSON, nullable=True)
    source = Column(String(20), nullable=True, default="unknown")
    is_enable = Column(Boolean, default=False, nullable=False)
