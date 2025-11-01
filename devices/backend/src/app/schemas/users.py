from pydantic import BaseModel
from uuid import UUID
from datetime import datetime

class UserCreate(BaseModel):
    device_id: str
    name: str | None = None

class UserOut(UserCreate):
    id: UUID
    created_at: datetime

    class Config:
        orm_mode = True
