from app.models.devices import DeviceScreenshot as Screenshot
from app.schemas.screenshots import ScreenshotCreate
from sqlalchemy import insert
from sqlalchemy.ext.asyncio import AsyncSession


async def create_screenshot(db: AsyncSession, screenshot: ScreenshotCreate):
    stmt = insert(Screenshot).values(**screenshot.model_dump()).returning(Screenshot)
    result = await db.execute(stmt)
    await db.commit()
    return result.scalar_one()
