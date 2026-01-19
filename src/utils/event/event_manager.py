from typing import Optional, Dict, Any, List
from sqlalchemy import select, delete, func, desc, asc, or_
import time
from src.database.connection import system_session_scope, system_async_session_scope
from src.database.models import SystemEvent, SystemConfig
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
    def _get_cleanup_config(session):
        try:
            max_count_config = session.query(SystemConfig).filter_by(key="event_max_count").first()
            max_days_config = session.query(SystemConfig).filter_by(key="event_max_days").first()
            
            max_count = int(max_count_config.value) if max_count_config else 100000
            if max_count <= 0: max_count = 100000
            
            max_days = int(max_days_config.value) if max_days_config else 30
            if max_days <= 0: max_days = 30
            
            return max_count, max_days
        except Exception as e:
            Log.w(TAG, f"Error reading cleanup config: {e}. Using defaults.")
            return 100000, 30

    @staticmethod
    async def _get_cleanup_config_async(session):
        try:
            result = await session.execute(select(SystemConfig).where(SystemConfig.key.in_(["event_max_count", "event_max_days"])))
            configs = {c.key: c.value for c in result.scalars().all()}
            
            max_count = int(configs.get("event_max_count", 100000))
            if max_count <= 0: max_count = 100000
            
            max_days = int(configs.get("event_max_days", 30))
            if max_days <= 0: max_days = 30
            
            return max_count, max_days
        except Exception as e:
            Log.w(TAG, f"Error reading cleanup config (async): {e}. Using defaults.")
            return 100000, 30

    @staticmethod
    def cleanup():
        deleted_count = 0
        try:
            with system_session_scope() as session:
                max_count, max_days = EventManager._get_cleanup_config(session)
                
                # 1. Cleanup by days (using milliseconds timestamp)
                current_time_ms = int(time.time() * 1000)
                retention_ms = max_days * 24 * 60 * 60 * 1000
                cutoff_time = current_time_ms - retention_ms
                
                res = session.query(SystemEvent).filter(SystemEvent.created_at < cutoff_time).delete(synchronize_session=False)
                deleted_count += res
                
                # 2. Cleanup by count
                count = session.query(func.count(SystemEvent.id)).scalar()
                if count > max_count:
                    cutoff_record = session.query(SystemEvent.created_at).order_by(SystemEvent.created_at.desc()).offset(max_count - 1).limit(1).scalar()
                    
                    if cutoff_record:
                        res = session.query(SystemEvent).filter(SystemEvent.created_at < cutoff_record).delete(synchronize_session=False)
                        deleted_count += res
            
            if deleted_count > 0:
                Log.w(TAG, f"Cleaned up {deleted_count} old system events")
                        
        except Exception as e:
            Log.e(TAG, "Failed to cleanup events", error=e)

    @staticmethod
    async def cleanup_async():
        deleted_count = 0
        try:
            async with system_async_session_scope() as session:
                max_count, max_days = await EventManager._get_cleanup_config_async(session)
                
                # 1. Cleanup by days (using milliseconds timestamp)
                current_time_ms = int(time.time() * 1000)
                retention_ms = max_days * 24 * 60 * 60 * 1000
                cutoff_time = current_time_ms - retention_ms
                
                res = await session.execute(delete(SystemEvent).where(SystemEvent.created_at < cutoff_time))
                deleted_count += res.rowcount
                
                # 2. Cleanup by count
                result = await session.execute(select(func.count(SystemEvent.id)))
                count = result.scalar()
                
                if count > max_count:
                    stmt = select(SystemEvent.created_at).order_by(SystemEvent.created_at.desc()).offset(max_count - 1).limit(1)
                    result = await session.execute(stmt)
                    cutoff_record = result.scalar()
                    
                    if cutoff_record:
                        res = await session.execute(delete(SystemEvent).where(SystemEvent.created_at < cutoff_record))
                        deleted_count += res.rowcount
            
            if deleted_count > 0:
                Log.w(TAG, f"Cleaned up {deleted_count} old system events (async)")
                    
        except Exception as e:
            Log.e(TAG, "Failed to cleanup events (async)", error=e)

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
        try:
            EventManager.cleanup()
            
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
    async def record_async(
        level: str,
        category: str,
        event_type: str,
        summary: str,
        details: Optional[Dict[str, Any]] = None,
        source_id: Optional[str] = None,
        is_resolved: bool = True
    ):
        try:
            await EventManager.cleanup_async()
            
            async with system_async_session_scope() as session:
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
                # commit is handled by context manager

            log_msg = f"[{category}] {summary}"
            if level == EventManager.LEVEL_NORMAL:
                Log.i(TAG, log_msg)
            elif level == EventManager.LEVEL_WARNING:
                Log.w(TAG, log_msg)
            else:
                Log.e(TAG, log_msg)
                    
        except Exception as e:
            Log.e(TAG, f"Failed to record async event: {summary}", error=e)

    @staticmethod
    def resolve(event_id: str):
        try:
            EventManager.cleanup()
            with system_session_scope() as session:
                event = session.query(SystemEvent).filter_by(id=event_id).first()
                if event:
                    event.is_resolved = True
                    Log.i(TAG, f"Event {event_id} marked as resolved")
        except Exception as e:
            Log.e(TAG, f"Failed to resolve event {event_id}", error=e)

    @staticmethod
    async def resolve_async(event_id: str):
        try:
            await EventManager.cleanup_async()
            async with system_async_session_scope() as session:
                result = await session.execute(select(SystemEvent).where(SystemEvent.id == event_id))
                event = result.scalars().first()
                if event:
                    event.is_resolved = True
                    Log.i(TAG, f"Event {event_id} marked as resolved (async)")
        except Exception as e:
            Log.e(TAG, f"Failed to resolve event {event_id} (async)", error=e)

    @staticmethod
    async def get_events_async(
        page: int = 1,
        page_size: int = 20,
        level: Optional[str] = None,
        category: Optional[str] = None,
        source_id: Optional[str] = None,
        is_resolved: Optional[bool] = None,
        search: Optional[str] = None,
        sort_by: Optional[str] = "created_at",
        sort_order: Optional[str] = "desc"
    ) -> Dict[str, Any]:
        
        # Cleanup on read as well, as requested
        await EventManager.cleanup_async()
        
        async with system_async_session_scope() as session:
            stmt = select(SystemEvent)

            if level:
                stmt = stmt.where(SystemEvent.level == level)
            if category:
                stmt = stmt.where(SystemEvent.category == category)
            if source_id:
                stmt = stmt.where(SystemEvent.source_id == source_id)
            if is_resolved is not None:
                stmt = stmt.where(SystemEvent.is_resolved == is_resolved)

            if search:
                search_term = f"%{search}%"
                stmt = stmt.where(
                    or_(
                        SystemEvent.summary.ilike(search_term),
                        SystemEvent.event_type.ilike(search_term),
                        SystemEvent.source_id.ilike(search_term)
                    )
                )

            # Count total
            count_stmt = select(func.count()).select_from(stmt.subquery())
            total_result = await session.execute(count_stmt)
            total = total_result.scalar()

            if sort_by and hasattr(SystemEvent, sort_by):
                column = getattr(SystemEvent, sort_by)
                if sort_order == "asc":
                    stmt = stmt.order_by(asc(column))
                else:
                    stmt = stmt.order_by(desc(column))
            else:
                stmt = stmt.order_by(desc(SystemEvent.created_at))
            
            stmt = stmt.offset((page - 1) * page_size).limit(page_size)
            result = await session.execute(stmt)
            events = result.scalars().all()
            
            # Convert to dict to avoid detached instance errors
            return {
                "items": [event.to_dict() for event in events],
                "total": total,
                "page": page,
                "page_size": page_size
            }

    @staticmethod
    async def resolve_event_by_id_async(event_id: str) -> bool:
        await EventManager.cleanup_async()
        try:
            async with system_async_session_scope() as session:
                result = await session.execute(select(SystemEvent).where(SystemEvent.id == event_id))
                event = result.scalars().first()
                if not event:
                    return False
                
                event.is_resolved = True
                return True
        except Exception as e:
            Log.e(TAG, f"Failed to resolve event {event_id}", error=e)
            return False
