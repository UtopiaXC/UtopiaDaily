from fastapi import APIRouter, Depends, HTTPException, Body, Request
from sqlalchemy.orm import Session
from typing import List, Dict, Any
import re
from src.database.connection import system_db_manager
from src.database.models import ScraperModule, ScraperModuleConfig
from src.web.dashboard.schemas import ScraperModuleResponse, ScraperModuleDetailResponse, TestModuleResponse, ScraperModuleConfigItem
from src.utils.logger.logger import Log
from src.scraper.modules.module_manager import ModuleManager
from src.web.dependencies import get_db
from src.utils.event import EventManager

router = APIRouter(prefix="/api/dashboard/scraper/modules", tags=["Scraper Modules"])

@router.get("/", response_model=List[ScraperModuleResponse])
async def get_modules(db: Session = Depends(get_db)):
    modules = db.query(ScraperModule).filter(ScraperModule.is_deleted == False).all()
    return modules

@router.get("/{module_id}", response_model=ScraperModuleDetailResponse)
async def get_module_detail(module_id: str, db: Session = Depends(get_db)):
    # Do NOT reload modules here. Only read from DB and local cache.
    local_manager = ModuleManager()
    
    module = db.query(ScraperModule).filter(ScraperModule.module_id == module_id, ScraperModule.is_deleted == False).first()
    if not module:
        raise HTTPException(status_code=404, detail="Module not found")

    # Fetch Configs from DB
    configs = db.query(ScraperModuleConfig).filter(ScraperModuleConfig.module_id == module_id).all()
    config_dict = {}
    for cfg in configs:
        config_dict[cfg.config_key] = {
            "value": cfg.value,
            "description": cfg.description,
            "type": cfg.type,
            "options": cfg.options,
            "hint": cfg.hint,
            "regular": cfg.regex,
            "source": cfg.source,
            "is_override": cfg.is_override
        }

    # Fetch Tasks from JSON (Legacy, TODO: Migrate to DB)
    tasks = local_manager.db.get(module_id, {}).get("tasks", {})
    
    return ScraperModuleDetailResponse(
        id=module.id,
        module_id=module.module_id,
        name=module.name,
        description=module.description,
        version=module.version,
        author=module.author,
        is_enable=module.is_enable,
        meta=module.meta,
        source=module.source,
        updated_at=module.updated_at,
        created_at=module.created_at,
        config=config_dict,
        tasks=tasks
    )

@router.post("/{module_id}/test_config", response_model=TestModuleResponse)
async def test_module_config(module_id: str, config: Dict[str, Any] = Body(...)):
    """
    Test the provided configuration against the module's validation logic.
    """
    local_manager = ModuleManager()
    success, message = local_manager.test_module_config(module_id, config)
    
    if not success:
        EventManager.record(
            level=EventManager.LEVEL_WARNING,
            category=EventManager.CATEGORY_MODULE,
            event_type="config_test_failed",
            summary=f"Config test failed for {module_id}",
            details={"message": message},
            source_id=module_id,
            is_resolved=False
        )
        
    return TestModuleResponse(success=success, message=message)

@router.post("/{module_id}/config")
async def update_module_config(module_id: str, req: Request, config: Dict[str, Any] = Body(...), db: Session = Depends(get_db)):
    module = db.query(ScraperModule).filter(ScraperModule.module_id == module_id).first()
    if not module:
        raise HTTPException(status_code=404, detail="Module not found")

    # 1. Basic Validation (Type & Regex)
    for key, value in config.items():
        cfg_item = db.query(ScraperModuleConfig).filter(
            ScraperModuleConfig.module_id == module_id,
            ScraperModuleConfig.config_key == key
        ).first()
        
        if cfg_item:
            if cfg_item.type in ['number', 'int', 'float', 'double']:
                if value is not None and value != "":
                    try:
                        float(value)
                    except (ValueError, TypeError):
                        raise HTTPException(status_code=400, detail=f"Config '{key}' must be a number")

            if cfg_item.regex:
                try:
                    val_str = str(value) if value is not None else ""
                    if not re.search(cfg_item.regex, val_str):
                        raise HTTPException(status_code=400, detail=f"Config '{key}' format is invalid")
                except re.error:
                    Log.w("CONFIG_UPDATE", f"Invalid regex for config {key}: {cfg_item.regex}")

    local_manager = ModuleManager()
    success, message = local_manager.test_module_config(module_id, config)
    if not success:
        EventManager.record(
            level=EventManager.LEVEL_WARNING,
            category=EventManager.CATEGORY_MODULE,
            event_type="config_save_failed",
            summary=f"Config save failed for {module_id}: Test failed",
            details={"message": message},
            source_id=module_id,
            is_resolved=False
        )
        raise HTTPException(status_code=400, detail=f"Configuration test failed: {message}")

    for key, value in config.items():
        cfg_item = db.query(ScraperModuleConfig).filter(
            ScraperModuleConfig.module_id == module_id,
            ScraperModuleConfig.config_key == key
        ).first()
        if cfg_item:
            cfg_item.value = value
    
    db.commit()
    
    current_user = getattr(req.state, "user", None)
    EventManager.record(
        level=EventManager.LEVEL_NORMAL,
        category=EventManager.CATEGORY_MODULE,
        event_type="config_updated",
        summary=f"Module config updated: {module_id}",
        details={"updated_by": current_user.username if current_user else "unknown"},
        source_id=module_id
    )
    
    return {"status": "success", "message": "Configuration updated"}

@router.post("/{module_id}/enable")
async def enable_module(module_id: str, req: Request):
    local_manager = ModuleManager()
    
    try:
        if local_manager.enable_module(module_id):
            current_user = getattr(req.state, "user", None)
            EventManager.record(
                level=EventManager.LEVEL_NORMAL,
                category=EventManager.CATEGORY_MODULE,
                event_type="module_enabled",
                summary=f"Module enabled: {module_id}",
                details={"enabled_by": current_user.username if current_user else "unknown"},
                source_id=module_id
            )
            return {"status": "success", "message": f"Module {module_id} enabled"}
        else:
            raise HTTPException(status_code=500, detail="Failed to enable module")
    except Exception as e:
        EventManager.record(
            level=EventManager.LEVEL_CRITICAL,
            category=EventManager.CATEGORY_MODULE,
            event_type="module_enable_failed",
            summary=f"Failed to enable module: {module_id}",
            details={"error": str(e)},
            source_id=module_id,
            is_resolved=False
        )
        raise HTTPException(status_code=400, detail=str(e))

@router.post("/{module_id}/disable")
async def disable_module(module_id: str, req: Request):
    local_manager = ModuleManager()
    
    local_manager.disable_module(module_id)
    
    current_user = getattr(req.state, "user", None)
    EventManager.record(
        level=EventManager.LEVEL_NORMAL,
        category=EventManager.CATEGORY_MODULE,
        event_type="module_disabled",
        summary=f"Module disabled: {module_id}",
        details={"disabled_by": current_user.username if current_user else "unknown"},
        source_id=module_id
    )
    
    return {"status": "success", "message": f"Module {module_id} disabled"}

@router.post("/{module_id}/test", response_model=TestModuleResponse)
async def test_module(module_id: str):
    local_manager = ModuleManager()
    
    success, message = local_manager.test_module(module_id)
    return TestModuleResponse(success=success, message=message)

@router.post("/reload")
async def reload_modules(req: Request):
    local_manager = ModuleManager()
    local_manager.reload_modules()
    
    current_user = getattr(req.state, "user", None)
    EventManager.record(
        level=EventManager.LEVEL_NORMAL,
        category=EventManager.CATEGORY_MODULE,
        event_type="modules_reloaded",
        summary="Modules reloaded",
        details={"triggered_by": current_user.username if current_user else "unknown"}
    )

    return {"status": "success", "message": "Modules reloaded"}
