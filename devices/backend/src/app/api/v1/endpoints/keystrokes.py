from fastapi import APIRouter
from pydantic import BaseModel
from typing import List
from datetime import datetime

router = APIRouter()

class KeystrokeCreate(BaseModel):
    user_id: str
    key: str
    timestamp: datetime | None = None

class KeystrokeResponse(BaseModel):
    id: int
    user_id: str
    key: str
    timestamp: datetime | None = None

# Mock data store
keystrokes_db: List[dict] = []
keystroke_id_counter = 1

@router.post("/", status_code=201, response_model=KeystrokeResponse)
async def create_keystroke(keystroke: KeystrokeCreate):
    global keystroke_id_counter
    keystroke_data = {
        "id": keystroke_id_counter,
        "user_id": keystroke.user_id,
        "key": keystroke.key,
        "timestamp": keystroke.timestamp
    }
    keystrokes_db.append(keystroke_data)
    keystroke_id_counter += 1
    return keystroke_data

@router.get("/", response_model=List[KeystrokeResponse])
async def get_keystrokes():
    return keystrokes_db
