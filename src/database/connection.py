import os
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, scoped_session, declarative_base
from contextlib import contextmanager
from src.utils.logger.logger import Log

TAG = "DB_CONNECTION"

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
SYSTEM_DB_DIR = os.path.join(PROJECT_ROOT, "database", "system")
DATA_DB_DIR = os.path.join(PROJECT_ROOT, "database", "data")

SYSTEM_DB_PATH = os.path.join(SYSTEM_DB_DIR, "system.db")
DATA_DB_PATH = os.path.join(DATA_DB_DIR, "data.db")

SYSTEM_DB_URL = f"sqlite:///{SYSTEM_DB_PATH}"
DATA_DB_URL = f"sqlite:///{DATA_DB_PATH}"

Base = declarative_base()

class DatabaseManager:
    def __init__(self, db_url, tag_suffix):
        self.db_url = db_url
        self.tag = f"{TAG}_{tag_suffix}"
        self._engine = None
        self._session_factory = None

    def _ensure_directory(self):
        if "sqlite" in self.db_url:
            path = self.db_url.replace("sqlite:///", "")
            directory = os.path.dirname(path)
            if not os.path.exists(directory):
                os.makedirs(directory)
                Log.i(self.tag, f"Created database directory: {directory}")

    def init_db(self):
        if self._engine:
            return

        self._ensure_directory()

        connect_args = {}
        if "sqlite" in self.db_url:
            connect_args = {"check_same_thread": False}

        self._engine = create_engine(
            self.db_url,
            connect_args=connect_args,
            pool_pre_ping=True,
            echo=False
        )

        self._session_factory = scoped_session(
            sessionmaker(autocommit=False, autoflush=False, bind=self._engine)
        )
        
        Log.i(self.tag, f"Database initialized at {self.db_url}")

    def create_tables(self, base=Base):
        """Explicitly create tables bound to this engine"""
        if not self._engine:
            self.init_db()
        base.metadata.create_all(bind=self._engine)

    def get_session(self):
        if not self._session_factory:
            self.init_db()
        return self._session_factory()

    def close(self):
        if self._session_factory:
            self._session_factory.remove()

system_db_manager = DatabaseManager(SYSTEM_DB_URL, "SYSTEM")
data_db_manager = DatabaseManager(DATA_DB_URL, "DATA")

@contextmanager
def system_session_scope():
    session = system_db_manager.get_session()
    try:
        yield session
        session.commit()
    except Exception as e:
        Log.e("DB_SESSION_SYSTEM", "Session rollback due to exception", error=e)
        session.rollback()
        raise
    finally:
        session.close()

@contextmanager
def data_session_scope():
    session = data_db_manager.get_session()
    try:
        yield session
        session.commit()
    except Exception as e:
        Log.e("DB_SESSION_DATA", "Session rollback due to exception", error=e)
        session.rollback()
        raise
    finally:
        session.close()
