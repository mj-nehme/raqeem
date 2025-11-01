from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()

class AppActivityCreate(BaseModel):
    user_id: str
    app_name: str
    activity: str

@router.post("/", status_code=201)
async def create_app_activity(activity: AppActivityCreate):
    return {"message": "App activity saved", "data": activity}
