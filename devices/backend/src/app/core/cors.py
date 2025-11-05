import os
from fastapi.middleware.cors import CORSMiddleware

def setup_cors(app):
    # Comma-separated list of allowed origins from env FRONTEND_ORIGINS
    raw = os.getenv("FRONTEND_ORIGINS", "")
    origins = [o.strip() for o in raw.split(",") if o.strip()]

    # If not provided, use common default ports for local development
    # Using specific origins instead of "*" to ensure browser compatibility with file uploads
    if len(origins) == 0:
        origins = [
            "http://localhost:4000",
            "http://localhost:4001",
            "http://localhost:4002",
            "http://localhost:5000",
            "http://localhost:5001",
            "http://localhost:5002",
        ]

    app.add_middleware(
        CORSMiddleware,
        allow_origins=origins,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
