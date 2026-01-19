import src.scraper.modules.default.bilibili_ranking_and_update.bilibili_service as service
from src.utils.logger.logger import Log
from src.scraper.modules.base_module import BaseModule

from datetime import datetime
from typing import Tuple, Dict, Any

TAG = "BILIBILI_RANKING_AND_UPDATE_CONTROLLER"

MODULE_META = {
    "id": "bilibili_ranking_and_update",
    "name": "module.bilibili_ranking_and_update.name",
    "description": "module.bilibili_ranking_and_update.description",
    "version": "1.0.0",
    "author": "UtopiaXC"
}


class BilibiliModule(BaseModule):
    def enable_module(self) -> bool:
        list_conf = service.get_init_configs()
        for conf in list_conf:
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
        return True, "success"

    def test_config(self, config: Dict[str, Any]) -> Tuple[bool, str]:
        return service.test_configuration(self, config)

    def execute_schedule_task(self, cron: str, task_key: str, timestamp: datetime) -> bool:
        return service.execute_schedule_task(self, cron, task_key)

    def generate_html(self, value) -> Any:
        # If content_type is html, return content directly
        if value.get("content_type") == "html":
            return {
                "id": value.get("id"), # Will be filled by router if missing
                "render_type": "html",
                "content": value.get("content"),
                "source": value.get("source")
            }
        return None

    def generate_markdown(self, value) -> None:
        return None


def create_module(context):
    return BilibiliModule(context)
