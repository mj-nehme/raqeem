"""Tests for circuit breaker implementation."""

import asyncio

import pytest
from app.core.reliability.circuit_breaker import (
    CircuitBreaker,
    CircuitBreakerConfig,
    CircuitBreakerError,
    CircuitState,
)


@pytest.mark.asyncio
async def test_circuit_breaker_initial_state():
    """Test that circuit breaker starts in closed state."""
    cb = CircuitBreaker("test", CircuitBreakerConfig())
    assert cb.get_state() == CircuitState.CLOSED


@pytest.mark.asyncio
async def test_circuit_breaker_opens_after_failures():
    """Test that circuit opens after max failures."""
    config = CircuitBreakerConfig(max_failures=3, timeout=1.0)
    cb = CircuitBreaker("test", config)

    # Cause failures
    for _ in range(3):
        try:
            await cb.call(async_error_operation)
        except ValueError:
            pass

    assert cb.get_state() == CircuitState.OPEN

    # Next call should fail immediately with CircuitBreakerError
    with pytest.raises(CircuitBreakerError):
        await cb.call(async_success_operation)


@pytest.mark.asyncio
async def test_circuit_breaker_half_open_transition():
    """Test transition from Open to Half-Open after timeout."""
    config = CircuitBreakerConfig(max_failures=2, timeout=0.1)
    cb = CircuitBreaker("test", config)

    # Open the circuit
    for _ in range(2):
        try:
            await cb.call(async_error_operation)
        except ValueError:
            pass

    assert cb.get_state() == CircuitState.OPEN

    # Wait for timeout
    await asyncio.sleep(0.15)

    # Next call should be allowed (transitions to Half-Open)
    result = await cb.call(async_success_operation)
    assert result == "success"


@pytest.mark.asyncio
async def test_circuit_breaker_closes_after_successes():
    """Test that circuit closes after successful requests in Half-Open."""
    config = CircuitBreakerConfig(
        max_failures=2, timeout=0.1, max_half_open_requests=2
    )
    cb = CircuitBreaker("test", config)

    # Open the circuit
    for _ in range(2):
        try:
            await cb.call(async_error_operation)
        except ValueError:
            pass

    # Wait for timeout
    await asyncio.sleep(0.15)

    # Succeed in Half-Open state
    for _ in range(2):
        await cb.call(async_success_operation)

    assert cb.get_state() == CircuitState.CLOSED


@pytest.mark.asyncio
async def test_circuit_breaker_reopens_on_half_open_failure():
    """Test that circuit reopens if request fails in Half-Open."""
    config = CircuitBreakerConfig(max_failures=2, timeout=0.1)
    cb = CircuitBreaker("test", config)

    # Open the circuit
    for _ in range(2):
        try:
            await cb.call(async_error_operation)
        except ValueError:
            pass

    # Wait for timeout
    await asyncio.sleep(0.15)

    # Fail in Half-Open state
    try:
        await cb.call(async_error_operation)
    except ValueError:
        pass

    assert cb.get_state() == CircuitState.OPEN


@pytest.mark.asyncio
async def test_circuit_breaker_reset():
    """Test manual reset of circuit breaker."""
    config = CircuitBreakerConfig(max_failures=2)
    cb = CircuitBreaker("test", config)

    # Open the circuit
    for _ in range(2):
        try:
            await cb.call(async_error_operation)
        except ValueError:
            pass

    assert cb.get_state() == CircuitState.OPEN

    # Reset
    await cb.reset()
    assert cb.get_state() == CircuitState.CLOSED

    # Should allow requests
    result = await cb.call(async_success_operation)
    assert result == "success"


@pytest.mark.asyncio
async def test_circuit_breaker_success_resets_count():
    """Test that success in Closed state resets failure count."""
    config = CircuitBreakerConfig(max_failures=3)
    cb = CircuitBreaker("test", config)

    # Some failures
    for _ in range(2):
        try:
            await cb.call(async_error_operation)
        except ValueError:
            pass

    # Success should reset count
    await cb.call(async_success_operation)

    # Should still be closed
    assert cb.get_state() == CircuitState.CLOSED

    # Two more failures shouldn't open it
    for _ in range(2):
        try:
            await cb.call(async_error_operation)
        except ValueError:
            pass

    assert cb.get_state() == CircuitState.CLOSED


# Helper functions for tests


async def async_success_operation():
    """Async operation that succeeds."""
    return "success"


async def async_error_operation():
    """Async operation that raises an error."""
    raise ValueError("Test error")
