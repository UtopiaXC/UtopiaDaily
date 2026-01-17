from fastapi import APIRouter, Depends, HTTPException, Request
from sqlalchemy.orm import Session
from typing import List, Optional
from pydantic import BaseModel
from src.database.connection import system_db_manager
from src.database.models import SystemConfig
from src.utils.logger.logger import Log
from src.utils.event import EventManager

router = APIRouter(prefix="/api/dashboard/system-config", tags=["System Config"])

# Schemas
class ConfigItem(BaseModel):
    key: str
    value: Optional[str]
    description: Optional[str]
    group: str
    type: str
    options: Optional[list] = None
    is_editable: bool = True
    order: int = 0

class ConfigUpdate(BaseModel):
    value: str

def get_db():
    db = system_db_manager.get_session()
    try:
        yield db
    finally:
        db.close()

@router.get("/", response_model=List[ConfigItem])
async def get_configs(group: Optional[str] = None, db: Session = Depends(get_db)):
    query = db.query(SystemConfig)
    if group:
        query = query.filter(SystemConfig.group == group)
    query = query.order_by(SystemConfig.order.asc())
    return query.all()

@router.put("/{key}")
async def update_config(key: str, update: ConfigUpdate, req: Request, db: Session = Depends(get_db)):
    config = db.query(SystemConfig).filter(SystemConfig.key == key).first()
    if not config:
        raise HTTPException(status_code=404, detail="Config not found")
    
    if not config.is_editable:
        raise HTTPException(status_code=403, detail="This config is read-only")

    new_value = update.value
    old_value = config.value
    
    if config.type == "boolean":
        if new_value.lower() not in ["true", "false"]:
            raise HTTPException(status_code=400, detail="Invalid boolean value")
            
    elif config.type == "int":
        if not new_value.isdigit():
            raise HTTPException(status_code=400, detail="Invalid integer value")
            
    elif config.type == "select":
        if config.options:
            valid_values = []
            for opt in config.options:
                if isinstance(opt, dict):
                    valid_values.append(opt.get("value"))
                else:
                    valid_values.append(str(opt))

            if valid_values and new_value not in valid_values:
                 raise HTTPException(status_code=400, detail=f"Invalid option. Must be one of {valid_values}")

    config.value = new_value
    db.commit()

    if key == "log_level":
        Log.set_level(new_value)
        
    current_user = getattr(req.state, "user", None)
    EventManager.record(
        level=EventManager.LEVEL_NORMAL,
        category=EventManager.CATEGORY_SYSTEM,
        event_type="config_updated",
        summary=f"System config updated: {key}",
        details={"key": key, "old_value": old_value, "new_value": new_value, "updated_by": current_user.username if current_user else "unknown"},
        source_id=current_user.id if current_user else None
    )

    return {"status": "success", "key": key, "new_value": update.value}
