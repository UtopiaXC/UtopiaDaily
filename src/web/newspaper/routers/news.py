from fastapi import APIRouter, Depends
from sqlalchemy import select, desc
from sqlalchemy.ext.asyncio import AsyncSession
from datetime import datetime, timedelta
import time

from src.database.connection import data_db_manager
from src.database.models.scraper_result import ScraperResult

router = APIRouter(prefix="/api/newspaper", tags=["Newspaper"])

async def get_data_db():
    async with data_db_manager.get_async_session() as session:
        yield session

@router.get("/latest")
async def get_latest_news(db: AsyncSession = Depends(get_data_db)):
    # Calculate timestamp for 3 days ago
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
        # row.value is the JSON content
        item = row.value
        # Ensure we have necessary fields, fallback to defaults if missing
        news_item = {
            "id": row.id,
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
