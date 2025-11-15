from . import users
from . import devices

# Provide both singular and plural aliases for compatibility with older imports
user = users
screenshot = devices.DeviceScreenshot
screenshots = devices  # Add plural alias for compatibility

__all__ = [
    "users",
    "user",
    "screenshot",
    "screenshots",
    "devices",
]
