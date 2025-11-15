class AlertService:
    """Stub AlertService for tests."""
    def should_create_alert(self, value: float, threshold: float) -> bool:
        return value >= threshold
    def determine_alert_level(self, value: float, threshold: float) -> str:
        delta = value - threshold
        if delta < 0:
            return "info"
        if delta < 10:
            return "warning"
        if delta < 20:
            return "error"
        return "critical"
