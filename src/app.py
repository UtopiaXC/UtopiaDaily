import multiprocessing
import time
import sys
import os
import signal
from src.utils.logger.logger import Log
from src.utils.env_manage.env_manager import EnvManager

TAG="APP"

current_dir = os.path.dirname(os.path.abspath(__file__))
project_root = os.path.dirname(current_dir)
sys.path.append(project_root)

EnvManager.init_env()

from src.database.migration_manager import migration_manager
from src.database.connection import system_session_scope
from src.database.models import SystemConfig
from src.scraper.scraper import run_scraper_service
from src.web.server import run_web_server
from src.utils.event import EventManager

children_processes = []

def signal_handler(sig, frame):
    Log.w(TAG,f"Received signal: {sig}, stopping system...")
    EventManager.record(
        level=EventManager.LEVEL_WARNING,
        category=EventManager.CATEGORY_SYSTEM,
        event_type="system_stopping",
        summary=f"Received signal {sig}, stopping system...",
        is_resolved=True
    )
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
    try:
        Log.i(TAG, "Initializing database and checking migrations...")
        migration_manager.run_migrations()
    except Exception as e:
        # Try to record event before dying, though DB might be broken
        try:
            EventManager.record(
                level=EventManager.LEVEL_FATAL,
                category=EventManager.CATEGORY_SYSTEM,
                event_type="migration_failed",
                summary="Critical error during database migration",
                details={"error": str(e)},
                is_resolved=False
            )
        except:
            pass
        Log.fatal(TAG, "Critical error during database migration. Exiting.", error=e)
    try:
        with system_session_scope() as session:
            log_level_config = session.query(SystemConfig).filter_by(key="log_level").first()
            if log_level_config and log_level_config.value:
                Log.i(TAG, f"Initializing log level to: {log_level_config.value}")
                Log.set_level(log_level_config.value)
            else:
                Log.i(TAG, "No log level config found, using default (INFO)")
                Log.set_level("INFO")
    except Exception as e:
        Log.e(TAG, "Failed to initialize system configs", error=e)
        Log.set_level("INFO")
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    scraper_process = multiprocessing.Process(
        target=run_scraper_service,
        name="ScraperService"
    )
    scraper_process.start()
    children_processes.append(scraper_process)
    Log.i(TAG,f"Scraper Process up: {scraper_process.pid}")
    web_process = multiprocessing.Process(
        target=run_web_server,
        kwargs={"host": "0.0.0.0", "port": 8000},
        name="WebServer"
    )
    web_process.start()
    children_processes.append(web_process)
    Log.i(TAG,f"Web Process up: {web_process.pid}")

    EventManager.record(
        level=EventManager.LEVEL_NORMAL,
        category=EventManager.CATEGORY_SYSTEM,
        event_type="system_started",
        summary="Utopia Daily System Started",
        details={"scraper_pid": scraper_process.pid, "web_pid": web_process.pid}
    )

    try:
        while True:
            time.sleep(2)
            for p in children_processes:
                if not p.is_alive():
                    Log.e(TAG,f"Process Died:{p.name}")
                    EventManager.record(
                        level=EventManager.LEVEL_FATAL,
                        category=EventManager.CATEGORY_SYSTEM,
                        event_type="process_died",
                        summary=f"Process Died: {p.name}",
                        is_resolved=False
                    )
                    cleanup_and_exit()

    except Exception as e:
        Log.e(TAG,f"Main Process Exception: {e}")
        EventManager.record(
            level=EventManager.LEVEL_FATAL,
            category=EventManager.CATEGORY_SYSTEM,
            event_type="main_process_exception",
            summary="Main Process Exception",
            details={"error": str(e)},
            is_resolved=False
        )
        cleanup_and_exit()


if __name__ == "__main__":
    Log.setup_global_exception_handler()
    Log.i(TAG,f"Utopia Daily starting...")
    main()
