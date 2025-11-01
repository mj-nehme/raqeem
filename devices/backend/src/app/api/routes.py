from fastapi import APIRouter
from app.api.v1.endpoints import (
    users,
    locations,
    screenshots,
    keystrokes,
    app_activity
    , devices
)

api_router = APIRouter()

api_router.include_router(users.router, prefix="/users", tags=["users"])
api_router.include_router(locations.router, prefix="/locations", tags=["locations"])
api_router.include_router(screenshots.router, prefix="/screenshots", tags=["screenshots"])
api_router.include_router(keystrokes.router, prefix="/keystrokes", tags=["keystrokes"])
api_router.include_router(app_activity.router, prefix="/app-activity", tags=["app_activity"])
api_router.include_router(devices.router, prefix="/devices", tags=["devices"])
