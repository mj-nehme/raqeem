"""Device service for validating and managing device data."""

from datetime import datetime
from typing import Any


class DeviceService:
    """Service for validating and managing device data.

    This is a stub service primarily used for testing.
    Methods can be patched in tests to customize behavior.
    """

    def validate_device_data(self, _data: dict[str, Any]) -> bool:
        """Validate device data structure.

        Args:
            _data: Dictionary containing device data fields

        Returns:
            True if device data is valid, False otherwise
        """
        return True

    def is_device_online(self, _last_seen: datetime) -> bool:
        """Check if a device is currently online based on last seen timestamp.

        Args:
            _last_seen: The last time the device was seen

        Returns:
            True if device is considered online, False otherwise
        """
        return True

    def validate_device_type(self, t: str) -> bool:
        """Validate that the device type is one of the supported types.

        Args:
            t: The device type string to validate

        Returns:
            True if device type is valid, False otherwise

        Note:
            Valid device types are: laptop, desktop, server, mobile, tablet
        """
        return t in {"laptop", "desktop", "server", "mobile", "tablet"}
