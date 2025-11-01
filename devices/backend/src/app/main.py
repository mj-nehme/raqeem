from contextlib import asynccontextmanager
from fastapi import FastAPI
from app.api.routes import api_router
from app.db.init_db import init_db
from app.core.cors import setup_cors


@asynccontextmanager
async def lifespan(app: FastAPI):
	# Startup
	try:
		await init_db()
	except Exception:
		# don't block startup if init fails here (DB may be initialized separately)
		pass
	yield
	# Shutdown: nothing to clean up currently


app = FastAPI(lifespan=lifespan)

# Setup CORS
setup_cors(app)

app.include_router(api_router, prefix="/api/v1")


@app.get("/health")
async def health_check():
	return {"status": "ok", "service": "devices-backend"}

