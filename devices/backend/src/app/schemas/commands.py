from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional


class CommandCreate(BaseModel):
    command: str = Field(..., min_length=1, max_length=500)
    
    class Config:
        json_schema_extra = {
            "example": {
                "command": "get_info"
            }
        }


class CommandResultSubmit(BaseModel):
    status: str = Field(..., pattern="^(completed|failed|running)$")
    result: Optional[str] = Field(None, max_length=10000)
    exit_code: Optional[int] = Field(0, ge=0, le=255)
    
    class Config:
        json_schema_extra = {
            "example": {
                "status": "completed",
                "result": "Command output here",
                "exit_code": 0
            }
        }


class CommandOut(BaseModel):
    id: int
    device_id: str
    command: str
    status: str
    created_at: datetime
    completed_at: Optional[datetime] = None
    result: Optional[str] = None
    exit_code: Optional[int] = None
    
    class Config:
        from_attributes = True
