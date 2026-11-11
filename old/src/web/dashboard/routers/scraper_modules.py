from fastapi import APIRouter, Depends, HTTPException, Body, Request
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from starlette.concurrency import run_in_threadpool
from typing import List, Dict, Any
import re
from src.database.models import ScraperModule, ScraperModuleConfig, ScraperModuleTask
from src.web.dashboard.schemas import ScraperModuleResponse, ScraperModuleDetailResponse, TestModuleResponse, ScraperModuleConfigItem, ScraperModuleTaskResponse, ScraperModuleTaskItem
from src.utils.logger.logger import Log
from src.scraper.modules.module_manager import ModuleManager
from src.web.dependencies import get_async_db
from src.utils.event import EventManager
from src.utils.classification.classification import list_local_models

router = APIRouter(prefix="/api/dashboard/scraper/modules", tags=["Scraper Modules"])

@router.get("/", response_model=List[ScraperModuleResponse])
async def get_modules(db: AsyncSession = Depends(get_async_db)):
    result = await db.execute(select(ScraperModule).where(ScraperModule.is_deleted == False))
    modules = result.scalars().all()

    response_modules = []
    for module in modules:
        response_modules.append(ScraperModuleResponse(
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
            created_at=module.created_at
        ))
        
    return response_modules

@router.get("/{module_id}", response_model=ScraperModuleDetailResponse)
async def get_module_detail(module_id: str, db: AsyncSession = Depends(get_async_db)):
    # Do NOT reload modules here. Only read from DB and local cache.
    local_manager = ModuleManager()
    
    result = await db.execute(select(ScraperModule).where(ScraperModule.module_id == module_id, ScraperModule.is_deleted == False))
    module = result.scalars().first()
    if not module:
        raise HTTPException(status_code=404, detail="Module not found")

    # Fetch Configs from DB
    config_result = await db.execute(select(ScraperModuleConfig).where(ScraperModuleConfig.module_id == module.id))
    configs = config_result.scalars().all()
    
    config_dict = {}
    for cfg in configs:
        options = cfg.options
        if cfg.config_key == "ai_model":
            options = await run_in_threadpool(list_local_models)

        value = cfg.value
        if cfg.type == "boolean":
            if isinstance(value, str):
                value = value.lower() == "true"
        
        config_dict[cfg.config_key] = {
            "value": value,
            "description": cfg.description,
            "type": cfg.type,
            "options": options,
            "hint": cfg.hint,
            "regular": cfg.regex,
            "source": cfg.source,
            "is_override": cfg.is_override
        }
    
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
        config=config_dict
    )

@router.get("/{module_id}/tasks", response_model=ScraperModuleTaskResponse)
async def get_module_tasks(module_id: str, db: AsyncSession = Depends(get_async_db)):
    result = await db.execute(select(ScraperModule).where(ScraperModule.module_id == module_id, ScraperModule.is_deleted == False))
    module = result.scalars().first()
    if not module:
        raise HTTPException(status_code=404, detail="Module not found")

    task_result = await db.execute(select(ScraperModuleTask).where(ScraperModuleTask.module_id == module.id))
    tasks = task_result.scalars().all()
    
    task_list = []
    for task in tasks:
        task_list.append({
            "id": task.id,
            "key": task.task_key,
            "name": task.name,
            "description": task.description
        })
    
    return ScraperModuleTaskResponse(
        module_id=module.module_id,
        module_name=module.name,
        tasks=task_list
    )

@router.post("/{module_id}/test_config", response_model=TestModuleResponse)
async def test_module_config(module_id: str, config: Dict[str, Any] = Body(...)):
    """
    Test the provided configuration against the module's validation logic.
    """
    local_manager = ModuleManager()
    success, message = await run_in_threadpool(local_manager.test_module_config, module_id, config)
    
    if not success:
        await EventManager.record_async(
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
async def update_module_config(module_id: str, req: Request, config: Dict[str, Any] = Body(...), db: AsyncSession = Depends(get_async_db)):
    result = await db.execute(select(ScraperModule).where(ScraperModule.module_id == module_id))
    module = result.scalars().first()
    if not module:
        raise HTTPException(status_code=404, detail="Module not found")
    module_internal_id = module.id

    for key, value in config.items():
        cfg_result = await db.execute(select(ScraperModuleConfig).where(
            ScraperModuleConfig.module_id == module_internal_id,
            ScraperModuleConfig.config_key == key
        ))
        cfg_item = cfg_result.scalars().first()
        
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
    success, message = await run_in_threadpool(local_manager.test_module_config, module_id, config)
    if not success:
        await EventManager.record_async(
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
        cfg_result = await db.execute(select(ScraperModuleConfig).where(
            ScraperModuleConfig.module_id == module_internal_id,
            ScraperModuleConfig.config_key == key
        ))
        cfg_item = cfg_result.scalars().first()
        
        if cfg_item:
            if cfg_item.type == "boolean" and isinstance(value, bool):
                cfg_item.value = str(value).lower()
            else:
                cfg_item.value = value
    
    await db.commit()
    
    current_user = getattr(req.state, "user", None)
    await EventManager.record_async(
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
        success = await run_in_threadpool(local_manager.enable_module, module_id)
        if success:
            current_user = getattr(req.state, "user", None)
            await EventManager.record_async(
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
        await EventManager.record_async(
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
    await run_in_threadpool(local_manager.disable_module, module_id)
    
    current_user = getattr(req.state, "user", None)
    await EventManager.record_async(
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
    success, message = await run_in_threadpool(local_manager.test_module, module_id)
    return TestModuleResponse(success=success, message=message)

@router.post("/reload")
async def reload_modules(req: Request):
    local_manager = ModuleManager()
    await run_in_threadpool(local_manager.reload_modules)
    
    current_user = getattr(req.state, "user", None)
    await EventManager.record_async(
        level=EventManager.LEVEL_NORMAL,
        category=EventManager.CATEGORY_MODULE,
        event_type="modules_reloaded",
        summary="Modules reloaded",
        details={"triggered_by": current_user.username if current_user else "unknown"}
    )

    return {"status": "success", "message": "Modules reloaded"}
