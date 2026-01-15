from src.database.connection import system_session_scope
from src.database.models import SystemConfig
from src.utils.logger.logger import Log

VERSION_CODE = 1
DESCRIPTION = "Initialize system configuration with default values"

TAG = "MIGRATION_001"

DEFAULT_CONFIGS = [
    {
        "key": "server_name",
        "value": "Utopia Daily",
        "default": "Utopia Daily",
        "description": "config.server_name.desc",
        "type": "string",
        "group": "system",
        "options": None,
        "is_editable": True,
        "is_public": True
    },
    {
        "key": "system_version",
        "value": "1.0.0",
        "default": "1.0.0",
        "description": "config.system_version.desc",
        "type": "string",
        "group": "system",
        "options": None,
        "is_editable": False
    },
    {
        "key": "log_level",
        "value": "INFO",
        "default": "INFO",
        "description": "config.log_level.desc",
        "type": "select",
        "group": "system",
        "options": ["DEBUG", "INFO", "WARNING", "ERROR"],
        "is_editable": True
    },
    {
        "key": "default_locale",
        "value": "auto",
        "default": "auto",
        "description": "config.default_locale.desc",
        "type": "select",
        "group": "system",
        "options": None, # Frontend will populate this dynamically
        "is_editable": True
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
                    description=config["description"],
                    type=config.get("type", "string"),
                    group=config.get("group", "system"),
                    options=config.get("options"),
                    is_editable=config.get("is_editable", True),
                    is_public=config.get("is_public", False)
                )
                session.add(new_config)
            else:
                # Update existing config structure if needed (dev mode)
                existing.type = config.get("type", "string")
                existing.options = config.get("options")
                existing.is_editable = config.get("is_editable", True)
                existing.is_public = config.get("is_public", False)
                # Reset value to auto if it was hardcoded before (optional, but good for dev)
                if existing.key == "default_locale" and existing.value not in ["auto", "en_US", "zh_CN"]:
                     existing.value = "auto"
                Log.i(TAG, f"Config {config['key']} updated.")
