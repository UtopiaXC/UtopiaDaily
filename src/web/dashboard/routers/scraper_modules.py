from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from src.database.connection import system_db_manager
from src.database.models import ScraperModule
from src.web.dashboard.schemas import ScraperModuleResponse, ScraperModuleDetailResponse, TestModuleResponse
from src.utils.logger.logger import Log
from src.scraper.modules.module_manager import ModuleManager

router = APIRouter(prefix="/api/dashboard/scraper/modules", tags=["Scraper Modules"])

def get_db():
    db = system_db_manager.get_session()
    try:
        yield db
    finally:
        db.close()

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
    
    config = local_manager.db.get(module_id, {}).get("config", {})
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
        config=config,
        tasks=tasks
    )

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
