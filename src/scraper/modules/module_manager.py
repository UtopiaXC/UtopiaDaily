import importlib.util
import os
import sys
import subprocess
from typing import Dict, Any, Tuple

from src.utils.logger.logger import Log
from src.utils.i18n import i18n
from src.database.connection import system_session_scope
from src.database.models import ScraperModule, ScraperModuleConfig, ScraperModuleTask
from src.utils.event import EventManager
from src.utils.cache_manage import cache_manager

TAG = "MODULE_MANAGER"

class ModuleContext:
    def __init__(self, module_id: str, manager: 'ModuleManager'):
        self.module_id = module_id
        self._manager = manager

    def set_module_config(self, key, description, value, value_type, options, force_init, hint, regular):
        self._manager.db_set_config(self.module_id, key, description, value, value_type, options, force_init, hint, regular)

    def get_module_config(self, key):
        return self._manager.db_get_config(self.module_id, key)

    def drop_module_config(self, key):
        self._manager.db_drop_config(self.module_id, key)

    def set_module_schedule_task(self, key, description, name="", force_init=False):
        self._manager.db_set_task(self.module_id, key, description, name=name, force_init=force_init)

    def get_module_schedule_task(self, key):
        return self._manager.db_get_task(self.module_id, key)

    def drop_module_schedule_task(self, key):
        self._manager.db_drop_task(self.module_id, key)

    def mark_message_tag(self, message):
        # TODO: Call actual AI tagging service
        return [{"tag": "news", "confidence": 0.9}]

    def save_structured_results(self, value, fingerprint=""):
        Log.i(TAG, f"[{self.module_id}] Data saved: {value.get('title', 'No Title')}")
        # TODO: Save to actual database
        return {"status": "success"}
    
    def install_requirements(self, requirements_file: str):
        return self._manager.install_module_requirements(self.module_id, requirements_file)


class ModuleManager:
    _instance = None
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(ModuleManager, cls).__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self):
        if self._initialized:
            return
            
        current_dir = os.path.dirname(os.path.abspath(__file__))
        self.dirs = {
            "default": os.path.join(current_dir, "default"),
            "external": os.path.join(current_dir, "external")
        }
        
        self._logged_conflicts = set()
        self._modules_cache = {}
        
        self._load_modules_cache()
        
        self._initialized = True

    def _load_modules_cache(self):
        data = cache_manager.get("modules")
        if data:
            self._modules_cache = data
            Log.i(TAG, f"Loaded {len(self._modules_cache)} modules from cache")

    def _save_modules_cache(self):
        cache_manager.set("modules", self._modules_cache)

    def db_set_config(self, module_id, key, description, value, value_type, options, force_init, hint, regular):
        with system_session_scope() as session:
            # Get module internal ID
            module = session.query(ScraperModule).filter_by(module_id=module_id).first()
            if not module:
                Log.e(TAG, f"Module {module_id} not found in DB")
                return

            config = session.query(ScraperModuleConfig).filter_by(module_id=module.id, config_key=key).first()
            
            if config:
                if force_init:
                    config.description = description
                    config.type = value_type
                    config.options = options
                    config.value = value
                    config.hint = hint
                    config.regex = regular
                    config.source = "custom"
                    Log.i(TAG, f"[{module_id}] Config '{key}' reset to default.")
                else:
                    config.description = description
                    config.type = value_type
                    config.options = options
                    config.hint = hint
                    config.regex = regular
                    if config.value is None:
                        config.value = value
            else:
                new_config = ScraperModuleConfig(
                    module_id=module.id,
                    config_key=key,
                    description=description,
                    type=value_type,
                    options=options,
                    value=value,
                    hint=hint,
                    regex=regular,
                    source="custom",
                    is_override=False
                )
                session.add(new_config)
                Log.i(TAG, f"[{module_id}] Config '{key}' initialized.")

    def db_get_config(self, module_id, key):
        with system_session_scope() as session:
            module = session.query(ScraperModule).filter_by(module_id=module_id).first()
            if not module:
                return None
            config = session.query(ScraperModuleConfig).filter_by(module_id=module.id, config_key=key).first()
            if config:
                return config.value
            return None

    def db_drop_config(self, module_id, key):
        with system_session_scope() as session:
            module = session.query(ScraperModule).filter_by(module_id=module_id).first()
            if not module:
                return
            config = session.query(ScraperModuleConfig).filter_by(module_id=module.id, config_key=key).first()
            if config:
                session.delete(config)
                Log.i(TAG, f"[{module_id}] Config '{key}' deleted.")

    def db_set_task(self, module_id, key, description, name, force_init):
        with system_session_scope() as session:
            module = session.query(ScraperModule).filter_by(module_id=module_id).first()
            if not module:
                Log.e(TAG, f"Module {module_id} not found in DB")
                return

            task = session.query(ScraperModuleTask).filter_by(module_id=module.id, task_key=key).first()
            
            if task:
                if force_init:
                    task.description = description
                    task.name = name
                    Log.i(TAG, f"[{module_id}] Task '{key}' reset.")
                else:
                    task.description = description
                    if not task.name:
                        task.name = name
            else:
                new_task = ScraperModuleTask(
                    module_id=module.id,
                    task_key=key,
                    name=name,
                    description=description
                )
                session.add(new_task)
                Log.i(TAG, f"[{module_id}] Task '{key}' initialized.")

    def db_get_task(self, module_id, key):
        with system_session_scope() as session:
            module = session.query(ScraperModule).filter_by(module_id=module_id).first()
            if not module:
                return None
            task = session.query(ScraperModuleTask).filter_by(module_id=module.id, task_key=key).first()
            if task:
                return {
                    "key": task.task_key,
                    "name": task.name,
                    "description": task.description
                }
            return None

    def db_drop_task(self, module_id, key):
        with system_session_scope() as session:
            module = session.query(ScraperModule).filter_by(module_id=module_id).first()
            if not module:
                return
            task = session.query(ScraperModuleTask).filter_by(module_id=module.id, task_key=key).first()
            if task:
                session.delete(task)
                Log.i(TAG, f"[{module_id}] Task '{key}' deleted.")

    def db_get_all_tasks(self, module_id):
        with system_session_scope() as session:
            module = session.query(ScraperModule).filter_by(module_id=module_id).first()
            if not module:
                return {}
            tasks = session.query(ScraperModuleTask).filter_by(module_id=module.id).all()
            return {t.task_key: {"name": t.name, "description": t.description} for t in tasks}
    
    def is_module_enabled(self, module_id):
        with system_session_scope() as session:
            module = session.query(ScraperModule).filter_by(module_id=module_id).first()
            return module.is_enable if module else False

    def scan_modules(self, force_reload=False):
        if self._modules_cache and not force_reload:
            return self._modules_cache

        found_modules = {}
        scan_order = ["default", "external"]
        
        for source_type in scan_order:
            path = self.dirs.get(source_type)
            if not path or not os.path.exists(path): continue
            
            for folder_name in os.listdir(path):
                module_path = os.path.join(path, folder_name, "controller.py")
                if os.path.isfile(module_path):
                    temp_id = folder_name
                    meta = self._read_module_meta(module_path, temp_id)

                    if not meta or "id" not in meta or "name" not in meta or "version" not in meta:
                        Log.w(TAG, f"Skipping invalid module at {module_path}: Missing required meta fields (id, name, version)")
                        continue
                    
                    module_id = meta["id"]

                    if module_id in found_modules:
                        Log.w(TAG, f"Duplicate module ID '{module_id}' found in {source_type}. Ignoring {module_path}. Keeping version from {found_modules[module_id]['source']}.")

                        if module_id not in self._logged_conflicts:
                            EventManager.record(
                                level=EventManager.LEVEL_WARNING,
                                category=EventManager.CATEGORY_MODULE,
                                event_type="module_conflict",
                                summary=f"Duplicate module ID detected: {module_id}",
                                details={
                                    "kept_path": found_modules[module_id]["path"],
                                    "ignored_path": module_path,
                                    "kept_source": found_modules[module_id]["source"],
                                    "ignored_source": source_type
                                },
                                source_id=module_id,
                                is_resolved=False
                            )
                            self._logged_conflicts.add(module_id)
                        continue

                    found_modules[module_id] = {
                        "path": module_path,
                        "source": source_type,
                        "meta": meta
                    }
        
        self._modules_cache = found_modules
        self._save_modules_cache()
        return found_modules

    def _read_module_meta(self, path, module_id):
        try:
            spec = importlib.util.spec_from_file_location(f"meta_{module_id}", path)
            if spec and spec.loader:
                module = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(module)
                return getattr(module, "MODULE_META", {})
        except Exception as e:
            Log.w(TAG, f"Failed to read meta for {module_id}: {e}")
        return {}

    def _load_all_module_locales(self):
        Log.i(TAG, "Loading all module locales...")
        scanned_modules = self.scan_modules()
        for module_id, info in scanned_modules.items():
            module_dir = os.path.dirname(info["path"])
            locales_dir = os.path.join(module_dir, "locales")
            if os.path.exists(locales_dir):
                i18n.load_module_locales(locales_dir)

    def reload_modules(self):
        Log.i(TAG, "Reloading modules...")
        scanned_modules = self.scan_modules(force_reload=True)

        locale_dirs = []
        for info in scanned_modules.values():
            module_dir = os.path.dirname(info["path"])
            locale_dirs.append(os.path.join(module_dir, "locales"))
        i18n.compile_locales(locale_dirs)

        with system_session_scope() as session:
            existing_modules = {m.module_id: m for m in session.query(ScraperModule).all()}

            for module_id, info in scanned_modules.items():
                meta = info["meta"]
                source = info["source"]
                
                if module_id in existing_modules:
                    module = existing_modules[module_id]
                    module.name = meta["name"]
                    module.description = meta.get("description")
                    module.version = meta["version"]
                    module.author = meta.get("author")
                    module.meta = meta
                    module.source = source
                    
                    if module.is_deleted:
                        Log.i(TAG, f"Restoring deleted module: {module_id}")
                        module.is_deleted = False
                        module.is_enable = False
                    
                    Log.i(TAG, f"Updated module: {module_id}")
                else:
                    existing = session.query(ScraperModule).filter_by(module_id=module_id).first()
                    if existing:
                        Log.e(TAG, f"Duplicate module ID detected during sync: {module_id}. Skipping {info['path']}")
                        continue

                    new_module = ScraperModule(
                        module_id=module_id,
                        name=meta["name"],
                        description=meta.get("description"),
                        version=meta["version"],
                        author=meta.get("author"),
                        meta=meta,
                        source=source,
                        is_enable=False
                    )
                    session.add(new_module)
                    Log.i(TAG, f"Registered new module: {module_id}")

            for module_id, module in existing_modules.items():
                if module_id not in scanned_modules and not module.is_deleted:
                    Log.w(TAG, f"Module {module_id} not found on disk. Marking as deleted.")
                    module.is_deleted = True
                    module.is_enable = False
                    self.disable_module(module_id)

    def get_module_info(self, module_id):
        return self._modules_cache.get(module_id)

    def enable_module(self, module_id):
        Log.i(TAG, f"Enabling module: {module_id}")
        module_info = self.get_module_info(module_id)
        if not module_info:
            Log.e(TAG, f"Module {module_id} not found in cache")
            return False
        
        success, message = self.test_module(module_id)
        if not success:
            Log.w(TAG, f"Module {module_id} failed pre-enable test: {message}")
            raise Exception(f"Module test failed: {message}")

        try:
            spec = importlib.util.spec_from_file_location(f"setup_{module_id}", module_info["path"])
            module_lib = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(module_lib)
            
            ctx = ModuleContext(module_id, self)
            instance = module_lib.create_module(ctx)
            
            if instance.enable_module():
                with system_session_scope() as session:
                    module = session.query(ScraperModule).filter_by(module_id=module_id).first()
                    if module:
                        module.is_enable = True

                return True
            else:
                Log.e(TAG, f"Module {module_id} failed to enable")
                return False
        except Exception as e:
            Log.e(TAG, f"Error enabling module {module_id}", error=e)
            return False

    def install_module_requirements(self, module_id, requirements_file):
        module_info = self.get_module_info(module_id)
        if not module_info:
            Log.e(TAG, f"[{module_id}] Module not found for installing requirements")
            return False

        module_path = module_info["path"]
        module_dir = os.path.dirname(module_path)
        req_path = os.path.join(module_dir, requirements_file)
        libs_dir = os.path.join(module_dir, "libs")

        if not os.path.exists(req_path):
            Log.w(TAG, f"[{module_id}] Requirements file not found: {req_path}")
            return False

        Log.i(TAG, f"[{module_id}] Installing requirements from {requirements_file} to {libs_dir}...")
        try:
            subprocess.check_call(
                [sys.executable, "-m", "pip", "install", "-r", req_path, "-t", libs_dir],
                stdout=subprocess.DEVNULL,
                stderr=subprocess.PIPE
            )
            Log.i(TAG, f"[{module_id}] Dependencies installed successfully.")
            return True
        except subprocess.CalledProcessError as e:
            Log.e(TAG, f"[{module_id}] Failed to install dependencies", error=e)
            return False

    def start_module(self, module_id):
        pass

    def disable_module(self, module_id):
        with system_session_scope() as session:
            module = session.query(ScraperModule).filter_by(module_id=module_id).first()
            if module:
                module.is_enable = False
            
        Log.i(TAG, f"Module {module_id} disabled")

    def notify_module_disabled(self, module_id, reason):
        Log.e(TAG, f"SYSTEM ALERT: Module {module_id} disabled due to: {reason}")
        
        EventManager.record(
            level=EventManager.LEVEL_CRITICAL,
            category=EventManager.CATEGORY_MODULE,
            event_type="module_crashed",
            summary=f"Module {module_id} disabled due to crash",
            details={"reason": reason},
            source_id=module_id,
            is_resolved=False
        )

        self.disable_module(module_id)

    def test_module(self, module_id):
        Log.i(TAG, f"Testing module: {module_id}")
        module_info = self.get_module_info(module_id)
        if not module_info:
            return False, "Module not found"
            
        try:
            module_path = module_info["path"]
            module_dir = os.path.dirname(module_path)
            if module_dir not in sys.path:
                sys.path.insert(0, module_dir)
            libs_dir = os.path.join(module_dir, "libs")
            if os.path.exists(libs_dir) and libs_dir not in sys.path:
                sys.path.insert(0, libs_dir)
            
            spec = importlib.util.spec_from_file_location(f"test_{module_id}", module_path)
            if spec is None:
                return False, f"Could not load spec for {module_path}"
            module_lib = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(module_lib)
            if not hasattr(module_lib, 'create_module'):
                return False, "Module missing 'create_module' factory function"
            ctx = ModuleContext(module_id, self)
            instance = module_lib.create_module(ctx)
            if hasattr(instance, 'test_module'):
                return instance.test_module()
            else:
                return False, "Module does not support testing"
        except Exception as e:
            Log.e(TAG, f"Error testing module {module_id}", error=e)
            return False, str(e)

    def test_module_config(self, module_id: str, config: Dict[str, Any]) -> Tuple[bool, str]:
        module_info = self.get_module_info(module_id)
        if not module_info:
            return False, "Module not found"
            
        try:
            module_path = module_info["path"]
            module_dir = os.path.dirname(module_path)
            if module_dir not in sys.path:
                sys.path.insert(0, module_dir)
            libs_dir = os.path.join(module_dir, "libs")
            if os.path.exists(libs_dir) and libs_dir not in sys.path:
                sys.path.insert(0, libs_dir)

            spec = importlib.util.spec_from_file_location(f"test_cfg_{module_id}", module_path)
            if spec is None:
                return False, f"Could not load spec for {module_path}"
            module_lib = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(module_lib)
            if not hasattr(module_lib, 'create_module'):
                return False, "Module missing 'create_module' factory function"
            ctx = ModuleContext(module_id, self)
            instance = module_lib.create_module(ctx)
            if hasattr(instance, 'test_config'):
                return instance.test_config(config)
            else:
                return True, "Module does not support config testing (assumed valid)"
        except Exception as e:
            return False, f"Config test error: {str(e)}"
