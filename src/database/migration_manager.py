import importlib
import os
from sqlalchemy import inspect
from src.database.connection import system_db_manager, system_session_scope, Base
from src.database.models import MigrationVersion, SystemConfig, User, UserRole, UserSession, UserPushConfig, ScraperModule, ScraperModuleConfig, SystemEvent
from src.utils.logger.logger import Log
from src.utils.event import EventManager

TAG = "MIGRATION_MANAGER"

class MigrationManager:
    def __init__(self):
        self.migrations_dir = os.path.join(os.path.dirname(__file__), "migrations")

    def _ensure_system_tables(self):
        system_db_manager.init_db()
        engine = system_db_manager._engine
        inspector = inspect(engine)
        
        tables_to_create = [
            MigrationVersion, 
            SystemConfig,
            UserRole,
            User,
            UserSession,
            UserPushConfig,
            ScraperModule,
            ScraperModuleConfig,
            SystemEvent
        ]

        for model in tables_to_create:
            table_name = model.__tablename__
            if not inspector.has_table(table_name):
                Log.i(TAG, f"Creating {table_name} table...")
                model.__table__.create(engine)

    def _get_applied_versions(self):
        with system_session_scope() as session:
            versions = session.query(MigrationVersion.version_name).all()
            return {v[0] for v in versions}

    def _record_migration(self, version_name, version_code, description):
        with system_session_scope() as session:
            migration = MigrationVersion(
                version_name=version_name,
                version_code=version_code,
                description=description
            )
            session.add(migration)

    def run_migrations(self):
        Log.i(TAG, "Checking for pending migrations...")

        self._ensure_system_tables()
        
        applied_versions = self._get_applied_versions()

        if not os.path.exists(self.migrations_dir):
            os.makedirs(self.migrations_dir)
            
        migration_files = sorted([f for f in os.listdir(self.migrations_dir) if f.endswith(".py") and f != "__init__.py"])
        
        for filename in migration_files:
            module_name = filename[:-3]
            if module_name in applied_versions:
                continue
            Log.i(TAG, f"Applying migration: {module_name}")
            try:
                module_path = f"src.database.migrations.{module_name}"
                migration_module = importlib.import_module(module_path)
                if hasattr(migration_module, 'upgrade'):
                    migration_module.upgrade()
                    version_code = getattr(migration_module, 'VERSION_CODE', 0)
                    description = getattr(migration_module, 'DESCRIPTION', '')
                    self._record_migration(module_name, version_code, description)
                    Log.i(TAG, f"Migration {module_name} applied successfully.")
                    
                    EventManager.record(
                        level=EventManager.LEVEL_NORMAL,
                        category=EventManager.CATEGORY_SYSTEM,
                        event_type="migration_success",
                        summary=f"Applied migration: {module_name}",
                        details={"version": version_code, "desc": description}
                    )
                else:
                    Log.w(TAG, f"Migration {module_name} missing 'upgrade' function. Skipped.")
                    
            except Exception as e:
                Log.e(TAG, f"Failed to apply migration {module_name}", error=e)
                try:
                    EventManager.record(
                        level=EventManager.LEVEL_CRITICAL,
                        category=EventManager.CATEGORY_SYSTEM,
                        event_type="migration_failed",
                        summary=f"Failed to apply migration: {module_name}",
                        details={"error": str(e)},
                        is_resolved=False
                    )
                except:
                    pass
                raise e

        Log.i(TAG, "Migration check completed.")

migration_manager = MigrationManager()
