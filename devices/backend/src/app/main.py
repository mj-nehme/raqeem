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


@asynccontextmanager
async def lifespan(_app: FastAPI):
    # Startup
    yield
    # Shutdown: nothing to clean up currently


app = FastAPI(
    title="Raqeem Devices Backend API",
    description="""
# High-Throughput Telemetry Ingestion API

The Raqeem Devices Backend handles device registration, metrics collection, activity logging, 
and alert submission for IoT device monitoring.

## Key Features

* **Device Registration**: Register and update device information
* **Metrics Collection**: Collect CPU, memory, disk, and network metrics
* **Activity Logging**: Track user activities and application usage
* **Alert Management**: Submit and query device alerts
* **Screenshot Management**: Upload and retrieve device screenshots
* **Remote Commands**: Execute commands on devices remotely

## Data Flow

1. Devices register via `/api/v1/devices/register`
2. Devices continuously push metrics, activities, and alerts
3. Data is stored in PostgreSQL and optionally forwarded to Mentor Backend
4. Dashboard queries data through Mentor Backend API

## Authentication

**MVP**: No authentication required. All endpoints are publicly accessible.

**Future**: JWT-based authentication recommended for production deployments.

## Error Handling

All endpoints return standard error responses:
* **400**: Bad Request - Invalid input or legacy fields
* **404**: Not Found - Resource does not exist
* **422**: Validation Error - Pydantic validation failed
* **500**: Internal Server Error - Server-side error

## Legacy Field Support

This API has migrated to canonical field names. Legacy field names are **rejected** with clear error messages:
* `id` → `deviceid`
* `name` → `device_name` (devices) or `process_name` (processes)
* `location` → `device_location`
* `type` → `activity_type` (activities) or `alert_type` (alerts)
* `command` → `command_text`
    """,
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
    contact={
        "name": "Raqeem API Support",
        "url": "https://github.com/mj-nehme/raqeem",
    },
    license_info={
        "name": "MIT",
        "url": "https://opensource.org/licenses/MIT",
    },
)

# Setup CORS
setup_cors(app)

app.include_router(api_router, prefix="/api/v1")


@app.get("/health", tags=["health"], summary="Health check endpoint")
async def health_check():
    """Check if the service is running and healthy."""
    return {"status": "ok", "service": "devices-backend"}
