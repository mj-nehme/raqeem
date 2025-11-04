from fastapi import APIRouter, Depends, File, UploadFile, Form, HTTPException
from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
import uuid

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

@router.post("/", status_code=201)
async def create_screenshot(
    device_id: str = Form(...),
    file: UploadFile = File(...),
    db: AsyncSession = Depends(get_db)
):
    try:
        # Generate unique filename
        file_id = str(uuid.uuid4())
        filename = f"{file_id}.png"
        
        # For now, just store the filename as image_path
        # In production, you'd upload to S3 or file storage
        obj = ScreenshotModel(user_id=device_id, image_path=filename)
        async with db.begin():
            db.add(obj)
        
        return {
            "id": str(obj.id),
            "user_id": str(obj.user_id),
            "image_url": obj.image_path,
            "status": "success"
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Screenshot upload failed: {str(e)}")

@router.post("/json", status_code=201, response_model=ScreenshotResponse)
async def create_screenshot_json(screenshot: ScreenshotCreate, db: AsyncSession = Depends(get_db)):
    # Original JSON endpoint for compatibility
    obj = ScreenshotModel(user_id=screenshot.user_id, image_path=screenshot.image_url)
    async with db.begin():
        db.add(obj)
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
