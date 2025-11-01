from pydantic import BaseModel
from uuid import UUID
from datetime import datetime

class KeystrokeCreate(BaseModel):
    user_id: UUID
    key: str

class KeystrokeOut(KeystrokeCreate):
    id: UUID
    created_at: datetime

    class Config:
        orm_mode = True
