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
        "value": "sentence-transformers/all-MiniLM-L6-v2",
        "default": "sentence-transformers/all-MiniLM-L6-v2",
        "description": "config.ai.model",
        "hint": "config.ai.model.hint",
        "tab": "ai_config",
        "group": "basic_ai",
        "type": "select",
        "options": [
            "albert-base-v1", "albert-base-v2", "albert-large-v1", "albert-large-v2", "albert-xlarge-v1", "albert-xlarge-v2", "albert-xxlarge-v1", "albert-xxlarge-v2",
            "bert-base-cased-finetuned-mrpc", "bert-base-cased", "bert-base-chinese", "bert-base-german-cased", "bert-base-german-dbmdz-cased", "bert-base-german-dbmdz-uncased",
            "bert-base-multilingual-cased", "bert-base-multilingual-uncased", "bert-base-uncased", "bert-large-cased-whole-word-masking-finetuned-squad",
            "bert-large-cased-whole-word-masking", "bert-large-cased", "bert-large-uncased-whole-word-masking-finetuned-squad", "bert-large-uncased-whole-word-masking",
            "bert-large-uncased", "camembert-base", "ctrl", "distilbert-base-cased-distilled-squad", "distilbert-base-cased", "distilbert-base-german-cased",
            "distilbert-base-multilingual-cased", "distilbert-base-uncased-distilled-squad", "distilbert-base-uncased-finetuned-sst-2-english", "distilbert-base-uncased",
            "distilgpt2", "distilroberta-base", "gpt2-large", "gpt2-medium", "gpt2-xl", "gpt2", "openai-gpt", "roberta-base-openai-detector", "roberta-base",
            "roberta-large-mnli", "roberta-large-openai-detector", "roberta-large", "sentence-transformers/all-MiniLM-L6-v2", "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2",
            "t5-11b", "t5-3b", "t5-base", "t5-large", "t5-small", "transfo-xl-wt103", "xlm-clm-ende-1024", "xlm-clm-enfr-1024", "xlm-mlm-100-1280", "xlm-mlm-17-1280",
            "xlm-mlm-en-2048", "xlm-mlm-ende-1024", "xlm-mlm-enfr-1024", "xlm-mlm-enro-1024", "xlm-mlm-tlm-xnli15-1024", "xlm-mlm-xnli15-1024", "xlm-roberta-base",
            "xlm-roberta-large-finetuned-conll02-dutch", "xlm-roberta-large-finetuned-conll02-spanish", "xlm-roberta-large-finetuned-conll03-english",
            "xlm-roberta-large-finetuned-conll03-german", "xlm-roberta-large", "xlnet-base-cased", "xlnet-large-cased"
        ],
        "order": 1
    },
    {
        "key": "ai_confidence",
        "value": "80",
        "default": "80",
        "description": "config.ai.confidence",
        "hint": "config.ai.confidence.hint",
        "tab": "ai_config",
        "group": "basic_ai",
        "type": "int",
        "order": 2
    },
    {
        "key": "ai_max_categories",
        "value": "3",
        "default": "3",
        "description": "config.ai.max_categories",
        "hint": "config.ai.max_categories.hint",
        "tab": "ai_config",
        "group": "basic_ai",
        "type": "int",
        "order": 3
    },
    {
        "key": "ai_tags",
        "value": "[]",
        "default": "[]",
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
