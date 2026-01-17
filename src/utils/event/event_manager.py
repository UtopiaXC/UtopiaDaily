from typing import Optional, Dict, Any
from src.database.connection import system_session_scope
from src.database.models import SystemEvent
from src.utils.logger.logger import Log

TAG = "EVENT_MANAGER"

class EventManager:
    
    LEVEL_NORMAL = "NORMAL"
    LEVEL_WARNING = "WARNING"
    LEVEL_CRITICAL = "CRITICAL"
    LEVEL_FATAL = "FATAL"

    CATEGORY_SYSTEM = "SYSTEM"
    CATEGORY_USER = "USER"
    CATEGORY_MODULE = "MODULE"
    CATEGORY_TASK = "TASK"
    CATEGORY_SECURITY = "SECURITY"

    @staticmethod
    def record(
        level: str,
        category: str,
        event_type: str,
        summary: str,
        details: Optional[Dict[str, Any]] = None,
        source_id: Optional[str] = None,
        is_resolved: bool = True
    ):
        """
        Record a system event.
        :param level: Event level (NORMAL, WARNING, CRITICAL, FATAL)
        :param category: Event category (SYSTEM, USER, MODULE, TASK, SECURITY)
        :param event_type: Specific event type identifier (e.g., 'user_login')
        :param summary: Short description
        :param details: Structured details (JSON serializable)
        :param source_id: Related User ID or Module ID
        :param is_resolved: Whether the issue is resolved (default True for NORMAL events)
        """
        try:
            with system_session_scope() as session:
                event = SystemEvent(
                    level=level,
                    category=category,
                    event_type=event_type,
                    summary=summary,
                    details=details,
                    source_id=source_id,
                    is_resolved=is_resolved
                )
                session.add(event)
                
                # Log to console as well for immediate visibility
                log_msg = f"[{category}] {summary}"
                if level == EventManager.LEVEL_NORMAL:
                    Log.i(TAG, log_msg)
                elif level == EventManager.LEVEL_WARNING:
                    Log.w(TAG, log_msg)
                else:
                    Log.e(TAG, log_msg)
                    
        except Exception as e:
            Log.e(TAG, f"Failed to record event: {summary}", error=e)

    @staticmethod
    def resolve(event_id: str):
        """Mark an event as resolved."""
        try:
            with system_session_scope() as session:
                event = session.query(SystemEvent).filter_by(id=event_id).first()
                if event:
                    event.is_resolved = True
                    Log.i(TAG, f"Event {event_id} marked as resolved")
        except Exception as e:
            Log.e(TAG, f"Failed to resolve event {event_id}", error=e)
