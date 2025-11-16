"""Pydantic schemas for device-related API endpoints."""
from datetime import datetime
from typing import Optional
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


class DeviceRegister(BaseModel):
    """Device registration request schema."""

    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "deviceid": "550e8400-e29b-41d4-a716-446655440000",
                "device_name": "laptop-001",
                "device_type": "laptop",
                "os": "Ubuntu 22.04",
                "device_location": "Office",
                "ip_address": "192.168.1.100",
                "mac_address": "00:1B:63:84:45:E6",
                "current_user": "john_doe",
            }
        }
    )

    deviceid: UUID = Field(..., description="Unique device identifier (UUID format)")
    device_name: Optional[str] = Field(None, description="Human-readable device name", max_length=255)
    device_type: Optional[str] = Field(None, description="Device type (e.g., laptop, desktop, server)", max_length=100)
    os: Optional[str] = Field(None, description="Operating system version", max_length=255)
    device_location: Optional[str] = Field(None, description="Physical location of device", max_length=255)
    ip_address: Optional[str] = Field(None, description="Current IP address", max_length=45)
    mac_address: Optional[str] = Field(None, description="MAC address", max_length=17)
    current_user: Optional[str] = Field(None, description="Currently logged-in user", max_length=255)


class DeviceRegisterResponse(BaseModel):
    """Device registration response schema."""

    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "deviceid": "550e8400-e29b-41d4-a716-446655440000",
                "created": True,
            }
        }
    )

    deviceid: UUID = Field(..., description="Device identifier")
    created: Optional[bool] = Field(None, description="True if device was created (False or absent if updated)")
    updated: Optional[bool] = Field(None, description="True if device was updated (False or absent if created)")


class DeviceMetric(BaseModel):
    """Device metrics submission schema."""

    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "cpu_usage": 45.2,
                "cpu_temp": 62.5,
                "memory_total": 16384,
                "memory_used": 8192,
                "swap_used": 512,
                "disk_total": 512000,
                "disk_used": 256000,
                "net_bytes_in": 1024000,
                "net_bytes_out": 512000,
            }
        }
    )

    cpu_usage: Optional[float] = Field(None, description="CPU usage percentage (0-100)", ge=0, le=100)
    cpu_temp: Optional[float] = Field(None, description="CPU temperature in Celsius", ge=-50, le=150)
    memory_total: Optional[int] = Field(None, description="Total memory in MB", ge=0)
    memory_used: Optional[int] = Field(None, description="Used memory in MB", ge=0)
    swap_used: Optional[int] = Field(None, description="Swap memory used in MB", ge=0)
    disk_total: Optional[int] = Field(None, description="Total disk space in MB", ge=0)
    disk_used: Optional[int] = Field(None, description="Used disk space in MB", ge=0)
    net_bytes_in: Optional[int] = Field(None, description="Network bytes received", ge=0)
    net_bytes_out: Optional[int] = Field(None, description="Network bytes sent", ge=0)


class DeviceProcess(BaseModel):
    """Device process information schema."""

    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "pid": 1234,
                "process_name": "python3",
                "cpu": 2.5,
                "memory": 150,
                "command_text": "python3 /usr/bin/app.py",
            }
        }
    )

    pid: int = Field(..., description="Process ID", ge=0)
    process_name: str = Field(..., description="Process name", max_length=255)
    cpu: Optional[float] = Field(None, description="CPU usage percentage", ge=0)
    memory: Optional[int] = Field(None, description="Memory usage in MB", ge=0)
    command_text: Optional[str] = Field(None, description="Full command line", max_length=1000)


class DeviceActivity(BaseModel):
    """Device activity log schema."""

    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "activity_type": "app_usage",
                "description": "User opened web browser",
                "app": "firefox",
                "duration": 3600,
            }
        }
    )

    activity_type: str = Field(..., description="Type of activity", max_length=100)
    description: Optional[str] = Field(None, description="Activity description", max_length=500)
    app: Optional[str] = Field(None, description="Application name", max_length=255)
    duration: Optional[int] = Field(None, description="Duration in seconds", ge=0)


class DeviceAlert(BaseModel):
    """Device alert submission schema."""

    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "level": "warning",
                "alert_type": "high_cpu",
                "message": "CPU usage exceeded threshold",
                "value": 85.5,
                "threshold": 80.0,
            }
        }
    )

    level: str = Field(..., description="Alert level (info, warning, error, critical)", pattern="^(info|warning|error|critical)$")
    alert_type: str = Field(..., description="Type of alert", max_length=100)
    message: str = Field(..., description="Alert message", max_length=500)
    value: Optional[float] = Field(None, description="Current value that triggered alert")
    threshold: Optional[float] = Field(None, description="Threshold value")


class DeviceOut(BaseModel):
    """Device information output schema."""

    model_config = ConfigDict(from_attributes=True)

    deviceid: UUID
    device_name: Optional[str] = None
    device_type: Optional[str] = None
    os: Optional[str] = None
    last_seen: Optional[datetime] = None
    is_online: Optional[bool] = None
    device_location: Optional[str] = None
    ip_address: Optional[str] = None
    mac_address: Optional[str] = None
    current_user: Optional[str] = None


class ErrorResponse(BaseModel):
    """Standard error response schema."""

    model_config = ConfigDict(json_schema_extra={"example": {"detail": "Error message describing the issue"}})

    detail: str = Field(..., description="Error message")


class SuccessResponse(BaseModel):
    """Standard success response schema."""

    model_config = ConfigDict(json_schema_extra={"example": {"status": "ok"}})

    status: str = Field(..., description="Status of the operation")


class InsertResponse(BaseModel):
    """Response for batch insert operations."""

    model_config = ConfigDict(json_schema_extra={"example": {"inserted": 5}})

    inserted: int = Field(..., description="Number of records inserted", ge=0)
