from abc import ABC
from typing import Any, List, Dict, Optional, Tuple, Union
from datetime import datetime

class BaseModule(ABC):


    def __init__(self, context):
        """
        Initialize the module with the system context.
        :param context: The system context providing the execute interface implementation.
        """
        self._context = context

    def set_module_config(self, key: str, description: str, value: str, value_type: str = "string", options: Optional[Union[List, Dict]] = None, force_init: bool = False, hint: str = "", regular: str = ""):
        """
        Set module configuration parameter.
        :param key: Configuration key
        :param description: Human-readable description
        :param value: Default value
        :param value_type: Data type (text, number, switch, date, datetime, select, password, array)
        :param options: Options for select/switch types (e.g., ["A", "B"] or {"A": "Label A"})
        :param force_init: If True, resets the value to default even if it exists
        :param hint: UI hint text
        :param regular: Regex for validation
        """
        return self._context.set_module_config(key, description, value, value_type, options, force_init, hint, regular)

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

    def set_module_schedule_task(self, key: str, description: str, name: str = "", force_init: bool = False):
        """
        Set a scheduled task preset for the module.
        :param key: Task key (unique within module)
        :param description: Task description
        :param name: Human-readable name for the task
        :param force_init: If True, resets the task preset
        """
        return self._context.set_module_schedule_task(key, description, name, force_init)

    def get_module_schedule_task(self, key: str):
        """
        Get scheduled task details.
        """
        return self._context.get_module_schedule_task(key)

    def drop_module_schedule_task(self, key: str):
        """
        Delete a scheduled task.
        """
        return self._context.drop_module_schedule_task(key)

    def mark_message_tag(self, message: str) -> List[Dict[str, Any]]:
        """
        Get tags for a message using system's tagging model.
        """
        return self._context.mark_message_tag(message)

    def save_structured_results(self, value: Any, fingerprint: str = "") -> Dict[str, Any]:
        """
        Save structured results to the database.
        """
        return self._context.save_structured_results(value, fingerprint)

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

    def test_module(self) -> Tuple[bool, str]:
        """
        Called to test if the module is available/working correctly.
        Returns a tuple (success, message).
        """
        return True, "No self-test implemented"

    def test_config(self, config: Dict[str, Any]) -> Tuple[bool, str]:
        """
        Called to test if the provided configuration is valid.
        :param config: A dictionary of configuration key-value pairs.
        :return: (success, message)
        """
        return True, "Configuration valid"

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
