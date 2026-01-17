from src.database.connection import system_db_manager
from src.database.models import ScraperModule
from src.utils.logger.logger import Log
from sqlalchemy import inspect

VERSION_CODE = 1
DESCRIPTION = "Create scraper_modules table"

TAG = "MIGRATION_003"

def upgrade():
    Log.i(TAG, "Starting upgrade...")
    engine = system_db_manager._engine
    inspector = inspect(engine)
    
    if not inspector.has_table(ScraperModule.__tablename__):
        Log.i(TAG, f"Creating table: {ScraperModule.__tablename__}")
        ScraperModule.__table__.create(engine)
    else:
        Log.i(TAG, f"Table {ScraperModule.__tablename__} already exists.")
        columns = [c['name'] for c in inspector.get_columns(ScraperModule.__tablename__)]
        if 'source' not in columns:
            Log.i(TAG, "Adding source column to scraper_modules table")
            from sqlalchemy import text
            with engine.connect() as conn:
                conn.execute(text(f"ALTER TABLE {ScraperModule.__tablename__} ADD COLUMN source VARCHAR(20) DEFAULT 'unknown'"))
                conn.commit()
