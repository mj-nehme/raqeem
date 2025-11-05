"""
Custom exception handlers to ensure CORS headers are present on all responses,
including error responses.
"""
from fastapi import Request, status
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from starlette.exceptions import HTTPException as StarletteHTTPException
import logging
import os
import traceback

logger = logging.getLogger(__name__)


def get_allowed_origins() -> list:
    """
    Get list of allowed origins from environment variable.
    This should match the CORS middleware configuration.
    """
    raw = os.getenv("FRONTEND_ORIGINS", "")
    origins = [o.strip() for o in raw.split(",") if o.strip()]
    
    if len(origins) == 0:
        origins = [
            "http://localhost:4000",
            "http://localhost:4001",
            "http://localhost:4002",
            "http://localhost:5000",
            "http://localhost:5001",
            "http://localhost:5002",
        ]
    
    return origins


def get_cors_headers(request: Request) -> dict:
    """
    Get CORS headers that should be added to error responses.
    This ensures CORS headers are present even when exceptions occur before
    the CORS middleware can process the response.
    """
    origin = request.headers.get("origin", "")
    allowed_origins = get_allowed_origins()
    
    headers = {}
    
    # Only add CORS headers if origin is in allowed list
    if origin in allowed_origins:
        headers["Access-Control-Allow-Origin"] = origin
        headers["Access-Control-Allow-Methods"] = "*"
        headers["Access-Control-Allow-Headers"] = "*"
        headers["Access-Control-Expose-Headers"] = "*"
        logger.debug(f"Adding CORS headers for origin: {origin}")
    else:
        logger.debug(f"Origin not in allowed list: {origin}")
    
    return headers


async def http_exception_handler(request: Request, exc: StarletteHTTPException) -> JSONResponse:
    """
    Handle HTTP exceptions and ensure CORS headers are present.
    """
    logger.error(
        f"HTTP exception on {request.method} {request.url.path}: "
        f"{exc.status_code} - {exc.detail}"
    )
    
    headers = get_cors_headers(request)
    
    return JSONResponse(
        status_code=exc.status_code,
        content={"detail": exc.detail},
        headers=headers
    )


async def validation_exception_handler(request: Request, exc: RequestValidationError) -> JSONResponse:
    """
    Handle validation errors and ensure CORS headers are present.
    """
    logger.error(
        f"Validation error on {request.method} {request.url.path}: {exc.errors()}"
    )
    
    headers = get_cors_headers(request)
    
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content={"detail": exc.errors()},
        headers=headers
    )


async def general_exception_handler(request: Request, exc: Exception) -> JSONResponse:
    """
    Catch-all exception handler to ensure CORS headers are present on all errors.
    This is critical for preventing CORS errors when internal server errors occur.
    """
    # Log full exception with traceback
    logger.exception(
        f"Unhandled exception on {request.method} {request.url.path}: "
        f"{type(exc).__name__}: {str(exc)}"
    )
    
    # Log traceback for debugging
    tb = "".join(traceback.format_exception(type(exc), exc, exc.__traceback__))
    logger.error(f"Traceback:\n{tb}")
    
    headers = get_cors_headers(request)
    
    # Provide more detailed error messages in development
    error_detail = "Internal server error"
    
    # Include exception type and message for debugging
    # In production, you might want to hide these details
    error_detail = f"{type(exc).__name__}: {str(exc)}"
    
    # Add hints for common errors
    hints = []
    if "database" in str(exc).lower() or "connection" in str(exc).lower():
        hints.append("Database connection issue detected. Check DATABASE_URL and database availability.")
    
    if "asyncpg" in str(exc).lower():
        hints.append("PostgreSQL async driver issue. Verify database is running and accessible.")
    
    response_content = {
        "detail": error_detail,
        "error_type": type(exc).__name__
    }
    
    if hints:
        response_content["hints"] = hints
    
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content=response_content,
        headers=headers
    )


def setup_exception_handlers(app):
    """
    Register all exception handlers with the FastAPI app.
    This must be called after CORS middleware is added.
    """
    logger.info("Setting up exception handlers for CORS error prevention")
    app.add_exception_handler(StarletteHTTPException, http_exception_handler)
    app.add_exception_handler(RequestValidationError, validation_exception_handler)
    app.add_exception_handler(Exception, general_exception_handler)
