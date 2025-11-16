"""Alert service for determining alert levels and thresholds.

This service provides logic for:
- Determining when to create alerts based on metric values
- Classifying alert severity levels
- Calculating alert thresholds with configurable deltas
"""

from app.core.logging_config import get_logger

logger = get_logger(__name__)

# Threshold deltas for alert level classification
WARNING_DELTA = 10  # Threshold exceeded by 0-10 units
ERROR_DELTA = 20  # Threshold exceeded by 10-20 units
# Anything >= 20 units over threshold is CRITICAL


class AlertService:
    """Service for alert creation and severity determination.

    This service evaluates metric values against thresholds and determines:
    - Whether an alert should be created
    - The severity level of the alert (info, warning, error, critical)

    Example:
        >>> service = AlertService()
        >>> service.should_create_alert(85.0, 80.0)  # True
        >>> service.determine_alert_level(85.0, 80.0)  # "warning"
        >>> service.determine_alert_level(105.0, 80.0)  # "critical"
    """

    def __init__(self):
        """Initialize the AlertService."""
        logger.debug("AlertService initialized")

    def should_create_alert(self, value: float, threshold: float) -> bool:
        """Determine if an alert should be created.

        Args:
            value: The current metric value.
            threshold: The threshold value to compare against.

        Returns:
            True if value meets or exceeds threshold, False otherwise.

        Example:
            >>> service.should_create_alert(85.0, 80.0)
            True
            >>> service.should_create_alert(75.0, 80.0)
            False
        """
        should_alert = value >= threshold

        if should_alert:
            logger.debug(
                "Alert threshold exceeded",
                extra={
                    "value": value,
                    "threshold": threshold,
                    "delta": value - threshold,
                },
            )

        return should_alert

    def determine_alert_level(self, value: float, threshold: float) -> str:
        """Determine the severity level of an alert based on how much the threshold is exceeded.

        Alert levels are determined by the delta (difference) between value and threshold:
        - info: Below threshold (delta < 0)
        - warning: 0 <= delta < WARNING_DELTA (default: 10)
        - error: WARNING_DELTA <= delta < ERROR_DELTA (default: 10-20)
        - critical: delta >= ERROR_DELTA (default: >= 20)

        Args:
            value: The current metric value.
            threshold: The threshold value to compare against.

        Returns:
            String indicating alert level: "info", "warning", "error", or "critical".

        Example:
            >>> service.determine_alert_level(75.0, 80.0)  # Below threshold
            "info"
            >>> service.determine_alert_level(85.0, 80.0)  # 5 over
            "warning"
            >>> service.determine_alert_level(95.0, 80.0)  # 15 over
            "error"
            >>> service.determine_alert_level(105.0, 80.0)  # 25 over
            "critical"
        """
        delta = value - threshold

        # Determine level based on delta
        if delta < 0:
            level = "info"
        elif delta < WARNING_DELTA:
            level = "warning"
        elif delta < ERROR_DELTA:
            level = "error"
        else:
            level = "critical"

        logger.debug(
            "Alert level determined",
            extra={
                "value": value,
                "threshold": threshold,
                "delta": delta,
                "level": level,
            },
        )

        return level
