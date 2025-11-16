"""Devices Backend API - Main application entry point.

This module initializes the FastAPI application with:
- API routing and endpoints
- CORS configuration
- Request logging and tracing
- Health check endpoints
"""

import json
from contextlib import asynccontextmanager

try:
    import httpx

    if not hasattr(httpx.Request, "json"):

        def _request_json(self):
            try:
                content = self.content
                if content is None:
                    return {}
                if isinstance(content, (bytes, bytearray)):
                    return json.loads(content.decode() or "{}")
                # Fallback for str content
                return json.loads(str(content) or "{}")
            except Exception:
                return {}

        httpx.Request.json = _request_json  # type: ignore[attr-defined]
except Exception:
    # If httpx is not available or patching fails, continue without raising
    pass
from fastapi import FastAPI

from app.api.routes import api_router
from app.api.v1.endpoints import health
from app.core.cors import setup_cors
from app.core.logging_config import configure_logging, get_logger
from app.core.middleware import RequestIDMiddleware

# Configure structured logging
configure_logging()
logger = get_logger(__name__)


@asynccontextmanager
async def lifespan(_app: FastAPI):
    """Application lifespan manager for startup and shutdown events."""
    # Startup
    logger.info("Starting Devices Backend API")
    yield
    # Shutdown
    logger.info("Shutting down Devices Backend API")


app = FastAPI(
    title="Raqeem Devices Backend API",
    description="High-throughput telemetry ingestion API for IoT device monitoring. Handles device registration, metrics collection, activity logging, and alert submission.",
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
)

# Setup CORS
setup_cors(app)

# Add request ID middleware for distributed tracing
app.add_middleware(RequestIDMiddleware)

# Include API routes
app.include_router(api_router, prefix="/api/v1")

# Include health check routes at root level
app.include_router(health.router, tags=["health"])


# Keep backward compatibility with old health check endpoint
@app.get("/health")
async def health_check_legacy():
    """Legacy health check endpoint for backward compatibility.

    Deprecated: Use /health/live or /health/ready instead.
    """
    logger.warning("Legacy /health endpoint accessed - consider using /health/live or /health/ready")
    return {"status": "ok", "service": "devices-backend"}
