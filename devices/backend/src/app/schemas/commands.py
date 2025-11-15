from pydantic import BaseModel, Field, ConfigDict
from datetime import datetime
from typing import Optional
from uuid import UUID


class CommandCreate(BaseModel):
    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "command": "get_info"
            }
        }
    )
    
    command: str = Field(..., min_length=1, max_length=500)


class CommandResultSubmit(BaseModel):
    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "status": "completed",
                "result": "Command output here",
                "exit_code": 0
            }
        }
    )
    
    status: str = Field(..., pattern="^(completed|failed|running)$")
    result: Optional[str] = Field(None, max_length=10000)
    exit_code: Optional[int] = Field(0, ge=0, le=255)


class CommandOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)
    
    commandid: UUID
    deviceid: UUID
    command_text: str
    status: str
    created_at: datetime
    completed_at: Optional[datetime] = None
    result: Optional[str] = None
    exit_code: Optional[int] = None
