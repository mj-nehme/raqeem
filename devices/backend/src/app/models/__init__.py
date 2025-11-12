from . import users
from . import devices

# Provide both singular and plural aliases for compatibility with older imports
user = users
screenshot = devices.DeviceScreenshot

__all__ = [
    "users",
    "user",
    "screenshot",
    "devices",
]
