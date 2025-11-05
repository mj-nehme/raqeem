import os
from fastapi.middleware.cors import CORSMiddleware
import logging

logger = logging.getLogger(__name__)


def setup_cors(app):
    """
    Setup CORS middleware with dynamic origin detection.
    
    CORS headers must be present on ALL responses, including error responses.
    The exception handlers in app.core.exceptions are configured to ensure
    CORS headers are added even when exceptions occur.
    """
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
    
    logger.info(f"CORS configured for origins: {origins}")

    app.add_middleware(
        CORSMiddleware,
        allow_origins=origins,
        allow_credentials=False,  # False to avoid CORS issues with FormData/file uploads
        allow_methods=["*"],
        allow_headers=["*"],
        expose_headers=["*"],  # Allow browsers to access all response headers
    )
