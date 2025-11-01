from pydantic import BaseModel
from uuid import UUID
from datetime import datetime

class ScreenshotCreate(BaseModel):
    user_id: UUID
    image_path: str

class ScreenshotOut(ScreenshotCreate):
    id: UUID
    created_at: datetime

    class Config:
        orm_mode = True
