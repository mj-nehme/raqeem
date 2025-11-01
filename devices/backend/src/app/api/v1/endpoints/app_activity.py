from fastapi import APIRouter
from pydantic import BaseModel
from typing import List
from datetime import datetime

router = APIRouter()

class AppActivityCreate(BaseModel):
    user_id: str
    activity: str
    timestamp: datetime | None = None

class AppActivityResponse(BaseModel):
    id: int
    user_id: str
    activity: str
    timestamp: datetime | None = None

# Mock data store
activities_db: List[dict] = []
activity_id_counter = 1

@router.post("/", status_code=201, response_model=AppActivityResponse)
async def create_app_activity(activity: AppActivityCreate):
    global activity_id_counter
    activity_data = {
        "id": activity_id_counter,
        "user_id": activity.user_id,
        "activity": activity.activity,
        "timestamp": activity.timestamp
    }
    activities_db.append(activity_data)
    activity_id_counter += 1
    return activity_data

@router.get("/", response_model=List[AppActivityResponse])
async def get_app_activities():
    return activities_db
