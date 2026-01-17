from sqlalchemy import Column, String, Boolean, Text
from sqlalchemy.dialects.sqlite import JSON
from src.database.models.base_model import BaseModel

class SystemEvent(BaseModel):
    __tablename__ = 'system_events'

    # Levels: NORMAL, WARNING, CRITICAL, FATAL
    level = Column(String(20), nullable=False, default="NORMAL", index=True)
    
    # Categories: SYSTEM, USER, MODULE, TASK, SECURITY
    category = Column(String(20), nullable=False, default="SYSTEM", index=True)
    
    # Event Type: user_login, module_conflict, etc.
    event_type = Column(String(50), nullable=False, index=True)
    
    # Summary: Short description for list view
    summary = Column(String(255), nullable=False)
    
    # Details: Structured data for detailed view
    details = Column(JSON, nullable=True)
    
    # Source ID: User ID or Module ID (optional)
    source_id = Column(String(100), nullable=True, index=True)
    
    # Is Resolved: For warnings/errors, track if they are handled
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
