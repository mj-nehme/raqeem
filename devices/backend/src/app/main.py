import json
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

import logging

from fastapi import FastAPI

from app.api.routes import api_router
from app.core.cors import setup_cors
from app.core.middleware import CorrelationIDMiddleware
from app.db.session import engine

logger = logging.getLogger(__name__)


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
    description="High-throughput telemetry ingestion API for IoT device monitoring. Handles device registration, metrics collection, activity logging, and alert submission.",
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
)

# Add correlation ID middleware
app.add_middleware(CorrelationIDMiddleware)

# Setup CORS
setup_cors(app)

app.include_router(api_router, prefix="/api/v1")


# Import health check router
from app.api.v1.endpoints.health import router as health_router
app.include_router(health_router)

# Keep legacy health endpoint for backwards compatibility
@app.get("/health")
async def legacy_health_check():
    return {"status": "ok", "service": "devices-backend"}
