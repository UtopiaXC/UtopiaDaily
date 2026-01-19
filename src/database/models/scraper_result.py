from sqlalchemy import Column, String, JSON, Index
from src.database.models.data_model import DataModel

class ScraperResult(DataModel):
    __tablename__ = "scraper_results"

    module_id = Column(String(50), nullable=False, index=True)
    fingerprint = Column(String(255), nullable=False, index=True)
    value = Column(JSON, nullable=False)

    __table_args__ = (
        Index('idx_module_fingerprint', 'module_id', 'fingerprint', unique=True),
    )

    def __repr__(self):
        return f"<ScraperResult(id={self.id}, module_id={self.module_id}, fingerprint={self.fingerprint})>"
