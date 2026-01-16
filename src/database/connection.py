import os
import sys
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, scoped_session, declarative_base
from sqlalchemy.exc import OperationalError
from contextlib import contextmanager
from dotenv import load_dotenv
from src.utils.logger.logger import Log

TAG = "DB_CONNECTION"

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
ENV_PATH = os.path.join(PROJECT_ROOT, "./config/.env")

load_dotenv(ENV_PATH)

SYSTEM_DB_DIR = os.path.join(PROJECT_ROOT, "database", "system")
DATA_DB_DIR = os.path.join(PROJECT_ROOT, "database", "data")

SYSTEM_DB_PATH = os.path.join(SYSTEM_DB_DIR, "system.db")
DATA_DB_PATH = os.path.join(DATA_DB_DIR, "data.db")

def get_db_url(db_name_suffix=""):
    db_type = os.getenv("DB_TYPE", "SQLITE").upper()
    
    if db_type == "SQLITE":
        if db_name_suffix == "system":
            return f"sqlite:///{SYSTEM_DB_PATH}"
        else:
            return f"sqlite:///{DATA_DB_PATH}"
            
    db_host = os.getenv("DB_HOST", "localhost")
    db_port = os.getenv("DB_PORT", "5432")
    db_user = os.getenv("DB_USER", "root")
    db_password = os.getenv("DB_PASSWORD", "root")
    db_name = os.getenv("DB_NAME", "utopia_daily")
    
    target_db_name = f"{db_name}_{db_name_suffix}" if db_name_suffix else db_name

    if db_type == "MYSQL":
        port_str = f":{db_port}" if db_port else ""
        return f"mysql+pymysql://{db_user}:{db_password}@{db_host}{port_str}/{target_db_name}"
    
    elif db_type == "POSTGRESQL":
        port_str = f":{db_port}" if db_port else ""
        return f"postgresql://{db_user}:{db_password}@{db_host}{port_str}/{target_db_name}"
    
    else:
        Log.w(TAG, f"Unknown DB_TYPE: {db_type}, falling back to SQLITE")
        if db_name_suffix == "system":
            return f"sqlite:///{SYSTEM_DB_PATH}"
        else:
            return f"sqlite:///{DATA_DB_PATH}"

SYSTEM_DB_URL = get_db_url("system")
DATA_DB_URL = get_db_url("data")

Log.w(TAG,SYSTEM_DB_URL)

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

        if "sqlite" not in self.db_url:
            try:
                with self._engine.connect() as conn:
                    pass
            except OperationalError as e:
                error_msg = str(e).lower()
                if "database" in error_msg and ("does not exist" in error_msg or "unknown" in error_msg):
                    try:
                        from sqlalchemy.engine.url import make_url
                        url = make_url(self.db_url)
                        db_name = url.database
                    except:
                        db_name = "unknown"
                    Log.e(self.tag, f"Database '{db_name}' does not exist. Please create it manually before running the application.", stack_trace=False)
                    sys.exit(1)
                else:
                    Log.e(self.tag, f"Failed to connect to database: {e}", stack_trace=False)
                    sys.exit(1)

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
