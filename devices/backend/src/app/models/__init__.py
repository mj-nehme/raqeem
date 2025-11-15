from . import devices, users

# Provide both singular and plural aliases for compatibility with older imports
user = users
screenshot = devices.DeviceScreenshot
screenshots = devices  # Add plural alias for compatibility

__all__ = [
    "devices",
    "screenshot",
    "screenshots",
    "user",
    "users",
]
