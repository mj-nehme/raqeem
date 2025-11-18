import logging
import os
import tempfile
import uuid

from app.core.config import settings
from app.db.session import get_db
from app.models import devices as dev_models
from app.schemas.devices import ErrorResponse
from app.services.minio_service import MinioService
from app.util import post_with_retry
from fastapi import APIRouter, Depends, File, Form, HTTPException, UploadFile
from sqlalchemy.ext.asyncio import AsyncSession

router = APIRouter()
logger = logging.getLogger(__name__)

# Default screenshot resolution when not provided
DEFAULT_SCREENSHOT_RESOLUTION = "800x600"


@router.post(
    "/",
    status_code=201,
    responses={
        201: {
            "description": "Screenshot uploaded successfully",
            "content": {
                "application/json": {
                    "example": {
                        "id": "123e4567-e89b-12d3-a456-426614174000",
                        "image_url": "123e4567-e89b-12d3-a456-426614174000.png",
                        "status": "success",
                    }
                }
            },
        },
        422: {
            "description": "Validation error - missing required fields",
            "model": ErrorResponse,
        },
        500: {
            "description": "Internal server error during upload",
            "model": ErrorResponse,
        },
    },
    summary="Upload a device screenshot",
    description="""
    Upload a screenshot image file for a device.

    This endpoint:
    - Accepts PNG or JPG image files
    - Generates a unique identifier for each screenshot
    - Uploads the image to MinIO object storage
    - Stores screenshot metadata in the database
    - Forwards metadata to mentor backend if configured

    **Form Fields:**
    - `device_id` or `deviceid`: Device identifier (required)
    - `file`: Image file to upload (required)

    **Supported formats:** PNG, JPG/JPEG

    **Returns:** Screenshot identifier and image URL/path
    """,
    tags=["Screenshots"],
)
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

    temp_file_path = None
    try:
        # Generate unique filename
        file_id = str(uuid.uuid4())
        filename = f"{file_id}.png"

        # Read file to get size
        content = await file.read()
        file_size = len(content)

        # Save to temporary file for MinIO upload
        with tempfile.NamedTemporaryFile(mode='wb', delete=False, suffix='.png') as temp_file:
            temp_file.write(content)
            temp_file_path = temp_file.name

        # Upload to MinIO
        try:
            minio_service = MinioService()
            minio_service.upload_file(temp_file_path, filename)
            logger.info(f"Screenshot uploaded to MinIO: {filename}")
        except Exception as minio_error:
            logger.error(f"MinIO upload failed: {minio_error}")
            raise HTTPException(status_code=500, detail=f"MinIO upload failed: {minio_error!s}") from minio_error

        # Store in device_screenshots table
        device_screenshot = dev_models.DeviceScreenshot(
            deviceid=device_identifier, path=filename, resolution=DEFAULT_SCREENSHOT_RESOLUTION, size=file_size
        )
        db.add(device_screenshot)
        await db.commit()
        await db.refresh(device_screenshot)

        # Forward to mentor backend if configured
        if settings.mentor_api_url:
            payload = {
                "deviceid": device_identifier,
                "path": filename,
                "resolution": DEFAULT_SCREENSHOT_RESOLUTION,
                "size": file_size,
            }
            # Forward screenshot metadata to mentor backend with retry
            await post_with_retry(
                f"{settings.mentor_api_url}/devices/screenshots",
                json=payload,
                max_retries=2,
            )

        return {"id": str(device_screenshot.screenshotid), "image_url": device_screenshot.path, "status": "success"}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Screenshot upload failed: {e!s}") from e
    finally:
        # Clean up temporary file
        if temp_file_path and os.path.exists(temp_file_path):
            try:
                os.unlink(temp_file_path)
            except Exception as cleanup_error:
                logger.warning(f"Failed to delete temporary file {temp_file_path}: {cleanup_error}")
