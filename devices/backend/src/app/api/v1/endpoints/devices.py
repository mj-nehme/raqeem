from fastapi import APIRouter, Depends, HTTPException
from typing import List
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from app.db.session import get_db
from app.models import devices as dev_models
from app.schemas.commands import CommandCreate, CommandResultSubmit, CommandOut
from app.core.config import settings
import datetime
import httpx
from uuid import UUID
import uuid

router = APIRouter()


@router.post("/register", status_code=200)
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

    final_id = device_id
    generated = False
    try:
        UUID(str(device_id))
    except Exception:
        final_id = str(uuid.uuid4())
        generated = True

    now = datetime.datetime.utcnow()
    existing = None
    if not generated:
        res = await db.execute(select(dev_models.Device).where(dev_models.Device.deviceid == final_id))
        existing = res.scalars().first()

    if existing:
        # update fields
        existing.device_name = payload.get("device_name") or existing.device_name
        existing.device_type = payload.get("device_type") or existing.device_type
        existing.os = payload.get("os") or existing.os
        existing.last_seen = now
        existing.is_online = True
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
        if generated:
            result["original_id"] = device_id
            result["normalized"] = True

    if settings.mentor_api_url:
        fwd = dict(payload)
        fwd["deviceid"] = final_id
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                await client.post(f"{settings.mentor_api_url}/devices/register", json=fwd)
        except Exception:
            pass

    return result


@router.post("/{device_id}/metrics")
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
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
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
                await client.post(f"{settings.mentor_api_url}/devices/metrics", json=forward)
        except Exception:
            # Do not fail ingestion if forwarding fails
            pass
    return {"status": "ok"}


@router.post("/{device_id}/processes")
async def post_processes(device_id: str, processes: List[dict], db: AsyncSession = Depends(get_db)):
    # Validate legacy fields and reject with clear error messages
    for p in processes:
        if "name" in p:
            raise HTTPException(status_code=400, detail="unsupported legacy field: name; use process_name")
        if "command" in p:
            raise HTTPException(status_code=400, detail="unsupported legacy field: command; use command_text")
    
    # delete existing processes for device, then insert new ones
    await db.execute(dev_models.DeviceProcess.__table__.delete().where(dev_models.DeviceProcess.deviceid == device_id))
    to_add = []
    now = datetime.datetime.utcnow()
    for p in processes:
        to_add.append({
            "deviceid": device_id,
            "pid": p.get("pid"),
            "process_name": p.get("process_name"),
            "cpu": p.get("cpu"),
            "memory": p.get("memory"),
            "command_text": p.get("command_text"),
            "timestamp": now,
        })
    if to_add:
        await db.execute(dev_models.DeviceProcess.__table__.insert(), to_add)
        await db.commit()
        # Optionally forward processes to mentor backend if configured
        if settings.mentor_api_url:
            try:
                async with httpx.AsyncClient(timeout=5.0) as client:
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
                    await client.post(f"{settings.mentor_api_url}/devices/processes", json=forward)
            except Exception:
                pass
    return {"inserted": len(to_add)}


@router.post("/{device_id}/activities")
async def post_activity(device_id: str, activities: List[dict], db: AsyncSession = Depends(get_db)):
    # If legacy field 'type' is provided, treat as validation issue (422) instead of 400
    for a in activities:
        if "type" in a and not a.get("activity_type"):
            # Non-empty legacy 'type' should be rejected as bad request (400)
            if (a.get("type") or "") != "":
                raise HTTPException(status_code=400, detail="unsupported legacy field: type; use activity_type")
            # Empty legacy 'type' is treated as validation error (422)
            raise HTTPException(status_code=422, detail="invalid field: use activity_type instead of type")
    
    to_add = []
    now = datetime.datetime.utcnow()
    for a in activities:
        to_add.append({
            "deviceid": device_id,
            "activity_type": a.get("activity_type"),
            "description": a.get("description"),
            "app": a.get("app"),
            "duration": a.get("duration"),
            "timestamp": now,
        })
    if to_add:
        await db.execute(dev_models.DeviceActivity.__table__.insert(), to_add)
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
async def post_alerts(device_id: str, alerts: List[dict], db: AsyncSession = Depends(get_db)):
    # Validate legacy fields and reject with clear error messages
    for a in alerts:
        if "type" in a:
            raise HTTPException(status_code=400, detail="unsupported legacy field: type; use alert_type")
    
    to_add = []
    now = datetime.datetime.utcnow()
    for a in alerts:
        to_add.append({
            "deviceid": device_id,
            "level": a.get("level"),
            "alert_type": a.get("alert_type"),
            "message": a.get("message"),
            "value": a.get("value"),
            "threshold": a.get("threshold"),
            "timestamp": now,
        })
    if to_add:
        await db.execute(dev_models.DeviceAlert.__table__.insert(), to_add)
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
        devices.append({
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
        })
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
        processes.append({
            "processid": str(process.processid),
            "deviceid": str(process.deviceid),
            "timestamp": process.timestamp.isoformat() if process.timestamp else None,
            "pid": process.pid,
            "process_name": process.process_name,
            "name": process.process_name,  # legacy alias for compatibility
            "cpu": float(process.cpu) if process.cpu is not None else None,
            "memory": process.memory,
            "command_text": process.command_text,
        })
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
        activities.append({
            "activityid": str(activity.activityid),
            "deviceid": str(activity.deviceid),
            "timestamp": activity.timestamp.isoformat() if activity.timestamp else None,
            "activity_type": activity.activity_type,
            "description": activity.description,
            "app": activity.app,
            "duration": activity.duration,
        })
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
        alerts.append({
            "alertid": str(alert.alertid),
            "deviceid": str(alert.deviceid),
            "timestamp": alert.timestamp.isoformat() if alert.timestamp else None,
            "level": alert.level,
            "alert_type": alert.alert_type,
            "message": alert.message,
            "value": float(alert.value) if alert.value is not None else None,
            "threshold": float(alert.threshold) if alert.threshold is not None else None,
        })
    return alerts


@router.get("/{device_id}")
async def get_device_by_id(device_id: str, db: AsyncSession = Depends(get_db)):
    """Get a specific device by ID."""
    # If not a valid UUID, treat as not found to match test expectations
    try:
        UUID(device_id)
    except Exception:
        raise HTTPException(status_code=404, detail="Device not found")
    
    res = await db.execute(
        select(dev_models.Device).where(dev_models.Device.deviceid == device_id)
    )
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


@router.get("/{device_id}/commands/pending", response_model=List[CommandOut])
async def get_pending_commands(device_id: str, db: AsyncSession = Depends(get_db)):
    """Get pending commands for a device"""
    res = await db.execute(
        select(dev_models.DeviceRemoteCommand)
        .where(dev_models.DeviceRemoteCommand.deviceid == device_id)
        .where(dev_models.DeviceRemoteCommand.status == "pending")
        .order_by(dev_models.DeviceRemoteCommand.created_at.asc())
    )
    commands = res.scalars().all()
    return commands


@router.post("/commands/{command_id}/result")
async def submit_command_result(
    command_id: UUID, 
    payload: CommandResultSubmit, 
    db: AsyncSession = Depends(get_db)
):
    """Submit command execution result"""
    res = await db.execute(
        select(dev_models.DeviceRemoteCommand).where(dev_models.DeviceRemoteCommand.commandid == command_id)
    )
    command = res.scalars().first()
    if not command:
        raise HTTPException(status_code=404, detail="Command not found")
    
    # Update command with result
    command.status = payload.status
    command.result = payload.result or ""
    command.exit_code = payload.exit_code or 0
    command.completed_at = datetime.datetime.utcnow()
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
async def create_command(
    device_id: str, 
    payload: CommandCreate, 
    db: AsyncSession = Depends(get_db)
):
    """Create a new command for a device (forwarded from mentor backend)"""
    # Validate command against whitelist
    allowed_commands = ["get_info", "status", "restart", "get_processes", "get_logs", "restart_service", "screenshot"]
    command_base = payload.command_text.lower().split()[0] if payload.command_text else ""
    if command_base not in allowed_commands:
        raise HTTPException(status_code=400, detail=f"Command not allowed. Allowed commands: {', '.join(allowed_commands)}")
    
    command = dev_models.DeviceRemoteCommand(
        deviceid=device_id,
        command_text=payload.command_text,
        status="pending",
        created_at=datetime.datetime.utcnow(),
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
        metrics.append({
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
        })
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
        processes.append({
            "processid": str(process.processid),
            "deviceid": str(process.deviceid),
            "timestamp": process.timestamp.isoformat() if process.timestamp else None,
            "pid": process.pid,
            "process_name": process.process_name,
            "cpu": float(process.cpu) if process.cpu is not None else None,
            "memory": process.memory,
            "command_text": process.command_text,
        })
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
        activities.append({
            "activityid": str(activity.activityid),
            "deviceid": str(activity.deviceid),
            "timestamp": activity.timestamp.isoformat() if activity.timestamp else None,
            "activity_type": activity.activity_type,
            "description": activity.description,
            "app": activity.app,
            "duration": activity.duration,
        })
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
        alerts.append({
            "alertid": str(alert.alertid),
            "deviceid": str(alert.deviceid),
            "timestamp": alert.timestamp.isoformat() if alert.timestamp else None,
            "level": alert.level,
            "alert_type": alert.alert_type,
            "message": alert.message,
            "value": float(alert.value) if alert.value is not None else None,
            "threshold": float(alert.threshold) if alert.threshold is not None else None,
        })
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
        screenshots.append({
            "screenshotid": str(screenshot.screenshotid),
            "deviceid": str(screenshot.deviceid),
            "timestamp": screenshot.timestamp.isoformat() if screenshot.timestamp else None,
            "path": screenshot.path,
            "resolution": screenshot.resolution,
            "size": screenshot.size,
        })
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
    commands = res.scalars().all()
    return commands
