"""Tests for retry logic."""

import asyncio
import time

import pytest
from app.core.reliability.retry import (
    RetryConfig,
    default_retry_config,
    retry_async,
    retry_with_backoff,
)


def test_retry_with_backoff_success():
    """Test successful retry after initial failure."""
    attempts = []

    def operation():
        attempts.append(1)
        if len(attempts) < 2:
            msg = "Temporary error"
            raise ValueError(msg)
        return "success"

    config = RetryConfig(max_attempts=3, initial_delay=0.01, max_delay=0.1)
    result = retry_with_backoff(config, operation, "test_operation")

    assert result == "success"
    assert len(attempts) == 2


def test_retry_with_backoff_max_attempts():
    """Test that retry stops after max attempts."""
    attempts = []

    def operation():
        attempts.append(1)
        msg = "Persistent error"
        raise ValueError(msg)

    config = RetryConfig(max_attempts=3, initial_delay=0.01, max_delay=0.1)

    with pytest.raises(Exception) as exc_info:
        retry_with_backoff(config, operation, "test_operation")

    assert "failed after 3 attempts" in str(exc_info.value)
    assert len(attempts) == 3


def test_retry_config_presets():
    """Test that preset configs have expected values."""
    default = default_retry_config()
    assert default.max_attempts == 3
    assert default.initial_delay == 0.1


@pytest.mark.asyncio
async def test_retry_async_success():
    """Test successful async retry after initial failure."""
    attempts = []

    async def operation():
        attempts.append(1)
        if len(attempts) < 2:
            msg = "Temporary error"
            raise ValueError(msg)
        return "success"

    config = RetryConfig(max_attempts=3, initial_delay=0.01, max_delay=0.1)
    result = await retry_async(config, operation, "test_async_operation")

    assert result == "success"
    assert len(attempts) == 2


@pytest.mark.asyncio
async def test_retry_async_with_timeout():
    """Test async retry with timeout."""
    attempts = []

    async def operation():
        attempts.append(1)
        await asyncio.sleep(0.5)  # Simulate slow operation
        return "success"

    config = RetryConfig(
        max_attempts=3, initial_delay=0.01, max_delay=0.1, timeout=0.1
    )

    with pytest.raises(Exception) as exc_info:
        await retry_async(config, operation, "test_timeout_operation")

    assert "failed after" in str(exc_info.value)


@pytest.mark.asyncio
async def test_retry_async_exponential_backoff():
    """Test that delays increase exponentially."""
    delays = []
    attempts = []

    async def operation():
        if len(attempts) > 0:
            # Record time since last attempt
            delays.append(time.time() - attempts[-1])
        attempts.append(time.time())
        msg = "Error"
        raise ValueError(msg)

    config = RetryConfig(
        max_attempts=3, initial_delay=0.05, max_delay=1.0, backoff_factor=2.0
    )

    with pytest.raises(Exception):
        await retry_async(config, operation, "test_backoff")

    # Verify delays increase (with some tolerance for timing variations)
    assert len(delays) == 2
    assert delays[1] > delays[0]
