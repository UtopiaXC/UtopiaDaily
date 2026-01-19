from fastapi import APIRouter, Depends, HTTPException, Query
from typing import Optional
from src.web.dependencies import get_current_user
from src.database.models import User
from src.web.dashboard.schemas import SystemEventResponse, PaginatedResponse
from src.utils.event import EventManager

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
    current_user: User = Depends(get_current_user)
):
    result = await EventManager.get_events_async(
        page=page,
        page_size=page_size,
        level=level,
        category=category,
        source_id=source_id,
        is_resolved=is_resolved,
        search=search,
        sort_by=sort_by,
        sort_order=sort_order
    )
    
    return PaginatedResponse(
        items=result["items"],
        total=result["total"],
        page=result["page"],
        page_size=result["page_size"]
    )

@router.post("/{event_id}/resolve")
async def resolve_event(
    event_id: str,
    current_user: User = Depends(get_current_user)
):
    success = await EventManager.resolve_event_by_id_async(event_id)
    if not success:
        raise HTTPException(status_code=404, detail="Event not found or failed to resolve")
        
    return {"status": "success"}
