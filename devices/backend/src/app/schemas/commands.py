from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


class CommandCreate(BaseModel):
    model_config = ConfigDict(
        json_schema_extra={
            "examples": [
                {
                    "summary": "Simple info command",
                    "description": "Request basic device information",
                    "value": {"command_text": "get_info"},
                },
                {
                    "summary": "System diagnostic",
                    "description": "Run system diagnostics",
                    "value": {"command_text": "run_diagnostics --verbose"},
                },
                {
                    "summary": "Configuration update",
                    "description": "Update device configuration",
                    "value": {"command_text": "update_config --key=telemetry_interval --value=30"},
                },
            ]
        }
    )

    command_text: str = Field(
        ...,
        min_length=1,
        max_length=500,
        description="Command to execute on the device. Must be 1-500 characters.",
    )


class CommandResultSubmit(BaseModel):
    model_config = ConfigDict(
        json_schema_extra={
            "examples": [
                {
                    "summary": "Successful command execution",
                    "description": "Command completed successfully with output",
                    "value": {
                        "status": "completed",
                        "result": "Device info: CPU=Intel i7, Memory=16GB, OS=Ubuntu 22.04",
                        "exit_code": 0,
                    },
                },
                {
                    "summary": "Failed command execution",
                    "description": "Command failed with error message",
                    "value": {
                        "status": "failed",
                        "result": "Error: Permission denied. Requires administrator privileges.",
                        "exit_code": 1,
                    },
                },
                {
                    "summary": "Running command status",
                    "description": "Command still executing (progress update)",
                    "value": {
                        "status": "running",
                        "result": "Diagnostics in progress... 45% complete",
                        "exit_code": None,
                    },
                },
            ]
        }
    )

    status: str = Field(
        ...,
        pattern="^(completed|failed|running)$",
        description="Command execution status. Must be one of: completed, failed, running",
    )
    result: str | None = Field(
        None, max_length=10000, description="Command output or error message. Maximum 10,000 characters."
    )
    exit_code: int | None = Field(
        None,
        ge=0,
        le=255,
        description="Unix-style exit code. 0 indicates success, non-zero indicates error. Only applicable when status is 'completed' or 'failed'.",
    )


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
