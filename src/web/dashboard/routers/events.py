from fastapi import APIRouter, Depends, Query, HTTPException
from sqlalchemy.orm import Session
from sqlalchemy import desc
from typing import Optional, List
from src.web.dependencies import get_db, get_current_user
from src.database.models import SystemEvent, User
from src.web.dashboard.schemas import SystemEventResponse, PaginatedResponse

router = APIRouter(prefix="/events", tags=["events"])

@router.get("", response_model=PaginatedResponse[SystemEventResponse])
def get_events(
    page: int = 1,
    page_size: int = 20,
    level: Optional[str] = None,
    category: Optional[str] = None,
    source_id: Optional[str] = None,
    is_resolved: Optional[bool] = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    # Check permission (assuming all logged-in users can view events for now, or restrict to admin)
    # if current_user.role.name != "admin": ...
    
    query = db.query(SystemEvent)
    
    if level:
        query = query.filter(SystemEvent.level == level)
    if category:
        query = query.filter(SystemEvent.category == category)
    if source_id:
        query = query.filter(SystemEvent.source_id == source_id)
    if is_resolved is not None:
        query = query.filter(SystemEvent.is_resolved == is_resolved)
        
    total = query.count()
    
    events = query.order_by(desc(SystemEvent.created_at))\
        .offset((page - 1) * page_size)\
        .limit(page_size)\
        .all()
        
    return PaginatedResponse(
        items=[event.to_dict() for event in events],
        total=total,
        page=page,
        page_size=page_size
    )

@router.post("/{event_id}/resolve")
def resolve_event(
    event_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    event = db.query(SystemEvent).filter(SystemEvent.id == event_id).first()
    if not event:
        raise HTTPException(status_code=404, detail="Event not found")
        
    event.is_resolved = True
    db.commit()
    return {"status": "success"}
