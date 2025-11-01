from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import insert
from app.models.locations import Location
from app.schemas.locations import LocationCreate

async def create_location(db: AsyncSession, location: LocationCreate):
    stmt = insert(Location).values(**location.model_dump()).returning(Location)
    result = await db.execute(stmt)
    await db.commit()
    return result.scalar_one()
