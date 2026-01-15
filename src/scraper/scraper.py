import time
import os
from src.scraper.modules.module_manager import ModuleManager
from src.utils.logger.logger import Log

TAG="SCRAPER_SERVICE"


def run_scraper_service():
    pid = os.getpid()
    Log.i(TAG,f"Process started (PID: {pid})")
    manager = ModuleManager()
    available_modules = manager.scan_modules()
    Log.i(TAG, f"Scanned {len(available_modules)} modules:")
    for mod_id, info in available_modules.items():
        meta = info.get('meta', {})
        Log.i(TAG, f" - [{mod_id}] {meta.get('name', mod_id)}")
    # TODO: In production, this loop just checks DB status.
    if "telegram_channel" in available_modules:
        if not manager.is_module_enabled("telegram_channel"):
            Log.i(TAG, "DEMO: Simulating user enabling 'telegram_channel'...")
            manager.enable_module("telegram_channel")
        else:
            Log.i(TAG, "'telegram_channel' is already enabled. Starting...")
            manager.start_module("telegram_channel")
    for mod_id in available_modules:
        if mod_id != "telegram_channel" and manager.is_module_enabled(mod_id):
            manager.start_module(mod_id)
    Log.i(TAG,"Inited, starting loop...")
    try:
        while True:
            time.sleep(10)
    except KeyboardInterrupt:
        Log.w(TAG,"Interrupted, stopping service...")
        for mod_id in list(manager.runners.keys()):
            runner = manager.runners[mod_id]
            runner.stop()
