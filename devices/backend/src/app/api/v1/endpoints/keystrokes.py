from fastapi import APIRouter, Depends
from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select

from app.db.session import get_db
from app.models.keystrokes import Keystroke as KeystrokeModel

router = APIRouter()

class KeystrokeCreate(BaseModel):
    user_id: str
    key: str
    timestamp: Optional[datetime] = None

class KeystrokeResponse(BaseModel):
    id: str
    user_id: str
    key: str
    timestamp: Optional[datetime] = None

@router.post("/", status_code=201, response_model=KeystrokeResponse)
async def create_keystroke(keystroke: KeystrokeCreate, db: AsyncSession = Depends(get_db)):
    obj = KeystrokeModel(user_id=keystroke.user_id, key=keystroke.key)
    db.add(obj)
    await db.commit()
    await db.refresh(obj)
    return {"id": str(obj.id), "user_id": str(obj.user_id), "key": obj.key, "timestamp": keystroke.timestamp}

@router.get("/", response_model=List[KeystrokeResponse])
async def get_keystrokes(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(KeystrokeModel))
    items = result.scalars().all()
    return [
        {"id": str(x.id), "user_id": str(x.user_id), "key": x.key, "timestamp": None}
        for x in items
    ]
