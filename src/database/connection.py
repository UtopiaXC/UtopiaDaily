import os
from sqlalchemy import create_engine, event
from sqlalchemy.engine import URL
from sqlalchemy.orm import sessionmaker, scoped_session, declarative_base
from sqlalchemy.exc import OperationalError
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
from contextlib import contextmanager, asynccontextmanager
from src.utils.logger.logger import Log
from src.utils.env_manage.env_manager import EnvManager

TAG = "DB_CONNECTION"

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))

SYSTEM_DB_DIR = os.path.join(PROJECT_ROOT, "database", "system")
DATA_DB_DIR = os.path.join(PROJECT_ROOT, "database", "data")

SYSTEM_DB_PATH = os.path.join(SYSTEM_DB_DIR, "system.db")
DATA_DB_PATH = os.path.join(DATA_DB_DIR, "data.db")

def get_db_url(db_name_suffix="", is_async=False):
    db_type = EnvManager.get_env("DB_TYPE", "SQLITE").upper()
    
    if db_type == "SQLITE":
        driver = "sqlite+aiosqlite" if is_async else "sqlite"
        if db_name_suffix == "system":
            return f"{driver}:///{SYSTEM_DB_PATH}"
        else:
            return f"{driver}:///{DATA_DB_PATH}"
            
    db_host = EnvManager.get_env("DB_HOST", "")
    db_port = EnvManager.get_env("DB_PORT", "")

    try:
        db_port = int(db_port) if db_port else None
    except ValueError:
        Log.w(TAG, f"Invalid DB_PORT: {db_port}, using default")
        db_port = None

    db_user = EnvManager.get_env("DB_USER", "")
    db_password = EnvManager.get_env("DB_PASSWORD", "")
    db_name = EnvManager.get_env("DB_NAME", "")
    
    target_db_name = f"{db_name}_{db_name_suffix}" if db_name_suffix else db_name

    if db_type == "MYSQL":
        driver = "mysql+aiomysql" if is_async else "mysql+pymysql"
        return URL.create(
            drivername=driver,
            username=db_user,
            password=db_password,
            host=db_host,
            port=db_port,
            database=target_db_name
        ).render_as_string(hide_password=False)
    
    elif db_type == "POSTGRESQL":
        driver = "postgresql+asyncpg" if is_async else "postgresql"
        return URL.create(
            drivername=driver,
            username=db_user,
            password=db_password,
            host=db_host,
            port=db_port,
            database=target_db_name
        ).render_as_string(hide_password=False)
    
    else:
        Log.w(TAG, f"Unknown DB_TYPE: {db_type}, falling back to SQLITE")
        driver = "sqlite+aiosqlite" if is_async else "sqlite"
        if db_name_suffix == "system":
            return f"{driver}:///{SYSTEM_DB_PATH}"
        else:
            return f"{driver}:///{DATA_DB_PATH}"

SYSTEM_DB_URL = get_db_url("system", is_async=False)
DATA_DB_URL = get_db_url("data", is_async=False)
SYSTEM_ASYNC_DB_URL = get_db_url("system", is_async=True)
DATA_ASYNC_DB_URL = get_db_url("data", is_async=True)

Log.w(TAG, f"Sync URL: {SYSTEM_DB_URL}")
Log.w(TAG, f"Async URL: {SYSTEM_ASYNC_DB_URL}")

Base = declarative_base()

class DatabaseManager:
    def __init__(self, db_url, async_db_url, tag_suffix):
        self.db_url = db_url
        self.async_db_url = async_db_url
        self.tag = f"{TAG}_{tag_suffix}"
        self._engine = None
        self._async_engine = None
        self._session_factory = None
        self._async_session_factory = None

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

        # Sync Engine Initialization
        connect_args = {}
        if "sqlite" in self.db_url:
            connect_args = {"check_same_thread": False}

        self._engine = create_engine(
            self.db_url,
            connect_args=connect_args,
            pool_pre_ping=True,
            echo=False
        )

        # Enable WAL mode for SQLite (Sync)
        if "sqlite" in self.db_url:
            @event.listens_for(self._engine, "connect")
            def set_sqlite_pragma(dbapi_connection, connection_record):
                cursor = dbapi_connection.cursor()
                try:
                    cursor.execute("PRAGMA journal_mode=WAL")
                    cursor.execute("PRAGMA synchronous=NORMAL")
                    Log.i(self.tag, "SQLite WAL mode enabled")
                except Exception as e:
                    Log.w(self.tag, f"Failed to set SQLite PRAGMA: {e}")
                finally:
                    cursor.close()

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
                    Log.fatal(self.tag, f"Database '{db_name}' does not exist. Please create it manually before running the application.", stack_trace=False)
                else:
                    Log.fatal(self.tag, f"Failed to connect to database: {e}", stack_trace=False)

        self._session_factory = scoped_session(
            sessionmaker(autocommit=False, autoflush=False, bind=self._engine)
        )

        self._async_engine = create_async_engine(
            self.async_db_url,
            echo=False,
            pool_pre_ping=True
        )
        
        self._async_session_factory = async_sessionmaker(
            self._async_engine, expire_on_commit=False, class_=AsyncSession
        )

        Log.i(self.tag, f"Database initialized at {self.db_url} (Async: {self.async_db_url})")

    def create_tables(self, base=Base):
        if not self._engine:
            self.init_db()
        base.metadata.create_all(bind=self._engine)

    def get_session(self):
        if not self._session_factory:
            self.init_db()
        return self._session_factory()
        
    def get_async_session(self) -> AsyncSession:
        if not self._async_session_factory:
            self.init_db()
        return self._async_session_factory()

    def close(self):
        if self._session_factory:
            self._session_factory.remove()

system_db_manager = DatabaseManager(SYSTEM_DB_URL, SYSTEM_ASYNC_DB_URL, "SYSTEM")
data_db_manager = DatabaseManager(DATA_DB_URL, DATA_ASYNC_DB_URL, "DATA")

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

@asynccontextmanager
async def system_async_session_scope():
    session = system_db_manager.get_async_session()
    try:
        yield session
        await session.commit()
    except Exception as e:
        Log.e("DB_ASYNC_SESSION_SYSTEM", "Async session rollback due to exception", error=e)
        await session.rollback()
        raise
    finally:
        await session.close()
