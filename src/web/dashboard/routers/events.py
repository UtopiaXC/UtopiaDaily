from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import desc, asc, or_, select, func
from typing import Optional
from src.web.dependencies import get_async_db, get_current_user
from src.database.models import SystemEvent, User
from src.web.dashboard.schemas import SystemEventResponse, PaginatedResponse

router = APIRouter(prefix="/events", tags=["events"])

@router.get("", response_model=PaginatedResponse[SystemEventResponse])
async def get_events(
    page: int = 1,
    page_size: int = 20,
    level: Optional[str] = None,
    category: Optional[str] = None,
    source_id: Optional[str] = None,
    is_resolved: Optional[bool] = None,
    search: Optional[str] = None,
    sort_by: Optional[str] = "created_at",
    sort_order: Optional[str] = "desc",
    db: AsyncSession = Depends(get_async_db),
    current_user: User = Depends(get_current_user)
):
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
    total_result = await db.execute(count_stmt)
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
    result = await db.execute(stmt)
    events = result.scalars().all()
        
    return PaginatedResponse(
        items=[event.to_dict() for event in events],
        total=total,
        page=page,
        page_size=page_size
    )

@router.post("/{event_id}/resolve")
async def resolve_event(
    event_id: str,
    db: AsyncSession = Depends(get_async_db),
    current_user: User = Depends(get_current_user)
):
    result = await db.execute(select(SystemEvent).where(SystemEvent.id == event_id))
    event = result.scalars().first()
    if not event:
        raise HTTPException(status_code=404, detail="Event not found")
        
    event.is_resolved = True
    await db.commit()
    return {"status": "success"}
