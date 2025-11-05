from contextlib import asynccontextmanager
from fastapi import FastAPI, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import text
from app.api.routes import api_router
from app.db.init_db import init_db
from app.db.session import get_db
from app.core.cors import setup_cors
from app.core.exceptions import setup_exception_handlers
import logging

# Configure logging
logging.basicConfig(
	level=logging.INFO,
	format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
	# Startup
	try:
		logger.info("Initializing database...")
		await init_db()
		logger.info("Database initialized successfully")
	except Exception as e:
		# don't block startup if init fails here (DB may be initialized separately)
		logger.warning(f"Database initialization failed: {e}. Continuing startup...")
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

# Setup CORS - must be done before exception handlers
setup_cors(app)

# Setup exception handlers to ensure CORS headers on all responses
setup_exception_handlers(app)

app.include_router(api_router, prefix="/api/v1")


@app.get("/health")
async def health_check(db: AsyncSession = Depends(get_db)):
	"""
	Health check endpoint that verifies both API and database connectivity.
	"""
	health_status = {
		"status": "ok",
		"service": "devices-backend",
		"database": "unknown"
	}
	
	# Check database connectivity
	try:
		await db.execute(text("SELECT 1"))
		health_status["database"] = "connected"
	except Exception as e:
		logger.error(f"Database health check failed: {e}")
		health_status["status"] = "degraded"
		health_status["database"] = f"error: {str(e)}"
	
	return health_status

