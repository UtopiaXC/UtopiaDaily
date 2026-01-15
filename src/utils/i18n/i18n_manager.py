import json
import os
from typing import Dict, Optional
from src.utils.logger.logger import Log

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
        self.locales_dir = os.path.join(os.path.dirname(__file__), "locales")
        
        self._load_locales()
        self._initialized = True

    def _load_locales(self):
        if not os.path.exists(self.locales_dir):
            os.makedirs(self.locales_dir)
            return

        for filename in os.listdir(self.locales_dir):
            if filename.endswith(".json"):
                locale_code = filename[:-5] # remove .json
                try:
                    with open(os.path.join(self.locales_dir, filename), 'r', encoding='utf-8') as f:
                        self.translations[locale_code] = json.load(f)
                    Log.i(TAG, f"Loaded locale: {locale_code}")
                except Exception as e:
                    Log.e(TAG, f"Failed to load locale {filename}", error=e)

    def load_module_locales(self, module_locales_dir: str):
        """
        Load locale files from a module's directory and merge them into the main translations.
        """
        if not os.path.exists(module_locales_dir):
            return

        for filename in os.listdir(module_locales_dir):
            if filename.endswith(".json"):
                locale_code = filename[:-5] # remove .json
                try:
                    with open(os.path.join(module_locales_dir, filename), 'r', encoding='utf-8') as f:
                        module_trans = json.load(f)
                        
                        if locale_code not in self.translations:
                            self.translations[locale_code] = {}

                        self.translations[locale_code].update(module_trans)
                        Log.i(TAG, f"Merged module locale: {locale_code} from {module_locales_dir}")
                except Exception as e:
                    Log.e(TAG, f"Failed to load module locale {filename} from {module_locales_dir}", error=e)

    def set_locale(self, locale: str):
        if locale in self.translations:
            self.current_locale = locale
        else:
            Log.w(TAG, f"Locale {locale} not found, keeping {self.current_locale}")

    def t(self, key: str, locale: Optional[str] = None, **kwargs) -> str:
        """
        Translate a key.
        :param key: The translation key (e.g., 'error.not_found')
        :param locale: Optional locale override. If None, uses current_locale.
        :param kwargs: Arguments for string formatting (e.g., name="User")
        :return: Translated string or the key itself if not found.
        """
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
