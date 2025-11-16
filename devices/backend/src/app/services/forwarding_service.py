"""Service for forwarding data to mentor backend with retry and circuit breaker."""

import logging
from typing import Any, Dict, List

import httpx
from app.core.config import settings
from app.core.reliability import (
    CircuitBreaker,
    CircuitBreakerConfig,
    CircuitBreakerError,
    RetryConfig,
    retry_async,
)

logger = logging.getLogger(__name__)

# Global circuit breaker for mentor backend forwarding
_mentor_circuit_breaker = CircuitBreaker(
    "mentor-backend",
    CircuitBreakerConfig(max_failures=5, timeout=60.0, max_half_open_requests=2),
)


async def forward_alerts_to_mentor(
    device_id: str,
    alerts: List[Dict[str, Any]],
) -> bool:
    """
    Forward alerts to mentor backend with retry logic and circuit breaker.

    Args:
        device_id: Device identifier
        alerts: List of alert dictionaries

    Returns:
        True if forwarding succeeded, False otherwise
    """
    if not settings.mentor_api_url:
        return False

    retry_config = RetryConfig(
        max_attempts=3,
        initial_delay=0.5,
        max_delay=3.0,
        backoff_factor=2.0,
        timeout=10.0,
    )

    try:
        async def forward_operation():
            async with httpx.AsyncClient(timeout=10.0) as client:
                for alert in alerts:
                    payload = {
                        "deviceid": device_id,
                        "level": alert.get("level"),
                        "alert_type": alert.get("alert_type"),
                        "message": alert.get("message"),
                        "value": alert.get("value"),
                        "threshold": alert.get("threshold"),
                    }
                    response = await client.post(
                        f"{settings.mentor_api_url}/devices/{device_id}/alerts",
                        json=payload,
                    )
                    response.raise_for_status()

        # Execute through circuit breaker
        await _mentor_circuit_breaker.call(
            lambda: retry_async(
                retry_config,
                forward_operation,
                f"Forward alerts for device {device_id}",
            )
        )

        logger.info(f"Successfully forwarded {len(alerts)} alerts for device {device_id}")
        return True

    except CircuitBreakerError:
        logger.warning(
            f"Circuit breaker open for mentor backend, skipping alert forwarding for device {device_id}"
        )
        return False
    except Exception as e:
        logger.error(
            f"Failed to forward alerts to mentor backend for device {device_id}: {e}"
        )
        return False


async def forward_registration_to_mentor(device_data: Dict[str, Any]) -> bool:
    """
    Forward device registration to mentor backend with retry logic.

    Args:
        device_data: Device registration data

    Returns:
        True if forwarding succeeded, False otherwise
    """
    if not settings.mentor_api_url:
        return False

    retry_config = RetryConfig(
        max_attempts=3,
        initial_delay=0.5,
        max_delay=3.0,
        backoff_factor=2.0,
        timeout=10.0,
    )

    try:
        async def forward_operation():
            async with httpx.AsyncClient(timeout=10.0) as client:
                response = await client.post(
                    f"{settings.mentor_api_url}/devices/register",
                    json=device_data,
                )
                response.raise_for_status()

        await _mentor_circuit_breaker.call(
            lambda: retry_async(
                retry_config,
                forward_operation,
                f"Forward registration for device {device_data.get('deviceid')}",
            )
        )

        logger.info(f"Successfully forwarded registration for device {device_data.get('deviceid')}")
        return True

    except CircuitBreakerError:
        logger.warning("Circuit breaker open for mentor backend, skipping registration forwarding")
        return False
    except Exception as e:
        logger.error(f"Failed to forward registration to mentor backend: {e}")
        return False
