import json
import os
from fastapi import APIRouter, Query, Depends
from sqlalchemy.orm import Session
from src.utils.i18n import i18n
from src.database.connection import system_db_manager
from src.database.models import SystemConfig

router = APIRouter(prefix="/api/common", tags=["Common"])

def get_db():
    db = system_db_manager.get_session()
    try:
        yield db
    finally:
        db.close()

@router.get("/system-info")
async def get_system_info(db: Session = Depends(get_db)):
    server_name_config = db.query(SystemConfig).filter(SystemConfig.key == "server_name").first()
    server_version_config = db.query(SystemConfig).filter(SystemConfig.key == "server_version").first()
    server_name = server_name_config.value if server_name_config else None
    server_version = server_version_config.value if server_version_config else None
    return {
        "server_name": server_name,
        "version": server_version
    }

@router.get("/i18n/{locale}")
async def get_translations(locale: str):
    translations = i18n.translations.get(locale)
    if not translations:
        translations = i18n.translations.get(i18n.default_locale, {})
    
    return translations

@router.get("/locales")
async def get_available_locales():
    try:
        lang_file = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "utils", "i18n", "languages.json")
        if os.path.exists(lang_file):
            with open(lang_file, 'r', encoding='utf-8') as f:
                languages = json.load(f)
                return [{"label": label, "value": code} for code, label in languages.items()]
    except Exception as e:
        print(f"Error loading languages.json: {e}")
    return [
        {"label": "Auto", "value": "auto"},
        {"label": "English", "value": "en_US"},
        {"label": "中文", "value": "zh_CN"}
    ]
