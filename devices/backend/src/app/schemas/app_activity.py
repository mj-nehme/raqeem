from pydantic import BaseModel
from uuid import UUID
from datetime import datetime

class AppActivityCreate(BaseModel):
    user_id: UUID
    app_name: str
    action: str

class AppActivityOut(AppActivityCreate):
    id: UUID
    activity_time: datetime

    class Config:
        orm_mode = True
