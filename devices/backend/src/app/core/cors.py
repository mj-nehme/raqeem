import os
from fastapi.middleware.cors import CORSMiddleware

def setup_cors(app):
    # Comma-separated list of allowed origins from env FRONTEND_ORIGINS
    raw = os.getenv("FRONTEND_ORIGINS", "")
    origins = [o.strip() for o in raw.split(",") if o.strip()]

    # If not provided, allow all origins without credentials (dev-friendly, no fixed ports)
    allow_all = len(origins) == 0

    app.add_middleware(
        CORSMiddleware,
        allow_origins=origins if not allow_all else ["*"],
        allow_credentials=False if allow_all else True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
