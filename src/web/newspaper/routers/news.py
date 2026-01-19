from fastapi import APIRouter, Depends
from sqlalchemy import select, desc
from sqlalchemy.ext.asyncio import AsyncSession
from starlette.concurrency import run_in_threadpool
from datetime import datetime, timedelta
import time
import importlib.util

from src.database.connection import data_db_manager, system_db_manager
from src.database.models.scraper_result import ScraperResult
from src.database.models.scraper_module import ScraperModule
from src.scraper.modules.module_manager import ModuleManager, ModuleContext
from src.utils.logger.logger import Log

router = APIRouter(prefix="/api/newspaper", tags=["Newspaper"])

async def get_data_db():
    async with data_db_manager.get_async_session() as session:
        yield session

async def get_system_db():
    async with system_db_manager.get_async_session() as session:
        yield session

@router.get("/latest")
async def get_latest_news(
    db: AsyncSession = Depends(get_data_db),
    sys_db: AsyncSession = Depends(get_system_db)
):
    # 1. Get Module UUID -> ID mapping
    stmt_modules = select(ScraperModule.id, ScraperModule.module_id)
    result_modules = await sys_db.execute(stmt_modules)
    module_map = {str(row.id): row.module_id for row in result_modules.all()}
    
    manager = ModuleManager()

    # 2. Get News
    three_days_ago_ms = int((time.time() - 3 * 24 * 60 * 60) * 1000)
    
    stmt = (
        select(ScraperResult)
        .where(ScraperResult.created_at >= three_days_ago_ms)
        .order_by(desc(ScraperResult.created_at))
        .limit(20)
    )
    
    result = await db.execute(stmt)
    rows = result.scalars().all()
    
    news_list = []
    for row in rows:
        item = row.value
        module_uuid = str(row.module_id)
        module_id_str = module_map.get(module_uuid)
        
        generated_html = None
        
        # Try to generate HTML via module
        if module_id_str:
            try:
                def generate_html_sync(mod_id, data):
                    mod_info = manager.get_module_info(mod_id)
                    if not mod_info: return None
                    
                    try:
                        spec = importlib.util.spec_from_file_location(f"gen_html_{mod_id}", mod_info["path"])
                        module_lib = importlib.util.module_from_spec(spec)
                        spec.loader.exec_module(module_lib)
                        
                        ctx = ModuleContext(mod_id, manager)
                        instance = module_lib.create_module(ctx)
                        
                        if hasattr(instance, 'generate_html'):
                            return instance.generate_html(data)
                    except Exception as e:
                        Log.w("NEWS_ROUTER", f"Failed to generate HTML for {mod_id}: {e}")
                    return None

                generated_html = await run_in_threadpool(generate_html_sync, module_id_str, item)
                    
            except Exception as e:
                Log.e("NEWS_ROUTER", f"Error processing module {module_id_str}", error=e)

        if generated_html:
            # HTML Mode: Minimal response, module controls everything
            news_item = {
                "id": row.id,
                "render_type": "html",
                "content": generated_html,
                # We still need source for grouping in frontend, unless we change frontend logic
                "source": item.get("source", "Unknown") 
            }
        else:
            # Raw Mode: Full metadata for frontend rendering
            news_item = {
                "id": row.id,
                "render_type": "raw",
                "title": item.get("title", "Untitled"),
                "summary": item.get("summary", ""),
                "source": item.get("source", "Unknown"),
                "datetime_released": item.get("datetime_released", ""),
                "content": item.get("content", ""),
                "tags": item.get("tags", []),
                "from_url": item.get("from_url", ""),
                "attachments": item.get("attachments", [])
            }
            
        news_list.append(news_item)

    return {"news": news_list}
