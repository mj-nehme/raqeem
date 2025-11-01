from fastapi import FastAPI
from app.api.routes import api_router
from app.db.init_db import init_db
from app.core.cors import setup_cors

app = FastAPI()

# Setup CORS
setup_cors(app)

app.include_router(api_router, prefix="/api/v1")


@app.get("/health")
async def health_check():
	return {"status": "ok", "service": "devices-backend"}


@app.on_event("startup")
async def on_startup():
	# ensure DB tables exist
	try:
		await init_db()
	except Exception:
		# don't block startup if init fails here (DB may be initialized separately)
		pass
