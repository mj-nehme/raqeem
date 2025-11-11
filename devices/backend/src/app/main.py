from contextlib import asynccontextmanager
from fastapi import FastAPI
from app.api.routes import api_router
from app.db.init_db import init_db
from app.core.cors import setup_cors

@asynccontextmanager
async def lifespan(app: FastAPI):
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
	openapi_url="/openapi.json"
)

# Setup CORS
setup_cors(app)

app.include_router(api_router, prefix="/api/v1")


@app.get("/health")
async def health_check():
	return {"status": "ok", "service": "devices-backend"}

