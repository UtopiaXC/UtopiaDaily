import importlib.util
import os
import sys
import threading
import time
import traceback
import json
import subprocess
from datetime import datetime
from typing import Dict, Any, Optional, List
from croniter import croniter

from src.utils.logger.logger import Log
from src.utils.i18n import i18n
from src.database.connection import system_session_scope
from src.database.models import ScraperModule

TAG = "MODULE_MANAGER"
DB_FILE = "modules_db.json"

class ModuleContext:
    def __init__(self, module_id: str, manager: 'ModuleManager'):
        self.module_id = module_id
        self._manager = manager

    def set_module_config(self, key, description, value, value_type, force_init, hint, regular):
        self._manager.db_set_config(self.module_id, key, description, value, value_type, force_init, hint, regular)

    def get_module_config(self, key):
        return self._manager.db_get_config(self.module_id, key)

    def drop_module_config(self, key):
        self._manager.db_drop_config(self.module_id, key)

    def set_module_schedule_task(self, key, description, cron, force_init):
        self._manager.db_set_task(self.module_id, key, description, cron, force_init)

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
        """
        Install requirements from a file into a local 'libs' directory.
        This allows modules to have their own dependencies without polluting the global environment.
        """
        return self._manager.install_module_requirements(self.module_id, requirements_file)


class ModuleRunner(threading.Thread):
    def __init__(self, module_id: str, module_path: str, manager: 'ModuleManager'):
        super().__init__(name=f"Runner-{module_id}")
        self.module_id = module_id
        self.module_path = module_path
        self.manager = manager
        self.daemon = True
        
        self.running = False
        self.module_instance = None
        self.retry_count = 0
        self.max_retries = 3
        self.is_crashed = False
        
        # Scheduling state
        self.task_states = {}
        self.task_crons = {}

    def _load_module(self):
        try:
            module_dir = os.path.dirname(self.module_path)
            libs_dir = os.path.join(module_dir, "libs")
            if os.path.exists(libs_dir) and libs_dir not in sys.path:
                sys.path.insert(0, libs_dir)
                Log.i(TAG, f"[{self.module_id}] Added local libs to path: {libs_dir}")
            if module_dir not in sys.path:
                sys.path.insert(0, module_dir)
            
            # Locales are now loaded globally by ModuleManager
            
            spec = importlib.util.spec_from_file_location(f"module_{self.module_id}", self.module_path)
            if spec is None:
                raise ImportError(f"Could not load spec for {self.module_path}")
            
            module_lib = importlib.util.module_from_spec(spec)
            sys.modules[f"module_{self.module_id}"] = module_lib
            spec.loader.exec_module(module_lib)

            if not hasattr(module_lib, 'create_module'):
                raise ImportError("Module missing 'create_module' factory function")

            context = ModuleContext(self.module_id, self.manager)
            self.module_instance = module_lib.create_module(context)
            return True
        except Exception as e:
            Log.e(TAG, f"[{self.module_id}] Load failed", error=e)
            return False

    def run(self):
        Log.i(TAG, f"[{self.module_id}] Runner started")
        self.running = True
        if not self._load_module():
            self._handle_crash("Initialization failed")
            return
        while self.running:
            try:
                earliest_next_run = self._check_schedules()
                sleep_time = 5.0
                if earliest_next_run:
                    now = datetime.now()
                    delta = (earliest_next_run - now).total_seconds()
                    sleep_time = max(0.1, min(delta, 5.0))
                time.sleep(sleep_time)
            except Exception as e:
                Log.e(TAG, f"[{self.module_id}] Loop exception", error=e)
                if not self._handle_crash(str(e)):
                    break
        
        Log.i(TAG, f"[{self.module_id}] Runner stopped")

    def _check_schedules(self) -> Optional[datetime]:
        tasks = self.manager.db_get_all_tasks(self.module_id)
        now = datetime.now()
        next_runs: List[datetime] = []
        for task_key, task_info in tasks.items():
            cron_expr = task_info.get('cron')
            if not cron_expr: continue
            if task_key in self.task_crons and self.task_crons[task_key] != cron_expr:
                Log.i(TAG, f"[{self.module_id}] Cron changed for {task_key}, rescheduling.")
                if task_key in self.task_states:
                    del self.task_states[task_key]
            self.task_crons[task_key] = cron_expr
            if task_key not in self.task_states:
                try:
                    self.task_states[task_key] = croniter(cron_expr, now).get_next(datetime)
                    Log.i(TAG, f"[{self.module_id}] Scheduled {task_key} for {self.task_states[task_key]}")
                except Exception as e:
                    Log.w(TAG, f"[{self.module_id}] Invalid cron '{cron_expr}' for {task_key}: {e}")
                    self.task_states[task_key] = None
            
            next_run = self.task_states.get(task_key)
            if next_run and now >= next_run:
                self._execute_task(cron_expr, task_key, now)
                try:
                    self.task_states[task_key] = croniter(cron_expr, now).get_next(datetime)
                    Log.i(TAG, f"[{self.module_id}] Next {task_key} at {self.task_states[task_key]}")
                except Exception as e:
                    Log.e(TAG, f"[{self.module_id}] Failed to schedule next run for {task_key}", error=e)
                    self.task_states[task_key] = None
            if self.task_states.get(task_key):
                next_runs.append(self.task_states[task_key])
        
        if next_runs:
            return min(next_runs)
        return None

    def _execute_task(self, cron, task_key, now):
        Log.i(TAG, f"[{self.module_id}] Executing task: {task_key}")
        try:
            self.module_instance.execute_schedule_task(cron, task_key, now)
        except Exception as e:
            Log.e(TAG, f"[{self.module_id}] Task {task_key} failed", error=e)

    def _handle_crash(self, reason):
        self.retry_count += 1
        Log.e(TAG, f"[{self.module_id}] CRASHED ({self.retry_count}/{self.max_retries}): {reason}")
        Log.e(TAG, traceback.format_exc())

        if self.retry_count > self.max_retries:
            Log.e(TAG, f"[{self.module_id}] Max retries exceeded. Disabling module.")
            self.is_crashed = True
            self.running = False
            self.manager.notify_module_disabled(self.module_id, reason)
            return False
        
        time.sleep(2)
        if self._load_module():
             return True
        return False

    def stop(self):
        self.running = False


class ModuleManager:
    def __init__(self):
        current_dir = os.path.dirname(os.path.abspath(__file__))
        self.dirs = {
            "default": os.path.join(current_dir, "default"),
            "external": os.path.join(current_dir, "external")
        }
        self.db_path = os.path.join(current_dir, DB_FILE)
        
        self.db = {} 
        self.runners: Dict[str, ModuleRunner] = {}
        
        self._load_db()
        self.reload_modules()
        self._load_all_module_locales()

    # ==========================================
    # Database Persistence (Mock)
    # ==========================================
    # TODO: replace real db
    def _load_db(self):
        if os.path.exists(self.db_path):
            try:
                with open(self.db_path, 'r', encoding='utf-8') as f:
                    self.db = json.load(f)
            except Exception as e:
                Log.e(TAG, "Failed to load DB", error=e)

    def _save_db(self):
        try:
            with open(self.db_path, 'w', encoding='utf-8') as f:
                json.dump(self.db, f, indent=2)
        except Exception as e:
            Log.e(TAG, "Failed to save DB", error=e)

    def _ensure_db_entry(self, module_id):
        if module_id not in self.db:
            self.db[module_id] = {"config": {}, "tasks": {}, "enabled": False}

    def db_set_config(self, module_id, key, description, value, value_type, force_init, hint, regular):
        self._ensure_db_entry(module_id)
        config_store = self.db[module_id]["config"]
        if key not in config_store or force_init:
            config_store[key] = {
                "value": value,
                "value_type": value_type,
                "description": description,
                "hint": hint,
                "regular": regular
            }
        self._save_db()

    def db_get_config(self, module_id, key):
        self._ensure_db_entry(module_id)
        cfg = self.db[module_id]["config"].get(key)
        return cfg["value"] if cfg else None

    def db_drop_config(self, module_id, key):
        if module_id in self.db and key in self.db[module_id]["config"]:
            del self.db[module_id]["config"][key]
            self._save_db()

    def db_set_task(self, module_id, key, description, cron, force_init):
        self._ensure_db_entry(module_id)
        task_store = self.db[module_id]["tasks"]
        if key not in task_store or force_init:
            task_store[key] = {
                "cron": cron,
                "description": description
            }
        self._save_db()

    def db_get_task(self, module_id, key):
        self._ensure_db_entry(module_id)
        task = self.db[module_id]["tasks"].get(key)
        return task["cron"] if task else None

    def db_drop_task(self, module_id, key):
        if module_id in self.db and key in self.db[module_id]["tasks"]:
            del self.db[module_id]["tasks"][key]
            self._save_db()

    def db_get_all_tasks(self, module_id):
        self._ensure_db_entry(module_id)
        return self.db[module_id]["tasks"]
    
    def is_module_enabled(self, module_id):
        with system_session_scope() as session:
            module = session.query(ScraperModule).filter_by(module_id=module_id).first()
            return module.is_enable if module else False

    def scan_modules(self):
        found_modules = {}
        # Define scan order: default first, then external
        scan_order = ["default", "external"]
        
        for source_type in scan_order:
            path = self.dirs.get(source_type)
            if not path or not os.path.exists(path): continue
            
            for folder_name in os.listdir(path):
                module_path = os.path.join(path, folder_name, "controller.py")
                if os.path.isfile(module_path):
                    # Use folder name as temporary ID until meta is read
                    temp_id = folder_name
                    meta = self._read_module_meta(module_path, temp_id)
                    
                    # Validate Meta
                    if not meta or "id" not in meta or "name" not in meta or "version" not in meta:
                        Log.w(TAG, f"Skipping invalid module at {module_path}: Missing required meta fields (id, name, version)")
                        continue
                    
                    module_id = meta["id"]
                    
                    # Check if module ID already exists (from previous source)
                    if module_id in found_modules:
                        Log.w(TAG, f"Duplicate module ID '{module_id}' found in {source_type}. Ignoring {module_path}. Keeping version from {found_modules[module_id]['source']}.")
                        continue

                    found_modules[module_id] = {
                        "path": module_path,
                        "source": source_type,
                        "meta": meta
                    }
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
        """
        Scans modules and updates the database.
        Handles duplicates by keeping the first loaded one.
        Marks missing modules as deleted.
        """
        Log.i(TAG, "Reloading modules...")
        scanned_modules = self.scan_modules()
        
        with system_session_scope() as session:
            # Get existing modules
            existing_modules = {m.module_id: m for m in session.query(ScraperModule).all()}
            
            # 1. Update or Create scanned modules
            for module_id, info in scanned_modules.items():
                meta = info["meta"]
                source = info["source"]
                
                if module_id in existing_modules:
                    # Update existing module
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
                        # Keep is_enable as False for safety when restoring
                        module.is_enable = False
                    
                    Log.i(TAG, f"Updated module: {module_id}")
                else:
                    # Create new module
                    # Check if module_id already exists in DB (handled by existing_modules check above, but double check for safety)
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
                        is_enable=False # Default to disabled
                    )
                    session.add(new_module)
                    Log.i(TAG, f"Registered new module: {module_id}")

            # 2. Mark missing modules as deleted
            for module_id, module in existing_modules.items():
                if module_id not in scanned_modules and not module.is_deleted:
                    Log.w(TAG, f"Module {module_id} not found on disk. Marking as deleted.")
                    module.is_deleted = True
                    module.is_enable = False
                    self.disable_module(module_id) # Stop if running
        
        # Reload locales after DB sync
        self._load_all_module_locales()

    def enable_module(self, module_id):
        Log.i(TAG, f"Enabling module: {module_id}")
        modules = self.scan_modules()
        if module_id not in modules:
            Log.e(TAG, f"Module {module_id} not found on disk")
            return False
        
        # Test module first
        success, message = self.test_module(module_id)
        if not success:
            Log.w(TAG, f"Module {module_id} failed pre-enable test: {message}")
            raise Exception(f"Module test failed: {message}")

        try:
            module_info = modules[module_id]
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
                        self._ensure_db_entry(module_id)
                        self.db[module_id]["enabled"] = True
                        self._save_db()
                
                self.start_module(module_id)
                return True
            else:
                Log.e(TAG, f"Module {module_id} failed to enable")
                return False
        except Exception as e:
            Log.e(TAG, f"Error enabling module {module_id}", error=e)
            return False

    def install_module_requirements(self, module_id, requirements_file):
        """
        Installs requirements for a specific module into its local 'libs' directory.
        """
        modules = self.scan_modules()
        if module_id not in modules:
            Log.e(TAG, f"[{module_id}] Module not found for installing requirements")
            return False

        module_path = modules[module_id]["path"]
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
        if not self.is_module_enabled(module_id):
            Log.w(TAG, f"Cannot start module {module_id}: Not enabled in DB")
            return False
        if module_id in self.runners and self.runners[module_id].is_alive():
            return True
        modules = self.scan_modules()
        if module_id not in modules:
            return False
        Log.i(TAG, f"Starting module: {module_id}")
        module_info = modules[module_id]
        runner = ModuleRunner(module_id, module_info["path"], self)
        self.runners[module_id] = runner
        runner.start()
        return True

    def disable_module(self, module_id):
        if module_id in self.runners:
            runner = self.runners[module_id]
            runner.stop()
            runner.join(timeout=5)
            del self.runners[module_id]

        with system_session_scope() as session:
            module = session.query(ScraperModule).filter_by(module_id=module_id).first()
            if module:
                module.is_enable = False

        if module_id in self.db:
            self.db[module_id]["enabled"] = False
            self._save_db()
            
        Log.i(TAG, f"Module {module_id} disabled")

    def notify_module_disabled(self, module_id, reason):
        Log.e(TAG, f"SYSTEM ALERT: Module {module_id} disabled due to: {reason}")
        self.disable_module(module_id)

    def test_module(self, module_id):
        """
        Test the availability of a module.
        Returns: (success: bool, message: str)
        """
        Log.i(TAG, f"Testing module: {module_id}")
        
        # Try to use running instance first
        if module_id in self.runners and self.runners[module_id].is_alive():
            runner = self.runners[module_id]
            if runner.module_instance:
                try:
                    if hasattr(runner.module_instance, 'test_module'):
                        return runner.module_instance.test_module()
                    else:
                        return False, "Module instance does not support testing"
                except Exception as e:
                    Log.e(TAG, f"Error testing running module {module_id}", error=e)
                    return False, str(e)

        modules = self.scan_modules()
        if module_id not in modules:
            return False, "Module not found"
        try:
            module_info = modules[module_id]
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
