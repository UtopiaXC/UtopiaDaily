from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from pydantic import BaseModel
from src.database.connection import system_db_manager
from src.database.models import SystemConfig
from src.web.dependencies import RequirePermission
from src.utils.constants.permissions import Permissions
from src.utils.logger.logger import Log

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

@router.get("/", response_model=List[ConfigItem], dependencies=[Depends(RequirePermission(Permissions.SYSTEM_CONFIG_VIEW))])
async def get_configs(group: Optional[str] = None, db: Session = Depends(get_db)):
    query = db.query(SystemConfig)
    if group:
        query = query.filter(SystemConfig.group == group)
    query = query.order_by(SystemConfig.order.asc())
    return query.all()

@router.put("/{key}", dependencies=[Depends(RequirePermission(Permissions.SYSTEM_CONFIG_EDIT))])
async def update_config(key: str, update: ConfigUpdate, db: Session = Depends(get_db)):
    config = db.query(SystemConfig).filter(SystemConfig.key == key).first()
    if not config:
        raise HTTPException(status_code=404, detail="Config not found")
    
    if not config.is_editable:
        raise HTTPException(status_code=403, detail="This config is read-only")

    # Input Validation
    new_value = update.value
    
    if config.type == "boolean":
        if new_value.lower() not in ["true", "false"]:
            raise HTTPException(status_code=400, detail="Invalid boolean value")
            
    elif config.type == "int":
        if not new_value.isdigit():
            raise HTTPException(status_code=400, detail="Invalid integer value")
            
    elif config.type == "select":
        # If options are defined, validate against them
        if config.options:
            # Options can be list of strings or list of dicts {label, value}
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
    
    # Hot Reload Logic
    if key == "log_level":
        Log.set_level(new_value)

    return {"status": "success", "key": key, "new_value": update.value}
