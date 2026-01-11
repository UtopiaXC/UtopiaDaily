import multiprocessing
import time
import sys
import os
import signal
from src.utils.logger.logger import Log

TAG="APP"

current_dir = os.path.dirname(os.path.abspath(__file__))
project_root = os.path.dirname(current_dir)
sys.path.append(project_root)

from src.scraper.scraper import run_scraper_service
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
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    scraper_process = multiprocessing.Process(
        target=run_scraper_service,
        name="ScraperService"
    )

    scraper_process.start()
    children_processes.append(scraper_process)
    Log.i(TAG,f"Process up: {scraper_process.pid}")

    # TODO: Other Services

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
    main()