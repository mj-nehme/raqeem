"""Devices Backend API - Main application entry point.

This module initializes the FastAPI application with:
- API routing and endpoints
- CORS configuration
- Request logging and tracing
- Health check endpoints
"""

import json
import logging
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
from app.db import session

# Configure structured logging
configure_logging()
logger = get_logger(__name__)


@asynccontextmanager
async def lifespan(_app: FastAPI):
    """Application lifespan manager for startup and shutdown events."""
    # Startup
    logger.info("Starting Raqeem Devices Backend API")
    logger.info("API documentation available at /docs")
    yield
    # Shutdown: Close database connections gracefully
    logger.info("Shutting down Raqeem Devices Backend API")
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
        {
            "name": "Health Check",
            "description": "Service health and readiness endpoints",
        },
    ],
)

# Setup CORS
setup_cors(app)

# Add request ID middleware for distributed tracing
app.add_middleware(RequestIDMiddleware)

# Include API routes
app.include_router(api_router, prefix="/api/v1")

# Include health check routes at root level
app.include_router(health.router, tags=["Health Check"])


# Keep backward compatibility with old health check endpoint
@app.get("/health", tags=["Health Check"], summary="Health check endpoint")
async def health_check_legacy():
    """
    Health check endpoint for monitoring and load balancer probes.
    
    **Deprecated**: Use `/health/live` for liveness checks or `/health/ready` for readiness checks instead.
    
    Returns service status and name for verification.
    """
    logger.warning("Legacy /health endpoint accessed - consider using /health/live or /health/ready")
    return {"status": "ok", "service": "devices-backend"}
