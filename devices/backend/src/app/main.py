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
    description="High-throughput telemetry ingestion API for IoT device monitoring. Handles device registration, metrics collection, activity logging, and alert submission.",
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
)

# Setup CORS
setup_cors(app)

app.include_router(api_router, prefix="/api/v1")


@app.get("/health")
async def health_check():
    return {"status": "ok", "service": "devices-backend"}
