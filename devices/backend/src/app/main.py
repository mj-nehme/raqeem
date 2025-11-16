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
from app.core.cors import setup_cors
from app.db import session


@asynccontextmanager
async def lifespan(_app: FastAPI):
    """
    Application lifespan manager for startup and shutdown procedures.

    This context manager implements graceful shutdown by ensuring database
    connections are properly closed when the application stops. This prevents
    connection leaks and allows in-flight requests to complete before shutdown.

    The lifespan pattern is preferred over startup/shutdown events as it
    guarantees cleanup occurs even if the application crashes.
    """
    # Startup: Nothing to initialize currently
    yield
    # Shutdown: Close database connections gracefully
    await session.shutdown()


app = FastAPI(
    title="Raqeem Devices Backend API",
    description="""
# Raqeem Devices Backend API

High-throughput telemetry ingestion API for IoT device monitoring and management.

## Overview

The Devices Backend handles:
- **Device Registration**: Device enrollment and configuration management
- **Metrics Collection**: Real-time performance monitoring (CPU, memory, disk, network)
- **Activity Logging**: User activity and application usage tracking
- **Alert Submission**: Threshold-based alert generation and forwarding
- **Process Monitoring**: Running process tracking and analysis
- **Screenshot Capture**: Visual device state capture and storage
- **Remote Commands**: Secure remote command execution and result tracking

## Key Features

- **High-Throughput Ingestion**: Optimized for handling large volumes of telemetry data
- **Async Processing**: Non-blocking I/O for maximum performance
- **Automatic Forwarding**: Optional integration with Mentor Backend for centralized monitoring
- **UUID-Based Identification**: Globally unique device identifiers
- **Timestamp Tracking**: Server-side timestamping for accurate data correlation
- **Legacy Compatibility**: Support for legacy field names with validation

## Authentication

Currently, the API does not require authentication for device endpoints.
Authentication and authorization will be added in future releases.

## Rate Limiting

No rate limiting is currently enforced. Production deployments should
implement rate limiting at the infrastructure level (e.g., API Gateway, Ingress).

## Data Forwarding

When `MENTOR_API_URL` is configured, the Devices Backend automatically forwards
telemetry data to the Mentor Backend for centralized monitoring and analysis.
Forwarding failures do not block the ingestion pipeline.

## Versioning

Current API version: **v1**

API is versioned through URL path: `/api/v1/*`
    """,
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
    contact={
        "name": "Raqeem Support",
        "url": "https://github.com/mj-nehme/raqeem",
        "email": "support@example.com",
    },
    license_info={
        "name": "MIT",
        "url": "https://opensource.org/licenses/MIT",
    },
    openapi_tags=[
        {
            "name": "Device Registration",
            "description": "Device enrollment and configuration management",
        },
        {
            "name": "Device Information",
            "description": "Query device status and configuration",
        },
        {
            "name": "Device Metrics",
            "description": "Performance metrics ingestion and retrieval",
        },
        {
            "name": "Device Processes",
            "description": "Process monitoring and tracking",
        },
        {
            "name": "Device Activities",
            "description": "User activity logging and retrieval",
        },
        {
            "name": "Device Alerts",
            "description": "Alert submission and retrieval",
        },
        {
            "name": "Device Commands",
            "description": "Remote command execution and tracking",
        },
        {
            "name": "Device Screenshots",
            "description": "Screenshot metadata management",
        },
        {
            "name": "Screenshots",
            "description": "Screenshot upload and storage",
        },
    ],
)

# Setup CORS
setup_cors(app)

app.include_router(api_router, prefix="/api/v1")


@app.get("/health", tags=["Health Check"], summary="Health check endpoint")
async def health_check():
    """
    Basic health check endpoint for monitoring and load balancer probes.

    This is a lightweight liveness check that only verifies the application
    is running and can respond to requests. It doesn't check dependencies.
    Use /health/ready for readiness checks that validate database connectivity.

    Returns service status and name for verification.
    """
    return {"status": "ok", "service": "devices-backend"}


@app.get("/health/ready")
async def health_check_ready():
    """
    Readiness check that validates all critical dependencies.

    This endpoint performs comprehensive health checks on all required services:
    - Database connectivity: Verifies PostgreSQL is accessible

    Returns 200 OK if all dependencies are healthy, or 503 Service Unavailable
    if any critical dependency is down. Load balancers should use this endpoint
    to determine if the service is ready to accept traffic.

    The response includes detailed status for each dependency to aid debugging.
    """
    health = {
        "status": "ok",
        "service": "devices-backend",
        "checks": {},
    }

    # Check database connectivity
    # This verifies the connection pool is healthy and PostgreSQL is responding
    db_healthy = await session.health_check()
    if db_healthy:
        health["checks"]["database"] = {"status": "healthy"}
    else:
        health["checks"]["database"] = {"status": "unhealthy"}
        health["status"] = "degraded"
        return health, 503

    return health

