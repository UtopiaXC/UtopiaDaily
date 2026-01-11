from sqlalchemy import Column, Integer, DateTime
from sqlalchemy.sql import func
from src.database.connection import db_manager, session_scope

class CRUDMixin:
    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    @classmethod
    def create(cls, **kwargs):
        instance = cls(**kwargs)
        return instance.save()

    def save(self, commit=True):
        with session_scope() as session:
            session.add(self)
            if commit:
                try:
                    session.commit()
                    session.refresh(self) # 刷新以获取 ID 等自动生成的字段
                except Exception:
                    session.rollback()
                    raise
            else:
                session.flush()
        return self

    def delete(self, commit=True):
        with session_scope() as session:
            session.delete(self)
            if commit:
                session.commit()

    @classmethod
    def get_by_id(cls, id):
        with session_scope() as session:
            return session.query(cls).filter(cls.id == id).first()

    @classmethod
    def get_all(cls, skip=0, limit=100):
        with session_scope() as session:
            return session.query(cls).offset(skip).limit(limit).all()

    def update(self, **kwargs):
        with session_scope() as session:
            session.add(self)
            for key, value in kwargs.items():
                if hasattr(type(self), key):
                    setattr(self, key, value)
            session.commit()
            session.refresh(self)
        return self