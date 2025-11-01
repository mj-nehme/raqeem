from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()

class LocationCreate(BaseModel):
    user_id: str
    latitude: float
    longitude: float

@router.post("/")
async def create_location(location: LocationCreate):
    return {"message": "Location added", "data": location}
