from fastapi import APIRouter, Depends, HTTPException, Body
from sqlalchemy.orm import Session
from typing import List, Dict, Any
import re
from src.database.connection import system_db_manager
from src.database.models import ScraperModule, ScraperModuleConfig
from src.web.dashboard.schemas import ScraperModuleResponse, ScraperModuleDetailResponse, TestModuleResponse, ScraperModuleConfigItem
from src.utils.logger.logger import Log
from src.scraper.modules.module_manager import ModuleManager
from src.web.dependencies import get_db

router = APIRouter(prefix="/api/dashboard/scraper/modules", tags=["Scraper Modules"])

@router.get("/", response_model=List[ScraperModuleResponse])
async def get_modules(db: Session = Depends(get_db)):
    modules = db.query(ScraperModule).filter(ScraperModule.is_deleted == False).all()
    return modules

@router.get("/{module_id}", response_model=ScraperModuleDetailResponse)
async def get_module_detail(module_id: str, db: Session = Depends(get_db)):
    # Reload modules to ensure we have the latest config/source/meta
    local_manager = ModuleManager()
    local_manager.reload_modules()
    
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
    return TestModuleResponse(success=success, message=message)

@router.post("/{module_id}/config")
async def update_module_config(module_id: str, config: Dict[str, Any] = Body(...), db: Session = Depends(get_db)):
    """
    Update module configuration.
    First validates types and regex.
    Then calls module.test_config().
    If all pass, saves to DB.
    """
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
            # Type Validation
            if cfg_item.type in ['number', 'int', 'float', 'double']:
                if value is not None and value != "":
                    try:
                        float(value)
                    except (ValueError, TypeError):
                        raise HTTPException(status_code=400, detail=f"Config '{key}' must be a number")
            
            # Regex Validation
            if cfg_item.regex:
                try:
                    val_str = str(value) if value is not None else ""
                    if not re.search(cfg_item.regex, val_str):
                        raise HTTPException(status_code=400, detail=f"Config '{key}' format is invalid")
                except re.error:
                    Log.w("CONFIG_UPDATE", f"Invalid regex for config {key}: {cfg_item.regex}")

    # 2. Module Logic Validation
    local_manager = ModuleManager()
    success, message = local_manager.test_module_config(module_id, config)
    if not success:
        raise HTTPException(status_code=400, detail=f"Configuration test failed: {message}")

    # 3. Save to DB
    for key, value in config.items():
        cfg_item = db.query(ScraperModuleConfig).filter(
            ScraperModuleConfig.module_id == module_id,
            ScraperModuleConfig.config_key == key
        ).first()
        if cfg_item:
            cfg_item.value = value
    
    db.commit()
    
    # No restart needed as per new requirement. 
    # Modules should read config dynamically or handle updates if needed.

    return {"status": "success", "message": "Configuration updated"}

@router.post("/{module_id}/enable")
async def enable_module(module_id: str):
    local_manager = ModuleManager()
    
    try:
        if local_manager.enable_module(module_id):
            return {"status": "success", "message": f"Module {module_id} enabled"}
        else:
            raise HTTPException(status_code=500, detail="Failed to enable module")
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@router.post("/{module_id}/disable")
async def disable_module(module_id: str):
    local_manager = ModuleManager()
    
    local_manager.disable_module(module_id)
    return {"status": "success", "message": f"Module {module_id} disabled"}

@router.post("/{module_id}/test", response_model=TestModuleResponse)
async def test_module(module_id: str):
    local_manager = ModuleManager()
    
    success, message = local_manager.test_module(module_id)
    return TestModuleResponse(success=success, message=message)

@router.post("/reload")
async def reload_modules():
    local_manager = ModuleManager()
    local_manager.reload_modules()
    return {"status": "success", "message": "Modules reloaded"}
