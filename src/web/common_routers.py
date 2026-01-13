from fastapi import APIRouter
from src.utils.i18n import i18n

router = APIRouter(prefix="/api/common", tags=["Common"])

@router.get("/i18n/{locale}")
async def get_translations(locale: str):
    """
    Returns the full translation dictionary for a specific locale.
    Frontend can use this to load translations dynamically.
    """
    # If locale doesn't exist, fallback to default or return empty
    translations = i18n.translations.get(locale)
    if not translations:
        # Try default locale if requested one is missing
        translations = i18n.translations.get(i18n.default_locale, {})
    
    return translations

@router.get("/i18n/locales")
async def get_available_locales():
    """
    Returns a list of available locale codes.
    """
    return list(i18n.translations.keys())
