import importlib.util
import os
import sys
import multiprocessing
from src.utils.logger.logger import Log

TAG = "ModuleManager"

class StandaloneContext:
    def __init__(self, module_name):
        self.module_name = module_name
        # TODO: Load from real database
        self.mock_db_config = {
            "channels": "https://t.me/s/tnews365, https://t.me/s/scitech_fans",
            "lookback_days": 3
        }

    def get_config(self, key):
        val = self.mock_db_config.get(key)
        # Log.d("Context", f"Read config: {key}={val}")
        return val

    def register_config(self, **kwargs):
        # Log.d("Context", f"Register config: {kwargs.get('key')}")
        pass

    def register_task(self, **kwargs):
        pass

    def save_result(self, data):
        Log.i("Context", f"[{self.module_name}] Save data requested (Mock Storage)")


def _subprocess_runner(module_path, module_name, task_params):
    TAG = "Runner"

    try:
        module_dir = os.path.dirname(module_path)
        if module_dir not in sys.path:
            sys.path.insert(0, module_dir)

        spec = importlib.util.spec_from_file_location(module_name, module_path)
        if spec is None:
            Log.e(TAG, f"Failed to load module spec: {module_path}")
            return

        module = importlib.util.module_from_spec(spec)
        sys.modules[module_name] = module
        spec.loader.exec_module(module)

        if hasattr(module, 'on_task_execute'):
            ctx = StandaloneContext(module_name)
            task_key = task_params.get("task_key", "fetch_news")

            Log.i(TAG, f"Executing {module_name} -> {task_key}")
            module.on_task_execute(ctx, task_key, task_params)

        elif hasattr(module, 'execute_task'):
            # Backward compatibility
            module.execute_task(task_params)
        else:
            Log.e(TAG, f"Module {module_name} missing entry point (on_task_execute)")

    except Exception as e:
        Log.e(TAG, "Worker crashed", error=e)


class ModuleManager:
    def __init__(self):
        current_dir = os.path.dirname(os.path.abspath(__file__))
        self.dirs = {
            "default": os.path.join(current_dir, "default"),
            "external": os.path.join(current_dir, "external")
        }
        for d in self.dirs.values():
            os.makedirs(d, exist_ok=True)
        self.loaded_modules = {}

    def scan_and_load(self):
        Log.i(TAG, "Scanning modules...")
        for source_type, path in self.dirs.items():
            if not os.path.exists(path): continue
            for folder_name in os.listdir(path):
                module_path = os.path.join(path, folder_name, "api.py")
                if os.path.isfile(module_path):
                    self.loaded_modules[folder_name] = {
                        "path": module_path,
                        "source": source_type
                    }
                    Log.i(TAG, f"Module found: {folder_name} ({source_type})")

    def run_module_task(self, module_name, task_params):
        if module_name not in self.loaded_modules:
            Log.w(TAG, f"Module {module_name} not found")
            return

        module_info = self.loaded_modules[module_name]

        p = multiprocessing.Process(
            target=_subprocess_runner,
            args=(module_info["path"], module_name, task_params)
        )
        p.daemon = True
        p.start()
        p.join()