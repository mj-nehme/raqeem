from . import users
from . import screenshots
from . import devices

# Provide both singular and plural aliases for compatibility with older imports
user = users
screenshot = screenshots

__all__ = [
    "users",
    "user",
    "screenshots",
    "screenshot",
    "devices",
]
