from fastapi import APIRouter, Depends
from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select

from app.db.session import get_db
from app.models.screenshots import Screenshot as ScreenshotModel

router = APIRouter()

class ScreenshotCreate(BaseModel):
    user_id: str
    image_url: str
    timestamp: Optional[datetime] = None

class ScreenshotResponse(BaseModel):
    id: str
    user_id: str
    image_url: str
    timestamp: Optional[datetime] = None

@router.post("/", status_code=201, response_model=ScreenshotResponse)
async def create_screenshot(screenshot: ScreenshotCreate, db: AsyncSession = Depends(get_db)):
    # Model uses image_path, map from image_url
    obj = ScreenshotModel(user_id=screenshot.user_id, image_path=screenshot.image_url)
    db.add(obj)
    await db.commit()
    await db.refresh(obj)
    return {
        "id": str(obj.id),
        "user_id": str(obj.user_id),
        "image_url": obj.image_path,
        "timestamp": screenshot.timestamp,
    }

@router.get("/", response_model=List[ScreenshotResponse])
async def get_screenshots(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(ScreenshotModel))
    items = result.scalars().all()
    return [
        {"id": str(x.id), "user_id": str(x.user_id), "image_url": x.image_path, "timestamp": None}
        for x in items
    ]
