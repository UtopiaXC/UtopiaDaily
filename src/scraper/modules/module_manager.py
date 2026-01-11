import importlib.util
import os
import sys
import threading
import time
import traceback
import json
from datetime import datetime
from typing import Dict, Any, Optional, List
from croniter import croniter

from src.utils.logger.logger import Log

TAG = "MODULE_MANAGER"
DB_FILE = "modules_db.json"

class ModuleContext:
    def __init__(self, module_id: str, manager: 'ModuleManager'):
        self.module_id = module_id
        self._manager = manager

    def set_module_config(self, key, description, value, force_init, hint, regular):
        self._manager.db_set_config(self.module_id, key, description, value, force_init, hint, regular)

    def get_module_config(self, key):
        return self._manager.db_get_config(self.module_id, key)

    def drop_module_config(self, key):
        self._manager.db_drop_config(self.module_id, key)

    def set_module_schedule_task(self, key, description, cron, force_init, hint="", regular=""):
        self._manager.db_set_task(self.module_id, key, description, cron, force_init, hint, regular)

    def get_module_schedule_task(self, key):
        return self._manager.db_get_task(self.module_id, key)

    def mark_message_tag(self, message):
        # TODO: Call actual AI tagging service
        return [{"tag": "news", "confidence": 0.9}]

    def save_structured_results(self, value):
        Log.i(TAG, f"[{self.module_id}] Data saved: {value.get('title', 'No Title')}")
        # TODO: Save to actual database
        return {"status": "success"}


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
        self.task_states = {} # { task_key: next_run_datetime }
        self.task_crons = {}  # { task_key: cron_expression_string }

    def _load_module(self):
        try:
            module_dir = os.path.dirname(self.module_path)
            if module_dir not in sys.path:
                sys.path.insert(0, module_dir)

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

    # ==========================================
    # Database Persistence (Mock)
    # ==========================================
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

    def db_set_config(self, module_id, key, description, value, force_init, hint, regular):
        self._ensure_db_entry(module_id)
        config_store = self.db[module_id]["config"]
        if key not in config_store or force_init:
            config_store[key] = {
                "value": value,
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

    def db_set_task(self, module_id, key, description, cron, force_init, hint, regular):
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

    def db_get_all_tasks(self, module_id):
        self._ensure_db_entry(module_id)
        return self.db[module_id]["tasks"]
    
    def is_module_enabled(self, module_id):
        return self.db.get(module_id, {}).get("enabled", False)

    # ==========================================
    # Management Methods
    # ==========================================

    def scan_modules(self):
        found_modules = {}
        for source_type, path in self.dirs.items():
            if not os.path.exists(path): continue
            for folder_name in os.listdir(path):
                module_path = os.path.join(path, folder_name, "api.py")
                if os.path.isfile(module_path):
                    module_id = folder_name
                    meta = self._read_module_meta(module_path, module_id)
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

    def enable_module(self, module_id):
        Log.i(TAG, f"Enabling module: {module_id}")
        modules = self.scan_modules()
        if module_id not in modules:
            Log.e(TAG, f"Module {module_id} not found")
            return False
        try:
            module_info = modules[module_id]
            spec = importlib.util.spec_from_file_location(f"setup_{module_id}", module_info["path"])
            module_lib = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(module_lib)
            
            ctx = ModuleContext(module_id, self)
            instance = module_lib.create_module(ctx)
            
            if instance.enable_module():
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

        if module_id in self.db:
            self.db[module_id]["enabled"] = False
            self._save_db()
            
        Log.i(TAG, f"Module {module_id} disabled")

    def notify_module_disabled(self, module_id, reason):
        Log.e(TAG, f"SYSTEM ALERT: Module {module_id} disabled due to: {reason}")
        self.disable_module(module_id)
