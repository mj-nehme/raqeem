import contextlib
import datetime
from typing import Any, cast
from uuid import UUID

from app.core.config import settings
from app.db.session import get_db
from app.models import devices as dev_models
from app.schemas.commands import CommandCreate, CommandOut, CommandResultResponse, CommandResultSubmit
from app.schemas.devices import (
    DeviceRegisterResponse,
    ErrorResponse,
    InsertedResponse,
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

    now = datetime.datetime.now(datetime.timezone.utc)
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
        result = {"deviceid": str(final_id), "updated": True}
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
        result = {"deviceid": str(final_id), "created": True}

    # Forward registration to mentor backend if configured (best-effort, non-blocking)
    if settings.mentor_api_url:
        fwd = dict(payload)
        fwd["deviceid"] = str(final_id)
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
    now = datetime.datetime.now(datetime.timezone.utc)
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
    now = datetime.datetime.now(datetime.timezone.utc)
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
    now = datetime.datetime.now(datetime.timezone.utc)
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


# NOTE: Read-only endpoints (list_devices, list_all_processes, list_all_activities,
# list_all_alerts, get_device_by_id) have been moved to mentor backend.
# Devices backend is write-focused and only provides command-related read endpoints.


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
    response_model=CommandResultResponse,
    responses={
        200: {
            "description": "Command result submitted successfully",
            "model": CommandResultResponse,
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
    # Diagnostic: log incoming command result
    with contextlib.suppress(Exception):
        print(f"[devices] submit_command_result: id={command_id} status={payload.status} exit_code={payload.exit_code}")
    res = await db.execute(
        select(dev_models.DeviceRemoteCommand).where(dev_models.DeviceRemoteCommand.commandid == command_id)
    )
    command = res.scalars().first()
    if not command:
        with contextlib.suppress(Exception):
            print(f"[devices] submit_command_result: command not found id={command_id}")
        raise HTTPException(status_code=404, detail="Command not found")

    # Update command with result
    command.status = payload.status  # type: ignore[assignment]
    command.result = payload.result or ""  # type: ignore[assignment]
    command.exit_code = payload.exit_code or 0  # type: ignore[assignment]
    command.completed_at = datetime.datetime.now(datetime.timezone.utc)  # type: ignore[assignment]
    db.add(command)
    await db.commit()

    # Forward result to mentor backend if configured
    if settings.mentor_api_url:
        # Use unified field name 'commandid' for consistency with mentor backend
        forward_payload = {
            "commandid": str(command.commandid),
            "status": command.status,
            "result": command.result,
            "exit_code": command.exit_code,
        }
        with contextlib.suppress(Exception):
            print(f"[devices] forwarding result to mentor: {forward_payload}")
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
        created_at=datetime.datetime.now(datetime.timezone.utc),
    )
    db.add(command)
    await db.commit()
    await db.refresh(command)

    return command


# NOTE: Device-specific read endpoints (get_device_metrics, get_device_processes,
# get_device_activities, get_device_alerts, get_device_screenshots, get_device_commands)
# have been moved to mentor backend. Devices backend is write-focused.
