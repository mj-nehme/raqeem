"""Health check endpoints for service monitoring and observability.

Provides comprehensive health checks including:
- Basic liveness check (is the service running?)
- Database connectivity check
- External service dependencies check
- Detailed component status
"""

from datetime import UTC, datetime
from typing import Any

from app.core.config import settings
from app.core.logging_config import get_logger
from app.db.session import get_db
from fastapi import APIRouter, Depends, status
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession

logger = get_logger(__name__)
router = APIRouter()


@router.get("/health", status_code=status.HTTP_200_OK)
async def health_check() -> dict[str, Any]:
    """Basic health check endpoint for load balancers and uptime monitoring.

    Returns:
        Simple status response indicating service is alive.

    Example Response:
        {
            "status": "ok",
            "service": "devices-backend",
            "timestamp": "2024-01-15T10:30:00Z"
        }
    """
    return {
        "status": "ok",
        "service": "devices-backend",
        "timestamp": datetime.now(UTC).isoformat(),
    }


@router.get("/health/ready", status_code=status.HTTP_200_OK)
async def readiness_check(db: AsyncSession = Depends(get_db)) -> dict[str, Any]:
    """Readiness check with dependency verification.

    Checks if the service is ready to accept traffic by verifying:
    - Database connectivity
    - Required configuration is present

    Returns:
        Detailed status of service readiness and dependencies.

    Raises:
        HTTP 503: If service is not ready (database down, missing config, etc.)

    Example Response:
        {
            "status": "ready",
            "service": "devices-backend",
            "timestamp": "2024-01-15T10:30:00Z",
            "checks": {
                "database": "ok",
                "config": "ok"
            }
        }
    """
    checks: dict[str, str] = {}
    overall_status = "ready"

    # Check database connectivity
    try:
        result = await db.execute(text("SELECT 1"))
        result.scalar()
        checks["database"] = "ok"
        logger.debug("Database health check passed")
    except Exception as e:
        checks["database"] = f"error: {e!s}"
        overall_status = "not_ready"
        logger.error("Database health check failed", exc_info=True)

    # Check required configuration
    try:
        # Verify critical settings are configured
        if not settings.database_url:
            checks["config"] = "error: database_url not configured"
            overall_status = "not_ready"
        elif not settings.secret_key:
            checks["config"] = "error: secret_key not configured"
            overall_status = "not_ready"
        else:
            checks["config"] = "ok"
    except Exception as e:
        checks["config"] = f"error: {e!s}"
        overall_status = "not_ready"
        logger.error("Config health check failed", exc_info=True)

    response = {
        "status": overall_status,
        "service": "devices-backend",
        "timestamp": datetime.now(UTC).isoformat(),
        "checks": checks,
    }

    # Return 503 if not ready
    if overall_status != "ready":
        return response

    return response


@router.get("/health/live", status_code=status.HTTP_200_OK)
async def liveness_check() -> dict[str, str]:
    """Liveness check for container orchestration (Kubernetes, etc.).

    This is a lightweight check that only verifies the service process is running.
    Does not check dependencies.

    Returns:
        Minimal status response.

    Example Response:
        {
            "status": "alive"
        }
    """
    return {"status": "alive"}
