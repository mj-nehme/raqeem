from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import insert
from app.models.users import User
from app.schemas.users import UserCreate

async def create_user(db: AsyncSession, user: UserCreate):
    stmt = insert(User).values(**user.model_dump()).returning(User)
    result = await db.execute(stmt)
    await db.commit()
    return result.scalar_one()
