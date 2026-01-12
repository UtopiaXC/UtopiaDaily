from sqlalchemy import Column, String, Boolean, BigInteger
from src.database.connection import Base
import uuid
import time

def get_utc_timestamp_ms():
    return int(time.time() * 1000)

class BaseModel(Base):
    __abstract__ = True

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    is_deleted = Column(Boolean, default=False, nullable=False)
    created_at = Column(BigInteger, default=get_utc_timestamp_ms, nullable=False)
    updated_at = Column(BigInteger, default=get_utc_timestamp_ms, onupdate=get_utc_timestamp_ms, nullable=False)
