import json

import src.scraper.modules.default.telegram_channel.telegram_channel_scraper as scraper
from src.utils.logger.logger import Log
from datetime import datetime

TAG = "TELEGRAM_CHANNEL_MODULE_SERVICE"

channel_conf = {
    "key": "channels",
    "description": "module.telegram_channel.config.channels.desc",
    "value": [],
    "force_init": False,
    "hint": "module.telegram_channel.config.channels.hint",
    "regular": ".*",
    "value_type": "array"
}
lookback_conf = {
    "key": "lookback_days",
    "description": "module.telegram_channel.config.lookback.desc",
    "value": 3,
    "force_init": False,
    "hint": "module.telegram_channel.config.lookback.hint",
    "regular": "^\\d+$",
    "value_type": "int"
}

timezone_conf = {
    "key": "timezone",
    "description": "module.telegram_channel.config.timezone.desc",
    "value": 0,
    "force_init": False,
    "hint": "module.telegram_channel.config.timezone.hint",
    "regular": "^\\d+$",
    "value_type": "int"
}

max_pages_conf = {
    "key": "max_pages",
    "description": "module.telegram_channel.config.max_pages.desc",
    "value": 5,
    "force_init": False,
    "hint": "module.telegram_channel.config.max_pages.hint",
    "regular": "^\\d+$",
    "value_type": "int"
}

task_fetch = {
    "key": "fetch_news",
    "name": "module.telegram_channel.task.fetch_news.name",
    "description": "module.telegram_channel.task.fetch_news.desc",
    "force_init": False
}


def get_init_configs():
    Log.d(TAG, "Configs loaded")
    return [channel_conf, lookback_conf, timezone_conf, max_pages_conf]


def get_init_schedule_tasks():
    Log.d(TAG, "Tasks loaded")
    return [task_fetch]


def test_telegram_access():
    try:
        test_url = "https://t.me/s/telegram"
        success, msg = scraper.test_connection(test_url)
        return success
    except Exception as e:
        Log.e(TAG, e)
        return False

def test_configuration(config):
    channels = config.get("channels", [])
    if not channels:
        return True, "No channels configured"
    
    if isinstance(channels, str):
        channels = [c.strip() for c in channels.split('\n') if c.strip()]
    
    failed_channels = []
    for channel in channels:
        channel = str(channel).strip()
        if not channel: continue
        if not channel.startswith("http"):
            if channel.startswith("t.me/"):
                channel = f"https://{channel}"
            else:
                channel = f"https://t.me/s/{channel}"
        success, msg = scraper.test_connection(channel)
        if not success:
            failed_channels.append(channel)
            
    if failed_channels:
        return False, f"Failed to connect to: {', '.join(failed_channels)}"
    
    return True, "All channels accessible"

def execute_schedule_task(module, cron: str, task_key: str, timestamp: datetime):
    if task_key == task_fetch.get("key"):
        Log.i(TAG, f"Executing task: {task_key}")
        channels = module.get_module_config("channels")
        if not channels:
            Log.w(TAG, "No channels configured")
            return False

        # Ensure channels is list
        if isinstance(channels, str):
             channels = [c.strip() for c in channels.split('\n') if c.strip()]

        for channel in channels:
            channel = str(channel).strip()
            if not channel: continue

            if not channel.startswith("http"):
                if channel.startswith("t.me/"):
                    channel = f"https://{channel}"
                else:
                    channel = f"https://t.me/s/{channel}"
            if "t.me/" in channel and "/s/" not in channel:
                channel = channel.replace("t.me/", "t.me/s/")

            messages = scraper.fetch_channel(channel,
                                             int(module.get_module_config(lookback_conf.get("key"))),
                                             int(module.get_module_config(timezone_conf.get("key"))),
                                             int(module.get_module_config(max_pages_conf.get("key"))),
                                             )
            try:
                for message in messages:
                    try:
                        tags = module.mark_message_tag(message['content'])
                        message['tags'] = tags
                    except Exception as e:
                        Log.w(TAG, f"Tagging failed: {e}")
                    module.save_structured_results(message, fingerprint=message['from_url'])
            except Exception as e:
                Log.e(TAG, e)
                return False
        Log.i(TAG, "Task completed successfully")
        return True
    return False
