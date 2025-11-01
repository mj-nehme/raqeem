from fastapi import APIRouter
from pydantic import BaseModel
from typing import List
from datetime import datetime

router = APIRouter()

class LocationCreate(BaseModel):
    user_id: str
    latitude: float
    longitude: float
    timestamp: datetime | None = None

class LocationResponse(BaseModel):
    id: int
    user_id: str
    latitude: float
    longitude: float
    timestamp: datetime | None = None

# Mock data store
locations_db: List[dict] = []
location_id_counter = 1

@router.post("/", status_code=201, response_model=LocationResponse)
async def create_location(location: LocationCreate):
    global location_id_counter
    location_data = {
        "id": location_id_counter,
        "user_id": location.user_id,
        "latitude": location.latitude,
        "longitude": location.longitude,
        "timestamp": location.timestamp
    }
    locations_db.append(location_data)
    location_id_counter += 1
    return location_data

@router.get("/", response_model=List[LocationResponse])
async def get_locations():
    return locations_db
