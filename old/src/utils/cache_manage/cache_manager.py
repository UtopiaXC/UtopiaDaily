import os
import json
import shutil
from typing import Any, Optional
from src.utils.logger.logger import Log

TAG = "CACHE_MANAGER"

class CacheManager:
    _instance = None
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(CacheManager, cls).__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self):
        if self._initialized:
            return
            
        # Determine project root
        current_file = os.path.abspath(__file__)
        # src/utils/cache_manage/cache_manager.py -> src/utils/cache_manage -> src/utils -> src -> project_root
        self.project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(current_file))))
        self.cache_root = os.path.join(self.project_root, "cache")
        
        if not os.path.exists(self.cache_root):
            try:
                os.makedirs(self.cache_root)
            except Exception as e:
                Log.e(TAG, "Failed to create cache root", error=e)
                
        self._initialized = True

    def get_cache_dir(self, subdir: Optional[str] = None) -> str:
        """
        Get the absolute path to a cache directory. Creates it if not exists.
        """
        path = self.cache_root
        if subdir:
            path = os.path.join(path, subdir)
        
        if not os.path.exists(path):
            try:
                os.makedirs(path)
            except Exception as e:
                Log.e(TAG, f"Failed to create cache subdir {subdir}", error=e)
        return path

    def set(self, key: str, data: Any):
        """
        Save data to cache as JSON. Key can contain slashes for subdirectories.
        """
        try:
            file_path = os.path.join(self.cache_root, f"{key}.json")
            dir_path = os.path.dirname(file_path)
            if not os.path.exists(dir_path):
                os.makedirs(dir_path)
                
            with open(file_path, 'w', encoding='utf-8') as f:
                json.dump(data, f, ensure_ascii=False, indent=2)
        except Exception as e:
            Log.e(TAG, f"Failed to set cache for {key}", error=e)

    def get(self, key: str) -> Optional[Any]:
        """
        Load data from cache.
        """
        try:
            file_path = os.path.join(self.cache_root, f"{key}.json")
            if not os.path.exists(file_path):
                return None
                
            with open(file_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            Log.e(TAG, f"Failed to get cache for {key}", error=e)
            return None

    def delete(self, key: str):
        """
        Delete a cache file.
        """
        try:
            file_path = os.path.join(self.cache_root, f"{key}.json")
            if os.path.exists(file_path):
                os.remove(file_path)
        except Exception as e:
            Log.e(TAG, f"Failed to delete cache for {key}", error=e)

    def clear(self):
        """
        Clear entire cache directory.
        """
        try:
            if os.path.exists(self.cache_root):
                shutil.rmtree(self.cache_root)
                os.makedirs(self.cache_root)
        except Exception as e:
            Log.e(TAG, "Failed to clear cache", error=e)

cache_manager = CacheManager()
