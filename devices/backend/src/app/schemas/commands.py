from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


class CommandCreate(BaseModel):
    model_config = ConfigDict(json_schema_extra={"example": {"command_text": "get_info"}})

    command_text: str = Field(..., min_length=1, max_length=500)


class CommandResultSubmit(BaseModel):
    model_config = ConfigDict(
        json_schema_extra={"example": {"status": "completed", "result": "Command output here", "exit_code": 0}}
    )

    status: str = Field(..., pattern="^(completed|failed|running)$")
    result: str | None = Field(None, max_length=10000)
    exit_code: int | None = Field(0, ge=0, le=255)


class CommandOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    commandid: UUID
    deviceid: UUID
    command_text: str
    status: str
    created_at: datetime
    completed_at: datetime | None = None
    result: str | None = None
    exit_code: int | None = None
