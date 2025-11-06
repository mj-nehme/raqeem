from pydantic import BaseModel
from uuid import UUID
from datetime import datetime

class LocationCreate(BaseModel):
    user_id: UUID
    latitude: float
    longitude: float

class LocationOut(LocationCreate):
    id: UUID
    timestamp: datetime

    class Config:
        orm_mode = True
