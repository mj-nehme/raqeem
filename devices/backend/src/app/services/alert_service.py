WARNING_DELTA = 10
ERROR_DELTA = 20


class AlertService:
    """Stub AlertService for tests."""

    def should_create_alert(self, value: float, threshold: float) -> bool:
        return value >= threshold

    def determine_alert_level(self, value: float, threshold: float) -> str:
        delta = value - threshold
        if delta < 0:
            return "info"
        if delta < WARNING_DELTA:
            return "warning"
        if delta < ERROR_DELTA:
            return "error"
        return "critical"
