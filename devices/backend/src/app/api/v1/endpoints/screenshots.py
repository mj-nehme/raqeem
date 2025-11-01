from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()

class ScreenshotCreate(BaseModel):
    user_id: str
    image_path: str  # or maybe a file upload later?

@router.post("/")
async def create_screenshot(screenshot: ScreenshotCreate):
    return {"message": "Screenshot recorded", "data": screenshot}
