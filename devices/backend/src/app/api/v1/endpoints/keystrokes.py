from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()

class KeystrokeCreate(BaseModel):
    user_id: str
    key: str

@router.post("/", status_code=201)
async def create_keystroke(keystroke: KeystrokeCreate):
    return {"message": "Keystroke logged", "data": keystroke}
