import time
import os
from src.scraper.modules.module_manager import ModuleManager
from src.utils.logger.logger import Log

TAG="SCRAPER SERVICE"


def run_scraper_service():
    pid = os.getpid()
    Log.i(TAG,f"Process started (PID: {pid})")
    manager = ModuleManager()
    manager.scan_and_load()
    Log.i(TAG,"Inited, starting loop...")
    try:
        while True:
            Log.i(TAG,"Simulate task triggered")
            manager.run_module_task("telegram_channel", {"target": "news"})

            time.sleep(50)

    except KeyboardInterrupt:
        Log.w(TAG,"Interrupted, stopping service...")