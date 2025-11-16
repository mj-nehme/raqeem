"""Devices Backend API - Main application entry point.

This module initializes the FastAPI application with:
- API routing and endpoints
- CORS configuration
- Request logging and tracing
- Health check endpoints
"""

import json
import logging
import signal
import sys
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
from app.core.middleware import CorrelationIDMiddleware, RequestIDMiddleware
from app.db.session import engine

# Configure structured logging
configure_logging()
logger = get_logger(__name__)


@asynccontextmanager
async def lifespan(_app: FastAPI):
    """Manage application lifecycle with graceful startup and shutdown."""
    # Startup
    logger.info("Starting up Devices Backend...")

    # Setup signal handlers for graceful shutdown
    def handle_shutdown(signum, frame):
        logger.info(f"Received signal {signum}, initiating graceful shutdown...")
        sys.exit(0)

    signal.signal(signal.SIGTERM, handle_shutdown)
    signal.signal(signal.SIGINT, handle_shutdown)

    yield

    # Shutdown
    logger.info("Shutting down Devices Backend...")
    await engine.dispose()
    logger.info("Database connections closed")


app = FastAPI(
    title="Raqeem Devices Backend API",
    description="""High-throughput telemetry ingestion API for IoT device monitoring.

## Overview
The Devices Backend is the primary data ingestion point for IoT device telemetry.
It handles high-volume data streams and forwards critical alerts to the Mentor Backend.

## Key Features
- **Device Registration**: Register and manage device information
- **Metrics Ingestion**: High-throughput telemetry data collection
- **Activity Logging**: Track user activities and system events
- **Alert Processing**: Process and forward critical alerts
- **Process Monitoring**: Track running processes on devices
- **Screenshot Storage**: Upload and manage device screenshots (MinIO/S3)
- **Remote Commands**: Execute commands on devices remotely

## Data Flow
Devices POST data directly to this backend, which:
1. Stores metrics in PostgreSQL
2. Uploads screenshots to MinIO (S3-compatible)
3. Forwards alerts to Mentor Backend (optional)

## Authentication
Currently, the API does not require authentication.
Authentication and authorization will be added in future releases.""",
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
)

# Add correlation ID and request ID middleware
app.add_middleware(CorrelationIDMiddleware)
app.add_middleware(RequestIDMiddleware)

# Setup CORS
setup_cors(app)

app.include_router(api_router, prefix="/api/v1")

# Include health check router
app.include_router(health.router)

# Keep legacy health endpoint for backwards compatibility
@app.get("/health")
async def legacy_health_check():
    return {"status": "ok", "service": "devices-backend"}
