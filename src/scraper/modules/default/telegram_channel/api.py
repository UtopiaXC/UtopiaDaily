from src.utils.logger.logger import Log
from telegram_channel_scraper import TelegramScraper

TAG = "TelegramAPI"


def on_register(context):
    context.register_config(
        key="channels",
        default_value="https://t.me/s/tnews365,https://t.me/s/scitech_fans",
        description="Telegram channel list (comma separated)",
        hint="Public channels only",
        regex=".*"
    )
    context.register_config(
        key="lookback_days",
        default_value="3",
        description="Lookback days",
        hint="Integer",
        regex="^\d+$"
    )
    context.register_task(
        task_key="fetch_news",
        default_cron="0 8 * * *",
        description="Fetch Telegram daily news"
    )


def on_enable(context):
    channels = context.get_config("channels")
    if not channels:
        Log.w(TAG, "Missing channel configuration")
        return False, "Missing channel configuration"

    Log.i(TAG, "Module enabled")
    return True, ""


def on_task_execute(context, task_key, task_params):
    if task_key == "fetch_news":
        Log.i(TAG, f"Executing task: {task_key}")

        channels_str = context.get_config("channels")
        channels = [url.strip() for url in channels_str.split(',') if url.strip()]

        config = {
            "lookback_days": context.get_config("lookback_days"),
            "enable_ai": True
        }

        try:
            scraper = TelegramScraper(config)

            def save_callback(item_data):
                Log.i(TAG, f"New item: {item_data['source']} - {item_data['title']}")

                # TODO: Implement database storage
                # context.save_result({
                #     "unique_hash": context.generate_hash(item_data['raw_data']['link']),
                #     "content": item_data,
                #     "tags": []
                # })

            scraper.run(channels, save_callback)
            Log.i(TAG, "Task completed successfully")

        except Exception as e:
            Log.e(TAG, f"Task execution failed: {e}")