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
    
    Log.i(TAG,"Inited, starting loop...")
    try:
        while True:
            time.sleep(10)
    except KeyboardInterrupt:
        Log.w(TAG,"Interrupted, stopping service...")
