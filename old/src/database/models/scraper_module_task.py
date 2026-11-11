from sqlalchemy import Column, String, Text, ForeignKey, UniqueConstraint
from src.database.models.base_model import BaseModel

class ScraperModuleTask(BaseModel):
    __tablename__ = 'scraper_module_tasks'

    module_id = Column(String(36), ForeignKey('scraper_modules.id'), nullable=False, index=True)
    task_key = Column(String(100), nullable=False)
    name = Column(String(100), nullable=False)
    description = Column(Text, nullable=True)

    __table_args__ = (
        UniqueConstraint('module_id', 'task_key', name='uix_module_task_key'),
    )

    def to_dict(self):
        return {
            "id": self.id,
            "module_id": self.module_id,
            "task_key": self.task_key,
            "name": self.name,
            "description": self.description,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }
