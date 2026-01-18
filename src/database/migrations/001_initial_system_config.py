import json
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
        "hint": "config.server_name.hint",
        "type": "string",
        "tab": "system",
        "group": "basic",
        "options": None,
        "is_editable": True,
        "is_public": True,
        "order": 2
    },
    {
        "key": "system_version",
        "value": "1.0.0",
        "default": "1.0.0",
        "description": "config.system_version.desc",
        "hint": "config.system_version.hint",
        "type": "string",
        "tab": "system",
        "group": "basic",
        "options": None,
        "is_editable": False,
        "order": 1
    },
    {
        "key": "log_level",
        "value": "INFO",
        "default": "INFO",
        "description": "config.log_level.desc",
        "hint": "config.log_level.hint",
        "type": "select",
        "tab": "system",
        "group": "basic",
        "options": ["DEBUG", "INFO", "WARNING", "ERROR"],
        "is_editable": True,
        "order": 4
    },
    {
        "key": "default_locale",
        "value": "auto",
        "default": "auto",
        "description": "config.default_locale.desc",
        "hint": "config.default_locale.hint",
        "type": "select",
        "tab": "system",
        "group": "basic",
        "options": None,
        "is_editable": True,
        "order": 3
    },
    # AI Configs
    {
        "key": "ai_model",
        "value": "None",
        "default": "None",
        "description": "config.ai.model",
        "hint": "config.ai.model.hint",
        "tab": "ai_config",
        "group": "basic_ai",
        "type": "model_select", # Custom type for model selection with download
        "options": [], # Will be populated dynamically
        "order": 1
    },
    {
        "key": "ai_confidence",
        "value": "60",
        "default": "60",
        "description": "config.ai.confidence",
        "hint": "config.ai.confidence.hint",
        "tab": "ai_config",
        "group": "basic_ai",
        "type": "int",
        "order": 2
    },
    {
        "key": "ai_tags",
        "value": "{\"groups\": [], \"tags\": []}",
        "default": "{\"groups\": [], \"tags\": []}",
        "description": "config.ai.tags",
        "hint": "config.ai.tags.hint",
        "tab": "ai_config",
        "group": "tag_config",
        "type": "json",
        "order": 1
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
                    hint=config.get("hint", ""),
                    type=config.get("type", "string"),
                    tab=config.get("tab", "system"),
                    group=config.get("group", "basic"),
                    options=config.get("options"),
                    is_editable=config.get("is_editable", True),
                    is_public=config.get("is_public", False),
                    order=config.get("order", 0)
                )
                session.add(new_config)
            else:
                existing.type = config.get("type", "string")
                existing.tab = config.get("tab", "system")
                existing.group = config.get("group", "basic")
                existing.options = config.get("options")
                existing.hint = config.get("hint", "")
                existing.is_editable = config.get("is_editable", True)
                existing.is_public = config.get("is_public", False)
                existing.order = config.get("order", 0)
                if existing.key == "default_locale" and existing.value not in ["auto", "en_US", "zh_CN"]:
                     existing.value = "auto"
                Log.i(TAG, f"Config {config['key']} updated.")
