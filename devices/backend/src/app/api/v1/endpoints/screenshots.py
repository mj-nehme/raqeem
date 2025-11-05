from fastapi import APIRouter, Depends, File, UploadFile, Form, HTTPException
from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
import uuid
import httpx
import logging

from app.db.session import get_db
from app.models.screenshots import Screenshot as ScreenshotModel
from app.models import devices as dev_models
from app.core.config import settings

router = APIRouter()
logger = logging.getLogger(__name__)

# Default screenshot resolution when not provided
DEFAULT_SCREENSHOT_RESOLUTION = "800x600"

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
        
        # Read file to get size
        content = await file.read()
        file_size = len(content)
        
        # Store in legacy screenshots table for backward compatibility
        obj = ScreenshotModel(user_id=device_id, image_path=filename)
        db.add(obj)
        
        # Also store in device_screenshots table with proper schema
        device_screenshot = dev_models.DeviceScreenshot(
            device_id=device_id,
            path=filename,
            resolution=DEFAULT_SCREENSHOT_RESOLUTION,
            size=file_size
        )
        db.add(device_screenshot)
        await db.commit()
        
        # Forward to mentor backend if configured
        if settings.mentor_api_url:
            try:
                async with httpx.AsyncClient(timeout=5.0) as client:
                    payload = {
                        "device_id": device_id,
                        "path": filename,
                        "resolution": DEFAULT_SCREENSHOT_RESOLUTION,
                        "size": file_size
                    }
                    # Forward screenshot metadata to mentor backend
                    await client.post(f"{settings.mentor_api_url}/devices/screenshots", json=payload)
            except (httpx.RequestError, httpx.TimeoutException) as e:
                # Log forwarding errors but don't fail the screenshot upload
                logger.warning(f"Failed to forward screenshot to mentor backend: {e}")
            except Exception as e:
                # Catch any other unexpected errors
                logger.error(f"Unexpected error forwarding screenshot: {e}")
                # Don't fail if forwarding fails
                pass
        
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
