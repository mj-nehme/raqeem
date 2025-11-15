import logging
import uuid

import httpx
from app.core.config import settings
from app.db.session import get_db
from app.models import devices as dev_models
from fastapi import APIRouter, Depends, File, Form, HTTPException, UploadFile
from sqlalchemy.ext.asyncio import AsyncSession

router = APIRouter()
logger = logging.getLogger(__name__)

# Default screenshot resolution when not provided
DEFAULT_SCREENSHOT_RESOLUTION = "800x600"


@router.post("/", status_code=201)
async def create_screenshot(
    device_id: str | None = Form(None),
    deviceid: str | None = Form(None),
    file: UploadFile = File(...),
    db: AsyncSession = Depends(get_db),
):
    # Support both 'device_id' and 'deviceid' form field names
    device_identifier = device_id or deviceid
    if not device_identifier:
        raise HTTPException(status_code=422, detail="device_id is required")

    try:
        # Generate unique filename
        file_id = str(uuid.uuid4())
        filename = f"{file_id}.png"

        # Read file to get size
        content = await file.read()
        file_size = len(content)

        # Store in device_screenshots table
        device_screenshot = dev_models.DeviceScreenshot(
            deviceid=device_identifier, path=filename, resolution=DEFAULT_SCREENSHOT_RESOLUTION, size=file_size
        )
        db.add(device_screenshot)
        await db.commit()
        await db.refresh(device_screenshot)

        # Forward to mentor backend if configured
        if settings.mentor_api_url:
            try:
                async with httpx.AsyncClient(timeout=5.0) as client:
                    payload = {
                        "deviceid": device_identifier,
                        "path": filename,
                        "resolution": DEFAULT_SCREENSHOT_RESOLUTION,
                        "size": file_size,
                    }
                    # Forward screenshot metadata to mentor backend
                    await client.post(f"{settings.mentor_api_url}/devices/screenshots", json=payload)
            except (httpx.RequestError, httpx.TimeoutException) as e:
                # Log forwarding errors but don't fail the screenshot upload
                logger.warning("Failed to forward screenshot to mentor backend: %s", e)
            except Exception:
                # Catch any other unexpected errors
                logger.exception("Unexpected error forwarding screenshot")
                # Don't fail if forwarding fails

        return {"id": str(device_screenshot.screenshotid), "image_url": device_screenshot.path, "status": "success"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Screenshot upload failed: {e!s}") from e
