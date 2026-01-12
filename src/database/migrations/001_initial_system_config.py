from src.database.connection import system_session_scope
from src.database.models import SystemConfig
from src.utils.logger.logger import Log

VERSION_CODE = 1
DESCRIPTION = "Initialize system configuration with default values"

TAG = "MIGRATION_001"

# Use i18n keys for description instead of raw text
DEFAULT_CONFIGS = [
    {
        "key": "system_version",
        "value": "1.0.0",
        "default": "1.0.0",
        "description": "config.system_version.desc" 
    },
    {
        "key": "log_level",
        "value": "INFO",
        "default": "INFO",
        "description": "config.log_level.desc"
    }
]

def upgrade():
    Log.i(TAG, "Starting upgrade...")
    with system_session_scope() as session:
        for config in DEFAULT_CONFIGS:
            existing = session.query(SystemConfig).filter_by(key=config["key"]).first()
            if not existing:
                Log.i(TAG, f"Adding config: {config['key']}")
                new_config = SystemConfig(
                    key=config["key"],
                    value=config["value"],
                    default=config["default"],
                    description=config["description"]
                )
                session.add(new_config)
            else:
                Log.i(TAG, f"Config {config['key']} already exists. Skipping.")
