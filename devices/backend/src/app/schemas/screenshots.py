from pydantic import BaseModel, ConfigDict
from uuid import UUID
from datetime import datetime

class ScreenshotCreate(BaseModel):
    user_id: UUID
    image_path: str

class ScreenshotOut(ScreenshotCreate):
    model_config = ConfigDict(from_attributes=True)
    
    id: UUID
    created_at: datetime
