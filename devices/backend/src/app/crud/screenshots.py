from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import insert
from app.models.screenshots import Screenshot
from app.schemas.screenshots import ScreenshotCreate

async def create_screenshot(db: AsyncSession, screenshot: ScreenshotCreate):
    stmt = insert(Screenshot).values(**screenshot.model_dump()).returning(Screenshot)
    result = await db.execute(stmt)
    await db.commit()
    return result.scalar_one()
