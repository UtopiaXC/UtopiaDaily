import multiprocessing
import time
import sys
import os
import signal
from src.utils.logger.logger import Log
from src.utils.i18n import i18n

TAG="APP"

current_dir = os.path.dirname(os.path.abspath(__file__))
project_root = os.path.dirname(current_dir)
sys.path.append(project_root)

from src.database.migration_manager import migration_manager
from src.scraper.scraper import run_scraper_service
from src.web.server import run_web_server

children_processes = []

def signal_handler(sig, frame):
    Log.w(TAG,f"Received signal: {sig}, stopping system...")
    cleanup_and_exit()


def cleanup_and_exit():
    for p in children_processes:
        if p.is_alive():
            Log.w(TAG, f"Stopping process: {p.name} (PID: {p.pid})...")
            p.terminate()
            p.join(timeout=2)
            if p.is_alive():
                Log.w(TAG, f"Killing process: {p.name}")
                p.kill()

    Log.w(TAG,"System Stopped")
    sys.exit(0)


def main():
    # 1. Run Database Migrations
    try:
        Log.i(TAG, "Initializing database and checking migrations...")
        migration_manager.run_migrations()
    except Exception as e:
        Log.e(TAG, "Critical error during database migration. Exiting.", error=e)
        sys.exit(1)

    # 2. Start Services
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    # Start Scraper Service
    scraper_process = multiprocessing.Process(
        target=run_scraper_service,
        name="ScraperService"
    )
    scraper_process.start()
    children_processes.append(scraper_process)
    Log.i(TAG,f"Scraper Process up: {scraper_process.pid}")

    # Start Web Server (Dashboard + Newspaper)
    web_process = multiprocessing.Process(
        target=run_web_server,
        kwargs={"host": "0.0.0.0", "port": 8000},
        name="WebServer"
    )
    web_process.start()
    children_processes.append(web_process)
    Log.i(TAG,f"Web Process up: {web_process.pid}")

    try:
        while True:
            time.sleep(2)
            for p in children_processes:
                if not p.is_alive():
                    Log.e(TAG,f"Process Died:{p.name}")
                    cleanup_and_exit()

    except Exception as e:
        Log.e(TAG,f"Main Process Exception: {e}")
        cleanup_and_exit()


if __name__ == "__main__":
    Log.i(TAG,f"Utopia Daily starting...")
    Log.i(TAG, f"I18n Test (EN): {i18n.t('config.system_version.desc')}")
    Log.i(TAG, f"I18n Test (CN): {i18n.t('config.system_version.desc', locale='zh_CN')}")
    main()
