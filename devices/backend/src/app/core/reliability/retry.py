"""Retry logic with exponential backoff for resilient external calls."""

import asyncio
import logging
import time
from collections.abc import Callable
from dataclasses import dataclass
from typing import Any, TypeVar

logger = logging.getLogger(__name__)

T = TypeVar("T")


@dataclass
class RetryConfig:
    """Configuration for retry behavior."""

    max_attempts: int = 3
    initial_delay: float = 0.1
    max_delay: float = 5.0
    backoff_factor: float = 2.0
    timeout: float | None = None


def default_retry_config() -> RetryConfig:
    """Returns sensible default retry configuration."""
    return RetryConfig(
        max_attempts=3,
        initial_delay=0.1,
        max_delay=5.0,
        backoff_factor=2.0,
    )


def database_retry_config() -> RetryConfig:
    """Returns retry configuration optimized for database operations."""
    return RetryConfig(
        max_attempts=5,
        initial_delay=0.2,
        max_delay=10.0,
        backoff_factor=2.0,
    )


def external_service_retry_config() -> RetryConfig:
    """Returns retry configuration for external service calls."""
    return RetryConfig(
        max_attempts=3,
        initial_delay=0.5,
        max_delay=5.0,
        backoff_factor=2.0,
    )


def retry_with_backoff(
    config: RetryConfig,
    operation: Callable[[], T],
    operation_name: str = "operation",
) -> T:
    """
    Execute a synchronous function with exponential backoff retry logic.

    Args:
        config: Retry configuration
        operation: Function to execute
        operation_name: Name for logging purposes

    Returns:
        Result from the operation

    Raises:
        Exception: The last exception if all retries failed
    """
    delay = config.initial_delay
    last_exception = None

    for attempt in range(1, config.max_attempts + 1):
        try:
            result = operation()
            if attempt > 1:
                logger.info(f"{operation_name} succeeded after {attempt} attempts")
            return result
        except Exception as e:
            last_exception = e

            if attempt == config.max_attempts:
                break

            logger.warning(
                f"{operation_name} failed (attempt {attempt}/{config.max_attempts}): {e}. "
                f"Retrying in {delay:.2f}s..."
            )

            time.sleep(delay)

            # Calculate next delay with exponential backoff
            delay = min(delay * config.backoff_factor, config.max_delay)

    error_msg = f"{operation_name} failed after {config.max_attempts} attempts: {last_exception}"
    logger.error(error_msg)
    raise Exception(error_msg) from last_exception


async def retry_async(
    config: RetryConfig,
    operation: Callable[[], Any],
    operation_name: str = "operation",
) -> Any:
    """
    Execute an async function with exponential backoff retry logic.

    Args:
        config: Retry configuration
        operation: Async function to execute
        operation_name: Name for logging purposes

    Returns:
        Result from the operation

    Raises:
        Exception: The last exception if all retries failed
    """
    delay = config.initial_delay
    last_exception = None

    for attempt in range(1, config.max_attempts + 1):
        try:
            if config.timeout:
                result = await asyncio.wait_for(operation(), timeout=config.timeout)
            else:
                result = await operation()

            if attempt > 1:
                logger.info(f"{operation_name} succeeded after {attempt} attempts")
            return result
        except TimeoutError as e:
            last_exception = e
            if attempt == config.max_attempts:
                break

            logger.warning(
                f"{operation_name} timed out (attempt {attempt}/{config.max_attempts}). "
                f"Retrying in {delay:.2f}s..."
            )
        except Exception as e:
            last_exception = e

            if attempt == config.max_attempts:
                break

            logger.warning(
                f"{operation_name} failed (attempt {attempt}/{config.max_attempts}): {e}. "
                f"Retrying in {delay:.2f}s..."
            )

        await asyncio.sleep(delay)

        # Calculate next delay with exponential backoff
        delay = min(delay * config.backoff_factor, config.max_delay)

    error_msg = f"{operation_name} failed after {config.max_attempts} attempts: {last_exception}"
    logger.error(error_msg)
    raise Exception(error_msg) from last_exception
