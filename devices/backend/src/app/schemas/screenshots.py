from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict


class ScreenshotCreate(BaseModel):
    user_id: UUID
    image_path: str


class ScreenshotOut(ScreenshotCreate):
    model_config = ConfigDict(from_attributes=True)

    id: UUID
    created_at: datetime
