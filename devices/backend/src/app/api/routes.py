from app.api.v1.endpoints import devices, screenshots
from fastapi import APIRouter

api_router = APIRouter()

api_router.include_router(screenshots.router, prefix="/screenshots", tags=["screenshots"])
api_router.include_router(devices.router, prefix="/devices", tags=["devices"])
