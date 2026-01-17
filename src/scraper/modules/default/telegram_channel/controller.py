import src.scraper.modules.default.telegram_channel.telegram_channel_service as service
from src.utils.logger.logger import Log
from src.scraper.modules.base_module import BaseModule

from datetime import datetime
from typing import Tuple, Dict, Any

TAG = "TELEGRAM_CHANNEL_MODULE_API"

MODULE_META = {
    "id": "telegram_channel_scraper",
    "name": "module.telegram_channel.name",
    "description": "module.telegram_channel.description",
    "version": "1.0.0",
    "author": "UtopiaXC"
}


class TelegramChannelModule(BaseModule):
    def enable_module(self) -> bool:
        list_conf = service.get_init_configs()

        for conf in list_conf:
            Log.w(TAG, conf)
            self.set_module_config(
                key=conf["key"],
                description=conf["description"],
                value=conf["value"],
                hint=conf["hint"],
                regular=conf["regular"],
                value_type=conf["value_type"]
            )
        list_tasks = service.get_init_schedule_tasks()
        for task in list_tasks:
            self.set_module_schedule_task(
                key=task["key"],
                description=task["description"],
                name=task["name"]
            )
        Log.i(TAG, "Module enabled")
        return True

    def disable_module(self) -> bool:
        Log.i(TAG, "Module disabled")
        return True

    def test_module(self) -> Tuple[bool, str]:
        if service.test_telegram_access():
            return True, "module.telegram_channel.test.success"
        else:
            return False, "module.telegram_channel.test.failed"

    def test_config(self, config: Dict[str, Any]) -> Tuple[bool, str]:
        return service.test_configuration(self, config)

    def execute_schedule_task(self, cron: str, task_key: str, timestamp: datetime) -> bool:
        return service.execute_schedule_task(self, cron, task_key, timestamp)

    def generate_html(self, value) -> None:
        return None

    def generate_markdown(self, value) -> None:
        return None


def create_module(context):
    return TelegramChannelModule(context)
