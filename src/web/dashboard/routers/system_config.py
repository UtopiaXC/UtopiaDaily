from fastapi import APIRouter, Depends, HTTPException, Request, BackgroundTasks
from sqlalchemy.orm import Session
from typing import List, Optional
import json
from pydantic import BaseModel
from src.database.connection import system_db_manager
from src.database.models import SystemConfig
from src.utils.logger.logger import Log
from src.utils.event import EventManager
from src.utils.classification.classification import list_local_models
from src.utils.classification.download_model import download_model_from_hf, get_default_models_json

router = APIRouter(prefix="/api/dashboard/system-config", tags=["System Config"])

# Schemas
class ConfigItem(BaseModel):
    key: str
    value: Optional[str]
    description: Optional[str]
    hint: Optional[str]
    tab: str
    group: str
    type: str
    options: Optional[list] = None
    is_editable: bool = True
    order: int = 0

class ConfigUpdate(BaseModel):
    value: str

class ModelDownloadRequest(BaseModel):
    repo_id: str

def get_db():
    db = system_db_manager.get_session()
    try:
        yield db
    finally:
        db.close()

@router.get("/", response_model=List[ConfigItem])
def get_configs(tab: Optional[str] = None, group: Optional[str] = None, db: Session = Depends(get_db)):
    query = db.query(SystemConfig)
    if tab:
        query = query.filter(SystemConfig.tab == tab)
    if group:
        query = query.filter(SystemConfig.group == group)
    query = query.order_by(SystemConfig.order.asc())
    
    configs = query.all()
    
    # Dynamic options for ai_model
    for conf in configs:
        if conf.key == "ai_model":
            conf.options = list_local_models()
            
    return configs

@router.put("/{key}")
def update_config(key: str, update: ConfigUpdate, req: Request, db: Session = Depends(get_db)):
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
    
    elif config.type == "json":
        try:
            json.loads(new_value)
        except json.JSONDecodeError:
            raise HTTPException(status_code=400, detail="Invalid JSON format")

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

@router.get("/ai-models")
def get_ai_models():
    return list_local_models()

@router.get("/default-models")
def get_default_models():
    return get_default_models_json()

@router.post("/download-model")
def download_model(request: ModelDownloadRequest, req: Request):
    current_user = getattr(req.state, "user", None)
    
    try:
        # Synchronous download as requested
        success, msg = download_model_from_hf(request.repo_id)
        
        if success:
            EventManager.record(
                level=EventManager.LEVEL_NORMAL,
                category=EventManager.CATEGORY_SYSTEM,
                event_type="model_download_success",
                summary=f"Model downloaded: {request.repo_id}",
                source_id=current_user.id if current_user else None
            )
            return {"status": "success", "message": msg}
        else:
            EventManager.record(
                level=EventManager.LEVEL_CRITICAL,
                category=EventManager.CATEGORY_SYSTEM,
                event_type="model_download_failed",
                summary=f"Model download failed: {request.repo_id}",
                details={"error": msg},
                source_id=current_user.id if current_user else None
            )
            raise HTTPException(status_code=500, detail=f"Download failed: {msg}")
            
    except Exception as e:
         EventManager.record(
            level=EventManager.LEVEL_CRITICAL,
            category=EventManager.CATEGORY_SYSTEM,
            event_type="model_download_failed",
            summary=f"Model download failed: {request.repo_id}",
            details={"error": str(e)},
            source_id=current_user.id if current_user else None
        )
         raise HTTPException(status_code=500, detail=str(e))
