from src.database.connection import system_db_manager
from src.database.models import ScraperModuleTask
from sqlalchemy import inspect

VERSION_CODE = 1
DESCRIPTION = "Create scraper_module_tasks table"

def upgrade():
    engine = system_db_manager._engine
    inspector = inspect(engine)
    
    if not inspector.has_table(ScraperModuleTask.__tablename__):
        ScraperModuleTask.__table__.create(engine)
