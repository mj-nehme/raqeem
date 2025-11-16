"""HTTP utility functions with retry logic"""

import asyncio
import logging
from typing import Any

import httpx

logger = logging.getLogger(__name__)

# HTTP status code constants
HTTP_STATUS_SERVER_ERROR = 500
HTTP_STATUS_TOO_MANY_REQUESTS = 429


async def post_with_retry(
    url: str,
    json: Any = None,
    max_retries: int = 3,
    timeout: float = 5.0,
    backoff_factor: float = 0.5,
) -> httpx.Response | None:
    """
    Post to a URL with exponential backoff retry logic.

    This function implements resilient HTTP communication with automatic retries
    for transient failures. It uses exponential backoff to avoid overwhelming
    servers during recovery and distinguishes between retryable errors (network
    issues, server errors, rate limits) and permanent failures (client errors).

    Retryable conditions:
    - Network errors (timeouts, connection failures)
    - Server errors (5xx status codes)
    - Rate limiting (429 Too Many Requests)

    Non-retryable conditions:
    - Successful responses (2xx, 3xx)
    - Client errors (4xx except 429)
    - Unexpected exceptions

    Args:
        url: The URL to post to
        json: JSON payload to send
        max_retries: Maximum number of retry attempts (default: 3)
        timeout: Request timeout in seconds (default: 5.0)
        backoff_factor: Factor for exponential backoff (default: 0.5)
                       delay = backoff_factor * (2 ** attempt)
                       e.g., delays: 0.5s, 1s, 2s for attempts 0, 1, 2

    Returns:
        Response object if successful, None if all retries failed
    """
    last_exception = None

    for attempt in range(max_retries + 1):
        try:
            async with httpx.AsyncClient(timeout=timeout) as client:
                response = await client.post(url, json=json)
                # Retry on 5xx errors and 429 rate limit
                # Success and client errors (4xx except 429) are not retried
                if response.status_code < HTTP_STATUS_SERVER_ERROR and response.status_code != HTTP_STATUS_TOO_MANY_REQUESTS:
                    return response

                logger.warning(
                    f"Request to {url} returned status {response.status_code} (attempt {attempt + 1}/{max_retries + 1})"
                )
                last_exception = Exception(f"HTTP {response.status_code}")
        except (httpx.TimeoutException, httpx.ConnectError, httpx.NetworkError) as e:
            # Network errors are transient and should be retried
            logger.warning(
                f"Request to {url} failed with {type(e).__name__}: {e} (attempt {attempt + 1}/{max_retries + 1})"
            )
            last_exception = e
        except Exception:
            # For unexpected errors (e.g., serialization errors), don't retry
            # as they're likely permanent issues with the request
            logger.exception(f"Unexpected error posting to {url}")
            return None

        # Implement exponential backoff between retries
        # Don't sleep after the last attempt since we're giving up
        if attempt < max_retries:
            delay = backoff_factor * (2**attempt)
            await asyncio.sleep(delay)

    logger.error(f"Request to {url} failed after {max_retries + 1} attempts: {last_exception}")
    return None
