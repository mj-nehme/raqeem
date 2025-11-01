
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import List

router = APIRouter()

class UserCreate(BaseModel):
    device_id: str
    name: str | None = None

class UserResponse(BaseModel):
    id: int
    device_id: str
    name: str | None = None

# Mock data store (replace with DB later)
users_db: List[dict] = []
user_id_counter = 1

@router.post("/", status_code=201, response_model=UserResponse)
async def create_user(user: UserCreate):
    global user_id_counter
    user_data = {
        "id": user_id_counter,
        "device_id": user.device_id,
        "name": user.name
    }
    users_db.append(user_data)
    user_id_counter += 1
    return user_data

@router.get("/", response_model=List[UserResponse])
async def get_users():
    return users_db
