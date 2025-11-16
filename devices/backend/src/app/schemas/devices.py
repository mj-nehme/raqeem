"""Pydantic schemas for device-related endpoints."""

from datetime import datetime
from typing import Literal
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


# Device Registration Schemas
class DeviceRegister(BaseModel):
    """Device registration request schema."""

    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "deviceid": "a843a399-701f-5011-aff3-4b69d8f21b11",
                "device_name": "Office Laptop",
                "device_type": "laptop",
                "os": "Ubuntu 22.04",
                "device_location": "Office Building A",
                "ip_address": "192.168.1.100",
                "mac_address": "00:1B:44:11:3A:B7",
                "current_user": "john.doe",
            }
        }
    )

    deviceid: str = Field(..., description="Unique device identifier (UUID format)")
    device_name: str | None = Field(None, description="Human-readable device name")
    device_type: str | None = Field(None, description="Type of device (laptop, desktop, server, etc.)")
    os: str | None = Field(None, description="Operating system name and version")
    device_location: str | None = Field(None, description="Physical location of the device")
    ip_address: str | None = Field(None, description="IP address of the device")
    mac_address: str | None = Field(None, description="MAC address of the device")
    current_user: str | None = Field(None, description="Currently logged in user")


class DeviceRegisterResponse(BaseModel):
    """Device registration response schema."""

    model_config = ConfigDict(
        json_schema_extra={"example": {"deviceid": "a843a399-701f-5011-aff3-4b69d8f21b11", "created": True}}
    )

    deviceid: str = Field(..., description="Device identifier")
    created: bool | None = Field(None, description="True if device was created")
    updated: bool | None = Field(None, description="True if device was updated")


# Device Metrics Schemas
class DeviceMetricsSubmit(BaseModel):
    """Device metrics submission schema."""

    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "cpu_usage": 45.5,
                "cpu_temp": 65.0,
                "memory_total": 16000000000,
                "memory_used": 8000000000,
                "swap_used": 100000000,
                "disk_total": 500000000000,
                "disk_used": 250000000000,
                "net_bytes_in": 1048576,
                "net_bytes_out": 524288,
            }
        }
    )

    cpu_usage: float | None = Field(None, description="CPU usage percentage (0-100)", ge=0, le=100)
    cpu_temp: float | None = Field(None, description="CPU temperature in Celsius", ge=-50, le=150)
    memory_total: int | None = Field(None, description="Total memory in bytes", ge=0)
    memory_used: int | None = Field(None, description="Used memory in bytes", ge=0)
    swap_used: int | None = Field(None, description="Used swap space in bytes", ge=0)
    disk_total: int | None = Field(None, description="Total disk space in bytes", ge=0)
    disk_used: int | None = Field(None, description="Used disk space in bytes", ge=0)
    net_bytes_in: int | None = Field(None, description="Network bytes received", ge=0)
    net_bytes_out: int | None = Field(None, description="Network bytes sent", ge=0)


class DeviceMetrics(BaseModel):
    """Device metrics response schema."""

    model_config = ConfigDict(from_attributes=True)

    metricid: str = Field(..., description="Unique metric identifier")
    deviceid: str = Field(..., description="Device identifier")
    timestamp: str = Field(..., description="Metric timestamp (ISO 8601)")
    cpu_usage: float | None = Field(None, description="CPU usage percentage")
    cpu_temp: float | None = Field(None, description="CPU temperature in Celsius")
    memory_total: int | None = Field(None, description="Total memory in bytes")
    memory_used: int | None = Field(None, description="Used memory in bytes")
    swap_used: int | None = Field(None, description="Used swap space in bytes")
    disk_total: int | None = Field(None, description="Total disk space in bytes")
    disk_used: int | None = Field(None, description="Used disk space in bytes")
    net_bytes_in: int | None = Field(None, description="Network bytes received")
    net_bytes_out: int | None = Field(None, description="Network bytes sent")


# Device Process Schemas
class ProcessSubmit(BaseModel):
    """Process submission schema."""

    model_config = ConfigDict(
        json_schema_extra={"example": {"pid": 1234, "process_name": "chrome", "cpu": 25.5, "memory": 512000000, "command_text": "/usr/bin/chrome --flag"}}
    )

    pid: int | None = Field(None, description="Process ID", ge=0)
    process_name: str | None = Field(None, description="Process name")
    cpu: float | None = Field(None, description="CPU usage percentage", ge=0, le=100)
    memory: int | None = Field(None, description="Memory usage in bytes", ge=0)
    command_text: str | None = Field(None, description="Full command line")


class DeviceProcess(BaseModel):
    """Device process response schema."""

    model_config = ConfigDict(from_attributes=True)

    processid: str = Field(..., description="Unique process record identifier")
    deviceid: str = Field(..., description="Device identifier")
    timestamp: str = Field(..., description="Process snapshot timestamp (ISO 8601)")
    pid: int | None = Field(None, description="Process ID")
    process_name: str = Field(..., description="Process name")
    name: str | None = Field(None, description="Process name (legacy alias)")
    cpu: float | None = Field(None, description="CPU usage percentage")
    memory: int | None = Field(None, description="Memory usage in bytes")
    command_text: str | None = Field(None, description="Full command line")


# Device Activity Schemas
class ActivitySubmit(BaseModel):
    """Activity submission schema."""

    model_config = ConfigDict(
        json_schema_extra={"example": {"activity_type": "file_access", "description": "Opened document.pdf", "app": "Adobe Reader", "duration": 300}}
    )

    activity_type: str | None = Field(None, description="Type of activity (file_access, app_launch, etc.)")
    description: str | None = Field(None, description="Activity description")
    app: str | None = Field(None, description="Application name")
    duration: int | None = Field(None, description="Activity duration in seconds", ge=0)


class DeviceActivity(BaseModel):
    """Device activity response schema."""

    model_config = ConfigDict(from_attributes=True)

    activityid: str = Field(..., description="Unique activity identifier")
    deviceid: str = Field(..., description="Device identifier")
    timestamp: str = Field(..., description="Activity timestamp (ISO 8601)")
    activity_type: str | None = Field(None, description="Type of activity")
    description: str | None = Field(None, description="Activity description")
    app: str | None = Field(None, description="Application name")
    duration: int | None = Field(None, description="Activity duration in seconds")


# Device Alert Schemas
class AlertSubmit(BaseModel):
    """Alert submission schema."""

    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "level": "warning",
                "alert_type": "high_cpu",
                "message": "CPU usage exceeded threshold",
                "value": 95.5,
                "threshold": 90.0,
            }
        }
    )

    level: Literal["info", "warning", "critical"] | None = Field(None, description="Alert severity level")
    alert_type: str | None = Field(None, description="Type of alert (high_cpu, low_memory, etc.)")
    message: str | None = Field(None, description="Alert message")
    value: float | None = Field(None, description="Current value that triggered the alert")
    threshold: float | None = Field(None, description="Threshold value")


class DeviceAlert(BaseModel):
    """Device alert response schema."""

    model_config = ConfigDict(from_attributes=True)

    alertid: str = Field(..., description="Unique alert identifier")
    deviceid: str = Field(..., description="Device identifier")
    timestamp: str = Field(..., description="Alert timestamp (ISO 8601)")
    level: str | None = Field(None, description="Alert severity level")
    alert_type: str | None = Field(None, description="Type of alert")
    message: str | None = Field(None, description="Alert message")
    value: float | None = Field(None, description="Current value")
    threshold: float | None = Field(None, description="Threshold value")


# Device Info Schemas
class DeviceInfo(BaseModel):
    """Device information response schema."""

    model_config = ConfigDict(
        from_attributes=True,
        json_schema_extra={
            "example": {
                "deviceid": "a843a399-701f-5011-aff3-4b69d8f21b11",
                "device_name": "Office Laptop",
                "device_type": "laptop",
                "os": "Ubuntu 22.04",
                "last_seen": "2025-11-16T12:00:00Z",
                "is_online": True,
                "device_location": "Office Building A",
                "ip_address": "192.168.1.100",
                "mac_address": "00:1B:44:11:3A:B7",
                "current_user": "john.doe",
            }
        },
    )

    deviceid: str = Field(..., description="Device identifier")
    id: str | None = Field(None, description="Device identifier (legacy alias)")
    device_name: str | None = Field(None, description="Device name")
    name: str | None = Field(None, description="Device name (legacy alias)")
    device_type: str | None = Field(None, description="Device type")
    os: str | None = Field(None, description="Operating system")
    last_seen: str | None = Field(None, description="Last seen timestamp (ISO 8601)")
    is_online: bool | None = Field(None, description="Online status")
    device_location: str | None = Field(None, description="Physical location")
    ip_address: str | None = Field(None, description="IP address")
    mac_address: str | None = Field(None, description="MAC address")
    current_user: str | None = Field(None, description="Current user")


# Screenshot Schemas
class DeviceScreenshot(BaseModel):
    """Device screenshot response schema."""

    model_config = ConfigDict(from_attributes=True)

    screenshotid: str = Field(..., description="Unique screenshot identifier")
    deviceid: str = Field(..., description="Device identifier")
    timestamp: str = Field(..., description="Screenshot timestamp (ISO 8601)")
    path: str = Field(..., description="Screenshot file path or URL")
    resolution: str | None = Field(None, description="Screenshot resolution (e.g., 1920x1080)")
    size: int | None = Field(None, description="File size in bytes")


# Standard Response Schemas
class StatusResponse(BaseModel):
    """Standard status response."""

    model_config = ConfigDict(json_schema_extra={"example": {"status": "ok"}})

    status: str = Field(..., description="Operation status")


class InsertedResponse(BaseModel):
    """Response for bulk insert operations."""

    model_config = ConfigDict(json_schema_extra={"example": {"inserted": 5}})

    inserted: int = Field(..., description="Number of records inserted")


class ErrorResponse(BaseModel):
    """Standard error response."""

    model_config = ConfigDict(json_schema_extra={"example": {"detail": "Error message describing what went wrong"}})

    detail: str = Field(..., description="Error message")
