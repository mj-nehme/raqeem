"""Metrics service for validating and calculating device metrics."""

from typing import Any

MAX_PERCENT = 100
MIN_VALUE = 0


class MetricsService:
    """Service for validating and calculating device metrics.

    This is a stub service primarily used for testing.
    Methods can be patched in tests to customize behavior.
    """

    def validate_cpu_usage(self, v: float) -> bool:
        """Validate that CPU usage is within valid range.

        Args:
            v: CPU usage percentage value

        Returns:
            True if CPU usage is between 0 and 100, False otherwise
        """
        return MIN_VALUE <= v <= MAX_PERCENT

    def validate_memory_usage(self, used: float, total: float | None) -> bool:
        """Validate that memory usage values are valid.

        Args:
            used: Amount of memory used (in bytes)
            total: Total memory available (in bytes)

        Returns:
            True if memory values are valid (used >= 0, used <= total), False otherwise

        Note:
            Both values must be non-negative and used cannot exceed total
        """
        if total is None or total < MIN_VALUE or used < MIN_VALUE:
            return False
        return used <= total

    def validate_disk_usage(self, used: float, total: float) -> bool:
        """Validate that disk usage values are valid.

        Args:
            used: Amount of disk space used (in bytes)
            total: Total disk space available (in bytes)

        Returns:
            True if disk values are valid (used >= 0, used <= total), False otherwise

        Note:
            Both values must be non-negative and used cannot exceed total
        """
        if total < MIN_VALUE or used < MIN_VALUE:
            return False
        return used <= total

    def validate_network_bytes(self, bytes_value: float) -> bool:
        """Validate that network bytes value is valid.

        Args:
            bytes_value: Network bytes value to validate

        Returns:
            True if value is non-negative, False otherwise
        """
        return bytes_value >= MIN_VALUE

    def calculate_average_metrics(self, data: list[dict[str, Any]]) -> dict[str, float]:
        """Calculate average CPU and memory metrics from a list of data points.

        Args:
            data: List of metric dictionaries containing cpu_usage and memory_used fields

        Returns:
            Dictionary with avg_cpu and avg_memory keys containing calculated averages

        Note:
            Returns zero values if input data is empty
        """
        if not data:
            return {"avg_cpu": 0.0, "avg_memory": 0.0}

        total_cpu = sum(d.get("cpu_usage", 0) for d in data)
        total_mem = sum(d.get("memory_used", 0) for d in data)
        count = len(data)

        return {
            "avg_cpu": total_cpu / count,
            "avg_memory": total_mem / count,
        }
