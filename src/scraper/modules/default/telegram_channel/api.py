from src.utils.logger.logger import Log
from src.scraper.modules.base_module import BaseModule
from src.scraper.modules.default.telegram_channel.telegram_channel_scraper import TelegramScraper

from datetime import datetime
from typing import Tuple

TAG = "TELEGRAM_CHANNEL_MODULE_API"

MODULE_META = {
    "name": "Telegram Channel Scraper",
    "description": "Scrapes public Telegram channels for news and updates.",
    "version": "1.0.0",
    "author": "UtopiaXC"
}

class TelegramChannelModule(BaseModule):
    def enable_module(self) -> bool:
        self.set_module_config(
            key="channels",
            description="Telegram channel list (comma separated)",
            value="https://t.me/s/tnews365,https://t.me/s/scitech_fans",
            force_init=False,
            hint="Public channels only",
            regular=".*"
        )
        self.set_module_config(
            key="lookback_days",
            description="Lookback days",
            value="3",
            force_init=False,
            hint="Integer",
            regular="^\\d+$"
        )
        self.set_module_schedule_task(
            key="fetch_news",
            description="Fetch Telegram daily news",
            cron="0 8 * * *",
            force_init=False
        )
        
        channels = self.get_module_config("channels")
        if not channels:
            Log.w(TAG, "Missing channel configuration")
            return False

        Log.i(TAG, "Module enabled")
        return True

    def disable_module(self) -> bool:
        Log.i(TAG, "Module disabled")
        return True

    def test_module(self) -> Tuple[bool, str]:
        """
        Test if the module can successfully connect to at least one channel.
        """
        channels_str = self.get_module_config("channels")
        if not channels_str:
            # If config is missing, try a default one just for testing connectivity
            test_channels = ["https://t.me/s/telegram"]
        else:
            test_channels = [url.strip() for url in channels_str.split(',') if url.strip()]
            if not test_channels:
                return False, "No channels configured"

        # Use just the first channel for a quick connectivity test
        target_channel = test_channels[0]
        
        try:
            # Use the scraper's own test method
            scraper = TelegramScraper({})
            return scraper.test_connection(target_channel)
        except Exception as e:
            return False, f"Connection test failed: {str(e)}"

    def execute_schedule_task(self, cron: str, task_key: str, timestamp: datetime) -> bool:
        if task_key == "fetch_news":
            Log.i(TAG, f"Executing task: {task_key}")

            channels_str = self.get_module_config("channels")
            if not channels_str:
                Log.e(TAG, "No channels configured")
                return False
                
            channels = [url.strip() for url in channels_str.split(',') if url.strip()]

            lookback_days = self.get_module_config("lookback_days")
            config = {
                "lookback_days": int(lookback_days) if lookback_days else 3,
                "enable_ai": True
            }

            try:
                scraper = TelegramScraper(config)

                def save_callback(item_data):
                    Log.i(TAG, f"New item: {item_data['source']} - {item_data['title']}")
                    
                    # Tagging
                    raw_text = item_data.get('raw_data', {}).get('text', '')
                    if raw_text:
                        try:
                            tags = self.mark_message_tag(raw_text)
                            item_data['tags'] = tags
                        except Exception as e:
                            Log.w(TAG, f"Tagging failed: {e}")
                    
                    # Save
                    self.save_structured_results(item_data)

                scraper.run(channels, save_callback)
                Log.i(TAG, "Task completed successfully")
                return True

            except Exception as e:
                Log.e(TAG, f"Task execution failed: {e}")
                return False
        return False

    def generate_html(self, value) -> str:
        # Optional: Implement custom HTML generation if needed
        return None

    def generate_markdown(self, value) -> str:
        # Optional: Implement custom Markdown generation if needed
        return None


def create_module(context):
    return TelegramChannelModule(context)
