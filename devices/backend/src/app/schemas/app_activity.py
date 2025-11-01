from pydantic import BaseModel
from uuid import UUID
from datetime import datetime

class AppActivityCreate(BaseModel):
    user_id: UUID
    app_name: str
    activity: str

class AppActivityOut(AppActivityCreate):
    id: UUID
    created_at: datetime

    class Config:
        orm_mode = True
