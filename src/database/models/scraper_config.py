from sqlalchemy import Column, String, Boolean, Text, ForeignKey, UniqueConstraint
from sqlalchemy.dialects.sqlite import JSON
from src.database.models.base_model import BaseModel

class ScraperModuleConfig(BaseModel):
    __tablename__ = 'scraper_module_configs'

    module_id = Column(String(100), ForeignKey('scraper_modules.module_id'), nullable=False, index=True)
    config_key = Column(String(100), nullable=False)
    description = Column(String(255), nullable=True)
    type = Column(String(50), default="text", nullable=False)  # text, number, switch, date, datetime, select, password, array
    options = Column(JSON, nullable=True)  # For select/switch types
    value = Column(JSON, nullable=True)  # The actual value
    hint = Column(String(255), nullable=True)
    regex = Column(String(255), nullable=True)
    source = Column(String(20), default="default", nullable=False)  # default, custom
    is_override = Column(Boolean, default=False, nullable=False)

    # Ensure unique config key per module
    __table_args__ = (
        UniqueConstraint('module_id', 'config_key', name='uix_module_config_key'),
    )

    def to_dict(self):
        return {
            "id": self.id,
            "module_id": self.module_id,
            "config_key": self.config_key,
            "description": self.description,
            "type": self.type,
            "options": self.options,
            "value": self.value,
            "hint": self.hint,
            "regex": self.regex,
            "source": self.source,
            "is_override": self.is_override,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }
