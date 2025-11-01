
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

router = APIRouter()

class UserCreate(BaseModel):
    device_id: str
    name: str | None = None

@router.post("/", status_code=201)
async def create_user(user: UserCreate):
    # You'll replace this with DB logic later
    return {"message": "User created", "data": user}
