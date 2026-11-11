from sqlalchemy import Column, String, Boolean, Text
from sqlalchemy.dialects.sqlite import JSON
from src.database.models.base_model import BaseModel

class SystemEvent(BaseModel):
    __tablename__ = 'system_events'
    level = Column(String(20), nullable=False, default="NORMAL", index=True)
    category = Column(String(20), nullable=False, default="SYSTEM", index=True)
    event_type = Column(String(50), nullable=False, index=True)
    summary = Column(String(255), nullable=False)
    details = Column(JSON, nullable=True)
    source_id = Column(String(100), nullable=True, index=True)
    is_resolved = Column(Boolean, default=True, nullable=False)

    def to_dict(self):
        return {
            "id": self.id,
            "level": self.level,
            "category": self.category,
            "event_type": self.event_type,
            "summary": self.summary,
            "details": self.details,
            "source_id": self.source_id,
            "is_resolved": self.is_resolved,
            "created_at": self.created_at,
            "updated_at": self.updated_at
        }
