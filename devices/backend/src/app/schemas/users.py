from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict


class UserCreate(BaseModel):
    device_id: str
    name: str | None = None


class UserOut(UserCreate):
    model_config = ConfigDict(from_attributes=True)

    id: UUID
    created_at: datetime
