import json
import src.scraper.modules.default.bilibili_ranking_and_update.bilibili_scraper as scraper
from src.utils.logger.logger import Log
from datetime import datetime

TAG = "BILIBILI_MODULE_SERVICE"

uploader_id_conf = {
    "key": "uploader_id",
    "description": "module.bilibili_ranking_and_update.config.uploader_id.desc",
    "value": [],
    "force_init": False,
    "hint": "module.bilibili_ranking_and_update.config.uploader_id.hint",
    "regular": ".*",
    "value_type": "array"
}

ranking_limit_conf = {
    "key": "ranking_limit",
    "description": "Ranking Limit (Not used yet, hardcoded to 20)",
    "value": 20,
    "force_init": False,
    "hint": "Limit",
    "regular": "\\d+",
    "value_type": "int"
}

task_fetch_ranking = {
    "key": "fetch_ranking",
    "name": "module.bilibili_ranking_and_update.task.fetch_ranking.name",
    "description": "module.bilibili_ranking_and_update.task.fetch_ranking.desc",
    "force_init": False
}

task_fetch_update = {
    "key": "fetch_update",
    "name": "module.bilibili_ranking_and_update.task.fetch_update.name",
    "description": "module.bilibili_ranking_and_update.task.fetch_update.desc",
    "force_init": False
}


def get_init_configs():
    Log.d(TAG, "Configs loaded")
    return [uploader_id_conf, ranking_limit_conf]


def get_init_schedule_tasks():
    Log.d(TAG, "Tasks loaded")
    return [task_fetch_ranking, task_fetch_update]


def test_configuration(module, config):
    return True, "success"

def execute_schedule_task(module, cron: str, task_key: str):
    if task_key == "fetch_ranking":
        Log.i(TAG, "Fetching Bilibili Ranking...")
        items = scraper.get_ranking()
        for item in items:
            # Fingerprint: bvid
            fingerprint = item["metadata"]["bvid"]
            if not module.is_fingerprint_stored(fingerprint):
                module.save_structured_results(item, fingerprint=fingerprint)
        Log.i(TAG, f"Fetched {len(items)} ranking items")
        return True
        
    elif task_key == "fetch_update":
        Log.i(TAG, "Fetching Bilibili Updates...")
        uids = module.get_module_config("uploader_id")
        if not uids:
            Log.w(TAG, "No UIDs configured")
            return False
            
        if isinstance(uids, str):
            uids = [u.strip() for u in uids.split('\n') if u.strip()]
            
        total_count = 0
        for uid in uids:
            items = scraper.get_uploader_videos(uid)
            for item in items:
                fingerprint = item["metadata"]["bvid"]
                if not module.is_fingerprint_stored(fingerprint):
                    module.save_structured_results(item, fingerprint=fingerprint)
            total_count += len(items)
        Log.i(TAG, f"Fetched {total_count} updates")
        return True
        
    return False
