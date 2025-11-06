from fastapi import APIRouter, Depends
from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select

from app.db.session import get_db
from app.models.app_activity import AppActivity as AppActivityModel

router = APIRouter()

class AppActivityCreate(BaseModel):
    user_id: str
    action: str
    timestamp: Optional[datetime] = None
    app_name: Optional[str] = None

class AppActivityResponse(BaseModel):
    id: str
    user_id: str
    action: str
    timestamp: Optional[datetime] = None

@router.post("/", status_code=201, response_model=AppActivityResponse)
async def create_app_activity(activity: AppActivityCreate, db: AsyncSession = Depends(get_db)):
    obj = AppActivityModel(
        user_id=activity.user_id,
        app_name=activity.app_name or "unknown",
        action=activity.action,
    )
    async with db.begin():
        db.add(obj)
    return {"id": str(obj.id), "user_id": str(obj.user_id), "action": obj.action, "timestamp": obj.activity_time}

@router.get("/", response_model=List[AppActivityResponse])
async def get_app_activities(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(AppActivityModel))
    items = result.scalars().all()
    return [
        {"id": str(x.id), "user_id": str(x.user_id), "action": x.action, "timestamp": x.activity_time}
        for x in items
    ]
