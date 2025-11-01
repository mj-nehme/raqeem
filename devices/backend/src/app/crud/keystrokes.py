from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import insert
from app.models.keystrokes import Keystroke
from app.schemas.keystrokes import KeystrokeCreate

async def create_keystroke(db: AsyncSession, keystroke: KeystrokeCreate):
    stmt = insert(Keystroke).values(**keystroke.model_dump()).returning(Keystroke)
    result = await db.execute(stmt)
    await db.commit()
    return result.scalar_one()
