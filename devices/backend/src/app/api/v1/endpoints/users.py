
from fastapi import APIRouter, Depends
from pydantic import BaseModel
from typing import List, Optional
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select

from app.db.session import get_db
from app.models.users import User as UserModel

router = APIRouter()

class UserCreate(BaseModel):
    device_id: str
    name: Optional[str] = None

class UserResponse(BaseModel):
    id: str
    device_id: str
    name: Optional[str] = None

@router.post("/", status_code=201, response_model=UserResponse)
async def create_user(user: UserCreate, db: AsyncSession = Depends(get_db)):
    obj = UserModel(deviceid=user.device_id, username=user.name)
    # Use a transaction context and avoid an immediate refresh to prevent overlapping operations
    async with db.begin():
        db.add(obj)
    # At this point, the object has been flushed/committed; UUID is generated client-side
    return {"id": str(obj.userid), "deviceid": str(obj.deviceid), "name": obj.username}

@router.get("/", response_model=List[UserResponse])
async def get_users(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(UserModel))
    items = result.scalars().all()
    return [{"id": str(u.userid), "deviceid": str(u.deviceid), "name": u.username} for u in items]
