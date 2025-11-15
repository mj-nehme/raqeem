from pydantic import BaseModel, ConfigDict
from uuid import UUID
from datetime import datetime

class UserCreate(BaseModel):
    device_id: str
    name: str | None = None

class UserOut(UserCreate):
    model_config = ConfigDict(from_attributes=True)
    
    id: UUID
    created_at: datetime
