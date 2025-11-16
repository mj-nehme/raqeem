"""Alert service for determining alert levels based on metric thresholds."""

# Threshold deltas for alert level determination
WARNING_DELTA = 10
ERROR_DELTA = 20
CRITICAL_DELTA = 30


class AlertService:
    """Service for creating and managing device alerts.

    Provides functionality to determine when alerts should be created
    and what severity level they should have based on metric values
    and configured thresholds.
    """

    def should_create_alert(self, value: float, threshold: float) -> bool:
        """Determine if an alert should be created based on value and threshold.

        Args:
            value: The current metric value
            threshold: The configured threshold for this metric

        Returns:
            True if value meets or exceeds threshold, False otherwise
        """
        return value >= threshold

    def determine_alert_level(self, value: float, threshold: float) -> str:
        """Determine the severity level of an alert.

        Args:
            value: The current metric value
            threshold: The configured threshold for this metric

        Returns:
            Alert level string: "info", "warning", "error", or "critical"

        Note:
            - Below threshold: "info"
            - 0-10 above threshold: "warning"
            - 10-20 above threshold: "error"
            - 20+ above threshold: "critical"
        """
        delta = value - threshold

        if delta < 0:
            return "info"
        if delta < WARNING_DELTA:
            return "warning"
        if delta < ERROR_DELTA:
            return "error"

        return "critical"

    def calculate_threshold_percentage(self, value: float, threshold: float) -> float:
        """Calculate how much a value exceeds the threshold as a percentage.

        Args:
            value: The current metric value
            threshold: The configured threshold for this metric

        Returns:
            Percentage over threshold (0 if below threshold)

        Note:
            Returns 0 if threshold is 0 to avoid division by zero
        """
        if threshold == 0:
            return 0.0

        if value < threshold:
            return 0.0

        return ((value - threshold) / threshold) * 100
