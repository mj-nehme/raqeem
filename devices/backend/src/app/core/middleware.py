"""Middleware components for request handling and observability.

This module provides middleware for:
- Request ID generation and propagation
- Request/response logging
- Error tracking and monitoring
"""

import time
import uuid
from collections.abc import Awaitable, Callable

from app.core.logging_config import get_logger
from fastapi import Request, Response
from starlette.middleware.base import BaseHTTPMiddleware

logger = get_logger(__name__)


class RequestIDMiddleware(BaseHTTPMiddleware):
    """Middleware to generate and propagate request IDs for distributed tracing.

    Adds a unique request ID to each request, which can be used for:
    - Correlating logs across services
    - Tracking requests through the system
    - Debugging and troubleshooting

    The request ID is:
    - Generated as a UUID if not provided
    - Extracted from X-Request-ID header if present
    - Added to response headers for client tracking
    """

    async def dispatch(self, request: Request, call_next: Callable[[Request], Awaitable[Response]]) -> Response:
        # Extract or generate request ID
        request_id = request.headers.get("X-Request-ID") or str(uuid.uuid4())

        # Store request ID in request state for access in handlers
        request.state.request_id = request_id

        # Track request timing
        start_time = time.time()

        # Log incoming request
        logger.info(
            "Request started",
            extra={
                "request_id": request_id,
                "method": request.method,
                "path": request.url.path,
                "client": request.client.host if request.client else "unknown",
            },
        )

        # Process request
        try:
            response = await call_next(request)

            # Calculate request duration
            duration = time.time() - start_time

            # Log completed request
            logger.info(
                "Request completed",
                extra={
                    "request_id": request_id,
                    "method": request.method,
                    "path": request.url.path,
                    "status_code": response.status_code,
                    "duration_ms": round(duration * 1000, 2),
                },
            )

            # Add request ID to response headers
            response.headers["X-Request-ID"] = request_id

            return response

        except Exception as e:
            # Log failed request
            duration = time.time() - start_time
            logger.error(
                "Request failed",
                extra={
                    "request_id": request_id,
                    "method": request.method,
                    "path": request.url.path,
                    "error": str(e),
                    "error_type": type(e).__name__,
                    "duration_ms": round(duration * 1000, 2),
                },
                exc_info=True,
            )
            raise


def get_request_id(request: Request) -> str:
    """Get the request ID from the current request.

    Args:
        request: FastAPI request object.

    Returns:
        Request ID string, or "unknown" if not available.

    Example:
        >>> @router.get("/example")
        >>> async def example(request: Request):
        ...     request_id = get_request_id(request)
        ...     logger.info("Processing", extra={"request_id": request_id})
    """
    return getattr(request.state, "request_id", "unknown")
