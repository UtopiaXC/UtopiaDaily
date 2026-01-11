from abc import ABC
from typing import Any, List, Dict, Optional
from datetime import datetime

class BaseModule(ABC):


    def __init__(self, context):
        """
        Initialize the module with the system context.
        :param context: The system context providing the execute interface implementation.
        """
        self._context = context

    # ==========================================
    # Execute Interface (System API)
    # ==========================================

    def set_module_config(self, key: str, description: str, value: str, force_init: bool = False, hint: str = "", regular: str = ""):
        """
        Set module configuration parameter.
        """
        return self._context.set_module_config(key, description, value, force_init, hint, regular)

    def get_module_config(self, key: str) -> str:
        """
        Get configuration from database.
        """
        return self._context.get_module_config(key)

    def drop_module_config(self, key: str):
        """
        Delete a configuration entry.
        """
        return self._context.drop_module_config(key)

    def set_module_schedule_task(self, key: str, description: str, cron: str, force_init: bool = False, hint: str = "", regular: str = ""):
        """
        Set a scheduled task for the module.
        """
        return self._context.set_module_schedule_task(key, description, cron, force_init, hint, regular)

    def get_module_schedule_task(self, key: str):
        """
        Get scheduled task details.
        """
        return self._context.get_module_schedule_task(key)

    def mark_message_tag(self, message: str) -> List[Dict[str, Any]]:
        """
        Get tags for a message using system's tagging model.
        """
        return self._context.mark_message_tag(message)

    def save_structured_results(self, value: Any) -> Dict[str, Any]:
        """
        Save structured results to the database.
        """
        return self._context.save_structured_results(value)

    # ==========================================
    # Listener Interface (To be implemented)
    # ==========================================

    def enable_module(self) -> bool:
        """
        Called when module is enabled.
        Should set up configurations here.
        """
        return True

    def disable_module(self) -> bool:
        """
        Called when module is disabled.
        """
        return True

    def execute_schedule_task(self, cron: str, task_key: str, timestamp: datetime) -> bool:
        """
        Called when a scheduled task is triggered.
        """
        return True

    def generate_html(self, value: Any) -> Optional[str]:
        """
        Generate HTML representation of the data.
        """
        return None

    def generate_markdown(self, value: Any) -> Optional[str]:
        """
        Generate Markdown representation of the data.
        """
        return None
