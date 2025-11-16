"""Health check endpoint with dependency validation."""

import datetime
import logging

from app.db.session import engine
from app.services.minio_service import check_minio_health
from fastapi import APIRouter, status
from sqlalchemy import text

logger = logging.getLogger(__name__)

router = APIRouter()


@router.get("/health", status_code=status.HTTP_200_OK)
async def health_check():
    """
    Comprehensive health check endpoint.

    Checks:
    - Service availability
    - Database connection
    - MinIO/S3 connection

    Returns:
        Dict with health status and details
    """
    health_status = {
        "status": "healthy",
        "service": "devices-backend",
        "timestamp": datetime.datetime.now(datetime.UTC).isoformat(),
        "details": {},
    }

    all_healthy = True

    # Check database health
    db_health = await check_database_health()
    health_status["details"]["database"] = db_health
    if db_health["status"] != "healthy":
        all_healthy = False

    # Check MinIO health
    minio_health = await check_minio_health()
    health_status["details"]["minio"] = minio_health
    if minio_health["status"] != "healthy":
        all_healthy = False

    if not all_healthy:
        health_status["status"] = "unhealthy"
        return health_status, status.HTTP_503_SERVICE_UNAVAILABLE

    return health_status


async def check_database_health() -> dict:
    """Check database connection health."""
    try:
        async with engine.connect() as conn:
            await conn.execute(text("SELECT 1"))
            return {
                "status": "healthy",
                "pool_size": engine.pool.size(),
                "checked_in": engine.pool.checkedin(),
            }
    except Exception as e:
        logger.error(f"Database health check failed: {e}")
        return {"status": "unhealthy", "error": str(e)}
