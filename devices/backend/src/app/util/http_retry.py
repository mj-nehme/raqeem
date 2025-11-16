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

    Args:
        url: The URL to post to
        json: JSON payload to send
        max_retries: Maximum number of retry attempts
        timeout: Request timeout in seconds
        backoff_factor: Factor for exponential backoff (delay = backoff_factor * (2 ** attempt))

    Returns:
        Response object if successful, None if all retries failed
    """
    last_exception = None

    for attempt in range(max_retries + 1):
        try:
            async with httpx.AsyncClient(timeout=timeout) as client:
                response = await client.post(url, json=json)
                # Retry on 5xx errors and 429 rate limit
                if response.status_code < HTTP_STATUS_SERVER_ERROR and response.status_code != HTTP_STATUS_TOO_MANY_REQUESTS:
                    return response

                logger.warning(
                    f"Request to {url} returned status {response.status_code} (attempt {attempt + 1}/{max_retries + 1})"
                )
                last_exception = Exception(f"HTTP {response.status_code}")
        except (httpx.TimeoutException, httpx.ConnectError, httpx.NetworkError) as e:
            logger.warning(
                f"Request to {url} failed with {type(e).__name__}: {e} (attempt {attempt + 1}/{max_retries + 1})"
            )
            last_exception = e
        except Exception:
            # For unexpected errors, don't retry
            logger.exception(f"Unexpected error posting to {url}")
            return None

        # Don't sleep after the last attempt
        if attempt < max_retries:
            delay = backoff_factor * (2**attempt)
            await asyncio.sleep(delay)

    logger.error(f"Request to {url} failed after {max_retries + 1} attempts: {last_exception}")
    return None
