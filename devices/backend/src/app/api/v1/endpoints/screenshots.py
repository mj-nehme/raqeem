from fastapi import APIRouter
from pydantic import BaseModel
from typing import List
from datetime import datetime

router = APIRouter()

class ScreenshotCreate(BaseModel):
    user_id: str
    image_url: str  # Changed from image_path to image_url to match test
    timestamp: datetime | None = None

class ScreenshotResponse(BaseModel):
    id: int
    user_id: str
    image_url: str
    timestamp: datetime | None = None

# Mock data store
screenshots_db: List[dict] = []
screenshot_id_counter = 1

@router.post("/", status_code=201, response_model=ScreenshotResponse)
async def create_screenshot(screenshot: ScreenshotCreate):
    global screenshot_id_counter
    screenshot_data = {
        "id": screenshot_id_counter,
        "user_id": screenshot.user_id,
        "image_url": screenshot.image_url,
        "timestamp": screenshot.timestamp
    }
    screenshots_db.append(screenshot_data)
    screenshot_id_counter += 1
    return screenshot_data

@router.get("/", response_model=List[ScreenshotResponse])
async def get_screenshots():
    return screenshots_db
