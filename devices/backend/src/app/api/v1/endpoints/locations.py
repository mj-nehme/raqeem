from fastapi import APIRouter, Depends
from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select

from app.db.session import get_db
from app.models.locations import Location as LocationModel

router = APIRouter()

class LocationCreate(BaseModel):
    user_id: str
    latitude: float
    longitude: float
    timestamp: Optional[datetime] = None

class LocationResponse(BaseModel):
    id: str
    user_id: str
    latitude: float
    longitude: float
    timestamp: Optional[datetime] = None

@router.post("/", status_code=201, response_model=LocationResponse)
async def create_location(location: LocationCreate, db: AsyncSession = Depends(get_db)):
    obj = LocationModel(user_id=location.user_id, latitude=location.latitude, longitude=location.longitude)
    db.add(obj)
    await db.commit()
    await db.refresh(obj)
    return {
        "id": str(obj.id),
        "user_id": str(obj.user_id),
        "latitude": obj.latitude,
        "longitude": obj.longitude,
        "timestamp": location.timestamp,
    }

@router.get("/", response_model=List[LocationResponse])
async def get_locations(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(LocationModel))
    items = result.scalars().all()
    return [
        {
            "id": str(x.id),
            "user_id": str(x.user_id),
            "latitude": x.latitude,
            "longitude": x.longitude,
            "timestamp": None,
        }
        for x in items
    ]
