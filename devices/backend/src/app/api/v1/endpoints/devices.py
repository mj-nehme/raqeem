from fastapi import APIRouter, Depends, HTTPException
from typing import List
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from app.db.session import get_db
from app.models import devices as dev_models
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
        existing.current_user = payload.get("current_user") or existing.current_user
        db.add(existing)
        await db.commit()
        return {"device_id": device_id, "updated": True}
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
            current_user=payload.get("current_user"),
        )
        db.add(obj)
        await db.commit()
        return {"device_id": device_id, "created": True}


@router.post("/{device_id}/metrics")
async def post_metrics(device_id: str, payload: dict, db: AsyncSession = Depends(get_db)):
    obj = dev_models.DeviceMetrics(
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
    return {"status": "ok"}


@router.post("/{device_id}/processes")
async def post_processes(device_id: str, processes: List[dict], db: AsyncSession = Depends(get_db)):
    # delete existing processes for device, then insert new ones
    await db.execute(dev_models.Process.__table__.delete().where(dev_models.Process.device_id == device_id))
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
        await db.execute(dev_models.Process.__table__.insert(), to_add)
        await db.commit()
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
        await db.execute(dev_models.ActivityLog.__table__.insert(), to_add)
        await db.commit()
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
        await db.execute(dev_models.Alert.__table__.insert(), to_add)
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
            "current_user": device.current_user,
        })
    return devices
