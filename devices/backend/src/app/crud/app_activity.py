from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import insert
from app.models.app_activity import AppActivity
from app.schemas.app_activity import AppActivityCreate

async def create_app_activity(db: AsyncSession, activity: AppActivityCreate):
    stmt = insert(AppActivity).values(**activity.model_dump()).returning(AppActivity)
    result = await db.execute(stmt)
    await db.commit()
    return result.scalar_one()
