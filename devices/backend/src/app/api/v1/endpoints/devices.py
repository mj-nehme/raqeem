import datetime
from typing import Any, cast
from uuid import UUID

import httpx
from app.core.config import settings
from app.db.session import get_db
from app.models import devices as dev_models
from app.schemas.commands import CommandCreate, CommandOut, CommandResultSubmit
from app.schemas.devices import (
    DeviceActivity,
    DeviceAlert,
    DeviceMetric,
    DeviceProcess,
    DeviceRegister,
    DeviceRegisterResponse,
    ErrorResponse,
    InsertResponse,
    SuccessResponse,
)
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

router = APIRouter()


@router.post(
    "/register",
    status_code=200,
    response_model=DeviceRegisterResponse,
    summary="Register or update a device",
    description="""
    Register a new device or update an existing device's information.
    
    **Key Points:**
    - Device ID must be a valid UUID
    - If device exists, updates its information
    - Automatically forwards registration to mentor backend (if configured)
    - Updates device's last_seen timestamp and online status
    
    **Legacy Field Handling:**
    - Legacy field 'id' → use 'deviceid'
    - Legacy field 'name' → use 'device_name'
    - Legacy field 'location' → use 'device_location'
    """,
    responses={
        200: {
            "description": "Device successfully registered or updated",
            "model": DeviceRegisterResponse,
        },
        400: {
            "description": "Invalid request (bad UUID, legacy fields, or missing deviceid)",
            "model": ErrorResponse,
        },
    },
)
async def register_device(payload: DeviceRegister, db: AsyncSession = Depends(get_db)):
    # Convert Pydantic model to dict for processing
    payload_dict = payload.model_dump()
    device_id = payload.deviceid
    final_id = device_id

    now = datetime.datetime.now(datetime.UTC)
    res = await db.execute(select(dev_models.Device).where(dev_models.Device.deviceid == final_id))
    existing = res.scalars().first()

    if existing:
        # update fields
        existing.device_name = payload.device_name or existing.device_name
        existing.device_type = payload.device_type or existing.device_type
        existing.os = payload.os or existing.os
        existing.last_seen = now  # type: ignore[assignment]
        existing.is_online = True  # type: ignore[assignment]
        existing.device_location = payload.device_location or existing.device_location
        existing.ip_address = payload.ip_address or existing.ip_address
        existing.mac_address = payload.mac_address or existing.mac_address
        existing.current_user = payload.current_user or existing.current_user
        db.add(existing)
        await db.commit()
        result = {"deviceid": final_id, "updated": True}
    else:
        obj = dev_models.Device(
            deviceid=device_id,
            device_name=payload.device_name,
            device_type=payload.device_type,
            os=payload.os,
            last_seen=now,
            is_online=True,
            device_location=payload.device_location,
            ip_address=payload.ip_address,
            mac_address=payload.mac_address,
            current_user=payload.current_user,
        )
        db.add(obj)
        await db.commit()
        result = {"deviceid": final_id, "created": True}

    if settings.mentor_api_url:
        fwd = payload_dict.copy()
        fwd["deviceid"] = str(final_id)
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                await client.post(f"{settings.mentor_api_url}/devices/register", json=fwd)
        except Exception:
            pass

    return DeviceRegisterResponse(deviceid=final_id, **result)


@router.post(
    "/{device_id}/metrics",
    response_model=SuccessResponse,
    summary="Submit device metrics",
    description="""
    Submit system metrics for a device including CPU, memory, disk, and network usage.
    
    **Key Points:**
    - All fields are optional
    - Metrics are automatically forwarded to mentor backend (if configured)
    - Timestamps are automatically recorded
    
    **Metrics Tracked:**
    - CPU: usage percentage and temperature
    - Memory: total, used, and swap
    - Disk: total and used space
    - Network: bytes in/out
    """,
    responses={
        200: {"description": "Metrics successfully recorded", "model": SuccessResponse},
        400: {"description": "Invalid request data", "model": ErrorResponse},
    },
)
async def post_metrics(device_id: str, payload: DeviceMetric, db: AsyncSession = Depends(get_db)):
    obj = dev_models.DeviceMetric(
        deviceid=device_id,
        cpu_usage=payload.cpu_usage,
        cpu_temp=payload.cpu_temp,
        memory_total=payload.memory_total,
        memory_used=payload.memory_used,
        swap_used=payload.swap_used,
        disk_total=payload.disk_total,
        disk_used=payload.disk_used,
        net_bytes_in=payload.net_bytes_in,
        net_bytes_out=payload.net_bytes_out,
    )
    db.add(obj)
    await db.commit()
    # Optionally forward metrics to mentor backend if configured
    if settings.mentor_api_url:
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                forward = payload.model_dump()
                forward["deviceid"] = device_id
                await client.post(f"{settings.mentor_api_url}/devices/metrics", json=forward)
        except Exception:
            # Do not fail ingestion if forwarding fails
            pass
    return SuccessResponse(status="ok")


@router.post(
    "/{device_id}/processes",
    response_model=InsertResponse,
    summary="Update device process list",
    description="""
    Replace the current process list for a device with a new snapshot.
    
    **Key Points:**
    - Replaces all existing processes for the device
    - Use for periodic process list updates
    - Automatically forwards to mentor backend (if configured)
    
    **Legacy Field Handling:**
    - Legacy field 'name' → use 'process_name'
    - Legacy field 'command' → use 'command_text'
    """,
    responses={
        200: {"description": "Process list successfully updated", "model": InsertResponse},
        400: {"description": "Invalid request data or legacy fields", "model": ErrorResponse},
    },
)
async def post_processes(device_id: str, processes: list[DeviceProcess], db: AsyncSession = Depends(get_db)):

    # delete existing processes for device, then insert new ones
    _proc_table = cast("Any", dev_models.DeviceProcess.__table__)
    await db.execute(_proc_table.delete().where(dev_models.DeviceProcess.deviceid == device_id))
    to_add = []
    now = datetime.datetime.now(datetime.UTC)
    for p in processes:
        p_dict = p.model_dump()
        to_add.append(
            {
                "deviceid": device_id,
                "pid": p_dict.get("pid"),
                "process_name": p_dict.get("process_name"),
                "cpu": p_dict.get("cpu"),
                "memory": p_dict.get("memory"),
                "command_text": p_dict.get("command_text"),
                "timestamp": now,
            }
        )
    if to_add:
        _proc_table = cast("Any", dev_models.DeviceProcess.__table__)
        await db.execute(_proc_table.insert(), to_add)
        await db.commit()
        # Optionally forward processes to mentor backend if configured
        if settings.mentor_api_url:
            try:
                async with httpx.AsyncClient(timeout=5.0) as client:
                    forward = [p.model_dump() for p in processes]
                    for proc in forward:
                        proc["deviceid"] = device_id
                    await client.post(f"{settings.mentor_api_url}/devices/processes", json=forward)
            except Exception:
                pass
    return InsertResponse(inserted=len(to_add))


@router.post(
    "/{device_id}/activities",
    response_model=InsertResponse,
    summary="Log device activities",
    description="""
    Submit one or more activity logs for a device.
    
    **Key Points:**
    - Activities track user actions and app usage
    - All activities are timestamped automatically
    - Automatically forwards to mentor backend (if configured)
    
    **Activity Types:**
    - app_usage: Application usage tracking
    - system_event: System-level events
    - user_action: User interactions
    
    **Legacy Field Handling:**
    - Legacy field 'type' → use 'activity_type'
    """,
    responses={
        200: {"description": "Activities successfully logged", "model": InsertResponse},
        400: {"description": "Invalid request data or legacy fields", "model": ErrorResponse},
        422: {"description": "Validation error", "model": ErrorResponse},
    },
)
async def post_activity(device_id: str, activities: list[DeviceActivity], db: AsyncSession = Depends(get_db)):

    to_add = []
    now = datetime.datetime.now(datetime.UTC)
    for a in activities:
        to_add.append(
            {
                "deviceid": device_id,
                "activity_type": a.get("activity_type"),
                "description": a.get("description"),
                "app": a.get("app"),
                "duration": a.get("duration"),
                "timestamp": now,
            }
        )
    if to_add:
        _act_table = cast("Any", dev_models.DeviceActivity.__table__)
        await db.execute(_act_table.insert(), to_add)
        await db.commit()
        # Optionally forward activities to mentor backend if configured
        if settings.mentor_api_url:
            try:
                async with httpx.AsyncClient(timeout=5.0) as client:
                    for a in activities:
                        forward = {
                            "deviceid": device_id,
                            "activity_type": a.get("activity_type"),
                            "description": a.get("description"),
                            "app": a.get("app"),
                            "duration": a.get("duration"),
                        }
                        await client.post(f"{settings.mentor_api_url}/devices/activity", json=forward)
            except Exception:
                pass
    return {"inserted": len(to_add)}


@router.post("/{device_id}/alerts")
async def post_alerts(device_id: str, alerts: list[dict], db: AsyncSession = Depends(get_db)):
    # Validate legacy fields and reject with clear error messages
    for a in alerts:
        if "type" in a:
            raise HTTPException(status_code=400, detail="unsupported legacy field: type; use alert_type")

    to_add = []
    now = datetime.datetime.now(datetime.UTC)
    for a in alerts:
        to_add.append(
            {
                "deviceid": device_id,
                "level": a.get("level"),
                "alert_type": a.get("alert_type"),
                "message": a.get("message"),
                "value": a.get("value"),
                "threshold": a.get("threshold"),
                "timestamp": now,
            }
        )
    if to_add:
        _alert_table = cast("Any", dev_models.DeviceAlert.__table__)
        await db.execute(_alert_table.insert(), to_add)
        await db.commit()
        # Optionally forward alerts to mentor backend if configured
        if settings.mentor_api_url:
            try:
                async with httpx.AsyncClient(timeout=5.0) as client:
                    for a in alerts:
                        payload = {
                            "deviceid": device_id,
                            "level": a.get("level"),
                            "alert_type": a.get("alert_type"),
                            "message": a.get("message"),
                            "value": a.get("value"),
                            "threshold": a.get("threshold"),
                        }
                        # Mentor API path accepts /devices/:id/alerts but uses JSON body for device_id
                        await client.post(f"{settings.mentor_api_url}/devices/{device_id}/alerts", json=payload)
            except Exception:
                # Swallow forwarding errors to avoid impacting device ingestion
                pass
    return {"inserted": len(to_add)}


@router.get("/")
async def list_devices(db: AsyncSession = Depends(get_db)):
    res = await db.execute(select(dev_models.Device))
    devices_list = res.scalars().all()
    # Convert SQLAlchemy models to dictionaries
    devices = []
    for device in devices_list:
        devices.append(
            {
                "deviceid": str(device.deviceid),
                "id": str(device.deviceid),  # legacy alias for compatibility
                "device_name": device.device_name,
                "device_type": device.device_type,
                "os": device.os,
                "last_seen": device.last_seen.isoformat() if device.last_seen else None,
                "is_online": device.is_online,
                "device_location": device.device_location,
                "ip_address": device.ip_address,
                "mac_address": device.mac_address,
                "current_user": device.current_user,
            }
        )
    return devices


@router.get("/processes")
async def list_all_processes(db: AsyncSession = Depends(get_db)):
    """Get all processes across all devices.

    Returns up to 1000 most recent processes ordered by timestamp descending.
    """
    res = await db.execute(
        select(dev_models.DeviceProcess)
        .order_by(dev_models.DeviceProcess.timestamp.desc())
        .limit(1000)  # Limit to prevent overwhelming the API
    )
    processes_list = res.scalars().all()
    processes = []
    for process in processes_list:
        processes.append(
            {
                "processid": str(process.processid),
                "deviceid": str(process.deviceid),
                "timestamp": process.timestamp.isoformat() if process.timestamp else None,
                "pid": process.pid,
                "process_name": process.process_name,
                "name": process.process_name,  # legacy alias for compatibility
                "cpu": float(process.cpu) if process.cpu is not None else None,
                "memory": process.memory,
                "command_text": process.command_text,
            }
        )
    return processes


@router.get("/activities")
async def list_all_activities(db: AsyncSession = Depends(get_db)):
    """Get all activities across all devices.

    Returns up to 1000 most recent activities ordered by timestamp descending.
    """
    res = await db.execute(
        select(dev_models.DeviceActivity)
        .order_by(dev_models.DeviceActivity.timestamp.desc())
        .limit(1000)  # Limit to prevent overwhelming the API
    )
    activities_list = res.scalars().all()
    activities = []
    for activity in activities_list:
        activities.append(
            {
                "activityid": str(activity.activityid),
                "deviceid": str(activity.deviceid),
                "timestamp": activity.timestamp.isoformat() if activity.timestamp else None,
                "activity_type": activity.activity_type,
                "description": activity.description,
                "app": activity.app,
                "duration": activity.duration,
            }
        )
    return activities


@router.get("/alerts")
async def list_all_alerts(db: AsyncSession = Depends(get_db)):
    """Get all alerts across all devices.

    Returns up to 1000 most recent alerts ordered by timestamp descending.
    """
    res = await db.execute(
        select(dev_models.DeviceAlert)
        .order_by(dev_models.DeviceAlert.timestamp.desc())
        .limit(1000)  # Limit to prevent overwhelming the API
    )
    alerts_list = res.scalars().all()
    alerts = []
    for alert in alerts_list:
        alerts.append(
            {
                "alertid": str(alert.alertid),
                "deviceid": str(alert.deviceid),
                "timestamp": alert.timestamp.isoformat() if alert.timestamp else None,
                "level": alert.level,
                "alert_type": alert.alert_type,
                "message": alert.message,
                "value": float(alert.value) if alert.value is not None else None,
                "threshold": float(alert.threshold) if alert.threshold is not None else None,
            }
        )
    return alerts


@router.get("/{device_id}")
async def get_device_by_id(device_id: str, db: AsyncSession = Depends(get_db)):
    """Get a specific device by ID."""
    # If not a valid UUID, treat as not found to match test expectations
    try:
        UUID(device_id)
    except Exception as e:
        raise HTTPException(status_code=404, detail="Device not found") from e

    res = await db.execute(select(dev_models.Device).where(dev_models.Device.deviceid == device_id))
    device = res.scalars().first()

    if not device:
        raise HTTPException(status_code=404, detail="Device not found")

    return {
        "deviceid": str(device.deviceid),
        "id": str(device.deviceid),  # legacy alias for compatibility
        "name": device.device_name,  # legacy alias for compatibility
        "device_name": device.device_name,
        "device_type": device.device_type,
        "os": device.os,
        "last_seen": device.last_seen.isoformat() if device.last_seen else None,
        "is_online": device.is_online,
        "device_location": device.device_location,
        "ip_address": device.ip_address,
        "mac_address": device.mac_address,
        "current_user": device.current_user,
    }


@router.get("/{device_id}/commands/pending", response_model=list[CommandOut])
async def get_pending_commands(device_id: str, db: AsyncSession = Depends(get_db)):
    """Get pending commands for a device"""
    res = await db.execute(
        select(dev_models.DeviceRemoteCommand)
        .where(dev_models.DeviceRemoteCommand.deviceid == device_id)
        .where(dev_models.DeviceRemoteCommand.status == "pending")
        .order_by(dev_models.DeviceRemoteCommand.created_at.asc())
    )
    return res.scalars().all()


@router.post("/commands/{command_id}/result")
async def submit_command_result(command_id: UUID, payload: CommandResultSubmit, db: AsyncSession = Depends(get_db)):
    """Submit command execution result"""
    res = await db.execute(
        select(dev_models.DeviceRemoteCommand).where(dev_models.DeviceRemoteCommand.commandid == command_id)
    )
    command = res.scalars().first()
    if not command:
        raise HTTPException(status_code=404, detail="Command not found")

    # Update command with result
    command.status = payload.status  # type: ignore[assignment]
    command.result = payload.result or ""  # type: ignore[assignment]
    command.exit_code = payload.exit_code or 0  # type: ignore[assignment]
    command.completed_at = datetime.datetime.now(datetime.UTC)  # type: ignore[assignment]
    db.add(command)
    await db.commit()

    # Forward result to mentor backend if configured
    if settings.mentor_api_url:
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                forward_payload = {
                    "id": str(command.commandid),
                    "status": command.status,
                    "result": command.result,
                    "exit_code": command.exit_code,
                }
                await client.post(f"{settings.mentor_api_url}/commands/status", json=forward_payload)
        except Exception:
            # Don't fail if forwarding fails
            pass

    return {"status": "ok", "commandid": str(command_id)}


@router.post("/{device_id}/commands", response_model=CommandOut)
async def create_command(device_id: str, payload: CommandCreate, db: AsyncSession = Depends(get_db)):
    """Create a new command for a device (forwarded from mentor backend)"""
    # Validate command against whitelist
    allowed_commands = ["get_info", "status", "restart", "get_processes", "get_logs", "restart_service", "screenshot"]
    command_base = payload.command_text.lower().split()[0] if payload.command_text else ""
    if command_base not in allowed_commands:
        raise HTTPException(
            status_code=400, detail=f"Command not allowed. Allowed commands: {', '.join(allowed_commands)}"
        )

    command = dev_models.DeviceRemoteCommand(
        deviceid=device_id,
        command_text=payload.command_text,
        status="pending",
        created_at=datetime.datetime.now(datetime.UTC),
    )
    db.add(command)
    await db.commit()
    await db.refresh(command)

    return command


@router.get("/{device_id}/metrics")
async def get_device_metrics(device_id: str, limit: int = 60, db: AsyncSession = Depends(get_db)):
    """Get recent metrics for a specific device.

    Args:
        device_id: Device identifier
        limit: Number of records to return (default: 60)

    Returns up to 'limit' most recent metrics ordered by timestamp descending.
    """
    res = await db.execute(
        select(dev_models.DeviceMetric)
        .where(dev_models.DeviceMetric.deviceid == device_id)
        .order_by(dev_models.DeviceMetric.timestamp.desc())
        .limit(limit)
    )
    metrics_list = res.scalars().all()
    metrics = []
    for metric in metrics_list:
        metrics.append(
            {
                "metricid": str(metric.metricid),
                "deviceid": str(metric.deviceid),
                "timestamp": metric.timestamp.isoformat() if metric.timestamp else None,
                "cpu_usage": float(metric.cpu_usage) if metric.cpu_usage is not None else None,
                "cpu_temp": float(metric.cpu_temp) if metric.cpu_temp is not None else None,
                "memory_total": metric.memory_total,
                "memory_used": metric.memory_used,
                "swap_used": metric.swap_used,
                "disk_total": metric.disk_total,
                "disk_used": metric.disk_used,
                "net_bytes_in": metric.net_bytes_in,
                "net_bytes_out": metric.net_bytes_out,
            }
        )
    return metrics


@router.get("/{device_id}/processes")
async def get_device_processes(device_id: str, limit: int = 100, db: AsyncSession = Depends(get_db)):
    """Get latest known processes for a specific device.

    Args:
        device_id: Device identifier
        limit: Number of records to return (default: 100)

    Returns up to 'limit' most recent processes ordered by timestamp descending.
    """
    res = await db.execute(
        select(dev_models.DeviceProcess)
        .where(dev_models.DeviceProcess.deviceid == device_id)
        .order_by(dev_models.DeviceProcess.timestamp.desc())
        .limit(limit)
    )
    processes_list = res.scalars().all()
    processes = []
    for process in processes_list:
        processes.append(
            {
                "processid": str(process.processid),
                "deviceid": str(process.deviceid),
                "timestamp": process.timestamp.isoformat() if process.timestamp else None,
                "pid": process.pid,
                "process_name": process.process_name,
                "cpu": float(process.cpu) if process.cpu is not None else None,
                "memory": process.memory,
                "command_text": process.command_text,
            }
        )
    return processes


@router.get("/{device_id}/activities")
async def get_device_activities(device_id: str, limit: int = 100, db: AsyncSession = Depends(get_db)):
    """Get recent activity logs for a specific device.

    Args:
        device_id: Device identifier
        limit: Number of records to return (default: 100)

    Returns up to 'limit' most recent activities ordered by timestamp descending.
    """
    res = await db.execute(
        select(dev_models.DeviceActivity)
        .where(dev_models.DeviceActivity.deviceid == device_id)
        .order_by(dev_models.DeviceActivity.timestamp.desc())
        .limit(limit)
    )
    activities_list = res.scalars().all()
    activities = []
    for activity in activities_list:
        activities.append(
            {
                "activityid": str(activity.activityid),
                "deviceid": str(activity.deviceid),
                "timestamp": activity.timestamp.isoformat() if activity.timestamp else None,
                "activity_type": activity.activity_type,
                "description": activity.description,
                "app": activity.app,
                "duration": activity.duration,
            }
        )
    return activities


@router.get("/{device_id}/alerts")
async def get_device_alerts(device_id: str, limit: int = 100, db: AsyncSession = Depends(get_db)):
    """Get recent alerts for a specific device.

    Args:
        device_id: Device identifier
        limit: Number of records to return (default: 100)

    Returns up to 'limit' most recent alerts ordered by timestamp descending.
    """
    res = await db.execute(
        select(dev_models.DeviceAlert)
        .where(dev_models.DeviceAlert.deviceid == device_id)
        .order_by(dev_models.DeviceAlert.timestamp.desc())
        .limit(limit)
    )
    alerts_list = res.scalars().all()
    alerts = []
    for alert in alerts_list:
        alerts.append(
            {
                "alertid": str(alert.alertid),
                "deviceid": str(alert.deviceid),
                "timestamp": alert.timestamp.isoformat() if alert.timestamp else None,
                "level": alert.level,
                "alert_type": alert.alert_type,
                "message": alert.message,
                "value": float(alert.value) if alert.value is not None else None,
                "threshold": float(alert.threshold) if alert.threshold is not None else None,
            }
        )
    return alerts


@router.get("/{device_id}/screenshots")
async def get_device_screenshots(device_id: str, limit: int = 50, db: AsyncSession = Depends(get_db)):
    """Get recent screenshots metadata for a specific device.

    Args:
        device_id: Device identifier
        limit: Number of records to return (default: 50)

    Returns up to 'limit' most recent screenshots ordered by timestamp descending.
    """
    res = await db.execute(
        select(dev_models.DeviceScreenshot)
        .where(dev_models.DeviceScreenshot.deviceid == device_id)
        .order_by(dev_models.DeviceScreenshot.timestamp.desc())
        .limit(limit)
    )
    screenshots_list = res.scalars().all()
    screenshots = []
    for screenshot in screenshots_list:
        screenshots.append(
            {
                "screenshotid": str(screenshot.screenshotid),
                "deviceid": str(screenshot.deviceid),
                "timestamp": screenshot.timestamp.isoformat() if screenshot.timestamp else None,
                "path": screenshot.path,
                "resolution": screenshot.resolution,
                "size": screenshot.size,
            }
        )
    return screenshots


@router.get("/{device_id}/commands")
async def get_device_commands(device_id: str, limit: int = 100, db: AsyncSession = Depends(get_db)):
    """Get command history for a specific device.

    Args:
        device_id: Device identifier
        limit: Number of records to return (default: 100)

    Returns up to 'limit' most recent commands ordered by creation time descending.
    """
    res = await db.execute(
        select(dev_models.DeviceRemoteCommand)
        .where(dev_models.DeviceRemoteCommand.deviceid == device_id)
        .order_by(dev_models.DeviceRemoteCommand.created_at.desc())
        .limit(limit)
    )
    return res.scalars().all()
