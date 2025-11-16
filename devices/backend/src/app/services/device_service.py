"""Device service for validating and managing device data."""

from datetime import datetime, timedelta, timezone
from typing import Any

# Device will be considered offline if not seen for this duration
DEVICE_ONLINE_THRESHOLD_MINUTES = 5

# Supported device types
VALID_DEVICE_TYPES = {"laptop", "desktop", "server", "mobile", "tablet"}


class DeviceService:
    """Service for validating and managing device data.

    This is a stub service primarily used for testing.
    Methods can be patched in tests to customize behavior.
    """

    def validate_device_data(self, data: dict[str, Any]) -> bool:
        """Validate device data structure.

        Args:
            data: Dictionary containing device data fields

        Returns:
            True if device data is valid, False otherwise

        Note:
            Checks for required fields and validates device_type if present
        """
        # Check for required deviceid field
        if "deviceid" not in data:
            return False

        # Validate device_type if provided
        if "device_type" in data and data["device_type"] is not None:
            if not self.validate_device_type(data["device_type"]):
                return False

        return True

    def is_device_online(self, last_seen: datetime) -> bool:
        """Check if a device is currently online based on last seen timestamp.

        Args:
            last_seen: The last time the device was seen

        Returns:
            True if device is considered online (seen within threshold), False otherwise

        Note:
            Considers device offline if not seen within DEVICE_ONLINE_THRESHOLD_MINUTES
        """
        now = datetime.now(timezone.utc)
        # Ensure last_seen is timezone-aware for comparison
        if last_seen.tzinfo is None:
            last_seen = last_seen.replace(tzinfo=timezone.utc)
        threshold = now - timedelta(minutes=DEVICE_ONLINE_THRESHOLD_MINUTES)
        return last_seen >= threshold

    def validate_device_type(self, t: str) -> bool:
        """Validate that the device type is one of the supported types.

        Args:
            t: The device type string to validate

        Returns:
            True if device type is valid, False otherwise

        Note:
            Valid device types are: laptop, desktop, server, mobile, tablet
        """
        return t in VALID_DEVICE_TYPES
