import json
import os
from typing import Dict, Optional, List
from src.utils.logger.logger import Log
from src.utils.cache_manage import cache_manager

TAG = "I18N"

class I18nManager:
    _instance = None
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(I18nManager, cls).__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self):
        if self._initialized:
            return
        
        self.default_locale = "en_US"
        self.current_locale = self.default_locale
        self.translations: Dict[str, Dict[str, str]] = {}
        self.base_locales_dir = os.path.join(os.path.dirname(__file__), "locales")
        self._load_from_cache()
        if not self.translations:
            self._load_locales_from_dir(self.base_locales_dir, "base")
            
        self._initialized = True

    def _load_from_cache(self):
        cache_dir = cache_manager.get_cache_dir("locales")
        if not os.path.exists(cache_dir):
            return

        for filename in os.listdir(cache_dir):
            if filename.endswith(".json"):
                locale_code = filename[:-5]
                data = cache_manager.get(f"locales/{locale_code}")
                if data:
                    self.translations[locale_code] = data
                    Log.i(TAG, f"Loaded locale {locale_code} from cache")

    def _load_locales_from_dir(self, directory: str, source_name: str):
        if not os.path.exists(directory):
            return

        for filename in os.listdir(directory):
            if filename.endswith(".json"):
                locale_code = filename[:-5] # remove .json
                try:
                    with open(os.path.join(directory, filename), 'r', encoding='utf-8') as f:
                        self.translations[locale_code] = json.load(f)
                    Log.i(TAG, f"Loaded locale {locale_code} from {source_name}")
                except Exception as e:
                    Log.e(TAG, f"Failed to load locale {filename} from {source_name}", error=e)

    def compile_locales(self, module_locale_dirs: List[str]):
        Log.i(TAG, "Compiling locales...")
        merged_translations = {}
        if os.path.exists(self.base_locales_dir):
            for filename in os.listdir(self.base_locales_dir):
                if filename.endswith(".json"):
                    locale_code = filename[:-5]
                    with open(os.path.join(self.base_locales_dir, filename), 'r', encoding='utf-8') as f:
                        merged_translations[locale_code] = json.load(f)

        for module_dir in module_locale_dirs:
            if not os.path.exists(module_dir):
                continue
            for filename in os.listdir(module_dir):
                if filename.endswith(".json"):
                    locale_code = filename[:-5]
                    try:
                        with open(os.path.join(module_dir, filename), 'r', encoding='utf-8') as f:
                            module_trans = json.load(f)
                            if locale_code not in merged_translations:
                                merged_translations[locale_code] = {}
                            merged_translations[locale_code].update(module_trans)
                    except Exception as e:
                        Log.w(TAG, f"Failed to merge locale from {module_dir}/{filename}: {e}")

        for locale_code, trans in merged_translations.items():
            cache_manager.set(f"locales/{locale_code}", trans)

        self.translations = merged_translations
        Log.i(TAG, "Locales compiled and cached.")

    def set_locale(self, locale: str):
        if locale in self.translations:
            self.current_locale = locale
        else:
            Log.w(TAG, f"Locale {locale} not found, keeping {self.current_locale}")

    def t(self, key: str, locale: Optional[str] = None, **kwargs) -> str:
        target_locale = locale or self.current_locale
        if target_locale not in self.translations:
            target_locale = self.default_locale
        trans_dict = self.translations.get(target_locale, {})
        text = trans_dict.get(key)
        if text is None:
            trans_dict = self.translations.get(self.default_locale, {})
            text = trans_dict.get(key)
        if text is None:
            return key
        try:
            return text.format(**kwargs)
        except Exception as e:
            Log.w(TAG, f"Format error for key '{key}': {e}")
            return text

i18n = I18nManager()
