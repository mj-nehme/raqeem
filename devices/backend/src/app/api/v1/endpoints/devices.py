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

router = APIRouter()


@router.post("/register", status_code=200)
async def register_device(payload: dict, db: AsyncSession = Depends(get_db)):
    # payload expected to contain id and optional fields
    device_id = payload.get("id") or payload.get("device_id")
    if not device_id:
        raise HTTPException(status_code=400, detail="missing device id")

    # Upsert device row (simple read-then-update or insert)
    now = datetime.datetime.utcnow()
    # try to find existing device
    res = await db.execute(select(dev_models.Device).where(dev_models.Device.id == device_id))
    existing = res.scalars().first()
    if existing:
        # update fields
        existing.name = payload.get("name") or existing.name
        existing.type = payload.get("type") or existing.type
        existing.os = payload.get("os") or existing.os
        existing.last_seen = now
        existing.is_online = True
        existing.location = payload.get("location") or existing.location
        existing.ip_address = payload.get("ip_address") or existing.ip_address
        existing.mac_address = payload.get("mac_address") or existing.mac_address
        existing.current_user_text = payload.get("current_user") or existing.current_user_text
        db.add(existing)
        await db.commit()
        result = {"device_id": device_id, "updated": True}
    else:
        obj = dev_models.Device(
            id=device_id,
            name=payload.get("name"),
            type=payload.get("type"),
            os=payload.get("os"),
            last_seen=now,
            is_online=True,
            location=payload.get("location"),
            ip_address=payload.get("ip_address"),
            mac_address=payload.get("mac_address"),
            current_user_text=payload.get("current_user"),
        )
        db.add(obj)
        await db.commit()
        result = {"device_id": device_id, "created": True}
    
    # Forward device registration to mentor backend if configured
    # Note: Input is already validated by this endpoint, and mentor backend
    # will perform its own validation via BindJSON
    if settings.mentor_api_url:
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                await client.post(f"{settings.mentor_api_url}/devices/register", json=payload)
        except Exception:
            # Do not fail registration if forwarding fails
            pass
    
    return result


@router.post("/{device_id}/metrics")
async def post_metrics(device_id: str, payload: dict, db: AsyncSession = Depends(get_db)):
    obj = dev_models.DeviceMetric(
        device_id=device_id,
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
                    "device_id": device_id,
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
    # delete existing processes for device, then insert new ones
    await db.execute(dev_models.DeviceProcess.__table__.delete().where(dev_models.DeviceProcess.device_id == device_id))
    to_add = []
    now = datetime.datetime.utcnow()
    for p in processes:
        to_add.append({
            "device_id": device_id,
            "pid": p.get("pid"),
            "name": p.get("name"),
            "cpu": p.get("cpu"),
            "memory": p.get("memory"),
            "command": p.get("command"),
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
                            "device_id": device_id,
                            "pid": p.get("pid"),
                            "name": p.get("name"),
                            "cpu": p.get("cpu"),
                            "memory": p.get("memory"),
                            "command": p.get("command"),
                        }
                        for p in processes
                    ]
                    await client.post(f"{settings.mentor_api_url}/devices/processes", json=forward)
            except Exception:
                pass
    return {"inserted": len(to_add)}


@router.post("/{device_id}/activities")
async def post_activity(device_id: str, activities: List[dict], db: AsyncSession = Depends(get_db)):
    to_add = []
    now = datetime.datetime.utcnow()
    for a in activities:
        to_add.append({
            "device_id": device_id,
            "type": a.get("type"),
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
                            "device_id": device_id,
                            "type": a.get("type"),
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
    to_add = []
    now = datetime.datetime.utcnow()
    for a in alerts:
        to_add.append({
            "device_id": device_id,
            "level": a.get("level"),
            "type": a.get("type"),
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
                            "device_id": device_id,
                            "level": a.get("level"),
                            "type": a.get("type"),
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
            "id": device.id,
            "name": device.name,
            "type": device.type,
            "os": device.os,
            "last_seen": device.last_seen.isoformat() if device.last_seen else None,
            "is_online": device.is_online,
            "location": device.location,
            "ip_address": device.ip_address,
            "mac_address": device.mac_address,
            "current_user": device.current_user_text,
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
            "id": str(process.id),
            "device_id": process.device_id,
            "timestamp": process.timestamp.isoformat() if process.timestamp else None,
            "pid": process.pid,
            "name": process.name,
            "cpu": float(process.cpu) if process.cpu is not None else None,
            "memory": process.memory,
            "command": process.command,
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
            "id": str(activity.id),
            "device_id": activity.device_id,
            "timestamp": activity.timestamp.isoformat() if activity.timestamp else None,
            "type": activity.type,
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
            "id": str(alert.id),
            "device_id": alert.device_id,
            "timestamp": alert.timestamp.isoformat() if alert.timestamp else None,
            "level": alert.level,
            "type": alert.type,
            "message": alert.message,
            "value": float(alert.value) if alert.value is not None else None,
            "threshold": float(alert.threshold) if alert.threshold is not None else None,
        })
    return alerts


@router.get("/{device_id}/commands/pending", response_model=List[CommandOut])
async def get_pending_commands(device_id: str, db: AsyncSession = Depends(get_db)):
    """Get pending commands for a device"""
    res = await db.execute(
        select(dev_models.DeviceRemoteCommand)
        .where(dev_models.DeviceRemoteCommand.device_id == device_id)
        .where(dev_models.DeviceRemoteCommand.status == "pending")
        .order_by(dev_models.DeviceRemoteCommand.created_at.asc())
    )
    commands = res.scalars().all()
    return commands


@router.post("/commands/{command_id}/result")
async def submit_command_result(
    command_id: int, 
    payload: CommandResultSubmit, 
    db: AsyncSession = Depends(get_db)
):
    """Submit command execution result"""
    res = await db.execute(
        select(dev_models.DeviceRemoteCommand).where(dev_models.DeviceRemoteCommand.id == command_id)
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
                    "id": command.id,
                    "status": command.status,
                    "result": command.result,
                    "exit_code": command.exit_code,
                }
                await client.post(f"{settings.mentor_api_url}/commands/status", json=forward_payload)
        except Exception:
            # Don't fail if forwarding fails
            pass
    
    return {"status": "ok", "command_id": command_id}


@router.post("/{device_id}/commands", response_model=CommandOut)
async def create_command(
    device_id: str, 
    payload: CommandCreate, 
    db: AsyncSession = Depends(get_db)
):
    """Create a new command for a device (forwarded from mentor backend)"""
    # Validate command against whitelist
    allowed_commands = ["get_info", "status", "restart", "get_processes", "get_logs", "restart_service", "screenshot"]
    command_base = payload.command.lower().split()[0] if payload.command else ""
    if command_base not in allowed_commands:
        raise HTTPException(status_code=400, detail=f"Command not allowed. Allowed commands: {', '.join(allowed_commands)}")
    
    command = dev_models.DeviceRemoteCommand(
        device_id=device_id,
        command=payload.command,
        status="pending",
        created_at=datetime.datetime.utcnow(),
    )
    db.add(command)
    await db.commit()
    await db.refresh(command)
    
    return command


@router.get("/processes")
async def get_all_processes(db: AsyncSession = Depends(get_db)):
    """Get all processes from all devices"""
    res = await db.execute(select(dev_models.DeviceProcess).order_by(dev_models.DeviceProcess.timestamp.desc()).limit(100))
    processes = res.scalars().all()
    return [
        {
            "id": str(process.id),
            "device_id": process.device_id,
            "pid": process.pid,
            "name": process.name,
            "cpu": process.cpu,
            "memory": process.memory,
            "command": process.command,
            "timestamp": process.timestamp.isoformat() if process.timestamp else None,
        }
        for process in processes
    ]


@router.get("/activities")
async def get_all_activities(db: AsyncSession = Depends(get_db)):
    """Get all activities from all devices"""
    res = await db.execute(select(dev_models.DeviceActivity).order_by(dev_models.DeviceActivity.timestamp.desc()).limit(100))
    activities = res.scalars().all()
    return [
        {
            "id": str(activity.id),
            "device_id": activity.device_id,
            "type": activity.type,
            "description": activity.description,
            "app": activity.app,
            "duration": activity.duration,
            "timestamp": activity.timestamp.isoformat() if activity.timestamp else None,
        }
        for activity in activities
    ]


@router.get("/alerts")
async def get_all_alerts(db: AsyncSession = Depends(get_db)):
    """Get all alerts from all devices"""
    res = await db.execute(select(dev_models.DeviceAlert).order_by(dev_models.DeviceAlert.timestamp.desc()).limit(100))
    alerts = res.scalars().all()
    return [
        {
            "id": str(alert.id),
            "device_id": alert.device_id,
            "level": alert.level,
            "type": alert.type,
            "message": alert.message,
            "value": alert.value,
            "threshold": alert.threshold,
            "timestamp": alert.timestamp.isoformat() if alert.timestamp else None,
        }
        for alert in alerts
    ]
