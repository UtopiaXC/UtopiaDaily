import os
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, scoped_session, declarative_base
from contextlib import contextmanager
from src.utils.logger.logger import Log

TAG = "DB_CONNECTION"

DB_PATH = os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(__file__))), "database/data.db")
SQLALCHEMY_DATABASE_URL = os.getenv("DATABASE_URL", f"sqlite:///{DB_PATH}")

Base = declarative_base()

class DatabaseManager:
    def __init__(self):
        self._engine = None
        self._session_factory = None

    def init_db(self, db_url=SQLALCHEMY_DATABASE_URL):
        if self._engine:
            return

        connect_args = {}
        if "sqlite" in db_url:
            connect_args = {"check_same_thread": False}

        self._engine = create_engine(
            db_url,
            connect_args=connect_args,
            pool_pre_ping=True,
            echo=False
        )

        self._session_factory = scoped_session(
            sessionmaker(autocommit=False, autoflush=False, bind=self._engine)
        )

        # 自动建表（仅用于开发阶段，生产环境建议使用 Alembic 迁移）
        Base.metadata.create_all(bind=self._engine)
        Log.i(TAG, f"Database initialized at {db_url}")

    def get_session(self):
        if not self._session_factory:
            self.init_db()
        return self._session_factory()

    def close(self):
        if self._session_factory:
            self._session_factory.remove()

db_manager = DatabaseManager()

@contextmanager
def session_scope():
    session = db_manager.get_session()
    try:
        yield session
        session.commit()
    except Exception as e:
        Log.e(TAG, "Session rollback due to exception", error=e)
        session.rollback()
        raise
    finally:
        session.close()