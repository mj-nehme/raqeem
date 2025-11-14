from fastapi import APIRouter
from app.api.v1.endpoints import (
    screenshots,
    devices
)

api_router = APIRouter()

api_router.include_router(screenshots.router, prefix="/screenshots", tags=["screenshots"])
api_router.include_router(devices.router, prefix="/devices", tags=["devices"])
