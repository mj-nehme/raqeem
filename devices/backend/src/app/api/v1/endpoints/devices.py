import datetime
from typing import Any, cast
from uuid import UUID

from app.core.config import settings
from app.db.session import get_db
from app.models import devices as dev_models
from app.schemas.commands import CommandCreate, CommandOut, CommandResultSubmit
from app.schemas.devices import (
    ActivitySubmit,
    AlertSubmit,
    DeviceActivity,
    DeviceAlert,
    DeviceInfo,
    DeviceMetrics,
    DeviceMetricsSubmit,
    DeviceProcess,
    DeviceRegister,
    DeviceRegisterResponse,
    DeviceScreenshot,
    ErrorResponse,
    InsertedResponse,
    ProcessSubmit,
    StatusResponse,
)
from app.util import post_with_retry
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

router = APIRouter()


@router.post(
    "/register",
    status_code=200,
    response_model=DeviceRegisterResponse,
    responses={
        200: {
            "description": "Device registered or updated successfully",
            "model": DeviceRegisterResponse,
        },
        400: {
            "description": "Bad request - validation error or legacy fields used",
            "model": ErrorResponse,
        },
        500: {
            "description": "Internal server error during database operation",
            "model": ErrorResponse,
        },
    },
    summary="Register or update a device",
    description="""
    Register a new device or update an existing device's information.
    
    This endpoint performs an upsert operation:
    - If the device exists (by deviceid), it updates the provided fields
    - If it doesn't exist, it creates a new device record
    
    **Features:**
    - Automatically forwards registration to mentor backend if configured
    - Updates last_seen timestamp and sets device as online
    - Validates against legacy field names for backwards compatibility
    
    **Legacy Field Handling:**
    - `id` → use `deviceid` instead
    - `name` → use `device_name` instead  
    - `location` → use `device_location` instead
    """,
    tags=["Device Registration"],
)
async def register_device(payload: dict, db: AsyncSession = Depends(get_db)):
    # Validate legacy fields and reject with clear error messages
    if "id" in payload:
        raise HTTPException(status_code=400, detail="unsupported legacy field: id; use deviceid")
    if "name" in payload:
        raise HTTPException(status_code=400, detail="unsupported legacy field: name; use device_name")
    if "location" in payload:
        raise HTTPException(status_code=400, detail="unsupported legacy field: location; use device_location")

    # payload expected to contain deviceid and optional fields
    device_id = payload.get("deviceid")
    if not device_id:
        raise HTTPException(status_code=400, detail="missing required field: deviceid")

    # Validate that deviceid is a valid UUID
    try:
        final_id = UUID(str(device_id))
    except (ValueError, AttributeError, TypeError) as e:
        raise HTTPException(status_code=400, detail=f"deviceid must be a valid UUID format: {e!s}") from e

    now = datetime.datetime.now(datetime.UTC)
    res = await db.execute(select(dev_models.Device).where(dev_models.Device.deviceid == final_id))
    existing = res.scalars().first()

    if existing:
        # update fields
        existing.device_name = payload.get("device_name") or existing.device_name
        existing.device_type = payload.get("device_type") or existing.device_type
        existing.os = payload.get("os") or existing.os
        existing.last_seen = now  # type: ignore[assignment]
        existing.is_online = True  # type: ignore[assignment]
        existing.device_location = payload.get("device_location") or existing.device_location
        existing.ip_address = payload.get("ip_address") or existing.ip_address
        existing.mac_address = payload.get("mac_address") or existing.mac_address
        existing.current_user = payload.get("current_user") or existing.current_user
        db.add(existing)
        await db.commit()
        result = {"deviceid": final_id, "updated": True}
    else:
        obj = dev_models.Device(
            deviceid=device_id,
            device_name=payload.get("device_name"),
            device_type=payload.get("device_type"),
            os=payload.get("os"),
            last_seen=now,
            is_online=True,
            device_location=payload.get("device_location"),
            ip_address=payload.get("ip_address"),
            mac_address=payload.get("mac_address"),
            current_user=payload.get("current_user"),
        )
        db.add(obj)
        await db.commit()
        result = {"deviceid": final_id, "created": True}

    # Forward registration to mentor backend if configured (best-effort, non-blocking)
    if settings.mentor_api_url:
        fwd = dict(payload)
        fwd["deviceid"] = final_id
        # Use retry logic for forwarding to mentor backend
        await post_with_retry(
            f"{settings.mentor_api_url}/devices/register",
            json=fwd,
            max_retries=3,
            timeout=5.0,
        )

    return result


@router.post(
    "/{device_id}/metrics",
    response_model=StatusResponse,
    responses={
        200: {
            "description": "Metrics stored successfully",
            "model": StatusResponse,
        },
        400: {
            "description": "Bad request - invalid device ID or metrics data",
            "model": ErrorResponse,
        },
        500: {
            "description": "Internal server error during database operation",
            "model": ErrorResponse,
        },
    },
    summary="Submit device performance metrics",
    description="""
    Store device performance metrics for monitoring and analysis.
    
    Ingests and stores metrics such as:
    - CPU usage and temperature
    - Memory and swap usage
    - Disk space usage
    - Network traffic (bytes in/out)
    
    **Features:**
    - Automatically forwards metrics to mentor backend if configured
    - All metric fields are optional
    - Metrics are timestamped server-side upon ingestion
    """,
    tags=["Device Metrics"],
)
async def post_metrics(device_id: str, payload: dict, db: AsyncSession = Depends(get_db)):
    obj = dev_models.DeviceMetric(
        deviceid=device_id,
        cpu_usage=payload.get("cpu_usage"),
        cpu_temp=payload.get("cpu_temp"),
        memory_total=payload.get("memory_total"),
        memory_used=payload.get("memory_used"),
        swap_used=payload.get("swap_used"),
        disk_total=payload.get("disk_total"),
        disk_used=payload.get("disk_used"),
        net_bytes_in=payload.get("net_bytes_in"),
        net_bytes_out=payload.get("net_bytes_out"),
    )
    db.add(obj)
    await db.commit()
    # Optionally forward metrics to mentor backend if configured
    if settings.mentor_api_url:
        forward = {
            "deviceid": device_id,
            "cpu_usage": payload.get("cpu_usage"),
            "cpu_temp": payload.get("cpu_temp"),
            "memory_total": payload.get("memory_total"),
            "memory_used": payload.get("memory_used"),
            "swap_used": payload.get("swap_used"),
            "disk_total": payload.get("disk_total"),
            "disk_used": payload.get("disk_used"),
            "net_bytes_in": payload.get("net_bytes_in"),
            "net_bytes_out": payload.get("net_bytes_out"),
        }
        await post_with_retry(
            f"{settings.mentor_api_url}/devices/metrics",
            json=forward,
            max_retries=2,
        )
    return {"status": "ok"}


@router.post(
    "/{device_id}/processes",
    response_model=InsertedResponse,
    responses={
        200: {
            "description": "Process list updated successfully",
            "model": InsertedResponse,
        },
        400: {
            "description": "Bad request - legacy fields or invalid data",
            "model": ErrorResponse,
        },
        500: {
            "description": "Internal server error during database operation",
            "model": ErrorResponse,
        },
    },
    summary="Update device process list",
    description="""
    Update the current process list for a device.
    
    This endpoint:
    - Replaces the existing process list with the new snapshot
    - Stores process information including PID, name, CPU, memory, and command
    - Forwards the process list to mentor backend if configured
    
    **Legacy Field Handling:**
    - `name` → use `process_name` instead
    - `command` → use `command_text` instead
    """,
    tags=["Device Processes"],
)
async def post_processes(device_id: str, processes: list[dict], db: AsyncSession = Depends(get_db)):
    # Validate legacy fields and reject with clear error messages
    for p in processes:
        if "name" in p:
            raise HTTPException(status_code=400, detail="unsupported legacy field: name; use process_name")
        if "command" in p:
            raise HTTPException(status_code=400, detail="unsupported legacy field: command; use command_text")

    # delete existing processes for device, then insert new ones
    _proc_table = cast("Any", dev_models.DeviceProcess.__table__)
    await db.execute(_proc_table.delete().where(dev_models.DeviceProcess.deviceid == device_id))
    to_add = []
    now = datetime.datetime.now(datetime.UTC)
    for p in processes:
        to_add.append(
            {
                "deviceid": device_id,
                "pid": p.get("pid"),
                "process_name": p.get("process_name"),
                "cpu": p.get("cpu"),
                "memory": p.get("memory"),
                "command_text": p.get("command_text"),
                "timestamp": now,
            }
        )
    if to_add:
        _proc_table = cast("Any", dev_models.DeviceProcess.__table__)
        await db.execute(_proc_table.insert(), to_add)
        await db.commit()
        # Optionally forward processes to mentor backend if configured
        if settings.mentor_api_url:
            forward = [
                {
                    "deviceid": device_id,
                    "pid": p.get("pid"),
                    "process_name": p.get("process_name"),
                    "cpu": p.get("cpu"),
                    "memory": p.get("memory"),
                    "command_text": p.get("command_text"),
                }
                for p in processes
            ]
            await post_with_retry(
                f"{settings.mentor_api_url}/devices/processes",
                json=forward,
                max_retries=2,
            )
    return {"inserted": len(to_add)}


@router.post(
    "/{device_id}/activities",
    response_model=InsertedResponse,
    responses={
        200: {
            "description": "Activities logged successfully",
            "model": InsertedResponse,
        },
        400: {
            "description": "Bad request - legacy fields or invalid data",
            "model": ErrorResponse,
        },
        422: {
            "description": "Validation error - invalid field usage",
            "model": ErrorResponse,
        },
        500: {
            "description": "Internal server error during database operation",
            "model": ErrorResponse,
        },
    },
    summary="Log device activities",
    description="""
    Log user activities on the device.
    
    Records activities such as:
    - File access and modifications
    - Application launches and usage
    - User sessions and interactions
    
    **Features:**
    - Timestamps are set server-side
    - Forwards activities to mentor backend if configured
    - Supports duration tracking for time-based activities
    
    **Legacy Field Handling:**
    - `type` → use `activity_type` instead
    """,
    tags=["Device Activities"],
)
async def post_activity(device_id: str, activities: list[dict], db: AsyncSession = Depends(get_db)):
    # If legacy field 'type' is provided, treat as validation issue (422) instead of 400
    for a in activities:
        if "type" in a and not a.get("activity_type"):
            # Non-empty legacy 'type' should be rejected as bad request (400)
            if (a.get("type") or "") != "":
                raise HTTPException(status_code=400, detail="unsupported legacy field: type; use activity_type")
            # Empty legacy 'type' is treated as validation error (422)
            raise HTTPException(status_code=422, detail="invalid field: use activity_type instead of type")

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
            for a in activities:
                forward = {
                    "deviceid": device_id,
                    "activity_type": a.get("activity_type"),
                    "description": a.get("description"),
                    "app": a.get("app"),
                    "duration": a.get("duration"),
                }
                await post_with_retry(
                    f"{settings.mentor_api_url}/devices/activity",
                    json=forward,
                    max_retries=2,
                )
    return {"inserted": len(to_add)}


@router.post(
    "/{device_id}/alerts",
    response_model=InsertedResponse,
    responses={
        200: {
            "description": "Alerts submitted successfully",
            "model": InsertedResponse,
        },
        400: {
            "description": "Bad request - legacy fields or invalid data",
            "model": ErrorResponse,
        },
        500: {
            "description": "Internal server error during database operation",
            "model": ErrorResponse,
        },
    },
    summary="Submit device alerts",
    description="""
    Submit alerts triggered by device monitoring thresholds.
    
    Supports alert types:
    - Performance alerts (high CPU, low memory, disk space)
    - Temperature warnings
    - Network connectivity issues
    - Custom application alerts
    
    **Alert Levels:**
    - `info`: Informational messages
    - `warning`: Warning conditions
    - `critical`: Critical issues requiring immediate attention
    
    **Features:**
    - Timestamps are set server-side
    - Forwards alerts to mentor backend if configured
    - Includes current value and threshold for context
    
    **Legacy Field Handling:**
    - `type` → use `alert_type` instead
    """,
    tags=["Device Alerts"],
)
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
                await post_with_retry(
                    f"{settings.mentor_api_url}/devices/{device_id}/alerts",
                    json=payload,
                    max_retries=2,
                )
    return {"inserted": len(to_add)}


@router.get(
    "/",
    response_model=list[DeviceInfo],
    responses={
        200: {
            "description": "List of all registered devices",
            "model": list[DeviceInfo],
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="List all devices",
    description="""
    Get a list of all registered devices with their current status.
    
    Returns device information including:
    - Device identifiers and names
    - Online status and last seen timestamp
    - Location and network information
    - Current logged-in user
    
    **Note:** Both new (`deviceid`) and legacy (`id`) identifiers are included for backwards compatibility.
    """,
    tags=["Device Information"],
)
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


@router.get(
    "/processes",
    response_model=list[DeviceProcess],
    responses={
        200: {
            "description": "List of recent processes across all devices",
            "model": list[DeviceProcess],
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="List all processes across devices",
    description="""
    Get all processes across all devices.
    
    Returns up to 1000 most recent processes ordered by timestamp descending.
    
    Useful for:
    - System-wide process monitoring
    - Security auditing
    - Resource usage analysis
    
    **Note:** For device-specific processes, use GET /devices/{device_id}/processes
    """,
    tags=["Device Processes"],
)
async def list_all_processes(db: AsyncSession = Depends(get_db)):
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


@router.get(
    "/activities",
    response_model=list[DeviceActivity],
    responses={
        200: {
            "description": "List of recent activities across all devices",
            "model": list[DeviceActivity],
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="List all activities across devices",
    description="""
    Get all activities across all devices.
    
    Returns up to 1000 most recent activities ordered by timestamp descending.
    
    Useful for:
    - User behavior analysis
    - Security monitoring
    - Compliance auditing
    
    **Note:** For device-specific activities, use GET /devices/{device_id}/activities
    """,
    tags=["Device Activities"],
)
async def list_all_activities(db: AsyncSession = Depends(get_db)):
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


@router.get(
    "/alerts",
    response_model=list[DeviceAlert],
    responses={
        200: {
            "description": "List of recent alerts across all devices",
            "model": list[DeviceAlert],
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="List all alerts across devices",
    description="""
    Get all alerts across all devices.
    
    Returns up to 1000 most recent alerts ordered by timestamp descending.
    
    Useful for:
    - System-wide monitoring dashboard
    - Alert aggregation and analysis
    - Incident response
    
    **Note:** For device-specific alerts, use GET /devices/{device_id}/alerts
    """,
    tags=["Device Alerts"],
)
async def list_all_alerts(db: AsyncSession = Depends(get_db)):
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


@router.get(
    "/{device_id}",
    response_model=DeviceInfo,
    responses={
        200: {
            "description": "Device information",
            "model": DeviceInfo,
        },
        404: {
            "description": "Device not found",
            "model": ErrorResponse,
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="Get device by ID",
    description="""
    Get detailed information about a specific device.
    
    Returns:
    - Device identifiers and configuration
    - Online status and last seen timestamp
    - Location and network information
    - Current logged-in user
    
    **Note:** Device ID must be a valid UUID format.
    """,
    tags=["Device Information"],
)
async def get_device_by_id(device_id: str, db: AsyncSession = Depends(get_db)):
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


@router.get(
    "/{device_id}/commands/pending",
    response_model=list[CommandOut],
    responses={
        200: {
            "description": "List of pending commands for the device",
            "model": list[CommandOut],
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="Get pending commands for a device",
    description="""
    Retrieve all pending remote commands for a specific device.
    
    Commands are ordered by creation time (oldest first) to ensure
    proper execution order. Devices should poll this endpoint
    periodically to check for new commands.
    
    **Workflow:**
    1. Device polls this endpoint
    2. Device executes commands in order
    3. Device submits results via POST /devices/commands/{command_id}/result
    """,
    tags=["Device Commands"],
)
async def get_pending_commands(device_id: str, db: AsyncSession = Depends(get_db)):
    res = await db.execute(
        select(dev_models.DeviceRemoteCommand)
        .where(dev_models.DeviceRemoteCommand.deviceid == device_id)
        .where(dev_models.DeviceRemoteCommand.status == "pending")
        .order_by(dev_models.DeviceRemoteCommand.created_at.asc())
    )
    return res.scalars().all()


@router.post(
    "/commands/{command_id}/result",
    response_model=StatusResponse,
    responses={
        200: {
            "description": "Command result submitted successfully",
            "model": StatusResponse,
        },
        404: {
            "description": "Command not found",
            "model": ErrorResponse,
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="Submit command execution result",
    description="""
    Submit the result of a remote command execution.
    
    Devices should call this endpoint after executing a command
    retrieved from GET /devices/{device_id}/commands/pending.
    
    **Status Values:**
    - `completed`: Command executed successfully
    - `failed`: Command execution failed
    - `running`: Command is still executing (not typically used)
    
    **Features:**
    - Records execution result and exit code
    - Updates command completion timestamp
    - Forwards result to mentor backend if configured
    """,
    tags=["Device Commands"],
)
async def submit_command_result(command_id: UUID, payload: CommandResultSubmit, db: AsyncSession = Depends(get_db)):
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
        forward_payload = {
            "id": str(command.commandid),
            "status": command.status,
            "result": command.result,
            "exit_code": command.exit_code,
        }
        await post_with_retry(
            f"{settings.mentor_api_url}/commands/status",
            json=forward_payload,
            max_retries=2,
        )

    return {"status": "ok", "commandid": str(command_id)}


@router.post(
    "/{device_id}/commands",
    response_model=CommandOut,
    responses={
        200: {
            "description": "Command created successfully",
            "model": CommandOut,
        },
        400: {
            "description": "Command not allowed or invalid",
            "model": ErrorResponse,
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="Create a remote command for a device",
    description="""
    Create a new remote command for a device to execute.
    
    This endpoint is typically called by the mentor backend to
    send commands to devices. The command will appear in the
    device's pending commands list.
    
    **Allowed Commands:**
    - `get_info`: Get device information
    - `status`: Get device status
    - `restart`: Restart the device
    - `get_processes`: Get running processes
    - `get_logs`: Retrieve logs
    - `restart_service`: Restart a specific service
    - `screenshot`: Take a screenshot
    
    **Security:** Only whitelisted commands are accepted to prevent
    arbitrary command execution.
    """,
    tags=["Device Commands"],
)
async def create_command(device_id: str, payload: CommandCreate, db: AsyncSession = Depends(get_db)):
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


@router.get(
    "/{device_id}/metrics",
    response_model=list[DeviceMetrics],
    responses={
        200: {
            "description": "List of recent metrics for the device",
            "model": list[DeviceMetrics],
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="Get device metrics",
    description="""
    Get recent performance metrics for a specific device.
    
    Returns up to `limit` most recent metrics ordered by timestamp descending.
    
    **Default limit:** 60 (approximately 1 hour of data if metrics are sent every minute)
    
    Metrics include:
    - CPU usage and temperature
    - Memory and swap usage
    - Disk space usage
    - Network traffic
    """,
    tags=["Device Metrics"],
)
async def get_device_metrics(device_id: str, limit: int = 60, db: AsyncSession = Depends(get_db)):
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


@router.get(
    "/{device_id}/processes",
    response_model=list[DeviceProcess],
    responses={
        200: {
            "description": "List of recent processes for the device",
            "model": list[DeviceProcess],
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="Get device processes",
    description="""
    Get the latest known process list for a specific device.
    
    Returns up to `limit` most recent process records ordered by timestamp descending.
    
    **Default limit:** 100 processes
    
    Process information includes:
    - Process ID (PID)
    - Process name and full command
    - CPU usage percentage
    - Memory usage in bytes
    """,
    tags=["Device Processes"],
)
async def get_device_processes(device_id: str, limit: int = 100, db: AsyncSession = Depends(get_db)):
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


@router.get(
    "/{device_id}/activities",
    response_model=list[DeviceActivity],
    responses={
        200: {
            "description": "List of recent activities for the device",
            "model": list[DeviceActivity],
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="Get device activities",
    description="""
    Get recent activity logs for a specific device.
    
    Returns up to `limit` most recent activities ordered by timestamp descending.
    
    **Default limit:** 100 activities
    
    Activity information includes:
    - Activity type (file_access, app_launch, etc.)
    - Description of the activity
    - Associated application
    - Duration (if applicable)
    """,
    tags=["Device Activities"],
)
async def get_device_activities(device_id: str, limit: int = 100, db: AsyncSession = Depends(get_db)):
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


@router.get(
    "/{device_id}/alerts",
    response_model=list[DeviceAlert],
    responses={
        200: {
            "description": "List of recent alerts for the device",
            "model": list[DeviceAlert],
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="Get device alerts",
    description="""
    Get recent alerts for a specific device.
    
    Returns up to `limit` most recent alerts ordered by timestamp descending.
    
    **Default limit:** 100 alerts
    
    Alert information includes:
    - Alert level (info, warning, critical)
    - Alert type (high_cpu, low_memory, etc.)
    - Alert message
    - Current value and threshold
    """,
    tags=["Device Alerts"],
)
async def get_device_alerts(device_id: str, limit: int = 100, db: AsyncSession = Depends(get_db)):
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


@router.get(
    "/{device_id}/screenshots",
    response_model=list[DeviceScreenshot],
    responses={
        200: {
            "description": "List of recent screenshots for the device",
            "model": list[DeviceScreenshot],
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="Get device screenshots metadata",
    description="""
    Get recent screenshot metadata for a specific device.
    
    Returns up to `limit` most recent screenshot records ordered by timestamp descending.
    
    **Default limit:** 50 screenshots
    
    Screenshot metadata includes:
    - Screenshot identifier
    - File path or URL
    - Resolution
    - File size in bytes
    - Timestamp when screenshot was taken
    
    **Note:** This endpoint returns metadata only. To upload screenshots,
    use POST /api/v1/screenshots/
    """,
    tags=["Device Screenshots"],
)
async def get_device_screenshots(device_id: str, limit: int = 50, db: AsyncSession = Depends(get_db)):
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


@router.get(
    "/{device_id}/commands",
    response_model=list[CommandOut],
    responses={
        200: {
            "description": "List of command history for the device",
            "model": list[CommandOut],
        },
        500: {
            "description": "Internal server error",
            "model": ErrorResponse,
        },
    },
    summary="Get device command history",
    description="""
    Get command execution history for a specific device.
    
    Returns up to `limit` most recent commands ordered by creation time descending.
    
    **Default limit:** 100 commands
    
    Command information includes:
    - Command identifier
    - Command text
    - Status (pending, completed, failed)
    - Creation and completion timestamps
    - Execution result and exit code (if completed)
    
    **Status Values:**
    - `pending`: Command waiting to be executed
    - `completed`: Command executed successfully
    - `failed`: Command execution failed
    """,
    tags=["Device Commands"],
)
async def get_device_commands(device_id: str, limit: int = 100, db: AsyncSession = Depends(get_db)):
    res = await db.execute(
        select(dev_models.DeviceRemoteCommand)
        .where(dev_models.DeviceRemoteCommand.deviceid == device_id)
        .order_by(dev_models.DeviceRemoteCommand.created_at.desc())
        .limit(limit)
    )
    return res.scalars().all()
