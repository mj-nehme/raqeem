from pydantic import BaseModel
from uuid import UUID
from datetime import datetime

class KeystrokeCreate(BaseModel):
    user_id: UUID
    keylog: str

class KeystrokeOut(KeystrokeCreate):
    id: UUID
    logged_at: datetime

    class Config:
        orm_mode = True
