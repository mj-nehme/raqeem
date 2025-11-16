"""Circuit breaker pattern implementation for fault tolerance."""

import asyncio
import logging
import time
from collections.abc import Callable
from dataclasses import dataclass
from enum import Enum
from typing import Any

logger = logging.getLogger(__name__)


class CircuitState(Enum):
    """Circuit breaker states."""

    CLOSED = "closed"  # Normal operation, requests pass through
    OPEN = "open"  # Circuit is open, requests fail immediately
    HALF_OPEN = "half_open"  # Testing if service recovered


class CircuitBreakerError(Exception):
    """Raised when circuit breaker is open."""



@dataclass
class CircuitBreakerConfig:
    """Configuration for circuit breaker behavior."""

    max_failures: int = 5
    timeout: float = 30.0  # seconds before transitioning from Open to Half-Open
    max_half_open_requests: int = 3


class CircuitBreaker:
    """
    Circuit breaker implementation for protecting against cascading failures.

    Usage:
        breaker = CircuitBreaker("my-service")
        result = await breaker.call(my_async_function)
    """

    def __init__(
        self,
        name: str,
        config: CircuitBreakerConfig | None = None,
    ):
        """
        Initialize circuit breaker.

        Args:
            name: Name for logging and identification
            config: Configuration (uses defaults if not provided)
        """
        self.name = name
        self.config = config or CircuitBreakerConfig()
        self.state = CircuitState.CLOSED
        self.failures = 0
        self.last_failure_time = 0.0
        self.half_open_requests = 0
        self.half_open_successes = 0
        self._lock = asyncio.Lock()

    async def call(self, operation: Callable[[], Any]) -> Any:
        """
        Execute an async operation through the circuit breaker.

        Args:
            operation: Async function to execute

        Returns:
            Result from the operation

        Raises:
            CircuitBreakerError: If circuit is open
            Exception: Any exception from the operation
        """
        if not await self._allow_request():
            raise CircuitBreakerError(f"Circuit breaker '{self.name}' is open")

        try:
            result = await operation()
            await self._on_success()
            return result
        except Exception:
            await self._on_failure()
            raise

    async def _allow_request(self) -> bool:
        """Check if a request should be allowed based on current state."""
        async with self._lock:
            now = time.time()

            if self.state == CircuitState.CLOSED:
                return True

            if self.state == CircuitState.OPEN:
                # Check if we should transition to Half-Open
                if now - self.last_failure_time >= self.config.timeout:
                    logger.info(
                        f"Circuit breaker '{self.name}': Transitioning from Open to Half-Open"
                    )
                    self.state = CircuitState.HALF_OPEN
                    self.half_open_requests = 0
                    self.half_open_successes = 0
                    return True
                return False

            if self.state == CircuitState.HALF_OPEN:
                # Allow limited requests in Half-Open state
                if self.half_open_requests < self.config.max_half_open_requests:
                    self.half_open_requests += 1
                    return True
                return False

            return False

    async def _on_success(self) -> None:
        """Handle a successful operation."""
        async with self._lock:
            if self.state == CircuitState.CLOSED:
                # Reset failure count on success in Closed state
                self.failures = 0

            elif self.state == CircuitState.HALF_OPEN:
                self.half_open_successes += 1
                # If we've had enough successes, close the circuit
                if self.half_open_successes >= self.config.max_half_open_requests:
                    logger.info(
                        f"Circuit breaker '{self.name}': Transitioning from Half-Open to Closed"
                    )
                    self.state = CircuitState.CLOSED
                    self.failures = 0
                    self.half_open_requests = 0
                    self.half_open_successes = 0

    async def _on_failure(self) -> None:
        """Handle a failed operation."""
        async with self._lock:
            self.failures += 1
            self.last_failure_time = time.time()

            if self.state == CircuitState.CLOSED:
                if self.failures >= self.config.max_failures:
                    logger.warning(
                        f"Circuit breaker '{self.name}': Transitioning from Closed to Open "
                        f"after {self.failures} failures"
                    )
                    self.state = CircuitState.OPEN

            elif self.state == CircuitState.HALF_OPEN:
                # Any failure in Half-Open state reopens the circuit
                logger.warning(
                    f"Circuit breaker '{self.name}': Transitioning from Half-Open to Open "
                    f"due to failure"
                )
                self.state = CircuitState.OPEN
                self.half_open_requests = 0
                self.half_open_successes = 0

    def get_state(self) -> CircuitState:
        """Get current circuit breaker state."""
        return self.state

    async def reset(self) -> None:
        """Manually reset circuit breaker to Closed state."""
        async with self._lock:
            logger.info(f"Circuit breaker '{self.name}': Manual reset to Closed state")
            self.state = CircuitState.CLOSED
            self.failures = 0
            self.half_open_requests = 0
            self.half_open_successes = 0
